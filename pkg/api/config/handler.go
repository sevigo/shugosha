package config

import (
	"encoding/json"
	"net/http"

	"github.com/sevigo/shugosha/pkg/model"
)

type configHandler struct {
	configManager model.ConfigManager
}

func NewConfigHandler(configManger model.ConfigManager) *configHandler {
	return &configHandler{
		configManager: configManger,
	}
}

// readConfigHandler handles requests to read the configuration.
func (h *configHandler) ReadConfigHandler(w http.ResponseWriter, r *http.Request) {
	config, err := h.configManager.LoadConfig()
	if err != nil {
		http.Error(w, "Failed to read config: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(config)
}

// updateConfigHandler handles requests to update the configuration.
func (h *configHandler) UpdateConfigHandler(w http.ResponseWriter, r *http.Request) {
	var newConfig model.BackupConfig
	if err := json.NewDecoder(r.Body).Decode(&newConfig); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.configManager.SaveConfig(&newConfig); err != nil {
		http.Error(w, "Failed to update config: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
