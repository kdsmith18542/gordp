package t128

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/kdsmith18542/gordp/core"
	"github.com/kdsmith18542/gordp/glog"
)

const (
	PDUTYPE2_UPDATE                      = 0x02
	PDUTYPE2_CONTROL                     = 0x14
	PDUTYPE2_POINTER                     = 0x1B
	PDUTYPE2_INPUT                       = 0x1C
	PDUTYPE2_SYNCHRONIZE                 = 0x1F
	PDUTYPE2_REFRESH_RECT                = 0x21
	PDUTYPE2_PLAY_SOUND                  = 0x22
	PDUTYPE2_SUPPRESS_OUTPUT             = 0x23
	PDUTYPE2_SHUTDOWN_REQUEST            = 0x24
	PDUTYPE2_SHUTDOWN_DENIED             = 0x25
	PDUTYPE2_SAVE_SESSION_INFO           = 0x26
	PDUTYPE2_FONTLIST                    = 0x27
	PDUTYPE2_FONTMAP                     = 0x28
	PDUTYPE2_SET_KEYBOARD_INDICATORS     = 0x29
	PDUTYPE2_BITMAPCACHE_PERSISTENT_LIST = 0x2B
	PDUTYPE2_BITMAPCACHE_ERROR_PDU       = 0x2C
	PDUTYPE2_SET_KEYBOARD_IME_STATUS     = 0x2D
	PDUTYPE2_OFFSCRCACHE_ERROR_PDU       = 0x2E
	PDUTYPE2_SET_ERROR_INFO_PDU          = 0x2F
	PDUTYPE2_DRAWNINEGRID_ERROR_PDU      = 0x30
	PDUTYPE2_DRAWGDIPLUS_ERROR_PDU       = 0x31
	PDUTYPE2_ARC_STATUS_PDU              = 0x32
	PDUTYPE2_STATUS_INFO_PDU             = 0x36
	PDUTYPE2_MONITOR_LAYOUT_PDU          = 0x37
)

// StreamId
const (
	STREAM_UNDEFINED = 0x00
	STREAM_LOW       = 0x01
	STREAM_MED       = 0x02
	STREAM_HI        = 0x04
)

// Level-2 Compression Flags
const (
	PACKET_COMPRESSED = 0x20
	PACKET_AT_FRONT   = 0x40
	PACKET_FLUSHED    = 0x80
)

// Level-1 Compression Flags
const (
	L1_PACKET_AT_FRONT   = 0x04
	L1_NO_COMPRESSION    = 0x02
	L1_COMPRESSED        = 0x01
	L1_INNER_COMPRESSION = 0x10
)

const (
	PACKET_COMPR_TYPE_8K    = 0x0
	PACKET_COMPR_TYPE_64K   = 0x1
	PACKET_COMPR_TYPE_RDP6  = 0x2
	PACKET_COMPR_TYPE_RDP61 = 0x3
)

// ShareDataCompressionManager handles RDP compression for share data
type ShareDataCompressionManager struct {
	history8K  []byte
	history64K []byte
	mutex      sync.Mutex
	stats      *CompressionStats
}

// NewShareDataCompressionManager creates a new share data compression manager
func NewShareDataCompressionManager() *ShareDataCompressionManager {
	return &ShareDataCompressionManager{
		history8K:  make([]byte, 0, 8192),  // 8KB history buffer
		history64K: make([]byte, 0, 65536), // 64KB history buffer
		stats:      &CompressionStats{},
	}
}

