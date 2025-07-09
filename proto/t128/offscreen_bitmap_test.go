package t128

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOffscreenBitmapCache_GetPut(t *testing.T) {
	cache := NewOffscreenBitmapCache(10, 100)

	// Test putting and getting an entry
	data := []byte{0x01, 0x02, 0x03, 0x04}
	id := cache.AddEntry(data, 2, 2, 16)
	assert.NotEqual(t, uint16(0), id)

	// Test getting the entry
	entry := cache.GetEntry(id)
	assert.NotNil(t, entry)
	assert.Equal(t, data, entry.Data)
	assert.Equal(t, uint16(2), entry.Width)
	assert.Equal(t, uint16(2), entry.Height)
	assert.Equal(t, uint16(16), entry.Bpp)
}

func TestOffscreenBitmapCache_Eviction(t *testing.T) {
	cache := NewOffscreenBitmapCache(10, 2) // Capacity of 2 entries

	data1 := []byte{0x01, 0x02, 0x03, 0x04}
	data2 := []byte{0x05, 0x06, 0x07, 0x08}
	data3 := []byte{0x09, 0x0A, 0x0B, 0x0C}

	id1 := cache.AddEntry(data1, 2, 2, 16)
	id2 := cache.AddEntry(data2, 2, 2, 16)
	id3 := cache.AddEntry(data3, 2, 2, 16)

	// When we add the third entry, the first entry should be evicted
	// Check that only 2 entries remain
	entry1 := cache.GetEntry(id1)
	entry2 := cache.GetEntry(id2)
	entry3 := cache.GetEntry(id3)

	// One of the first two entries should be evicted, the third should exist
	// Since we can't predict which one gets evicted due to timing, we check that:
	// 1. Only 2 entries exist in total
	// 2. The third entry definitely exists
	// 3. One of the first two entries is missing
	count := 0
	if entry1 != nil {
		count++
	}
	if entry2 != nil {
		count++
	}
	if entry3 != nil {
		count++
	}

	assert.Equal(t, 2, count, "Should have exactly 2 entries after eviction")
	assert.NotNil(t, entry3, "Third entry should exist")
	assert.True(t, entry1 == nil || entry2 == nil, "One of the first two entries should be evicted")
}

func TestOffscreenBitmapCache_Clear(t *testing.T) {
	cache := NewOffscreenBitmapCache(10, 100)

	// Add some entries
	data := []byte{0x01, 0x02, 0x03, 0x04}
	id := cache.AddEntry(data, 2, 2, 16)

	// Verify entry exists
	entry := cache.GetEntry(id)
	assert.NotNil(t, entry)

	// Clear cache
	cache.Clear()

	// Verify entry is gone
	entry = cache.GetEntry(id)
	assert.Nil(t, entry)
}

func TestOffscreenBitmapManager_ProcessBitmap(t *testing.T) {
	manager := NewOffscreenBitmapManager(100, 10)

	// Create bitmap data
	data := &TsOffscreenBitmapData{
		CacheId:    1,
		CacheIndex: 0,
		Width:      100,
		Height:     100,
		Bpp:        32,
		Data:       make([]byte, 100*100*4),
	}

	// Process the bitmap
	id := manager.ProcessOffscreenBitmap(data)
	assert.NotEqual(t, uint16(0), id)

	// Get the bitmap
	entry := manager.GetOffscreenBitmap(id)
	assert.NotNil(t, entry)
	assert.Equal(t, uint16(100), entry.Width)
	assert.Equal(t, uint16(100), entry.Height)
	assert.Equal(t, uint16(32), entry.Bpp)
}

