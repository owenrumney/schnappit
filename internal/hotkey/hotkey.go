package hotkey

import (
	"fmt"
	"log"
	"os"
	"strings"

	"golang.design/x/hotkey"

	"github.com/owenrumney/schnappit/internal/config"
)

// EnvHotkey allows overriding the config file via environment variable
const EnvHotkey = "SCHNAPPIT_HOTKEY"

// Shortcut represents a registered global hotkey
type Shortcut struct {
	hk       *hotkey.Hotkey
	onTrigger func()
	stop     chan struct{}
}

// modifierMap maps string names to hotkey modifiers
// Note: On macOS, ModCmd is Command, ModOption is Option/Alt
var modifierMap = map[string]hotkey.Modifier{
	"cmd":     hotkey.ModCmd,
	"command": hotkey.ModCmd,
	"ctrl":    hotkey.ModCtrl,
	"control": hotkey.ModCtrl,
	"shift":   hotkey.ModShift,
	"alt":     hotkey.ModOption,
	"option":  hotkey.ModOption,
	"opt":     hotkey.ModOption,
}

// keyMap maps string names to hotkey keys
var keyMap = map[string]hotkey.Key{
	"a": hotkey.KeyA, "b": hotkey.KeyB, "c": hotkey.KeyC, "d": hotkey.KeyD,
	"e": hotkey.KeyE, "f": hotkey.KeyF, "g": hotkey.KeyG, "h": hotkey.KeyH,
	"i": hotkey.KeyI, "j": hotkey.KeyJ, "k": hotkey.KeyK, "l": hotkey.KeyL,
	"m": hotkey.KeyM, "n": hotkey.KeyN, "o": hotkey.KeyO, "p": hotkey.KeyP,
	"q": hotkey.KeyQ, "r": hotkey.KeyR, "s": hotkey.KeyS, "t": hotkey.KeyT,
	"u": hotkey.KeyU, "v": hotkey.KeyV, "w": hotkey.KeyW, "x": hotkey.KeyX,
	"y": hotkey.KeyY, "z": hotkey.KeyZ,
	"0": hotkey.Key0, "1": hotkey.Key1, "2": hotkey.Key2, "3": hotkey.Key3,
	"4": hotkey.Key4, "5": hotkey.Key5, "6": hotkey.Key6, "7": hotkey.Key7,
	"8": hotkey.Key8, "9": hotkey.Key9,
	"space":  hotkey.KeySpace,
	"return": hotkey.KeyReturn,
	"enter":  hotkey.KeyReturn,
	"escape": hotkey.KeyEscape,
	"esc":    hotkey.KeyEscape,
	"tab":    hotkey.KeyTab,
	"delete": hotkey.KeyDelete,
	"f1":     hotkey.KeyF1, "f2": hotkey.KeyF2, "f3": hotkey.KeyF3,
	"f4":     hotkey.KeyF4, "f5": hotkey.KeyF5, "f6": hotkey.KeyF6,
	"f7":     hotkey.KeyF7, "f8": hotkey.KeyF8, "f9": hotkey.KeyF9,
	"f10":    hotkey.KeyF10, "f11": hotkey.KeyF11, "f12": hotkey.KeyF12,
}

// ParseHotkey parses a hotkey string like "cmd+shift+x" into modifiers and key
func ParseHotkey(s string) ([]hotkey.Modifier, hotkey.Key, error) {
	parts := strings.Split(strings.ToLower(s), "+")
	if len(parts) < 2 {
		return nil, 0, fmt.Errorf("invalid hotkey format: %s (expected format: modifier+key, e.g., cmd+shift+x)", s)
	}

	var mods []hotkey.Modifier
	var key hotkey.Key
	var keyFound bool

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if mod, ok := modifierMap[part]; ok {
			mods = append(mods, mod)
		} else if k, ok := keyMap[part]; ok {
			if keyFound {
				return nil, 0, fmt.Errorf("multiple keys specified in hotkey: %s", s)
			}
			key = k
			keyFound = true
		} else {
			return nil, 0, fmt.Errorf("unknown modifier or key: %s", part)
		}
	}

	if !keyFound {
		return nil, 0, fmt.Errorf("no key specified in hotkey: %s", s)
	}
	if len(mods) == 0 {
		return nil, 0, fmt.Errorf("no modifiers specified in hotkey: %s", s)
	}

	return mods, key, nil
}

// GetConfiguredHotkey returns the hotkey string from environment, config file, or default
func GetConfiguredHotkey() string {
	// Environment variable takes precedence
	if env := os.Getenv(EnvHotkey); env != "" {
		return env
	}

	// Load from config file
	cfg, _ := config.Load()
	return cfg.Hotkey
}

// New creates and registers a new global hotkey shortcut
func New(onTrigger func()) (*Shortcut, error) {
	hotkeyStr := GetConfiguredHotkey()

	mods, key, err := ParseHotkey(hotkeyStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse hotkey %q: %w", hotkeyStr, err)
	}

	hk := hotkey.New(mods, key)
	if err := hk.Register(); err != nil {
		return nil, fmt.Errorf("failed to register hotkey %q: %w (you may need to grant accessibility permissions in System Settings > Privacy & Security > Accessibility)", hotkeyStr, err)
	}

	s := &Shortcut{
		hk:        hk,
		onTrigger: onTrigger,
		stop:      make(chan struct{}),
	}

	log.Printf("Global hotkey registered: %s", hotkeyStr)

	// Start listening for hotkey events
	go s.listen()

	return s, nil
}

// listen waits for hotkey events and triggers the callback
func (s *Shortcut) listen() {
	for {
		select {
		case <-s.stop:
			return
		case <-s.hk.Keydown():
			log.Println("Hotkey triggered")
			if s.onTrigger != nil {
				s.onTrigger()
			}
		}
	}
}

// Unregister unregisters the hotkey and stops listening
func (s *Shortcut) Unregister() {
	close(s.stop)
	if err := s.hk.Unregister(); err != nil {
		log.Printf("Failed to unregister hotkey: %v", err)
	}
}
