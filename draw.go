package main

import (
	"bytes"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"log"
	"os"

	"github.com/fogleman/gg"
	"github.com/nfnt/resize"
)

func drawImage(template string, fontSize float64, text string) (*bytes.Buffer, error) {
	// Define image width and padding
	const (
		imageWidth = 512
		padding    = 8
	)

	// Create a new context with the correct width and height
	textImage := gg.NewContext(imageWidth, 0)

	// Set the font face and size
	fontPath := "./assets/fonts/Anton-Regular.ttf"
	if err := textImage.LoadFontFace(fontPath, fontSize); err != nil {
		log.Printf("Could not load font: %v", err)
		return nil, err
	}

	// Word wrap the text and get the height required for rendering
	lines := textImage.WordWrap(text, imageWidth-2*padding)
	textHeight := float64(len(lines)) * fontSize * 1.5

	// Adjust text image height for padding
	textImageHeight := int(textHeight + 3*padding)
	textImage = gg.NewContext(imageWidth, textImageHeight)

	// Set the background to white
	textImage.SetRGB(1, 1, 1)
	textImage.Clear()

	// Set the font face and size
	if err := textImage.LoadFontFace(fontPath, fontSize); err != nil {
		log.Printf("Could not load font: %v", err)
		return nil, err
	}

	// Set the color to black for the text
	textImage.SetRGB(0, 0, 0)

	// Draw the wrapped text justified center with padding
	drawTextCenter(textImage, lines, fontSize, padding)

	// Load the image to place below the text
	img, err := loadImage(fmt.Sprintf("./assets/img/%s.png", template))
	if err != nil {
		log.Printf("Could not load image: %v", err)
		return nil, err
	}

	// Resize the template image to match the width of the text image
	templateImage := resize.Resize(uint(imageWidth), 0, img, resize.Lanczos3)

	// Get the height of the loaded image
	imgHeight := templateImage.Bounds().Dy()

	// Create a new context for the final image
	finalHeight := int(textHeight) + imgHeight + 2*padding
	finalImage := image.NewRGBA(image.Rect(0, 0, imageWidth, finalHeight))

	// Draw the text image onto the final image
	draw.Draw(finalImage, image.Rect(0, 0, imageWidth, textImageHeight), textImage.Image(), image.Point{0, 0}, draw.Src)

	// Draw the scaled coffee image below the text image
	draw.Draw(finalImage, image.Rect(0, textImageHeight, imageWidth, finalHeight), templateImage, image.Point{0, 0}, draw.Src)

	// Save the final image to a file
	var buf bytes.Buffer
	if err := png.Encode(&buf, finalImage); err != nil {
		log.Printf("Could not save final image")
		return nil, err
	}
	return &buf, nil
}

func drawTextCenter(dc *gg.Context, lines []string, fontSize, padding float64) {
	totalHeight := float64(len(lines)) * fontSize * 1.5
	y := (float64(dc.Height()) - totalHeight) / 2

	// Adjust y for padding
	y += padding

	for _, line := range lines {
		width, _ := dc.MeasureString(line)
		x := (float64(dc.Width()) - width) / 2
		dc.DrawString(line, x, y+fontSize)
		y += fontSize * 1.5
	}
}

func loadImage(filename string) (image.Image, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}

	return img, nil
}
