package models

import "time"

type AgentRequest struct {
	Message        string `json:"message"`
	ConversationID string `json:"conversation_id,omitempty"`
	Source         string `json:"source,omitempty"`
}

type AgentResponse struct {
	ID             string    `json:"id"`
	ConversationID string    `json:"conversation_id"`
	Role           string    `json:"role"`
	Content        string    `json:"content"`
	Actions        []Action  `json:"actions,omitempty"`
	Timestamp      time.Time `json:"timestamp"`
	Status         string    `json:"status"`
}

type Action struct {
	Type        string                 `json:"type"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters,omitempty"`
	Result      interface{}            `json:"result,omitempty"`
	Status      string                 `json:"status"`
	Error       string                 `json:"error,omitempty"`
	Duration    int64                  `json:"duration_ms,omitempty"`
}

type Conversation struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Messages  []Message `json:"messages"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Message struct {
	ID        string    `json:"id"`
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	Actions   []Action  `json:"actions,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

type FileInfo struct {
	Name       string    `json:"name"`
	Path       string    `json:"path"`
	Size       int64     `json:"size"`
	IsDir      bool      `json:"is_dir"`
	Extension  string    `json:"extension"`
	ModifiedAt time.Time `json:"modified_at"`
	Permission string    `json:"permission"`
}

type AntigravityConfig struct {
	Temperature        float64 `json:"temperature"`
	MaxTokens          int     `json:"max_tokens"`
	SelectedModel      string  `json:"selected_model"`
	ThinkingBudget     int     `json:"thinking_budget"`
	GoogleClientID     string  `json:"google_client_id"`
	GoogleClientSecret string  `json:"google_client_secret"`
}

type FileContent struct {
	Path     string `json:"path"`
	Content  string `json:"content"`
	Encoding string `json:"encoding"`
	Size     int64  `json:"size"`
}

type FileWriteRequest struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

type SearchRequest struct {
	Path          string `json:"path"`
	Query         string `json:"query"`
	FileType      string `json:"file_type,omitempty"`
	MaxDepth      int    `json:"max_depth,omitempty"`
	SearchContent bool   `json:"search_content,omitempty"`
}

type SearchResult struct {
	Path      string `json:"path"`
	Name      string `json:"name"`
	IsDir     bool   `json:"is_dir"`
	MatchLine int    `json:"match_line,omitempty"`
	MatchText string `json:"match_text,omitempty"`
}

type CommandRequest struct {
	Command    string `json:"command"`
	Shell      string `json:"shell,omitempty"`
	WorkingDir string `json:"working_dir,omitempty"`
	Timeout    int    `json:"timeout,omitempty"`
}

type CommandResponse struct {
	ID         string    `json:"id"`
	Command    string    `json:"command"`
	Output     string    `json:"output"`
	Error      string    `json:"error,omitempty"`
	ExitCode   int       `json:"exit_code"`
	Duration   int64     `json:"duration_ms"`
	StartedAt  time.Time `json:"started_at"`
	FinishedAt time.Time `json:"finished_at"`
}

type CommandHistory struct {
	Commands []CommandResponse `json:"commands"`
}

type SystemInfo struct {
	OS          OSInfo            `json:"os"`
	CPU         CPUInfo           `json:"cpu"`
	Memory      MemoryInfo        `json:"memory"`
	Disks       []DiskInfo        `json:"disks"`
	Network     []NetInfo         `json:"network"`
	Uptime      uint64            `json:"uptime"`
	HostName    string            `json:"hostname"`
	Bios        map[string]string `json:"bios,omitempty"`
	CollectedAt time.Time         `json:"collected_at"`
}

type OSInfo struct {
	Platform string `json:"platform"`
	Family   string `json:"family"`
	Version  string `json:"Version"`
	Arch     string `json:"arch"`
}

type CPUInfo struct {
	ModelName    string  `json:"model_name"`
	Cores        int     `json:"cores"`
	Threads      int     `json:"threads"`
	UsagePercent float64 `json:"usage_percent"`
	Frequency    float64 `json:"frequency_mhz"`
}

type MemoryInfo struct {
	Total       uint64  `json:"total"`
	Used        uint64  `json:"used"`
	Free        uint64  `json:"free"`
	UsedPercent float64 `json:"used_percent"`
}

type DiskInfo struct {
	Device      string  `json:"device"`
	MountPoint  string  `json:"mount_point"`
	FSType      string  `json:"fs_type"`
	Total       uint64  `json:"total"`
	Used        uint64  `json:"used"`
	Free        uint64  `json:"free"`
	UsedPercent float64 `json:"used_percent"`
}

type NetInfo struct {
	Name       string   `json:"name"`
	MacAddress string   `json:"mac_address"`
	Addresses  []string `json:"addresses"`
	BytesSent  uint64   `json:"bytes_sent"`
	BytesRecv  uint64   `json:"bytes_recv"`
}

type ProcessInfo struct {
	PID        int32   `json:"pid"`
	Name       string  `json:"name"`
	Status     string  `json:"status"`
	CPUPercent float64 `json:"cpu_percent"`
	MemPercent float32 `json:"mem_percent"`
	MemoryMB   float64 `json:"memory_mb"`
	Username   string  `json:"username"`
	CreateTime int64   `json:"create_time"`
	CmdLine    string  `json:"cmdline"`
}

type WSMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

type WSCommandOutput struct {
	CommandID string `json:"command_id"`
	Output    string `json:"output"`
	IsError   bool   `json:"is_error"`
	Done      bool   `json:"done"`
}

type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Message string      `json:"message,omitempty"`
}

type Settings struct {
	Theme          string `json:"theme"`
	DefaultShell   string `json:"default_shell"`
	MaxFileSize    int64  `json:"max_file_size"`
	CommandTimeout int    `json:"command_timeout"`
	APIKey         string `json:"api_key,omitempty"`
	AIProvider     string `json:"ai_provider"`
	AIModel        string `json:"ai_model"`
	ExecutionMode  string `json:"execution_mode"` // "auto", "yolo", "safe"
}
