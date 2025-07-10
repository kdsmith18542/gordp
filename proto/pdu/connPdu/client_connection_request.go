package connPdu

import (
	"bytes"
	"io"

	"github.com/kdsmith18542/gordp/core"
	"github.com/kdsmith18542/gordp/proto/x224"
)

// ClientConnectionRequestPDU
// https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-rdpbcgr/18a27ef9-6f9a-4501-b000-94b1fe3c2c10
type ClientConnectionRequestPDU struct {
	Cookie      string
	ProtocolNeg Negotiation
}

func NewClientConnectionRequestPDU() *ClientConnectionRequestPDU {
	return &ClientConnectionRequestPDU{
		Cookie: "Cookie: mstshash=DESKTOP-0",
		ProtocolNeg: Negotiation{
			Type:   TYPE_RDP_NEG_REQ,
			Length: 8,
			Result: PROTOCOL_RDP | PROTOCOL_SSL | PROTOCOL_HYBRID,
		}}
}

func (pdu *ClientConnectionRequestPDU) Serialize() []byte {
	buff := new(bytes.Buffer)
	core.WriteFull(buff, []byte(pdu.Cookie+"\r\n"))
	core.WriteLE(buff, &pdu.ProtocolNeg)
	return buff.Bytes()
}

func (pdu *ClientConnectionRequestPDU) Write(w io.Writer) {
	x224.Connect(w, x224.TPDU_CONNECTION_REQUEST, pdu.Serialize())
}
