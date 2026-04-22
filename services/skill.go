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
	Name        string   `json:"name"`
	Description string   `json:"desc"`
	Command     string   `json:"cmd"`
	Template    string   `json:"template,omitempty"`
	Category    string   `json:"category,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	Enabled     bool     `json:"enabled"`
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
					if !skill.Enabled {
						skill.Enabled = true
					}
					ss.skills = append(ss.skills, skill)
				}
			} else {
				skill := ss.parseMDSkill(f.Name(), string(data))
				ss.skills = append(ss.skills, skill)
			}
		}
	}
}

func (ss *SkillService) parseMDSkill(filename string, content string) SkillDefinition {
	baseName := strings.TrimSuffix(filename, ".md")
	skill := SkillDefinition{
		Name:        baseName,
		Description: "Skill: " + baseName,
		Category:    "custom",
		Enabled:     true,
	}

	if !strings.HasPrefix(content, "---") {
		skill.Template = content
		return skill
	}

	rest := content[3:]
	endIdx := strings.Index(rest, "\n---")
	if endIdx == -1 {
		skill.Template = content
		return skill
	}

	frontmatter := strings.TrimSpace(rest[:endIdx])
	body := strings.TrimSpace(rest[endIdx+4:])

	for _, line := range strings.Split(frontmatter, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		colonIdx := strings.Index(line, ":")
		if colonIdx == -1 {
			continue
		}
		key := strings.TrimSpace(line[:colonIdx])
		val := strings.TrimSpace(line[colonIdx+1:])
		val = strings.Trim(val, "\"'")

		switch strings.ToLower(key) {
		case "name":
			if val != "" {
				skill.Name = val
			}
		case "description", "desc":
			skill.Description = val
		case "category":
			skill.Category = val
		case "tags":
			rawTags := strings.Split(val, ",")
			for _, t := range rawTags {
				t = strings.TrimSpace(t)
				t = strings.Trim(t, "[]\"' ")
				if t != "" {
					skill.Tags = append(skill.Tags, t)
				}
			}
		case "enabled":
			skill.Enabled = strings.ToLower(val) != "false"
		case "command", "cmd":
			skill.Command = val
		}
	}

	skill.Template = body
	return skill
}

func (ss *SkillService) GetSkills() []SkillDefinition {
	ss.mu.RLock()
	defer ss.mu.RUnlock()
	return ss.skills
}

func (ss *SkillService) GetSkillsByCategory(category string) []SkillDefinition {
	ss.mu.RLock()
	defer ss.mu.RUnlock()
	var result []SkillDefinition
	for _, s := range ss.skills {
		if strings.EqualFold(s.Category, category) {
			result = append(result, s)
		}
	}
	return result
}

func (ss *SkillService) EnableSkill(name string, enabled bool) bool {
	ss.mu.Lock()
	defer ss.mu.Unlock()
	for i := range ss.skills {
		if strings.EqualFold(ss.skills[i].Name, name) {
			ss.skills[i].Enabled = enabled
			return true
		}
	}
	return false
}

func (ss *SkillService) GetSkill(name string) *SkillDefinition {
	ss.mu.RLock()
	defer ss.mu.RUnlock()
	for i := range ss.skills {
		if strings.EqualFold(ss.skills[i].Name, name) {
			return &ss.skills[i]
		}
	}
	return nil
}

func (ss *SkillService) getDefaultSkills() []SkillDefinition {
	return []SkillDefinition{
		{Name: "list_directory", Description: "List files in a directory", Command: "list <path>", Category: "filesystem", Enabled: true},
		{Name: "read_file", Description: "Read file contents", Command: "read <path>", Category: "filesystem", Enabled: true},
		{Name: "write_file", Description: "Write to a file", Command: "write <path> <content>", Category: "filesystem", Enabled: true},
		{Name: "delete_file", Description: "Delete a file or folder", Command: "delete <path>", Category: "filesystem", Enabled: true},
		{Name: "search_files", Description: "Search for files", Command: "search <query>", Category: "filesystem", Enabled: true},
		{Name: "create_directory", Description: "Create a directory", Command: "mkdir <path>", Category: "filesystem", Enabled: true},
		{Name: "execute_command", Description: "Execute a shell command", Command: "run <command>", Category: "shell", Enabled: true},
		{Name: "system_info", Description: "Get system information", Command: "sysinfo", Category: "system", Enabled: true},
		{Name: "list_processes", Description: "List running processes", Command: "processes", Category: "system", Enabled: true},
		{Name: "kill_process", Description: "Kill a process by PID", Command: "kill <pid>", Category: "system", Enabled: true},
		{Name: "cpu_info", Description: "Get CPU information", Command: "cpu", Category: "system", Enabled: true},
		{Name: "memory_info", Description: "Get memory information", Command: "memory", Category: "system", Enabled: true},
		{Name: "list_drives", Description: "List available drives", Command: "drives", Category: "system", Enabled: true},
		{Name: "file_info", Description: "Get file/directory info", Command: "info <path>", Category: "filesystem", Enabled: true},
		{Name: "web_fetch", Description: "Fetch web page content", Command: "fetch <url>", Category: "web", Enabled: true},
		{Name: "web_search", Description: "Search the web via DuckDuckGo", Command: "websearch <query>", Category: "web", Enabled: true},
		{Name: "grep", Description: "Search file contents with regex", Command: "grep <pattern> [path]", Category: "code", Enabled: true},
		{Name: "glob", Description: "Find files by name pattern", Command: "glob <pattern> [path]", Category: "filesystem", Enabled: true},
		{Name: "remember", Description: "Store long-term memory", Command: "remember <key> <value>", Category: "memory", Enabled: true},
		{
			Name:        "code_review",
			Description: "Perform a comprehensive code review on the specified file or directory",
			Command:     "grep <pattern> <path>",
			Category:    "code",
			Tags:        []string{"review", "quality", "analysis"},
			Enabled:     true,
			Template: `Analyze code for: bugs, security issues, performance bottlenecks, style violations.
