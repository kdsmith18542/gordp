package t128

import (
	"io"

	"github.com/kdsmith18542/gordp/core"
)

// TsFontMapPDU
// https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-rdpbcgr/b4e557f3-7540-46fc-815d-0c12299cf1ee
type TsFontMapPDU struct {
	NumberEntries   uint16
	TotalNumEntries uint16
	MapFlags        uint16
	EntrySize       uint16
}

func (t *TsFontMapPDU) iDataPDU() {}

func (t *TsFontMapPDU) Read(r io.Reader) DataPDU {
	return core.ReadLE(r, t)
}

func (t *TsFontMapPDU) Serialize() []byte {
	// Serialize all fields as little-endian
	return core.ToLE(t)
}

func (t *TsFontMapPDU) Type2() uint8 {
	return PDUTYPE2_FONTMAP
}
