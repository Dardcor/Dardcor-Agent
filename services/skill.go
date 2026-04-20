package services

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
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

func (ss *SkillService) RunSkill(name string, args string) (string, error) {
	ss.mu.RLock()
	var skill *SkillDefinition
	for _, s := range ss.skills {
		if s.Name == name {
			skill = &s
			break
		}
	}
	ss.mu.RUnlock()

	if skill == nil {
		return "", fmt.Errorf("skill not found: %s", name)
	}

	if skill.Command == "" {
		return "", fmt.Errorf("skill %s is documentation-only and cannot be executed", name)
	}

	cmdStr := skill.Command
	if args != "" {
		if strings.Contains(cmdStr, "$1") {
			cmdStr = strings.Replace(cmdStr, "$1", args, -1)
		} else {
			cmdStr += " " + args
		}
	}

	out, err := exec.Command("sh", "-c", cmdStr).CombinedOutput()
	if err != nil {
		if runtime.GOOS == "windows" {
			out, err = exec.Command("cmd", "/c", cmdStr).CombinedOutput()
		}
	}

	if err != nil {
		return string(out), fmt.Errorf("skill execution failed: %v", err)
	}
	return string(out), nil
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
