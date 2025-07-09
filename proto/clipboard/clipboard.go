package clipboard

import (
	"bytes"
	"fmt"
	"io"

	"github.com/GoFeGroup/gordp/core"
	"github.com/GoFeGroup/gordp/glog"
)

// ClipboardFormat represents a clipboard format
type ClipboardFormat uint32

const (
	CLIPRDR_FORMAT_RAW_BITMAP  ClipboardFormat = 0x0002
	CLIPRDR_FORMAT_PALETTE     ClipboardFormat = 0x0009
	CLIPRDR_FORMAT_METAFILE    ClipboardFormat = 0x0003
	CLIPRDR_FORMAT_SYLK        ClipboardFormat = 0x0004
	CLIPRDR_FORMAT_DIF         ClipboardFormat = 0x0005
	CLIPRDR_FORMAT_TIFF        ClipboardFormat = 0x0006
	CLIPRDR_FORMAT_OEMTEXT     ClipboardFormat = 0x0007
	CLIPRDR_FORMAT_DIB         ClipboardFormat = 0x0008
	CLIPRDR_FORMAT_UNICODETEXT ClipboardFormat = 0x000D
	CLIPRDR_FORMAT_HTML        ClipboardFormat = 0x000F
	CLIPRDR_FORMAT_CSV         ClipboardFormat = 0x0010
	CLIPRDR_FORMAT_BIFF        ClipboardFormat = 0x0011
	CLIPRDR_FORMAT_RTF         ClipboardFormat = 0x0012
	CLIPRDR_FORMAT_PNG         ClipboardFormat = 0x0013
	CLIPRDR_FORMAT_JPEG        ClipboardFormat = 0x0014
	CLIPRDR_FORMAT_GIF         ClipboardFormat = 0x0015
	CLIPRDR_FORMAT_FILE_LIST   ClipboardFormat = 0x0016
)

// ClipboardMessageType represents the type of clipboard message
type ClipboardMessageType uint16

const (
	CLIPRDR_MSG_TYPE_CAPABILITIES          ClipboardMessageType = 0x0001
	CLIPRDR_MSG_TYPE_MONITOR_READY         ClipboardMessageType = 0x0002
	CLIPRDR_MSG_TYPE_FORMAT_LIST           ClipboardMessageType = 0x0003
	CLIPRDR_MSG_TYPE_FORMAT_LIST_RESPONSE  ClipboardMessageType = 0x0004
	CLIPRDR_MSG_TYPE_FORMAT_DATA_REQUEST   ClipboardMessageType = 0x0005
	CLIPRDR_MSG_TYPE_FORMAT_DATA_RESPONSE  ClipboardMessageType = 0x0006
	CLIPRDR_MSG_TYPE_TEMP_DIRECTORY        ClipboardMessageType = 0x0007
	CLIPRDR_MSG_TYPE_CLIP_CAPS             ClipboardMessageType = 0x0008
	CLIPRDR_MSG_TYPE_FILECONTENTS_REQUEST  ClipboardMessageType = 0x0009
	CLIPRDR_MSG_TYPE_FILECONTENTS_RESPONSE ClipboardMessageType = 0x000A
	CLIPRDR_MSG_TYPE_LOCK_CLIPDATA         ClipboardMessageType = 0x000B
	CLIPRDR_MSG_TYPE_UNLOCK_CLIPDATA       ClipboardMessageType = 0x000C
)

// ClipboardMessage represents a clipboard message header
type ClipboardMessage struct {
	MessageType  ClipboardMessageType
	MessageFlags uint16
	DataLength   uint32
	Data         []byte
}

// ClipboardCapabilities represents clipboard capabilities
type ClipboardCapabilities struct {
	GeneralFlags uint32
	Pad1         uint16
	Pad2         uint16
	Pad3         uint16
	Pad4         uint16
	Pad5         uint16
	Pad6         uint16
}

// ClipboardFormatList represents a list of clipboard formats
type ClipboardFormatList struct {
	Formats []ClipboardFormat
}

