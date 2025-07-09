package t128

import (
	"bytes"
	"fmt"
	"testing"
	"time"
)

func TestNewBitmapCacheManager(t *testing.T) {
	manager := NewBitmapCacheManager()
	if manager == nil {
		t.Fatal("NewBitmapCacheManager returned nil")
	}

	// Check that all three caches are initialized
	for i, cache := range manager.caches {
		if cache == nil {
			t.Fatalf("Cache %d is nil", i)
		}
	}

	// Check cache sizes
	expectedSizes := []int{600, 300, 100}
	for i, expectedSize := range expectedSizes {
		if manager.caches[i].MaxEntries != expectedSize {
			t.Errorf("Cache %d expected size %d, got %d", i, expectedSize, manager.caches[i].MaxEntries)
		}
	}
}

func TestBitmapCacheManager_GetCacheIndex(t *testing.T) {
	manager := NewBitmapCacheManager()

	tests := []struct {
		width, height uint16
		expected      uint8
	}{
		{32, 32, 0},   // 1024 pixels - small cache
		{64, 16, 0},   // 1024 pixels - small cache
		{128, 128, 1}, // 16384 pixels - medium cache
		{256, 64, 1},  // 16384 pixels - medium cache
		{256, 256, 2}, // 65536 pixels - large cache
		{512, 128, 2}, // 65536 pixels - large cache
	}

	for _, test := range tests {
		result := manager.GetCacheIndex(test.width, test.height)
		if result != test.expected {
			t.Errorf("GetCacheIndex(%d, %d) = %d, expected %d",
				test.width, test.height, result, test.expected)
		}
	}
}

func TestBitmapCacheManager_ProcessBitmap(t *testing.T) {
	manager := NewBitmapCacheManager()

	// Test data
	testData := []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
	width, height, bpp := uint16(2), uint16(2), uint16(16)

	// First call should be a miss
	data, cached, key, cacheIndex, compressedLength := manager.ProcessBitmap(testData, width, height, bpp)
	if cached {
		t.Error("First call should not be cached")
	}
	if len(data) == 0 {
		t.Error("ProcessBitmap returned empty data")
	}
	if key == 0 {
		t.Error("ProcessBitmap returned zero key")
	}
	if cacheIndex != 0 {
		t.Errorf("Expected cache index 0 for 2x2 bitmap, got %d", cacheIndex)
	}

	// Second call with same data should be a hit
	_, cached2, key2, cacheIndex2, _ := manager.ProcessBitmap(testData, width, height, bpp)
	if !cached2 {
		t.Error("Second call should be cached")
	}
	if key != key2 {
		t.Error("Keys should be the same for identical data")
	}
	if cacheIndex != cacheIndex2 {
		t.Error("Cache indices should be the same for identical data")
	}

	// Check that compressed length is reasonable
	if compressedLength > uint16(len(testData)) {
		t.Errorf("Compressed length %d should not be larger than original %d",
			compressedLength, len(testData))
	}
}

func TestCompressionManager_Compress(t *testing.T) {
	compressor := NewCompressionManager()

	// Test with compressible data (repeating pattern)
	compressibleData := bytes.Repeat([]byte{0x00, 0xFF, 0x00, 0xFF}, 1000)
	compressed := compressor.Compress(compressibleData)

	if len(compressed) == 0 {
		t.Error("Compression returned empty data")
	}

	// Test with incompressible data (random-like)
	incompressibleData := make([]byte, 1000)
	for i := range incompressibleData {
		incompressibleData[i] = byte(i % 256)
	}
	compressed2 := compressor.Compress(incompressibleData)

	// Incompressible data might not compress well, but should still return data
	if len(compressed2) == 0 {
		t.Error("Compression of incompressible data returned empty data")
	}
}

func TestCompressionManager_Decompress(t *testing.T) {
	compressor := NewCompressionManager()

	// Test data
	originalData := []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}

	// Compress the data
	compressed := compressor.Compress(originalData)

	// Decompress the data
	decompressed, err := compressor.Decompress(compressed)
	if err != nil {
		t.Fatalf("Decompress failed: %v", err)
	}

	// Check that decompressed data matches original
	if !bytes.Equal(originalData, decompressed) {
		t.Errorf("Decompressed data doesn't match original: %v vs %v",
			originalData, decompressed)
	}

	// Test decompressing uncompressed data
	decompressed2, err := compressor.Decompress(originalData)
	if err != nil {
		t.Fatalf("Decompress of uncompressed data failed: %v", err)
	}
	if !bytes.Equal(originalData, decompressed2) {
		t.Error("Decompressing uncompressed data should return original data")
	}
}

