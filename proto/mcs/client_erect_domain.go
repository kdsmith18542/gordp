package mcs

import (
	"bytes"
	"github.com/kdsmith18542/gordp/proto/mcs/per"
	"io"
)

type ClientErectDomain struct{}

func (e *ClientErectDomain) Write(w io.Writer) {
	WriteMcsPduHeader(w, MCS_PDUTYPE_ERECT_DOMAIN_REQUEST, 0)
	per.WriteInteger(w, 0) // subHeight
	per.WriteInteger(w, 0)
}

func (e *ClientErectDomain) Serialize() []byte {
	// Pre-allocate buffer: small fixed-size message (typically < 16 bytes)
	buff := bytes.NewBuffer(make([]byte, 0, 16))
	e.Write(buff)
	return buff.Bytes()
}
