package virtualchannels

import (
	"bytes"
	"fmt"
	"sync"
	"time"

	"github.com/kdsmith18542/gordp/core"
	"github.com/kdsmith18542/gordp/proto/clipboard"
)

// ClipboardHandler handles clipboard synchronization
type ClipboardHandler struct {
	manager *VirtualChannelManager

	// Clipboard state
	mu              sync.RWMutex
	localClipboard  string
	remoteClipboard string
	isEnabled       bool

	// Enhanced clipboard management
	clipboardManager *clipboard.ClipboardManager
	lastSyncTime     time.Time
	syncInterval     time.Duration
	formatCache      map[clipboard.ClipboardFormat][]byte
	requestQueue     []clipboard.ClipboardFormat
}

// NewClipboardHandler creates a new clipboard handler
func NewClipboardHandler(manager *VirtualChannelManager) *ClipboardHandler {
	handler := &ClipboardHandler{
		manager:          manager,
		isEnabled:        true,
		clipboardManager: clipboard.NewClipboardManager(nil),
		syncInterval:     5 * time.Second,
		formatCache:      make(map[clipboard.ClipboardFormat][]byte),
		requestQueue:     make([]clipboard.ClipboardFormat, 0),
	}

	// Set up periodic synchronization
	go handler.periodicSync()

	return handler
}

// OnClipboardOpen is called when the clipboard channel opens
func (h *ClipboardHandler) OnClipboardOpen() {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.manager.SetChannelOpen("clipboard", true)
	fmt.Println("Clipboard channel opened")

	// Send capabilities and monitor ready
	h.sendCapabilities()
	h.sendMonitorReady()
}

// OnClipboardClose is called when the clipboard channel closes
func (h *ClipboardHandler) OnClipboardClose() {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.manager.SetChannelOpen("clipboard", false)
	fmt.Println("Clipboard channel closed")
}

// OnFormatList handles clipboard format list events
func (h *ClipboardHandler) OnFormatList(formats []clipboard.ClipboardFormat) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	fmt.Printf("Received clipboard format list: %d formats\n", len(formats))

	// Cache available formats
	for _, format := range formats {
		h.formatCache[format] = nil // Mark as available but not yet requested
	}

	// Request data for text formats we're interested in
	for _, format := range formats {
		if format == clipboard.CLIPRDR_FORMAT_UNICODETEXT ||
			format == clipboard.CLIPRDR_FORMAT_OEMTEXT ||
			format == clipboard.CLIPRDR_FORMAT_HTML {
			h.requestFormatData(format)
		}
	}

	return nil
}

// OnFormatDataRequest handles clipboard format data request events
func (h *ClipboardHandler) OnFormatDataRequest(formatID clipboard.ClipboardFormat) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	fmt.Printf("Received clipboard format data request: %d (%s)\n", formatID, clipboard.GetFormatName(formatID))

	// Check if we have data for this format
	if data, exists := h.formatCache[formatID]; exists && data != nil {
		h.sendFormatDataResponse(formatID, data)
	} else {
		// Send empty response if we don't have data
		h.sendFormatDataResponse(formatID, []byte{})
	}

	return nil
}

// OnFormatDataResponse handles clipboard format data response events
func (h *ClipboardHandler) OnFormatDataResponse(formatID clipboard.ClipboardFormat, data []byte) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Cache the received data
	h.formatCache[formatID] = data

	// Update remote clipboard content for text formats
	if formatID == clipboard.CLIPRDR_FORMAT_UNICODETEXT || formatID == clipboard.CLIPRDR_FORMAT_OEMTEXT {
		h.remoteClipboard = string(data)
		fmt.Printf("Received clipboard text data from remote: %d bytes\n", len(data))

		// Update local clipboard if enabled and different
		if h.isEnabled && h.localClipboard != h.remoteClipboard {
			h.updateLocalClipboard(h.remoteClipboard)
		}
	} else if formatID == clipboard.CLIPRDR_FORMAT_HTML {
		fmt.Printf("Received clipboard HTML data from remote: %d bytes\n", len(data))
	} else {
		fmt.Printf("Received clipboard data from remote: format=%d (%s), %d bytes\n",
			formatID, clipboard.GetFormatName(formatID), len(data))
	}

	return nil
}

