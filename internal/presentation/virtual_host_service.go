package presentation

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/janmbaco/go-reverseproxy-ssl/v3/internal/domain"
	"github.com/janmbaco/go-reverseproxy-ssl/v3/internal/infrastructure/grpcutil"
)

// VirtualHostService implementa la responsabilidad de gestionar operaciones de virtual hosts
type VirtualHostService struct {
	certificateService ICertificateService
	fileService        IFileService
	formHandler        IFormHandler
	vhResolver         domain.VirtualHostResolver
	logger             domain.Logger
}

// NewVirtualHostService crea una nueva instancia de VirtualHostService
func NewVirtualHostService(certificateService ICertificateService, fileService IFileService, formHandler IFormHandler, vhResolver domain.VirtualHostResolver, logger domain.Logger) IVirtualHostService {
	return &VirtualHostService{
		certificateService: certificateService,
		fileService:        fileService,
		formHandler:        formHandler,
		vhResolver:         vhResolver,
		logger:             logger,
	}
}

// CreateVirtualHost implementa IVirtualHostService.CreateVirtualHost
func (vhs *VirtualHostService) CreateVirtualHost(r *http.Request, config *domain.Config) (interface{}, error) {
	// Parse multipart form
	if err := r.ParseMultipartForm(32 << 20); err != nil { // 32MB max
		return nil, fmt.Errorf("failed to parse form: %v", err)
	}

	// Get virtual host type
	vhType := r.FormValue("virtualHostType")
	if vhType == "" {
		vhType = "web" // default to web
	}

	switch vhType {
	case "web":
		return vhs.createWebVirtualHost(r, config)
	case "grpc-web":
		return vhs.createGrpcWebVirtualHost(r, config)
	default:
		return nil, fmt.Errorf("unsupported virtual host type: %s", vhType)
	}
}

// createWebVirtualHost creates a WebVirtualHost from the form data
func (vhs *VirtualHostService) createWebVirtualHost(r *http.Request, config *domain.Config) (*domain.WebVirtualHost, error) {
	// Get form values
	from := r.FormValue("from")
	scheme := r.FormValue("scheme")
	hostName := r.FormValue("hostName")
	pathValue := r.FormValue("path")

	portStr := r.FormValue("port")
	port := uint(8080)
	if portStr != "" {
		if p, err := strconv.ParseUint(portStr, 10, 32); err == nil {
			port = uint(p)
		}
	}

	vhs.logger.Info(fmt.Sprintf("Creating web virtual host: from=%s, scheme=%s, host=%s, port=%d", from, scheme, hostName, port))

	// Create certificate directory
	certDir, err := vhs.fileService.CreateCertDirectory(config, from)
	if err != nil {
		return nil, fmt.Errorf("failed to create cert directory: %v", err)
	}

	// Handle server certificates
	serverCert, err := vhs.certificateService.HandleServerCertificates(r, certDir)
	if err != nil {
		return nil, fmt.Errorf("failed to handle server certificates: %v", err)
	}

	// Handle client certificates
	clientCert, err := vhs.certificateService.HandleClientCertificates(r, certDir)
	if err != nil {
		return nil, fmt.Errorf("failed to handle client certificates: %v", err)
	}

	// Create and return virtual host
	newVH := &domain.WebVirtualHost{
		ClientCertificateHost: domain.ClientCertificateHost{
			VirtualHostBase: domain.VirtualHostBase{
				From:              from,
				Scheme:            scheme,
				HostName:          hostName,
				Port:              port,
				Path:              pathValue,
				ServerCertificate: serverCert,
			},
			ClientCertificate: clientCert,
		},
		ResponseHeaders:  make(map[string]string), // Initialize empty map
		NeedPkFromClient: false,                   // Default to false
	}
	newVH.EnsureID()
	newVH.SetURLToReplace() // Initialize URL replacement fields

	vhs.logger.Info(fmt.Sprintf("Created new web virtual host with ID=%s", newVH.GetID()))

	return newVH, nil
}

