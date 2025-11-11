package presentation

import (
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"strings"

	"github.com/janmbaco/go-infrastructure/v2/configuration"
	"github.com/janmbaco/go-reverseproxy-ssl/v3/internal/domain"
)

//go:embed static/css/* static/js/*
var staticFS embed.FS

//go:embed templates/layouts/*.html templates/pages/*.html
var templatesFS embed.FS

type ConfigUI struct {
	configHandler      configuration.ConfigHandler
	vhResolver         domain.VirtualHostResolver
	virtualHostService IVirtualHostService
	templates          *template.Template
	logger             domain.Logger
}

func NewConfigUI(configHandler configuration.ConfigHandler, vhResolver domain.VirtualHostResolver, virtualHostService IVirtualHostService, logger domain.Logger) *ConfigUI {
	// Load templates with error handling
	templates, err := template.ParseFS(templatesFS, "templates/layouts/*.html", "templates/pages/*.html")
	if err != nil {
		logger.Error("Failed to load templates: " + err.Error())
		// Don't panic - return nil and let the caller handle it
		return nil
	}

	logger.Info("Templates loaded successfully")

	return &ConfigUI{
		configHandler:      configHandler,
		vhResolver:         vhResolver,
		virtualHostService: virtualHostService,
		templates:          templates,
		logger:             logger,
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
	vhCollection, err := cui.virtualHostService.GetVirtualHosts(config)
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
		Title              string
		ActivePage         string
		Template           string
		IsEdit             bool
		VirtualHost        interface{}
		VirtualHostType    string
		WebVirtualHost     *domain.WebVirtualHost
		GrpcWebVirtualHost *domain.GrpcWebVirtualHost
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
		VirtualHostType: "web",
		WebVirtualHost: &domain.WebVirtualHost{
			ClientCertificateHost: domain.ClientCertificateHost{
				VirtualHostBase: domain.VirtualHostBase{
					Scheme: "http",
					Port:   8080,
				},
			},
		},
		GrpcWebVirtualHost: nil,
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

	// Find the virtual host in WebVirtualHosts
	var foundVH interface{}
	var virtualHostType string
	var webVH *domain.WebVirtualHost
	var grpcWebVH *domain.GrpcWebVirtualHost

	for _, vh := range config.WebVirtualHosts {
		if vh.GetID() == id {
			// Create a copy of the virtual host
			foundVH = &domain.WebVirtualHost{
				ClientCertificateHost: domain.ClientCertificateHost{
					VirtualHostBase: vh.VirtualHostBase,
				},
			}
			virtualHostType = "web"
			webVH = foundVH.(*domain.WebVirtualHost)
			break
		}
	}

	// If not found in WebVirtualHosts, search in GrpcWebVirtualHosts
	if foundVH == nil {
		for _, vh := range config.GrpcWebVirtualHosts {
			if vh.GetID() == id {
				// Create a copy of the virtual host
				foundVH = &domain.GrpcWebVirtualHost{
					ClientCertificateHost: domain.ClientCertificateHost{
						VirtualHostBase: vh.VirtualHostBase,
					},
					GrpcWebProxy: vh.GrpcWebProxy,
				}
				virtualHostType = "grpc-web"
				grpcWebVH = foundVH.(*domain.GrpcWebVirtualHost)
				break
			}
		}
	}

	if foundVH == nil {
		http.Error(w, "Virtual host not found", http.StatusNotFound)
		return
	}

	data := struct {
		Title              string
		ActivePage         string
		Template           string
		IsEdit             bool
		VirtualHost        interface{}
		VirtualHostType    string
		WebVirtualHost     *domain.WebVirtualHost
		GrpcWebVirtualHost *domain.GrpcWebVirtualHost
	}{
		Title:              "Edit Virtual Host - Reverse Proxy Config",
		ActivePage:         "virtualhosts",
		Template:           "virtualhost-form-content",
		IsEdit:             true,
		VirtualHost:        foundVH,
		VirtualHostType:    virtualHostType,
		WebVirtualHost:     webVH,
		GrpcWebVirtualHost: grpcWebVH,
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
	vhCollection, err := cui.virtualHostService.GetVirtualHosts(config)
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
	config := cui.configHandler.GetConfig().(*domain.Config)

	newVH, err := cui.virtualHostService.CreateVirtualHost(r, config)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create virtual host: %v", err), http.StatusInternalServerError)
		return
	}

	// Add to appropriate array based on type
	switch vh := newVH.(type) {
	case *domain.WebVirtualHost:
		config.WebVirtualHosts = append(config.WebVirtualHosts, vh)
	case *domain.GrpcWebVirtualHost:
		config.GrpcWebVirtualHosts = append(config.GrpcWebVirtualHosts, vh)
	}

	if err := cui.configHandler.SetConfig(config); err != nil {
		cui.logger.Error(fmt.Sprintf("Failed to save config: %v", err))
		http.Error(w, fmt.Sprintf("Failed to save config: %v", err), http.StatusInternalServerError)
		return
	}

	cui.logger.Info("Virtual host created successfully")

	// Get ID using type assertion
	var id string
	if vh, ok := newVH.(*domain.WebVirtualHost); ok {
		id = vh.GetID()
	} else if vh, ok := newVH.(*domain.GrpcWebVirtualHost); ok {
		id = vh.GetID()
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Virtual host created successfully",
		"id":      id,
	})
}

