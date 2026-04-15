package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"dardcor-agent/services"
)

type EgoHandler struct {
	egoService   *services.EgoService
	dreamService *services.DreamService
}

func NewEgoHandler(ego *services.EgoService, dream *services.DreamService) *EgoHandler {
	return &EgoHandler{egoService: ego, dreamService: dream}
}

func (h *EgoHandler) GetState(w http.ResponseWriter, r *http.Request) {
	state := h.egoService.GetState()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    state,
	})
}

func (h *EgoHandler) GetDreams(w http.ResponseWriter, r *http.Request) {
	count := 10
	if c := r.URL.Query().Get("count"); c != "" {
		if v, err := strconv.Atoi(c); err == nil && v > 0 {
			count = v
		}
	}
	dreams := h.dreamService.GetRecentDreams(count)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    dreams,
	})
}
