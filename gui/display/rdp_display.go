package display

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"sync"

	"github.com/kdsmith18542/gordp/proto/bitmap"
)

// RDPDisplayWidget represents the widget that displays RDP content
type RDPDisplayWidget struct {
	// Current display state
	currentImage image.Image
	zoomLevel    float64

	// Display properties
	width  int
	height int

	// Display file for saving screenshots
	displayFile string

	// Thread safety
	mu sync.RWMutex

	// Callback for Qt integration
	onBitmapUpdate func(image.Image)
}

// NewRDPDisplayWidget creates a new RDP display widget
func NewRDPDisplayWidget() *RDPDisplayWidget {
	widget := &RDPDisplayWidget{
		zoomLevel:   1.0,
		width:       1024,
		height:      768,
		displayFile: "rdp_display.png",
	}

	fmt.Println("RDP Display Widget created")
	return widget
}

// UpdateDisplay updates the display with new image data
func (w *RDPDisplayWidget) UpdateDisplay(img image.Image) {
	if img == nil {
		return
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	w.currentImage = img
	w.width = img.Bounds().Dx()
	w.height = img.Bounds().Dy()

	// Save the image to file for display
	w.saveImage(img)

	// Notify Qt GUI if callback is set
	if w.onBitmapUpdate != nil {
		w.onBitmapUpdate(img)
	}

	fmt.Printf("Display updated: %dx%d\n", w.width, w.height)
}

// HandleBitmapUpdate handles bitmap updates from the RDP client
func (w *RDPDisplayWidget) HandleBitmapUpdate(option *bitmap.Option, bitmap *bitmap.BitMap) {
	if option == nil || bitmap == nil {
		fmt.Println("Invalid bitmap update: nil option or bitmap")
		return
	}

	// Convert GoRDP bitmap to Go image
	img := w.convertBitmapToImage(option, bitmap)
	if img == nil {
		fmt.Printf("Failed to convert bitmap: %dx%d at (%d,%d)\n",
			option.Width, option.Height, option.Left, option.Top)
		return
	}

	// Update the display with the new image
	w.UpdateDisplay(img)

	fmt.Printf("Bitmap update handled: %dx%d at (%d,%d)\n",
		option.Width, option.Height, option.Left, option.Top)
}

// convertBitmapToImage converts GoRDP bitmap to Go image.Image
func (w *RDPDisplayWidget) convertBitmapToImage(option *bitmap.Option, bitmap *bitmap.BitMap) image.Image {
	if bitmap == nil || bitmap.Image == nil {
		return nil
	}

	// Check if the bitmap already has an image
	if img, ok := bitmap.Image.(image.Image); ok {
		return img
	}

	// If we have raw data, convert it to an image
	if option != nil && len(option.Data) > 0 {
		return w.convertRawDataToImage(option)
	}

	return nil
}

// convertRawDataToImage converts raw bitmap data to an image
func (w *RDPDisplayWidget) convertRawDataToImage(option *bitmap.Option) image.Image {
	if option == nil || len(option.Data) == 0 {
		return nil
	}

	// Create a new RGBA image
	bounds := image.Rect(option.Left, option.Top, option.Left+option.Width, option.Top+option.Height)
	rgba := image.NewRGBA(bounds)

	// Convert bitmap data to RGBA based on bits per pixel
	switch option.BitPerPixel {
	case 8:
		w.convert8BitToRGBA(option, rgba)
	case 16:
		w.convert16BitToRGBA(option, rgba)
	case 24:
		w.convert24BitToRGBA(option, rgba)
	case 32:
		w.convert32BitToRGBA(option, rgba)
	default:
		fmt.Printf("Unsupported bits per pixel: %d\n", option.BitPerPixel)
		return nil
	}

	return rgba
}

// convert8BitToRGBA converts 8-bit bitmap to RGBA
func (w *RDPDisplayWidget) convert8BitToRGBA(option *bitmap.Option, rgba *image.RGBA) {
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
func (w *RDPDisplayWidget) convert16BitToRGBA(option *bitmap.Option, rgba *image.RGBA) {
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
func (w *RDPDisplayWidget) convert24BitToRGBA(option *bitmap.Option, rgba *image.RGBA) {
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
func (w *RDPDisplayWidget) convert32BitToRGBA(option *bitmap.Option, rgba *image.RGBA) {
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

// SetBitmapUpdateCallback sets a callback for bitmap updates
func (w *RDPDisplayWidget) SetBitmapUpdateCallback(callback func(image.Image)) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.onBitmapUpdate = callback
}

// GetCurrentImage returns the current display image
func (w *RDPDisplayWidget) GetCurrentImage() image.Image {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.currentImage
}

// saveImage saves the image to a file
func (w *RDPDisplayWidget) saveImage(img image.Image) {
	file, err := os.Create(w.displayFile)
	if err != nil {
		fmt.Printf("Error creating display file: %v\n", err)
		return
	}
	defer file.Close()

	err = png.Encode(file, img)
	if err != nil {
		fmt.Printf("Error encoding image: %v\n", err)
		return
	}

	fmt.Printf("Display saved to %s\n", w.displayFile)
}

// SetZoom sets the zoom level
func (w *RDPDisplayWidget) SetZoom(zoom float64) {
	if zoom < 0.1 || zoom > 5.0 {
		return // Limit zoom range
	}

	w.mu.Lock()
	defer w.mu.Unlock()
	w.zoomLevel = zoom
	fmt.Printf("Zoom level set to %.2fx\n", zoom)
}

// GetZoom returns the current zoom level
func (w *RDPDisplayWidget) GetZoom() float64 {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.zoomLevel
}

// ZoomIn increases the zoom level
func (w *RDPDisplayWidget) ZoomIn() {
	w.SetZoom(w.zoomLevel * 1.25)
}

// ZoomOut decreases the zoom level
func (w *RDPDisplayWidget) ZoomOut() {
	w.SetZoom(w.zoomLevel / 1.25)
}

// ResetZoom resets zoom to 100%
func (w *RDPDisplayWidget) ResetZoom() {
	w.SetZoom(1.0)
}

// GetSize returns the current display size
func (w *RDPDisplayWidget) GetSize() (width, height int) {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.width, w.height
}

// SetConnectionStatus updates the connection status display
func (w *RDPDisplayWidget) SetConnectionStatus(connected bool, message string) {
	if !connected {
		fmt.Printf("Connection status: %s\n", message)
		// Create a simple "disconnected" image
		w.createDisconnectedImage(message)
	}
}

// createDisconnectedImage creates a simple image showing disconnected status
func (w *RDPDisplayWidget) createDisconnectedImage(message string) {
	// Create a simple gray image with text
	img := image.NewRGBA(image.Rect(0, 0, 800, 600))

	// Fill with gray background
	for y := 0; y < 600; y++ {
		for x := 0; x < 800; x++ {
			img.Set(x, y, color.RGBA{128, 128, 128, 255})
		}
	}

	w.mu.Lock()
	w.currentImage = img
	w.width = 800
	w.height = 600
	w.mu.Unlock()

	// Save the disconnected image
	w.saveImage(img)
	fmt.Printf("Disconnected image created: %s\n", message)
}

// GetDisplayFile returns the current display file path
func (w *RDPDisplayWidget) GetDisplayFile() string {
	return w.displayFile
}
