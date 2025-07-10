package virtualchannels

import (
	"bytes"
	"fmt"
	"sync"
	"time"

	"github.com/kdsmith18542/gordp/core"
	"github.com/kdsmith18542/gordp/proto/audio"
)

// AudioHandler handles audio redirection
type AudioHandler struct {
	manager *VirtualChannelManager

	// Audio state
	mu        sync.RWMutex
	isEnabled bool
	isPlaying bool
	volume    float64 // 0.0 to 1.0

	// Enhanced audio management
	audioManager   *audio.AudioManager
	selectedFormat *audio.AudioFormat
	audioBuffer    []byte
	bufferSize     int
	lastTimestamp  uint32
	audioStats     map[string]interface{}
	formatCache    map[uint16]*audio.AudioFormat
}

// NewAudioHandler creates a new audio handler
func NewAudioHandler(manager *VirtualChannelManager) *AudioHandler {
	handler := &AudioHandler{
		manager:      manager,
		isEnabled:    true,
		volume:       1.0,
		audioManager: audio.NewAudioManager(nil),
		bufferSize:   4096,
		audioBuffer:  make([]byte, 0),
		audioStats:   make(map[string]interface{}),
		formatCache:  make(map[uint16]*audio.AudioFormat),
	}

	// Initialize default audio format
	handler.selectedFormat = &audio.AudioFormat{
		FormatTag:      0x0001, // WAVE_FORMAT_PCM
		Channels:       2,      // Stereo
		SamplesPerSec:  48000,  // 48 kHz
		AvgBytesPerSec: 192000, // 2 channels * 2 bytes * 48000
		BlockAlign:     4,      // 2 channels * 2 bytes
		BitsPerSample:  16,     // 16-bit
		ExtraSize:      0,
	}

	return handler
}

// OnAudioOpen is called when the audio channel opens
func (h *AudioHandler) OnAudioOpen() {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.manager.SetChannelOpen("audio", true)
	fmt.Println("Audio channel opened")

	// Send audio capabilities and formats
	h.sendAudioCapabilities()
	h.sendAudioFormats()
}

// OnAudioClose is called when the audio channel closes
func (h *AudioHandler) OnAudioClose() {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.manager.SetChannelOpen("audio", false)
	h.isPlaying = false
	h.audioBuffer = h.audioBuffer[:0] // Clear buffer
	fmt.Println("Audio channel closed")
}

// OnAudioData is called when audio data is received
func (h *AudioHandler) OnAudioData(data []byte, format string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if !h.isEnabled {
		return
	}

	fmt.Printf("Audio data received: %d bytes, format: %s\n", len(data), format)

	// Process and buffer the audio data
	h.processAudioData(data, format)

	// Send audio data to remote server via RDP audio channel
	client := h.manager.client
	if client != nil && h.manager.IsChannelOpen("audio") {
		h.sendAudioData(data, format)
	}

	h.isPlaying = true
	h.updateAudioStats("data_received", len(data))
}

// processAudioData processes incoming audio data
func (h *AudioHandler) processAudioData(data []byte, format string) {
	// Apply volume adjustment
	if h.volume != 1.0 {
		data = h.applyVolume(data)
	}

	// Buffer the audio data
	h.audioBuffer = append(h.audioBuffer, data...)

	// Limit buffer size
	if len(h.audioBuffer) > h.bufferSize {
		h.audioBuffer = h.audioBuffer[len(h.audioBuffer)-h.bufferSize:]
	}

	h.updateAudioStats("buffer_size", len(h.audioBuffer))
}

// applyVolume applies volume adjustment to audio data
func (h *AudioHandler) applyVolume(data []byte) []byte {
	if h.volume == 1.0 {
		return data
	}

	// For 16-bit PCM audio, adjust volume by scaling samples
	if h.selectedFormat != nil && h.selectedFormat.BitsPerSample == 16 {
		adjusted := make([]byte, len(data))
		copy(adjusted, data)

		for i := 0; i < len(adjusted); i += 2 {
			if i+1 < len(adjusted) {
				sample := int16(adjusted[i]) | (int16(adjusted[i+1]) << 8)
				sample = int16(float64(sample) * h.volume)
				adjusted[i] = byte(sample & 0xFF)
				adjusted[i+1] = byte((sample >> 8) & 0xFF)
			}
		}
		return adjusted
	}

	return data
}

