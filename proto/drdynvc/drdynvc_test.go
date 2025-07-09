package drdynvc

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDynamicVirtualChannelMessage_Serialize(t *testing.T) {
	msg := &DynamicVirtualChannelMessage{
		MessageType: DVCCREATE_REQ,
		Data:        []byte{0x01, 0x02, 0x03, 0x04},
	}

	serialized := msg.Serialize()
	expected := []byte{0x01, 0x01, 0x02, 0x03, 0x04}
	assert.Equal(t, expected, serialized)
}

func TestReadDynamicVirtualChannelMessage(t *testing.T) {
	data := []byte{0x02, 0x01, 0x02, 0x03, 0x04}
	reader := bytes.NewReader(data)

	msg, err := ReadDynamicVirtualChannelMessage(reader)
	require.NoError(t, err)
	assert.Equal(t, uint8(0x02), msg.MessageType)
	assert.Equal(t, []byte{0x01, 0x02, 0x03, 0x04}, msg.Data)
}

func TestCreateRequest_Serialize(t *testing.T) {
	req := &CreateRequest{
		RequestId:   0x12345678,
		ChannelId:   0x87654321,
		ChannelName: "test_channel",
	}

	serialized := req.Serialize()
	expected := []byte{
		0x78, 0x56, 0x34, 0x12, // RequestId (little endian)
		0x21, 0x43, 0x65, 0x87, // ChannelId (little endian)
		0x74, 0x65, 0x73, 0x74, 0x5f, 0x63, 0x68, 0x61, 0x6e, 0x6e, 0x65, 0x6c, // "test_channel"
		0x00, // null terminator
	}
	assert.Equal(t, expected, serialized)
}

func TestParseCreateRequest(t *testing.T) {
	data := []byte{
		0x78, 0x56, 0x34, 0x12, // RequestId
		0x21, 0x43, 0x65, 0x87, // ChannelId
		0x74, 0x65, 0x73, 0x74, 0x5f, 0x63, 0x68, 0x61, 0x6e, 0x6e, 0x65, 0x6c, // "test_channel"
		0x00, // null terminator
	}

	req, err := ParseCreateRequest(data)
	require.NoError(t, err)
	assert.Equal(t, uint32(0x12345678), req.RequestId)
	assert.Equal(t, uint32(0x87654321), req.ChannelId)
	assert.Equal(t, "test_channel", req.ChannelName)
}

func TestCreateResponse_Serialize(t *testing.T) {
	resp := &CreateResponse{
		RequestId: 0x12345678,
		ChannelId: 0x87654321,
		Status:    DVCCREATE_SUCCESS,
	}

	serialized := resp.Serialize()
	expected := []byte{
		0x78, 0x56, 0x34, 0x12, // RequestId
		0x21, 0x43, 0x65, 0x87, // ChannelId
		0x00, 0x00, 0x00, 0x00, // Status (success)
	}
	assert.Equal(t, expected, serialized)
}

func TestParseCreateResponse(t *testing.T) {
	data := []byte{
		0x78, 0x56, 0x34, 0x12, // RequestId
		0x21, 0x43, 0x65, 0x87, // ChannelId
		0x00, 0x00, 0x00, 0x00, // Status
	}

	resp, err := ParseCreateResponse(data)
	require.NoError(t, err)
	assert.Equal(t, uint32(0x12345678), resp.RequestId)
	assert.Equal(t, uint32(0x87654321), resp.ChannelId)
	assert.Equal(t, uint32(DVCCREATE_SUCCESS), resp.Status)
}

func TestOpenRequest_Serialize(t *testing.T) {
	req := &OpenRequest{
		RequestId: 0x12345678,
		ChannelId: 0x87654321,
	}

	serialized := req.Serialize()
	expected := []byte{
		0x78, 0x56, 0x34, 0x12, // RequestId
		0x21, 0x43, 0x65, 0x87, // ChannelId
	}
	assert.Equal(t, expected, serialized)
}

func TestParseOpenRequest(t *testing.T) {
	data := []byte{
		0x78, 0x56, 0x34, 0x12, // RequestId
		0x21, 0x43, 0x65, 0x87, // ChannelId
	}

	req, err := ParseOpenRequest(data)
	require.NoError(t, err)
	assert.Equal(t, uint32(0x12345678), req.RequestId)
	assert.Equal(t, uint32(0x87654321), req.ChannelId)
}

