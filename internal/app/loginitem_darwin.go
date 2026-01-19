//go:build darwin

package app

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

const (
	launchAgentLabel = "com.owenrumney.schnappit"
	launchAgentFile  = "com.owenrumney.schnappit.plist"
)

// IsLoginItemEnabled checks if the app is set to start on login
func IsLoginItemEnabled() bool {
	plistPath, err := getLaunchAgentPath()
	if err != nil {
		log.Printf("Failed to get launch agent path: %v", err)
		return false
	}
	_, err = os.Stat(plistPath)
	enabled := err == nil
	log.Printf("Login item enabled: %v (path: %s)", enabled, plistPath)
	return enabled
}

// SetLoginItemEnabled enables or disables starting on login
func SetLoginItemEnabled(enabled bool) error {
	log.Printf("Setting login item enabled: %v", enabled)
	plistPath, err := getLaunchAgentPath()
	if err != nil {
		log.Printf("Failed to get launch agent path: %v", err)
		return err
	}

	if enabled {
		return createLaunchAgent(plistPath)
	}
	return removeLaunchAgent(plistPath)
}

// getLaunchAgentPath returns the path to the LaunchAgent plist
func getLaunchAgentPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(home, "Library", "LaunchAgents", launchAgentFile), nil
}

// createLaunchAgent creates the LaunchAgent plist file
func createLaunchAgent(plistPath string) error {
	log.Printf("Creating launch agent at: %s", plistPath)
	appPath, err := os.Executable()
	if err != nil {
		log.Printf("Failed to get executable path: %v", err)
		return fmt.Errorf("failed to get executable path: %w", err)
	}
	log.Printf("Executable path: %s", appPath)

	// Ensure LaunchAgents directory exists
	dir := filepath.Dir(plistPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Printf("Failed to create LaunchAgents directory: %v", err)
		return fmt.Errorf("failed to create LaunchAgents directory: %w", err)
	}

	plistContent := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>Label</key>
	<string>%s</string>
	<key>ProgramArguments</key>
	<array>
		<string>%s</string>
	</array>
	<key>RunAtLoad</key>
	<true/>
	<key>KeepAlive</key>
	<false/>
</dict>
</plist>
`, launchAgentLabel, appPath)

	if err := os.WriteFile(plistPath, []byte(plistContent), 0644); err != nil {
		log.Printf("Failed to write LaunchAgent plist: %v", err)
		return fmt.Errorf("failed to write LaunchAgent plist: %w", err)
	}

	log.Printf("Successfully created launch agent")
	return nil
}

// removeLaunchAgent removes the LaunchAgent plist file
func removeLaunchAgent(plistPath string) error {
	log.Printf("Removing launch agent at: %s", plistPath)
	err := os.Remove(plistPath)
	if err != nil && !os.IsNotExist(err) {
		log.Printf("Failed to remove LaunchAgent plist: %v", err)
		return fmt.Errorf("failed to remove LaunchAgent plist: %w", err)
	}
	log.Printf("Successfully removed launch agent")
	return nil
}