func (cui *ConfigUI) getVirtualHost(w http.ResponseWriter, r *http.Request, id string) {
	config := cui.configHandler.GetConfig().(*domain.Config)

	foundVH, vhType, err := cui.virtualHostService.GetVirtualHost(id, config)
	if err != nil {
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
	config := cui.configHandler.GetConfig().(*domain.Config)

	newVH, oldServerPaths, oldClientPaths, err := cui.virtualHostService.UpdateVirtualHost(r, id, config)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to update virtual host: %v", err), http.StatusInternalServerError)
		return
	}

	// Remove old virtual host from both arrays
	for i, vh := range config.WebVirtualHosts {
		if vh.GetID() == id {
			config.WebVirtualHosts = append(config.WebVirtualHosts[:i], config.WebVirtualHosts[i+1:]...)
			break
		}
	}
	for i, vh := range config.GrpcWebVirtualHosts {
		if vh.GetID() == id {
			config.GrpcWebVirtualHosts = append(config.GrpcWebVirtualHosts[:i], config.GrpcWebVirtualHosts[i+1:]...)
			break
		}
	}

	// Add new virtual host to appropriate array
	switch vh := newVH.(type) {
	case *domain.WebVirtualHost:
		config.WebVirtualHosts = append(config.WebVirtualHosts, vh)
	case *domain.GrpcWebVirtualHost:
		config.GrpcWebVirtualHosts = append(config.GrpcWebVirtualHosts, vh)
	}

	if err := cui.configHandler.SetConfig(config); err != nil {
		http.Error(w, fmt.Sprintf("Failed to save config: %v", err), http.StatusInternalServerError)
		return
	}

	// Clean up old certificates that are no longer used
	cui.virtualHostService.CleanupUnusedCertificates(config, oldServerPaths, oldClientPaths)

	cui.logger.Info("Virtual host updated successfully with new ID")

	// Get new ID using type assertion
	var newId string
	if vh, ok := newVH.(*domain.WebVirtualHost); ok {
		newId = vh.GetID()
	} else if vh, ok := newVH.(*domain.GrpcWebVirtualHost); ok {
		newId = vh.GetID()
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"newId":   newId,
	})
}

func (cui *ConfigUI) deleteVirtualHost(w http.ResponseWriter, r *http.Request, id string) {
	config := cui.configHandler.GetConfig().(*domain.Config)

	deletedFrom, err := cui.virtualHostService.DeleteVirtualHost(id, config)
	if err != nil {
		cui.logger.Error("Virtual host not found with ID: " + id)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Virtual host not found",
		})
		return
	}

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

func (cui *ConfigUI) Start(port string) error {
	cui.logger.Info("Starting ConfigUI server on port " + port)

	mux := http.NewServeMux()
	cui.SetupRoutes(mux)

	server := &http.Server{
		Addr:    port,
		Handler: mux,
	}

	cui.logger.Info("ConfigUI server started successfully")
	return server.ListenAndServe()
}