// createGrpcWebVirtualHost creates a GrpcWebVirtualHost from the form data
func (vhs *VirtualHostService) createGrpcWebVirtualHost(r *http.Request, config *domain.Config) (*domain.GrpcWebVirtualHost, error) {
	// Get form values
	from := r.FormValue("from")
	grpcHostName := r.FormValue("grpcHostName")
	grpcPortStr := r.FormValue("grpcPort")
	authority := r.FormValue("authority")

	grpcPort := uint(7656) // default gRPC port
	if grpcPortStr != "" {
		if p, err := strconv.ParseUint(grpcPortStr, 10, 32); err == nil {
			grpcPort = uint(p)
		}
	}

	vhs.logger.Info(fmt.Sprintf("Creating gRPC-Web virtual host: from=%s, grpcHost=%s, grpcPort=%d", from, grpcHostName, grpcPort))

	// Parse gRPC services and methods
	grpcServices := make(map[string][]string)
	serviceNames := r.Form["grpcServiceName[]"]
	methods := r.Form["grpcMethods[]"]

	// Group methods by service - each method corresponds to the service at the same index
	for i, method := range methods {
		if method == "" {
			continue
		}
		serviceName := ""
		if i < len(serviceNames) {
			serviceName = serviceNames[i]
		}
		if serviceName == "" {
			continue
		}
		if grpcServices[serviceName] == nil {
			grpcServices[serviceName] = []string{}
		}
		grpcServices[serviceName] = append(grpcServices[serviceName], method)
	}

	// Parse CORS settings
	allowAllOrigins := r.FormValue("allowAllOrigins") == "on"
	allowedOrigins := r.Form["allowedOrigins[]"]
	useWebSockets := r.FormValue("useWebSockets") == "on"
	allowedHeaders := r.Form["allowedHeaders[]"]

	// Create certificate directory
	certDir, err := vhs.fileService.CreateCertDirectory(config, from)
	if err != nil {
		return nil, fmt.Errorf("failed to create cert directory: %v", err)
	}

	// Handle server certificates
	serverCert, err := vhs.certificateService.HandleServerCertificates(r, certDir)
	if err != nil {
		return nil, fmt.Errorf("failed to handle server certificates: %v", err)
	}

	// Handle client certificates
	clientCert, err := vhs.certificateService.HandleClientCertificates(r, certDir)
	if err != nil {
		return nil, fmt.Errorf("failed to handle client certificates: %v", err)
	}

	// Create and return virtual host
	newVH := &domain.GrpcWebVirtualHost{
		ClientCertificateHost: domain.ClientCertificateHost{
			VirtualHostBase: domain.VirtualHostBase{
				From:              from,
				Scheme:            "https", // gRPC-Web always uses HTTPS
				HostName:          grpcHostName,
				Port:              grpcPort,
				ServerCertificate: serverCert,
			},
			ClientCertificate: clientCert,
		},
		GrpcWebProxy: &grpcutil.GrpcWebProxy{
			GrpcProxy: grpcutil.GrpcProxy{
				GrpcServices:        grpcServices,
				IsTransparentServer: r.FormValue("isTransparentServer") == "on",
				Authority:           authority,
			},
			AllowAllOrigins: allowAllOrigins,
			AllowedOrigins:  allowedOrigins,
			UseWebSockets:   useWebSockets,
			AllowedHeaders:  allowedHeaders,
		},
	}
	newVH.EnsureID()
	newVH.SetURLToReplace() // Initialize URL replacement fields

	vhs.logger.Info(fmt.Sprintf("Created new gRPC-Web virtual host with ID=%s", newVH.GetID()))

	return newVH, nil
}

