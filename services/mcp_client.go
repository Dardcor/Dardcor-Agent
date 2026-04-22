package services

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"dardcor-agent/config"
)

type MCPClientService struct {
	mu         sync.RWMutex
	servers    map[string]*MCPServer
	configPath string
}

type MCPServer struct {
	Name      string            `json:"name"`
	Command   string            `json:"command"`
	Args      []string          `json:"args"`
	Env       map[string]string `json:"env,omitempty"`
	Status    string            `json:"status"`
	LastError string            `json:"last_error,omitempty"`
	AddedAt   time.Time         `json:"added_at"`
	Available bool              `json:"available"`
}

type MCPConfig struct {
	Servers map[string]MCPServerConfig `json:"servers"`
}

type MCPServerConfig struct {
	Command string            `json:"command"`
	Args    []string          `json:"args,omitempty"`
	Env     map[string]string `json:"env,omitempty"`
}

func NewMCPClientService() *MCPClientService {
	homeDir, _ := os.UserHomeDir()
	configPath := filepath.Join(homeDir, ".dardcor", "mcp.json")
	if config.AppConfig != nil {
		configPath = filepath.Join(filepath.Dir(config.AppConfig.DataDir), "mcp.json")
	}

	svc := &MCPClientService{
		servers:    make(map[string]*MCPServer),
		configPath: configPath,
	}
	svc.loadConfig()
	return svc
}

func (m *MCPClientService) loadConfig() {
	data, err := os.ReadFile(m.configPath)
	if err != nil {
		return
	}
	var cfg MCPConfig
	if json.Unmarshal(data, &cfg) != nil {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	for name, s := range cfg.Servers {
		available := false
		if _, err := exec.LookPath(s.Command); err == nil {
			available = true
		}
		if s.Command == "npx" || s.Command == "node" {
			if _, err := exec.LookPath(s.Command); err == nil {
				available = true
			}
		}
		m.servers[name] = &MCPServer{
			Name:      name,
			Command:   s.Command,
			Args:      s.Args,
			Env:       s.Env,
			Status:    "idle",
			Available: available,
			AddedAt:   time.Now(),
		}
	}
}

func (m *MCPClientService) saveConfig() {
	m.mu.RLock()
	defer m.mu.RUnlock()

	cfg := MCPConfig{Servers: make(map[string]MCPServerConfig)}
	for name, srv := range m.servers {
		cfg.Servers[name] = MCPServerConfig{
			Command: srv.Command,
			Args:    srv.Args,
			Env:     srv.Env,
		}
	}

	os.MkdirAll(filepath.Dir(m.configPath), 0755)
	data, _ := json.MarshalIndent(cfg, "", "  ")
	os.WriteFile(m.configPath, data, 0644)
}

func (m *MCPClientService) AddServer(name, command string, args []string, env map[string]string) {
	available := false
	if _, err := exec.LookPath(command); err == nil {
		available = true
	}

	m.mu.Lock()
	m.servers[name] = &MCPServer{
		Name:      name,
		Command:   command,
		Args:      args,
		Env:       env,
		Status:    "idle",
		Available: available,
		AddedAt:   time.Now(),
	}
	m.mu.Unlock()
	m.saveConfig()
}

func (m *MCPClientService) RemoveServer(name string) bool {
	m.mu.Lock()
	_, existed := m.servers[name]
	delete(m.servers, name)
	m.mu.Unlock()
	if existed {
		m.saveConfig()
	}
	return existed
}

func (m *MCPClientService) ListServers() []*MCPServer {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var list []*MCPServer
	for _, srv := range m.servers {
		copy := *srv
		list = append(list, &copy)
	}
	return list
}

func (m *MCPClientService) GetServer(name string) *MCPServer {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if s, ok := m.servers[name]; ok {
		copy := *s
		return &copy
	}
	return nil
}

func (m *MCPClientService) FormatForAgentPrompt() string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(m.servers) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("\n\n[MCP SERVERS AVAILABLE]\n")
	for name, srv := range m.servers {
		status := "✓"
		if !srv.Available {
			status = "✗"
		}
		sb.WriteString(fmt.Sprintf("  %s %s: %s %s\n", status, name, srv.Command, strings.Join(srv.Args, " ")))
	}
	sb.WriteString("[END MCP SERVERS]\n")
	return sb.String()
}

func GetBuiltinMCPServers() []MCPServerConfig {
	return []MCPServerConfig{
		{Command: "npx", Args: []string{"-y", "@upstash/context7-mcp@latest"}},
		{Command: "npx", Args: []string{"-y", "@grep-app/mcp@latest"}},
		{Command: "npx", Args: []string{"-y", "@modelcontextprotocol/server-brave-search"}},
		{Command: "npx", Args: []string{"-y", "@modelcontextprotocol/server-filesystem", "."}},
		{Command: "npx", Args: []string{"-y", "@modelcontextprotocol/server-github"}},
		{Command: "npx", Args: []string{"-y", "@modelcontextprotocol/server-memory"}},
	}
}
