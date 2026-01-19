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

	if desk, ok := a.fyneApp.(desktop.App); ok {
		desk.SetSystemTrayIcon(assets.MenuBarIcon())

		// Create login item toggle
		loginItemLabel := "Start on Login"
		if IsLoginItemEnabled() {
			loginItemLabel = "✓ Start on Login"
		}
		loginItem := fyne.NewMenuItem(loginItemLabel, nil)
		loginItem.Action = func() {
			enabled := IsLoginItemEnabled()
			if err := SetLoginItemEnabled(!enabled); err != nil {
				log.Printf("Failed to toggle login item: %v", err)
				return
			}
			// Update menu label
			if !enabled {
				loginItem.Label = "✓ Start on Login"
			} else {
				loginItem.Label = "Start on Login"
			}
			desk.SetSystemTrayMenu(fyne.NewMenu("Schnappit",
				fyne.NewMenuItem("Capture Screenshot ("+hotkeyInfo+")", a.onCapture),
				fyne.NewMenuItemSeparator(),
				loginItem,
				fyne.NewMenuItemSeparator(),
				fyne.NewMenuItem("Quit", func() {
					a.fyneApp.Quit()
				}),
			))
		}

		menu := fyne.NewMenu("Schnappit",
			fyne.NewMenuItem("Capture Screenshot ("+hotkeyInfo+")", a.onCapture),
			fyne.NewMenuItemSeparator(),
			loginItem,
			fyne.NewMenuItemSeparator(),
			fyne.NewMenuItem("Quit", func() {
				a.fyneApp.Quit()
			}),
		)
		desk.SetSystemTrayMenu(menu)
	}

	a.fyneApp.Lifecycle().SetOnStarted(func() {
		hideDockIcon()

		go func() {
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

	a.fyneApp.Lifecycle().SetOnStopped(func() {
		if a.shortcut != nil {
			a.shortcut.Unregister()
		}
	})

	a.fyneApp.Run()

	return nil
}

// onCapture is called when the user triggers a screenshot capture
func (a *App) onCapture() {
	if !a.capturing.CompareAndSwap(false, true) {
		return
	}

	// Detect which display contains the mouse cursor
	displayIndex := capture.GetDisplayAtMousePosition()

	log.Printf("Capturing display %d (where mouse cursor is located)...", displayIndex)
	fullScreenshot, err := capture.CaptureDisplay(displayIndex)
	if err != nil {
		log.Printf("Failed to capture screenshot: %v", err)
		a.capturing.Store(false)
		return
	}

	displayBounds := capture.GetDisplayBounds(displayIndex)
	scaleFactor := capture.GetDisplayScaleFactor(displayIndex)

	log.Printf("Display bounds: %v, scale factor: %v", displayBounds, scaleFactor)

	sel := selector.New(a.fyneApp, displayBounds, scaleFactor, fullScreenshot,
		func(rect image.Rectangle) {
			a.openEditorWithRegion(fullScreenshot, rect, scaleFactor)
		},
		func() {
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

	rect = rect.Intersect(fullScreenshot.Bounds())
	if rect.Empty() {
		log.Printf("Selection region is empty or out of bounds")
		return
	}

	log.Printf("Adjusted region: %v", rect)

	subImg, ok := fullScreenshot.SubImage(rect).(*image.RGBA)
	if !ok {
		log.Printf("Failed to get subimage")
		return
	}

	cropped := image.NewRGBA(image.Rect(0, 0, rect.Dx(), rect.Dy()))
	draw.Draw(cropped, cropped.Bounds(), subImg, rect.Min, draw.Src)

	ed := editor.New(a.fyneApp, cropped, scaleFactor)
	ed.Show()
}
