package nla

import (
	"bytes"
	"testing"
)

func TestNewNTLMv2ClientChallenge(t *testing.T) {
	// Test data
	serverInfo := []byte{0x01, 0x00, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	timestamp := []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}

	// Create client challenge
	challenge := NewNTLMv2ClientChallenge(serverInfo, timestamp)

	// Verify basic structure
	if challenge.Must.RespType != 0x01 {
		t.Errorf("Expected RespType 0x01, got 0x%02x", challenge.Must.RespType)
	}

	if challenge.Must.HiRespType != 0x01 {
		t.Errorf("Expected HiRespType 0x01, got 0x%02x", challenge.Must.HiRespType)
	}

	if !bytes.Equal(challenge.Must.Timestamp[:], timestamp[:8]) {
		t.Errorf("Expected timestamp %x, got %x", timestamp[:8], challenge.Must.Timestamp[:])
	}

	// Verify client challenge is not zero
	zeroChallenge := [8]byte{}
	if bytes.Equal(challenge.Must.ChallengeFromClient[:], zeroChallenge[:]) {
		t.Error("Expected non-zero client challenge")
	}
}

func TestAddChannelBinding(t *testing.T) {
	// Test data
	serverInfo := []byte{0x01, 0x00, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	timestamp := []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
	channelBindingToken := []byte{0x01, 0x02, 0x03, 0x04, 0x05}

	// Create client challenge
	challenge := NewNTLMv2ClientChallenge(serverInfo, timestamp)

	// Count initial AVPairs
	initialCount := len(challenge.Optional.AvPairs)

	// Add channel binding
	challenge.AddChannelBinding(channelBindingToken)

	// Verify AVPairs count increased by 1
	if len(challenge.Optional.AvPairs) != initialCount+1 {
		t.Errorf("Expected %d AVPairs, got %d", initialCount+1, len(challenge.Optional.AvPairs))
	}

	// Verify channel binding is present
	found := false
	for _, pair := range challenge.Optional.AvPairs {
		if pair.Must.Id == MsvChannelBindings {
			if !bytes.Equal(pair.Optional.Value, channelBindingToken) {
				t.Errorf("Expected channel binding token %x, got %x", channelBindingToken, pair.Optional.Value)
			}
			found = true
			break
		}
	}

	if !found {
		t.Error("Channel binding AVPair not found")
	}
}

func TestAddChannelBindingEmptyToken(t *testing.T) {
	// Test data
	serverInfo := []byte{0x01, 0x00, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	timestamp := []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}

	// Create client challenge
	challenge := NewNTLMv2ClientChallenge(serverInfo, timestamp)

	// Count initial AVPairs
	initialCount := len(challenge.Optional.AvPairs)

	// Add empty channel binding
	challenge.AddChannelBinding([]byte{})

	// Verify AVPairs count didn't change
	if len(challenge.Optional.AvPairs) != initialCount {
		t.Errorf("Expected %d AVPairs, got %d", initialCount, len(challenge.Optional.AvPairs))
	}
}

func TestSerializeWithChannelBinding(t *testing.T) {
	// Test data
	serverInfo := []byte{0x01, 0x00, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	timestamp := []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
	channelBindingToken := []byte{0x01, 0x02, 0x03, 0x04, 0x05}

	// Create client challenge
	challenge := NewNTLMv2ClientChallenge(serverInfo, timestamp)
	challenge.AddChannelBinding(channelBindingToken)

	// Serialize
	data := challenge.Serialize()

	// Verify serialized data contains channel binding
	if len(data) == 0 {
		t.Error("Serialized data is empty")
	}

	// The serialized data should be longer than without channel binding
	if len(data) < 32 {
		t.Errorf("Serialized data too short: %d bytes", len(data))
	}
}
