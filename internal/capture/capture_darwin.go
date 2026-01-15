//go:build darwin

package capture

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework CoreGraphics -framework CoreFoundation -framework ScreenCaptureKit -framework AppKit

#import <CoreGraphics/CoreGraphics.h>
#import <ScreenCaptureKit/ScreenCaptureKit.h>
#import <AppKit/AppKit.h>

static int displayCount = 0;
static CGDirectDisplayID *displays = NULL;
static CGRect *displayBounds = NULL;

// Initialize and get display information using ScreenCaptureKit
void SCK_Initialize() {
    @autoreleasepool {
        [NSApplication sharedApplication];

        dispatch_semaphore_t semaphore = dispatch_semaphore_create(0);

        [SCShareableContent getShareableContentWithCompletionHandler:^(SCShareableContent *content, NSError *error) {
            if (error) {
                NSLog(@"Error getting shareable content: %@", error);
                dispatch_semaphore_signal(semaphore);
                return;
            }

            NSArray<SCDisplay *> *scDisplays = content.displays;
            displayCount = (int)[scDisplays count];

            if (displays) free(displays);
            if (displayBounds) free(displayBounds);

            displays = (CGDirectDisplayID *)malloc(sizeof(CGDirectDisplayID) * displayCount);
            displayBounds = (CGRect *)malloc(sizeof(CGRect) * displayCount);

            for (int i = 0; i < displayCount; i++) {
                SCDisplay *display = scDisplays[i];
                displays[i] = display.displayID;
                displayBounds[i] = CGRectMake(0, 0, display.width, display.height);
            }

            dispatch_semaphore_signal(semaphore);
        }];

        dispatch_semaphore_wait(semaphore, DISPATCH_TIME_FOREVER);
    }
}

int SCK_GetDisplayCount() {
    if (displayCount == 0) {
        SCK_Initialize();
    }
    return displayCount;
}

void SCK_GetDisplayBounds(int index, int *x, int *y, int *width, int *height) {
    if (displayCount == 0) {
        SCK_Initialize();
    }
    if (index >= 0 && index < displayCount) {
        CGRect bounds = CGDisplayBounds(displays[index]);
        *x = (int)bounds.origin.x;
        *y = (int)bounds.origin.y;
        // Use pixel dimensions for Retina/HiDPI displays
        CGDisplayModeRef mode = CGDisplayCopyDisplayMode(displays[index]);
        if (mode) {
            *width = (int)CGDisplayModeGetPixelWidth(mode);
            *height = (int)CGDisplayModeGetPixelHeight(mode);
            CGDisplayModeRelease(mode);
        } else {
            *width = (int)bounds.size.width;
            *height = (int)bounds.size.height;
        }
    }
}

// Get the scale factor for a display (for coordinate conversion)
float SCK_GetDisplayScaleFactor(int index) {
    if (displayCount == 0) {
        SCK_Initialize();
    }
    if (index >= 0 && index < displayCount) {
        CGRect bounds = CGDisplayBounds(displays[index]);
        CGDisplayModeRef mode = CGDisplayCopyDisplayMode(displays[index]);
        if (mode) {
            float scale = (float)CGDisplayModeGetPixelWidth(mode) / bounds.size.width;
            CGDisplayModeRelease(mode);
            return scale;
        }
    }
    return 1.0;
}

