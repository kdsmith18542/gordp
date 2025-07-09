package bitmap

import (
	"bytes"
	"image"
	"image/color"
	"io"

	"github.com/GoFeGroup/gordp/core"
)

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

	// Handle chroma subsampling (not implemented)
	if cs {
		// Chroma subsampling is not supported in this implementation
		core.ThrowError("RemoteFX chroma subsampling is not supported")
	}

	// Handle color loss (not implemented)
	if cll != 0 {
		// Color loss is not supported in this implementation
		core.ThrowError("RemoteFX color loss is not supported")
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
