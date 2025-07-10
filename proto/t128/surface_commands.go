package t128

import (
	"bytes"
	"io"

	"github.com/kdsmith18542/gordp/core"
	"github.com/kdsmith18542/gordp/glog"
)

// Surface Command Types
const (
	SURFCMD_SET_SURFACE_BITS             = 0x0001
	SURFCMD_FRAME_MARKER                 = 0x0004
	SURFCMD_STREAM_SURFACE_BITS          = 0x0006
	SURFCMD_CREATE_SURFACE               = 0x000C
	SURFCMD_DELETE_SURFACE               = 0x000D
	SURFCMD_SOLID_FILL                   = 0x000E
	SURFCMD_SURFACE_TO_SURFACE           = 0x000F
	SURFCMD_SURFACE_TO_CACHE             = 0x0010
	SURFCMD_CACHE_TO_SURFACE             = 0x0011
	SURFCMD_EVICT_CACHE_ENTRY            = 0x0012
	SURFCMD_CACHE_IMPORT_OFFER           = 0x0013
	SURFCMD_CACHE_IMPORT_REPLY           = 0x0014
	SURFCMD_CAPS_ADVERTISE               = 0x0015
	SURFCMD_CAPS_CONFIRM                 = 0x0016
	SURFCMD_MAP_SURFACE_TO_OUTPUT        = 0x0017
	SURFCMD_MAP_SURFACE_TO_WINDOW        = 0x0018
	SURFCMD_MAP_SURFACE_TO_SCALED_OUTPUT = 0x0019
	SURFCMD_MAP_SURFACE_TO_SCALED_WINDOW = 0x001A
	SURFCMD_OVERLAY_SURFACE              = 0x001B
	SURFCMD_REMOVE_SURFACE               = 0x001C
	SURFCMD_MAP_SURFACE_TO_OUTPUT_SCALED = 0x001D
)

// Surface Command Flags
const (
	SURFCMD_FLAG_SET_SURFACE_BITS_ETWING    = 0x0001
	SURFCMD_FLAG_SET_SURFACE_BITS_V2        = 0x0002
	SURFCMD_FLAG_STREAM_SURFACE_BITS_ETWING = 0x0001
	SURFCMD_FLAG_STREAM_SURFACE_BITS_V2     = 0x0002
	SURFCMD_FLAG_FRAME_MARKER_ETWING        = 0x0001
	SURFCMD_FLAG_FRAME_MARKER_V2            = 0x0002
)

// Surface Command Header
// https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-rdpegdi/2c3c3c41-1d54-4254-bb62-bc082a3c1f10
type TsSurfaceCommandHeader struct {
	CommandType uint16
	CommandSize uint16
}

func (h *TsSurfaceCommandHeader) Read(r io.Reader) {
	core.ReadLE(r, &h.CommandType)
	core.ReadLE(r, &h.CommandSize)
}

func (h *TsSurfaceCommandHeader) Write(w io.Writer) {
	core.WriteLE(w, h.CommandType)
	core.WriteLE(w, h.CommandSize)
}

// Surface Command Interface
type SurfaceCommand interface {
	Type() uint16
	Read(r io.Reader) SurfaceCommand
	Write(w io.Writer)
	Serialize() []byte
}

// Set Surface Bits Command
// https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-rdpegdi/2c3c3c41-1d54-4254-bb62-bc082a3c1f10
type TsSetSurfaceBitsCommand struct {
	Header     TsSurfaceCommandHeader
	DestLeft   uint16
	DestTop    uint16
	DestRight  uint16
	DestBottom uint16
	BitmapData TsBitmapData
}

func (c *TsSetSurfaceBitsCommand) Type() uint16 {
	return SURFCMD_SET_SURFACE_BITS
}

