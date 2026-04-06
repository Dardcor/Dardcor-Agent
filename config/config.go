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
	Port       string `json:"port"`
	DataDir    string `json:"data_dir"`
	Settings   models.Settings `json:"settings"`
}

func Init() (*Config, error) {
	execPath, err := os.Executable()
	if err != nil {
		execPath = "."
	}
	baseDir := filepath.Dir(execPath)
	dataDir := filepath.Join(baseDir, "data")
	if _, err := os.Stat(dataDir); os.IsNotExist(err) {
		dataDir = filepath.Join(baseDir, "..", "data")
	}

	// Try current working directory fallback
	if _, err := os.Stat(dataDir); os.IsNotExist(err) {
		cwd, _ := os.Getwd()
		dataDir = filepath.Join(cwd, "data")
		if _, err := os.Stat(dataDir); os.IsNotExist(err) {
			dataDir = filepath.Join(cwd, "..", "data")
		}
	}

	defaultShell := "cmd.exe"
	if runtime.GOOS != "windows" {
		defaultShell = "/bin/bash"
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "25000"
	}

	cfg := &Config{
		Port:    port,
		DataDir: dataDir,
		Settings: models.Settings{
			Theme:          "dark",
			DefaultShell:   defaultShell,
			MaxFileSize:    50 * 1024 * 1024, // 50MB
			CommandTimeout: 30,
			AIProvider:     "local",
			AIModel:        "dardcor-agent-v1",
		},
	}

	// Load settings from file if exists
	settingsPath := filepath.Join(dataDir, "settings.json")
	if data, err := os.ReadFile(settingsPath); err == nil {
		json.Unmarshal(data, &cfg.Settings)
	}

	// Ensure data directories exist
	dirs := []string{
		dataDir,
		filepath.Join(dataDir, "conversations"),
		filepath.Join(dataDir, "commands"),
	}
	for _, dir := range dirs {
		os.MkdirAll(dir, 0755)
	}

	AppConfig = cfg
	return cfg, nil
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
	return filepath.Join(c.DataDir, "conversations")
}

func (c *Config) GetCommandsDir() string {
	return filepath.Join(c.DataDir, "commands")
}