// UpdateVirtualHost implementa IVirtualHostService.UpdateVirtualHost
func (vhs *VirtualHostService) UpdateVirtualHost(r *http.Request, id string, config *domain.Config) (interface{}, []string, []string, error) {
	// Parse multipart form
	if err := r.ParseMultipartForm(32 << 20); err != nil { // 32MB max
		return nil, nil, nil, fmt.Errorf("failed to parse form: %v", err)
	}

	// Get virtual host type - try to determine from existing or form
	vhType := r.FormValue("virtualHostType")
	if vhType == "" {
		// Try to determine from existing virtual host
		if existingVH, _, err := vhs.findVirtualHostByID(config, id); err == nil {
			switch existingVH.(type) {
			case *domain.WebVirtualHost:
				vhType = "web"
			case *domain.GrpcWebVirtualHost:
				vhType = "grpc-web"
			}
		} else {
			vhType = "web" // default
		}
	}

	switch vhType {
	case "web":
		return vhs.updateWebVirtualHost(r, id, config)
	case "grpc-web":
		return vhs.updateGrpcWebVirtualHost(r, id, config)
	default:
		return nil, nil, nil, fmt.Errorf("unsupported virtual host type: %s", vhType)
	}
}

// updateWebVirtualHost updates a WebVirtualHost from the form data
func (vhs *VirtualHostService) updateWebVirtualHost(r *http.Request, id string, config *domain.Config) (*domain.WebVirtualHost, []string, []string, error) {
	// Find existing virtual host
	oldVH, _, err := vhs.findVirtualHostByID(config, id)
	if err != nil {
		return nil, nil, nil, err
	}
	webVH, ok := oldVH.(*domain.WebVirtualHost)
	if !ok {
		return nil, nil, nil, fmt.Errorf("existing virtual host is not a WebVirtualHost")
	}

	vhs.logger.Info(fmt.Sprintf("Updating web virtual host ID=%s, From=%s", id, webVH.From))

	// Get form values with defaults from existing
	from := vhs.getFormValueOrDefault(r, "from", webVH.From)
	port := vhs.parsePortFromForm(r, webVH.Port)

	// Create certificate directory
	certDir, err := vhs.fileService.CreateCertDirectory(config, from)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to create cert directory: %v", err)
	}

	// Handle certificate updates
	serverCert, clientCert, oldServerPaths, oldClientPaths, err := vhs.certificateService.HandleCertificateUpdates(r, certDir, webVH)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to handle certificates: %v", err)
	}

	// Create new virtual host
	newVH := &domain.WebVirtualHost{
		ClientCertificateHost: domain.ClientCertificateHost{
			VirtualHostBase: domain.VirtualHostBase{
				From:              from,
				Scheme:            r.FormValue("scheme"),
				HostName:          r.FormValue("hostName"),
				Port:              port,
				Path:              r.FormValue("path"),
				ServerCertificate: serverCert,
			},
			ClientCertificate: clientCert,
		},
		ResponseHeaders:  make(map[string]string), // Initialize empty map
		NeedPkFromClient: false,                   // Default to false
	}
	newVH.EnsureID()        // Generate new ID
	newVH.SetURLToReplace() // Initialize URL replacement fields

	vhs.logger.Info(fmt.Sprintf("Created new web virtual host with ID=%s", newVH.GetID()))

	return newVH, oldServerPaths, oldClientPaths, nil
}

