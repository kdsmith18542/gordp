package audio

import (
	"bytes"
	"fmt"
	"io"

	"github.com/GoFeGroup/gordp/core"
	"github.com/GoFeGroup/gordp/glog"
)

// AudioFormat represents audio format information
type AudioFormat struct {
	FormatTag      uint16 // WAVE_FORMAT_PCM = 0x0001
	Channels       uint16 // Number of channels (1 = mono, 2 = stereo)
	SamplesPerSec  uint32 // Sample rate (e.g., 44100, 48000)
	AvgBytesPerSec uint32 // Average bytes per second
	BlockAlign     uint16 // Block alignment
	BitsPerSample  uint16 // Bits per sample (8, 16, 24, 32)
	ExtraSize      uint16 // Size of extra data
	ExtraData      []byte // Extra format data
}

// AudioMessageType represents the type of audio message
type AudioMessageType uint16

const (
	RDPSND_MSG_TYPE_SERVER_AUDIO_VERSION_AND_FORMATS AudioMessageType = 0x0001
	RDPSND_MSG_TYPE_CLIENT_AUDIO_VERSION_AND_FORMATS AudioMessageType = 0x0002
	RDPSND_MSG_TYPE_SERVER_FORMATS                   AudioMessageType = 0x0003
	RDPSND_MSG_TYPE_CLIENT_FORMATS                   AudioMessageType = 0x0004
	RDPSND_MSG_TYPE_SERVER_FORMAT_CONFIRM            AudioMessageType = 0x0005
	RDPSND_MSG_TYPE_CLIENT_FORMAT_CONFIRM            AudioMessageType = 0x0006
	RDPSND_MSG_TYPE_SERVER_TRAINING                  AudioMessageType = 0x0007
	RDPSND_MSG_TYPE_CLIENT_TRAINING_CONFIRM          AudioMessageType = 0x0008
	RDPSND_MSG_TYPE_SERVER_TRAINING_CONFIRM          AudioMessageType = 0x0009
	RDPSND_MSG_TYPE_CLIENT_TRAINING                  AudioMessageType = 0x000A
	RDPSND_MSG_TYPE_SERVER_WAVE_INFO                 AudioMessageType = 0x000B
	RDPSND_MSG_TYPE_CLIENT_WAVE_INFO                 AudioMessageType = 0x000C
	RDPSND_MSG_TYPE_SERVER_WAVE                      AudioMessageType = 0x000D
	RDPSND_MSG_TYPE_CLIENT_WAVE                      AudioMessageType = 0x000E
	RDPSND_MSG_TYPE_SERVER_CLOSE                     AudioMessageType = 0x000F
	RDPSND_MSG_TYPE_CLIENT_CLOSE                     AudioMessageType = 0x0010
	RDPSND_MSG_TYPE_SERVER_FORMATS_NEW               AudioMessageType = 0x0011
	RDPSND_MSG_TYPE_CLIENT_FORMATS_NEW               AudioMessageType = 0x0012
	RDPSND_MSG_TYPE_SERVER_DVR_SUBSCRIBE             AudioMessageType = 0x0013
	RDPSND_MSG_TYPE_CLIENT_DVR_SUBSCRIBE             AudioMessageType = 0x0014
	RDPSND_MSG_TYPE_SERVER_DVR_START                 AudioMessageType = 0x0015
	RDPSND_MSG_TYPE_CLIENT_DVR_START                 AudioMessageType = 0x0016
	RDPSND_MSG_TYPE_SERVER_DVR_STOP                  AudioMessageType = 0x0017
	RDPSND_MSG_TYPE_CLIENT_DVR_STOP                  AudioMessageType = 0x0018
)

// AudioMessage represents an audio message header
type AudioMessage struct {
	MessageType  AudioMessageType
	MessageFlags uint16
	DataLength   uint32
	Data         []byte
}

// AudioVersionAndFormats represents audio version and formats message
type AudioVersionAndFormats struct {
	Version    uint16
	NumFormats uint16
	Formats    []AudioFormat
}

// AudioWaveInfo represents audio wave info message
type AudioWaveInfo struct {
	Timestamp uint32
	FormatID  uint16
	Data      []byte
}

// AudioManager manages audio operations
type AudioManager struct {
	version uint16
	formats []AudioFormat
	handler AudioHandler
}

// AudioHandler handles audio events
type AudioHandler interface {
	OnAudioData(formatID uint16, data []byte, timestamp uint32) error
	OnAudioFormatList(formats []AudioFormat) error
	OnAudioFormatConfirm(formatID uint16) error
}

// DefaultAudioHandler provides a default implementation
type DefaultAudioHandler struct{}

// NewDefaultAudioHandler creates a new default audio handler
func NewDefaultAudioHandler() *DefaultAudioHandler {
	return &DefaultAudioHandler{}
}

