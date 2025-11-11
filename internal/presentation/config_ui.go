package presentation

import (
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/janmbaco/go-infrastructure/v2/configuration"
	"github.com/janmbaco/go-reverseproxy-ssl/v3/internal/domain"
	certificates "github.com/janmbaco/go-reverseproxy-ssl/v3/internal/infrastructure/certificates"
)

//go:embed static/css/* static/js/*
var staticFS embed.FS

//go:embed templates/layouts/*.html templates/pages/*.html
var templatesFS embed.FS

type ConfigUI struct {
	configHandler configuration.ConfigHandler
	vhResolver    domain.VirtualHostResolver
	templates     *template.Template
	logger        domain.Logger
}

func NewConfigUI(configHandler configuration.ConfigHandler, vhResolver domain.VirtualHostResolver, logger domain.Logger) *ConfigUI {
	// Load templates with error handling
	templates, err := template.ParseFS(templatesFS, "templates/layouts/*.html", "templates/pages/*.html")
	if err != nil {
		logger.Error("Failed to load templates: " + err.Error())
		// Don't panic - return nil and let the caller handle it
		return nil
	}

	logger.Info("Templates loaded successfully")

	return &ConfigUI{
		configHandler: configHandler,
		vhResolver:    vhResolver,
		templates:     templates,
		logger:        logger,
	}
}

func (cui *ConfigUI) SetupRoutes(mux *http.ServeMux) {
	cui.logger.Info("Setting up ConfigUI routes")

	// Static files - create sub filesystem
	staticSub, err := fs.Sub(staticFS, "static")
	if err != nil {
		cui.logger.Error("Failed to create static sub FS: " + err.Error())
	} else {
		mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticSub))))
		cui.logger.Info("Static files handler set up")
	}

	// Wrap all handlers with recovery middleware to prevent panics from crashing the server
	recoverFunc := func(handler http.HandlerFunc) http.HandlerFunc {
		return RecoveryHandlerFunc(cui.logger, handler)
	}

	// Page routes (all wrapped with recovery)
	mux.HandleFunc("/", recoverFunc(cui.handleDashboard))
	mux.HandleFunc("/config", recoverFunc(cui.handleConfig))
	mux.HandleFunc("/virtualhosts", recoverFunc(cui.handleVirtualHosts))
	mux.HandleFunc("/virtualhosts/new", recoverFunc(cui.handleNewVirtualHost))
	mux.HandleFunc("/virtualhosts/edit/", recoverFunc(func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/virtualhosts/edit/")
		cui.handleEditVirtualHost(w, r, path)
	}))
	mux.HandleFunc("/virtualhosts/", recoverFunc(cui.handleVirtualHostActions))

	// API routes (all wrapped with recovery)
	mux.HandleFunc("/api/config", recoverFunc(cui.handleGetConfig))
	mux.HandleFunc("/api/config/update", recoverFunc(cui.handleUpdateConfig))
	mux.HandleFunc("/api/virtualhosts", recoverFunc(cui.handleVirtualHostsAPI))
	mux.HandleFunc("/api/virtualhosts/", recoverFunc(cui.handleVirtualHostAPI))

	cui.logger.Info("ConfigUI routes set up with panic recovery")
}

func (cui *ConfigUI) handleDashboard(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	config := cui.configHandler.GetConfig().(*domain.Config)
	vhCollection, err := cui.vhResolver.Resolve(config)
	if err != nil {
		http.Error(w, "Failed to resolve virtual hosts", http.StatusInternalServerError)
		return
	}

	data := struct {
		Title        string
		ActivePage   string
		Template     string
		Config       *domain.Config
		VirtualHosts []domain.IVirtualHost
		IsLocalhost  bool
	}{
		Title:        "Dashboard - Reverse Proxy Config",
		ActivePage:   "dashboard",
		Template:     "dashboard-content",
		Config:       config,
		VirtualHosts: vhCollection,
		IsLocalhost:  strings.Contains(r.Host, "localhost") || strings.Contains(r.Host, "127.0.0.1"),
	}

	w.Header().Set("Content-Type", "text/html")
	if err := cui.templates.ExecuteTemplate(w, "base", data); err != nil {
		cui.logger.Error("Template execution error: " + err.Error())
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
	}
}

