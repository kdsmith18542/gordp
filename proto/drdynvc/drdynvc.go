package drdynvc

import (
	"bytes"
	"fmt"
	"io"

	"github.com/GoFeGroup/gordp/core"
	"github.com/GoFeGroup/gordp/glog"
)

// Dynamic Virtual Channel Message Types
const (
	DVCCREATE_REQ      = 0x01 // Create Request
	DVCCREATE_RSP      = 0x02 // Create Response
	DVCOPEN_REQ        = 0x03 // Open Request
	DVCOPEN_RSP        = 0x04 // Open Response
	DVCCLOSE_REQ       = 0x05 // Close Request
	DVCCLOSE_RSP       = 0x06 // Close Response
	DVCDATA_FIRST      = 0x10 // Data First
	DVCDATA            = 0x20 // Data
	DVCDATA_LAST       = 0x30 // Data Last
	DVCDATA_FIRST_LAST = 0x40 // Data First and Last
)

// Dynamic Virtual Channel Status Codes
const (
	DVCCREATE_SUCCESS = 0x00000000
	DVCCREATE_FAILED  = 0x00000001
	DVCOPEN_SUCCESS   = 0x00000000
	DVCOPEN_FAILED    = 0x00000001
	DVCCLOSE_SUCCESS  = 0x00000000
	DVCCLOSE_FAILED   = 0x00000001
)

// DynamicVirtualChannelMessage represents a dynamic virtual channel message
type DynamicVirtualChannelMessage struct {
	MessageType uint8
	Data        []byte
}

// CreateRequest represents a DVC create request
type CreateRequest struct {
	RequestId   uint32
	ChannelId   uint32
	ChannelName string
}

// CreateResponse represents a DVC create response
type CreateResponse struct {
	RequestId uint32
	ChannelId uint32
	Status    uint32
}

// OpenRequest represents a DVC open request
type OpenRequest struct {
	RequestId uint32
	ChannelId uint32
}

// OpenResponse represents a DVC open response
type OpenResponse struct {
	RequestId uint32
	ChannelId uint32
	Status    uint32
}

// CloseRequest represents a DVC close request
type CloseRequest struct {
	RequestId uint32
	ChannelId uint32
}

// CloseResponse represents a DVC close response
type CloseResponse struct {
	RequestId uint32
	ChannelId uint32
	Status    uint32
}

// DataMessage represents a DVC data message
type DataMessage struct {
	ChannelId uint32
	Data      []byte
}

// DynamicVirtualChannelManager manages dynamic virtual channels
type DynamicVirtualChannelManager struct {
	channels      map[uint32]*DynamicVirtualChannel
	requests      map[uint32]chan interface{} // Request ID to response channel
	nextRequestId uint32
}

// DynamicVirtualChannel represents a dynamic virtual channel
type DynamicVirtualChannel struct {
	ChannelId   uint32
	ChannelName string
	IsOpen      bool
	Handler     DynamicVirtualChannelHandler
}

// DynamicVirtualChannelHandler handles dynamic virtual channel events
type DynamicVirtualChannelHandler interface {
	OnChannelCreated(channelId uint32, channelName string) error
	OnChannelOpened(channelId uint32) error
	OnChannelClosed(channelId uint32) error
	OnDataReceived(channelId uint32, data []byte) error
}

// DefaultDynamicVirtualChannelHandler provides a default implementation
type DefaultDynamicVirtualChannelHandler struct{}

// NewDefaultDynamicVirtualChannelHandler creates a new default handler
func NewDefaultDynamicVirtualChannelHandler() *DefaultDynamicVirtualChannelHandler {
	return &DefaultDynamicVirtualChannelHandler{}
}

// OnChannelCreated handles channel creation events
func (h *DefaultDynamicVirtualChannelHandler) OnChannelCreated(channelId uint32, channelName string) error {
	glog.Debugf("Dynamic virtual channel created: %s (ID: %d)", channelName, channelId)
	return nil
}

// OnChannelOpened handles channel open events
func (h *DefaultDynamicVirtualChannelHandler) OnChannelOpened(channelId uint32) error {
	glog.Debugf("Dynamic virtual channel opened: ID: %d", channelId)
	return nil
}