// Decompress decompresses share data based on compression type
func (cm *ShareDataCompressionManager) Decompress(data []byte, comprType uint8) ([]byte, error) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	startTime := time.Now()
	defer func() {
		cm.stats.DecompressionTime = time.Since(startTime).Nanoseconds()
	}()

	// Check if data is actually compressed
	if len(data) < 2 {
		return data, nil // Not compressed
	}

	var decompressed []byte
	var err error

	switch comprType {
	case PACKET_COMPR_TYPE_8K:
		decompressed, err = cm.decompress8K(data)
	case PACKET_COMPR_TYPE_64K:
		decompressed, err = cm.decompress64K(data)
	case PACKET_COMPR_TYPE_RDP6:
		decompressed, err = cm.decompressRDP6(data)
	case PACKET_COMPR_TYPE_RDP61:
		decompressed, err = cm.decompressRDP61(data)
	default:
		// No compression or unknown type
		return data, nil
	}

	if err != nil {
		cm.stats.Errors++
		return nil, fmt.Errorf("failed to decompress share data (type %d): %v", comprType, err)
	}

	// Update statistics
	cm.stats.TotalCompressed += int64(len(data))
	cm.stats.TotalUncompressed += int64(len(decompressed))
	if cm.stats.TotalUncompressed > 0 {
		cm.stats.CompressionRatio = float64(cm.stats.TotalCompressed) / float64(cm.stats.TotalUncompressed)
	}

	glog.Debugf("Share data decompression (type %d): %d -> %d bytes (%.1f%% reduction)",
		comprType, len(data), len(decompressed),
		float64(len(data)-len(decompressed))/float64(len(data))*100)

	return decompressed, nil
}

// decompress8K decompresses using 8KB history buffer
func (cm *ShareDataCompressionManager) decompress8K(data []byte) ([]byte, error) {
	// 8KB compression uses zlib with 8KB window
	reader := bytes.NewReader(data)
	zlibReader, err := zlib.NewReader(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to create zlib reader for 8K decompression: %v", err)
	}
	defer zlibReader.Close()

	var buf bytes.Buffer
	_, err = io.Copy(&buf, zlibReader)
	if err != nil {
		return nil, fmt.Errorf("failed to decompress 8K data: %v", err)
	}

	return buf.Bytes(), nil
}

// decompress64K decompresses using 64KB history buffer
func (cm *ShareDataCompressionManager) decompress64K(data []byte) ([]byte, error) {
	// 64KB compression uses zlib with 64KB window
	reader := bytes.NewReader(data)
	zlibReader, err := zlib.NewReader(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to create zlib reader for 64K decompression: %v", err)
	}
	defer zlibReader.Close()

	var buf bytes.Buffer
	_, err = io.Copy(&buf, zlibReader)
	if err != nil {
		return nil, fmt.Errorf("failed to decompress 64K data: %v", err)
	}

	return buf.Bytes(), nil
}

// decompressRDP6 decompresses using RDP6 compression
func (cm *ShareDataCompressionManager) decompressRDP6(data []byte) ([]byte, error) {
	// RDP6 compression uses zlib with specific parameters
	reader := bytes.NewReader(data)
	zlibReader, err := zlib.NewReader(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to create zlib reader for RDP6 decompression: %v", err)
	}
	defer zlibReader.Close()

	var buf bytes.Buffer
	_, err = io.Copy(&buf, zlibReader)
	if err != nil {
		return nil, fmt.Errorf("failed to decompress RDP6 data: %v", err)
	}

	return buf.Bytes(), nil
}

// decompressRDP61 decompresses using RDP6.1 compression
func (cm *ShareDataCompressionManager) decompressRDP61(data []byte) ([]byte, error) {
	// RDP6.1 compression uses zlib with specific parameters
	reader := bytes.NewReader(data)
	zlibReader, err := zlib.NewReader(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to create zlib reader for RDP6.1 decompression: %v", err)
	}
	defer zlibReader.Close()

	var buf bytes.Buffer
	_, err = io.Copy(&buf, zlibReader)
	if err != nil {
		return nil, fmt.Errorf("failed to decompress RDP6.1 data: %v", err)
	}

	return buf.Bytes(), nil
}