func (cui *ConfigUI) handleConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	config := cui.configHandler.GetConfig().(*domain.Config)

	// Convert config to JSON for raw display
	configJSON, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		configJSON = []byte("Error marshaling config to JSON")
	}

	data := struct {
		Title      string
		ActivePage string
		Template   string
		Config     *domain.Config
		RawConfig  string
	}{
		Title:      "Configuration - Reverse Proxy Config",
		ActivePage: "config",
		Template:   "config-content",
		Config:     config,
		RawConfig:  string(configJSON),
	}

	w.Header().Set("Content-Type", "text/html")
	if err := cui.templates.ExecuteTemplate(w, "base", data); err != nil {
		cui.logger.Error("Template execution error: " + err.Error())
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
	}
}

func (cui *ConfigUI) handleVirtualHosts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	config := cui.configHandler.GetConfig().(*domain.Config)
	vhCollection, err := cui.vhResolver.Resolve(config)
	if err != nil {
		http.Error(w, "Failed to resolve virtual hosts", http.StatusInternalServerError)
		return
	}

	data := struct {
		Title        string
		ActivePage   string
		Template     string
		VirtualHosts []domain.IVirtualHost
	}{
		Title:        "Virtual Hosts - Reverse Proxy Config",
		ActivePage:   "virtualhosts",
		Template:     "virtualhosts-content",
		VirtualHosts: vhCollection,
	}

	w.Header().Set("Content-Type", "text/html")
	if err := cui.templates.ExecuteTemplate(w, "base", data); err != nil {
		cui.logger.Error("Template execution error: " + err.Error())
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
	}
}

func (cui *ConfigUI) handleNewVirtualHost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	data := struct {
		Title       string
		ActivePage  string
		Template    string
		IsEdit      bool
		VirtualHost *domain.WebVirtualHost
	}{
		Title:      "New Virtual Host - Reverse Proxy Config",
		ActivePage: "virtualhosts",
		Template:   "virtualhost-form-content",
		IsEdit:     false,
		VirtualHost: &domain.WebVirtualHost{
			ClientCertificateHost: domain.ClientCertificateHost{
				VirtualHostBase: domain.VirtualHostBase{
					Scheme: "http",
					Port:   8080,
				},
			},
		},
	}

	w.Header().Set("Content-Type", "text/html")
	if err := cui.templates.ExecuteTemplate(w, "base", data); err != nil {
		cui.logger.Error("Template execution error: " + err.Error())
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
	}
}

func (cui *ConfigUI) handleVirtualHostActions(w http.ResponseWriter, r *http.Request) {
	// Extract action and ID from URL path
	path := strings.TrimPrefix(r.URL.Path, "/virtualhosts/")
	parts := strings.Split(path, "/")

	if len(parts) < 2 {
		http.Redirect(w, r, "/virtualhosts", http.StatusSeeOther)
		return
	}

	action := parts[0]
	id := parts[1]

	switch action {
	case "edit":
		cui.handleEditVirtualHost(w, r, id)
	case "delete":
		if r.Method == http.MethodPost {
			cui.handleDeleteVirtualHost(w, r, id)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	default:
		http.Redirect(w, r, "/virtualhosts", http.StatusSeeOther)
	}
}

func (cui *ConfigUI) handleEditVirtualHost(w http.ResponseWriter, r *http.Request, id string) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	config := cui.configHandler.GetConfig().(*domain.Config)

	// Find the virtual host
	var foundVH *domain.WebVirtualHost
	for _, vh := range config.WebVirtualHosts {
		if vh.GetID() == id {
			// Create a copy of the virtual host
			foundVH = &domain.WebVirtualHost{
				ClientCertificateHost: domain.ClientCertificateHost{
					VirtualHostBase: vh.VirtualHostBase,
				},
			}
			break
		}
	}

	if foundVH == nil {
		http.Error(w, "Virtual host not found", http.StatusNotFound)
		return
	}

	data := struct {
		Title       string
		ActivePage  string
		Template    string
		IsEdit      bool
		VirtualHost *domain.WebVirtualHost
	}{
		Title:       "Edit Virtual Host - Reverse Proxy Config",
		ActivePage:  "virtualhosts",
		Template:    "virtualhost-form-content",
		IsEdit:      true,
		VirtualHost: foundVH,
	}

	w.Header().Set("Content-Type", "text/html")
	if err := cui.templates.ExecuteTemplate(w, "base", data); err != nil {
		cui.logger.Error("Template execution error: " + err.Error())
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
	}
}