// OnFileContentsRequest handles file contents request events
func (h *ClipboardHandler) OnFileContentsRequest(streamID uint32, listIndex uint32, dwFlags uint32, nPositionLow uint32, nPositionHigh uint32, cbRequested uint32, clipDataID uint32) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	fmt.Printf("Received file contents request: streamID=%d, listIndex=%d, flags=%d, position=%d, requested=%d, clipDataID=%d\n",
		streamID, listIndex, dwFlags, (uint64(nPositionHigh)<<32)|uint64(nPositionLow), cbRequested, clipDataID)

	// Send empty file contents response for now
	h.sendFileContentsResponse(streamID, []byte{})

	return nil
}

// SetLocalClipboard sets the local clipboard content
func (h *ClipboardHandler) SetLocalClipboard(content string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.localClipboard = content
	fmt.Printf("Local clipboard updated: %d bytes\n", len(content))

	// Cache the data
	h.formatCache[clipboard.CLIPRDR_FORMAT_UNICODETEXT] = []byte(content)

	// Send to remote if channel is open
	if h.manager.IsChannelOpen("clipboard") {
		h.sendToRemote(content)
	}
}

// GetLocalClipboard returns the local clipboard content
func (h *ClipboardHandler) GetLocalClipboard() string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.localClipboard
}

// GetRemoteClipboard returns the remote clipboard content
func (h *ClipboardHandler) GetRemoteClipboard() string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.remoteClipboard
}

// IsEnabled returns whether clipboard synchronization is enabled
func (h *ClipboardHandler) IsEnabled() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.isEnabled
}

// SetEnabled enables or disables clipboard synchronization
func (h *ClipboardHandler) SetEnabled(enabled bool) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.isEnabled = enabled
	if enabled {
		fmt.Println("Clipboard synchronization enabled")
	} else {
		fmt.Println("Clipboard synchronization disabled")
	}
}

// sendToRemote sends clipboard data to the remote system
func (h *ClipboardHandler) sendToRemote(content string) {
	h.mu.RLock()
	client := h.manager.client
	h.mu.RUnlock()

	if client == nil {
		fmt.Println("RDP client is not initialized; cannot send clipboard data")
		return
	}

	// Announce available formats
	formats := []clipboard.ClipboardFormat{
		clipboard.CLIPRDR_FORMAT_UNICODETEXT,
		clipboard.CLIPRDR_FORMAT_OEMTEXT,
		clipboard.CLIPRDR_FORMAT_HTML,
	}

	formatListMsg := h.clipboardManager.CreateFormatListMessage(formats)
	err := client.SendVirtualChannelData("CLIPRDR", formatListMsg.Serialize(), 0)
	if err != nil {
		fmt.Printf("Failed to send format list: %v\n", err)
		return
	}

	// Send clipboard data as FORMAT_DATA_RESPONSE
	data := []byte(content)
	dataMsg := h.clipboardManager.CreateFormatDataResponseMessage(clipboard.CLIPRDR_FORMAT_UNICODETEXT, data)
	err = client.SendVirtualChannelData("CLIPRDR", dataMsg.Serialize(), 0)
	if err != nil {
		fmt.Printf("Failed to send clipboard data: %v\n", err)
		return
	}

	fmt.Printf("Sent clipboard data to remote: %d bytes\n", len(content))
}

// sendCapabilities sends clipboard capabilities to the remote system
func (h *ClipboardHandler) sendCapabilities() {
	client := h.manager.client
	if client == nil {
		return
	}

	capabilitiesMsg := h.clipboardManager.CreateCapabilitiesMessage()
	err := client.SendVirtualChannelData("CLIPRDR", capabilitiesMsg.Serialize(), 0)
	if err != nil {
		fmt.Printf("Failed to send capabilities: %v\n", err)
	}
}

// sendMonitorReady sends monitor ready message
func (h *ClipboardHandler) sendMonitorReady() {
	client := h.manager.client
	if client == nil {
		return
	}

	monitorReadyMsg := h.clipboardManager.CreateMonitorReadyMessage()
	err := client.SendVirtualChannelData("CLIPRDR", monitorReadyMsg.Serialize(), 0)
	if err != nil {
		fmt.Printf("Failed to send monitor ready: %v\n", err)
	}
}

