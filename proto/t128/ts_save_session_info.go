package t128

import (
	"io"

	"github.com/kdsmith18542/gordp/core"
	"github.com/kdsmith18542/gordp/glog"
)

// TsSaveSessionInfoPDU
// https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-rdpbcgr/d892bc5b-aecd-4aee-99b6-5f43b5a63d75
type TsSaveSessionInfoPDU struct {
	InfoType uint32
	InfoData []byte
}

func (t *TsSaveSessionInfoPDU) iDataPDU() {}

func (t *TsSaveSessionInfoPDU) Read(r io.Reader) DataPDU {
	glog.Warnf("not implement")
	return t
}

func (t *TsSaveSessionInfoPDU) Serialize() []byte {
	// Serialize InfoType as little-endian 32-bit integer
	data := core.ToLE(t.InfoType)

	// Append InfoData if present
	if len(t.InfoData) > 0 {
		data = append(data, t.InfoData...)
	}

	return data
}

func (t *TsSaveSessionInfoPDU) Type2() uint8 {
	return PDUTYPE2_SAVE_SESSION_INFO
}
