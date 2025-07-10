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

// update code
const (
	FASTPATH_UPDATETYPE_ORDERS        = 0x0
	FASTPATH_UPDATETYPE_BITMAP        = 0x1
	FASTPATH_UPDATETYPE_PALETTE       = 0x2
	FASTPATH_UPDATETYPE_SYNCHRONIZE   = 0x3
	FASTPATH_UPDATETYPE_SURFCMDS      = 0x4
	FASTPATH_UPDATETYPE_PTR_NULL      = 0x5
	FASTPATH_UPDATETYPE_PTR_DEFAULT   = 0x6
	FASTPATH_UPDATETYPE_PTR_POSITION  = 0x8
	FASTPATH_UPDATETYPE_COLOR         = 0x9
	FASTPATH_UPDATETYPE_CACHED        = 0xA
	FASTPATH_UPDATETYPE_POINTER       = 0xB
	FASTPATH_UPDATETYPE_LARGE_POINTER = 0xC
)

// fragmentation
const (
	FASTPATH_FRAGMENT_SINGLE = 0x0
	FASTPATH_FRAGMENT_LAST   = 0x1
	FASTPATH_FRAGMENT_FIRST  = 0x2
	FASTPATH_FRAGMENT_NEXT   = 0x3
)

// compression
const (
	FASTPATH_OUTPUT_COMPRESSION_USED = 0x2
)

// FastPathCompressionManager handles RDP6.1 compression for FastPath
type FastPathCompressionManager struct {
	history    []byte
	maxHistory int
	mutex      sync.Mutex
	stats      *CompressionStats
}

// CompressionStats tracks compression performance
type CompressionStats struct {
	TotalCompressed   int64
	TotalUncompressed int64
	CompressionRatio  float64
	DecompressionTime int64 // nanoseconds
	Errors            int64
}

// NewFastPathCompressionManager creates a new FastPath compression manager
func NewFastPathCompressionManager() *FastPathCompressionManager {
	return &FastPathCompressionManager{
		history:    make([]byte, 0, 65536), // 64KB history buffer
		maxHistory: 65536,
		stats:      &CompressionStats{},
	}
}

// Decompress decompresses FastPath data using RDP6.1 compression
func (cm *FastPathCompressionManager) Decompress(data []byte) ([]byte, error) {
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

	// RDP6.1 compression uses zlib with specific parameters
	reader := bytes.NewReader(data)
	zlibReader, err := zlib.NewReader(reader)
	if err != nil {
		cm.stats.Errors++
		return nil, fmt.Errorf("failed to create zlib reader for FastPath decompression: %v", err)
	}
	defer zlibReader.Close()

	var buf bytes.Buffer
	_, err = io.Copy(&buf, zlibReader)
	if err != nil {
		cm.stats.Errors++
		return nil, fmt.Errorf("failed to decompress FastPath data: %v", err)
	}

	decompressed := buf.Bytes()

	// Update statistics
	cm.stats.TotalCompressed += int64(len(data))
	cm.stats.TotalUncompressed += int64(len(decompressed))
	if cm.stats.TotalUncompressed > 0 {
		cm.stats.CompressionRatio = float64(cm.stats.TotalCompressed) / float64(cm.stats.TotalUncompressed)
	}

	glog.Debugf("FastPath decompression: %d -> %d bytes (%.1f%% reduction)",
		len(data), len(decompressed),
		float64(len(data)-len(decompressed))/float64(len(data))*100)

	return decompressed, nil
}

// Compress compresses FastPath data using RDP6.1 compression
func (cm *FastPathCompressionManager) Compress(data []byte) ([]byte, error) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	// Use zlib with RDP6.1 parameters
	var buf bytes.Buffer
	writer, err := zlib.NewWriterLevel(&buf, zlib.BestCompression)
	if err != nil {
		cm.stats.Errors++
		return nil, fmt.Errorf("failed to create zlib writer for FastPath compression: %v", err)
	}

	_, err = writer.Write(data)
	if err != nil {
		cm.stats.Errors++
		writer.Close()
		return nil, fmt.Errorf("failed to compress FastPath data: %v", err)
	}

	err = writer.Close()
	if err != nil {
		cm.stats.Errors++
		return nil, fmt.Errorf("failed to close zlib writer: %v", err)
	}

	compressed := buf.Bytes()

	// Only return compressed data if it's actually smaller
	if len(compressed) < len(data) {
		glog.Debugf("FastPath compression: %d -> %d bytes (%.1f%% reduction)",
			len(data), len(compressed),
			float64(len(data)-len(compressed))/float64(len(data))*100)
		return compressed, nil
	}

	// Return uncompressed data if compression doesn't help
	return data, nil
}

