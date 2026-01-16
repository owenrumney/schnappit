package app

import (
	"image"
	"image/draw"
	"log"
	"sync/atomic"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/driver/desktop"

	"github.com/owenrumney/schnappit/internal/assets"
	"github.com/owenrumney/schnappit/internal/capture"
	"github.com/owenrumney/schnappit/internal/editor"
	"github.com/owenrumney/schnappit/internal/hotkey"
	"github.com/owenrumney/schnappit/internal/selector"
)

// App represents the main Schnappit application
type App struct {
	fyneApp   fyne.App
	capturing atomic.Bool
	shortcut  *hotkey.Shortcut
}

// New creates a new Schnappit application
func New() *App {
	return &App{
		fyneApp: app.New(),
	}
}

// Run starts the application
func (a *App) Run() error {
	hotkeyInfo := hotkey.GetConfiguredHotkey()

	// Set up system tray if supported (keeps app running without visible window)
	if desk, ok := a.fyneApp.(desktop.App); ok {
		// Set custom menu bar icon
		desk.SetSystemTrayIcon(assets.MenuBarIcon())

		menu := fyne.NewMenu("Schnappit",
			fyne.NewMenuItem("Capture Screenshot ("+hotkeyInfo+")", a.onCapture),
			fyne.NewMenuItemSeparator(),
			fyne.NewMenuItem("Quit", func() {
				a.fyneApp.Quit()
			}),
		)
		desk.SetSystemTrayMenu(menu)
	}

	// Register global hotkey after Fyne event loop starts (required on macOS)
	// Use a goroutine with delay to ensure the event loop is fully running
	a.fyneApp.Lifecycle().SetOnStarted(func() {
		// Hide dock icon now that Fyne is fully initialized
		hideDockIcon()

		go func() {
			// Small delay to ensure event loop is fully initialized
			time.Sleep(500 * time.Millisecond)

			shortcut, err := hotkey.New(a.onCapture)
			if err != nil {
				log.Printf("Warning: %v", err)
				log.Println("Continuing without global hotkey - use system tray menu instead")
			} else {
				a.shortcut = shortcut
				log.Printf("Schnappit running. Press %s or use system tray to capture.", hotkeyInfo)
			}
		}()
	})

	// Cleanup hotkey on app stop
	a.fyneApp.Lifecycle().SetOnStopped(func() {
		if a.shortcut != nil {
			a.shortcut.Unregister()
		}
	})

	// Run the Fyne application (this blocks)
	a.fyneApp.Run()

	return nil
}

// onCapture is called when the user triggers a screenshot capture
func (a *App) onCapture() {
	// Use atomic swap to prevent multiple concurrent captures
	if !a.capturing.CompareAndSwap(false, true) {
		return // Already capturing
	}

	log.Println("Capturing screen for region selection...")

	// Capture the full screen first for the selector background
	fullScreenshot, err := capture.CaptureDisplay(0)
	if err != nil {
		log.Printf("Failed to capture screenshot: %v", err)
		a.capturing.Store(false)
		return
	}

	// Get display bounds and scale factor for region selection
	displayBounds := capture.GetDisplayBounds(0)
	scaleFactor := capture.GetDisplayScaleFactor(0)

	log.Printf("Display bounds: %v, scale factor: %v", displayBounds, scaleFactor)

	// Show region selector with the captured screenshot as background
	sel := selector.New(a.fyneApp, displayBounds, scaleFactor, fullScreenshot,
		func(rect image.Rectangle) {
			// Region selected - crop from the already-captured screenshot
			a.openEditorWithRegion(fullScreenshot, rect, scaleFactor)
		},
		func() {
			// Cancelled
			a.capturing.Store(false)
			log.Println("Region selection cancelled")
		},
	)
	sel.Show()
}

// openEditorWithRegion crops the screenshot to the selected region and opens the editor
func (a *App) openEditorWithRegion(fullScreenshot *image.RGBA, rect image.Rectangle, scaleFactor float64) {
	defer func() { a.capturing.Store(false) }()

	log.Printf("Opening editor with region: %v", rect)
	log.Printf("Screenshot bounds: %v", fullScreenshot.Bounds())

	// Ensure rect is within screenshot bounds
	rect = rect.Intersect(fullScreenshot.Bounds())
	if rect.Empty() {
		log.Printf("Selection region is empty or out of bounds")
		return
	}

	log.Printf("Adjusted region: %v", rect)

	// Create cropped image using SubImage for efficiency and safety
	subImg := fullScreenshot.SubImage(rect).(*image.RGBA)

	// Copy to a new image with origin at (0,0)
	cropped := image.NewRGBA(image.Rect(0, 0, rect.Dx(), rect.Dy()))
	draw.Draw(cropped, cropped.Bounds(), subImg, rect.Min, draw.Src)

	// Open editor with the cropped image and scale factor
	ed := editor.New(a.fyneApp, cropped, scaleFactor)
	ed.Show()
}
