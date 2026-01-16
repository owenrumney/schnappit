package tools

import (
	"image"
	"image/color"
	"testing"
)

func TestNewArrow(t *testing.T) {
	start := image.Pt(10, 20)
	end := image.Pt(100, 200)
	c := color.RGBA{R: 255, G: 0, B: 0, A: 255}
	strokeWidth := 3

	arrow := NewArrow(start, end, c, strokeWidth)

	if arrow.Start != start {
		t.Errorf("Expected start %v, got %v", start, arrow.Start)
	}
	if arrow.End != end {
		t.Errorf("Expected end %v, got %v", end, arrow.End)
	}
	if arrow.Color != c {
		t.Errorf("Expected color %v, got %v", c, arrow.Color)
	}
	if arrow.StrokeWidth != strokeWidth {
		t.Errorf("Expected strokeWidth %d, got %d", strokeWidth, arrow.StrokeWidth)
	}
}

func TestArrowBounds(t *testing.T) {
	tests := []struct {
		name        string
		start       image.Point
		end         image.Point
		strokeWidth int
		wantMinX    int
		wantMinY    int
		wantMaxX    int
		wantMaxY    int
	}{
		{
			name:        "normal arrow",
			start:       image.Pt(10, 20),
			end:         image.Pt(100, 200),
			strokeWidth: 3,
			wantMinX:    7,  // 10 - 3
			wantMinY:    17, // 20 - 3
			wantMaxX:    103,
			wantMaxY:    203,
		},
		{
			name:        "reversed arrow",
			start:       image.Pt(100, 200),
			end:         image.Pt(10, 20),
			strokeWidth: 5,
			wantMinX:    5,   // 10 - 5
			wantMinY:    15,  // 20 - 5
			wantMaxX:    105, // 100 + 5
			wantMaxY:    205, // 200 + 5
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			arrow := NewArrow(tt.start, tt.end, color.Black, tt.strokeWidth)
			bounds := arrow.Bounds()

			if bounds.Min.X != tt.wantMinX {
				t.Errorf("Bounds().Min.X = %d, want %d", bounds.Min.X, tt.wantMinX)
			}
			if bounds.Min.Y != tt.wantMinY {
				t.Errorf("Bounds().Min.Y = %d, want %d", bounds.Min.Y, tt.wantMinY)
			}
			if bounds.Max.X != tt.wantMaxX {
				t.Errorf("Bounds().Max.X = %d, want %d", bounds.Max.X, tt.wantMaxX)
			}
			if bounds.Max.Y != tt.wantMaxY {
				t.Errorf("Bounds().Max.Y = %d, want %d", bounds.Max.Y, tt.wantMaxY)
			}
		})
	}
}

func TestArrowContains(t *testing.T) {
	arrow := NewArrow(image.Pt(10, 10), image.Pt(100, 100), color.Black, 5)

	tests := []struct {
		name string
		x    int
		y    int
		want bool
	}{
		{"inside bounds", 50, 50, true},
		{"at start", 10, 10, true},
		{"at end", 100, 100, true},
		{"outside left", 0, 50, false},
		{"outside right", 200, 50, false},
		{"outside top", 50, 0, false},
		{"outside bottom", 50, 200, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := arrow.Contains(tt.x, tt.y); got != tt.want {
				t.Errorf("Contains(%d, %d) = %v, want %v", tt.x, tt.y, got, tt.want)
			}
		})
	}
}

func TestNewRect(t *testing.T) {
	rect := image.Rect(10, 20, 100, 200)
	c := color.RGBA{R: 0, G: 255, B: 0, A: 255}
	strokeWidth := 2
	filled := true

	annotation := NewRect(rect, c, strokeWidth, filled)

	if annotation.Rect != rect {
		t.Errorf("Expected rect %v, got %v", rect, annotation.Rect)
	}
	if annotation.Color != c {
		t.Errorf("Expected color %v, got %v", c, annotation.Color)
	}
	if annotation.StrokeWidth != strokeWidth {
		t.Errorf("Expected strokeWidth %d, got %d", strokeWidth, annotation.StrokeWidth)
	}
	if annotation.Filled != filled {
		t.Errorf("Expected filled %v, got %v", filled, annotation.Filled)
	}
}