func TestOpenResponse_Serialize(t *testing.T) {
	resp := &OpenResponse{
		RequestId: 0x12345678,
		ChannelId: 0x87654321,
		Status:    DVCOPEN_SUCCESS,
	}

	serialized := resp.Serialize()
	expected := []byte{
		0x78, 0x56, 0x34, 0x12, // RequestId
		0x21, 0x43, 0x65, 0x87, // ChannelId
		0x00, 0x00, 0x00, 0x00, // Status
	}
	assert.Equal(t, expected, serialized)
}

func TestParseOpenResponse(t *testing.T) {
	data := []byte{
		0x78, 0x56, 0x34, 0x12, // RequestId
		0x21, 0x43, 0x65, 0x87, // ChannelId
		0x00, 0x00, 0x00, 0x00, // Status
	}

	resp, err := ParseOpenResponse(data)
	require.NoError(t, err)
	assert.Equal(t, uint32(0x12345678), resp.RequestId)
	assert.Equal(t, uint32(0x87654321), resp.ChannelId)
	assert.Equal(t, uint32(DVCOPEN_SUCCESS), resp.Status)
}

func TestCloseRequest_Serialize(t *testing.T) {
	req := &CloseRequest{
		RequestId: 0x12345678,
		ChannelId: 0x87654321,
	}

	serialized := req.Serialize()
	expected := []byte{
		0x78, 0x56, 0x34, 0x12, // RequestId
		0x21, 0x43, 0x65, 0x87, // ChannelId
	}
	assert.Equal(t, expected, serialized)
}

func TestParseCloseRequest(t *testing.T) {
	data := []byte{
		0x78, 0x56, 0x34, 0x12, // RequestId
		0x21, 0x43, 0x65, 0x87, // ChannelId
	}

	req, err := ParseCloseRequest(data)
	require.NoError(t, err)
	assert.Equal(t, uint32(0x12345678), req.RequestId)
	assert.Equal(t, uint32(0x87654321), req.ChannelId)
}

func TestCloseResponse_Serialize(t *testing.T) {
	resp := &CloseResponse{
		RequestId: 0x12345678,
		ChannelId: 0x87654321,
		Status:    DVCCLOSE_SUCCESS,
	}

	serialized := resp.Serialize()
	expected := []byte{
		0x78, 0x56, 0x34, 0x12, // RequestId
		0x21, 0x43, 0x65, 0x87, // ChannelId
		0x00, 0x00, 0x00, 0x00, // Status
	}
	assert.Equal(t, expected, serialized)
}

func TestParseCloseResponse(t *testing.T) {
	data := []byte{
		0x78, 0x56, 0x34, 0x12, // RequestId
		0x21, 0x43, 0x65, 0x87, // ChannelId
		0x00, 0x00, 0x00, 0x00, // Status
	}

	resp, err := ParseCloseResponse(data)
	require.NoError(t, err)
	assert.Equal(t, uint32(0x12345678), resp.RequestId)
	assert.Equal(t, uint32(0x87654321), resp.ChannelId)
	assert.Equal(t, uint32(DVCCLOSE_SUCCESS), resp.Status)
}

func TestDataMessage_Serialize(t *testing.T) {
	msg := &DataMessage{
		ChannelId: 0x87654321,
		Data:      []byte{0x01, 0x02, 0x03, 0x04},
	}

	serialized := msg.Serialize()
	expected := []byte{
		0x21, 0x43, 0x65, 0x87, // ChannelId
		0x01, 0x02, 0x03, 0x04, // Data
	}
	assert.Equal(t, expected, serialized)
}

func TestParseDataMessage(t *testing.T) {
	data := []byte{
		0x21, 0x43, 0x65, 0x87, // ChannelId
		0x01, 0x02, 0x03, 0x04, // Data
	}

	msg, err := ParseDataMessage(data)
	require.NoError(t, err)
	assert.Equal(t, uint32(0x87654321), msg.ChannelId)
	assert.Equal(t, []byte{0x01, 0x02, 0x03, 0x04}, msg.Data)
}

