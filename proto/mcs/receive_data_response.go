package mcs

import (
	"bytes"
	"fmt"
	"github.com/kdsmith18542/gordp/core"
	"github.com/kdsmith18542/gordp/glog"
	"github.com/kdsmith18542/gordp/proto/mcs/per"
	"github.com/kdsmith18542/gordp/proto/x224"
	"io"
)

type ReceiveDataResponse struct{}

func (res *ReceiveDataResponse) Read(r io.Reader) (uint16, []byte) {
	data := x224.Read(r)
	r = bytes.NewReader(data)
	pduHeader := ReadMcsPduHeader(r)
	core.ThrowIf(pduHeader != MCS_PDUTYPE_SEND_DATA_INDICATION, fmt.Errorf("invalid pdu header: %v", pduHeader))
	userId := per.ReadInteger16(r, MCS_CHANNEL_USERID_BASE) // UserId
	channelId := per.ReadInteger16(r, 0)
	glog.Debugf("userId: %v, channelId: %v", userId, channelId)
	enumerated := per.ReadEnumerated(r)
	glog.Debugf("enumerated: %v", enumerated)
	return channelId, per.ReadOctetString(r, 0)
}
