package mcsPdu

import (
	"bytes"
	"github.com/kdsmith18542/gordp/glog"
	"github.com/kdsmith18542/gordp/proto/mcs"
	"github.com/kdsmith18542/gordp/proto/x224"
	"io"
)

// ServerMcsChannelJoinConfirmPDU
// https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-rdpbcgr/cfc938b5-041d-4c15-9909-81dd035b914e
type ServerMcsChannelJoinConfirmPDU struct {
	McsCJcf mcs.ServerChannelJoinConfirm
}

func (pdu *ServerMcsChannelJoinConfirmPDU) Read(r io.Reader) {
	data := x224.Read(r)
	glog.Debugf("read channel join confirm: %v - %x", len(data), data)
	pdu.McsCJcf.Read(bytes.NewReader(data))
	glog.Debugf("mcsCJcf: %+v", pdu.McsCJcf)
}