func (c *TsSetSurfaceBitsCommand) Read(r io.Reader) SurfaceCommand {
	c.Header.Read(r)
	core.ReadLE(r, &c.DestLeft)
	core.ReadLE(r, &c.DestTop)
	core.ReadLE(r, &c.DestRight)
	core.ReadLE(r, &c.DestBottom)
	c.BitmapData.Read(r)
	return c
}

func (c *TsSetSurfaceBitsCommand) Write(w io.Writer) {
	c.Header.Write(w)
	core.WriteLE(w, c.DestLeft)
	core.WriteLE(w, c.DestTop)
	core.WriteLE(w, c.DestRight)
	core.WriteLE(w, c.DestBottom)
	// BitmapData is written separately
}

func (c *TsSetSurfaceBitsCommand) Serialize() []byte {
	buff := new(bytes.Buffer)
	c.Write(buff)
	// Add bitmap data
	buff.Write(c.BitmapData.BitmapDataStream)
	return buff.Bytes()
}

// Frame Marker Command
// https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-rdpegdi/2c3c3c41-1d54-4254-bb62-bc082a3c1f10
type TsFrameMarkerCommand struct {
	Header      TsSurfaceCommandHeader
	FrameAction uint16
	FrameId     uint32
}

func (c *TsFrameMarkerCommand) Type() uint16 {
	return SURFCMD_FRAME_MARKER
}

func (c *TsFrameMarkerCommand) Read(r io.Reader) SurfaceCommand {
	c.Header.Read(r)
	core.ReadLE(r, &c.FrameAction)
	core.ReadLE(r, &c.FrameId)
	return c
}

func (c *TsFrameMarkerCommand) Write(w io.Writer) {
	c.Header.Write(w)
	core.WriteLE(w, c.FrameAction)
	core.WriteLE(w, c.FrameId)
}

func (c *TsFrameMarkerCommand) Serialize() []byte {
	buff := new(bytes.Buffer)
	c.Write(buff)
	return buff.Bytes()
}

// Create Surface Command
// https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-rdpegdi/2c3c3c41-1d54-4254-bb62-bc082a3c1f10
type TsCreateSurfaceCommand struct {
	Header         TsSurfaceCommandHeader
	SurfaceId      uint16
	Width          uint16
	Height         uint16
	PixelFormat    uint8
	ScanLineStride uint16
	HeapIndex      uint8
	SurfaceData    []byte
}

func (c *TsCreateSurfaceCommand) Type() uint16 {
	return SURFCMD_CREATE_SURFACE
}

func (c *TsCreateSurfaceCommand) Read(r io.Reader) SurfaceCommand {
	c.Header.Read(r)
	core.ReadLE(r, &c.SurfaceId)
	core.ReadLE(r, &c.Width)
	core.ReadLE(r, &c.Height)
	core.ReadLE(r, &c.PixelFormat)
	core.ReadLE(r, &c.ScanLineStride)
	core.ReadLE(r, &c.HeapIndex)

	// Read surface data if present
	if c.Header.CommandSize > 12 {
		dataSize := int(c.Header.CommandSize) - 12
		c.SurfaceData = core.ReadBytes(r, dataSize)
	}

	return c
}

func (c *TsCreateSurfaceCommand) Write(w io.Writer) {
	c.Header.Write(w)
	core.WriteLE(w, c.SurfaceId)
	core.WriteLE(w, c.Width)
	core.WriteLE(w, c.Height)
	core.WriteLE(w, c.PixelFormat)
	core.WriteLE(w, c.ScanLineStride)
	core.WriteLE(w, c.HeapIndex)
	if len(c.SurfaceData) > 0 {
		core.WriteFull(w, c.SurfaceData)
	}
}

func (c *TsCreateSurfaceCommand) Serialize() []byte {
	buff := new(bytes.Buffer)
	c.Write(buff)
	return buff.Bytes()
}

// Delete Surface Command
type TsDeleteSurfaceCommand struct {
	Header    TsSurfaceCommandHeader
	SurfaceId uint16
}

