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

		menu := fyne.NewMenu("Schnappit",
			fyne.NewMenuItem("Capture Screenshot ("+hotkeyInfo+")", a.onCapture),
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

	log.Println("Capturing screen for region selection...")
	fullScreenshot, err := capture.CaptureDisplay(0)
	if err != nil {
		log.Printf("Failed to capture screenshot: %v", err)
		a.capturing.Store(false)
		return
	}

	displayBounds := capture.GetDisplayBounds(0)
	scaleFactor := capture.GetDisplayScaleFactor(0)

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
