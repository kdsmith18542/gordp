package t128

import (
	"io"

	"github.com/kdsmith18542/gordp/core"
)

// TsFontListPDU
// https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-rdpbcgr/e373575a-01e2-43a7-a6d8-e1952b83e787
type TsFontListPDU struct {
	NumberFonts   uint16
	TotalNumFonts uint16
	ListFlags     uint16 //This field SHOULD be set to 0x0003
	EntrySize     uint16 //This field SHOULD be set to 0x0032 (50 bytes).
}

func (t *TsFontListPDU) Read(r io.Reader) DataPDU {
	// Read all fields as little-endian
	t.NumberFonts = core.ReadLE(r, t.NumberFonts)
	t.TotalNumFonts = core.ReadLE(r, t.TotalNumFonts)
	t.ListFlags = core.ReadLE(r, t.ListFlags)
	t.EntrySize = core.ReadLE(r, t.EntrySize)
	return t
}

func (t *TsFontListPDU) iDataPDU() {}

func (t *TsFontListPDU) Serialize() []byte {
	return core.ToLE(t)
}

func (t *TsFontListPDU) Type2() uint8 {
	return PDUTYPE2_FONTLIST
}
