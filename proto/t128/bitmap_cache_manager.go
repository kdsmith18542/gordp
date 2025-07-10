package t128

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"io"
	"sync"

	"github.com/kdsmith18542/gordp/glog"
)

// BitmapCacheManager manages multiple bitmap caches and provides compression
type BitmapCacheManager struct {
	caches     [3]*BitmapCache // Three bitmap caches as per RDP spec
	mutex      sync.RWMutex
	compressor *CompressionManager
}

// CompressionManager handles RDP compression
type CompressionManager struct {
	history    []byte
	maxHistory int
	mutex      sync.Mutex
}

// NewBitmapCacheManager creates a new bitmap cache manager
func NewBitmapCacheManager() *BitmapCacheManager {
	manager := &BitmapCacheManager{
		compressor: NewCompressionManager(),
	}

	// Initialize three bitmap caches with different sizes
	// These sizes are typical for RDP bitmap caching
	manager.caches[0] = NewBitmapCache(600) // Small bitmaps
	manager.caches[1] = NewBitmapCache(300) // Medium bitmaps
	manager.caches[2] = NewBitmapCache(100) // Large bitmaps

	glog.Debugf("Bitmap cache manager initialized with 3 caches")
	return manager
}

// NewCompressionManager creates a new compression manager
func NewCompressionManager() *CompressionManager {
	return &CompressionManager{
		history:    make([]byte, 0, 65536), // 64KB history buffer
		maxHistory: 65536,
	}
}

// GetCache returns the appropriate cache for the given bitmap size
func (bcm *BitmapCacheManager) GetCache(width, height uint16) *BitmapCache {
	bcm.mutex.RLock()
	defer bcm.mutex.RUnlock()

	// Determine cache based on bitmap size
	size := int(width) * int(height)

	if size <= 1024 { // Small bitmaps (32x32 or smaller)
		return bcm.caches[0]
	} else if size <= 16384 { // Medium bitmaps (128x128 or smaller)
		return bcm.caches[1]
	} else { // Large bitmaps
		return bcm.caches[2]
	}
}

// GetCacheIndex returns the cache index (0, 1, or 2) for the given bitmap size
func (bcm *BitmapCacheManager) GetCacheIndex(width, height uint16) uint8 {
	size := int(width) * int(height)

	if size <= 1024 {
		return 0
	} else if size <= 16384 {
		return 1
	} else {
		return 2
	}
}

// ProcessBitmap processes a bitmap and returns either cached data or new data
func (bcm *BitmapCacheManager) ProcessBitmap(data []byte, width, height, bpp uint16) ([]byte, bool, uint64, uint8, uint16) {
	// Generate cache key
	key := GenerateCacheKey(data, width, height, bpp)

	// Get appropriate cache
	cache := bcm.GetCache(width, height)
	cacheIndex := bcm.GetCacheIndex(width, height)

	// Try to get from cache
	if entry, found := cache.Get(key); found {
		glog.Debugf("Bitmap cache hit: cache=%d, key=%016X, size=%dx%d", cacheIndex, key, width, height)
		return entry.Data, true, key, cacheIndex, 0
	}

	// Not in cache, store it
	cache.Put(key, data, width, height, bpp)
	glog.Debugf("Bitmap cache miss: cache=%d, key=%016X, size=%dx%d", cacheIndex, key, width, height)

	// Return compressed data
	compressedData := bcm.compressor.Compress(data)
	return compressedData, false, key, cacheIndex, uint16(len(compressedData))
}

// Compress compresses data using RDP compression
func (cm *CompressionManager) Compress(data []byte) []byte {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	// Use zlib compression for RDP compression
	var buf bytes.Buffer
	writer, err := zlib.NewWriterLevel(&buf, zlib.BestCompression)
	if err != nil {
		glog.Errorf("Failed to create zlib writer: %v", err)
		return data // Return uncompressed data on error
	}

	_, err = writer.Write(data)
	if err != nil {
		glog.Errorf("Failed to compress data: %v", err)
		writer.Close()
		return data
	}

	err = writer.Close()
	if err != nil {
		glog.Errorf("Failed to close zlib writer: %v", err)
		return data
	}

	compressed := buf.Bytes()

	// Only return compressed data if it's actually smaller
	if len(compressed) < len(data) {
		glog.Debugf("Compression: %d -> %d bytes (%.1f%% reduction)",
			len(data), len(compressed),
			float64(len(data)-len(compressed))/float64(len(data))*100)
		return compressed
	}

	return data
}

// Decompress decompresses data using RDP compression
func (cm *CompressionManager) Decompress(data []byte) ([]byte, error) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	// Check if data is compressed (simple heuristic)
	if len(data) < 2 || data[0] != 0x78 || (data[1] != 0x9C && data[1] != 0xDA && data[1] != 0x01) {
		// Not zlib compressed, return as-is
		return data, nil
	}

	reader := bytes.NewReader(data)
	zlibReader, err := zlib.NewReader(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to create zlib reader: %v", err)
	}
	defer zlibReader.Close()

	var buf bytes.Buffer
	_, err = io.Copy(&buf, zlibReader)
	if err != nil {
		return nil, fmt.Errorf("failed to decompress data: %v", err)
	}

	return buf.Bytes(), nil
}

