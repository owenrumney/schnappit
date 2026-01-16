//go:build darwin

package app

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa

#import <Cocoa/Cocoa.h>

void HideDockIcon() {
    [NSApp setActivationPolicy:NSApplicationActivationPolicyAccessory];
}
*/
import "C"

// hideDockIcon hides the dock icon on macOS
func hideDockIcon() {
	C.HideDockIcon()
}
