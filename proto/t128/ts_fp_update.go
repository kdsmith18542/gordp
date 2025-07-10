package t128

import (
	"bytes"
	"io"

	"github.com/kdsmith18542/gordp/core"
	"github.com/kdsmith18542/gordp/glog"
)

type UpdatePDU interface {
	iUpdatePDU()
	Read(r io.Reader) UpdatePDU
}

// TsFpUpdatePDU
// https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-rdpbcgr/68b5ee54-d0d5-4d65-8d81-e1c4025f7597
type TsFpUpdatePDU struct {
	Header FpOutputHeader
	Length uint16
	PDU    UpdatePDU
}

func (p *TsFpUpdatePDU) iPDU() {}

func (p *TsFpUpdatePDU) Serialize() []byte {
	var buf bytes.Buffer

	// Serialize header manually since it doesn't have a Serialize method
	updateHeader := uint8(p.Header.UpdateCode | (p.Header.Fragmentation << 4) | (p.Header.Compression << 6))
	buf.Write([]byte{updateHeader})

	// Serialize length
	buf.Write(core.ToLE(p.Length))

	// Note: PDU serialization would need to be implemented per type
	// For now, return the basic structure without PDU data

	return buf.Bytes()
}

func (p *TsFpUpdatePDU) Type() uint16 {
	return 0x0 // FASTPATH_OUTPUT_ACTION_FASTPATH equivalent
}

func (p *TsFpUpdatePDU) Read(r io.Reader) PDU {
	p.Header.Read(r)

	core.ReadLE(r, &p.Length)
	if p.Length == 0 {
		glog.Debugf("length = 0")
		return p
	}

	data := core.ReadBytes(r, int(p.Length))
	//glog.Debugf("fastpath pdu data: %v - %x", len(data), data)

	glog.Debugf("updateCode: %v", p.Header.UpdateCode)
	switch p.Header.UpdateCode {
	case FASTPATH_UPDATETYPE_BITMAP:
		p.PDU = (&TsFpUpdateBitmap{}).Read(bytes.NewReader(data))
	case FASTPATH_UPDATETYPE_CACHED:
		p.PDU = (&TsFpUpdateCachedBitmap{}).Read(bytes.NewReader(data))
	case FASTPATH_UPDATETYPE_SURFCMDS:
		p.PDU = (&TsFpUpdateSurfaceCommands{}).Read(bytes.NewReader(data))
	default:
		glog.Warnf("updateCode [%x] not implement", p.Header.UpdateCode)
	}

	glog.Debugf("p.PDU: %T", p.PDU)
	return p
}