// sendAudioData sends audio data to the remote system
func (h *AudioHandler) sendAudioData(data []byte, format string) {
	client := h.manager.client
	if client == nil {
		return
	}

	// Determine format ID based on format string
	formatID := h.getFormatID(format)

	// Create audio wave message
	waveMsg := h.audioManager.CreateClientWaveMessage(formatID, data, h.lastTimestamp)

	// Send via virtual channel
	err := client.SendVirtualChannelData("RDPSND", waveMsg.Serialize(), 0)
	if err != nil {
		fmt.Printf("Failed to send audio data: %v\n", err)
		h.updateAudioStats("send_errors", 1)
	} else {
		fmt.Printf("Audio data sent to remote: %d bytes\n", len(data))
		h.updateAudioStats("data_sent", len(data))
	}

	h.lastTimestamp += uint32(len(data) / 4) // Approximate timestamp increment
}

// getFormatID converts format string to format ID
func (h *AudioHandler) getFormatID(format string) uint16 {
	switch format {
	case "PCM":
		return 1
	case "ADPCM":
		return 2
	case "DVI":
		return 3
	case "GSM":
		return 4
	default:
		return 1 // Default to PCM
	}
}

// sendAudioCapabilities sends audio capabilities to the remote system
func (h *AudioHandler) sendAudioCapabilities() {
	client := h.manager.client
	if client == nil {
		return
	}

	capabilitiesMsg := h.audioManager.CreateClientVersionAndFormatsMessage()
	err := client.SendVirtualChannelData("RDPSND", capabilitiesMsg.Serialize(), 0)
	if err != nil {
		fmt.Printf("Failed to send audio capabilities: %v\n", err)
	}
}

// sendAudioFormats sends audio formats to the remote system
func (h *AudioHandler) sendAudioFormats() {
	client := h.manager.client
	if client == nil {
		return
	}

	formatsMsg := h.audioManager.CreateClientFormatsMessage()
	err := client.SendVirtualChannelData("RDPSND", formatsMsg.Serialize(), 0)
	if err != nil {
		fmt.Printf("Failed to send audio formats: %v\n", err)
	}
}

// sendFormatConfirm sends format confirmation
func (h *AudioHandler) sendFormatConfirm(formatID uint16) {
	client := h.manager.client
	if client == nil {
		return
	}

	confirmMsg := h.audioManager.CreateClientFormatConfirmMessage(formatID)
	err := client.SendVirtualChannelData("RDPSND", confirmMsg.Serialize(), 0)
	if err != nil {
		fmt.Printf("Failed to send format confirm: %v\n", err)
	}
}

// IsEnabled returns whether audio redirection is enabled
func (h *AudioHandler) IsEnabled() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.isEnabled
}

// SetEnabled enables or disables audio redirection
func (h *AudioHandler) SetEnabled(enabled bool) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.isEnabled = enabled
	if enabled {
		fmt.Println("Audio redirection enabled")
	} else {
		fmt.Println("Audio redirection disabled")
		h.isPlaying = false
	}

	h.updateAudioStats("enabled", enabled)
}

// IsPlaying returns whether audio is currently playing
func (h *AudioHandler) IsPlaying() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.isPlaying
}

// GetVolume returns the current volume level
func (h *AudioHandler) GetVolume() float64 {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.volume
}

// SetVolume sets the volume level (0.0 to 1.0)
func (h *AudioHandler) SetVolume(volume float64) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if volume < 0.0 {
		volume = 0.0
	} else if volume > 1.0 {
		volume = 1.0
	}

	h.volume = volume
	fmt.Printf("Audio volume set to: %.2f\n", volume)
	h.updateAudioStats("volume", volume)
}

// PlayAudio plays audio data
func (h *AudioHandler) PlayAudio(data []byte, format string) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if !h.isEnabled {
		return fmt.Errorf("audio redirection is disabled")
	}

	client := h.manager.client
	if client == nil {
		return fmt.Errorf("RDP client is not initialized")
	}

	if !h.manager.IsChannelOpen("audio") {
		return fmt.Errorf("audio channel is not open")
	}

	// Process the audio data
	h.processAudioData(data, format)

	// Send audio data
	h.sendAudioData(data, format)

	fmt.Printf("Playing audio: %d bytes, format: %s\n", len(data), format)
	h.isPlaying = true
	h.updateAudioStats("play_requests", 1)

	return nil
}

