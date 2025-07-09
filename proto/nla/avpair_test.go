package nla

import (
	"bytes"
	"testing"
)

func TestCreateChannelBindingAVPair(t *testing.T) {
	// Test data
	token := []byte{0x01, 0x02, 0x03, 0x04, 0x05}

	// Create channel binding AVPair
	avPair := CreateChannelBindingAVPair(token)

	// Verify the AVPair structure
	if avPair.Must.Id != MsvChannelBindings {
		t.Errorf("Expected AVPair ID %d, got %d", MsvChannelBindings, avPair.Must.Id)
	}

	if avPair.Must.Len != uint16(len(token)) {
		t.Errorf("Expected AVPair length %d, got %d", len(token), avPair.Must.Len)
	}

	if !bytes.Equal(avPair.Optional.Value, token) {
		t.Errorf("Expected AVPair value %x, got %x", token, avPair.Optional.Value)
	}
}

func TestGetChannelBindings(t *testing.T) {
	// Create test AVPairs with channel binding
	token := []byte{0x01, 0x02, 0x03, 0x04, 0x05}
	channelBindingPair := CreateChannelBindingAVPair(token)

	avPairs := AVPairs{
		{Must: struct {
			Id  uint16
			Len uint16
		}{Id: MsvAvTimestamp, Len: 8}, Optional: struct{ Value []byte }{Value: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}}},
		channelBindingPair,
		{Must: struct {
			Id  uint16
			Len uint16
		}{Id: MsvAvEOL, Len: 0}, Optional: struct{ Value []byte }{Value: []byte{}}},
	}

	// Test getting channel binding
	result := avPairs.GetChannelBindings()
	if !bytes.Equal(result, token) {
		t.Errorf("Expected channel binding token %x, got %x", token, result)
	}
}

func TestGetChannelBindingsNotFound(t *testing.T) {
	// Create test AVPairs without channel binding
	avPairs := AVPairs{
		{Must: struct {
			Id  uint16
			Len uint16
		}{Id: MsvAvTimestamp, Len: 8}, Optional: struct{ Value []byte }{Value: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}}},
		{Must: struct {
			Id  uint16
			Len uint16
		}{Id: MsvAvEOL, Len: 0}, Optional: struct{ Value []byte }{Value: []byte{}}},
	}

	// Test getting channel binding when not present
	result := avPairs.GetChannelBindings()
	if result != nil {
		t.Errorf("Expected nil channel binding token, got %x", result)
	}
}

func TestAVPairWrite(t *testing.T) {
	// Test AVPair serialization
	token := []byte{0x01, 0x02, 0x03, 0x04, 0x05}
	avPair := CreateChannelBindingAVPair(token)

	var buf bytes.Buffer
	avPair.Write(&buf)

	// Verify serialized data
	expected := []byte{
		0x0A, 0x00, // ID (MsvChannelBindings)
		0x05, 0x00, // Length
		0x01, 0x02, 0x03, 0x04, 0x05, // Value
	}

	if !bytes.Equal(buf.Bytes(), expected) {
		t.Errorf("Expected serialized data %x, got %x", expected, buf.Bytes())
	}
}
