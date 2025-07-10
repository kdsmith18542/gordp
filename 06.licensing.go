package gordp

import "github.com/kdsmith18542/gordp/proto/pdu/licPdu"

func (c *Client) readLicensing() {
	licensing := licPdu.ServerLicensingPDU{}
	licensing.Read(c.stream)
}