func (cui *ConfigUI) handleDeleteVirtualHost(w http.ResponseWriter, r *http.Request, id string) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Call the API delete method
	cui.deleteVirtualHost(w, r, id)
}

func (cui *ConfigUI) handleGetConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	config := cui.configHandler.GetConfig()
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(config)
}

func (cui *ConfigUI) handleUpdateConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var newConfig domain.Config
	if err := json.NewDecoder(r.Body).Decode(&newConfig); err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
		return
	}
	if err := cui.configHandler.SetConfig(&newConfig); err != nil {
		http.Error(w, fmt.Sprintf("Failed to update config: %v", err), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func (cui *ConfigUI) handleVirtualHostsAPI(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		cui.getVirtualHosts(w, r)
	case http.MethodPost:
		cui.createVirtualHost(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (cui *ConfigUI) handleVirtualHostAPI(w http.ResponseWriter, r *http.Request) {
	// Extract ID from URL path
	path := r.URL.Path
	if len(path) < len("/api/virtualhosts/") {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}
	id := path[len("/api/virtualhosts/"):]

	switch r.Method {
	case http.MethodGet:
		cui.getVirtualHost(w, r, id)
	case http.MethodPost:
		cui.updateVirtualHost(w, r, id)
	case http.MethodDelete:
		cui.deleteVirtualHost(w, r, id)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (cui *ConfigUI) getVirtualHosts(w http.ResponseWriter, r *http.Request) {
	config := cui.configHandler.GetConfig().(*domain.Config)
	vhCollection, err := cui.vhResolver.Resolve(config)
	if err != nil {
		http.Error(w, "Failed to resolve virtual hosts", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success":      true,
		"virtualHosts": vhCollection,
	})
}

func (cui *ConfigUI) createVirtualHost(w http.ResponseWriter, r *http.Request) {
	// Parse multipart form
	if err := r.ParseMultipartForm(32 << 20); err != nil { // 32MB max
		http.Error(w, fmt.Sprintf("Failed to parse form: %v", err), http.StatusBadRequest)
		return
	}

	config := cui.configHandler.GetConfig().(*domain.Config)

	// Get cert base directory from config or use default
	certBaseDir := config.CertDir
	if certBaseDir == "" {
		certBaseDir = "/app/certs"
	}

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

	cui.logger.Info(fmt.Sprintf("Creating virtual host: from=%s, scheme=%s, host=%s, port=%d", from, scheme, hostName, port))

	// Create cert directory based on 'from' (sanitize for filesystem)
	safeName := sanitizePathName(from)
	certDir := filepath.Join(certBaseDir, safeName)

	if err := os.MkdirAll(certDir, 0755); err != nil {
		cui.logger.Error(fmt.Sprintf("Failed to create cert directory %s: %v", certDir, err))
		http.Error(w, fmt.Sprintf("Failed to create cert directory: %v", err), http.StatusInternalServerError)
		return
	}

	var serverCert *certificates.CertificateDefs
	var clientCert *certificates.CertificateDefs

	// Save server certificate if provided
	useServerCert := r.FormValue("useServerCert") == "on"
	if useServerCert {
		if _, header, err := r.FormFile("serverCertFile"); err == nil {
			serverCertPath := filepath.Join(certDir, header.Filename)
			if err := cui.saveCertificateFromForm(r, "serverCertFile", serverCertPath); err != nil {
				cui.logger.Error(fmt.Sprintf("Failed to save server cert: %v", err))
				http.Error(w, fmt.Sprintf("Failed to save server cert: %v", err), http.StatusInternalServerError)
				return
			}
			cui.logger.Info(fmt.Sprintf("Saved server certificate to %s", serverCertPath))

			if serverCert == nil {
				serverCert = &certificates.CertificateDefs{}
			}
			serverCert.PublicKey = serverCertPath
		}

		// Save server key if provided
		if _, header, err := r.FormFile("serverKeyFile"); err == nil {
			serverKeyPath := filepath.Join(certDir, header.Filename)
			if err := cui.saveCertificateFromForm(r, "serverKeyFile", serverKeyPath); err != nil {
				cui.logger.Error(fmt.Sprintf("Failed to save server key: %v", err))
				http.Error(w, fmt.Sprintf("Failed to save server key: %v", err), http.StatusInternalServerError)
				return
			}
			cui.logger.Info(fmt.Sprintf("Saved server key to %s", serverKeyPath))

			if serverCert == nil {
				serverCert = &certificates.CertificateDefs{}
			}
			serverCert.PrivateKey = serverKeyPath
		}
	}

	// Save client certificate if provided
	useClientCert := r.FormValue("useClientCert") == "on"
	if useClientCert {
		if _, header, err := r.FormFile("clientCertFile"); err == nil {
			clientCertPath := filepath.Join(certDir, header.Filename)
			if err := cui.saveCertificateFromForm(r, "clientCertFile", clientCertPath); err != nil {
				cui.logger.Error(fmt.Sprintf("Failed to save client cert: %v", err))
				http.Error(w, fmt.Sprintf("Failed to save client cert: %v", err), http.StatusInternalServerError)
				return
			}
			cui.logger.Info(fmt.Sprintf("Saved client CA certificate to %s", clientCertPath))

			if clientCert == nil {
				clientCert = &certificates.CertificateDefs{}
			}
			clientCert.CaPem = []string{clientCertPath}
		}
	}

	// Create new virtual host
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
	}
	newVH.EnsureID()

	cui.logger.Info(fmt.Sprintf("Created new virtual host with ID=%s", newVH.GetID()))

	// Add to config
	config.WebVirtualHosts = append(config.WebVirtualHosts, newVH)

	if err := cui.configHandler.SetConfig(config); err != nil {
		cui.logger.Error(fmt.Sprintf("Failed to save config: %v", err))
		http.Error(w, fmt.Sprintf("Failed to save config: %v", err), http.StatusInternalServerError)
		return
	}

	cui.logger.Info("Virtual host created successfully")

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Virtual host created successfully",
		"id":      newVH.GetID(),
	})
}

func (cui *ConfigUI) getVirtualHost(w http.ResponseWriter, r *http.Request, id string) {
	config := cui.configHandler.GetConfig().(*domain.Config)

	// Search for virtual host by "From" field (which acts as ID)
	var foundVH interface{}
	var vhType string

	// Check WebVirtualHosts
	for _, vh := range config.WebVirtualHosts {
		if vh.From == id {
			foundVH = vh
			vhType = "web"
			break
		}
	}

	// Check GrpcWebVirtualHosts
	if foundVH == nil {
		for _, vh := range config.GrpcWebVirtualHosts {
			if vh.From == id {
				foundVH = vh
				vhType = "grpc-web"
				break
			}
		}
	}

	if foundVH == nil {
		http.Error(w, "Virtual host not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success":     true,
		"virtualHost": foundVH,
		"type":        vhType,
	})
}

func (cui *ConfigUI) updateVirtualHost(w http.ResponseWriter, r *http.Request, id string) {
	// Parse multipart form
	if err := r.ParseMultipartForm(32 << 20); err != nil { // 32MB max
		http.Error(w, fmt.Sprintf("Failed to parse form: %v", err), http.StatusBadRequest)
		return
	}

	config := cui.configHandler.GetConfig().(*domain.Config)

	// Get cert base directory from config or use default
	certBaseDir := config.CertDir
	if certBaseDir == "" {
		certBaseDir = "/app/certs"
	}

	// Find and remove the old virtual host
	var oldVH *domain.WebVirtualHost
	oldIndex := -1
	for i, vh := range config.WebVirtualHosts {
		if vh.GetID() == id {
			oldVH = vh
			oldIndex = i
			break
		}
	}

	if oldVH == nil {
		http.Error(w, "Virtual host not found", http.StatusNotFound)
		return
	}

	cui.logger.Info(fmt.Sprintf("Updating virtual host ID=%s, From=%s", id, oldVH.From))

	// Get form values
	from := r.FormValue("from")
	if from == "" {
		from = oldVH.From
	}

	portStr := r.FormValue("port")
	port := uint(8080)
	if portStr != "" {
		if p, err := strconv.ParseUint(portStr, 10, 32); err == nil {
			port = uint(p)
		}
	}

	// Create cert directory based on 'from' (sanitize for filesystem)
	safeName := sanitizePathName(from)
	certDir := filepath.Join(certBaseDir, safeName)

	if err := os.MkdirAll(certDir, 0755); err != nil {
		cui.logger.Error(fmt.Sprintf("Failed to create cert directory %s: %v", certDir, err))
		http.Error(w, fmt.Sprintf("Failed to create cert directory: %v", err), http.StatusInternalServerError)
		return
	}

	// Track old certificate paths for cleanup
	var oldServerCertPaths []string
	var oldClientCertPaths []string

	if oldVH.ServerCertificate != nil {
		if oldVH.ServerCertificate.PublicKey != "" {
			oldServerCertPaths = append(oldServerCertPaths, oldVH.ServerCertificate.PublicKey)
		}
		if oldVH.ServerCertificate.PrivateKey != "" {
			oldServerCertPaths = append(oldServerCertPaths, oldVH.ServerCertificate.PrivateKey)
		}
	}
	if oldVH.ClientCertificate != nil && len(oldVH.ClientCertificate.CaPem) > 0 {
		oldClientCertPaths = append(oldClientCertPaths, oldVH.ClientCertificate.CaPem...)
	}

	// Prepare certificate paths - preserve existing ones
	var serverCert *certificates.CertificateDefs
	var clientCert *certificates.CertificateDefs

	// Copy existing certificates
	if oldVH.ServerCertificate != nil {
		serverCert = &certificates.CertificateDefs{
			PublicKey:  oldVH.ServerCertificate.PublicKey,
			PrivateKey: oldVH.ServerCertificate.PrivateKey,
			CaPem:      oldVH.ServerCertificate.CaPem,
		}
	}
	if oldVH.ClientCertificate != nil {
		clientCert = &certificates.CertificateDefs{
			CaPem: oldVH.ClientCertificate.CaPem,
		}
	}

	// Save server certificate if provided
	if _, header, err := r.FormFile("serverCertFile"); err == nil {
		serverCertPath := filepath.Join(certDir, header.Filename)
		if err := cui.saveCertificateFromForm(r, "serverCertFile", serverCertPath); err != nil {
			cui.logger.Error(fmt.Sprintf("Failed to save server cert: %v", err))
			http.Error(w, fmt.Sprintf("Failed to save server cert: %v", err), http.StatusInternalServerError)
			return
		}
		cui.logger.Info(fmt.Sprintf("Saved server certificate to %s", serverCertPath))

		if serverCert == nil {
			serverCert = &certificates.CertificateDefs{}
		}
		serverCert.PublicKey = serverCertPath
	}

	// Save server key if provided
	if _, header, err := r.FormFile("serverKeyFile"); err == nil {
		serverKeyPath := filepath.Join(certDir, header.Filename)
		if err := cui.saveCertificateFromForm(r, "serverKeyFile", serverKeyPath); err != nil {
			cui.logger.Error(fmt.Sprintf("Failed to save server key: %v", err))
			http.Error(w, fmt.Sprintf("Failed to save server key: %v", err), http.StatusInternalServerError)
			return
		}
		cui.logger.Info(fmt.Sprintf("Saved server key to %s", serverKeyPath))

		if serverCert == nil {
			serverCert = &certificates.CertificateDefs{}
		}
		serverCert.PrivateKey = serverKeyPath
	}

	// Save client certificate if provided
	if _, header, err := r.FormFile("clientCertFile"); err == nil {
		clientCertPath := filepath.Join(certDir, header.Filename)
		if err := cui.saveCertificateFromForm(r, "clientCertFile", clientCertPath); err != nil {
			cui.logger.Error(fmt.Sprintf("Failed to save client cert: %v", err))
			http.Error(w, fmt.Sprintf("Failed to save client cert: %v", err), http.StatusInternalServerError)
			return
		}
		cui.logger.Info(fmt.Sprintf("Saved client CA certificate to %s", clientCertPath))

		if clientCert == nil {
			clientCert = &certificates.CertificateDefs{}
		}
		clientCert.CaPem = []string{clientCertPath}
	}

	// Create NEW virtual host with NEW ID
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
	}
	newVH.EnsureID() // Generate new ID

	cui.logger.Info(fmt.Sprintf("Created new virtual host with ID=%s to replace old ID=%s", newVH.GetID(), id))

	// Remove old virtual host
	config.WebVirtualHosts = append(config.WebVirtualHosts[:oldIndex], config.WebVirtualHosts[oldIndex+1:]...)

	// Add new virtual host
	config.WebVirtualHosts = append(config.WebVirtualHosts, newVH)

	if err := cui.configHandler.SetConfig(config); err != nil {
		http.Error(w, fmt.Sprintf("Failed to save config: %v", err), http.StatusInternalServerError)
		return
	}

	// Now that the new config is saved, clean up old certificates if they were replaced
	certsInUse := cui.collectAllCertsInUse(config)
	cui.deleteUnusedCertFiles(oldServerCertPaths, oldClientCertPaths, certsInUse)

	cui.logger.Info("Virtual host updated successfully with new ID")

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"newId":   newVH.GetID(),
	})
} // saveCertificateFromForm saves a certificate file from form data
func (cui *ConfigUI) saveCertificateFromForm(r *http.Request, fieldName, destPath string) error {
	file, _, err := r.FormFile(fieldName)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()

	destFile, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer func() { _ = destFile.Close() }()

	_, err = io.Copy(destFile, file)
	return err
}

