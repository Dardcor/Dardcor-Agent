package handlers

import (
	"encoding/json"
	"net/http"

	"dardcor-agent/services"
)

type BackgroundHandler struct {
	svc *services.BackgroundAgentService
}

func NewBackgroundHandler(svc *services.BackgroundAgentService) *BackgroundHandler {
	return &BackgroundHandler{svc: svc}
}

func (h *BackgroundHandler) Submit(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var req struct {
		Prompt string `json:"prompt"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Prompt == "" {
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "prompt required"})
		return
	}
	taskID := h.svc.Submit(req.Prompt)
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "task_id": taskID, "status": "pending"})
}

func (h *BackgroundHandler) GetTask(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	id := r.URL.Query().Get("id")
	if id == "" {
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "id required"})
		return
	}
	task := h.svc.GetTask(id)
	if task == nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "task not found"})
		return
	}
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "data": task})
}

func (h *BackgroundHandler) ListTasks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	tasks := h.svc.ListTasks()
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "data": tasks, "count": len(tasks)})
}

func (h *BackgroundHandler) CancelTask(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var req struct {
		ID string `json:"id"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	ok := h.svc.CancelTask(req.ID)
	json.NewEncoder(w).Encode(map[string]interface{}{"success": ok})
}

func (h *BackgroundHandler) PurgeCompleted(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	count := h.svc.PurgeCompleted()
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "purged": count})
}
