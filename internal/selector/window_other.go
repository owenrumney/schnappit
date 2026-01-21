//go:build !darwin

package selector

// positionWindowOnDisplay is a no-op on non-darwin platforms
func positionWindowOnDisplay(x, y, width, height float32) {
	// No-op: platform-specific positioning not implemented
}
