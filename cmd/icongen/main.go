// Command icongen generates app icons for Schnappit
package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: icongen <output-dir>")
		os.Exit(1)
	}

	outDir := os.Args[1]
	if err := os.MkdirAll(outDir, 0755); err != nil {
		fmt.Printf("Failed to create output directory: %v\n", err)
		os.Exit(1)
	}

	// Generate icons at required sizes for .icns
	sizes := []int{16, 32, 64, 128, 256, 512, 1024}
	for _, size := range sizes {
		img := generateAppIcon(size)
		filename := filepath.Join(outDir, fmt.Sprintf("icon_%dx%d.png", size, size))
		if err := savePNG(filename, img); err != nil {
			fmt.Printf("Failed to save %s: %v\n", filename, err)
			os.Exit(1)
		}
		fmt.Printf("Generated %s\n", filename)
	}
}

func generateAppIcon(size int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, size, size))

	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			t := float64(y) / float64(size)
			r := uint8(88 + t*20)
			g := uint8(86 + t*40)
			b := uint8(214 - t*30)

			cornerRadius := size / 5
			if isInRoundedRect(x, y, size, size, cornerRadius) {
				img.Set(x, y, color.RGBA{r, g, b, 255})
			}
		}
	}

	drawViewfinder(img, size)

	return img
}

func isInRoundedRect(x, y, w, h, r int) bool {
	if x < r && y < r {
		return distSq(x, y, r, r) <= r*r
	}
	if x >= w-r && y < r {
		return distSq(x, y, w-r-1, r) <= r*r
	}
	if x < r && y >= h-r {
		return distSq(x, y, r, h-r-1) <= r*r
	}
	if x >= w-r && y >= h-r {
		return distSq(x, y, w-r-1, h-r-1) <= r*r
	}
	return true
}

func distSq(x1, y1, x2, y2 int) int {
	dx := x1 - x2
	dy := y1 - y2
	return dx*dx + dy*dy
}

func drawViewfinder(img *image.RGBA, size int) {
	c := color.RGBA{255, 255, 255, 255}

	lineWidth := max(size/11, 2)
	margin := size / 5
	cornerLen := size / 3

	drawHLine(img, margin, margin, margin+cornerLen, lineWidth, c)
	drawVLine(img, margin, margin, margin+cornerLen, lineWidth, c)

	drawHLine(img, size-margin-cornerLen, margin, size-margin, lineWidth, c)
	drawVLine(img, size-margin-lineWidth, margin, margin+cornerLen, lineWidth, c)

	drawHLine(img, margin, size-margin-lineWidth, margin+cornerLen, lineWidth, c)
	drawVLine(img, margin, size-margin-cornerLen, size-margin, lineWidth, c)

	drawHLine(img, size-margin-cornerLen, size-margin-lineWidth, size-margin, lineWidth, c)
	drawVLine(img, size-margin-lineWidth, size-margin-cornerLen, size-margin, lineWidth, c)

	centerX, centerY := size/2, size/2
	dotRadius := max(size/16, 2)
	for y := centerY - dotRadius; y <= centerY+dotRadius; y++ {
		for x := centerX - dotRadius; x <= centerX+dotRadius; x++ {
			if distSq(x, y, centerX, centerY) <= dotRadius*dotRadius {
				img.Set(x, y, c)
			}
		}
	}
}

func drawHLine(img *image.RGBA, x1, y, x2, thickness int, c color.Color) {
	for t := 0; t < thickness; t++ {
		for x := x1; x < x2; x++ {
			if y+t >= 0 && y+t < img.Bounds().Dy() && x >= 0 && x < img.Bounds().Dx() {
				img.Set(x, y+t, c)
			}
		}
	}
}

func drawVLine(img *image.RGBA, x, y1, y2, thickness int, c color.Color) {
	for t := 0; t < thickness; t++ {
		for y := y1; y < y2; y++ {
			if x+t >= 0 && x+t < img.Bounds().Dx() && y >= 0 && y < img.Bounds().Dy() {
				img.Set(x+t, y, c)
			}
		}
	}
}

func savePNG(filename string, img image.Image) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}
