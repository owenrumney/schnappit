//go:build !darwin

package app

import "fmt"

// IsLoginItemEnabled checks if the app is set to start on login
func IsLoginItemEnabled() bool {
	return false
}

// SetLoginItemEnabled enables or disables starting on login
func SetLoginItemEnabled(enabled bool) error {
	return fmt.Errorf("login items not supported on this platform")
}