// StopAudio stops audio playback
func (h *AudioHandler) StopAudio() {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.isPlaying = false
	h.audioBuffer = h.audioBuffer[:0] // Clear buffer
	fmt.Println("Audio playback stopped")
	h.updateAudioStats("stop_requests", 1)
}

// PauseAudio pauses audio playback
func (h *AudioHandler) PauseAudio() {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.isPlaying {
		h.isPlaying = false
		fmt.Println("Audio playback paused")
		h.updateAudioStats("pause_requests", 1)
	}
}

// ResumeAudio resumes audio playback
func (h *AudioHandler) ResumeAudio() {
	h.mu.Lock()
	defer h.mu.Unlock()

	if !h.isPlaying && h.isEnabled {
		h.isPlaying = true
		fmt.Println("Audio playback resumed")
		h.updateAudioStats("resume_requests", 1)
	}
}

// SetAudioFormat sets the audio format
func (h *AudioHandler) SetAudioFormat(format *audio.AudioFormat) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.selectedFormat = format
	h.formatCache[1] = format // Cache with format ID 1

	fmt.Printf("Audio format set: %d channels, %d Hz, %d bits\n",
		format.Channels, format.SamplesPerSec, format.BitsPerSample)

	h.updateAudioStats("format_changed", 1)
}

// GetAudioFormat returns the current audio format
func (h *AudioHandler) GetAudioFormat() *audio.AudioFormat {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.selectedFormat
}

// GetAudioBuffer returns the current audio buffer
func (h *AudioHandler) GetAudioBuffer() []byte {
	h.mu.RLock()
	defer h.mu.RUnlock()

	buffer := make([]byte, len(h.audioBuffer))
	copy(buffer, h.audioBuffer)
	return buffer
}

// ClearAudioBuffer clears the audio buffer
func (h *AudioHandler) ClearAudioBuffer() {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.audioBuffer = h.audioBuffer[:0]
	fmt.Println("Audio buffer cleared")
	h.updateAudioStats("buffer_cleared", 1)
}

// updateAudioStats updates audio statistics
func (h *AudioHandler) updateAudioStats(key string, value interface{}) {
	h.audioStats[key] = value
	h.audioStats["last_update"] = time.Now()
}

// GetAudioStats returns audio statistics
func (h *AudioHandler) GetAudioStats() map[string]interface{} {
	h.mu.RLock()
	defer h.mu.RUnlock()

	stats := map[string]interface{}{
		"enabled":      h.isEnabled,
		"playing":      h.isPlaying,
		"volume":       h.volume,
		"channel_open": h.manager.IsChannelOpen("audio"),
		"buffer_size":  len(h.audioBuffer),
		"format_cache": len(h.formatCache),
	}

	// Merge with detailed stats
	for key, value := range h.audioStats {
		stats[key] = value
	}

	return stats
}

// GetAudioFormats returns available audio formats
func (h *AudioHandler) GetAudioFormats() []*audio.AudioFormat {
	h.mu.RLock()
	defer h.mu.RUnlock()

	formats := make([]*audio.AudioFormat, 0, len(h.formatCache))
	for _, format := range h.formatCache {
		formats = append(formats, format)
	}

	return formats
}

// HandleAudioMessage handles incoming audio messages
func (h *AudioHandler) HandleAudioMessage(msg *audio.AudioMessage) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	switch msg.MessageType {
	case audio.RDPSND_MSG_TYPE_SERVER_AUDIO_VERSION_AND_FORMATS:
		return h.handleServerVersionAndFormats(msg)
	case audio.RDPSND_MSG_TYPE_SERVER_FORMATS:
		return h.handleServerFormats(msg)
	case audio.RDPSND_MSG_TYPE_SERVER_FORMAT_CONFIRM:
		return h.handleServerFormatConfirm(msg)
	case audio.RDPSND_MSG_TYPE_SERVER_WAVE:
		return h.handleServerWave(msg)
	default:
		fmt.Printf("Unhandled audio message type: %d\n", msg.MessageType)
		return nil
	}
}