// requestFormatData requests data for a specific format
func (h *ClipboardHandler) requestFormatData(formatID clipboard.ClipboardFormat) {
	client := h.manager.client
	if client == nil {
		return
	}

	// Create format data request message
	buf := new(bytes.Buffer)
	core.WriteLE(buf, formatID)

	requestMsg := &clipboard.ClipboardMessage{
		MessageType:  clipboard.CLIPRDR_MSG_TYPE_FORMAT_DATA_REQUEST,
		MessageFlags: 0,
		DataLength:   uint32(buf.Len()),
		Data:         buf.Bytes(),
	}

	err := client.SendVirtualChannelData("CLIPRDR", requestMsg.Serialize(), 0)
	if err != nil {
		fmt.Printf("Failed to request format data: %v\n", err)
	}
}

// sendFormatDataResponse sends format data response
func (h *ClipboardHandler) sendFormatDataResponse(formatID clipboard.ClipboardFormat, data []byte) {
	client := h.manager.client
	if client == nil {
		return
	}

	dataMsg := h.clipboardManager.CreateFormatDataResponseMessage(formatID, data)
	err := client.SendVirtualChannelData("CLIPRDR", dataMsg.Serialize(), 0)
	if err != nil {
		fmt.Printf("Failed to send format data response: %v\n", err)
	}
}

// sendFileContentsResponse sends file contents response
func (h *ClipboardHandler) sendFileContentsResponse(streamID uint32, data []byte) {
	client := h.manager.client
	if client == nil {
		return
	}

	// Create file contents response message
	buf := new(bytes.Buffer)
	core.WriteLE(buf, streamID)
	core.WriteLE(buf, uint32(0))         // listIndex
	core.WriteLE(buf, uint32(0))         // dwFlags
	core.WriteLE(buf, uint32(0))         // nPositionLow
	core.WriteLE(buf, uint32(0))         // nPositionHigh
	core.WriteLE(buf, uint32(len(data))) // cbRequested
	core.WriteLE(buf, uint32(0))         // clipDataID
	buf.Write(data)

	responseMsg := &clipboard.ClipboardMessage{
		MessageType:  clipboard.CLIPRDR_MSG_TYPE_FILECONTENTS_RESPONSE,
		MessageFlags: 0,
		DataLength:   uint32(buf.Len()),
		Data:         buf.Bytes(),
	}

	err := client.SendVirtualChannelData("CLIPRDR", responseMsg.Serialize(), 0)
	if err != nil {
		fmt.Printf("Failed to send file contents response: %v\n", err)
	}
}

// updateLocalClipboard updates the local clipboard (platform-specific)
func (h *ClipboardHandler) updateLocalClipboard(content string) {
	// This would require platform-specific clipboard access
	// For now, just update our internal state
	h.localClipboard = content
	fmt.Printf("Local clipboard updated from remote: %d bytes\n", len(content))
}

// periodicSync performs periodic clipboard synchronization
func (h *ClipboardHandler) periodicSync() {
	ticker := time.NewTicker(h.syncInterval)
	defer ticker.Stop()

	for range ticker.C {
		if h.isEnabled && h.manager.IsChannelOpen("clipboard") {
			h.SyncClipboard()
		}
	}
}

// SyncClipboard synchronizes clipboard between local and remote
func (h *ClipboardHandler) SyncClipboard() {
	h.mu.Lock()
	defer h.mu.Unlock()

	if !h.isEnabled {
		return
	}

	client := h.manager.client
	if client == nil {
		fmt.Println("RDP client is not initialized; cannot sync clipboard")
		return
	}

	// Send local clipboard to remote if channel is open and content has changed
	if h.manager.IsChannelOpen("clipboard") && h.localClipboard != "" {
		h.sendToRemote(h.localClipboard)
	}

	// Request remote clipboard formats to check for changes
	if h.manager.IsChannelOpen("clipboard") {
		// Send monitor ready to trigger format list
		h.sendMonitorReady()
	}

	h.lastSyncTime = time.Now()
	fmt.Println("Clipboard synchronization complete")
}

// GetClipboardStats returns clipboard statistics
func (h *ClipboardHandler) GetClipboardStats() map[string]interface{} {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return map[string]interface{}{
		"enabled":           h.isEnabled,
		"local_size":        len(h.localClipboard),
		"remote_size":       len(h.remoteClipboard),
		"last_sync":         h.lastSyncTime,
		"format_cache_size": len(h.formatCache),
		"channel_open":      h.manager.IsChannelOpen("clipboard"),
	}
}

// ClearClipboardCache clears the clipboard format cache
func (h *ClipboardHandler) ClearClipboardCache() {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.formatCache = make(map[clipboard.ClipboardFormat][]byte)
	fmt.Println("Clipboard cache cleared")
}