func (cui *ConfigUI) deleteVirtualHost(w http.ResponseWriter, r *http.Request, id string) {
	cui.logger.Info("Attempting to delete virtual host with ID: " + id)

	config := cui.configHandler.GetConfig().(*domain.Config)

	// Collect certificate paths before deletion
	serverCertPaths, clientCertPaths := cui.collectCertPathsFromVH(config, id)

	// Delete virtual host from config
	deletedFrom, deleted := cui.removeVirtualHostByID(config, id)

	if !deleted {
		cui.logger.Error("Virtual host not found with ID: " + id)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Virtual host not found",
		})
		return
	}

	// Check which certificates are still in use
	certsInUse := cui.collectAllCertsInUse(config)

	// Delete unused certificate files
	cui.deleteUnusedCertFiles(serverCertPaths, clientCertPaths, certsInUse)

	// Delete certificate directory if empty
	cui.cleanupCertDirectory(config, deletedFrom)

	cui.logger.Info("Saving updated configuration...")
	if err := cui.configHandler.SetConfig(config); err != nil {
		cui.logger.Error("Failed to save config: " + err.Error())
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Failed to save configuration: " + err.Error(),
		})
		return
	}

	cui.logger.Info(fmt.Sprintf("Virtual host deleted successfully: From='%s', ID='%s'", deletedFrom, id))
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
	})
}

