package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefault(t *testing.T) {
	cfg := Default()

	if cfg.Hotkey != "cmd+shift+x" {
		t.Errorf("Default().Hotkey = %q, want %q", cfg.Hotkey, "cmd+shift+x")
	}
}

func TestLoadCreatesDefault(t *testing.T) {
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Hotkey == "" {
		t.Error("Load() returned empty hotkey")
	}
	configPath := filepath.Join(tmpDir, configDir, configFile)
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("Load() should create default config file")
	}
}

func TestSaveAndLoad(t *testing.T) {
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	cfg := &Config{Hotkey: "ctrl+alt+p"}
	if err := cfg.Save(); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if loaded.Hotkey != "ctrl+alt+p" {
		t.Errorf("Load().Hotkey = %q, want %q", loaded.Hotkey, "ctrl+alt+p")
	}
}

func TestLoadWithEmptyHotkey(t *testing.T) {

	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	configPath := filepath.Join(tmpDir, configDir, configFile)
	os.MkdirAll(filepath.Dir(configPath), 0755)
	os.WriteFile(configPath, []byte(`{"hotkey": ""}`), 0644)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Hotkey != Default().Hotkey {
		t.Errorf("Load().Hotkey = %q, want default %q", cfg.Hotkey, Default().Hotkey)
	}
}

func TestPath(t *testing.T) {
	path := Path()
	if path == "" {
		t.Error("Path() should not return empty string")
	}
	if !filepath.IsAbs(path) && path[0] != '~' {
		t.Errorf("Path() should return absolute path or start with ~, got %q", path)
	}
}