func TestRectBounds(t *testing.T) {
	rect := image.Rect(10, 20, 100, 200)
	strokeWidth := 5

	annotation := NewRect(rect, color.Black, strokeWidth, false)
	bounds := annotation.Bounds()

	// Bounds should be inset by -strokeWidth (expanded)
	if bounds.Min.X != 5 { // 10 - 5
		t.Errorf("Bounds().Min.X = %d, want 5", bounds.Min.X)
	}
	if bounds.Min.Y != 15 { // 20 - 5
		t.Errorf("Bounds().Min.Y = %d, want 15", bounds.Min.Y)
	}
	if bounds.Max.X != 105 { // 100 + 5
		t.Errorf("Bounds().Max.X = %d, want 105", bounds.Max.X)
	}
	if bounds.Max.Y != 205 { // 200 + 5
		t.Errorf("Bounds().Max.Y = %d, want 205", bounds.Max.Y)
	}
}

func TestRectContains(t *testing.T) {
	rect := NewRect(image.Rect(10, 10, 100, 100), color.Black, 5, false)

	tests := []struct {
		name string
		x    int
		y    int
		want bool
	}{
		{"inside", 50, 50, true},
		{"at corner", 10, 10, true},
		{"in stroke area", 7, 50, true}, // Within expanded bounds
		{"outside", 0, 0, false},
		{"outside right", 200, 50, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := rect.Contains(tt.x, tt.y); got != tt.want {
				t.Errorf("Contains(%d, %d) = %v, want %v", tt.x, tt.y, got, tt.want)
			}
		})
	}
}

func TestArrowDraw(t *testing.T) {
	// Create a test image
	img := image.NewRGBA(image.Rect(0, 0, 200, 200))

	// Create and draw an arrow
	arrow := NewArrow(image.Pt(10, 10), image.Pt(100, 100), color.RGBA{R: 255, A: 255}, 3)
	arrow.Draw(img)

	// Check that pixels along the line are colored
	// The line goes from (10,10) to (100,100), so (50,50) should be colored
	pixel := img.RGBAAt(50, 50)
	if pixel.R == 0 {
		t.Error("Expected pixel at (50,50) to be colored red")
	}
}

func TestRectDraw(t *testing.T) {
	// Create a test image
	img := image.NewRGBA(image.Rect(0, 0, 200, 200))

	// Create and draw a filled rectangle
	rect := NewRect(image.Rect(50, 50, 150, 150), color.RGBA{G: 255, A: 255}, 2, true)
	rect.Draw(img)

	// Check that pixels inside are colored
	pixel := img.RGBAAt(100, 100)
	if pixel.G == 0 {
		t.Error("Expected pixel at (100,100) to be colored green for filled rect")
	}
}

func TestRectDrawOutline(t *testing.T) {
	// Create a test image
	img := image.NewRGBA(image.Rect(0, 0, 200, 200))

	// Create and draw an outlined rectangle
	rect := NewRect(image.Rect(50, 50, 150, 150), color.RGBA{B: 255, A: 255}, 3, false)
	rect.Draw(img)

	// Check that border pixels are colored
	pixel := img.RGBAAt(50, 100) // Left edge
	if pixel.B == 0 {
		t.Error("Expected pixel at left edge to be colored blue")
	}

	// Check that center is not colored (outline only)
	centerPixel := img.RGBAAt(100, 100)
	if centerPixel.B != 0 {
		t.Error("Expected center pixel to not be colored for outline rect")
	}
}

func TestAbs(t *testing.T) {
	tests := []struct {
		input int
		want  int
	}{
		{5, 5},
		{-5, 5},
		{0, 0},
		{-100, 100},
	}

	for _, tt := range tests {
		if got := abs(tt.input); got != tt.want {
			t.Errorf("abs(%d) = %d, want %d", tt.input, got, tt.want)
		}
	}
}
