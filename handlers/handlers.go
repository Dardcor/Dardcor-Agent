package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"dardcor-agent/models"
	"dardcor-agent/services"
)

type FileSystemHandler struct {
	service *services.FileSystemService
}

func NewFileSystemHandler(service *services.FileSystemService) *FileSystemHandler {
	return &FileSystemHandler{service: service}
}

func (h *FileSystemHandler) safeJoin(path string) (string, error) {
	workspace, err := h.service.GetDefaultWorkspace()
	if err != nil {
		return "", err
	}
	cleanPath := filepath.Clean(path)
	if filepath.IsAbs(cleanPath) {
		rel, err := filepath.Rel(workspace, cleanPath)
		if err != nil || strings.HasPrefix(rel, "..") {
			return "", fmt.Errorf("access denied: path outside workspace")
		}
		return cleanPath, nil
	}
	target := filepath.Join(workspace, cleanPath)
	rel, err := filepath.Rel(workspace, target)
	if err != nil || strings.HasPrefix(rel, "..") {
		return "", fmt.Errorf("access denied: path outside workspace")
	}
	return target, nil
}

func (h *FileSystemHandler) ListDirectory(w http.ResponseWriter, r *http.Request) {
	path, err := h.safeJoin(r.URL.Query().Get("path"))
	if err != nil {
		path = "." // Fallback to workspace root if empty/invalid
		if workspace, err2 := h.service.GetDefaultWorkspace(); err2 == nil {
			path = workspace
		}
	}
	if path == "" {
		path = "."
	}

	files, err := h.service.ListDirectory(path)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    files,
	})
}

func (h *FileSystemHandler) ReadFile(w http.ResponseWriter, r *http.Request) {
	path, err := h.safeJoin(r.URL.Query().Get("path"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}
	if path == "" {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   "path is required",
		})
		return
	}

	content, err := h.service.ReadFile(path)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    content,
	})
}

func (h *FileSystemHandler) WriteFile(w http.ResponseWriter, r *http.Request) {
	var req models.FileWriteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   "invalid request body",
		})
		return
	}

	path, err := h.safeJoin(req.Path)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	if err := h.service.WriteFile(path, req.Content); err != nil {
		writeJSON(w, http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Message: "File written successfully",
	})
}

func (h *FileSystemHandler) DeleteFile(w http.ResponseWriter, r *http.Request) {
	path, err := h.safeJoin(r.URL.Query().Get("path"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	if path == "" {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   "path is required",
		})
		return
	}

	if err := h.service.DeleteFile(path); err != nil {
		writeJSON(w, http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Deleted successfully",
	})
}

func (h *FileSystemHandler) SearchFiles(w http.ResponseWriter, r *http.Request) {
	var req models.SearchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   "invalid request body",
		})
		return
	}

	path, err := h.safeJoin(req.Path)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}
	req.Path = path
	results, err := h.service.SearchFiles(req)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    results,
	})
}

func (h *FileSystemHandler) GetFileInfo(w http.ResponseWriter, r *http.Request) {
	path, err := h.safeJoin(r.URL.Query().Get("path"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}
	if path == "" {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   "path is required",
		})
		return
	}

	info, err := h.service.GetFileInfo(path)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    info,
	})
}

func (h *FileSystemHandler) CreateDirectory(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Path string `json:"path"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   "invalid request body",
		})
		return
	}

	path, err := h.safeJoin(req.Path)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}
	if err := h.service.CreateDirectory(path); err != nil {
		writeJSON(w, http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Directory created successfully",
	})
}

func (h *FileSystemHandler) GetDrives(w http.ResponseWriter, r *http.Request) {
	drives := h.service.GetDrives()
	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    drives,
	})
}

func (h *FileSystemHandler) GetDefaultWorkspace(w http.ResponseWriter, r *http.Request) {
	path, err := h.service.GetDefaultWorkspace()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}
	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    path,
	})
}

func (h *FileSystemHandler) MoveFile(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Source      string `json:"source"`
		Destination string `json:"destination"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   "invalid request body",
		})
		return
	}

	src, err := h.safeJoin(req.Source)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{Success: false, Error: "source: " + err.Error()})
		return
	}
	dst, err := h.safeJoin(req.Destination)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{Success: false, Error: "destination: " + err.Error()})
		return
	}

	if err := h.service.MoveFile(src, dst); err != nil {
		writeJSON(w, http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Message: "File moved successfully",
	})
}

func (h *FileSystemHandler) CopyFile(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Source      string `json:"source"`
		Destination string `json:"destination"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   "invalid request body",
		})
		return
	}

	src, err := h.safeJoin(req.Source)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{Success: false, Error: "source: " + err.Error()})
		return
	}
	dst, err := h.safeJoin(req.Destination)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{Success: false, Error: "destination: " + err.Error()})
		return
	}

	if err := h.service.CopyFile(src, dst); err != nil {
		writeJSON(w, http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Message: "File copied successfully",
	})
}

type CommandHandler struct {
	service *services.CommandService
}

func NewCommandHandler(service *services.CommandService) *CommandHandler {
	return &CommandHandler{service: service}
}

func (h *CommandHandler) ExecuteCommand(w http.ResponseWriter, r *http.Request) {
	var req models.CommandRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   "invalid request body",
		})
		return
	}

	result, err := h.service.ExecuteCommand(req)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    result,
	})
}

func (h *CommandHandler) GetCommandHistory(w http.ResponseWriter, r *http.Request) {
	limit := 50
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil {
			limit = parsed
		}
	}

	history, err := h.service.GetCommandHistory(limit)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    history,
	})
}

func (h *CommandHandler) GetShellInfo(w http.ResponseWriter, r *http.Request) {
	info := h.service.GetShellInfo()
	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    info,
	})
}

type SystemHandler struct {
	service *services.SystemService
}

func NewSystemHandler(service *services.SystemService) *SystemHandler {
	return &SystemHandler{service: service}
}

func (h *SystemHandler) GetSystemInfo(w http.ResponseWriter, r *http.Request) {
	info, err := h.service.GetSystemInfo()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    info,
	})
}

func (h *SystemHandler) GetProcesses(w http.ResponseWriter, r *http.Request) {
	sortBy := r.URL.Query().Get("sort")
	if sortBy == "" {
		sortBy = "cpu"
	}

	limit := 50
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil {
			limit = parsed
		}
	}

	procs, err := h.service.GetProcesses(sortBy, limit)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    procs,
	})
}

func (h *SystemHandler) KillProcess(w http.ResponseWriter, r *http.Request) {
	var req struct {
		PID int32 `json:"pid"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   "invalid request body",
		})
		return
	}

	if err := h.service.KillProcess(req.PID); err != nil {
		writeJSON(w, http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Process killed successfully",
	})
}

func (h *SystemHandler) GetCPUUsage(w http.ResponseWriter, r *http.Request) {
	usage, err := h.service.GetCPUUsageRealtime()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    usage,
	})
}

func (h *SystemHandler) GetMemoryUsage(w http.ResponseWriter, r *http.Request) {
	mem, err := h.service.GetMemoryUsage()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    mem,
	})
}

type AgentHandler struct {
	service *services.AgentService
}

func NewAgentHandler(service *services.AgentService) *AgentHandler {
	return &AgentHandler{service: service}
}

func (h *AgentHandler) ProcessMessage(w http.ResponseWriter, r *http.Request) {
	var req models.AgentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   "invalid request body",
		})
		return
	}

	response, err := h.service.ProcessMessage(req, nil)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    response,
	})
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