// OnAudioData handles audio data events
func (h *DefaultAudioHandler) OnAudioData(formatID uint16, data []byte, timestamp uint32) error {
	glog.Debugf("Received audio data: formatID=%d, size=%d bytes, timestamp=%d", formatID, len(data), timestamp)
	return nil
}

// OnAudioFormatList handles audio format list events
func (h *DefaultAudioHandler) OnAudioFormatList(formats []AudioFormat) error {
	glog.Debugf("Received audio format list: %d formats", len(formats))
	for i, format := range formats {
		glog.Debugf("  Format %d: %d channels, %d Hz, %d bits", i, format.Channels, format.SamplesPerSec, format.BitsPerSample)
	}
	return nil
}

// OnAudioFormatConfirm handles audio format confirm events
func (h *DefaultAudioHandler) OnAudioFormatConfirm(formatID uint16) error {
	glog.Debugf("Audio format confirmed: %d", formatID)
	return nil
}

// NewAudioManager creates a new audio manager
func NewAudioManager(handler AudioHandler) *AudioManager {
	if handler == nil {
		handler = NewDefaultAudioHandler()
	}

	return &AudioManager{
		version: 0x0001, // RDP 8.1
		formats: []AudioFormat{
			{
				FormatTag:      0x0001, // WAVE_FORMAT_PCM
				Channels:       2,      // Stereo
				SamplesPerSec:  48000,  // 48 kHz
				AvgBytesPerSec: 192000, // 2 channels * 2 bytes * 48000
				BlockAlign:     4,      // 2 channels * 2 bytes
				BitsPerSample:  16,     // 16-bit
				ExtraSize:      0,
			},
			{
				FormatTag:      0x0001, // WAVE_FORMAT_PCM
				Channels:       2,      // Stereo
				SamplesPerSec:  44100,  // 44.1 kHz
				AvgBytesPerSec: 176400, // 2 channels * 2 bytes * 44100
				BlockAlign:     4,      // 2 channels * 2 bytes
				BitsPerSample:  16,     // 16-bit
				ExtraSize:      0,
			},
		},
		handler: handler,
	}
}

// ReadAudioMessage reads an audio message from the stream
func ReadAudioMessage(r io.Reader) (*AudioMessage, error) {
	msg := &AudioMessage{}

	// Read message header
	if err := core.ReadLE(r, &msg.MessageType); err != nil {
		return nil, fmt.Errorf("failed to read message type: %v", err)
	}

	if err := core.ReadLE(r, &msg.MessageFlags); err != nil {
		return nil, fmt.Errorf("failed to read message flags: %v", err)
	}

	if err := core.ReadLE(r, &msg.DataLength); err != nil {
		return nil, fmt.Errorf("failed to read data length: %v", err)
	}

	// Read message data
	if msg.DataLength > 0 {
		msg.Data = make([]byte, msg.DataLength)
		if _, err := io.ReadFull(r, msg.Data); err != nil {
			return nil, fmt.Errorf("failed to read message data: %w", err)
		}
	}

	return msg, nil
}

// Serialize serializes the audio message
func (m *AudioMessage) Serialize() []byte {
	buf := new(bytes.Buffer)
	core.WriteLE(buf, m.MessageType)
	core.WriteLE(buf, m.MessageFlags)
	core.WriteLE(buf, m.DataLength)
	if len(m.Data) > 0 {
		buf.Write(m.Data)
	}
	return buf.Bytes()
}

// ProcessMessage processes an audio message
func (am *AudioManager) ProcessMessage(msg *AudioMessage) error {
	switch msg.MessageType {
	case RDPSND_MSG_TYPE_SERVER_AUDIO_VERSION_AND_FORMATS:
		return am.handleServerVersionAndFormats(msg)
	case RDPSND_MSG_TYPE_SERVER_FORMATS:
		return am.handleServerFormats(msg)
	case RDPSND_MSG_TYPE_SERVER_FORMAT_CONFIRM:
		return am.handleServerFormatConfirm(msg)
	case RDPSND_MSG_TYPE_SERVER_WAVE:
		return am.handleServerWave(msg)
	default:
		glog.Debugf("Unhandled audio message type: %d", msg.MessageType)
		return nil
	}
}

// handleServerVersionAndFormats handles server version and formats message
func (am *AudioManager) handleServerVersionAndFormats(msg *AudioMessage) error {
	if len(msg.Data) < 4 {
		return fmt.Errorf("invalid version and formats message size")
	}

	reader := bytes.NewReader(msg.Data)

	var version uint16
	var numFormats uint16

	core.ReadLE(reader, &version)
	core.ReadLE(reader, &numFormats)

	glog.Debugf("Server audio version: %d, formats: %d", version, numFormats)

	// Read formats
	formats := make([]AudioFormat, 0, numFormats)
	for i := uint16(0); i < numFormats; i++ {
		format := AudioFormat{}
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
	}

	am.formats = formats
	return am.handler.OnAudioFormatList(formats)
}

