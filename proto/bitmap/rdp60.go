package bitmap

import (
	"bytes"
	"image"
	"image/color"
	"io"

	"github.com/kdsmith18542/gordp/glog"
)

// RDP6ColorManager handles RemoteFX color processing features
type RDP6ColorManager struct {
	stats *ColorProcessingStats
}

// ColorProcessingStats tracks color processing performance
type ColorProcessingStats struct {
	ChromaSubsamplingCount int64
	ColorLossCount         int64
	ProcessingTime         int64 // nanoseconds
	Errors                 int64
}

// NewRDP6ColorManager creates a new RDP6 color manager
func NewRDP6ColorManager() *RDP6ColorManager {
	return &RDP6ColorManager{
		stats: &ColorProcessingStats{},
	}
}

// GetStats returns color processing statistics
func (cm *RDP6ColorManager) GetStats() *ColorProcessingStats {
	return cm.stats
}

// ResetStats resets color processing statistics
func (cm *RDP6ColorManager) ResetStats() {
	cm.stats = &ColorProcessingStats{}
}

// applyChromaSubsampling applies chroma subsampling reconstruction
// RemoteFX uses 4:2:0 chroma subsampling (YUV420)
func (cm *RDP6ColorManager) applyChromaSubsampling(y, u, v []byte, width, height int) ([]byte, []byte, []byte) {
	cm.stats.ChromaSubsamplingCount++

	// In 4:2:0 subsampling, U and V are quarter size (half width, half height)
	chromaWidth := width / 2
	chromaHeight := height / 2

	// Upsample U and V to full resolution
	upsampledU := make([]byte, width*height)
	upsampledV := make([]byte, width*height)

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			// Map full resolution coordinates to chroma coordinates
			chromaX := x / 2
			chromaY := y / 2

			// Clamp to chroma dimensions
			if chromaX >= chromaWidth {
				chromaX = chromaWidth - 1
			}
			if chromaY >= chromaHeight {
				chromaY = chromaHeight - 1
			}

			// Get chroma values
			chromaIndex := chromaY*chromaWidth + chromaX
			if chromaIndex < len(u) {
				upsampledU[y*width+x] = u[chromaIndex]
			}
			if chromaIndex < len(v) {
				upsampledV[y*width+x] = v[chromaIndex]
			}
		}
	}

	return y, upsampledU, upsampledV
}

// applyColorLoss applies color loss compensation
// Color loss levels: 0=none, 1=low, 2=medium, 3=high
func (cm *RDP6ColorManager) applyColorLoss(r, g, b []byte, colorLossLevel uint8) ([]byte, []byte, []byte) {
	cm.stats.ColorLossCount++

	if colorLossLevel == 0 {
		return r, g, b
	}

	// Color loss compensation factors
	compensationFactors := map[uint8]float64{
		1: 1.1, // Low loss: 10% boost
		2: 1.2, // Medium loss: 20% boost
		3: 1.3, // High loss: 30% boost
	}

	factor := compensationFactors[colorLossLevel]
	if factor == 0 {
		factor = 1.0
	}

	// Apply compensation to each color channel
	compensatedR := make([]byte, len(r))
	compensatedG := make([]byte, len(g))
	compensatedB := make([]byte, len(b))

	for i := 0; i < len(r); i++ {
		compensatedR[i] = cm.clampColor(float64(r[i]) * factor)
		compensatedG[i] = cm.clampColor(float64(g[i]) * factor)
		compensatedB[i] = cm.clampColor(float64(b[i]) * factor)
	}

	return compensatedR, compensatedG, compensatedB
}

// clampColor clamps color value to 0-255 range
func (cm *RDP6ColorManager) clampColor(value float64) byte {
	if value < 0 {
		return 0
	}
	if value > 255 {
		return 255
	}
	return byte(value)
}