// Capture a region of the screen
void SCK_CaptureRect(int displayIndex, int x, int y, int width, int height, void *buffer) {
    @autoreleasepool {
        if (displayCount == 0) {
            SCK_Initialize();
        }

        if (displayIndex < 0 || displayIndex >= displayCount) {
            return;
        }

        dispatch_semaphore_t semaphore = dispatch_semaphore_create(0);

        [SCShareableContent getShareableContentWithCompletionHandler:^(SCShareableContent *content, NSError *error) {
            if (error) {
                NSLog(@"Error: %@", error);
                dispatch_semaphore_signal(semaphore);
                return;
            }

            SCDisplay *targetDisplay = nil;
            for (SCDisplay *display in content.displays) {
                if (display.displayID == displays[displayIndex]) {
                    targetDisplay = display;
                    break;
                }
            }

            if (!targetDisplay) {
                dispatch_semaphore_signal(semaphore);
                return;
            }

            SCContentFilter *filter = [[SCContentFilter alloc] initWithDisplay:targetDisplay excludingWindows:@[]];
            SCStreamConfiguration *config = [[SCStreamConfiguration alloc] init];
            config.width = width;
            config.height = height;
            config.sourceRect = CGRectMake(x, y, width, height);
            config.showsCursor = NO;
            config.pixelFormat = kCVPixelFormatType_32BGRA;

            [SCScreenshotManager captureImageWithFilter:filter
                                          configuration:config
                                      completionHandler:^(CGImageRef image, NSError *error) {
                if (error || !image) {
                    NSLog(@"Capture error: %@", error);
                    dispatch_semaphore_signal(semaphore);
                    return;
                }

                size_t imgWidth = CGImageGetWidth(image);
                size_t imgHeight = CGImageGetHeight(image);

                CGColorSpaceRef colorSpace = CGColorSpaceCreateDeviceRGB();
                CGContextRef context = CGBitmapContextCreate(
                    buffer,
                    imgWidth,
                    imgHeight,
                    8,
                    imgWidth * 4,
                    colorSpace,
                    kCGImageAlphaPremultipliedLast | kCGBitmapByteOrder32Big
                );

                CGContextDrawImage(context, CGRectMake(0, 0, imgWidth, imgHeight), image);

                CGContextRelease(context);
                CGColorSpaceRelease(colorSpace);

                dispatch_semaphore_signal(semaphore);
            }];
        }];

        dispatch_semaphore_wait(semaphore, DISPATCH_TIME_FOREVER);
    }
}
*/
import "C"

import (
	"fmt"
	"image"
	"unsafe"
)

// NumDisplays returns the number of active displays
func NumDisplays() int {
	return int(C.SCK_GetDisplayCount())
}

// GetDisplayScaleFactor returns the scale factor for Retina/HiDPI displays
func GetDisplayScaleFactor(displayIndex int) float64 {
	return float64(C.SCK_GetDisplayScaleFactor(C.int(displayIndex)))
}

// GetDisplayBounds returns the bounds of the display at the given index
func GetDisplayBounds(displayIndex int) image.Rectangle {
	var x, y, width, height C.int
	C.SCK_GetDisplayBounds(C.int(displayIndex), &x, &y, &width, &height)
	return image.Rect(int(x), int(y), int(x)+int(width), int(y)+int(height))
}

// CaptureDisplay captures the entire display at the given index
func CaptureDisplay(displayIndex int) (*image.RGBA, error) {
	bounds := GetDisplayBounds(displayIndex)
	return CaptureRect(displayIndex, bounds)
}

// CaptureRect captures a rectangular region from the specified display
func CaptureRect(displayIndex int, rect image.Rectangle) (*image.RGBA, error) {
	width := rect.Dx()
	height := rect.Dy()

	if width <= 0 || height <= 0 {
		return nil, fmt.Errorf("invalid capture dimensions: %dx%d", width, height)
	}

	img := image.NewRGBA(image.Rect(0, 0, width, height))

	C.SCK_CaptureRect(
		C.int(displayIndex),
		C.int(rect.Min.X),
		C.int(rect.Min.Y),
		C.int(width),
		C.int(height),
		unsafe.Pointer(&img.Pix[0]),
	)

	return img, nil
}

// CaptureRegion captures a region across all displays (for multi-monitor setups)
func CaptureRegion(rect image.Rectangle) (*image.RGBA, error) {
	// For MVP, capture from primary display (index 0)
	// TODO: Handle multi-monitor region selection
	return CaptureRect(0, rect)
}