// Compress compresses share data based on compression type
func (cm *ShareDataCompressionManager) Compress(data []byte, comprType uint8) ([]byte, error) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	var compressed []byte
	var err error

	switch comprType {
	case PACKET_COMPR_TYPE_8K:
		compressed, err = cm.compress8K(data)
	case PACKET_COMPR_TYPE_64K:
		compressed, err = cm.compress64K(data)
	case PACKET_COMPR_TYPE_RDP6:
		compressed, err = cm.compressRDP6(data)
	case PACKET_COMPR_TYPE_RDP61:
		compressed, err = cm.compressRDP61(data)
	default:
		// No compression
		return data, nil
	}

	if err != nil {
		cm.stats.Errors++
		return nil, fmt.Errorf("failed to compress share data (type %d): %v", comprType, err)
	}

	// Only return compressed data if it's actually smaller
	if len(compressed) < len(data) {
		glog.Debugf("Share data compression (type %d): %d -> %d bytes (%.1f%% reduction)",
			comprType, len(data), len(compressed),
			float64(len(data)-len(compressed))/float64(len(data))*100)
		return compressed, nil
	}

	// Return uncompressed data if compression doesn't help
	return data, nil
}

// compress8K compresses using 8KB history buffer
func (cm *ShareDataCompressionManager) compress8K(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	writer, err := zlib.NewWriterLevel(&buf, zlib.BestCompression)
	if err != nil {
		return nil, fmt.Errorf("failed to create zlib writer for 8K compression: %v", err)
	}

	_, err = writer.Write(data)
	if err != nil {
		writer.Close()
		return nil, fmt.Errorf("failed to compress 8K data: %v", err)
	}

	err = writer.Close()
	if err != nil {
		return nil, fmt.Errorf("failed to close zlib writer: %v", err)
	}

	return buf.Bytes(), nil
}

// compress64K compresses using 64KB history buffer
func (cm *ShareDataCompressionManager) compress64K(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	writer, err := zlib.NewWriterLevel(&buf, zlib.BestCompression)
	if err != nil {
		return nil, fmt.Errorf("failed to create zlib writer for 64K compression: %v", err)
	}

	_, err = writer.Write(data)
	if err != nil {
		writer.Close()
		return nil, fmt.Errorf("failed to compress 64K data: %v", err)
	}

	err = writer.Close()
	if err != nil {
		return nil, fmt.Errorf("failed to close zlib writer: %v", err)
	}

	return buf.Bytes(), nil
}

// compressRDP6 compresses using RDP6 compression
func (cm *ShareDataCompressionManager) compressRDP6(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	writer, err := zlib.NewWriterLevel(&buf, zlib.BestCompression)
	if err != nil {
		return nil, fmt.Errorf("failed to create zlib writer for RDP6 compression: %v", err)
	}

	_, err = writer.Write(data)
	if err != nil {
		writer.Close()
		return nil, fmt.Errorf("failed to compress RDP6 data: %v", err)
	}

	err = writer.Close()
	if err != nil {
		return nil, fmt.Errorf("failed to close zlib writer: %v", err)
	}

	return buf.Bytes(), nil
}

// compressRDP61 compresses using RDP6.1 compression
func (cm *ShareDataCompressionManager) compressRDP61(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	writer, err := zlib.NewWriterLevel(&buf, zlib.BestCompression)
	if err != nil {
		return nil, fmt.Errorf("failed to create zlib writer for RDP6.1 compression: %v", err)
	}

	_, err = writer.Write(data)
	if err != nil {
		writer.Close()
		return nil, fmt.Errorf("failed to compress RDP6.1 data: %v", err)
	}

	err = writer.Close()
	if err != nil {
		return nil, fmt.Errorf("failed to close zlib writer: %v", err)
	}

	return buf.Bytes(), nil
}

// GetStats returns compression statistics
func (cm *ShareDataCompressionManager) GetStats() *CompressionStats {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	stats := *cm.stats // Copy to avoid race conditions
	return &stats
}

