package selector

import (
	"image"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

// Handle positions
type HandlePos int

const (
	HandleNone HandlePos = iota
	HandleTopLeft
	HandleTopRight
	HandleBottomLeft
	HandleBottomRight
	HandleTop
	HandleBottom
	HandleLeft
	HandleRight
	HandleMove // For moving the entire selection
)

const handleSize = 8

// Selector represents the region selection overlay
type Selector struct {
	window        fyne.Window
	onSelect      func(image.Rectangle)
	onCancel      func()
	scaleFactor   float64

	// Selection state
	hasSelection  bool
	selectionMin  fyne.Position
	selectionMax  fyne.Position

	// Drag state
	dragging      bool
	dragHandle    HandlePos
	dragStart     fyne.Position
	dragSelMin    fyne.Position
	dragSelMax    fyne.Position

	// Background screenshot
	screenshot *image.RGBA

	// UI elements
	topDim        *canvas.Rectangle
	bottomDim     *canvas.Rectangle
	leftDim       *canvas.Rectangle
	rightDim      *canvas.Rectangle
	selectionRect *canvas.Rectangle

	// Resize handles
	handles       []*canvas.Rectangle

	// Instructions
	instructions  *canvas.Text

	screenWidth  float32
	screenHeight float32
}

// New creates a new region selector with a pre-captured screenshot as background
func New(app fyne.App, displayBounds image.Rectangle, scaleFactor float64, screenshot *image.RGBA, onSelect func(image.Rectangle), onCancel func()) *Selector {
	// Use logical coordinates for display
	screenWidth := float32(displayBounds.Dx()) / float32(scaleFactor)
	screenHeight := float32(displayBounds.Dy()) / float32(scaleFactor)

	s := &Selector{
		onSelect:     onSelect,
		onCancel:     onCancel,
		scaleFactor:  scaleFactor,
		screenshot:   screenshot,
		screenWidth:  screenWidth,
		screenHeight: screenHeight,
		handles:      make([]*canvas.Rectangle, 8),
	}

	s.window = app.NewWindow("Select Region")
	s.setupUI()

	return s
}

// setupUI creates the selection overlay UI
func (s *Selector) setupUI() {
	bgImage := canvas.NewImageFromImage(s.screenshot)
	bgImage.FillMode = canvas.ImageFillStretch
	bgImage.Resize(fyne.NewSize(s.screenWidth, s.screenHeight))
	bgImage.Move(fyne.NewPos(0, 0))

	dimColor := color.NRGBA{R: 0, G: 0, B: 0, A: 120}

	s.topDim = canvas.NewRectangle(dimColor)
	s.bottomDim = canvas.NewRectangle(dimColor)
	s.leftDim = canvas.NewRectangle(dimColor)
	s.rightDim = canvas.NewRectangle(dimColor)

	s.topDim.Resize(fyne.NewSize(s.screenWidth, s.screenHeight))
	s.topDim.Move(fyne.NewPos(0, 0))
	s.bottomDim.Resize(fyne.NewSize(0, 0))
	s.leftDim.Resize(fyne.NewSize(0, 0))
	s.rightDim.Resize(fyne.NewSize(0, 0))

	s.selectionRect = canvas.NewRectangle(color.Transparent)
	s.selectionRect.StrokeColor = color.NRGBA{R: 0, G: 120, B: 215, A: 255}
	s.selectionRect.StrokeWidth = 2
	s.selectionRect.Resize(fyne.NewSize(0, 0))

	handleColor := color.NRGBA{R: 255, G: 255, B: 255, A: 255}
	handleBorder := color.NRGBA{R: 0, G: 120, B: 215, A: 255}
	for i := 0; i < 8; i++ {
		h := canvas.NewRectangle(handleColor)
		h.StrokeColor = handleBorder
		h.StrokeWidth = 1
		h.Resize(fyne.NewSize(handleSize, handleSize))
		h.Hide()
		s.handles[i] = h
	}

	s.instructions = canvas.NewText("Click and drag to select region. Press Enter to capture, Escape to cancel.", color.White)
	s.instructions.TextSize = 14

	instructionsBg := canvas.NewRectangle(color.NRGBA{R: 0, G: 0, B: 0, A: 180})
	instructionsBg.Resize(fyne.NewSize(520, 30))
	instructionsBg.Move(fyne.NewPos(15, 15))
	s.instructions.Move(fyne.NewPos(20, 20))

	mouseArea := newMouseArea(s)
	mouseArea.Resize(fyne.NewSize(s.screenWidth, s.screenHeight))

	content := container.NewWithoutLayout(
		bgImage,
		s.topDim,
		s.bottomDim,
		s.leftDim,
		s.rightDim,
		s.selectionRect,
	)

	// Add handles
	for _, h := range s.handles {
		content.Add(h)
	}

	content.Add(instructionsBg)
	content.Add(s.instructions)
	content.Add(mouseArea)

	s.window.SetContent(content)
	s.window.SetFullScreen(true)

	s.window.Canvas().SetOnTypedKey(func(key *fyne.KeyEvent) {
		switch key.Name {
		case fyne.KeyEscape:
			s.window.Close()
			if s.onCancel != nil {
				s.onCancel()
			}
		case fyne.KeyReturn, fyne.KeyEnter:
			if s.hasSelection {
				s.confirmSelection()
			}
		}
	})
}

// confirmSelection finalizes the selection and calls the callback
func (s *Selector) confirmSelection() {
	if !s.hasSelection {
		return
	}

	scale := s.scaleFactor
	minX, minY, maxX, maxY := s.normalizedBounds()

	rect := image.Rect(
		int(float64(minX)*scale),
		int(float64(minY)*scale),
		int(float64(maxX)*scale),
		int(float64(maxY)*scale),
	)

	s.window.Close()
	if s.onSelect != nil {
		s.onSelect(rect)
	}
}

// updateSelection updates the visual selection feedback
func (s *Selector) updateSelection() {
	if !s.hasSelection {
		return
	}

	minX, minY, maxX, maxY := s.normalizedBounds()
	selWidth := maxX - minX
	selHeight := maxY - minY

	s.topDim.Move(fyne.NewPos(0, 0))
	s.topDim.Resize(fyne.NewSize(s.screenWidth, minY))

	s.bottomDim.Move(fyne.NewPos(0, maxY))
	s.bottomDim.Resize(fyne.NewSize(s.screenWidth, s.screenHeight-maxY))

	s.leftDim.Move(fyne.NewPos(0, minY))
	s.leftDim.Resize(fyne.NewSize(minX, selHeight))

	s.rightDim.Move(fyne.NewPos(maxX, minY))
	s.rightDim.Resize(fyne.NewSize(s.screenWidth-maxX, selHeight))

	s.selectionRect.Move(fyne.NewPos(minX, minY))
	s.selectionRect.Resize(fyne.NewSize(selWidth, selHeight))

	halfHandle := float32(handleSize / 2)

	s.handles[0].Move(fyne.NewPos(minX-halfHandle, minY-halfHandle))
	s.handles[1].Move(fyne.NewPos(maxX-halfHandle, minY-halfHandle))
	s.handles[2].Move(fyne.NewPos(minX-halfHandle, maxY-halfHandle))
	s.handles[3].Move(fyne.NewPos(maxX-halfHandle, maxY-halfHandle))

	midX := minX + selWidth/2
	midY := minY + selHeight/2
	s.handles[4].Move(fyne.NewPos(midX-halfHandle, minY-halfHandle))
	s.handles[5].Move(fyne.NewPos(midX-halfHandle, maxY-halfHandle))
	s.handles[6].Move(fyne.NewPos(minX-halfHandle, midY-halfHandle))
	s.handles[7].Move(fyne.NewPos(maxX-halfHandle, midY-halfHandle))

	for _, h := range s.handles {
		h.Show()
		h.Refresh()
	}

	s.topDim.Refresh()
	s.bottomDim.Refresh()
	s.leftDim.Refresh()
	s.rightDim.Refresh()
	s.selectionRect.Refresh()

	s.instructions.Text = "Drag handles to resize. Press Enter to capture, Escape to cancel."
	s.instructions.Refresh()
}

// hitTestHandle checks if a position is over a handle
func (s *Selector) hitTestHandle(pos fyne.Position) HandlePos {
	if !s.hasSelection {
		return HandleNone
	}

	minX, minY, maxX, maxY := s.normalizedBounds()
	midX := (minX + maxX) / 2
	midY := (minY + maxY) / 2

	hitRadius := float32(handleSize)

	if abs32(pos.X-minX) < hitRadius && abs32(pos.Y-minY) < hitRadius {
		return HandleTopLeft
	}
	if abs32(pos.X-maxX) < hitRadius && abs32(pos.Y-minY) < hitRadius {
		return HandleTopRight
	}
	if abs32(pos.X-minX) < hitRadius && abs32(pos.Y-maxY) < hitRadius {
		return HandleBottomLeft
	}
	if abs32(pos.X-maxX) < hitRadius && abs32(pos.Y-maxY) < hitRadius {
		return HandleBottomRight
	}

	if abs32(pos.X-midX) < hitRadius && abs32(pos.Y-minY) < hitRadius {
		return HandleTop
	}
	if abs32(pos.X-midX) < hitRadius && abs32(pos.Y-maxY) < hitRadius {
		return HandleBottom
	}
	if abs32(pos.X-minX) < hitRadius && abs32(pos.Y-midY) < hitRadius {
		return HandleLeft
	}
	if abs32(pos.X-maxX) < hitRadius && abs32(pos.Y-midY) < hitRadius {
		return HandleRight
	}

	if pos.X >= minX && pos.X <= maxX && pos.Y >= minY && pos.Y <= maxY {
		return HandleMove
	}

	return HandleNone
}

// normalizedBounds returns the selection bounds with min/max properly ordered
func (s *Selector) normalizedBounds() (minX, minY, maxX, maxY float32) {
	minX = min(s.selectionMin.X, s.selectionMax.X)
	minY = min(s.selectionMin.Y, s.selectionMax.Y)
	maxX = max(s.selectionMin.X, s.selectionMax.X)
	maxY = max(s.selectionMin.Y, s.selectionMax.Y)
	return
}

// Show displays the selector overlay
func (s *Selector) Show() {
	s.window.Show()
}

// Close closes the selector window
func (s *Selector) Close() {
	s.window.Close()
}

// mouseArea handles mouse events for region selection
type mouseArea struct {
	widget.BaseWidget
	selector *Selector
}

func newMouseArea(s *Selector) *mouseArea {
	m := &mouseArea{selector: s}
	m.ExtendBaseWidget(m)
	return m
}

func (m *mouseArea) CreateRenderer() fyne.WidgetRenderer {
	return &mouseAreaRenderer{}
}

func (m *mouseArea) Tapped(ev *fyne.PointEvent) {}

func (m *mouseArea) TappedSecondary(ev *fyne.PointEvent) {
	m.selector.window.Close()
	if m.selector.onCancel != nil {
		m.selector.onCancel()
	}
}

func (m *mouseArea) Dragged(ev *fyne.DragEvent) {
	if !m.selector.dragging {
		return
	}

	s := m.selector
	dx := ev.Position.X - s.dragStart.X
	dy := ev.Position.Y - s.dragStart.Y

	switch s.dragHandle {
	case HandleNone:
		// Creating new selection
		s.selectionMax = ev.Position

	case HandleTopLeft:
		s.selectionMin = fyne.NewPos(s.dragSelMin.X+dx, s.dragSelMin.Y+dy)

	case HandleTopRight:
		s.selectionMax = fyne.NewPos(s.dragSelMax.X+dx, s.dragSelMax.Y)
		s.selectionMin = fyne.NewPos(s.dragSelMin.X, s.dragSelMin.Y+dy)

	case HandleBottomLeft:
		s.selectionMin = fyne.NewPos(s.dragSelMin.X+dx, s.dragSelMin.Y)
		s.selectionMax = fyne.NewPos(s.dragSelMax.X, s.dragSelMax.Y+dy)

	case HandleBottomRight:
		s.selectionMax = fyne.NewPos(s.dragSelMax.X+dx, s.dragSelMax.Y+dy)

	case HandleTop:
		s.selectionMin = fyne.NewPos(s.dragSelMin.X, s.dragSelMin.Y+dy)

	case HandleBottom:
		s.selectionMax = fyne.NewPos(s.dragSelMax.X, s.dragSelMax.Y+dy)

	case HandleLeft:
		s.selectionMin = fyne.NewPos(s.dragSelMin.X+dx, s.dragSelMin.Y)

	case HandleRight:
		s.selectionMax = fyne.NewPos(s.dragSelMax.X+dx, s.dragSelMax.Y)

	case HandleMove:
		s.selectionMin = fyne.NewPos(s.dragSelMin.X+dx, s.dragSelMin.Y+dy)
		s.selectionMax = fyne.NewPos(s.dragSelMax.X+dx, s.dragSelMax.Y+dy)
	}

	s.updateSelection()
}

func (m *mouseArea) DragEnd() {
	m.selector.dragging = false
}

func (m *mouseArea) MouseDown(ev *desktop.MouseEvent) {
	s := m.selector

	handle := s.hitTestHandle(ev.Position)

	if handle != HandleNone {
		s.dragging = true
		s.dragHandle = handle
		s.dragStart = ev.Position
		s.dragSelMin = s.selectionMin
		s.dragSelMax = s.selectionMax
	} else {
		s.dragging = true
		s.dragHandle = HandleNone
		s.hasSelection = true
		s.selectionMin = ev.Position
		s.selectionMax = ev.Position
		s.dragStart = ev.Position

		for _, h := range s.handles {
			h.Hide()
		}
	}
}

func (m *mouseArea) MouseUp(ev *desktop.MouseEvent) {
	// DragEnd handles this
}

type mouseAreaRenderer struct{}

func (r *mouseAreaRenderer) Destroy()                     {}
func (r *mouseAreaRenderer) Layout(size fyne.Size)        {}
func (r *mouseAreaRenderer) MinSize() fyne.Size           { return fyne.NewSize(0, 0) }
func (r *mouseAreaRenderer) Objects() []fyne.CanvasObject { return nil }
func (r *mouseAreaRenderer) Refresh()                     {}

func abs32(x float32) float32 {
	if x < 0 {
		return -x
	}
	return x
}
