package mcs

import (
	"github.com/kdsmith18542/gordp/core"
	"github.com/kdsmith18542/gordp/glog"
	"io"
)

// ServerNetworkData
// https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-rdpbcgr/89fa11de-5275-4106-9cf1-e5aa7709436c
type ServerNetworkData struct {
	McsChannelId   uint16
	ChannelCount   uint16
	ChannelIdArray []uint16
}

func (d *ServerNetworkData) Read(r io.Reader) {
	core.ReadLE(r, &d.McsChannelId)
	core.ReadLE(r, &d.ChannelCount)
	d.ChannelIdArray = make([]uint16, d.ChannelCount)
	core.ReadLE(r, d.ChannelIdArray)
	glog.Debugf("server network data: %+v", d)
}
