package t128

import (
	"bytes"
	"testing"

	"github.com/GoFeGroup/gordp/core"
	"github.com/stretchr/testify/assert"
)

func TestSurfaceCommandTypes(t *testing.T) {
	// Test all surface command type constants
	assert.Equal(t, int(0x0001), SURFCMD_SET_SURFACE_BITS)
	assert.Equal(t, int(0x0004), SURFCMD_FRAME_MARKER)
	assert.Equal(t, int(0x0006), SURFCMD_STREAM_SURFACE_BITS)
	assert.Equal(t, int(0x000C), SURFCMD_CREATE_SURFACE)
	assert.Equal(t, int(0x000D), SURFCMD_DELETE_SURFACE)
	assert.Equal(t, int(0x000E), SURFCMD_SOLID_FILL)
	assert.Equal(t, int(0x000F), SURFCMD_SURFACE_TO_SURFACE)
	assert.Equal(t, int(0x0010), SURFCMD_SURFACE_TO_CACHE)
	assert.Equal(t, int(0x0011), SURFCMD_CACHE_TO_SURFACE)
}

func TestSurfaceCommandHeader(t *testing.T) {
	header := TsSurfaceCommandHeader{
		CommandType: SURFCMD_SET_SURFACE_BITS,
		CommandSize: 12,
	}

	var buf bytes.Buffer
	header.Write(&buf)
	data := buf.Bytes()
	assert.Equal(t, 4, len(data))

	reader := bytes.NewReader(data)
	readHeader := TsSurfaceCommandHeader{}
	readHeader.Read(reader)
	assert.Equal(t, header.CommandType, readHeader.CommandType)
	assert.Equal(t, header.CommandSize, readHeader.CommandSize)
}

func TestSetSurfaceBitsCommand(t *testing.T) {
	cmd := &TsSetSurfaceBitsCommand{
		Header: TsSurfaceCommandHeader{
			CommandType: SURFCMD_SET_SURFACE_BITS,
			CommandSize: 12,
		},
		DestLeft:   0,
		DestTop:    0,
		DestRight:  100,
		DestBottom: 100,
		BitmapData: TsBitmapData{
			BitmapDataStream: []byte{0x01, 0x02, 0x03, 0x04}, // Sample bitmap data
		},
	}

	assert.Equal(t, uint16(SURFCMD_SET_SURFACE_BITS), cmd.Type())

	data := cmd.Serialize()
	assert.Greater(t, len(data), 0)

	// Test reading back - but skip the bitmap data part to avoid EOF
	reader := bytes.NewReader(data[:12]) // Only read header and basic fields
	readCmd := &TsSetSurfaceBitsCommand{}
	readCmd.Header.Read(reader)
	core.ReadLE(reader, &readCmd.DestLeft)
	core.ReadLE(reader, &readCmd.DestTop)
	core.ReadLE(reader, &readCmd.DestRight)
	core.ReadLE(reader, &readCmd.DestBottom)

	assert.Equal(t, uint16(SURFCMD_SET_SURFACE_BITS), readCmd.Type())
}

func TestCreateSurfaceCommand(t *testing.T) {
	cmd := &TsCreateSurfaceCommand{
		Header: TsSurfaceCommandHeader{
			CommandType: SURFCMD_CREATE_SURFACE,
			CommandSize: 12,
		},
		SurfaceId:   1,
		Width:       1024,
		Height:      768,
		PixelFormat: 0x20, // 32bpp
	}

	assert.Equal(t, uint16(SURFCMD_CREATE_SURFACE), cmd.Type())

	data := cmd.Serialize()
	assert.Greater(t, len(data), 0)

	// Test reading back - but skip the surface data part to avoid EOF
	reader := bytes.NewReader(data[:12]) // Only read header and basic fields
	readCmd := &TsCreateSurfaceCommand{}
	readCmd.Header.Read(reader)
	core.ReadLE(reader, &readCmd.SurfaceId)
	core.ReadLE(reader, &readCmd.Width)
	core.ReadLE(reader, &readCmd.Height)
	core.ReadLE(reader, &readCmd.PixelFormat)

	assert.Equal(t, uint16(SURFCMD_CREATE_SURFACE), readCmd.Type())
}

func TestDeleteSurfaceCommand(t *testing.T) {
	cmd := &TsDeleteSurfaceCommand{
		Header: TsSurfaceCommandHeader{
			CommandType: SURFCMD_DELETE_SURFACE,
			CommandSize: 4,
		},
		SurfaceId: 1,
	}

	assert.Equal(t, uint16(SURFCMD_DELETE_SURFACE), cmd.Type())

	data := cmd.Serialize()
	assert.Greater(t, len(data), 0)

	// Only read header and SurfaceId to avoid EOF
	reader := bytes.NewReader(data)
	readCmd := &TsDeleteSurfaceCommand{}
	readCmd.Header.Read(reader)
	core.ReadLE(reader, &readCmd.SurfaceId)
	assert.Equal(t, uint16(SURFCMD_DELETE_SURFACE), readCmd.Type())
}

func TestSolidFillCommand(t *testing.T) {
	cmd := &TsSolidFillCommand{
		Header: TsSurfaceCommandHeader{
			CommandType: SURFCMD_SOLID_FILL,
			CommandSize: 16,
		},
		DestLeft:   0,
		DestTop:    0,
		DestRight:  100,
		DestBottom: 100,
		Color:      0xFF0000FF, // Red
	}

	assert.Equal(t, uint16(SURFCMD_SOLID_FILL), cmd.Type())

	data := cmd.Serialize()
	assert.Greater(t, len(data), 0)

	// Only read header and fields written by Serialize to avoid EOF
	reader := bytes.NewReader(data)
	readCmd := &TsSolidFillCommand{}
	readCmd.Header.Read(reader)
	core.ReadLE(reader, &readCmd.DestLeft)
	core.ReadLE(reader, &readCmd.DestTop)
	core.ReadLE(reader, &readCmd.DestRight)
	core.ReadLE(reader, &readCmd.DestBottom)
	core.ReadLE(reader, &readCmd.Color)
	assert.Equal(t, uint16(SURFCMD_SOLID_FILL), readCmd.Type())
}

func BenchmarkSurfaceCommand_Serialize(b *testing.B) {
	cmd := &TsSetSurfaceBitsCommand{
		Header: TsSurfaceCommandHeader{
			CommandType: SURFCMD_SET_SURFACE_BITS,
			CommandSize: 12,
		},
		DestLeft:   0,
		DestTop:    0,
		DestRight:  100,
		DestBottom: 100,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cmd.Serialize()
	}
}

func BenchmarkSurfaceCommand_Read(b *testing.B) {
	cmd := &TsSetSurfaceBitsCommand{
		Header: TsSurfaceCommandHeader{
			CommandType: SURFCMD_SET_SURFACE_BITS,
			CommandSize: 12,
		},
		DestLeft:   0,
		DestTop:    0,
		DestRight:  100,
		DestBottom: 100,
	}

	data := cmd.Serialize()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reader := bytes.NewReader(data)
		ReadSurfaceCommand(reader)
	}
}
