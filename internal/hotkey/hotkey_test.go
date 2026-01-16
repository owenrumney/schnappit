package hotkey

import (
	"os"
	"testing"

	"golang.design/x/hotkey"
)

func TestParseHotkey(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantMods    []hotkey.Modifier
		wantKey     hotkey.Key
		wantErr     bool
		errContains string
	}{
		{
			name:     "cmd+shift+x",
			input:    "cmd+shift+x",
			wantMods: []hotkey.Modifier{hotkey.ModCmd, hotkey.ModShift},
			wantKey:  hotkey.KeyX,
			wantErr:  false,
		},
		{
			name:     "ctrl+alt+s",
			input:    "ctrl+alt+s",
			wantMods: []hotkey.Modifier{hotkey.ModCtrl, hotkey.ModOption},
			wantKey:  hotkey.KeyS,
			wantErr:  false,
		},
		{
			name:     "command+f1",
			input:    "command+f1",
			wantMods: []hotkey.Modifier{hotkey.ModCmd},
			wantKey:  hotkey.KeyF1,
			wantErr:  false,
		},
		{
			name:     "uppercase input",
			input:    "CMD+SHIFT+X",
			wantMods: []hotkey.Modifier{hotkey.ModCmd, hotkey.ModShift},
			wantKey:  hotkey.KeyX,
			wantErr:  false,
		},
		{
			name:     "with spaces",
			input:    "cmd + shift + x",
			wantMods: []hotkey.Modifier{hotkey.ModCmd, hotkey.ModShift},
			wantKey:  hotkey.KeyX,
			wantErr:  false,
		},
		{
			name:        "no key",
			input:       "cmd+shift",
			wantErr:     true,
			errContains: "no key specified",
		},
		{
			name:        "no modifier",
			input:       "x",
			wantErr:     true,
			errContains: "invalid hotkey format",
		},
		{
			name:        "multiple keys",
			input:       "cmd+x+y",
			wantErr:     true,
			errContains: "multiple keys",
		},
		{
			name:        "unknown modifier",
			input:       "super+x",
			wantErr:     true,
			errContains: "unknown modifier or key",
		},
		{
			name:     "option modifier",
			input:    "option+shift+a",
			wantMods: []hotkey.Modifier{hotkey.ModOption, hotkey.ModShift},
			wantKey:  hotkey.KeyA,
			wantErr:  false,
		},
		{
			name:     "number key",
			input:    "cmd+1",
			wantMods: []hotkey.Modifier{hotkey.ModCmd},
			wantKey:  hotkey.Key1,
			wantErr:  false,
		},
		{
			name:     "space key",
			input:    "ctrl+space",
			wantMods: []hotkey.Modifier{hotkey.ModCtrl},
			wantKey:  hotkey.KeySpace,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mods, key, err := ParseHotkey(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseHotkey(%q) expected error containing %q, got nil", tt.input, tt.errContains)
				} else if tt.errContains != "" && !contains(err.Error(), tt.errContains) {
					t.Errorf("ParseHotkey(%q) error = %q, want error containing %q", tt.input, err.Error(), tt.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("ParseHotkey(%q) unexpected error: %v", tt.input, err)
				return
			}

			if key != tt.wantKey {
				t.Errorf("ParseHotkey(%q) key = %v, want %v", tt.input, key, tt.wantKey)
			}

			if len(mods) != len(tt.wantMods) {
				t.Errorf("ParseHotkey(%q) got %d modifiers, want %d", tt.input, len(mods), len(tt.wantMods))
				return
			}

			// Check all expected modifiers are present (order may vary)
			for _, wantMod := range tt.wantMods {
				found := false
				for _, gotMod := range mods {
					if gotMod == wantMod {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("ParseHotkey(%q) missing modifier %v", tt.input, wantMod)
				}
			}
		})
	}
}

func TestGetConfiguredHotkey(t *testing.T) {
	// Test that env var takes precedence
	os.Setenv(EnvHotkey, "ctrl+shift+s")
	defer os.Unsetenv(EnvHotkey)

	if got := GetConfiguredHotkey(); got != "ctrl+shift+s" {
		t.Errorf("GetConfiguredHotkey() = %q, want %q", got, "ctrl+shift+s")
	}

	// Test fallback to config (which returns default if no config file)
	os.Unsetenv(EnvHotkey)
	got := GetConfiguredHotkey()
	if got == "" {
		t.Error("GetConfiguredHotkey() should not return empty string")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && stringContains(s, substr)))
}

func stringContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