func TestBitmapCache_GetPut(t *testing.T) {
	cache := NewBitmapCache(10)

	// Test data
	testData := []byte{0x01, 0x02, 0x03, 0x04}
	key := uint64(12345)
	width, height, bpp := uint16(2), uint16(2), uint16(16)

	// Put data in cache
	cache.Put(key, testData, width, height, bpp)

	// Get data from cache
	entry, found := cache.Get(key)
	if !found {
		t.Error("Data not found in cache after Put")
	}

	if !bytes.Equal(entry.Data, testData) {
		t.Error("Retrieved data doesn't match original")
	}

	if entry.Width != width || entry.Height != height || entry.Bpp != bpp {
		t.Error("Retrieved entry metadata doesn't match original")
	}

	// Test cache miss
	_, found = cache.Get(99999)
	if found {
		t.Error("Non-existent key should not be found")
	}
}

func TestBitmapCache_Eviction(t *testing.T) {
	cache := NewBitmapCache(2) // Small cache for testing eviction

	// Add three entries to trigger eviction, with a small delay to ensure unique timestamps
	cache.Put(1, []byte{0x01}, 1, 1, 16)
	time.Sleep(2 * time.Millisecond)
	cache.Put(2, []byte{0x02}, 1, 1, 16)
	time.Sleep(2 * time.Millisecond)
	cache.Put(3, []byte{0x03}, 1, 1, 16)

	// First entry should be evicted
	_, found := cache.Get(1)
	if found {
		t.Error("First entry should have been evicted")
	}

	// Second and third entries should still be there
	_, found = cache.Get(2)
	if !found {
		t.Error("Second entry should still be in cache")
	}

	_, found = cache.Get(3)
	if !found {
		t.Error("Third entry should still be in cache")
	}

	// Check cache size
	if len(cache.Entries) != 2 {
		t.Errorf("Cache should have 2 entries, got %d", len(cache.Entries))
	}
}

func TestGenerateCacheKey(t *testing.T) {
	// Test that same data generates same key
	data1 := []byte{0x01, 0x02, 0x03, 0x04}
	key1 := GenerateCacheKey(data1, 2, 2, 16)
	key2 := GenerateCacheKey(data1, 2, 2, 16)

	if key1 != key2 {
		t.Error("Same data should generate same cache key")
	}

	// Test that different data generates different keys
	data2 := []byte{0x05, 0x06, 0x07, 0x08}
	key3 := GenerateCacheKey(data2, 2, 2, 16)

	if key1 == key3 {
		t.Error("Different data should generate different cache keys")
	}

	// Test that different dimensions generate different keys
	key4 := GenerateCacheKey(data1, 4, 4, 16)
	if key1 == key4 {
		t.Error("Different dimensions should generate different cache keys")
	}
}

func TestBitmapCacheManager_GetCacheStats(t *testing.T) {
	manager := NewBitmapCacheManager()

	// Add some data to generate stats
	testData := []byte{0x01, 0x02, 0x03, 0x04}
	manager.ProcessBitmap(testData, 2, 2, 16)
	manager.ProcessBitmap(testData, 2, 2, 16) // This should be a hit

	stats := manager.GetCacheStats()

	// Check that stats are returned for all caches
	for i := 0; i < 3; i++ {
		cacheName := fmt.Sprintf("cache_%d", i)
		cacheStats, exists := stats[cacheName]
		if !exists {
			t.Errorf("Stats for %s not found", cacheName)
			continue
		}

		statsMap, ok := cacheStats.(map[string]interface{})
		if !ok {
			t.Errorf("Cache stats for %s is not a map", cacheName)
			continue
		}

		// Check required fields
		requiredFields := []string{"entries", "max_entries", "hit_count", "miss_count", "hit_rate"}
		for _, field := range requiredFields {
			if _, exists := statsMap[field]; !exists {
				t.Errorf("Field %s missing from cache stats", field)
			}
		}
	}
}

func TestBitmapCacheManager_ClearCache(t *testing.T) {
	manager := NewBitmapCacheManager()

	// Add some data
	testData := []byte{0x01, 0x02, 0x03, 0x04}
	manager.ProcessBitmap(testData, 2, 2, 16)

	// Check that data is in cache
	stats := manager.GetCacheStats()
	cache0Stats := stats["cache_0"].(map[string]interface{})
	if cache0Stats["entries"].(int) == 0 {
		t.Error("Cache should have entries after adding data")
	}

	// Clear cache
	manager.ClearCache()

	// Check that cache is empty
	stats = manager.GetCacheStats()
	cache0Stats = stats["cache_0"].(map[string]interface{})
	if cache0Stats["entries"].(int) != 0 {
		t.Error("Cache should be empty after clearing")
	}

	if cache0Stats["hit_count"].(int64) != 0 {
		t.Error("Hit count should be reset after clearing")
	}

	if cache0Stats["miss_count"].(int64) != 0 {
		t.Error("Miss count should be reset after clearing")
	}
}
