package t128

import (
	"bytes"
	"crypto/md5"
	"io"
	"sync"

	"github.com/kdsmith18542/gordp/core"
	"github.com/kdsmith18542/gordp/glog"
)

// Offscreen Bitmap Support Level
const (
	OFFSCREEN_SUPPORT_LEVEL_NONE    = 0x00000000
	OFFSCREEN_SUPPORT_LEVEL_DEFAULT = 0x00000001
	OFFSCREEN_SUPPORT_LEVEL_CACHE   = 0x00000002
	OFFSCREEN_SUPPORT_LEVEL_FULL    = 0x00000003
)

// Offscreen Bitmap Cache Entry
type OffscreenCacheEntry struct {
	ID       uint16
	Data     []byte
	Width    uint16
	Height   uint16
	Bpp      uint16
	Hash     [16]byte
	LastUsed int64
}

// Offscreen Bitmap Cache
type OffscreenBitmapCache struct {
	mu         sync.RWMutex
	entries    map[uint16]*OffscreenCacheEntry
	maxSize    uint16
	maxEntries uint16
	nextID     uint16
}

// New Offscreen Bitmap Cache
func NewOffscreenBitmapCache(maxSize, maxEntries uint16) *OffscreenBitmapCache {
	return &OffscreenBitmapCache{
		entries:    make(map[uint16]*OffscreenCacheEntry),
		maxSize:    maxSize,
		maxEntries: maxEntries,
		nextID:     1,
	}
}

// Add Entry to Cache
func (c *OffscreenBitmapCache) AddEntry(data []byte, width, height, bpp uint16) uint16 {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Calculate hash
	hash := md5.Sum(data)

	// Check if entry already exists
	for id, entry := range c.entries {
		if bytes.Equal(entry.Hash[:], hash[:]) {
			entry.LastUsed = core.GetCurrentTimestamp()
			return id
		}
	}

	// Check cache limits
	if uint16(len(c.entries)) >= c.maxEntries {
		c.evictOldest()
	}

	// Create new entry
	id := c.nextID
	c.nextID++

	entry := &OffscreenCacheEntry{
		ID:       id,
		Data:     make([]byte, len(data)),
		Width:    width,
		Height:   height,
		Bpp:      bpp,
		Hash:     hash,
		LastUsed: core.GetCurrentTimestamp(),
	}
	copy(entry.Data, data)

	c.entries[id] = entry
	glog.Debugf("Added offscreen cache entry: ID=%d, size=%d bytes", id, len(data))

	return id
}

// Get Entry from Cache
func (c *OffscreenBitmapCache) GetEntry(id uint16) *OffscreenCacheEntry {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.entries[id]
	if !exists {
		return nil
	}

	entry.LastUsed = core.GetCurrentTimestamp()
	return entry
}

// Remove Entry from Cache
func (c *OffscreenBitmapCache) RemoveEntry(id uint16) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.entries[id]; exists {
		delete(c.entries, id)
		glog.Debugf("Removed offscreen cache entry: ID=%d", id)
		return true
	}
	return false
}

// Evict Oldest Entry
func (c *OffscreenBitmapCache) evictOldest() {
	var oldestID uint16
	var oldestTime int64 = 9223372036854775807 // Max int64

	for id, entry := range c.entries {
		if entry.LastUsed < oldestTime {
			oldestTime = entry.LastUsed
			oldestID = id
		}
	}

	if oldestID > 0 {
		delete(c.entries, oldestID)
		glog.Debugf("Evicted oldest offscreen cache entry: ID=%d", oldestID)
	}
}

// Clear Cache
func (c *OffscreenBitmapCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries = make(map[uint16]*OffscreenCacheEntry)
	c.nextID = 1
	glog.Debugf("Cleared offscreen bitmap cache")
}

// Get Cache Statistics
func (c *OffscreenBitmapCache) GetStats() (count, maxEntries uint16) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return uint16(len(c.entries)), c.maxEntries
}

