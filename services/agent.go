package services

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"dardcor-agent/models"
	"dardcor-agent/storage"

	"github.com/google/uuid"
)

type AgentService struct {
	fsService  *FileSystemService
	cmdService *CommandService
	sysService *SystemService
}

func NewAgentService(fs *FileSystemService, cmd *CommandService, sys *SystemService) *AgentService {
	return &AgentService{
		fsService:  fs,
		cmdService: cmd,
		sysService: sys,
	}
}

// ProcessMessage processes a user message and returns an agent response
func (as *AgentService) ProcessMessage(req models.AgentRequest) (*models.AgentResponse, error) {
	// Create or load conversation
	var convID string
	if req.ConversationID != "" {
		convID = req.ConversationID
	} else {
		conv, err := storage.Store.CreateConversation(as.generateTitle(req.Message))
		if err != nil {
			return nil, err
		}
		convID = conv.ID
	}

	// Save user message
	userMsg := models.Message{
		Role:    "user",
		Content: req.Message,
	}
	storage.Store.AddMessage(convID, userMsg)

	// Process the message and execute actions
	actions, responseText := as.interpretAndExecute(req.Message)

	// Create response
	response := &models.AgentResponse{
		ID:             uuid.New().String(),
		ConversationID: convID,
		Role:           "assistant",
		Content:        responseText,
		Actions:        actions,
		Timestamp:      time.Now(),
		Status:         "completed",
	}

	// Save assistant message
	assistantMsg := models.Message{
		Role:    "assistant",
		Content: responseText,
		Actions: actions,
	}
	storage.Store.AddMessage(convID, assistantMsg)

	return response, nil
}

// interpretAndExecute parses the user message and executes appropriate actions
func (as *AgentService) interpretAndExecute(message string) ([]models.Action, string) {
	msg := strings.ToLower(strings.TrimSpace(message))
	var actions []models.Action
	var responseText string

	switch {
	// File system operations
	case strings.HasPrefix(msg, "list ") || strings.HasPrefix(msg, "ls ") || strings.HasPrefix(msg, "dir "):
		actions, responseText = as.handleListDir(message)

	case strings.HasPrefix(msg, "read ") || strings.HasPrefix(msg, "cat ") || strings.HasPrefix(msg, "baca "):
		actions, responseText = as.handleReadFile(message)

	case strings.HasPrefix(msg, "write ") || strings.HasPrefix(msg, "tulis "):
		actions, responseText = as.handleWriteFile(message)

	case strings.HasPrefix(msg, "delete ") || strings.HasPrefix(msg, "hapus ") || strings.HasPrefix(msg, "rm "):
		actions, responseText = as.handleDeleteFile(message)

	case strings.HasPrefix(msg, "search ") || strings.HasPrefix(msg, "cari ") || strings.HasPrefix(msg, "find "):
		actions, responseText = as.handleSearch(message)

	case strings.HasPrefix(msg, "mkdir ") || strings.HasPrefix(msg, "buat folder "):
		actions, responseText = as.handleMkdir(message)

	case strings.HasPrefix(msg, "drives") || strings.HasPrefix(msg, "disk"):
		actions, responseText = as.handleDrives()

	// Command execution
	case strings.HasPrefix(msg, "run ") || strings.HasPrefix(msg, "exec ") || strings.HasPrefix(msg, "jalankan "):
		actions, responseText = as.handleRunCommand(message)

	case strings.HasPrefix(msg, "cmd ") || strings.HasPrefix(msg, "$"):
		actions, responseText = as.handleDirectCommand(message)

	// System information
	case msg == "sysinfo" || msg == "system info" || msg == "info sistem" || msg == "system":
		actions, responseText = as.handleSystemInfo()

	case msg == "processes" || msg == "proses" || msg == "ps" || strings.HasPrefix(msg, "top"):
		actions, responseText = as.handleProcesses(message)

	case strings.HasPrefix(msg, "kill ") || strings.HasPrefix(msg, "matikan "):
		actions, responseText = as.handleKillProcess(message)

	case msg == "cpu":
		actions, responseText = as.handleCPUInfo()

	case msg == "memory" || msg == "ram" || msg == "mem":
		actions, responseText = as.handleMemoryInfo()

	// Help
	case msg == "help" || msg == "bantuan" || msg == "?":
		responseText = as.getHelpText()

	// General info
	case msg == "whoami" || msg == "siapa":
		responseText = as.getAgentInfo()

	// File info
	case strings.HasPrefix(msg, "info "):
		actions, responseText = as.handleFileInfo(message)

	default:
		// Try to interpret as a command
		if strings.Contains(msg, "file") || strings.Contains(msg, "folder") || strings.Contains(msg, "directory") {
			responseText = "🤖 Saya mengerti Anda ingin bekerja dengan file/folder. Gunakan perintah seperti:\n" +
				"• `list <path>` - Melihat isi direktori\n" +
				"• `read <path>` - Membaca file\n" +
				"• `search <query>` - Mencari file\n" +
				"• `info <path>` - Info file/folder\n\n" +
				"Ketik `help` untuk daftar perintah lengkap."
		} else {
			responseText = fmt.Sprintf("🤖 **Dardcor Agent v1.0**\n\nSaya menerima pesan Anda: \"%s\"\n\n"+
				"Saya adalah AI Agent yang bisa mengakses seluruh komputer Anda. "+
				"Berikut yang bisa saya lakukan:\n\n"+
				"📁 **File System** - Browse, baca, tulis, hapus, cari file\n"+
				"💻 **Terminal** - Jalankan perintah sistem\n"+
				"📊 **System Monitor** - Info CPU, RAM, Disk, Network\n"+
				"⚙️ **Process Manager** - Lihat dan kelola proses\n\n"+
				"Ketik `help` untuk panduan lengkap.", message)
		}
	}

	return actions, responseText
}

