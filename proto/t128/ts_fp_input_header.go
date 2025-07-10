package t128

import (
	"io"

	"github.com/kdsmith18542/gordp/core"
)

// action
const (
	FASTPATH_INPUT_ACTION_FASTPATH = 0x0
	FASTPATH_INPUT_ACTION_X224     = 0x3
)

// flags
const (
	FASTPATH_INPUT_SECURE_CHECKSUM = 0x1
	FASTPATH_INPUT_ENCRYPTED       = 0x2
)

// FpInputHeader
// https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-rdpbcgr/b8e7c588-51cb-455b-bb73-92d480903133
type FpInputHeader struct {
	Action    uint8
	NumEvents uint8
	Flags     uint8
}

func (h *FpInputHeader) Read(r io.Reader) {
	var inputHeader uint8
	core.ReadLE(r, &inputHeader)

	// Extract fields from the packed header
	h.Action = (inputHeader >> 6) & 0x03
	h.NumEvents = (inputHeader >> 2) & 0x0F
	h.Flags = inputHeader & 0x03
}

func (h *FpInputHeader) Write(w io.Writer) {
	inputHeader := uint8(h.Action<<6 | h.NumEvents<<2 | h.Flags)
	core.WriteLE(w, inputHeader)
}