// updateGrpcWebVirtualHost updates a GrpcWebVirtualHost from the form data
func (vhs *VirtualHostService) updateGrpcWebVirtualHost(r *http.Request, id string, config *domain.Config) (*domain.GrpcWebVirtualHost, []string, []string, error) {
	// Find existing virtual host
	oldVH, _, err := vhs.findVirtualHostByID(config, id)
	if err != nil {
		return nil, nil, nil, err
	}
	grpcVH, ok := oldVH.(*domain.GrpcWebVirtualHost)
	if !ok {
		return nil, nil, nil, fmt.Errorf("existing virtual host is not a GrpcWebVirtualHost")
	}

	vhs.logger.Info(fmt.Sprintf("Updating gRPC-Web virtual host ID=%s, From=%s", id, grpcVH.From))

	// Get form values with defaults from existing
	from := vhs.getFormValueOrDefault(r, "from", grpcVH.From)
	authority := vhs.getFormValueOrDefault(r, "authority", grpcVH.GrpcWebProxy.Authority)
	grpcHostName := vhs.getFormValueOrDefault(r, "grpcHostName", grpcVH.HostName)
	grpcPort := vhs.parsePortFromFormField(r, "grpcPort", grpcVH.Port)

	// Parse gRPC services and methods
	grpcServices := make(map[string][]string)
	serviceNames := r.Form["grpcServiceName[]"]
	methods := r.Form["grpcMethods[]"]

	// Group methods by service - each method corresponds to the service at the same index
	for i, method := range methods {
		if method == "" {
			continue
		}
		serviceName := ""
		if i < len(serviceNames) {
			serviceName = serviceNames[i]
		}
		if serviceName == "" {
			continue
		}
		if grpcServices[serviceName] == nil {
			grpcServices[serviceName] = []string{}
		}
		grpcServices[serviceName] = append(grpcServices[serviceName], method)
	}

	// Parse CORS settings
	allowAllOrigins := r.FormValue("allowAllOrigins") == "on"
	allowedOrigins := r.Form["allowedOrigins[]"]
	useWebSockets := r.FormValue("useWebSockets") == "on"
	allowedHeaders := r.Form["allowedHeaders[]"]

	// Create certificate directory
	certDir, err := vhs.fileService.CreateCertDirectory(config, from)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to create cert directory: %v", err)
	}

	// Handle certificate updates
	serverCert, clientCert, oldServerPaths, oldClientPaths, err := vhs.certificateService.HandleCertificateUpdates(r, certDir, grpcVH)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to handle certificates: %v", err)
	}

	// Create new virtual host
	newVH := &domain.GrpcWebVirtualHost{
		ClientCertificateHost: domain.ClientCertificateHost{
			VirtualHostBase: domain.VirtualHostBase{
				From:              from,
				Scheme:            "https", // gRPC-Web always uses HTTPS
				HostName:          grpcHostName,
				Port:              grpcPort,
				ServerCertificate: serverCert,
			},
			ClientCertificate: clientCert,
		},
		GrpcWebProxy: &grpcutil.GrpcWebProxy{
			GrpcProxy: grpcutil.GrpcProxy{
				GrpcServices:        grpcServices,
				IsTransparentServer: r.FormValue("isTransparentServer") == "on",
				Authority:           authority,
			},
			AllowAllOrigins: allowAllOrigins,
			AllowedOrigins:  allowedOrigins,
			UseWebSockets:   useWebSockets,
			AllowedHeaders:  allowedHeaders,
		},
	}
	newVH.EnsureID()        // Generate new ID
	newVH.SetURLToReplace() // Initialize URL replacement fields

	vhs.logger.Info(fmt.Sprintf("Created new gRPC-Web virtual host with ID=%s", newVH.GetID()))

	return newVH, oldServerPaths, oldClientPaths, nil
}

// DeleteVirtualHost implementa IVirtualHostService.DeleteVirtualHost
func (vhs *VirtualHostService) DeleteVirtualHost(id string, config *domain.Config) (string, error) {
	vhs.logger.Info("Attempting to delete virtual host with ID: " + id)

	// Collect certificate paths before deletion
	serverCertPaths, clientCertPaths := vhs.certificateService.CollectCertPathsFromVH(config, id)

	// Delete virtual host from config
	deletedFrom, deleted := vhs.removeVirtualHostByID(config, id)

	if !deleted {
		vhs.logger.Error("Virtual host not found with ID: " + id)
		return "", fmt.Errorf("virtual host not found")
	}

	// Check which certificates are still in use
	certsInUse := vhs.certificateService.CollectAllCertsInUse(config)

	// Delete unused certificate files
	vhs.certificateService.DeleteUnusedCertFiles(serverCertPaths, clientCertPaths, certsInUse)

	// Delete certificate directory if empty
	vhs.certificateService.CleanupCertDirectory(config, deletedFrom)

	vhs.logger.Info(fmt.Sprintf("Virtual host deleted successfully: From='%s', ID='%s'", deletedFrom, id))
	return deletedFrom, nil
}

