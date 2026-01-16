package editor

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/owenrumney/schnappit/internal/editor/tools"
	"github.com/owenrumney/schnappit/internal/output"
)

// Tool represents the currently selected annotation tool
type Tool int

const (
	ToolArrow Tool = iota
	ToolRectangle
)

// Editor represents the screenshot annotation editor
type Editor struct {
	window      fyne.Window
	screenshot  *image.RGBA
	annotations []tools.Annotation
	currentTool Tool
	toolColor   color.Color
	scaleFactor float64
	onSave      func(image.Image)
	onCopy      func(image.Image)

	// Drawing state
	drawing      bool
	startPoint   image.Point
	currentPoint image.Point
	imgCanvas    *canvas.Image
	overlay      *image.RGBA // For compositing final image
	preview      *image.RGBA // For live preview during drawing
}

// New creates a new editor with the given screenshot and scale factor
func New(app fyne.App, screenshot *image.RGBA, scaleFactor float64) *Editor {
	e := &Editor{
		screenshot:  screenshot,
		annotations: make([]tools.Annotation, 0),
		currentTool: ToolArrow,
		toolColor:   color.RGBA{R: 255, G: 0, B: 0, A: 255}, // Default red
		scaleFactor: scaleFactor,
	}

	e.window = app.NewWindow("Schnappit - Edit Screenshot")
	e.setupUI()

	return e
}

// setupUI creates the editor UI
func (e *Editor) setupUI() {
	// Initialize overlay and preview for compositing
	e.overlay = image.NewRGBA(e.screenshot.Bounds())
	e.preview = image.NewRGBA(e.screenshot.Bounds())
	e.refreshOverlay()

	// Screenshot display with overlay
	e.imgCanvas = canvas.NewImageFromImage(e.overlay)
	e.imgCanvas.FillMode = canvas.ImageFillStretch // Stretch to fill the logical size

	// Create a tappable container for mouse events
	drawArea := newDrawArea(e)

	// Toolbar
	toolbar := e.createToolbar()

	// Calculate logical size (pixel size / scale factor)
	bounds := e.screenshot.Bounds()
	logicalWidth := float32(float64(bounds.Dx()) / e.scaleFactor)
	logicalHeight := float32(float64(bounds.Dy()) / e.scaleFactor)

	// Set explicit size on the image canvas
	e.imgCanvas.SetMinSize(fyne.NewSize(logicalWidth, logicalHeight))

	// Main layout - wrap the canvas in the draw area for mouse events
	imageContainer := container.NewStack(e.imgCanvas, drawArea)
	content := container.NewBorder(toolbar, nil, nil, nil, imageContainer)
	e.window.SetContent(content)

	// Size window to fit screenshot at logical size (with some padding for toolbar)
	e.window.Resize(fyne.NewSize(
		logicalWidth,
		logicalHeight+50, // Extra space for toolbar
	))
}

// refreshOverlay redraws the overlay with the screenshot and all annotations
func (e *Editor) refreshOverlay() {
	// Copy screenshot to overlay
	draw.Draw(e.overlay, e.overlay.Bounds(), e.screenshot, image.Point{}, draw.Src)

	// Draw all annotations
	for _, ann := range e.annotations {
		ann.Draw(e.overlay)
	}
}

// updateCanvas refreshes the canvas display
func (e *Editor) updateCanvas() {
	e.refreshOverlay()
	e.imgCanvas.Image = e.overlay
	e.imgCanvas.Refresh()
}

// updatePreview refreshes the canvas with a preview of the current annotation being drawn
func (e *Editor) updatePreview() {
	if !e.drawing {
		return
	}

	// Copy the current overlay (with existing annotations) to preview
	draw.Draw(e.preview, e.preview.Bounds(), e.overlay, image.Point{}, draw.Src)

	// Draw the preview annotation
	strokeWidth := int(3 * e.scaleFactor)

	switch e.currentTool {
	case ToolArrow:
		previewArrow := tools.NewArrow(e.startPoint, e.currentPoint, e.toolColor, strokeWidth)
		previewArrow.Draw(e.preview)
	case ToolRectangle:
		rect := image.Rectangle{
			Min: image.Pt(min(e.startPoint.X, e.currentPoint.X), min(e.startPoint.Y, e.currentPoint.Y)),
			Max: image.Pt(max(e.startPoint.X, e.currentPoint.X), max(e.startPoint.Y, e.currentPoint.Y)),
		}
		previewRect := tools.NewRect(rect, e.toolColor, strokeWidth, false)
		previewRect.Draw(e.preview)
	}

	e.imgCanvas.Image = e.preview
	e.imgCanvas.Refresh()
}

// drawArea is a custom widget that handles mouse events for drawing
type drawArea struct {
	widget.BaseWidget
	editor *Editor
}

func newDrawArea(e *Editor) *drawArea {
	d := &drawArea{editor: e}
	d.ExtendBaseWidget(d)
	return d
}

func (d *drawArea) CreateRenderer() fyne.WidgetRenderer {
	return &drawAreaRenderer{}
}

