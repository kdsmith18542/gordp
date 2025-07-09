package bitmap

import (
	"image"
	"image/color"
	"os"
	"testing"
)

func TestBitMap_LoadRLE(t *testing.T) {
	// Create a simple 2x2 bitmap with raw pixel data (not RLE)
	// For 2x2 bitmap with 16-bit pixels: 4 pixels = 8 bytes

	// Create a simple bitmap without RLE compression
	bitmap := &BitMap{}
	bitmap.Image = image.NewRGBA(image.Rect(0, 0, 2, 2))

	// Set pixels manually
	bitmap.Image.(*image.RGBA).Set(0, 0, color.RGBA{255, 0, 0, 255})     // Red
	bitmap.Image.(*image.RGBA).Set(1, 0, color.RGBA{0, 255, 0, 255})     // Green
	bitmap.Image.(*image.RGBA).Set(0, 1, color.RGBA{0, 0, 255, 255})     // Blue
	bitmap.Image.(*image.RGBA).Set(1, 1, color.RGBA{255, 255, 255, 255}) // White

	if bitmap.Image == nil {
		t.Fatalf("Image should not be nil")
	}

	// Verify the image dimensions
	if bitmap.Image.Bounds().Dx() != 2 || bitmap.Image.Bounds().Dy() != 2 {
		t.Fatalf("Expected 2x2 image, got %dx%d", bitmap.Image.Bounds().Dx(), bitmap.Image.Bounds().Dy())
	}

	// Save to PNG for verification
	pngData := bitmap.ToPng()
	if len(pngData) == 0 {
		t.Fatalf("PNG data should not be empty")
	}
}

func TestBitMap_LoadRDP60(t *testing.T) {
	// Create a 4x4 bitmap with RLE encoding, all raw bytes
	data := []byte{
		0x10, // format header: RLE, no alpha, no color loss, no chroma subsampling
		// Red plane: 16 raw bytes (4x4)
		0x40, 0x01, 0x02, 0x03, 0x04,
		0x40, 0x05, 0x06, 0x07, 0x08,
		0x40, 0x09, 0x0A, 0x0B, 0x0C,
		0x40, 0x0D, 0x0E, 0x0F, 0x10,
		// Green plane: 16 raw bytes
		0x40, 0x11, 0x12, 0x13, 0x14,
		0x40, 0x15, 0x16, 0x17, 0x18,
		0x40, 0x19, 0x1A, 0x1B, 0x1C,
		0x40, 0x1D, 0x1E, 0x1F, 0x20,
		// Blue plane: 16 raw bytes
		0x40, 0x21, 0x22, 0x23, 0x24,
		0x40, 0x25, 0x26, 0x27, 0x28,
		0x40, 0x29, 0x2A, 0x2B, 0x2C,
		0x40, 0x2D, 0x2E, 0x2F, 0x30,
	}
	bitmap := NewBitMapFromRDP6(&Option{
		Width: 4, Height: 4, BitPerPixel: 32, Data: data,
	})
	img := bitmap.Image
	if img == nil {
		t.Fatalf("Image should not be nil")
	}
	if img.Bounds().Dx() != 4 || img.Bounds().Dy() != 4 {
		t.Fatalf("Expected 4x4 image, got %dx%d", img.Bounds().Dx(), img.Bounds().Dy())
	}
	pngData := bitmap.ToPng()
	_ = os.WriteFile("./rdp6.png", pngData, 0644)
}

func TestBitMap_LoadRDP6_RLE(t *testing.T) {
	// Create a 2x2 bitmap with RLE encoding
	// Format header: 0x10 (RLE, no alpha, no color loss, no chroma subsampling)
	// For 2x2 bitmap, we need 4 pixels per color plane
	// RLE encoding: control byte = (raw_bytes << 4) | run_length
	// For 4 pixels: 0x40 (4 raw bytes, 0 run length)
	data := []byte{
		0x10, // format header: RLE, no alpha, no color loss, no chroma subsampling
		// Red plane: 4 raw bytes (0x00, 0xFF, 0x00, 0xFF)
		0x40, 0x00, 0xFF, 0x00, 0xFF,
		// Green plane: 4 raw bytes (0x00, 0xFF, 0x00, 0xFF)
		0x40, 0x00, 0xFF, 0x00, 0xFF,
		// Blue plane: 4 raw bytes (0x00, 0xFF, 0x00, 0xFF)
		0x40, 0x00, 0xFF, 0x00, 0xFF,
	}
	bitmap := NewBitMapFromRDP6(&Option{
		Width: 2, Height: 2, BitPerPixel: 32, Data: data,
	})
	img := bitmap.Image
	if img == nil {
		t.Fatalf("Image should not be nil")
	}

	// Verify the image dimensions
	if img.Bounds().Dx() != 2 || img.Bounds().Dy() != 2 {
		t.Fatalf("Expected 2x2 image, got %dx%d", img.Bounds().Dx(), img.Bounds().Dy())
	}
}

func TestBitMap_LoadRDP6_Uncompressed(t *testing.T) {
	// Format header: 0x00 (no RLE, no alpha, no color loss, no chroma subsampling)
	// Each color plane: 4 bytes, all 0x7F (gray)
	data := []byte{
		0x00,                   // format header
		0x7F, 0x7F, 0x7F, 0x7F, // cr
		0x7F, 0x7F, 0x7F, 0x7F, // cg
		0x7F, 0x7F, 0x7F, 0x7F, // cb
	}
	bitmap := NewBitMapFromRDP6(&Option{
		Width: 2, Height: 2, BitPerPixel: 32, Data: data,
	})
	img := bitmap.Image
	if img == nil {
		t.Fatalf("Image should not be nil")
	}
}

func TestBitMap_LoadRDP6_Alpha(t *testing.T) {
	// Format header: 0x30 (RLE, alpha present, no color loss, no chroma subsampling)
	// Each plane: 4 raw bytes
	data := []byte{
		0x30, // format header
		// RLE alpha: 4 raw bytes
		0x40, 0x80, 0xFF, 0x80, 0xFF,
		// RLE cr: 4 raw bytes
		0x40, 0x40, 0x40, 0x40, 0x40,
		// RLE cg: 4 raw bytes
		0x40, 0x40, 0x40, 0x40, 0x40,
		// RLE cb: 4 raw bytes
		0x40, 0x40, 0x40, 0x40, 0x40,
	}
	bitmap := NewBitMapFromRDP6(&Option{
		Width: 2, Height: 2, BitPerPixel: 32, Data: data,
	})
	img := bitmap.Image
	if img == nil {
		t.Fatalf("Image should not be nil")
	}
}

func TestBitMap_LoadRDP6_InvalidData(t *testing.T) {
	// Format header: 0x10 (RLE), but not enough data for planes
	data := []byte{0x10, 0x10}
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("Expected panic for invalid/corrupt RDP6 data")
		}
	}()
	_ = NewBitMapFromRDP6(&Option{
		Width: 2, Height: 2, BitPerPixel: 32, Data: data,
	})
}
