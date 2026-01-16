package assets

import (
	"bytes"
	"image"
	"image/color"
	"image/png"

	"fyne.io/fyne/v2"
)

// MenuBarIcon returns a template icon for the macOS menu bar
// It's a simple viewfinder/screenshot icon (crosshairs with corners)
func MenuBarIcon() fyne.Resource {
	img := generateMenuBarIcon(22) // Standard menu bar size
	return &fyne.StaticResource{
		StaticName:    "menubar-icon.png",
		StaticContent: imageToPNG(img),
	}
}

// generateMenuBarIcon creates a simple viewfinder icon
func generateMenuBarIcon(size int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, size, size))

	c := color.RGBA{255, 255, 255, 255}

	cornerLen := size / 3
	lineWidth := 2
	margin := 2

	drawHLine(img, margin, margin, margin+cornerLen, lineWidth, c)
	drawVLine(img, margin, margin, margin+cornerLen, lineWidth, c)

	drawHLine(img, size-margin-cornerLen, margin, size-margin, lineWidth, c)
	drawVLine(img, size-margin-lineWidth, margin, margin+cornerLen, lineWidth, c)

	drawHLine(img, margin, size-margin-lineWidth, margin+cornerLen, lineWidth, c)
	drawVLine(img, margin, size-margin-cornerLen, size-margin, lineWidth, c)

	drawHLine(img, size-margin-cornerLen, size-margin-lineWidth, size-margin, lineWidth, c)
	drawVLine(img, size-margin-lineWidth, size-margin-cornerLen, size-margin, lineWidth, c)

	centerX, centerY := size/2, size/2
	dotSize := 2
	for y := centerY - dotSize/2; y <= centerY+dotSize/2; y++ {
		for x := centerX - dotSize/2; x <= centerX+dotSize/2; x++ {
			if x >= 0 && x < size && y >= 0 && y < size {
				img.Set(x, y, c)
			}
		}
	}

	return img
}

func drawHLine(img *image.RGBA, x1, y, x2, thickness int, c color.Color) {
	for t := 0; t < thickness; t++ {
		for x := x1; x < x2; x++ {
			if y+t < img.Bounds().Dy() {
				img.Set(x, y+t, c)
			}
		}
	}
}

func drawVLine(img *image.RGBA, x, y1, y2, thickness int, c color.Color) {
	for t := 0; t < thickness; t++ {
		for y := y1; y < y2; y++ {
			if x+t < img.Bounds().Dx() {
				img.Set(x+t, y, c)
			}
		}
	}
}

func imageToPNG(img image.Image) []byte {
	var buf bytes.Buffer
	png.Encode(&buf, img)
	return buf.Bytes()
}