func (d *drawArea) Tapped(ev *fyne.PointEvent) {}

func (d *drawArea) TappedSecondary(ev *fyne.PointEvent) {}

func (d *drawArea) MouseDown(ev *desktop.MouseEvent) {
	d.editor.drawing = true
	// Convert logical coordinates to pixel coordinates
	scale := d.editor.scaleFactor
	d.editor.startPoint = image.Pt(
		int(float64(ev.Position.X)*scale),
		int(float64(ev.Position.Y)*scale),
	)
	d.editor.currentPoint = d.editor.startPoint
}

func (d *drawArea) MouseUp(ev *desktop.MouseEvent) {
	// DragEnd handles finalization
}

// Dragged implements fyne.Draggable for live preview
func (d *drawArea) Dragged(ev *fyne.DragEvent) {
	if !d.editor.drawing {
		return
	}

	// Convert logical coordinates to pixel coordinates
	scale := d.editor.scaleFactor
	d.editor.currentPoint = image.Pt(
		int(float64(ev.Position.X)*scale),
		int(float64(ev.Position.Y)*scale),
	)

	// Update preview
	d.editor.updatePreview()
}

// DragEnd implements fyne.Draggable - finalizes the annotation
func (d *drawArea) DragEnd() {
	if !d.editor.drawing {
		return
	}
	d.editor.drawing = false

	// Scale stroke width for Retina displays
	scale := d.editor.scaleFactor
	strokeWidth := int(3 * scale)

	// Create annotation based on current tool
	var ann tools.Annotation
	switch d.editor.currentTool {
	case ToolArrow:
		ann = tools.NewArrow(d.editor.startPoint, d.editor.currentPoint, d.editor.toolColor, strokeWidth)
	case ToolRectangle:
		rect := image.Rectangle{
			Min: image.Pt(min(d.editor.startPoint.X, d.editor.currentPoint.X), min(d.editor.startPoint.Y, d.editor.currentPoint.Y)),
			Max: image.Pt(max(d.editor.startPoint.X, d.editor.currentPoint.X), max(d.editor.startPoint.Y, d.editor.currentPoint.Y)),
		}
		ann = tools.NewRect(rect, d.editor.toolColor, strokeWidth, false)
	}

	if ann != nil {
		d.editor.annotations = append(d.editor.annotations, ann)
		d.editor.updateCanvas()
	}
}

type drawAreaRenderer struct{}

func (r *drawAreaRenderer) Destroy()                             {}
func (r *drawAreaRenderer) Layout(size fyne.Size)                {}
func (r *drawAreaRenderer) MinSize() fyne.Size                   { return fyne.NewSize(0, 0) }
func (r *drawAreaRenderer) Objects() []fyne.CanvasObject         { return nil }
func (r *drawAreaRenderer) Refresh()                             {}

// createToolbar creates the annotation toolbar with icons
func (e *Editor) createToolbar() *fyne.Container {
	arrowBtn := widget.NewButtonWithIcon("", theme.MoveDownIcon(), func() {
		e.currentTool = ToolArrow
	})
	arrowBtn.Importance = widget.MediumImportance

	rectBtn := widget.NewButtonWithIcon("", theme.CheckButtonIcon(), func() {
		e.currentTool = ToolRectangle
	})
	rectBtn.Importance = widget.MediumImportance

	copyBtn := widget.NewButtonWithIcon("", theme.ContentCopyIcon(), func() {
		e.copyToClipboard()
	})
	copyBtn.Importance = widget.HighImportance

	saveBtn := widget.NewButtonWithIcon("", theme.DocumentSaveIcon(), func() {
		e.saveToFile()
	})
	saveBtn.Importance = widget.HighImportance

	closeBtn := widget.NewButtonWithIcon("", theme.CancelIcon(), func() {
		e.window.Close()
	})

	return container.NewHBox(
		arrowBtn,
		rectBtn,
		widget.NewSeparator(),
		copyBtn,
		saveBtn,
		closeBtn,
	)
}

// Show displays the editor window
func (e *Editor) Show() {
	e.window.Show()
}

// copyToClipboard copies the annotated screenshot to clipboard
func (e *Editor) copyToClipboard() {
	finalImg := e.renderFinal()
	if err := output.CopyToClipboard(finalImg); err != nil {
		dialog.ShowError(fmt.Errorf("Failed to copy to clipboard: %w", err), e.window)
		return
	}
	e.window.Close()
}

// saveToFile saves the annotated screenshot to a file
func (e *Editor) saveToFile() {
	finalImg := e.renderFinal()
	path, err := output.SaveToFile(finalImg)
	if err != nil {
		dialog.ShowError(fmt.Errorf("Failed to save file: %w", err), e.window)
		return
	}
	dialog.ShowInformation("Saved", fmt.Sprintf("Screenshot saved to:\n%s", path), e.window)
}

// renderFinal renders the screenshot with all annotations
func (e *Editor) renderFinal() image.Image {
	e.refreshOverlay()
	return e.overlay
}
