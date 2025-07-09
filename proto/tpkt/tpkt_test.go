package tpkt

import (
	"bytes"
	"testing"

	"github.com/GoFeGroup/gordp/core"
	"github.com/stretchr/testify/assert"
)

func TestReadTPKTHeader(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		expected *Header
		wantErr  bool
	}{
		{
			name: "valid header",
			data: []byte{0x03, 0x00, 0x00, 0x08},
			expected: &Header{
				Version:  3,
				Reserved: 0,
				Length:   8,
			},
			wantErr: false,
		},
		{
			name:     "invalid version",
			data:     []byte{0x02, 0x00, 0x00, 0x08},
			expected: nil,
			wantErr:  true,
		},
		{
			name:     "incomplete header",
			data:     []byte{0x03, 0x00, 0x00},
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := bytes.NewReader(tt.data)
			header := &Header{}

			// Use TryCatch to handle panics
			var err error
			core.TryCatch(func() {
				header.Read(reader)
			}, func(e any) {
				err = e.(error)
			})

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected.Version, header.Version)
				assert.Equal(t, tt.expected.Length, header.Length)
			}
		})
	}
}

func TestWriteTPKTHeader(t *testing.T) {
	header := &Header{
		Version:  3,
		Reserved: 0,
		Length:   8,
	}

	var buf bytes.Buffer
	header.Write(&buf)

	expected := []byte{0x03, 0x00, 0x00, 0x08}
	assert.Equal(t, expected, buf.Bytes())
}

func TestReadTPKTPacket(t *testing.T) {
	// Create a complete TPKT packet
	packetData := []byte{0x03, 0x00, 0x00, 0x0A, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06}
	reader := bytes.NewReader(packetData)

	data := Read(reader)
	expected := []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06}
	assert.Equal(t, expected, data)
}

func TestWriteTPKTPacket(t *testing.T) {
	data := []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06}
	var buf bytes.Buffer

	Write(&buf, data)

	// Verify the packet can be read back correctly
	reader := bytes.NewReader(buf.Bytes())
	readData := Read(reader)
	assert.Equal(t, data, readData)
}

func TestTPKTPacketLargeData(t *testing.T) {
	// Test large packet
	largeData := make([]byte, 8192)
	for i := range largeData {
		largeData[i] = byte(i % 256)
	}

	var buf bytes.Buffer
	Write(&buf, largeData)

	// Verify the packet can be read back correctly
	reader := bytes.NewReader(buf.Bytes())
	readData := Read(reader)
	assert.Equal(t, largeData, readData)
}

func TestTPKTPacketInvalidLength(t *testing.T) {
	// Test packet with invalid length (too short)
	invalidPacket := []byte{0x03, 0x00, 0x00, 0x05}
	reader := bytes.NewReader(invalidPacket)

	// Should panic due to invalid length
	assert.Panics(t, func() {
		Read(reader)
	})
}

func TestTPKTPacketEOF(t *testing.T) {
	// Test reading from empty reader
	reader := bytes.NewReader([]byte{})

	// Should panic due to EOF
	assert.Panics(t, func() {
		Read(reader)
	})
}

func TestWriteTPKTPacketTooLarge(t *testing.T) {
	// Test writing data that's too large
	largeData := make([]byte, 0x10000) // 64KB

	var buf bytes.Buffer

	// Should panic due to data too large
	assert.Panics(t, func() {
		Write(&buf, largeData)
	})
}

func BenchmarkReadTPKTHeader(b *testing.B) {
	data := []byte{0x03, 0x00, 0x00, 0x08}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reader := bytes.NewReader(data)
		header := &Header{}
		header.Read(reader)
	}
}

func BenchmarkWriteTPKTHeader(b *testing.B) {
	header := &Header{
		Version:  3,
		Reserved: 0,
		Length:   8,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		header.Write(&buf)
	}
}

func BenchmarkReadTPKTPacket(b *testing.B) {
	// Create test packet
	data := make([]byte, 1024)
	for i := range data {
		data[i] = byte(i % 256)
	}

	var buf bytes.Buffer
	Write(&buf, data)
	packetData := buf.Bytes()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reader := bytes.NewReader(packetData)
		Read(reader)
	}
}

func BenchmarkWriteTPKTPacket(b *testing.B) {
	data := make([]byte, 1024)
	for i := range data {
		data[i] = byte(i % 256)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		Write(&buf, data)
	}
}