func (as *AgentService) handleListDir(message string) ([]models.Action, string) {
	path := as.extractPath(message, []string{"list ", "ls ", "dir "})
	if path == "" {
		path = "."
	}

	action := models.Action{
		Type:        "list_directory",
		Description: fmt.Sprintf("Listing directory: %s", path),
		Parameters:  map[string]interface{}{"path": path},
		Status:      "running",
	}

	start := time.Now()
	files, err := as.fsService.ListDirectory(path)
	action.Duration = time.Since(start).Milliseconds()

	if err != nil {
		action.Status = "error"
		action.Error = err.Error()
		return []models.Action{action}, fmt.Sprintf("❌ Gagal membaca direktori: %s", err.Error())
	}

	action.Status = "completed"
	action.Result = files

	// Format response
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("📂 **Isi direktori:** `%s`\n\n", path))
	sb.WriteString(fmt.Sprintf("Total: %d item\n\n", len(files)))

	dirCount, fileCount := 0, 0
	for _, f := range files {
		if f.IsDir {
			dirCount++
			sb.WriteString(fmt.Sprintf("📁 `%s/`\n", f.Name))
		} else {
			fileCount++
			size := formatSize(f.Size)
			sb.WriteString(fmt.Sprintf("📄 `%s` (%s)\n", f.Name, size))
		}
	}

	sb.WriteString(fmt.Sprintf("\n📊 %d folder, %d file", dirCount, fileCount))
	return []models.Action{action}, sb.String()
}

func (as *AgentService) handleReadFile(message string) ([]models.Action, string) {
	path := as.extractPath(message, []string{"read ", "cat ", "baca "})

	action := models.Action{
		Type:        "read_file",
		Description: fmt.Sprintf("Reading file: %s", path),
		Parameters:  map[string]interface{}{"path": path},
		Status:      "running",
	}

	start := time.Now()
	content, err := as.fsService.ReadFile(path)
	action.Duration = time.Since(start).Milliseconds()

	if err != nil {
		action.Status = "error"
		action.Error = err.Error()
		return []models.Action{action}, fmt.Sprintf("❌ Gagal membaca file: %s", err.Error())
	}

	action.Status = "completed"

	// Truncate long content for display
	displayContent := content.Content
	if len(displayContent) > 5000 {
		displayContent = displayContent[:5000] + "\n\n... (truncated)"
	}

	responseText := fmt.Sprintf("📄 **File:** `%s`\n📏 **Size:** %s\n\n```\n%s\n```",
		content.Path, formatSize(content.Size), displayContent)

	return []models.Action{action}, responseText
}