// Offscreen Bitmap Data
// https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-rdpegdi/2c3c3c41-1d54-4254-bb62-bc082a3c1f10
type TsOffscreenBitmapData struct {
	CacheId    uint16
	CacheIndex uint16
	Width      uint16
	Height     uint16
	Bpp        uint16
	Data       []byte
}

func (d *TsOffscreenBitmapData) Read(r io.Reader) {
	core.ReadLE(r, &d.CacheId)
	core.ReadLE(r, &d.CacheIndex)
	core.ReadLE(r, &d.Width)
	core.ReadLE(r, &d.Height)
	core.ReadLE(r, &d.Bpp)

	// Read bitmap data
	dataSize := int(d.Width) * int(d.Height) * int(d.Bpp) / 8
	d.Data = core.ReadBytes(r, dataSize)

	glog.Debugf("Offscreen bitmap: ID=%d, index=%d, %dx%d, %d bpp, %d bytes",
		d.CacheId, d.CacheIndex, d.Width, d.Height, d.Bpp, len(d.Data))
}

func (d *TsOffscreenBitmapData) Write(w io.Writer) {
	core.WriteLE(w, d.CacheId)
	core.WriteLE(w, d.CacheIndex)
	core.WriteLE(w, d.Width)
	core.WriteLE(w, d.Height)
	core.WriteLE(w, d.Bpp)
	core.WriteFull(w, d.Data)
}

// Offscreen Bitmap Cache PDU
type TsOffscreenBitmapCachePdu struct {
	Header TsShareDataHeader
	Data   TsOffscreenBitmapData
}

func (p *TsOffscreenBitmapCachePdu) Read(r io.Reader) {
	p.Header.Read(r)
	p.Data.Read(r)
}

func (p *TsOffscreenBitmapCachePdu) Write(w io.Writer) {
	// p.Header.Write(w) // TsShareDataHeader has no Write method
	p.Data.Write(w)
}

// Offscreen Bitmap Cache Manager
type OffscreenBitmapManager struct {
	cache *OffscreenBitmapCache
}

// New Offscreen Bitmap Manager
func NewOffscreenBitmapManager(cacheSize, maxEntries uint16) *OffscreenBitmapManager {
	return &OffscreenBitmapManager{
		cache: NewOffscreenBitmapCache(cacheSize, maxEntries),
	}
}

// Process Offscreen Bitmap Data
func (m *OffscreenBitmapManager) ProcessOffscreenBitmap(data *TsOffscreenBitmapData) uint16 {
	return m.cache.AddEntry(data.Data, data.Width, data.Height, data.Bpp)
}

// Get Offscreen Bitmap
func (m *OffscreenBitmapManager) GetOffscreenBitmap(id uint16) *OffscreenCacheEntry {
	return m.cache.GetEntry(id)
}

// Remove Offscreen Bitmap
func (m *OffscreenBitmapManager) RemoveOffscreenBitmap(id uint16) bool {
	return m.cache.RemoveEntry(id)
}

// Clear All Offscreen Bitmaps
func (m *OffscreenBitmapManager) Clear() {
	m.cache.Clear()
}

// Get Statistics
func (m *OffscreenBitmapManager) GetStats() (count, maxEntries uint16) {
	return m.cache.GetStats()
}

// Offscreen Bitmap Order
// https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-rdpegdi/2c3c3c41-1d54-4254-bb62-bc082a3c1f10
type TsOffscreenBitmapOrder struct {
	Header     TsOrderHeader
	CacheId    uint16
	CacheIndex uint16
	DestLeft   uint16
	DestTop    uint16
	DestRight  uint16
	DestBottom uint16
	SourceLeft uint16
	SourceTop  uint16
}

// Add TsOrderHeader implementation
// https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-rdpegdi/2c3c3c41-1d54-4254-bb62-bc082a3c1f10
type TsOrderHeader struct {
	ControlFlags uint8
	OrderType    uint8
}