func TestOffscreenBitmapManager_RemoveBitmap(t *testing.T) {
	manager := NewOffscreenBitmapManager(100, 10)

	// Create bitmap data
	data := &TsOffscreenBitmapData{
		CacheId:    1,
		CacheIndex: 0,
		Width:      100,
		Height:     100,
		Bpp:        32,
		Data:       make([]byte, 100*100*4),
	}

	// Process the bitmap
	id := manager.ProcessOffscreenBitmap(data)
	assert.NotEqual(t, uint16(0), id)

	// Verify bitmap exists
	entry := manager.GetOffscreenBitmap(id)
	assert.NotNil(t, entry)

	// Remove bitmap
	removed := manager.RemoveOffscreenBitmap(id)
	assert.True(t, removed)

	// Verify bitmap is gone
	entry = manager.GetOffscreenBitmap(id)
	assert.Nil(t, entry)
}

func TestOffscreenBitmapManager_GetStats(t *testing.T) {
	manager := NewOffscreenBitmapManager(100, 10)

	// Create two different bitmap data
	data1 := &TsOffscreenBitmapData{
		CacheId:    1,
		CacheIndex: 0,
		Width:      100,
		Height:     100,
		Bpp:        32,
		Data:       make([]byte, 100*100*4),
	}
	data1.Data[0] = 0x01 // Make it different

	data2 := &TsOffscreenBitmapData{
		CacheId:    2,
		CacheIndex: 1,
		Width:      100,
		Height:     100,
		Bpp:        32,
		Data:       make([]byte, 100*100*4),
	}
	data2.Data[0] = 0x02 // Make it different

	// Process bitmaps
	manager.ProcessOffscreenBitmap(data1)
	manager.ProcessOffscreenBitmap(data2)

	count, maxEntries := manager.GetStats()
	assert.Equal(t, uint16(2), count)
	assert.Equal(t, uint16(10), maxEntries)
}

func TestOffscreenBitmapOrder_ReadWrite(t *testing.T) {
	order := &TsOffscreenBitmapOrder{
		CacheId:    1,
		CacheIndex: 2,
		DestLeft:   10,
		DestTop:    20,
		DestRight:  110,
		DestBottom: 120,
		SourceLeft: 0,
		SourceTop:  0,
	}

	// Test that the order can be created and accessed
	assert.Equal(t, uint16(1), order.CacheId)
	assert.Equal(t, uint16(2), order.CacheIndex)
	assert.Equal(t, uint16(10), order.DestLeft)
	assert.Equal(t, uint16(20), order.DestTop)
	assert.Equal(t, uint16(110), order.DestRight)
	assert.Equal(t, uint16(120), order.DestBottom)
}

func TestOffscreenBitmapCachePdu_ReadWrite(t *testing.T) {
	pdu := &TsOffscreenBitmapCachePdu{
		Data: TsOffscreenBitmapData{
			CacheId:    1,
			CacheIndex: 2,
			Width:      100,
			Height:     100,
			Bpp:        32,
			Data:       make([]byte, 100*100*4),
		},
	}

	// Test that the PDU can be created and accessed
	assert.Equal(t, uint16(1), pdu.Data.CacheId)
	assert.Equal(t, uint16(2), pdu.Data.CacheIndex)
	assert.Equal(t, uint16(100), pdu.Data.Width)
	assert.Equal(t, uint16(100), pdu.Data.Height)
}

func BenchmarkOffscreenBitmapCache_Put(b *testing.B) {
	cache := NewOffscreenBitmapCache(100, 1000)
	data := make([]byte, 1024)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.AddEntry(data, 32, 32, 32)
	}
}

func BenchmarkOffscreenBitmapCache_Get(b *testing.B) {
	cache := NewOffscreenBitmapCache(100, 1000)
	data := make([]byte, 1024)
	id := cache.AddEntry(data, 32, 32, 32)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.GetEntry(id)
	}
}

func BenchmarkOffscreenBitmapManager_ProcessBitmap(b *testing.B) {
	manager := NewOffscreenBitmapManager(100, 1000)
	data := &TsOffscreenBitmapData{
		CacheId:    1,
		CacheIndex: 0,
		Width:      100,
		Height:     100,
		Bpp:        32,
		Data:       make([]byte, 100*100*4),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manager.ProcessOffscreenBitmap(data)
	}
}