// GetStats returns compression statistics
func (cm *FastPathCompressionManager) GetStats() *CompressionStats {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	stats := *cm.stats // Copy to avoid race conditions
	return &stats
}

// ResetStats resets compression statistics
func (cm *FastPathCompressionManager) ResetStats() {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	cm.stats = &CompressionStats{}
}

// FpOutputHeader
// https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-rdpbcgr/a1c4caa8-00ed-45bb-a06e-5177473766d3
type FpOutputHeader struct {
	UpdateCode    uint8
	Fragmentation uint8
	Compression   uint8
	compressor    *FastPathCompressionManager
}

// NewFpOutputHeader creates a new FastPath output header with compression support
func NewFpOutputHeader() *FpOutputHeader {
	return &FpOutputHeader{
		compressor: NewFastPathCompressionManager(),
	}
}

func (h *FpOutputHeader) Read(r io.Reader) {
	var updateHeader uint8
	core.ReadLE(r, &updateHeader)
	glog.Debugf("fpOutputHeader: %x", updateHeader)
	h.UpdateCode = updateHeader & 0xF
	h.Fragmentation = (updateHeader >> 4) & 0x03
	h.Compression = (updateHeader >> 6) & 0x03
	glog.Debugf("fpOutputHeader: %+v", h)

	if h.Compression == FASTPATH_OUTPUT_COMPRESSION_USED {
		glog.Debugf("FastPath compression detected, will decompress data")
		// Compression is handled by the caller after reading the header
		// The actual decompression happens when processing the data
	}
}

// ReadCompressedData reads and decompresses FastPath data if compression is used
func (h *FpOutputHeader) ReadCompressedData(r io.Reader, length uint16) ([]byte, error) {
	if h.Compression != FASTPATH_OUTPUT_COMPRESSION_USED {
		// No compression, read data as-is
		return core.ReadBytes(r, int(length)), nil
	}

	// Read compressed data
	compressedData := core.ReadBytes(r, int(length))

	// Decompress the data
	decompressedData, err := h.compressor.Decompress(compressedData)
	if err != nil {
		return nil, fmt.Errorf("failed to decompress FastPath data: %v", err)
	}

	return decompressedData, nil
}

// Write writes the FastPath output header
func (h *FpOutputHeader) Write(w io.Writer) {
	updateHeader := h.UpdateCode | (h.Fragmentation << 4) | (h.Compression << 6)
	core.WriteLE(w, updateHeader)
}

// WriteCompressedData writes and compresses FastPath data
func (h *FpOutputHeader) WriteCompressedData(w io.Writer, data []byte) error {
	if h.Compression != FASTPATH_OUTPUT_COMPRESSION_USED {
		// No compression, write data as-is
		_, err := w.Write(data)
		return err
	}

	// Compress the data
	compressedData, err := h.compressor.Compress(data)
	if err != nil {
		return fmt.Errorf("failed to compress FastPath data: %v", err)
	}

	// Write compressed data
	_, err = w.Write(compressedData)
	return err
}

// SetCompression enables or disables compression
func (h *FpOutputHeader) SetCompression(enabled bool) {
	if enabled {
		h.Compression = FASTPATH_OUTPUT_COMPRESSION_USED
	} else {
		h.Compression = 0
	}
}

// IsCompressed returns true if compression is enabled
func (h *FpOutputHeader) IsCompressed() bool {
	return h.Compression == FASTPATH_OUTPUT_COMPRESSION_USED
}

// GetCompressionStats returns compression statistics
func (h *FpOutputHeader) GetCompressionStats() *CompressionStats {
	if h.compressor == nil {
		return &CompressionStats{}
	}
	return h.compressor.GetStats()
}