func (as *AgentService) handleWriteFile(message string) ([]models.Action, string) {
	// Parse "write <path> <content>" format
	parts := strings.SplitN(message, " ", 3)
	if len(parts) < 3 {
		return nil, "❌ Format: `write <path> <content>`"
	}

	path := parts[1]
	content := parts[2]

	action := models.Action{
		Type:        "write_file",
		Description: fmt.Sprintf("Writing to file: %s", path),
		Parameters:  map[string]interface{}{"path": path},
		Status:      "running",
	}

	start := time.Now()
	err := as.fsService.WriteFile(path, content)
	action.Duration = time.Since(start).Milliseconds()

	if err != nil {
		action.Status = "error"
		action.Error = err.Error()
		return []models.Action{action}, fmt.Sprintf("❌ Gagal menulis file: %s", err.Error())
	}

	action.Status = "completed"
	return []models.Action{action}, fmt.Sprintf("✅ File berhasil ditulis: `%s` (%d bytes)", path, len(content))
}

func (as *AgentService) handleDeleteFile(message string) ([]models.Action, string) {
	path := as.extractPath(message, []string{"delete ", "hapus ", "rm "})

	action := models.Action{
		Type:        "delete_file",
		Description: fmt.Sprintf("Deleting: %s", path),
		Parameters:  map[string]interface{}{"path": path},
		Status:      "running",
	}

	start := time.Now()
	err := as.fsService.DeleteFile(path)
	action.Duration = time.Since(start).Milliseconds()

	if err != nil {
		action.Status = "error"
		action.Error = err.Error()
		return []models.Action{action}, fmt.Sprintf("❌ Gagal menghapus: %s", err.Error())
	}

	action.Status = "completed"
	return []models.Action{action}, fmt.Sprintf("✅ Berhasil dihapus: `%s`", path)
}

func (as *AgentService) handleSearch(message string) ([]models.Action, string) {
	query := as.extractPath(message, []string{"search ", "cari ", "find "})

	action := models.Action{
		Type:        "search_files",
		Description: fmt.Sprintf("Searching for: %s", query),
		Parameters:  map[string]interface{}{"query": query},
		Status:      "running",
	}

	start := time.Now()
	results, err := as.fsService.SearchFiles(models.SearchRequest{
		Path:     ".",
		Query:    query,
		MaxDepth: 5,
	})
	action.Duration = time.Since(start).Milliseconds()

	if err != nil {
		action.Status = "error"
		action.Error = err.Error()
		return []models.Action{action}, fmt.Sprintf("❌ Error pencarian: %s", err.Error())
	}

	action.Status = "completed"
	action.Result = results

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("🔍 **Hasil pencarian:** \"%s\"\n\n", query))
	sb.WriteString(fmt.Sprintf("Ditemukan %d hasil:\n\n", len(results)))

	for _, r := range results {
		icon := "📄"
		if r.IsDir {
			icon = "📁"
		}
		sb.WriteString(fmt.Sprintf("%s `%s`\n", icon, r.Path))
		if r.MatchText != "" {
			sb.WriteString(fmt.Sprintf("   Line %d: %s\n", r.MatchLine, r.MatchText))
		}
	}

	return []models.Action{action}, sb.String()
}

func (as *AgentService) handleMkdir(message string) ([]models.Action, string) {
	path := as.extractPath(message, []string{"mkdir ", "buat folder "})

	action := models.Action{
		Type:        "create_directory",
		Description: fmt.Sprintf("Creating directory: %s", path),
		Parameters:  map[string]interface{}{"path": path},
		Status:      "running",
	}

	start := time.Now()
	err := as.fsService.CreateDirectory(path)
	action.Duration = time.Since(start).Milliseconds()

	if err != nil {
		action.Status = "error"
		action.Error = err.Error()
		return []models.Action{action}, fmt.Sprintf("❌ Gagal membuat folder: %s", err.Error())
	}

	action.Status = "completed"
	return []models.Action{action}, fmt.Sprintf("✅ Folder berhasil dibuat: `%s`", path)
}