// GetVirtualHost implementa IVirtualHostService.GetVirtualHost
func (vhs *VirtualHostService) GetVirtualHost(id string, config *domain.Config) (interface{}, string, error) {
	// Search for virtual host by ID using GetID() method
	var foundVH interface{}
	var vhType string

	// Check WebVirtualHosts
	for _, vh := range config.WebVirtualHosts {
		if vh.GetID() == id {
			foundVH = vh
			vhType = "web"
			break
		}
	}

	// Check GrpcWebVirtualHosts
	if foundVH == nil {
		for _, vh := range config.GrpcWebVirtualHosts {
			if vh.GetID() == id {
				foundVH = vh
				vhType = "grpc-web"
				break
			}
		}
	}

	if foundVH == nil {
		return nil, "", fmt.Errorf("virtual host not found")
	}

	return foundVH, vhType, nil
}

// GetVirtualHosts implementa IVirtualHostService.GetVirtualHosts
func (vhs *VirtualHostService) GetVirtualHosts(config *domain.Config) ([]domain.IVirtualHost, error) {
	return vhs.vhResolver.Resolve(config)
}

// CleanupUnusedCertificates implementa IVirtualHostService.CleanupUnusedCertificates
func (vhs *VirtualHostService) CleanupUnusedCertificates(config *domain.Config, oldServerPaths, oldClientPaths []string) {
	certsInUse := vhs.certificateService.CollectAllCertsInUse(config)
	vhs.certificateService.DeleteUnusedCertFiles(oldServerPaths, oldClientPaths, certsInUse)
}

// Helper methods

func (vhs *VirtualHostService) findVirtualHostByID(config *domain.Config, id string) (interface{}, int, error) {
	// Check WebVirtualHosts
	for i, vh := range config.WebVirtualHosts {
		if vh.GetID() == id {
			return vh, i, nil
		}
	}
	// Check GrpcWebVirtualHosts
	for i, vh := range config.GrpcWebVirtualHosts {
		if vh.GetID() == id {
			return vh, i, nil
		}
	}
	return nil, -1, fmt.Errorf("virtual host not found")
}

func (vhs *VirtualHostService) getFormValueOrDefault(r *http.Request, field, defaultValue string) string {
	value := r.FormValue(field)
	if value == "" {
		return defaultValue
	}
	return value
}

func (vhs *VirtualHostService) parsePortFromForm(r *http.Request, defaultPort uint) uint {
	portStr := r.FormValue("port")
	if portStr == "" {
		return defaultPort
	}
	if p, err := strconv.ParseUint(portStr, 10, 32); err == nil {
		return uint(p)
	}
	return defaultPort
}

func (vhs *VirtualHostService) parsePortFromFormField(r *http.Request, fieldName string, defaultPort uint) uint {
	portStr := r.FormValue(fieldName)
	if portStr == "" {
		return defaultPort
	}
	if p, err := strconv.ParseUint(portStr, 10, 32); err == nil {
		return uint(p)
	}
	return defaultPort
}

func (vhs *VirtualHostService) removeVirtualHostByID(config *domain.Config, id string) (string, bool) {
	// Ensure all virtual hosts have IDs
	for _, vh := range config.WebVirtualHosts {
		vh.EnsureID()
	}
	for _, vh := range config.GrpcWebVirtualHosts {
		vh.EnsureID()
	}

	// Try WebVirtualHosts
	for i, vh := range config.WebVirtualHosts {
		if vh.GetID() == id {
			vhs.logger.Info(fmt.Sprintf("Found virtual host in WebVirtualHosts: From='%s', ID='%s'", vh.From, vh.GetID()))
			deletedFrom := vh.From
			config.WebVirtualHosts = append(config.WebVirtualHosts[:i], config.WebVirtualHosts[i+1:]...)
			return deletedFrom, true
		}
	}

	// Try GrpcWebVirtualHosts
	for i, vh := range config.GrpcWebVirtualHosts {
		if vh.GetID() == id {
			vhs.logger.Info(fmt.Sprintf("Found virtual host in GrpcWebVirtualHosts: From='%s', ID='%s'", vh.From, vh.GetID()))
			deletedFrom := vh.From
			config.GrpcWebVirtualHosts = append(config.GrpcWebVirtualHosts[:i], config.GrpcWebVirtualHosts[i+1:]...)
			return deletedFrom, true
		}
	}

	return "", false
}
