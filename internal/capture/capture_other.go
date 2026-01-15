//go:build !darwin

package capture

import (
	"fmt"
	"image"
	"runtime"
)

// NumDisplays returns the number of active displays
func NumDisplays() int {
	return 0
}

// GetDisplayBounds returns the bounds of the display at the given index
func GetDisplayBounds(displayIndex int) image.Rectangle {
	return image.Rectangle{}
}

// CaptureDisplay captures the entire display at the given index
func CaptureDisplay(displayIndex int) (*image.RGBA, error) {
	return nil, fmt.Errorf("capture not implemented for %s", runtime.GOOS)
}

// CaptureRect captures a rectangular region from the specified display
func CaptureRect(displayIndex int, rect image.Rectangle) (*image.RGBA, error) {
	return nil, fmt.Errorf("capture not implemented for %s", runtime.GOOS)
}

// CaptureRegion captures a region across all displays
func CaptureRegion(rect image.Rectangle) (*image.RGBA, error) {
	return nil, fmt.Errorf("capture not implemented for %s", runtime.GOOS)
}