func (as *AgentService) handleDrives() ([]models.Action, string) {
	drives := as.fsService.GetDrives()

	action := models.Action{
		Type:        "list_drives",
		Description: "Listing available drives",
		Status:      "completed",
		Result:      drives,
	}

	var sb strings.Builder
	sb.WriteString("💽 **Drive yang tersedia:**\n\n")
	for _, d := range drives {
		sb.WriteString(fmt.Sprintf("💿 `%s`\n", d))
	}

	return []models.Action{action}, sb.String()
}

func (as *AgentService) handleRunCommand(message string) ([]models.Action, string) {
	cmd := as.extractPath(message, []string{"run ", "exec ", "jalankan "})

	action := models.Action{
		Type:        "execute_command",
		Description: fmt.Sprintf("Executing: %s", cmd),
		Parameters:  map[string]interface{}{"command": cmd},
		Status:      "running",
	}

	start := time.Now()
	result, err := as.cmdService.ExecuteCommand(models.CommandRequest{
		Command: cmd,
		Timeout: 30,
	})
	action.Duration = time.Since(start).Milliseconds()

	if err != nil {
		action.Status = "error"
		action.Error = err.Error()
		return []models.Action{action}, fmt.Sprintf("❌ Gagal menjalankan perintah: %s", err.Error())
	}

	action.Status = "completed"
	action.Result = result

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("💻 **Command:** `%s`\n", cmd))
	sb.WriteString(fmt.Sprintf("⏱️ Duration: %dms | Exit Code: %d\n\n", result.Duration, result.ExitCode))

	if result.Output != "" {
		output := result.Output
		if len(output) > 5000 {
			output = output[:5000] + "\n... (truncated)"
		}
		sb.WriteString(fmt.Sprintf("```\n%s\n```\n", output))
	}

	if result.Error != "" {
		sb.WriteString(fmt.Sprintf("\n⚠️ **Error:**\n```\n%s\n```", result.Error))
	}

	return []models.Action{action}, sb.String()
}

func (as *AgentService) handleDirectCommand(message string) ([]models.Action, string) {
	cmd := strings.TrimSpace(message)
	if strings.HasPrefix(cmd, "cmd ") {
		cmd = strings.TrimPrefix(cmd, "cmd ")
	} else if strings.HasPrefix(cmd, "$") {
		cmd = strings.TrimPrefix(cmd, "$")
	}
	cmd = strings.TrimSpace(cmd)

	return as.handleRunCommand("run " + cmd)
}

func (as *AgentService) handleSystemInfo() ([]models.Action, string) {
	action := models.Action{
		Type:        "system_info",
		Description: "Getting system information",
		Status:      "running",
	}

	start := time.Now()
	info, err := as.sysService.GetSystemInfo()
	action.Duration = time.Since(start).Milliseconds()

	if err != nil {
		action.Status = "error"
		action.Error = err.Error()
		return []models.Action{action}, fmt.Sprintf("❌ Gagal mendapatkan info sistem: %s", err.Error())
	}

	action.Status = "completed"
	action.Result = info

	var sb strings.Builder
	sb.WriteString("🖥️ **System Information**\n\n")
	sb.WriteString(fmt.Sprintf("**Hostname:** %s\n", info.HostName))
	sb.WriteString(fmt.Sprintf("**OS:** %s %s (%s)\n", info.OS.Platform, info.OS.Version, info.OS.Arch))
	sb.WriteString(fmt.Sprintf("**Uptime:** %s\n\n", formatDuration(info.Uptime)))

	sb.WriteString(fmt.Sprintf("**CPU:** %s\n", info.CPU.ModelName))
	sb.WriteString(fmt.Sprintf("  Cores: %d | Threads: %d | Usage: %.1f%%\n\n", info.CPU.Cores, info.CPU.Threads, info.CPU.UsagePercent))

	sb.WriteString(fmt.Sprintf("**Memory:** %s / %s (%.1f%% used)\n\n",
		formatSize(int64(info.Memory.Used)), formatSize(int64(info.Memory.Total)), info.Memory.UsedPercent))

	sb.WriteString("**Disks:**\n")
	for _, d := range info.Disks {
		sb.WriteString(fmt.Sprintf("  💽 %s (%s) - %s / %s (%.1f%%)\n",
			d.MountPoint, d.FSType,
			formatSize(int64(d.Used)), formatSize(int64(d.Total)), d.UsedPercent))
	}

	return []models.Action{action}, sb.String()
}