Steps: read file -> grep for patterns -> check complexity -> summarize findings.
Output: severity-ranked list of issues with file:line references.`,
		},
		{
			Name:        "git_operations",
			Description: "Execute common git operations: status, diff, log, commit, branch management",
			Command:     "run git <subcommand>",
			Category:    "vcs",
			Tags:        []string{"git", "vcs", "version-control"},
			Enabled:     true,
			Template: `Git operations available: status, diff, log, add, commit, push, pull, branch, checkout, merge.
Always run 'git status' first, then proceed with requested operation.
Use 'run git <command>' syntax. Never force-push to main without confirmation.`,
		},
		{
			Name:        "testing",
			Description: "Run tests, analyze failures, and suggest fixes",
			Command:     "run <test-command>",
			Category:    "code",
			Tags:        []string{"test", "quality", "tdd"},
			Enabled:     true,
			Template: `Testing workflow: discover test files -> run test suite -> parse failures -> suggest fixes.
Common commands: 'run go test ./...', 'run npm test', 'run pytest'.
On failure: read the failing test, read the source, diagnose the root cause.`,
		},
		{
			Name:        "debugging",
			Description: "Systematic debugging: identify errors, trace root cause, propose and apply fixes",
			Command:     "grep <error-pattern> <path>",
			Category:    "code",
			Tags:        []string{"debug", "errors", "fix"},
			Enabled:     true,
			Template: `Debugging protocol:
1. Read error message carefully
2. Grep for error pattern in source
3. Read relevant source files
4. Trace call stack / data flow
5. Identify root cause
6. Apply minimal fix
7. Verify fix with test or run`,
		},
		{
			Name:        "refactoring",
			Description: "Refactor code: extract functions, rename variables, reduce complexity, improve structure",
			Command:     "read <path>",
			Category:    "code",
			Tags:        []string{"refactor", "clean", "structure"},
			Enabled:     true,
			Template: `Refactoring steps:
1. Read target file(s)
2. Identify code smells (duplication, long functions, magic numbers)
3. Plan refactoring (extract, rename, restructure)
4. Apply changes incrementally
5. Verify functionality preserved
Keep public API signatures intact unless explicitly requested.`,
		},
		{
			Name:        "project_init",
			Description: "Initialize a new project with proper structure, config files, and boilerplate",
			Command:     "mkdir <path>",
			Category:    "scaffold",
			Tags:        []string{"init", "scaffold", "boilerplate"},
			Enabled:     true,
			Template: `Project initialization workflow:
1. Ask/infer: language, framework, project name
2. Create directory structure
3. Write config files (package.json / go.mod / pyproject.toml)
4. Write entry point (main / index)
5. Write .gitignore
6. Run dependency install if applicable
7. Confirm project is runnable`,
		},
	}
}