// handleServerFormats handles server formats message
func (am *AudioManager) handleServerFormats(msg *AudioMessage) error {
	if len(msg.Data) < 2 {
		return fmt.Errorf("invalid formats message size")
	}

	reader := bytes.NewReader(msg.Data)

	var numFormats uint16
	core.ReadLE(reader, &numFormats)

	glog.Debugf("Server audio formats: %d", numFormats)

	// Read formats
	formats := make([]AudioFormat, 0, numFormats)
	for i := uint16(0); i < numFormats; i++ {
		format := AudioFormat{}
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
	}

	am.formats = formats
	return am.handler.OnAudioFormatList(formats)
}

// handleServerFormatConfirm handles server format confirm message
func (am *AudioManager) handleServerFormatConfirm(msg *AudioMessage) error {
	if len(msg.Data) < 2 {
		return fmt.Errorf("invalid format confirm message size")
	}

	reader := bytes.NewReader(msg.Data)

	var formatID uint16
	core.ReadLE(reader, &formatID)

	glog.Debugf("Server audio format confirm: %d", formatID)
	return am.handler.OnAudioFormatConfirm(formatID)
}

// handleServerWave handles server wave message
func (am *AudioManager) handleServerWave(msg *AudioMessage) error {
	if len(msg.Data) < 6 {
		return fmt.Errorf("invalid wave message size")
	}

	reader := bytes.NewReader(msg.Data)

	var timestamp uint32
	var formatID uint16

	core.ReadLE(reader, &timestamp)
	core.ReadLE(reader, &formatID)

	data := msg.Data[6:]

	glog.Debugf("Server audio wave: formatID=%d, size=%d bytes, timestamp=%d", formatID, len(data), timestamp)
	return am.handler.OnAudioData(formatID, data, timestamp)
}

// CreateClientVersionAndFormatsMessage creates a client version and formats message
func (am *AudioManager) CreateClientVersionAndFormatsMessage() *AudioMessage {
	buf := new(bytes.Buffer)
	core.WriteLE(buf, am.version)
	core.WriteLE(buf, uint16(len(am.formats)))

	for _, format := range am.formats {
		core.WriteLE(buf, format.FormatTag)
		core.WriteLE(buf, format.Channels)
		core.WriteLE(buf, format.SamplesPerSec)
		core.WriteLE(buf, format.AvgBytesPerSec)
		core.WriteLE(buf, format.BlockAlign)
		core.WriteLE(buf, format.BitsPerSample)
		core.WriteLE(buf, format.ExtraSize)
		if len(format.ExtraData) > 0 {
			buf.Write(format.ExtraData)
		}
	}

	return &AudioMessage{
		MessageType:  RDPSND_MSG_TYPE_CLIENT_AUDIO_VERSION_AND_FORMATS,
		MessageFlags: 0,
		DataLength:   uint32(buf.Len()),
		Data:         buf.Bytes(),
	}
}

// CreateClientFormatsMessage creates a client formats message
func (am *AudioManager) CreateClientFormatsMessage() *AudioMessage {
	buf := new(bytes.Buffer)
	core.WriteLE(buf, uint16(len(am.formats)))

	for _, format := range am.formats {
		core.WriteLE(buf, format.FormatTag)
		core.WriteLE(buf, format.Channels)
		core.WriteLE(buf, format.SamplesPerSec)
		core.WriteLE(buf, format.AvgBytesPerSec)
		core.WriteLE(buf, format.BlockAlign)
		core.WriteLE(buf, format.BitsPerSample)
		core.WriteLE(buf, format.ExtraSize)
		if len(format.ExtraData) > 0 {
			buf.Write(format.ExtraData)
		}
	}

	return &AudioMessage{
		MessageType:  RDPSND_MSG_TYPE_CLIENT_FORMATS,
		MessageFlags: 0,
		DataLength:   uint32(buf.Len()),
		Data:         buf.Bytes(),
	}
}

// CreateClientFormatConfirmMessage creates a client format confirm message
func (am *AudioManager) CreateClientFormatConfirmMessage(formatID uint16) *AudioMessage {
	buf := new(bytes.Buffer)
	core.WriteLE(buf, formatID)

	return &AudioMessage{
		MessageType:  RDPSND_MSG_TYPE_CLIENT_FORMAT_CONFIRM,
		MessageFlags: 0,
		DataLength:   uint32(buf.Len()),
		Data:         buf.Bytes(),
	}
}

// CreateClientWaveMessage creates a client wave message
func (am *AudioManager) CreateClientWaveMessage(formatID uint16, data []byte, timestamp uint32) *AudioMessage {
	buf := new(bytes.Buffer)
	core.WriteLE(buf, timestamp)
	core.WriteLE(buf, formatID)
	buf.Write(data)

	return &AudioMessage{
		MessageType:  RDPSND_MSG_TYPE_CLIENT_WAVE,
		MessageFlags: 0,
		DataLength:   uint32(buf.Len()),
		Data:         buf.Bytes(),
	}
}