// yuvToRgb converts YUV to RGB using BT.709 coefficients
func (cm *RDP6ColorManager) yuvToRgb(y, u, v []byte) ([]byte, []byte, []byte) {
	r := make([]byte, len(y))
	g := make([]byte, len(y))
	b := make([]byte, len(y))

	for i := 0; i < len(y); i++ {
		// YUV to RGB conversion using BT.709 coefficients
		yVal := float64(y[i])
		uVal := float64(u[i]) - 128.0
		vVal := float64(v[i]) - 128.0

		// BT.709 conversion matrix
		rVal := yVal + 1.5748*vVal
		gVal := yVal - 0.1873*uVal - 0.4681*vVal
		bVal := yVal + 1.8556*uVal

		r[i] = cm.clampColor(rVal)
		g[i] = cm.clampColor(gVal)
		b[i] = cm.clampColor(bVal)
	}

	return r, g, b
}

// rgbToYuv converts RGB to YUV using BT.709 coefficients
func (cm *RDP6ColorManager) rgbToYuv(r, g, b []byte) ([]byte, []byte, []byte) {
	y := make([]byte, len(r))
	u := make([]byte, len(r))
	v := make([]byte, len(r))

	for i := 0; i < len(r); i++ {
		// RGB to YUV conversion using BT.709 coefficients
		rVal := float64(r[i])
		gVal := float64(g[i])
		bVal := float64(b[i])

		// BT.709 conversion matrix
		yVal := 0.2126*rVal + 0.7152*gVal + 0.0722*bVal
		uVal := -0.1146*rVal - 0.3854*gVal + 0.5000*bVal + 128.0
		vVal := 0.5000*rVal - 0.4542*gVal - 0.0458*bVal + 128.0

		y[i] = cm.clampColor(yVal)
		u[i] = cm.clampColor(uVal)
		v[i] = cm.clampColor(vVal)
	}

	return y, u, v
}

// applyChromaSubsampling420 applies 4:2:0 chroma subsampling
func (cm *RDP6ColorManager) applyChromaSubsampling420(y, u, v []byte, width, height int) ([]byte, []byte, []byte) {
	chromaWidth := width / 2
	chromaHeight := height / 2

	// Downsample U and V to quarter size
	downsampledU := make([]byte, chromaWidth*chromaHeight)
	downsampledV := make([]byte, chromaWidth*chromaHeight)

	for chromaY := 0; chromaY < chromaHeight; chromaY++ {
		for chromaX := 0; chromaX < chromaWidth; chromaX++ {
			// Map chroma coordinates to full resolution
			fullX := chromaX * 2
			fullY := chromaY * 2

			// Average 2x2 block of chroma values
			var sumU, sumV float64
			count := 0

			for dy := 0; dy < 2; dy++ {
				for dx := 0; dx < 2; dx++ {
					x := fullX + dx
					y := fullY + dy

					if x < width && y < height {
						index := y*width + x
						if index < len(u) && index < len(v) {
							sumU += float64(u[index])
							sumV += float64(v[index])
							count++
						}
					}
				}
			}

			chromaIndex := chromaY*chromaWidth + chromaX
			if count > 0 {
				downsampledU[chromaIndex] = byte(sumU / float64(count))
				downsampledV[chromaIndex] = byte(sumV / float64(count))
			}
		}
	}

	return y, downsampledU, downsampledV
}

