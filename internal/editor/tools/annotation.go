package tools

import (
	"image"
	"image/color"
	"image/draw"
	"math"
)

// Annotation represents a drawable annotation on the screenshot
type Annotation interface {
	// Draw renders the annotation onto the image
	Draw(img *image.RGBA)

	// Bounds returns the bounding box of the annotation
	Bounds() image.Rectangle

	// Contains returns true if the point is within the annotation
	Contains(x, y int) bool
}

// BaseAnnotation contains common annotation properties
type BaseAnnotation struct {
	Color       color.Color
	StrokeWidth int
}

// ArrowAnnotation represents an arrow annotation
type ArrowAnnotation struct {
	BaseAnnotation
	Start image.Point
	End   image.Point
}

// NewArrow creates a new arrow annotation
func NewArrow(start, end image.Point, c color.Color, strokeWidth int) *ArrowAnnotation {
	return &ArrowAnnotation{
		BaseAnnotation: BaseAnnotation{Color: c, StrokeWidth: strokeWidth},
		Start:          start,
		End:            end,
	}
}

// Draw renders the arrow onto the image
func (a *ArrowAnnotation) Draw(img *image.RGBA) {
	drawLine(img, a.Start, a.End, a.Color, a.StrokeWidth)
	drawArrowHead(img, a.Start, a.End, a.Color, a.StrokeWidth)
}

// Bounds returns the bounding box of the arrow
func (a *ArrowAnnotation) Bounds() image.Rectangle {
	minX := min(a.Start.X, a.End.X) - a.StrokeWidth
	minY := min(a.Start.Y, a.End.Y) - a.StrokeWidth
	maxX := max(a.Start.X, a.End.X) + a.StrokeWidth
	maxY := max(a.Start.Y, a.End.Y) + a.StrokeWidth
	return image.Rect(minX, minY, maxX, maxY)
}

// Contains returns true if the point is near the arrow line
func (a *ArrowAnnotation) Contains(x, y int) bool {
	return image.Pt(x, y).In(a.Bounds())
}

// RectAnnotation represents a rectangle annotation
type RectAnnotation struct {
	BaseAnnotation
	Rect   image.Rectangle
	Filled bool
}

// NewRect creates a new rectangle annotation
func NewRect(rect image.Rectangle, c color.Color, strokeWidth int, filled bool) *RectAnnotation {
	return &RectAnnotation{
		BaseAnnotation: BaseAnnotation{Color: c, StrokeWidth: strokeWidth},
		Rect:           rect,
		Filled:         filled,
	}
}

// Draw renders the rectangle onto the image
func (r *RectAnnotation) Draw(img *image.RGBA) {
	if r.Filled {
		draw.Draw(img, r.Rect, &image.Uniform{r.Color}, image.Point{}, draw.Over)
	} else {
		drawRectOutline(img, r.Rect, r.Color, r.StrokeWidth)
	}
}

// Bounds returns the bounding box of the rectangle
func (r *RectAnnotation) Bounds() image.Rectangle {
	return r.Rect.Inset(-r.StrokeWidth)
}

// Contains returns true if the point is within the rectangle
func (r *RectAnnotation) Contains(x, y int) bool {
	return image.Pt(x, y).In(r.Bounds())
}

func drawLine(img *image.RGBA, start, end image.Point, c color.Color, width int) {
	dx := math.Abs(float64(end.X - start.X))
	dy := math.Abs(float64(end.Y - start.Y))
	sx, sy := 1, 1
	if start.X >= end.X {
		sx = -1
	}
	if start.Y >= end.Y {
		sy = -1
	}
	err := dx - dy

	x, y := start.X, start.Y
	for {
		for i := -width / 2; i <= width/2; i++ {
			for j := -width / 2; j <= width/2; j++ {
				if x+i >= 0 && y+j >= 0 && x+i < img.Bounds().Dx() && y+j < img.Bounds().Dy() {
					img.Set(x+i, y+j, c)
				}
			}
		}

		if x == end.X && y == end.Y {
			break
		}
		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x += sx
		}
		if e2 < dx {
			err += dx
			y += sy
		}
	}
}

func drawArrowHead(img *image.RGBA, start, end image.Point, c color.Color, strokeWidth int) {
	headLength := strokeWidth * 5
	headWidth := strokeWidth * 3

	dx := float64(end.X - start.X)
	dy := float64(end.Y - start.Y)
	length := math.Sqrt(dx*dx + dy*dy)
	if length < 1 {
		return
	}

	dx, dy = dx/length, dy/length

	baseX := float64(end.X) - dx*float64(headLength)
	baseY := float64(end.Y) - dy*float64(headLength)

	perpX, perpY := -dy, dx

	left := image.Pt(int(baseX+perpX*float64(headWidth)), int(baseY+perpY*float64(headWidth)))
	right := image.Pt(int(baseX-perpX*float64(headWidth)), int(baseY-perpY*float64(headWidth)))

	drawLine(img, end, left, c, strokeWidth)
	drawLine(img, end, right, c, strokeWidth)
}

func drawRectOutline(img *image.RGBA, rect image.Rectangle, c color.Color, width int) {
	drawLine(img, image.Pt(rect.Min.X, rect.Min.Y), image.Pt(rect.Max.X, rect.Min.Y), c, width)
	drawLine(img, image.Pt(rect.Min.X, rect.Max.Y), image.Pt(rect.Max.X, rect.Max.Y), c, width)
	drawLine(img, image.Pt(rect.Min.X, rect.Min.Y), image.Pt(rect.Min.X, rect.Max.Y), c, width)
	drawLine(img, image.Pt(rect.Max.X, rect.Min.Y), image.Pt(rect.Max.X, rect.Max.Y), c, width)
}