func (h *TsOrderHeader) Read(r io.Reader) {
	var headerByte uint8
	core.ReadLE(r, &headerByte)
	h.ControlFlags = headerByte >> 4
	h.OrderType = headerByte & 0x0F
}

func (h *TsOrderHeader) Write(w io.Writer) {
	headerByte := (h.ControlFlags << 4) | (h.OrderType & 0x0F)
	core.WriteLE(w, headerByte)
}

// Update TsOffscreenBitmapOrder to use header parsing
func (o *TsOffscreenBitmapOrder) Read(r io.Reader) {
	o.Header.Read(r)
	core.ReadLE(r, &o.CacheId)
	core.ReadLE(r, &o.CacheIndex)
	core.ReadLE(r, &o.DestLeft)
	core.ReadLE(r, &o.DestTop)
	core.ReadLE(r, &o.DestRight)
	core.ReadLE(r, &o.DestBottom)
	core.ReadLE(r, &o.SourceLeft)
	core.ReadLE(r, &o.SourceTop)

	glog.Debugf("Offscreen bitmap order: header=[flags=%d type=%d], cache=%d, index=%d, dest=[%d,%d,%d,%d], source=[%d,%d]",
		o.Header.ControlFlags, o.Header.OrderType, o.CacheId, o.CacheIndex, o.DestLeft, o.DestTop, o.DestRight, o.DestBottom, o.SourceLeft, o.SourceTop)
}

func (o *TsOffscreenBitmapOrder) Write(w io.Writer) {
	o.Header.Write(w)
	core.WriteLE(w, o.CacheId)
	core.WriteLE(w, o.CacheIndex)
	core.WriteLE(w, o.DestLeft)
	core.WriteLE(w, o.DestTop)
	core.WriteLE(w, o.DestRight)
	core.WriteLE(w, o.DestBottom)
	core.WriteLE(w, o.SourceLeft)
	core.WriteLE(w, o.SourceTop)
}

// Offscreen Bitmap Delete Order
type TsOffscreenBitmapDeleteOrder struct {
	Header     TsOrderHeader
	CacheId    uint16
	CacheIndex uint16
}

// Update TsOffscreenBitmapDeleteOrder to use header parsing
func (o *TsOffscreenBitmapDeleteOrder) Read(r io.Reader) {
	o.Header.Read(r)
	core.ReadLE(r, &o.CacheId)
	core.ReadLE(r, &o.CacheIndex)

	glog.Debugf("Offscreen bitmap delete order: header=[flags=%d type=%d], cache=%d, index=%d", o.Header.ControlFlags, o.Header.OrderType, o.CacheId, o.CacheIndex)
}

func (o *TsOffscreenBitmapDeleteOrder) Write(w io.Writer) {
	o.Header.Write(w)
	core.WriteLE(w, o.CacheId)
	core.WriteLE(w, o.CacheIndex)
}

// Offscreen Bitmap Cache Support
type OffscreenBitmapCacheSupport struct {
	SupportLevel uint32
	CacheSize    uint16
	CacheEntries uint16
}

func (s *OffscreenBitmapCacheSupport) Read(r io.Reader) {
	core.ReadLE(r, &s.SupportLevel)
	core.ReadLE(r, &s.CacheSize)
	core.ReadLE(r, &s.CacheEntries)
}

func (s *OffscreenBitmapCacheSupport) Write(w io.Writer) {
	core.WriteLE(w, s.SupportLevel)
	core.WriteLE(w, s.CacheSize)
	core.WriteLE(w, s.CacheEntries)
}

// Create Default Offscreen Bitmap Cache Support
func NewOffscreenBitmapCacheSupport() *OffscreenBitmapCacheSupport {
	return &OffscreenBitmapCacheSupport{
		SupportLevel: OFFSCREEN_SUPPORT_LEVEL_DEFAULT,
		CacheSize:    7680, // 7.5MB default
		CacheEntries: 100,  // 100 entries default
	}
}
