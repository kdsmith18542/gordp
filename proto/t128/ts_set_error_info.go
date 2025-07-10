package t128

import (
	"io"

	"github.com/kdsmith18542/gordp/core"
)

// TsSetErrorInfoPDU
// https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-rdpbcgr/a21a1bd9-2303-49c1-90ec-3932435c248c
type TsSetErrorInfoPDU struct {
	ErrorInfo uint32
}

func (t *TsSetErrorInfoPDU) iDataPDU() {}

func (t *TsSetErrorInfoPDU) Read(r io.Reader) DataPDU {
	return core.ReadLE(r, t)
}

func (t *TsSetErrorInfoPDU) Serialize() []byte {
	// Serialize ErrorInfo as little-endian 32-bit integer
	return core.ToLE(t.ErrorInfo)
}

func (t *TsSetErrorInfoPDU) Type2() uint8 {
	return PDUTYPE2_SET_ERROR_INFO_PDU
}