func (c *TsDeleteSurfaceCommand) Type() uint16 {
	return SURFCMD_DELETE_SURFACE
}

func (c *TsDeleteSurfaceCommand) Read(r io.Reader) SurfaceCommand {
	c.Header.Read(r)
	core.ReadLE(r, &c.SurfaceId)
	return c
}

func (c *TsDeleteSurfaceCommand) Write(w io.Writer) {
	c.Header.Write(w)
	core.WriteLE(w, c.SurfaceId)
}

func (c *TsDeleteSurfaceCommand) Serialize() []byte {
	buff := new(bytes.Buffer)
	c.Write(buff)
	return buff.Bytes()
}

// Solid Fill Command
type TsSolidFillCommand struct {
	Header     TsSurfaceCommandHeader
	DestLeft   uint16
	DestTop    uint16
	DestRight  uint16
	DestBottom uint16
	Color      uint32
}

func (c *TsSolidFillCommand) Type() uint16 {
	return SURFCMD_SOLID_FILL
}

func (c *TsSolidFillCommand) Read(r io.Reader) SurfaceCommand {
	c.Header.Read(r)
	core.ReadLE(r, &c.DestLeft)
	core.ReadLE(r, &c.DestTop)
	core.ReadLE(r, &c.DestRight)
	core.ReadLE(r, &c.DestBottom)
	core.ReadLE(r, &c.Color)
	return c
}

func (c *TsSolidFillCommand) Write(w io.Writer) {
	c.Header.Write(w)
	core.WriteLE(w, c.DestLeft)
	core.WriteLE(w, c.DestTop)
	core.WriteLE(w, c.DestRight)
	core.WriteLE(w, c.DestBottom)
	core.WriteLE(w, c.Color)
}

func (c *TsSolidFillCommand) Serialize() []byte {
	buff := new(bytes.Buffer)
	c.Write(buff)
	return buff.Bytes()
}

// Surface to Surface Copy Command
type TsSurfaceToSurfaceCommand struct {
	Header          TsSurfaceCommandHeader
	SourceSurfaceId uint16
	DestSurfaceId   uint16
	SourceRect      Rectangle
	DestRect        Rectangle
}

type Rectangle struct {
	Left   uint16
	Top    uint16
	Right  uint16
	Bottom uint16
}

func (c *TsSurfaceToSurfaceCommand) Type() uint16 {
	return SURFCMD_SURFACE_TO_SURFACE
}

func (c *TsSurfaceToSurfaceCommand) Read(r io.Reader) SurfaceCommand {
	c.Header.Read(r)
	core.ReadLE(r, &c.SourceSurfaceId)
	core.ReadLE(r, &c.DestSurfaceId)
	core.ReadLE(r, &c.SourceRect)
	core.ReadLE(r, &c.DestRect)
	return c
}

func (c *TsSurfaceToSurfaceCommand) Write(w io.Writer) {
	c.Header.Write(w)
	core.WriteLE(w, c.SourceSurfaceId)
	core.WriteLE(w, c.DestSurfaceId)
	core.WriteLE(w, c.SourceRect)
	core.WriteLE(w, c.DestRect)
}

func (c *TsSurfaceToSurfaceCommand) Serialize() []byte {
	buff := new(bytes.Buffer)
	c.Write(buff)
	return buff.Bytes()
}

// Surface to Cache Command
type TsSurfaceToCacheCommand struct {
	Header     TsSurfaceCommandHeader
	SurfaceId  uint16
	CacheKey   [8]byte
	CacheSlot  uint16
	SourceRect Rectangle
}

func (c *TsSurfaceToCacheCommand) Type() uint16 {
	return SURFCMD_SURFACE_TO_CACHE
}

func (c *TsSurfaceToCacheCommand) Read(r io.Reader) SurfaceCommand {
	c.Header.Read(r)
	core.ReadLE(r, &c.SurfaceId)
	core.ReadFull(r, c.CacheKey[:])
	core.ReadLE(r, &c.CacheSlot)
	core.ReadLE(r, &c.SourceRect)
	return c
}

