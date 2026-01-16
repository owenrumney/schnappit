package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

const (
	configDir  = ".config/schnappit"
	configFile = "config.json"
)

// Config represents the application configuration
type Config struct {
	Hotkey string `json:"hotkey"`
}

// Default returns the default configuration
func Default() *Config {
	return &Config{
		Hotkey: "cmd+shift+x",
	}
}

// Load reads the configuration from disk, creating a default if it doesn't exist
func Load() (*Config, error) {
	path, err := configPath()
	if err != nil {
		return Default(), nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Create default config
			cfg := Default()
			_ = cfg.Save() // Best effort to save default
			return cfg, nil
		}
		return Default(), nil
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Default(), nil
	}

	// Ensure hotkey has a value
	if cfg.Hotkey == "" {
		cfg.Hotkey = Default().Hotkey
	}

	return &cfg, nil
}

// Save writes the configuration to disk
func (c *Config) Save() error {
	path, err := configPath()
	if err != nil {
		return err
	}

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// configPath returns the full path to the config file
func configPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, configDir, configFile), nil
}

// Path returns the config file path for display purposes
func Path() string {
	path, err := configPath()
	if err != nil {
		return "~/" + configDir + "/" + configFile
	}
	return path
}