func (cui *ConfigUI) collectCertPathsFromVH(config *domain.Config, id string) ([]string, []string) {
	var serverCertPaths []string
	var clientCertPaths []string

	for _, vh := range config.WebVirtualHosts {
		if vh.GetID() == id {
			if vh.ServerCertificate != nil {
				if vh.ServerCertificate.PublicKey != "" {
					serverCertPaths = append(serverCertPaths, vh.ServerCertificate.PublicKey)
				}
				if vh.ServerCertificate.PrivateKey != "" {
					serverCertPaths = append(serverCertPaths, vh.ServerCertificate.PrivateKey)
				}
			}
			if vh.ClientCertificate != nil && len(vh.ClientCertificate.CaPem) > 0 {
				clientCertPaths = append(clientCertPaths, vh.ClientCertificate.CaPem...)
			}
			break
		}
	}

	return serverCertPaths, clientCertPaths
}

func (cui *ConfigUI) removeVirtualHostByID(config *domain.Config, id string) (string, bool) {
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
			cui.logger.Info(fmt.Sprintf("Found virtual host in WebVirtualHosts: From='%s', ID='%s'", vh.From, vh.GetID()))
			deletedFrom := vh.From
			config.WebVirtualHosts = append(config.WebVirtualHosts[:i], config.WebVirtualHosts[i+1:]...)
			return deletedFrom, true
		}
	}

	// Try GrpcWebVirtualHosts
	for i, vh := range config.GrpcWebVirtualHosts {
		if vh.GetID() == id {
			cui.logger.Info(fmt.Sprintf("Found virtual host in GrpcWebVirtualHosts: From='%s', ID='%s'", vh.From, vh.GetID()))
			deletedFrom := vh.From
			config.GrpcWebVirtualHosts = append(config.GrpcWebVirtualHosts[:i], config.GrpcWebVirtualHosts[i+1:]...)
			return deletedFrom, true
		}
	}

	return "", false
}

