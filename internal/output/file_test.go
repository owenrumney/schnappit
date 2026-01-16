package output

import (
	"image"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestGenerateFilename(t *testing.T) {
	filename := generateFilename()

	// Should start with "schnappit-"
	if !strings.HasPrefix(filename, "schnappit-") {
		t.Errorf("Filename should start with 'schnappit-', got %s", filename)
	}

	// Should end with ".png"
	if !strings.HasSuffix(filename, ".png") {
		t.Errorf("Filename should end with '.png', got %s", filename)
	}

	// Should contain today's date
	today := time.Now().Format("2006-01-02")
	if !strings.Contains(filename, today) {
		t.Errorf("Filename should contain today's date %s, got %s", today, filename)
	}
}

func TestGetOutputDir(t *testing.T) {
	dir, err := getOutputDir()
	if err != nil {
		t.Fatalf("getOutputDir() error = %v", err)
	}

	// Should be within home directory
	home, _ := os.UserHomeDir()
	if !strings.HasPrefix(dir, home) {
		t.Errorf("Output dir should be under home directory, got %s", dir)
	}

	// Should end with schnappit
	if !strings.HasSuffix(dir, "schnappit") {
		t.Errorf("Output dir should end with 'schnappit', got %s", dir)
	}

	// Directory should exist
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		t.Errorf("Output directory should exist after getOutputDir()")
	}
}

func TestSaveToFileWithName_PathTraversal(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))

	tests := []struct {
		name     string
		filename string
		wantErr  bool
	}{
		{
			name:     "valid filename",
			filename: "test.png",
			wantErr:  false,
		},
		{
			name:     "path traversal with ..",
			filename: "../../../etc/passwd",
			wantErr:  true,
		},
		{
			name:     "path traversal with forward slash",
			filename: "foo/bar.png",
			wantErr:  true,
		},
		{
			name:     "path traversal with backslash",
			filename: "foo\\bar.png",
			wantErr:  true,
		},
		{
			name:     "hidden parent directory",
			filename: "..test.png",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path, err := SaveToFileWithName(img, tt.filename)
			if (err != nil) != tt.wantErr {
				t.Errorf("SaveToFileWithName() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err == nil {
				// Clean up the test file
				os.Remove(path)
			}
		})
	}
}

func TestSaveToFile(t *testing.T) {
	// Create a simple test image
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))

	path, err := SaveToFile(img)
	if err != nil {
		t.Fatalf("SaveToFile() error = %v", err)
	}

	// Clean up
	defer os.Remove(path)

	// Verify file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("SaveToFile() file should exist at %s", path)
	}

	// Verify file permissions (owner read/write only)
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Failed to stat file: %v", err)
	}
	mode := info.Mode().Perm()
	if mode != FilePermissions {
		t.Errorf("File permissions = %o, want %o", mode, FilePermissions)
	}

	// Verify file is in correct directory
	expectedDir := filepath.Join(os.Getenv("HOME"), DefaultDir)
	if !strings.HasPrefix(path, expectedDir) {
		t.Errorf("File should be in %s, got %s", expectedDir, path)
	}
}

func TestSaveToFileWithName_CreatesValidPNG(t *testing.T) {
	// Create a test image with some content
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	for y := 0; y < 100; y++ {
		for x := 0; x < 100; x++ {
			img.Set(x, y, image.White)
		}
	}

	path, err := SaveToFileWithName(img, "test-image.png")
	if err != nil {
		t.Fatalf("SaveToFileWithName() error = %v", err)
	}

	// Clean up
	defer os.Remove(path)

	// Verify file is not empty
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Failed to stat file: %v", err)
	}
	if info.Size() == 0 {
		t.Error("Saved file should not be empty")
	}

	// Verify it starts with PNG magic bytes
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}
	pngMagic := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	for i, b := range pngMagic {
		if data[i] != b {
			t.Errorf("File doesn't have PNG magic bytes at position %d", i)
			break
		}
	}
}
