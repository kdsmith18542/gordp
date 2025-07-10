package x224

import (
	"bytes"
	"testing"

	"github.com/kdsmith18542/gordp/core"
	"github.com/stretchr/testify/assert"
)

func TestReadX224Header(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		expected *Header
		wantErr  bool
	}{
		{
			name: "valid header",
			data: []byte{0x02, 0xf0, 0x80, 0x7f, 0x65, 0x82, 0x01, 0x94},
			expected: &Header{
				Length:  0x02,
				PduType: 0xf0,
				DstRef:  0x807f,
				SrcRef:  0x6582,
				Flags:   0x01,
			},
			wantErr: false,
		},
		{
			name:     "incomplete header",
			data:     []byte{0x02, 0xf0, 0x80},
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := bytes.NewReader(tt.data)
			header := &Header{}

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
				assert.Equal(t, tt.expected.PduType, header.PduType)
				assert.Equal(t, tt.expected.Length, header.Length)
			}
		})
	}
}

func TestWriteX224Header(t *testing.T) {
	header := &Header{
		Length:  0x02,
		PduType: 0xf0,
		DstRef:  0x807f,
		SrcRef:  0x6582,
		Flags:   0x01,
	}

	var buf bytes.Buffer
	header.Write(&buf)

	// Verify header was written correctly
	assert.Greater(t, buf.Len(), 0)
}

func BenchmarkReadX224Header(b *testing.B) {
	data := []byte{0x02, 0xf0, 0x80, 0x7f, 0x65, 0x82, 0x01, 0x94}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reader := bytes.NewReader(data)
		header := &Header{}
		header.Read(reader)
	}
}

func BenchmarkWriteX224Header(b *testing.B) {
	header := &Header{
		Length:  0x02,
		PduType: 0xf0,
		DstRef:  0x807f,
		SrcRef:  0x6582,
		Flags:   0x01,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		header.Write(&buf)
	}
}