// OnChannelClosed handles channel close events
func (h *DefaultDynamicVirtualChannelHandler) OnChannelClosed(channelId uint32) error {
	glog.Debugf("Dynamic virtual channel closed: ID: %d", channelId)
	return nil
}

// OnDataReceived handles data received events
func (h *DefaultDynamicVirtualChannelHandler) OnDataReceived(channelId uint32, data []byte) error {
	glog.Debugf("Dynamic virtual channel data received: ID: %d, %d bytes", channelId, len(data))
	return nil
}

// NewDynamicVirtualChannelManager creates a new dynamic virtual channel manager
func NewDynamicVirtualChannelManager() *DynamicVirtualChannelManager {
	return &DynamicVirtualChannelManager{
		channels:      make(map[uint32]*DynamicVirtualChannel),
		requests:      make(map[uint32]chan interface{}),
		nextRequestId: 1,
	}
}

// RegisterChannel registers a dynamic virtual channel
func (m *DynamicVirtualChannelManager) RegisterChannel(channelName string, handler DynamicVirtualChannelHandler) error {
	channel := &DynamicVirtualChannel{
		ChannelName: channelName,
		Handler:     handler,
	}
	if handler == nil {
		channel.Handler = NewDefaultDynamicVirtualChannelHandler()
	}
	// Channel ID will be assigned when created
	return nil
}

// RegisterChannelWithID registers a dynamic virtual channel with a specific ID
func (m *DynamicVirtualChannelManager) RegisterChannelWithID(channelId uint32, channelName string, handler DynamicVirtualChannelHandler) error {
	channel := &DynamicVirtualChannel{
		ChannelId:   channelId,
		ChannelName: channelName,
		Handler:     handler,
		IsOpen:      false,
	}
	if handler == nil {
		channel.Handler = NewDefaultDynamicVirtualChannelHandler()
	}

	m.channels[channelId] = channel
	return nil
}

// GetChannel retrieves a dynamic virtual channel by ID
func (m *DynamicVirtualChannelManager) GetChannel(channelId uint32) (*DynamicVirtualChannel, bool) {
	channel, exists := m.channels[channelId]
	return channel, exists
}

// ReadDynamicVirtualChannelMessage reads a dynamic virtual channel message
func ReadDynamicVirtualChannelMessage(r io.Reader) (*DynamicVirtualChannelMessage, error) {
	msg := &DynamicVirtualChannelMessage{}

	// Read message type
	core.ReadLE(r, &msg.MessageType)

	// Read remaining data
	remainingData, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read message data: %w", err)
	}
	msg.Data = remainingData

	return msg, nil
}

// Serialize serializes the dynamic virtual channel message
func (m *DynamicVirtualChannelMessage) Serialize() []byte {
	buf := new(bytes.Buffer)
	core.WriteLE(buf, m.MessageType)
	if len(m.Data) > 0 {
		buf.Write(m.Data)
	}
	return buf.Bytes()
}

// ParseCreateRequest parses a create request from message data
func ParseCreateRequest(data []byte) (*CreateRequest, error) {
	if len(data) < 8 {
		return nil, fmt.Errorf("invalid create request data size")
	}

	reader := bytes.NewReader(data)
	req := &CreateRequest{}

	core.ReadLE(reader, &req.RequestId)
	core.ReadLE(reader, &req.ChannelId)

	// Read channel name (null-terminated string)
	nameBytes := data[8:]
	if len(nameBytes) > 0 {
		// Find null terminator
		nullIndex := bytes.IndexByte(nameBytes, 0)
		if nullIndex >= 0 {
			req.ChannelName = string(nameBytes[:nullIndex])
		} else {
			req.ChannelName = string(nameBytes)
		}
	}

	return req, nil
}

// ParseCreateResponse parses a create response from message data
func ParseCreateResponse(data []byte) (*CreateResponse, error) {
	if len(data) < 12 {
		return nil, fmt.Errorf("invalid create response data size")
	}

	reader := bytes.NewReader(data)
	resp := &CreateResponse{}

	core.ReadLE(reader, &resp.RequestId)
	core.ReadLE(reader, &resp.ChannelId)
	core.ReadLE(reader, &resp.Status)

	return resp, nil
}

