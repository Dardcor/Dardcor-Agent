package handlers

import (
	"encoding/json"
	"net/http"

	"dardcor-agent/services"
)

type StatsHandler struct {
	costSvc *services.CostTrackerService
	idxSvc  *services.IndexService
	lspSvc  *services.LSPService
	mcpSvc  *services.MCPClientService
}

func NewStatsHandler(cost *services.CostTrackerService, idx *services.IndexService, lsp *services.LSPService, mcp *services.MCPClientService) *StatsHandler {
	return &StatsHandler{costSvc: cost, idxSvc: idx, lspSvc: lsp, mcpSvc: mcp}
}

func (h *StatsHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	stats := h.costSvc.GetStats()
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    stats,
		"summary": h.costSvc.FormatStats(),
	})
}

func (h *StatsHandler) ResetStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	h.costSvc.Reset()
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "message": "Stats reset"})
}

func (h *StatsHandler) GetIndexStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	idx := h.idxSvc.GetIndex()
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"status":  h.idxSvc.GetStatus(),
		"index":   idx,
	})
}

func (h *StatsHandler) BuildIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var req struct {
		Path string `json:"path"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	idx, err := h.idxSvc.Build(req.Path)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": err.Error()})
		return
	}
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":    true,
		"file_count": idx.FileCount,
		"root_path":  idx.RootPath,
		"built_at":   idx.BuiltAt,
	})
}

func (h *StatsHandler) SearchIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	q := r.URL.Query().Get("q")
	results := h.idxSvc.Search(q, 20)
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "data": results, "count": len(results)})
}

func (h *StatsHandler) GetLSPStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"servers": h.lspSvc.GetStatus(),
	})
}

func (h *StatsHandler) GetMCPServers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	servers := h.mcpSvc.ListServers()
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "data": servers, "count": len(servers)})
}

func (h *StatsHandler) AddMCPServer(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var req struct {
		Name    string            `json:"name"`
		Command string            `json:"command"`
		Args    []string          `json:"args"`
		Env     map[string]string `json:"env"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Name == "" || req.Command == "" {
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "name and command required"})
		return
	}
	h.mcpSvc.AddServer(req.Name, req.Command, req.Args, req.Env)
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "message": "MCP server added: " + req.Name})
}

func (h *StatsHandler) RemoveMCPServer(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	name := r.URL.Query().Get("name")
	ok := h.mcpSvc.RemoveServer(name)
	json.NewEncoder(w).Encode(map[string]interface{}{"success": ok})
}