func (cui *ConfigUI) collectAllCertsInUse(config *domain.Config) map[string]bool {
	certsInUse := make(map[string]bool)

	for _, vh := range config.WebVirtualHosts {
		if vh.ServerCertificate != nil {
			if vh.ServerCertificate.PublicKey != "" {
				certsInUse[vh.ServerCertificate.PublicKey] = true
			}
			if vh.ServerCertificate.PrivateKey != "" {
				certsInUse[vh.ServerCertificate.PrivateKey] = true
			}
		}
		if vh.ClientCertificate != nil {
			for _, ca := range vh.ClientCertificate.CaPem {
				certsInUse[ca] = true
			}
		}
	}

	return certsInUse
}

func (cui *ConfigUI) deleteUnusedCertFiles(serverCertPaths, clientCertPaths []string, certsInUse map[string]bool) {
	for _, certPath := range serverCertPaths {
		if !certsInUse[certPath] {
			cui.logger.Info(fmt.Sprintf("Deleting unused certificate: %s", certPath))
			if err := os.Remove(certPath); err != nil {
				cui.logger.Error(fmt.Sprintf("Failed to delete certificate %s: %v", certPath, err))
			}
		} else {
			cui.logger.Info(fmt.Sprintf("Certificate %s is still in use, skipping deletion", certPath))
		}
	}

	for _, certPath := range clientCertPaths {
		if !certsInUse[certPath] {
			cui.logger.Info(fmt.Sprintf("Deleting unused client certificate: %s", certPath))
			if err := os.Remove(certPath); err != nil {
				cui.logger.Error(fmt.Sprintf("Failed to delete client certificate %s: %v", certPath, err))
			}
		} else {
			cui.logger.Info(fmt.Sprintf("Client certificate %s is still in use, skipping deletion", certPath))
		}
	}
}

