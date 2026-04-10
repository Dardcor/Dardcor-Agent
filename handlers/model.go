package handlers

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"dardcor-agent/models"
)

type ModelHandler struct {
	mu sync.Mutex
}

func NewModelHandler() *ModelHandler {

	os.MkdirAll(filepath.Join("database", "model"), 0755)
	os.MkdirAll(filepath.Join("database", "tools"), 0755)
	os.MkdirAll(filepath.Join("database", "skills"), 0755)
	os.MkdirAll(filepath.Join("database", "settings"), 0755)

	h := &ModelHandler{}

	h.ensureFileExists("tools", defaultTools)
	h.ensureFileExists("skills", defaultSkills)
	h.ensureFileExists("model", map[string]bool{"antigravity": false, "gemini": false, "openrouter": false})

	return h
}

func (h *ModelHandler) ensureFileExists(configType string, defaultData interface{}) {
	path := h.getPath(configType)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		data, _ := json.MarshalIndent(defaultData, "", "  ")
		os.WriteFile(path, data, 0644)
	}
}

var defaultTools = []map[string]interface{}{
	{"id": "fs", "name": "File System", "description": "Read, write, and manage local files and directories safely.", "icon": "📁", "status": "active", "category": "System"},
	{"id": "terminal", "name": "Terminal", "description": "Execute shell commands and manage system processes.", "icon": "💻", "status": "active", "category": "System"},
	{"id": "web", "name": "Web Browser", "description": "Search the web, browse websites, and extract information.", "icon": "🌐", "status": "active", "category": "Network"},
	{"id": "code", "name": "Code Interpreter", "description": "Execute Javascript/TypeScript code in a secure sandbox.", "icon": "📜", "status": "active", "category": "Logic"},
	{"id": "git", "name": "Git Manager", "description": "Manage repositories, commits, branches, and merges.", "icon": "🌿", "status": "inactive", "category": "Development"},
	{"id": "db", "name": "Database", "description": "Query and manipulate local or remote databases.", "icon": "🗄️", "status": "inactive", "category": "System"},
	{"id": "ai", "name": "Image Gen", "description": "Generate and edit images using AI models.", "icon": "🎨", "status": "active", "category": "Media"},
	{"id": "api", "name": "HTTP Client", "description": "Make API requests and handle various data formats.", "icon": "🔗", "status": "active", "category": "Network"},
}

var defaultSkills = []map[string]interface{}{
	{"id": "web-dev", "name": "Web Development", "description": "Expertise in modern frontend frameworks, backend systems, and responsive design.", "level": 95, "tags": []string{"React", "NodeJS", "CSS"}, "icon": "⚛️"},
	{"id": "data-sci", "name": "Data Engineering", "description": "Processing large datasets, building pipelines, and performing complex analysis.", "level": 82, "tags": []string{"Python", "SQL", "Pandas"}, "icon": "📊"},
	{"id": "devops", "name": "Cloud Architecture", "description": "Deploying applications, managing infrastructure, and CI/CD automation.", "level": 78, "tags": []string{"Docker", "AWS", "K8s"}, "icon": "☁️"},
	{"id": "sec", "name": "Cyber Security", "description": "Vulnerability assessment, secure coding practices, and threat mitigation.", "level": 65, "tags": []string{"PenTest", "Auth", "Encryption"}, "icon": "🛡️"},
	{"id": "ml", "name": "Machine Learning", "description": "Training models, fine-tuning LLMs, and implementing RAG systems.", "level": 88, "tags": []string{"LLM", "PyTorch", "VectorDB"}, "icon": "🧠"},
	{"id": "seo", "name": "SEO Strategy", "description": "Optimizing content for search engines and improving digital visibility.", "level": 92, "tags": []string{"Growth", "Analysis", "Content"}, "icon": "📈"},
}

func (h *ModelHandler) getPath(configType string) string {
	switch configType {
	case "tools":
		return filepath.Join("database", "tools", "config.json")
	case "skills":
		return filepath.Join("database", "skills", "config.json")
	case "workspace":
		return filepath.Join("database", "settings", "workspace.json")
	default:
		return filepath.Join("database", "model", "configure_active.json")
	}
}

func (h *ModelHandler) GetModelConfig(w http.ResponseWriter, r *http.Request) {
	h.getConfig(w, r, "model")
}

func (h *ModelHandler) GetToolsConfig(w http.ResponseWriter, r *http.Request) {
	h.getConfig(w, r, "tools")
}

func (h *ModelHandler) GetSkillsConfig(w http.ResponseWriter, r *http.Request) {
	h.getConfig(w, r, "skills")
}

func (h *ModelHandler) GetWorkspaceConfig(w http.ResponseWriter, r *http.Request) {
	h.getConfig(w, r, "workspace")
}

func (h *ModelHandler) getConfig(w http.ResponseWriter, r *http.Request, configType string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	path := h.getPath(configType)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			writeJSON(w, http.StatusOK, models.APIResponse{
				Success: true,
				Data:    nil,
			})
			return
		}
		writeJSON(w, http.StatusInternalServerError, models.APIResponse{Success: false, Error: err.Error()})
		return
	}

	var config interface{}
	if err := json.Unmarshal(data, &config); err != nil {
		writeJSON(w, http.StatusInternalServerError, models.APIResponse{Success: false, Error: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    config,
	})
}

func (h *ModelHandler) SaveModelConfig(w http.ResponseWriter, r *http.Request) {
	h.saveConfig(w, r, "model")
}

func (h *ModelHandler) SaveToolsConfig(w http.ResponseWriter, r *http.Request) {
	h.saveConfig(w, r, "tools")
}

func (h *ModelHandler) SaveSkillsConfig(w http.ResponseWriter, r *http.Request) {
	h.saveConfig(w, r, "skills")
}

func (h *ModelHandler) SaveWorkspaceConfig(w http.ResponseWriter, r *http.Request) {
	h.saveConfig(w, r, "workspace")
}

func (h *ModelHandler) saveConfig(w http.ResponseWriter, r *http.Request, configType string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	var config interface{}
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{Success: false, Error: "invalid request body"})
		return
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.APIResponse{Success: false, Error: err.Error()})
		return
	}

	path := h.getPath(configType)
	if err := os.WriteFile(path, data, 0644); err != nil {
		writeJSON(w, http.StatusInternalServerError, models.APIResponse{Success: false, Error: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Message: configType + " configuration saved successfully",
	})
}
