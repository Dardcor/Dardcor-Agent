package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

type MCPRequest struct {
	JSONRPC string                 `json:"jsonrpc"`
	Method  string                 `json:"method"`
	Params  map[string]interface{} `json:"params"`
	ID      interface{}            `json:"id"`
}

type MCPResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	Result  interface{} `json:"result,omitempty"`
	Error   *MCPError   `json:"error,omitempty"`
	ID      interface{} `json:"id"`
}

type MCPError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type MemMCPServer struct {
	searchSvc *MemorySearchService
}

func NewMemMCPServer(searchSvc *MemorySearchService) *MemMCPServer {
	return &MemMCPServer{searchSvc: searchSvc}
}

func (m *MemMCPServer) HandleMCP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req MCPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		resp := MCPResponse{
			JSONRPC: "2.0",
			Error:   &MCPError{Code: -32700, Message: "Parse error: " + err.Error()},
			ID:      nil,
		}
		json.NewEncoder(w).Encode(resp)
		return
	}

	var result interface{}
	var rpcErr *MCPError

	switch req.Method {
	case "initialize":
		result = map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities": map[string]interface{}{
				"tools": map[string]interface{}{},
			},
			"serverInfo": map[string]interface{}{
				"name":    "dardcor-agent",
				"version": "1.0.0",
			},
		}

	case "tools/list":
		result = m.handleListTools()

	case "tools/call":
		name, _ := req.Params["name"].(string)
		var args map[string]interface{}
		if a, ok := req.Params["arguments"]; ok {
			if aMap, ok := a.(map[string]interface{}); ok {
				args = aMap
			}
		}
		if args == nil {
			args = map[string]interface{}{}
		}
		toolResult, err := m.handleCallTool(name, args)
		if err != nil {
			rpcErr = &MCPError{Code: -32603, Message: err.Error()}
		} else {
			result = toolResult
		}

	default:
		rpcErr = &MCPError{Code: -32601, Message: "Method not found"}
	}

	resp := MCPResponse{
		JSONRPC: "2.0",
		Result:  result,
		Error:   rpcErr,
		ID:      req.ID,
	}
	json.NewEncoder(w).Encode(resp)
}

func (m *MemMCPServer) handleListTools() interface{} {
	return map[string]interface{}{
		"tools": []map[string]interface{}{
			{
				"name":        "search",
				"description": "Search Dardcor Agent memory. Returns observations matching the query. Step 1 of 3-layer retrieval.",
				"inputSchema": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"query": map[string]interface{}{
							"type":        "string",
							"description": "Search query",
						},
						"limit": map[string]interface{}{
							"type":        "number",
							"description": "Max results (default 20)",
						},
						"project": map[string]interface{}{
							"type":        "string",
							"description": "Filter by project",
						},
					},
				},
			},
			{
				"name":        "timeline",
				"description": "Get chronological context around an observation. Step 2 of 3-layer retrieval.",
				"inputSchema": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"anchor": map[string]interface{}{
							"type":        "number",
							"description": "Observation ID to center timeline around",
						},
						"depth_before": map[string]interface{}{
							"type":        "number",
							"description": "Items before anchor (default 3)",
						},
						"depth_after": map[string]interface{}{
							"type":        "number",
							"description": "Items after anchor (default 3)",
						},
						"project": map[string]interface{}{
							"type":        "string",
							"description": "Filter by project",
						},
					},
				},
			},
			{
				"name":        "get_observations",
				"description": "Fetch full observation details by IDs. Step 3 of 3-layer retrieval.",
				"inputSchema": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"ids": map[string]interface{}{
							"type":        "array",
							"description": "Observation IDs to fetch",
							"items": map[string]interface{}{
								"type": "number",
							},
						},
					},
					"required": []string{"ids"},
				},
			},
		},
	}
}

func (m *MemMCPServer) handleCallTool(name string, arguments map[string]interface{}) (interface{}, error) {
	switch name {
	case "search":
		return m.toolSearch(arguments)
	case "timeline":
		return m.toolTimeline(arguments)
	case "get_observations":
		return m.toolGetObservations(arguments)
	default:
		return nil, fmt.Errorf("unknown tool: %s", name)
	}
}

func (m *MemMCPServer) toolSearch(args map[string]interface{}) (interface{}, error) {
	query, _ := args["query"].(string)
	limit := 20
	if l, ok := args["limit"]; ok {
		limit = toInt(l)
	}
	project, _ := args["project"].(string)

	results, err := m.searchSvc.Search(query, limit, project)
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}

	resultJSON, err := json.Marshal(results)
	if err != nil {
		return nil, fmt.Errorf("marshal search results failed: %w", err)
	}

	return mcpTextContent(string(resultJSON)), nil
}

func (m *MemMCPServer) toolTimeline(args map[string]interface{}) (interface{}, error) {
	var anchorID int64
	if a, ok := args["anchor"]; ok {
		anchorID = int64(toInt(a))
	}
	depthBefore := 3
	if d, ok := args["depth_before"]; ok {
		depthBefore = toInt(d)
	}
	depthAfter := 3
	if d, ok := args["depth_after"]; ok {
		depthAfter = toInt(d)
	}
	project, _ := args["project"].(string)

	timeline, err := m.searchSvc.Timeline(anchorID, depthBefore, depthAfter, project)
	if err != nil {
		return nil, fmt.Errorf("timeline failed: %w", err)
	}

	resultJSON, err := json.Marshal(timeline)
	if err != nil {
		return nil, fmt.Errorf("marshal timeline failed: %w", err)
	}

	return mcpTextContent(string(resultJSON)), nil
}

func (m *MemMCPServer) toolGetObservations(args map[string]interface{}) (interface{}, error) {
	var ids []int64
	if rawIDs, ok := args["ids"]; ok {
		switch v := rawIDs.(type) {
		case []interface{}:
			for _, item := range v {
				ids = append(ids, int64(toInt(item)))
			}
		}
	}
	if len(ids) == 0 {
		return nil, fmt.Errorf("ids is required and must be a non-empty array")
	}

	obs, err := m.searchSvc.GetObservations(ids)
	if err != nil {
		return nil, fmt.Errorf("get observations failed: %w", err)
	}

	resultJSON, err := json.Marshal(obs)
	if err != nil {
		return nil, fmt.Errorf("marshal observations failed: %w", err)
	}

	return mcpTextContent(string(resultJSON)), nil
}

func mcpTextContent(text string) map[string]interface{} {
	return map[string]interface{}{
		"content": []map[string]interface{}{
			{"type": "text", "text": text},
		},
	}
}

func toInt(v interface{}) int {
	switch val := v.(type) {
	case float64:
		return int(val)
	case float32:
		return int(val)
	case int:
		return val
	case int64:
		return int(val)
	case int32:
		return int(val)
	case string:
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	return 0
}
