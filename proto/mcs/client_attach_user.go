package mcs

import (
	"bytes"
	"io"
)

type ClientAttachUser struct{}

func (c *ClientAttachUser) Write(w io.Writer) {
	WriteMcsPduHeader(w, MCS_PDUTYPE_ATTACH_USER_REQUEST, 0)
}

func (c *ClientAttachUser) Serialize() []byte {
	// Pre-allocate buffer: small fixed-size message (typically < 8 bytes)
	buff := bytes.NewBuffer(make([]byte, 0, 8))
	c.Write(buff)
	return buff.Bytes()
}