// handleServerVersionAndFormats handles server version and formats message
func (h *AudioHandler) handleServerVersionAndFormats(msg *audio.AudioMessage) error {
	if len(msg.Data) < 4 {
		return fmt.Errorf("invalid version and formats message size")
	}

	reader := bytes.NewReader(msg.Data)

	var version uint16
	var numFormats uint16

	core.ReadLE(reader, &version)
	core.ReadLE(reader, &numFormats)

	fmt.Printf("Server audio version: %d, formats: %d\n", version, numFormats)

	// Read formats
	formats := make([]audio.AudioFormat, 0, numFormats)
	for i := uint16(0); i < numFormats; i++ {
		format := audio.AudioFormat{}
		core.ReadLE(reader, &format.FormatTag)
		core.ReadLE(reader, &format.Channels)
		core.ReadLE(reader, &format.SamplesPerSec)
		core.ReadLE(reader, &format.AvgBytesPerSec)
		core.ReadLE(reader, &format.BlockAlign)
		core.ReadLE(reader, &format.BitsPerSample)
		core.ReadLE(reader, &format.ExtraSize)

		if format.ExtraSize > 0 {
			format.ExtraData = make([]byte, format.ExtraSize)
			reader.Read(format.ExtraData)
		}

		formats = append(formats, format)
		h.formatCache[uint16(i+1)] = &format
	}

	h.updateAudioStats("server_formats", len(formats))
	return nil
}

// handleServerFormats handles server formats message
func (h *AudioHandler) handleServerFormats(msg *audio.AudioMessage) error {
	if len(msg.Data) < 2 {
		return fmt.Errorf("invalid formats message size")
	}

	reader := bytes.NewReader(msg.Data)

	var numFormats uint16
	core.ReadLE(reader, &numFormats)

	fmt.Printf("Server audio formats: %d\n", numFormats)

	// Read formats
	formats := make([]audio.AudioFormat, 0, numFormats)
	for i := uint16(0); i < numFormats; i++ {
		format := audio.AudioFormat{}
		core.ReadLE(reader, &format.FormatTag)
		core.ReadLE(reader, &format.Channels)
		core.ReadLE(reader, &format.SamplesPerSec)
		core.ReadLE(reader, &format.AvgBytesPerSec)
		core.ReadLE(reader, &format.BlockAlign)
		core.ReadLE(reader, &format.BitsPerSample)
		core.ReadLE(reader, &format.ExtraSize)

		if format.ExtraSize > 0 {
			format.ExtraData = make([]byte, format.ExtraSize)
			reader.Read(format.ExtraData)
		}

		formats = append(formats, format)
		h.formatCache[uint16(i+1)] = &format
	}

	h.updateAudioStats("server_formats", len(formats))
	return nil
}

// handleServerFormatConfirm handles server format confirm message
func (h *AudioHandler) handleServerFormatConfirm(msg *audio.AudioMessage) error {
	if len(msg.Data) < 2 {
		return fmt.Errorf("invalid format confirm message size")
	}

	reader := bytes.NewReader(msg.Data)

	var formatID uint16
	core.ReadLE(reader, &formatID)

	fmt.Printf("Server audio format confirm: %d\n", formatID)
	h.updateAudioStats("format_confirmed", formatID)
	return nil
}

// handleServerWave handles server wave message
func (h *AudioHandler) handleServerWave(msg *audio.AudioMessage) error {
	if len(msg.Data) < 6 {
		return fmt.Errorf("invalid wave message size")
	}

	reader := bytes.NewReader(msg.Data)

	var timestamp uint32
	var formatID uint16

	core.ReadLE(reader, &timestamp)
	core.ReadLE(reader, &formatID)

	data := msg.Data[6:]

	fmt.Printf("Server audio wave: formatID=%d, size=%d bytes, timestamp=%d\n", formatID, len(data), timestamp)

	// Process the received audio data
	h.processAudioData(data, "PCM") // Assume PCM for now
	h.updateAudioStats("data_received", len(data))

	return nil
}
