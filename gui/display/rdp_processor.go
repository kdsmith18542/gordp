package display

import (
	"fmt"
	"image"
	"image/color"

	"github.com/kdsmith18542/gordp"
	"github.com/kdsmith18542/gordp/proto/bitmap"
)

// QtRDPProcessor implements the gordp.Processor interface for Qt display
type QtRDPProcessor struct {
	displayWidget *RDPDisplayWidget
	client        *gordp.Client
}

// NewQtRDPProcessor creates a new Qt RDP processor
func NewQtRDPProcessor(displayWidget *RDPDisplayWidget) *QtRDPProcessor {
	return &QtRDPProcessor{
		displayWidget: displayWidget,
	}
}

// ProcessBitmap processes bitmap data from GoRDP and updates the display
func (p *QtRDPProcessor) ProcessBitmap(option *bitmap.Option, bitmap *bitmap.BitMap) {
	if p.displayWidget == nil {
		return
	}

	// Convert GoRDP bitmap to Go image
	img := p.convertBitmapToImage(option, bitmap)
	if img == nil {
		return
	}

	// Update the display widget
	p.displayWidget.UpdateDisplay(img)
}

// convertBitmapToImage converts GoRDP bitmap to Go image.Image
func (p *QtRDPProcessor) convertBitmapToImage(option *bitmap.Option, bitmap *bitmap.BitMap) image.Image {
	if bitmap == nil || bitmap.Image == nil {
		return nil
	}

	// Check if the bitmap already has an image
	if img, ok := bitmap.Image.(image.Image); ok {
		return img
	}

	// If we have raw data, convert it to an image
	if option != nil && len(option.Data) > 0 {
		return p.convertRawDataToImage(option)
	}

	return nil
}

// convertRawDataToImage converts raw bitmap data to an image
func (p *QtRDPProcessor) convertRawDataToImage(option *bitmap.Option) image.Image {
	if option == nil || len(option.Data) == 0 {
		return nil
	}

	// Create a new RGBA image
	bounds := image.Rect(option.Left, option.Top, option.Left+option.Width, option.Top+option.Height)
	rgba := image.NewRGBA(bounds)

	// Convert bitmap data to RGBA based on bits per pixel
	switch option.BitPerPixel {
	case 8:
		p.convert8BitToRGBA(option, rgba)
	case 16:
		p.convert16BitToRGBA(option, rgba)
	case 24:
		p.convert24BitToRGBA(option, rgba)
	case 32:
		p.convert32BitToRGBA(option, rgba)
	default:
		fmt.Printf("Unsupported bits per pixel: %d\n", option.BitPerPixel)
		return nil
	}

	return rgba
}

// convert8BitToRGBA converts 8-bit bitmap to RGBA
func (p *QtRDPProcessor) convert8BitToRGBA(option *bitmap.Option, rgba *image.RGBA) {
	data := option.Data
	width := option.Width
	height := option.Height

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			idx := y*width + x
			if idx < len(data) {
				val := data[idx]
				rgba.Set(x, y, color.RGBA{val, val, val, 255})
			}
		}
	}
}

// convert16BitToRGBA converts 16-bit bitmap to RGBA
func (p *QtRDPProcessor) convert16BitToRGBA(option *bitmap.Option, rgba *image.RGBA) {
	data := option.Data
	width := option.Width
	height := option.Height

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			idx := (y*width + x) * 2
			if idx+1 < len(data) {
				// 16-bit RGB565 format
				pixel := uint16(data[idx]) | uint16(data[idx+1])<<8
				r := uint8((pixel >> 11) & 0x1F << 3)
				g := uint8((pixel >> 5) & 0x3F << 2)
				b := uint8(pixel & 0x1F << 3)
				rgba.Set(x, y, color.RGBA{r, g, b, 255})
			}
		}
	}
}

// convert24BitToRGBA converts 24-bit bitmap to RGBA
func (p *QtRDPProcessor) convert24BitToRGBA(option *bitmap.Option, rgba *image.RGBA) {
	data := option.Data
	width := option.Width
	height := option.Height

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			idx := (y*width + x) * 3
			if idx+2 < len(data) {
				// 24-bit RGB format
				r := data[idx]
				g := data[idx+1]
				b := data[idx+2]
				rgba.Set(x, y, color.RGBA{r, g, b, 255})
			}
		}
	}
}

// convert32BitToRGBA converts 32-bit bitmap to RGBA
func (p *QtRDPProcessor) convert32BitToRGBA(option *bitmap.Option, rgba *image.RGBA) {
	data := option.Data
	width := option.Width
	height := option.Height

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			idx := (y*width + x) * 4
			if idx+3 < len(data) {
				// 32-bit RGBA format
				r := data[idx]
				g := data[idx+1]
				b := data[idx+2]
				a := data[idx+3]
				rgba.Set(x, y, color.RGBA{r, g, b, a})
			}
		}
	}
}

// SetClient sets the RDP client reference
func (p *QtRDPProcessor) SetClient(client *gordp.Client) {
	p.client = client
}

// GetClient returns the RDP client reference
func (p *QtRDPProcessor) GetClient() *gordp.Client {
	return p.client
}