// ClipboardFormatData represents clipboard format data
type ClipboardFormatData struct {
	FormatID ClipboardFormat
	Data     []byte
}

// ClipboardManager manages clipboard operations
type ClipboardManager struct {
	capabilities *ClipboardCapabilities
	formats      []ClipboardFormat
	handler      ClipboardHandler
}

// ClipboardHandler handles clipboard events
type ClipboardHandler interface {
	OnFormatList(formats []ClipboardFormat) error
	OnFormatDataRequest(formatID ClipboardFormat) error
	OnFormatDataResponse(formatID ClipboardFormat, data []byte) error
	OnFileContentsRequest(streamID uint32, listIndex uint32, dwFlags uint32, nPositionLow uint32, nPositionHigh uint32, cbRequested uint32, clipDataID uint32) error
}

// DefaultClipboardHandler provides a default implementation
type DefaultClipboardHandler struct{}

// NewDefaultClipboardHandler creates a new default clipboard handler
func NewDefaultClipboardHandler() *DefaultClipboardHandler {
	return &DefaultClipboardHandler{}
}

// OnFormatList handles format list events
func (h *DefaultClipboardHandler) OnFormatList(formats []ClipboardFormat) error {
	glog.Debugf("Received clipboard format list: %v", formats)
	return nil
}

// OnFormatDataRequest handles format data request events
func (h *DefaultClipboardHandler) OnFormatDataRequest(formatID ClipboardFormat) error {
	glog.Debugf("Received clipboard format data request: %d", formatID)
	return nil
}

// OnFormatDataResponse handles format data response events
func (h *DefaultClipboardHandler) OnFormatDataResponse(formatID ClipboardFormat, data []byte) error {
	glog.Debugf("Received clipboard format data response: %d, %d bytes", formatID, len(data))
	return nil
}

// OnFileContentsRequest handles file contents request events
func (h *DefaultClipboardHandler) OnFileContentsRequest(streamID uint32, listIndex uint32, dwFlags uint32, nPositionLow uint32, nPositionHigh uint32, cbRequested uint32, clipDataID uint32) error {
	glog.Debugf("Received file contents request: streamID=%d, listIndex=%d, flags=%d, position=%d, requested=%d, clipDataID=%d",
		streamID, listIndex, dwFlags, (uint64(nPositionHigh)<<32)|uint64(nPositionLow), cbRequested, clipDataID)
	return nil
}

// NewClipboardManager creates a new clipboard manager
func NewClipboardManager(handler ClipboardHandler) *ClipboardManager {
	if handler == nil {
		handler = NewDefaultClipboardHandler()
	}

	return &ClipboardManager{
		capabilities: &ClipboardCapabilities{
			GeneralFlags: 0x00000001, // CB_USE_LONG_FORMAT_NAMES
		},
		handler: handler,
	}
}

