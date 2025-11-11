package presentation

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/janmbaco/go-infrastructure/v2/configuration"
)

func HandleGetConfig(w http.ResponseWriter, r *http.Request, configHandler configuration.ConfigHandler) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	config := configHandler.GetConfig()
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(config)
}

func HandleUpdateConfig(w http.ResponseWriter, r *http.Request, configHandler configuration.ConfigHandler) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var newConfig interface{} // placeholder
	if err := json.NewDecoder(r.Body).Decode(&newConfig); err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
		return
	}
	if err := configHandler.SetConfig(&newConfig); err != nil {
		http.Error(w, fmt.Sprintf("Failed to update config: %v", err), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}
