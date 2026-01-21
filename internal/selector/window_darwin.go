//go:build darwin

package selector

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa

#import <Cocoa/Cocoa.h>

// PositionWindowOnDisplay moves a window with the given title to cover the specified display area
// x, y are the display origin in screen coordinates (Quartz coordinates, origin at top-left of main display)
// width, height are the display dimensions
void PositionWindowOnDisplay(float x, float y, float width, float height) {
    // Must run on main thread for AppKit operations
    if (![NSThread isMainThread]) {
        dispatch_sync(dispatch_get_main_queue(), ^{
            PositionWindowOnDisplay(x, y, width, height);
        });
        return;
    }

    NSApplication *app = [NSApplication sharedApplication];

    // Find the selector window (titled "Select Region")
    NSWindow *targetWindow = nil;
    for (NSWindow *window in [app windows]) {
        if ([[window title] isEqualToString:@"Select Region"]) {
            targetWindow = window;
            break;
        }
    }

    if (!targetWindow) {
        NSLog(@"PositionWindowOnDisplay: Could not find 'Select Region' window");
        return;
    }

    // macOS screen coordinates have origin at bottom-left of the primary display
    // CoreGraphics (used by capture) has origin at top-left
    // We need to convert from CG coordinates to NS coordinates

    // Get the primary screen height for coordinate conversion
    NSScreen *primaryScreen = [[NSScreen screens] firstObject];
    CGFloat primaryHeight = [primaryScreen frame].size.height;

    // Convert y from top-left origin to bottom-left origin
    // In CG coords: y=0 is top of primary screen
    // In NS coords: y=0 is bottom of primary screen
    CGFloat nsY = primaryHeight - y - height;

    NSRect windowFrame = NSMakeRect(x, nsY, width, height);

    NSLog(@"PositionWindowOnDisplay: Setting frame to x=%f, y=%f (nsY=%f), w=%f, h=%f",
          x, y, nsY, width, height);

    // Make the window borderless and position it
    [targetWindow setStyleMask:NSWindowStyleMaskBorderless];
    [targetWindow setLevel:NSFloatingWindowLevel];
    [targetWindow setFrame:windowFrame display:YES animate:NO];
    [targetWindow makeKeyAndOrderFront:nil];
}
*/
import "C"

import (
	"log"
	"time"
)

// positionWindowOnDisplay moves the selector window to the specified display
func positionWindowOnDisplay(x, y, width, height float32) {
	// Small delay to ensure the window is created by Fyne
	time.Sleep(100 * time.Millisecond)
	log.Printf("Positioning window at (%f, %f) size %fx%f", x, y, width, height)
	C.PositionWindowOnDisplay(C.float(x), C.float(y), C.float(width), C.float(height))
}