// ReadClipboardMessage reads a clipboard message from the stream
func ReadClipboardMessage(r io.Reader) (*ClipboardMessage, error) {
	msg := &ClipboardMessage{}

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

// Serialize serializes the clipboard message
func (m *ClipboardMessage) Serialize() []byte {
	buf := new(bytes.Buffer)
	core.WriteLE(buf, m.MessageType)
	core.WriteLE(buf, m.MessageFlags)
	core.WriteLE(buf, m.DataLength)
	if len(m.Data) > 0 {
		buf.Write(m.Data)
	}
	return buf.Bytes()
}

// ProcessMessage processes a clipboard message
func (cm *ClipboardManager) ProcessMessage(msg *ClipboardMessage) error {
	switch msg.MessageType {
	case CLIPRDR_MSG_TYPE_CAPABILITIES:
		return cm.handleCapabilities(msg)
	case CLIPRDR_MSG_TYPE_MONITOR_READY:
		return cm.handleMonitorReady(msg)
	case CLIPRDR_MSG_TYPE_FORMAT_LIST:
		return cm.handleFormatList(msg)
	case CLIPRDR_MSG_TYPE_FORMAT_DATA_REQUEST:
		return cm.handleFormatDataRequest(msg)
	case CLIPRDR_MSG_TYPE_FORMAT_DATA_RESPONSE:
		return cm.handleFormatDataResponse(msg)
	case CLIPRDR_MSG_TYPE_FILECONTENTS_REQUEST:
		return cm.handleFileContentsRequest(msg)
	default:
		glog.Debugf("Unhandled clipboard message type: %d", msg.MessageType)
		return nil
	}
}

// handleCapabilities handles capabilities message
func (cm *ClipboardManager) handleCapabilities(msg *ClipboardMessage) error {
	if len(msg.Data) < 16 {
		return fmt.Errorf("invalid capabilities message size")
	}

	capabilities := &ClipboardCapabilities{}
	reader := bytes.NewReader(msg.Data)

	core.ReadLE(reader, &capabilities.GeneralFlags)
	core.ReadLE(reader, &capabilities.Pad1)
	core.ReadLE(reader, &capabilities.Pad2)
	core.ReadLE(reader, &capabilities.Pad3)
	core.ReadLE(reader, &capabilities.Pad4)
	core.ReadLE(reader, &capabilities.Pad5)
	core.ReadLE(reader, &capabilities.Pad6)

	cm.capabilities = capabilities
	glog.Debugf("Received clipboard capabilities: flags=0x%08X", capabilities.GeneralFlags)
	return nil
}

// handleMonitorReady handles monitor ready message
func (cm *ClipboardManager) handleMonitorReady(msg *ClipboardMessage) error {
	glog.Debugf("Clipboard monitor ready")
	return nil
}

// handleFormatList handles format list message
func (cm *ClipboardManager) handleFormatList(msg *ClipboardMessage) error {
	formats := make([]ClipboardFormat, 0)
	reader := bytes.NewReader(msg.Data)

	for reader.Len() > 0 {
		var formatID ClipboardFormat
		if err := core.ReadLE(reader, &formatID); err != nil {
			break
		}
		formats = append(formats, formatID)
	}

	cm.formats = formats
	return cm.handler.OnFormatList(formats)
}

// handleFormatDataRequest handles format data request message
func (cm *ClipboardManager) handleFormatDataRequest(msg *ClipboardMessage) error {
	if len(msg.Data) < 4 {
		return fmt.Errorf("invalid format data request message size")
	}

	var formatID ClipboardFormat
	reader := bytes.NewReader(msg.Data)
	core.ReadLE(reader, &formatID)

	return cm.handler.OnFormatDataRequest(formatID)
}

// handleFormatDataResponse handles format data response message
func (cm *ClipboardManager) handleFormatDataResponse(msg *ClipboardMessage) error {
	if len(msg.Data) < 4 {
		return fmt.Errorf("invalid format data response message size")
	}

	var formatID ClipboardFormat
	reader := bytes.NewReader(msg.Data)
	core.ReadLE(reader, &formatID)

	data := msg.Data[4:]
	return cm.handler.OnFormatDataResponse(formatID, data)
}

// handleFileContentsRequest handles file contents request message
func (cm *ClipboardManager) handleFileContentsRequest(msg *ClipboardMessage) error {
	if len(msg.Data) < 28 {
		return fmt.Errorf("invalid file contents request message size")
	}

	reader := bytes.NewReader(msg.Data)

	var streamID uint32
	var listIndex uint32
	var dwFlags uint32
	var nPositionLow uint32
	var nPositionHigh uint32
	var cbRequested uint32
	var clipDataID uint32

	core.ReadLE(reader, &streamID)
	core.ReadLE(reader, &listIndex)
	core.ReadLE(reader, &dwFlags)
	core.ReadLE(reader, &nPositionLow)
	core.ReadLE(reader, &nPositionHigh)
	core.ReadLE(reader, &cbRequested)
	core.ReadLE(reader, &clipDataID)

	return cm.handler.OnFileContentsRequest(streamID, listIndex, dwFlags, nPositionLow, nPositionHigh, cbRequested, clipDataID)
}

// CreateCapabilitiesMessage creates a capabilities message
func (cm *ClipboardManager) CreateCapabilitiesMessage() *ClipboardMessage {
	buf := new(bytes.Buffer)
	core.WriteLE(buf, cm.capabilities.GeneralFlags)
	core.WriteLE(buf, cm.capabilities.Pad1)
	core.WriteLE(buf, cm.capabilities.Pad2)
	core.WriteLE(buf, cm.capabilities.Pad3)
	core.WriteLE(buf, cm.capabilities.Pad4)
	core.WriteLE(buf, cm.capabilities.Pad5)
	core.WriteLE(buf, cm.capabilities.Pad6)

	return &ClipboardMessage{
		MessageType:  CLIPRDR_MSG_TYPE_CAPABILITIES,
		MessageFlags: 0,
		DataLength:   uint32(buf.Len()),
		Data:         buf.Bytes(),
	}
}

// CreateMonitorReadyMessage creates a monitor ready message
func (cm *ClipboardManager) CreateMonitorReadyMessage() *ClipboardMessage {
	return &ClipboardMessage{
		MessageType:  CLIPRDR_MSG_TYPE_MONITOR_READY,
		MessageFlags: 0,
		DataLength:   0,
		Data:         nil,
	}
}

// CreateFormatListMessage creates a format list message
func (cm *ClipboardManager) CreateFormatListMessage(formats []ClipboardFormat) *ClipboardMessage {
	buf := new(bytes.Buffer)
	for _, format := range formats {
		core.WriteLE(buf, format)
	}

	return &ClipboardMessage{
		MessageType:  CLIPRDR_MSG_TYPE_FORMAT_LIST,
		MessageFlags: 0,
		DataLength:   uint32(buf.Len()),
		Data:         buf.Bytes(),
	}
}

// CreateFormatDataResponseMessage creates a format data response message
func (cm *ClipboardManager) CreateFormatDataResponseMessage(formatID ClipboardFormat, data []byte) *ClipboardMessage {
	buf := new(bytes.Buffer)
	core.WriteLE(buf, formatID)
	buf.Write(data)

	return &ClipboardMessage{
		MessageType:  CLIPRDR_MSG_TYPE_FORMAT_DATA_RESPONSE,
		MessageFlags: 0,
		DataLength:   uint32(buf.Len()),
		Data:         buf.Bytes(),
	}
}

// GetFormatName returns the name of a clipboard format
func GetFormatName(format ClipboardFormat) string {
	switch format {
	case CLIPRDR_FORMAT_RAW_BITMAP:
		return "CF_BITMAP"
	case CLIPRDR_FORMAT_PALETTE:
		return "CF_PALETTE"
	case CLIPRDR_FORMAT_METAFILE:
		return "CF_METAFILEPICT"
	case CLIPRDR_FORMAT_SYLK:
		return "CF_SYLK"
	case CLIPRDR_FORMAT_DIF:
		return "CF_DIF"
	case CLIPRDR_FORMAT_TIFF:
		return "CF_TIFF"
	case CLIPRDR_FORMAT_OEMTEXT:
		return "CF_OEMTEXT"
	case CLIPRDR_FORMAT_DIB:
		return "CF_DIB"
	case CLIPRDR_FORMAT_UNICODETEXT:
		return "CF_UNICODETEXT"
	case CLIPRDR_FORMAT_HTML:
		return "CF_HTML"
	case CLIPRDR_FORMAT_CSV:
		return "CF_CSV"
	case CLIPRDR_FORMAT_BIFF:
		return "CF_BIFF"
	case CLIPRDR_FORMAT_RTF:
		return "CF_RTF"
	case CLIPRDR_FORMAT_PNG:
		return "CF_PNG"
	case CLIPRDR_FORMAT_JPEG:
		return "CF_JPEG"
	case CLIPRDR_FORMAT_GIF:
		return "CF_GIF"
	case CLIPRDR_FORMAT_FILE_LIST:
		return "CF_HDROP"
	default:
		return fmt.Sprintf("Unknown Format (0x%08X)", uint32(format))
	}
}