// ParseOpenRequest parses an open request from message data
func ParseOpenRequest(data []byte) (*OpenRequest, error) {
	if len(data) < 8 {
		return nil, fmt.Errorf("invalid open request data size")
	}

	reader := bytes.NewReader(data)
	req := &OpenRequest{}

	core.ReadLE(reader, &req.RequestId)
	core.ReadLE(reader, &req.ChannelId)

	return req, nil
}

// ParseOpenResponse parses an open response from message data
func ParseOpenResponse(data []byte) (*OpenResponse, error) {
	if len(data) < 12 {
		return nil, fmt.Errorf("invalid open response data size")
	}

	reader := bytes.NewReader(data)
	resp := &OpenResponse{}

	core.ReadLE(reader, &resp.RequestId)
	core.ReadLE(reader, &resp.ChannelId)
	core.ReadLE(reader, &resp.Status)

	return resp, nil
}

// ParseCloseRequest parses a close request from message data
func ParseCloseRequest(data []byte) (*CloseRequest, error) {
	if len(data) < 8 {
		return nil, fmt.Errorf("invalid close request data size")
	}

	reader := bytes.NewReader(data)
	req := &CloseRequest{}

	core.ReadLE(reader, &req.RequestId)
	core.ReadLE(reader, &req.ChannelId)

	return req, nil
}

// ParseCloseResponse parses a close response from message data
func ParseCloseResponse(data []byte) (*CloseResponse, error) {
	if len(data) < 12 {
		return nil, fmt.Errorf("invalid close response data size")
	}

	reader := bytes.NewReader(data)
	resp := &CloseResponse{}

	core.ReadLE(reader, &resp.RequestId)
	core.ReadLE(reader, &resp.ChannelId)
	core.ReadLE(reader, &resp.Status)

	return resp, nil
}

// ParseDataMessage parses a data message from message data
func ParseDataMessage(data []byte) (*DataMessage, error) {
	if len(data) < 4 {
		return nil, fmt.Errorf("invalid data message size")
	}

	reader := bytes.NewReader(data)
	msg := &DataMessage{}

	core.ReadLE(reader, &msg.ChannelId)
	msg.Data = data[4:]

	return msg, nil
}

// SerializeCreateRequest serializes a create request
func (req *CreateRequest) Serialize() []byte {
	buf := new(bytes.Buffer)
	core.WriteLE(buf, req.RequestId)
	core.WriteLE(buf, req.ChannelId)
	buf.WriteString(req.ChannelName)
	buf.WriteByte(0) // Null terminator
	return buf.Bytes()
}

// SerializeCreateResponse serializes a create response
func (resp *CreateResponse) Serialize() []byte {
	buf := new(bytes.Buffer)
	core.WriteLE(buf, resp.RequestId)
	core.WriteLE(buf, resp.ChannelId)
	core.WriteLE(buf, resp.Status)
	return buf.Bytes()
}

// SerializeOpenRequest serializes an open request
func (req *OpenRequest) Serialize() []byte {
	buf := new(bytes.Buffer)
	core.WriteLE(buf, req.RequestId)
	core.WriteLE(buf, req.ChannelId)
	return buf.Bytes()
}

// SerializeOpenResponse serializes an open response
func (resp *OpenResponse) Serialize() []byte {
	buf := new(bytes.Buffer)
	core.WriteLE(buf, resp.RequestId)
	core.WriteLE(buf, resp.ChannelId)
	core.WriteLE(buf, resp.Status)
	return buf.Bytes()
}

// SerializeCloseRequest serializes a close request
func (req *CloseRequest) Serialize() []byte {
	buf := new(bytes.Buffer)
	core.WriteLE(buf, req.RequestId)
	core.WriteLE(buf, req.ChannelId)
	return buf.Bytes()
}

// SerializeCloseResponse serializes a close response
func (resp *CloseResponse) Serialize() []byte {
	buf := new(bytes.Buffer)
	core.WriteLE(buf, resp.RequestId)
	core.WriteLE(buf, resp.ChannelId)
	core.WriteLE(buf, resp.Status)
	return buf.Bytes()
}

// SerializeDataMessage serializes a data message
func (msg *DataMessage) Serialize() []byte {
	buf := new(bytes.Buffer)
	core.WriteLE(buf, msg.ChannelId)
	buf.Write(msg.Data)
	return buf.Bytes()
}