func (as *AgentService) handleProcesses(message string) ([]models.Action, string) {
	sortBy := "cpu"
	limit := 20

	if strings.Contains(message, "mem") || strings.Contains(message, "ram") {
		sortBy = "memory"
	} else if strings.Contains(message, "name") || strings.Contains(message, "nama") {
		sortBy = "name"
	}

	action := models.Action{
		Type:        "list_processes",
		Description: "Listing processes",
		Status:      "running",
	}

	start := time.Now()
	procs, err := as.sysService.GetProcesses(sortBy, limit)
	action.Duration = time.Since(start).Milliseconds()

	if err != nil {
		action.Status = "error"
		action.Error = err.Error()
		return []models.Action{action}, fmt.Sprintf("❌ Error: %s", err.Error())
	}

	action.Status = "completed"
	action.Result = procs

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("⚙️ **Top %d Processes** (sorted by %s)\n\n", limit, sortBy))
	sb.WriteString("| PID | Name | CPU%% | Memory | User |\n")
	sb.WriteString("|-----|------|------|--------|------|\n")

	for _, p := range procs {
		sb.WriteString(fmt.Sprintf("| %d | %s | %.1f%% | %.1f MB | %s |\n",
			p.PID, truncate(p.Name, 20), p.CPUPercent, p.MemoryMB, truncate(p.Username, 15)))
	}

	return []models.Action{action}, sb.String()
}

func (as *AgentService) handleKillProcess(message string) ([]models.Action, string) {
	pidStr := as.extractPath(message, []string{"kill ", "matikan "})
	var pid int32
	fmt.Sscanf(pidStr, "%d", &pid)

	if pid == 0 {
		return nil, "❌ PID tidak valid. Format: `kill <PID>`"
	}

	action := models.Action{
		Type:        "kill_process",
		Description: fmt.Sprintf("Killing process PID: %d", pid),
		Parameters:  map[string]interface{}{"pid": pid},
		Status:      "running",
	}

	start := time.Now()
	err := as.sysService.KillProcess(pid)
	action.Duration = time.Since(start).Milliseconds()

	if err != nil {
		action.Status = "error"
		action.Error = err.Error()
		return []models.Action{action}, fmt.Sprintf("❌ Gagal mematikan proses: %s", err.Error())
	}

	action.Status = "completed"
	return []models.Action{action}, fmt.Sprintf("✅ Proses %d berhasil dimatikan", pid)
}

func (as *AgentService) handleCPUInfo() ([]models.Action, string) {
	info, err := as.sysService.GetSystemInfo()
	if err != nil {
		return nil, fmt.Sprintf("❌ Error: %s", err.Error())
	}

	action := models.Action{
		Type:   "cpu_info",
		Status: "completed",
	}

	responseText := fmt.Sprintf("🔧 **CPU Info**\n\n"+
		"**Model:** %s\n"+
		"**Cores:** %d\n"+
		"**Threads:** %d\n"+
		"**Usage:** %.1f%%\n"+
		"**Frequency:** %.0f MHz",
		info.CPU.ModelName, info.CPU.Cores, info.CPU.Threads,
		info.CPU.UsagePercent, info.CPU.Frequency)

	return []models.Action{action}, responseText
}

func (as *AgentService) handleMemoryInfo() ([]models.Action, string) {
	mem, err := as.sysService.GetMemoryUsage()
	if err != nil {
		return nil, fmt.Sprintf("❌ Error: %s", err.Error())
	}

	action := models.Action{
		Type:   "memory_info",
		Status: "completed",
	}

	responseText := fmt.Sprintf("🧠 **Memory Info**\n\n"+
		"**Total:** %s\n"+
		"**Used:** %s\n"+
		"**Free:** %s\n"+
		"**Usage:** %.1f%%",
		formatSize(int64(mem.Total)), formatSize(int64(mem.Used)),
		formatSize(int64(mem.Free)), mem.UsedPercent)

	return []models.Action{action}, responseText
}

