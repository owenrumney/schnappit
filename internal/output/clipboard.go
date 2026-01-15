package output

import (
	"bytes"
	"fmt"
	"image"
	"image/png"

	"golang.design/x/clipboard"
)

var clipboardInitialized bool

// initClipboard initializes the clipboard (must be called from main thread)
func initClipboard() error {
	if clipboardInitialized {
		return nil
	}
	if err := clipboard.Init(); err != nil {
		return fmt.Errorf("failed to initialize clipboard: %w", err)
	}
	clipboardInitialized = true
	return nil
}

// CopyToClipboard copies the image to the system clipboard
func CopyToClipboard(img image.Image) error {
	if err := initClipboard(); err != nil {
		return err
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return fmt.Errorf("failed to encode image: %w", err)
	}

	clipboard.Write(clipboard.FmtImage, buf.Bytes())
	return nil
}
