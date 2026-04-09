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
	Port     string         `json:"port"`
	DataDir  string         `json:"data_dir"`
	Settings models.Settings `json:"settings"`
	AI       AIConfig       `json:"ai"`
}

type AIConfig struct {
	Provider   string  `json:"provider"`
	Model      string  `json:"model"`
	APIKey     string  `json:"api_key"`
	BaseURL    string  `json:"base_url"`
	MaxTokens  int     `json:"max_tokens"`
	Streaming  bool    `json:"streaming"`
	Temperature float64 `json:"temperature"`
}

type UserConfig struct {
	Provider       string `json:"provider"`
	Model          string `json:"model"`
	APIKey         string `json:"api_key"`
	ProviderBaseURL string `json:"provider_base_url"`
	Port           string `json:"port"`
	Initialized    bool   `json:"initialized"`
}

func Init() (*Config, error) {
	execPath, err := os.Executable()
	if err != nil {
		execPath = "."
	}
	baseDir := filepath.Dir(execPath)
	dataDir := filepath.Join(baseDir, "database")
	if _, err := os.Stat(dataDir); os.IsNotExist(err) {
		dataDir = filepath.Join(baseDir, "..", "database")
	}
	if _, err := os.Stat(dataDir); os.IsNotExist(err) {
		cwd, _ := os.Getwd()
		dataDir = filepath.Join(cwd, "database")
		if _, err := os.Stat(dataDir); os.IsNotExist(err) {
			dataDir = filepath.Join(cwd, "..", "database")
		}
	}

	defaultShell := "cmd.exe"
	if runtime.GOOS != "windows" {
		defaultShell = "/bin/bash"
	}

	port := "25000"
	
	userCfg := loadUserConfig()

	aiCfg := buildAIConfig(userCfg)

	cfg := &Config{
		Port:    port,
		DataDir: dataDir,
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

	settingsPath := filepath.Join(dataDir, "settings.json")
	if data, err := os.ReadFile(settingsPath); err == nil {
		json.Unmarshal(data, &cfg.Settings)
	}

	dirs := []string{
		dataDir,
		filepath.Join(dataDir, "conversations-web"),
		filepath.Join(dataDir, "commands"),
		filepath.Join(dataDir, "settings"),
		filepath.Join(dataDir, "model", "antigravity"),
	}
	for _, dir := range dirs {
		os.MkdirAll(dir, 0755)
	}

	AppConfig = cfg
	return cfg, nil
}

func loadUserConfig() UserConfig {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return UserConfig{}
	}
	cfgPath := filepath.Join(homeDir, ".dardcor", "config.json")
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
		}
	}

	return ai
}

func (c *Config) SaveSettings() error {
	settingsPath := filepath.Join(c.DataDir, "settings.json")
	data, err := json.MarshalIndent(c.Settings, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(settingsPath, data, 0644)
}

func (c *Config) GetConversationsDir() string {
	return filepath.Join(c.DataDir, "conversations-web")
}

func (c *Config) GetCommandsDir() string {
	return filepath.Join(c.DataDir, "commands")
}

func (c *Config) IsAIEnabled() bool {
	return c.AI.Provider != "local" && c.AI.Provider != ""
}