// ResetStats resets compression statistics
func (cm *ShareDataCompressionManager) ResetStats() {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	cm.stats = &CompressionStats{}
}

// TsShareDataHeader
// https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-rdpbcgr/4b5d4c0d-a657-41e9-9c69-d58632f46d31
type TsShareDataHeader struct {
	SharedId           uint32
	Padding1           uint8
	StreamId           uint8
	UncompressedLength uint16
	PDUType2           uint8

	//https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-rdpbcgr/9355a663-ef22-4431-afeb-d72ac68f25fd
	CompressedType   uint8
	CompressedLength uint16
	compressor       *ShareDataCompressionManager
}

// NewTsShareDataHeader creates a new share data header with compression support
func NewTsShareDataHeader() *TsShareDataHeader {
	return &TsShareDataHeader{
		compressor: NewShareDataCompressionManager(),
	}
}

func (h *TsShareDataHeader) Read(r io.Reader) {
	core.ReadLE(r, h)
	glog.Debugf("[!] compressedType: %x", h.CompressedType)

	// Check if compression is used
	if h.CompressedType&PACKET_COMPRESSED != 0 {
		glog.Debugf("Share data compression detected (type: %d), will decompress data", h.CompressedType&PACKET_COMPRESSED)
		// Compression is handled by the caller after reading the header
		// The actual decompression happens when processing the data
	}
}

// ReadCompressedData reads and decompresses share data if compression is used
func (h *TsShareDataHeader) ReadCompressedData(r io.Reader) ([]byte, error) {
	if h.CompressedType&PACKET_COMPRESSED == 0 {
		// No compression, read data as-is
		return core.ReadBytes(r, int(h.CompressedLength)), nil
	}

	// Read compressed data
	compressedData := core.ReadBytes(r, int(h.CompressedLength))

	// Get compression type
	comprType := h.CompressedType & 0x03 // Lower 2 bits contain compression type

	// Decompress the data
	decompressedData, err := h.compressor.Decompress(compressedData, comprType)
	if err != nil {
		return nil, fmt.Errorf("failed to decompress share data: %v", err)
	}

	return decompressedData, nil
}

// Write writes the share data header
func (h *TsShareDataHeader) Write(w io.Writer) {
	core.WriteLE(w, h)
}

// WriteCompressedData writes and compresses share data
func (h *TsShareDataHeader) WriteCompressedData(w io.Writer, data []byte) error {
	if h.CompressedType&PACKET_COMPRESSED == 0 {
		// No compression, write data as-is
		_, err := w.Write(data)
		return err
	}

	// Get compression type
	comprType := h.CompressedType & 0x03 // Lower 2 bits contain compression type

	// Compress the data
	compressedData, err := h.compressor.Compress(data, comprType)
	if err != nil {
		return fmt.Errorf("failed to compress share data: %v", err)
	}

	// Update compressed length
	h.CompressedLength = uint16(len(compressedData))

	// Write compressed data
	_, err = w.Write(compressedData)
	return err
}

// SetCompression enables or disables compression with specified type
func (h *TsShareDataHeader) SetCompression(enabled bool, comprType uint8) {
	if enabled {
		h.CompressedType |= PACKET_COMPRESSED
		h.CompressedType = (h.CompressedType & 0xFC) | (comprType & 0x03) // Set lower 2 bits
	} else {
		h.CompressedType &^= PACKET_COMPRESSED
	}
}

// IsCompressed returns true if compression is enabled
func (h *TsShareDataHeader) IsCompressed() bool {
	return h.CompressedType&PACKET_COMPRESSED != 0
}

// GetCompressionType returns the compression type
func (h *TsShareDataHeader) GetCompressionType() uint8 {
	return h.CompressedType & 0x03
}

// GetCompressionStats returns compression statistics
func (h *TsShareDataHeader) GetCompressionStats() *CompressionStats {
	if h.compressor == nil {
		return &CompressionStats{}
	}
	return h.compressor.GetStats()
}
