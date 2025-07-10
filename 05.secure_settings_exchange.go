package gordp

import (
	"github.com/kdsmith18542/gordp/proto/mcs"
	"github.com/kdsmith18542/gordp/proto/pdu/licPdu"
)

func (c *Client) sendClientInfo() {
	clientInfo := licPdu.NewClientInfoPDU(c.userId, c.option.UserName, c.option.Password)
	clientInfo.Write(c.stream)

	// Send Monitor Layout PDU if multi-monitor is configured
	if len(c.monitors) > 0 {
		pdu := &mcs.MonitorLayoutPDU{
			UserDataHeader: mcs.UserDataHeader{
				Type: mcs.CS_MONITOR,
				Len:  uint16(8 + len(c.monitors)*40), // 8 bytes header + 40 bytes per monitor
			},
			NumMonitors: uint32(len(c.monitors)),
			Monitors:    c.monitors,
		}
		c.stream.Write(pdu.Serialize())
	}
}