func (as *AgentService) handleFileInfo(message string) ([]models.Action, string) {
	path := as.extractPath(message, []string{"info "})

	info, err := as.fsService.GetFileInfo(path)
	if err != nil {
		return nil, fmt.Sprintf("❌ Error: %s", err.Error())
	}

	action := models.Action{
		Type:   "file_info",
		Status: "completed",
		Result: info,
	}

	icon := "📄"
	if info.IsDir {
		icon = "📁"
	}

	responseText := fmt.Sprintf("%s **File Info**\n\n"+
		"**Name:** %s\n"+
		"**Path:** %s\n"+
		"**Size:** %s\n"+
		"**Type:** %s\n"+
		"**Modified:** %s\n"+
		"**Permissions:** %s",
		icon, info.Name, info.Path, formatSize(info.Size),
		info.Extension, info.ModifiedAt.Format("2006-01-02 15:04:05"),
		info.Permission)

	return []models.Action{action}, responseText
}

func (as *AgentService) getHelpText() string {
	return `🤖 **Dardcor Agent - Panduan Perintah**

📁 **File System:**
• ` + "`list <path>`" + ` - Lihat isi direktori
• ` + "`read <path>`" + ` - Baca isi file
• ` + "`write <path> <content>`" + ` - Tulis ke file
• ` + "`delete <path>`" + ` - Hapus file/folder
• ` + "`search <query>`" + ` - Cari file
• ` + "`mkdir <path>`" + ` - Buat folder
• ` + "`info <path>`" + ` - Info detail file
• ` + "`drives`" + ` - Lihat drive yang tersedia

💻 **Command Execution:**
• ` + "`run <command>`" + ` - Jalankan perintah
• ` + "`cmd <command>`" + ` - Jalankan cmd langsung
• ` + "`$<command>`" + ` - Shortcut perintah

📊 **System Monitor:**
• ` + "`sysinfo`" + ` - Info sistem lengkap
• ` + "`cpu`" + ` - Info CPU
• ` + "`memory`" + ` - Info RAM
• ` + "`processes`" + ` - Daftar proses
• ` + "`kill <PID>`" + ` - Matikan proses

ℹ️ **Lainnya:**
• ` + "`help`" + ` - Tampilkan bantuan ini
• ` + "`whoami`" + ` - Info agent`
}

func (as *AgentService) getAgentInfo() string {
	hostname, _ := os.Hostname()
	return fmt.Sprintf(`🤖 **Dardcor Agent v1.0**

**Platform:** %s/%s
**Hostname:** %s
**Go Version:** %s
**Agent Capabilities:**
• 📁 Full File System Access
• 💻 Command Execution
• 📊 System Monitoring
• ⚙️ Process Management
• 🔍 File Search
• 📝 Conversation History

Dibuat oleh **Dardcor** - AI Agent yang powerful untuk mengakses seluruh komputer Anda.`,
		runtime.GOOS, runtime.GOARCH, hostname, runtime.Version())
}

func (as *AgentService) generateTitle(message string) string {
	title := message
	if len(title) > 50 {
		title = title[:50] + "..."
	}
	return title
}

func (as *AgentService) extractPath(message string, prefixes []string) string {
	for _, prefix := range prefixes {
		if strings.HasPrefix(strings.ToLower(message), prefix) {
			return strings.TrimSpace(message[len(prefix):])
		}
	}
	parts := strings.Fields(message)
	if len(parts) > 1 {
		return strings.Join(parts[1:], " ")
	}
	return ""
}

// Utility functions
func formatSize(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
		TB = GB * 1024
	)

	switch {
	case bytes >= TB:
		return fmt.Sprintf("%.2f TB", float64(bytes)/float64(TB))
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

func formatDuration(seconds uint64) string {
	days := seconds / 86400
	hours := (seconds % 86400) / 3600
	mins := (seconds % 3600) / 60

	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm", days, hours, mins)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, mins)
	}
	return fmt.Sprintf("%dm", mins)
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