// GetCacheStats returns statistics about all caches
func (bcm *BitmapCacheManager) GetCacheStats() map[string]interface{} {
	bcm.mutex.RLock()
	defer bcm.mutex.RUnlock()

	stats := make(map[string]interface{})

	for i, cache := range bcm.caches {
		cacheName := fmt.Sprintf("cache_%d", i)
		stats[cacheName] = map[string]interface{}{
			"entries":     len(cache.Entries),
			"max_entries": cache.MaxEntries,
			"hit_count":   cache.HitCount,
			"miss_count":  cache.MissCount,
			"hit_rate":    float64(cache.HitCount) / float64(cache.HitCount+cache.MissCount) * 100,
		}
	}

	return stats
}

// ClearCache clears all bitmap caches
func (bcm *BitmapCacheManager) ClearCache() {
	bcm.mutex.Lock()
	defer bcm.mutex.Unlock()

	for i, cache := range bcm.caches {
		cache.Entries = make(map[uint64]*BitmapCacheEntry)
		cache.HitCount = 0
		cache.MissCount = 0
		glog.Debugf("Cleared bitmap cache %d", i)
	}
}

// OptimizeBitmapData optimizes bitmap data for transmission
func (bcm *BitmapCacheManager) OptimizeBitmapData(bitmapData *TsBitmapData) (*TsBitmapData, bool) {
	// Check if we can use cached data
	if len(bitmapData.BitmapDataStream) == 0 {
		return bitmapData, false
	}

	// Process the bitmap through cache manager
	optimizedData, cached, key, cacheIndex, _ := bcm.ProcessBitmap(
		bitmapData.BitmapDataStream,
		bitmapData.Width,
		bitmapData.Height,
		bitmapData.BitsPerPixel,
	)

	if cached {
		// Return cached bitmap data
		// Note: This would need to be handled differently in the actual implementation
		// as we need to return a different PDU type
		glog.Debugf("Using cached bitmap: key=%016X, cache=%d", key, cacheIndex)
		return bitmapData, true
	}

	// Update bitmap data with compressed data
	if len(optimizedData) != len(bitmapData.BitmapDataStream) {
		bitmapData.BitmapDataStream = optimizedData
		bitmapData.BitmapLength = uint16(len(optimizedData))
		bitmapData.Flags |= BITMAP_COMPRESSION
		glog.Debugf("Compressed bitmap: %d -> %d bytes", len(bitmapData.BitmapDataStream), len(optimizedData))
	}

	return bitmapData, false
}

// GetCachedBitmap retrieves a cached bitmap by cache ID and index
func (bcm *BitmapCacheManager) GetCachedBitmap(cacheId uint16, cacheIndex uint16, key1, key2 uint32) *TsBitmapData {
	bcm.mutex.RLock()
	defer bcm.mutex.RUnlock()

	// Convert cache ID to array index
	if cacheId >= 3 {
		glog.Warnf("Invalid cache ID: %d", cacheId)
		return nil
	}

	cache := bcm.caches[cacheId]
	if cache == nil {
		glog.Warnf("Cache %d not found", cacheId)
		return nil
	}

	// Reconstruct the key from key1 and key2
	key := uint64(key2)<<32 | uint64(key1)

	// Try to get from cache
	if entry, found := cache.Get(key); found {
		glog.Debugf("Retrieved cached bitmap: cache=%d, index=%d, key=%016X, size=%dx%d",
			cacheId, cacheIndex, key, entry.Width, entry.Height)

		return &TsBitmapData{
			DestLeft:         0, // Will be set by caller
			DestTop:          0, // Will be set by caller
			DestRight:        entry.Width,
			DestBottom:       entry.Height,
			Width:            entry.Width,
			Height:           entry.Height,
			BitsPerPixel:     entry.Bpp,
			Flags:            0,
			BitmapLength:     uint16(len(entry.Data)),
			BitmapDataStream: entry.Data,
		}
	}

	glog.Warnf("Cached bitmap not found: cache=%d, index=%d, key=%016X", cacheId, cacheIndex, key)
	return nil
}

// CreateCachedBitmapUpdate creates a cached bitmap update PDU
func (bcm *BitmapCacheManager) CreateCachedBitmapUpdate(bitmapData *TsBitmapData, key uint64, cacheIndex uint8) *TsFpUpdateCachedBitmap {
	return &TsFpUpdateCachedBitmap{
		UpdateType:       int16(FASTPATH_UPDATETYPE_CACHED),
		NumberRectangles: 1,
		Rectangles: []TsCachedBitmapData{
			{
				DestLeft:   bitmapData.DestLeft,
				DestTop:    bitmapData.DestTop,
				DestRight:  bitmapData.DestRight,
				DestBottom: bitmapData.DestBottom,
				CacheId:    cacheIndex,
				CacheIndex: 0, // Will be set by the server
				Key1:       uint32(key & 0xFFFFFFFF),
				Key2:       uint32(key >> 32),
			},
		},
	}
}
