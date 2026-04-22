package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"

	"dardcor-agent/models"
)

var (
	AppConfig *Config
)

type Config struct {
	Port          string          `json:"port"`
	DataDir       string          `json:"data_dir"`
	Settings      models.Settings `json:"settings"`
	AI            AIConfig        `json:"ai"`
	ProjectRules  []string        `json:"-"`
	ProjectPrompt string          `json:"-"`
}

type AIConfig struct {
	Provider    string  `json:"provider"`
	Model       string  `json:"model"`
	APIKey      string  `json:"api_key"`
	BaseURL     string  `json:"base_url"`
	MaxTokens   int     `json:"max_tokens"`
	Streaming   bool    `json:"streaming"`
	Temperature float64 `json:"temperature"`
}

type UserConfig struct {
	Provider        string `json:"provider"`
	Model           string `json:"model"`
	APIKey          string `json:"api_key"`
	ProviderBaseURL string `json:"provider_base_url"`
	Port            string `json:"port"`
	Initialized     bool   `json:"initialized"`
}

func Init() (*Config, error) {
	homeDir, _ := os.UserHomeDir()
	baseDir := filepath.Join(homeDir, ".dardcor")

	if envDir := os.Getenv("DARDCOR_DATA_DIR"); envDir != "" {
		baseDir = envDir
	}

	dirs := []string{
		baseDir,
		filepath.Join(baseDir, "session"),
		filepath.Join(baseDir, "cache"),
		filepath.Join(baseDir, "plugins"),
	}
	for _, dir := range dirs {
		os.MkdirAll(dir, 0755)
	}

	defaultShell := "cmd.exe"
	if runtime.GOOS != "windows" {
		defaultShell = "/bin/bash"
	}

	userCfg := loadUserConfig(baseDir)
	aiCfg := buildAIConfig(userCfg)

	cfg := &Config{
		Port:    "25000",
		DataDir: baseDir,
		AI:      aiCfg,
		Settings: models.Settings{
			Theme:          "dark",
			DefaultShell:   defaultShell,
			MaxFileSize:    50 * 1024 * 1024,
			CommandTimeout: 30,
			AIProvider:     aiCfg.Provider,
			AIModel:        aiCfg.Model,
		},
	}

	// Load settings if exists
	settingsPath := filepath.Join(baseDir, "settings.json")
	if data, err := os.ReadFile(settingsPath); err == nil {
		json.Unmarshal(data, &cfg.Settings)
	}

	// Load project-level config if exists
	loadProjectConfig(cfg)

	AppConfig = cfg
	return cfg, nil
}

func loadUserConfig(baseDir string) UserConfig {
	cfgPath := filepath.Join(baseDir, "config.json")
	data, err := os.ReadFile(cfgPath)
	if err != nil {
		return UserConfig{}
	}
	var cfg UserConfig
	json.Unmarshal(data, &cfg)
	return cfg
}

func buildAIConfig(userCfg UserConfig) AIConfig {
	ai := AIConfig{
		Provider:    "local",
		Model:       "dardcor-agent",
		MaxTokens:   4096,
		Streaming:   true,
		Temperature: 0.7,
	}

	if userCfg.Provider != "" {
		ai.Provider = userCfg.Provider
	}
	if userCfg.Model != "" {
		ai.Model = userCfg.Model
	}
	if userCfg.APIKey != "" {
		ai.APIKey = userCfg.APIKey
	}
	if userCfg.ProviderBaseURL != "" {
		ai.BaseURL = userCfg.ProviderBaseURL
	}

	if v := os.Getenv("DARDCOR_AI_PROVIDER"); v != "" {
		ai.Provider = v
	}
	if v := os.Getenv("DARDCOR_AI_MODEL"); v != "" {
		ai.Model = v
	}
	if v := os.Getenv("DARDCOR_API_KEY"); v != "" {
		ai.APIKey = v
	}
	if v := os.Getenv("DARDCOR_BASE_URL"); v != "" {
		ai.BaseURL = v
	}

	if ai.BaseURL == "" {
		switch ai.Provider {
		case "openai":
			ai.BaseURL = "https://api.openai.com/v1"
		case "anthropic":
			ai.BaseURL = "https://api.anthropic.com"
		case "gemini":
			ai.BaseURL = "https://generativelanguage.googleapis.com"
		case "deepseek":
			ai.BaseURL = "https://api.deepseek.com/v1"
		case "openrouter":
			ai.BaseURL = "https://openrouter.ai/api/v1"
		case "ollama":
			ai.BaseURL = "http://localhost:11434/v1"
		case "nvidia":
			ai.BaseURL = "https://integrate.api.nvidia.com/v1"
		}
	}

	return ai
}

func loadProjectConfig(cfg *Config) {
	// 1. Load Global Rules (~/.dardcor.json)
	homeDir, _ := os.UserHomeDir()
	globalPath := filepath.Join(homeDir, ".dardcor.json")
	if data, err := os.ReadFile(globalPath); err == nil {
		var gCfg struct {
			Model    string            `json:"model"`
			Rules    []string          `json:"rules"`
			Prompt   string            `json:"prompt"`
			Settings map[string]string `json:"settings"`
		}
		if err := json.Unmarshal(data, &gCfg); err == nil {
			if len(gCfg.Rules) > 0 {
				cfg.ProjectRules = append(cfg.ProjectRules, gCfg.Rules...)
			}
			if gCfg.Prompt != "" {
				cfg.ProjectPrompt = gCfg.Prompt
			}
		}
	}

	// 2. Load Project Rules (./.dardcor.json)
	projectPath := ".dardcor.json"
	if data, err := os.ReadFile(projectPath); err == nil {
		var pCfg struct {
			Model    string            `json:"model"`
			Rules    []string          `json:"rules"`
			Prompt   string            `json:"prompt"`
			Settings map[string]string `json:"settings"`
		}

		if err := json.Unmarshal(data, &pCfg); err == nil {
			if len(pCfg.Rules) > 0 {
				// Project rules extend global rules
				cfg.ProjectRules = append(cfg.ProjectRules, pCfg.Rules...)
			}
			if pCfg.Prompt != "" {
				// Project prompt overrides global prompt
				cfg.ProjectPrompt = pCfg.Prompt
			}
		}
	}
}

func (c *Config) ConfigPath() string  { return filepath.Join(c.DataDir, "config.json") }
func (c *Config) AccountPath() string { return filepath.Join(c.DataDir, "account.json") }
func (c *Config) SessionDir() string  { return filepath.Join(c.DataDir, "session") }
func (c *Config) CacheDir() string    { return filepath.Join(c.DataDir, "cache") }
func (c *Config) PluginsDir() string  { return filepath.Join(c.DataDir, "plugins") }

func (c *Config) SaveSettings() error {
	settingsPath := filepath.Join(c.DataDir, "settings.json")
	data, err := json.MarshalIndent(c.Settings, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(settingsPath, data, 0644)
}

func (c *Config) GetConversationsDir(source string) string {
	dir := filepath.Join(c.SessionDir(), source)
	os.MkdirAll(dir, 0755)
	return dir
}

func (c *Config) GetCommandsDir() string {
	dir := filepath.Join(c.CacheDir(), "commands")
	os.MkdirAll(dir, 0755)
	return dir
}

func (c *Config) IsAIEnabled() bool {
	return c.AI.Provider != "local" && c.AI.Provider != ""
}
