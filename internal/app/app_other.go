//go:build !darwin

package app

// hideDockIcon is a no-op on non-macOS platforms
func hideDockIcon() {}