func TestDynamicVirtualChannelManager(t *testing.T) {
	manager := NewDynamicVirtualChannelManager()

	// Test registering a channel
	err := manager.RegisterChannelWithID(1, "test_channel", nil)
	require.NoError(t, err)

	// Test getting a channel
	channel, exists := manager.GetChannel(1)
	require.True(t, exists)
	assert.Equal(t, "test_channel", channel.ChannelName)
	assert.Equal(t, uint32(1), channel.ChannelId)
	assert.False(t, channel.IsOpen)

	// Test getting non-existent channel
	_, exists = manager.GetChannel(999)
	assert.False(t, exists)
}

func TestDefaultDynamicVirtualChannelHandler(t *testing.T) {
	handler := NewDefaultDynamicVirtualChannelHandler()

	// Test all handler methods (they should not panic)
	err := handler.OnChannelCreated(1, "test_channel")
	assert.NoError(t, err)

	err = handler.OnChannelOpened(1)
	assert.NoError(t, err)

	err = handler.OnDataReceived(1, []byte{0x01, 0x02, 0x03})
	assert.NoError(t, err)

	err = handler.OnChannelClosed(1)
	assert.NoError(t, err)
}

func TestParseCreateRequest_InvalidData(t *testing.T) {
	// Test with insufficient data
	data := []byte{0x01, 0x02, 0x03} // Less than 8 bytes
	_, err := ParseCreateRequest(data)
	assert.Error(t, err)
}

func TestParseCreateResponse_InvalidData(t *testing.T) {
	// Test with insufficient data
	data := []byte{0x01, 0x02, 0x03} // Less than 12 bytes
	_, err := ParseCreateResponse(data)
	assert.Error(t, err)
}

func TestParseOpenRequest_InvalidData(t *testing.T) {
	// Test with insufficient data
	data := []byte{0x01, 0x02, 0x03} // Less than 8 bytes
	_, err := ParseOpenRequest(data)
	assert.Error(t, err)
}

func TestParseOpenResponse_InvalidData(t *testing.T) {
	// Test with insufficient data
	data := []byte{0x01, 0x02, 0x03} // Less than 12 bytes
	_, err := ParseOpenResponse(data)
	assert.Error(t, err)
}

func TestParseCloseRequest_InvalidData(t *testing.T) {
	// Test with insufficient data
	data := []byte{0x01, 0x02, 0x03} // Less than 8 bytes
	_, err := ParseCloseRequest(data)
	assert.Error(t, err)
}

func TestParseCloseResponse_InvalidData(t *testing.T) {
	// Test with insufficient data
	data := []byte{0x01, 0x02, 0x03} // Less than 12 bytes
	_, err := ParseCloseResponse(data)
	assert.Error(t, err)
}

func TestParseDataMessage_InvalidData(t *testing.T) {
	// Test with insufficient data
	data := []byte{0x01, 0x02, 0x03} // Less than 4 bytes
	_, err := ParseDataMessage(data)
	assert.Error(t, err)
}

func TestMessageTypeConstants(t *testing.T) {
	// Test that all message type constants are unique
	types := map[uint8]string{
		DVCCREATE_REQ:      "DVCCREATE_REQ",
		DVCCREATE_RSP:      "DVCCREATE_RSP",
		DVCOPEN_REQ:        "DVCOPEN_REQ",
		DVCOPEN_RSP:        "DVCOPEN_RSP",
		DVCCLOSE_REQ:       "DVCCLOSE_REQ",
		DVCCLOSE_RSP:       "DVCCLOSE_RSP",
		DVCDATA_FIRST:      "DVCDATA_FIRST",
		DVCDATA:            "DVCDATA",
		DVCDATA_LAST:       "DVCDATA_LAST",
		DVCDATA_FIRST_LAST: "DVCDATA_FIRST_LAST",
	}

	// Verify all constants are defined
	assert.Equal(t, 10, len(types))
}

func TestStatusConstants(t *testing.T) {
	// Test that status constants are properly defined
	assert.Equal(t, 0x00000000, DVCCREATE_SUCCESS)
	assert.Equal(t, 0x00000001, DVCCREATE_FAILED)
	assert.Equal(t, 0x00000000, DVCOPEN_SUCCESS)
	assert.Equal(t, 0x00000001, DVCOPEN_FAILED)
	assert.Equal(t, 0x00000000, DVCCLOSE_SUCCESS)
	assert.Equal(t, 0x00000001, DVCCLOSE_FAILED)
}
