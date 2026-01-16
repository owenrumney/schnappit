package output

import (
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	// DefaultDir is the default directory for saving screenshots
	DefaultDir = "Pictures/schnappit"
	// FilePermissions restricts screenshot files to owner read/write only
	FilePermissions = 0600
)

// SaveToFile saves the image to a file in the schnappit directory
func SaveToFile(img image.Image) (string, error) {
	return SaveToFileWithName(img, generateFilename())
}

// SaveToFileWithName saves the image to a file with the specified name
func SaveToFileWithName(img image.Image, filename string) (string, error) {
	if strings.Contains(filename, "..") || strings.ContainsAny(filename, `/\`) {
		return "", fmt.Errorf("invalid filename: path traversal not allowed")
	}

	dir, err := getOutputDir()
	if err != nil {
		return "", err
	}

	outPath := filepath.Join(dir, filename)

	absDir, err := filepath.Abs(dir)
	if err != nil {
		return "", fmt.Errorf("failed to resolve output directory: %w", err)
	}
	absPath, err := filepath.Abs(outPath)
	if err != nil {
		return "", fmt.Errorf("failed to resolve output path: %w", err)
	}
	if !strings.HasPrefix(absPath, absDir+string(filepath.Separator)) {
		return "", fmt.Errorf("invalid filename: path escapes output directory")
	}

	file, err := os.OpenFile(outPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, FilePermissions)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}

	defer file.Close()

	if err := png.Encode(file, img); err != nil {
		file.Close()
		os.Remove(outPath)
		return "", fmt.Errorf("failed to encode image: %w", err)
	}
	if err := file.Sync(); err != nil {
		file.Close()
		os.Remove(outPath)
		return "", fmt.Errorf("failed to sync file: %w", err)
	}

	return outPath, nil
}

// getOutputDir returns the output directory, creating it if necessary
func getOutputDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	dir := filepath.Join(home, DefaultDir)

	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create output directory: %w", err)
	}

	return dir, nil
}

// generateFilename generates a timestamp-based filename
func generateFilename() string {
	return fmt.Sprintf("schnappit-%s.png", time.Now().Format("2006-01-02-150405"))
}
