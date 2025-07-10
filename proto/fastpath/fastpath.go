package fastpath

import (
	"io"

	"github.com/kdsmith18542/gordp/core"
	"github.com/kdsmith18542/gordp/proto/mcs/per"
)

type Header struct {
	EncryptionFlags uint8
	NumberEvents    uint8
	Length          int
}

func (h *Header) Read(r io.Reader) {
	var b uint8
	core.ReadLE(r, &b)
	h.EncryptionFlags = (b & 0xc0) >> 6
	h.NumberEvents = (b & 0x3c) >> 2
	h.Length = per.ReadLength(r)
	h.Length = core.If(h.Length < 0x80, h.Length-2, h.Length-3)
}

func (h *Header) Write(w io.Writer) {
	b := uint8(h.EncryptionFlags<<6 | h.NumberEvents<<2)
	core.WriteLE(w, b)
	h.Length = core.If(h.Length < 0x80, h.Length+2, h.Length+3)
	per.WriteLength(w, h.Length)
}

type FastPathData struct {
	Header Header
	Data   []byte
}

func Read(r io.Reader) *FastPathData {
	fp := &FastPathData{}
	fp.Header.Read(r)
	//glog.Debugf("fastpath read header: %+v", fp.Header)
	fp.Data = core.ReadBytes(r, fp.Header.Length)
	//glog.Debugf("fastpath read data: %v - %x", len(fp.Data), fp.Data)
	return fp
}

func Write(w io.Writer, data []byte) {
	(&Header{Length: len(data)}).Write(w)
	core.WriteFull(w, data)
}
