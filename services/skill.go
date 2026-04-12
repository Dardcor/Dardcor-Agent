package services

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"dardcor-agent/config"
)

type SkillDefinition struct {
	Name        string `json:"name"`
	Description string `json:"desc"`
	Command     string `json:"cmd"`
	Template    string `json:"template,omitempty"`
}

type SkillService struct {
	skillsDir string
	skills    []SkillDefinition
	mu        sync.RWMutex
}

func NewSkillService() *SkillService {
	dir := "database/skills"
	if config.AppConfig != nil {
		dir = filepath.Join(config.AppConfig.DataDir, "skills")
	}
	os.MkdirAll(dir, 0755)

	ss := &SkillService{
		skillsDir: dir,
		skills:    []SkillDefinition{},
	}
	ss.Reload()
	return ss
}

func (ss *SkillService) Reload() {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	ss.skills = ss.getDefaultSkills()

	files, err := os.ReadDir(ss.skillsDir)
	if err != nil {
		return
	}

	for _, f := range files {
		if !f.IsDir() && (strings.HasSuffix(f.Name(), ".md") || strings.HasSuffix(f.Name(), ".json")) {
			path := filepath.Join(ss.skillsDir, f.Name())
			file, err := os.Open(path)
			if err != nil {
				continue
			}
			data, _ := io.ReadAll(file)
			file.Close()

			if strings.HasSuffix(f.Name(), ".json") {
				var skill SkillDefinition
				if err := json.Unmarshal(data, &skill); err == nil {
					ss.skills = append(ss.skills, skill)
				}
			} else {
				name := strings.TrimSuffix(f.Name(), ".md")
				ss.skills = append(ss.skills, SkillDefinition{
					Name:        name,
					Description: "Supreme skill: " + name,
					Template:    string(data),
				})
			}
		}
	}
}

func (ss *SkillService) GetSkills() []SkillDefinition {
	ss.mu.RLock()
	defer ss.mu.RUnlock()
	return ss.skills
}

func (ss *SkillService) getDefaultSkills() []SkillDefinition {
	return []SkillDefinition{
		{"list_directory", "List files in a directory", "list <path>", ""},
		{"read_file", "Read file contents", "read <path>", ""},
		{"write_file", "Write to a file", "write <path> <content>", ""},
		{"delete_file", "Delete a file or folder", "delete <path>", ""},
		{"search_files", "Search for files", "search <query>", ""},
		{"create_directory", "Create a directory", "mkdir <path>", ""},
		{"execute_command", "Execute a shell command", "run <command>", ""},
		{"system_info", "Get system information", "sysinfo", ""},
		{"list_processes", "List running processes", "processes", ""},
		{"kill_process", "Kill a process by PID", "kill <pid>", ""},
		{"cpu_info", "Get CPU information", "cpu", ""},
		{"memory_info", "Get memory information", "memory", ""},
		{"list_drives", "List available drives", "drives", ""},
		{"file_info", "Get file/directory info", "info <path>", ""},
		{"web_fetch", "Fetch web page content", "fetch <url>", ""},
		{"web_search", "Search the web via DuckDuckGo", "websearch <query>", ""},
		{"grep", "Search file contents with regex", "grep <pattern> [path]", ""},
		{"glob", "Find files by name pattern", "glob <pattern> [path]", ""},
		{"remember", "Store long-term memory", "remember <key> <value>", ""},
	}
}