func (cui *ConfigUI) cleanupCertDirectory(config *domain.Config, deletedFrom string) {
	certBaseDir := config.CertDir
	if certBaseDir == "" {
		certBaseDir = "/app/certs"
	}
	safeName := sanitizePathName(deletedFrom)
	certDir := filepath.Join(certBaseDir, safeName)

	cui.logger.Info(fmt.Sprintf("Checking certificate directory: %s", certDir))
	if entries, err := os.ReadDir(certDir); err == nil {
		if len(entries) == 0 {
			cui.logger.Info(fmt.Sprintf("Deleting empty certificate directory: %s", certDir))
			if err := os.Remove(certDir); err != nil {
				cui.logger.Error(fmt.Sprintf("Failed to delete directory %s: %v", certDir, err))
			}
		} else {
			cui.logger.Info(fmt.Sprintf("Directory %s still contains files, keeping it", certDir))
		}
	}
}

// sanitizePathName sanitizes a path name to prevent directory traversal attacks
// It extracts the domain/host part from URLs and sanitizes dangerous characters
func sanitizePathName(name string) string {
	// If name contains a URL-like format, extract just the host/domain part
	if strings.Contains(name, "://") {
		// Parse as URL to extract host
		if parts := strings.Split(name, "://"); len(parts) > 1 {
			hostAndPath := parts[1]
			// Take everything before the first '/' (path separator)
			if hostEnd := strings.Index(hostAndPath, "/"); hostEnd > 0 {
				name = hostAndPath[:hostEnd]
			} else {
				name = hostAndPath
			}
		}
	} else if strings.Contains(name, "/") {
		// If no protocol but has path, take only the domain part
		if hostEnd := strings.Index(name, "/"); hostEnd > 0 {
			name = name[:hostEnd]
		}
	}

	// Replace dangerous characters
	safeName := strings.ReplaceAll(name, "/", "-")
	safeName = strings.ReplaceAll(safeName, ":", "-")
	safeName = strings.ReplaceAll(safeName, "\\", "-")
	safeName = strings.ReplaceAll(safeName, "..", "-")
	safeName = strings.ReplaceAll(safeName, "<", "-")
	safeName = strings.ReplaceAll(safeName, ">", "-")
	safeName = strings.ReplaceAll(safeName, "|", "-")
	safeName = strings.ReplaceAll(safeName, "?", "-")
	safeName = strings.ReplaceAll(safeName, "*", "-")
	safeName = strings.ReplaceAll(safeName, "\"", "-")
	safeName = strings.ReplaceAll(safeName, "'", "-")

	// Remove any remaining dangerous sequences
	safeName = strings.ReplaceAll(safeName, "..", "")

	// Ensure it's not empty and doesn't start/end with dangerous chars
	safeName = strings.Trim(safeName, ".- ")

	// If empty after sanitization, use a default
	if safeName == "" {
		safeName = "default"
	}

	return safeName
}

func (cui *ConfigUI) Start(port string) {
	cui.logger.Info("ConfigUI Start() called with port: " + port)
	cui.logger.Info("Starting ConfigUI on " + "127.0.0.1" + port)

	mux := http.NewServeMux()
	cui.SetupRoutes(mux)

	server := &http.Server{
		Addr:    "127.0.0.1" + port,
		Handler: mux,
	}

	cui.logger.Info("ConfigUI server configured, starting...")

	go func() {
		cui.logger.Info("ConfigUI server goroutine started, listening on " + server.Addr)
		cui.logger.Info("ConfigUI server starting ListenAndServe...")
		if err := server.ListenAndServe(); err != nil {
			cui.logger.Error("ConfigUI server error: " + err.Error())
		}
		cui.logger.Info("ConfigUI server goroutine finished")
	}()

	cui.logger.Info("ConfigUI Start() completed")
}