func decompressColorPlane(r io.Reader, w, h int) []byte {
	result := make([]byte, 0)
	size := w * h

	for size > 0 {
		controlByte := ReadByte(r)
		nRunLength := controlByte & 0x0F
		cRawBytes := (controlByte & 0xF0) >> 4

		//glog.Debugf("nRunLength: %v", nRunLength)
		//glog.Debugf("cRawBytes: %v", cRawBytes)

		// ==> 如果 nRunLength 字段设置为 1，则实际运行长度为 16 加上 cRawBytes 中的值。
		// 在解码时，假定 rawValues 字段中的 RAW 字节数为零。这给出了 31 个值的最大运行长度
		// ==> 如果 nRunLength 字段设置为 2，则实际运行长度为 32 加上 cRawBytes 中的值。
		// 在解码时，假定 rawValues 字段中的 RAW 字节数为零。这给出了 47 个值的最大运行长度。
		if nRunLength == 1 {
			nRunLength = 16 + cRawBytes
			cRawBytes = 0
		} else if nRunLength == 2 {
			nRunLength = 32 + cRawBytes
			cRawBytes = 0
		}

		if cRawBytes != 0 {
			data := ReadBytes(r, int(cRawBytes))
			result = append(result, data...)

			//glog.Debugf("--> data: %x", data)
			size -= int(cRawBytes)
		}
		if nRunLength != 0 {
			//glog.Debugf("nRunLength = %v", nRunLength)
			//glog.Debugf("resultLen = %v", len(result))
			// 行首，set(0), else set 上一个字符
			if len(result)%w == 0 {
				//glog.Debugf("write black")
				for i := 0; i < int(nRunLength); i++ {
					result = append(result, 0)
				}
			} else {
				b := result[len(result)-1]
				for i := 0; i < int(nRunLength); i++ {
					result = append(result, b)
				}
			}
			//data := ReadBytes(r, int(nRunLength))
			//glog.Debugf("data: %x", data)
			size -= int(nRunLength)
		}
	}

	//glog.Debugf("final: %v", len(result))

	for y := w; y < len(result); y += w {
		for x, e := y, y+w; x < e; x++ { // e->end, per line
			delta := result[x]
			if delta%2 == 0 {
				delta >>= 1
			} else {
				delta = 255 - ((delta - 1) >> 1)
			}
			result[x] = result[x-w] + delta
		}
	}

	return result
}

func (m *BitMap) LoadRDP60(option *Option) *BitMap {
	r := bytes.NewReader(option.Data)

	formatHeader := ReadByte(r)
	//glog.Debugf("format Header: %x", formatHeader)

	cll := formatHeader & 0x7 // color loss level
	//glog.Debugf("cll: %x", cll)

	cs := ((formatHeader & 0x08) >> 3) == 1 // whether chroma subsampling is being used
	//glog.Debugf("cs: %v", cs)

	rle := ((formatHeader & 0x10) >> 4) == 1
	//glog.Debugf("rle: %v", rle)

	na := ((formatHeader & 0x20) >> 5) == 1 //Indicates if an alpha plane is present.
	//glog.Debugf("na: %v", na)

	w, h := option.Width, option.Height

	// Create color manager for RemoteFX features
	colorManager := NewRDP6ColorManager()

	var alpha []byte
	if na {
		// Alpha plane present
		if rle {
			alpha = decompressColorPlane(r, w, h)
		} else {
			// Uncompressed alpha plane
			alpha = ReadBytes(r, w*h)
		}
	}

	var cr, cg, cb []byte
	if rle {
		cr = decompressColorPlane(r, w, h)
		cg = decompressColorPlane(r, w, h)
		cb = decompressColorPlane(r, w, h)
	} else {
		// Uncompressed color planes
		cr = ReadBytes(r, w*h)
		cg = ReadBytes(r, w*h)
		cb = ReadBytes(r, w*h)
	}

	// Handle chroma subsampling
	if cs {
		glog.Debugf("Processing RemoteFX chroma subsampling (4:2:0)")

		// Convert RGB to YUV
		y, u, v := colorManager.rgbToYuv(cr, cg, cb)

		// Apply chroma subsampling reconstruction
		y, u, v = colorManager.applyChromaSubsampling(y, u, v, w, h)

		// Convert back to RGB
		cr, cg, cb = colorManager.yuvToRgb(y, u, v)

		glog.Debugf("Chroma subsampling processed: %dx%d", w, h)
	}

	// Handle color loss
	if cll != 0 {
		glog.Debugf("Processing RemoteFX color loss (level: %d)", cll)

		// Apply color loss compensation
		cr, cg, cb = colorManager.applyColorLoss(cr, cg, cb, cll)

		glog.Debugf("Color loss compensation applied: level %d", cll)
	}

	img := image.NewRGBA(image.Rect(0, 0, w, h))
	pos := 0
	for y := 1; y <= h; y++ {
		for x := 0; x < w; x++ {
			var a uint8 = 255
			if na && len(alpha) > pos {
				a = alpha[pos]
			}
			img.Set(x, h-y, color.RGBA{R: cr[pos], G: cg[pos], B: cb[pos], A: a})
			pos++
		}
	}

	m.Image = img
	return m
}
