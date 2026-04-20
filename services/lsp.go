package services

import (
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

// LSPService provides Language Server Protocol integration for code intelligence
type LSPService struct {
	mu      sync.RWMutex
	servers map[string]*LSPServerInfo
	active  map[string]*lspProcess
}

type LSPServerInfo struct {
	Language  string   `json:"language"`
	Command   string   `json:"command"`
	Args      []string `json:"args"`
	Available bool     `json:"available"`
}

type lspProcess struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout io.ReadCloser
	id     int
}

type LSPDiagnostic struct {
	File     string `json:"file"`
	Line     int    `json:"line"`
	Col      int    `json:"col"`
	Severity string `json:"severity"`
	Message  string `json:"message"`
	Source   string `json:"source"`
}

type LSPLocation struct {
	Path  string `json:"path"`
	Range struct {
		Start struct{ Line, Character int } `json:"start"`
		End   struct{ Line, Character int } `json:"end"`
	} `json:"range"`
}

func NewLSPService() *LSPService {
	svc := &LSPService{
		servers: map[string]*LSPServerInfo{
			"typescript": {Language: "typescript", Command: "typescript-language-server", Args: []string{"--stdio"}},
			"javascript": {Language: "javascript", Command: "typescript-language-server", Args: []string{"--stdio"}},
			"go":         {Language: "go", Command: "gopls", Args: []string{}},
			"python":     {Language: "python", Command: "pylsp", Args: []string{}},
			"rust":       {Language: "rust", Command: "rust-analyzer", Args: []string{}},
			"java":       {Language: "java", Command: "jdtls", Args: []string{}},
		},
		active: make(map[string]*lspProcess),
	}
	svc.detectAvailable()
	return svc
}

func (l *LSPService) detectAvailable() {
	l.mu.Lock()
	defer l.mu.Unlock()
	for _, srv := range l.servers {
		_, err := exec.LookPath(srv.Command)
		srv.Available = err == nil
	}
}

func (l *LSPService) GetAvailableServers() []*LSPServerInfo {
	l.mu.RLock()
	defer l.mu.RUnlock()
	var available []*LSPServerInfo
	for _, srv := range l.servers {
		if srv.Available {
			copy := *srv
			available = append(available, &copy)
		}
	}
	return available
}

func (l *LSPService) GetServerForFile(filePath string) *LSPServerInfo {
	l.mu.RLock()
	defer l.mu.RUnlock()

	ext := strings.ToLower(filepath.Ext(filePath))
	if ext != "" {
		ext = ext[1:]
	}

	langMap := map[string]string{
		"ts": "typescript", "tsx": "typescript",
		"js": "javascript", "jsx": "javascript",
		"go": "go", "py": "python",
		"rs": "rust", "java": "java",
	}
	if lang, ok := langMap[ext]; ok {
		if srv, ok := l.servers[lang]; ok && srv.Available {
			return srv
		}
	}
	return nil
}

func (l *LSPService) GetDefinition(filePath string, line, col int) ([]LSPLocation, error) {
	srv := l.GetServerForFile(filePath)
	if srv == nil {
		return nil, fmt.Errorf("no LSP server available for %s", filePath)
	}
	// This is a stub for the full JSON-RPC implementation
	return nil, fmt.Errorf("LSP method textDocument/definition not yet fully implemented for %s", srv.Command)
}

func (l *LSPService) FindReferences(filePath string, line, col int) ([]LSPLocation, error) {
	return nil, fmt.Errorf("LSP method textDocument/references not yet implemented")
}

func (l *LSPService) GetSymbols(filePath string) ([]string, error) {
	return nil, fmt.Errorf("LSP method textDocument/documentSymbol not yet implemented")
}

func (l *LSPService) FormatDiagnosticsForAgent(diagnostics []LSPDiagnostic) string {
	if len(diagnostics) == 0 {
		return "No LSP diagnostics."
	}
	var sb strings.Builder
	for _, d := range diagnostics {
		icon := "ℹ️"
		switch d.Severity {
		case "error":
			icon = "❌"
		case "warning":
			icon = "⚠️"
		case "hint":
			icon = "💡"
		}
		sb.WriteString(fmt.Sprintf("%s %s:%d:%d — %s\n", icon, d.File, d.Line, d.Col, d.Message))
	}
	return sb.String()
}

func (l *LSPService) GetStatus() map[string]interface{} {
	l.mu.RLock()
	defer l.mu.RUnlock()
	status := map[string]interface{}{}
	for lang, srv := range l.servers {
		status[lang] = map[string]interface{}{
			"command":   srv.Command,
			"available": srv.Available,
		}
	}
	return status
}

func (l *LSPService) MarshalJSON() ([]byte, error) {
	return json.Marshal(l.GetStatus())
}