func (c *TsSurfaceToCacheCommand) Write(w io.Writer) {
	c.Header.Write(w)
	core.WriteLE(w, c.SurfaceId)
	core.WriteFull(w, c.CacheKey[:])
	core.WriteLE(w, c.CacheSlot)
	core.WriteLE(w, c.SourceRect)
}

func (c *TsSurfaceToCacheCommand) Serialize() []byte {
	buff := new(bytes.Buffer)
	c.Write(buff)
	return buff.Bytes()
}

// Cache to Surface Command
type TsCacheToSurfaceCommand struct {
	Header    TsSurfaceCommandHeader
	CacheSlot uint16
	SurfaceId uint16
	DestRect  Rectangle
}

func (c *TsCacheToSurfaceCommand) Type() uint16 {
	return SURFCMD_CACHE_TO_SURFACE
}

func (c *TsCacheToSurfaceCommand) Read(r io.Reader) SurfaceCommand {
	c.Header.Read(r)
	core.ReadLE(r, &c.CacheSlot)
	core.ReadLE(r, &c.SurfaceId)
	core.ReadLE(r, &c.DestRect)
	return c
}

func (c *TsCacheToSurfaceCommand) Write(w io.Writer) {
	c.Header.Write(w)
	core.WriteLE(w, c.CacheSlot)
	core.WriteLE(w, c.SurfaceId)
	core.WriteLE(w, c.DestRect)
}

func (c *TsCacheToSurfaceCommand) Serialize() []byte {
	buff := new(bytes.Buffer)
	c.Write(buff)
	return buff.Bytes()
}

// Surface Command Map
var surfaceCommandMap = map[uint16]SurfaceCommand{
	SURFCMD_SET_SURFACE_BITS:   &TsSetSurfaceBitsCommand{},
	SURFCMD_FRAME_MARKER:       &TsFrameMarkerCommand{},
	SURFCMD_CREATE_SURFACE:     &TsCreateSurfaceCommand{},
	SURFCMD_DELETE_SURFACE:     &TsDeleteSurfaceCommand{},
	SURFCMD_SOLID_FILL:         &TsSolidFillCommand{},
	SURFCMD_SURFACE_TO_SURFACE: &TsSurfaceToSurfaceCommand{},
	SURFCMD_SURFACE_TO_CACHE:   &TsSurfaceToCacheCommand{},
	SURFCMD_CACHE_TO_SURFACE:   &TsCacheToSurfaceCommand{},
}

// Read Surface Command
func ReadSurfaceCommand(r io.Reader) SurfaceCommand {
	header := &TsSurfaceCommandHeader{}
	header.Read(r)

	command, exists := surfaceCommandMap[header.CommandType]
	if !exists {
		glog.Warnf("Unknown surface command type: 0x%04X", header.CommandType)
		return nil
	}

	return command.Read(r)
}

// FastPath Surface Commands Update
type TsFpUpdateSurfaceCommands struct {
	UpdateType     int16 // This field MUST be set to FASTPATH_UPDATETYPE_SURFCMDS (0x0004).
	NumberCommands uint16
	Commands       []SurfaceCommand
}

func (t *TsFpUpdateSurfaceCommands) iUpdatePDU() {}

func (t *TsFpUpdateSurfaceCommands) Read(r io.Reader) UpdatePDU {
	var updateType int16
	core.ReadLE(r, &updateType)
	core.ReadLE(r, &t.NumberCommands)

	glog.Debugf("Surface commands update: %d commands", t.NumberCommands)

	t.Commands = make([]SurfaceCommand, t.NumberCommands)
	for i := range t.Commands {
		command := ReadSurfaceCommand(r)
		if command != nil {
			t.Commands[i] = command
			glog.Debugf("Surface command %d: type=0x%04X", i, command.Type())
		}
	}

	return t
}
