package nla

import (
	"bytes"

	"github.com/GoFeGroup/gordp/core"
)

// NTLMv2ClientChallenge
// https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-nlmp/aee311d6-21a7-4470-92a5-c4ecb022a87b
type NTLMv2ClientChallenge struct {
	Must struct {
		RespType            uint8
		HiRespType          uint8
		Reserved1           uint16
		Reserved2           uint32
		Timestamp           [8]byte
		ChallengeFromClient [8]byte
		Reserved3           uint32
	}
	Optional struct {
		AvPairs AVPairs
	}
}

func NewNTLMv2ClientChallenge(serverInfo, timestamp []byte) *NTLMv2ClientChallenge {
	clientChallenge := &NTLMv2ClientChallenge{}
	clientChallenge.Must.RespType = 0x01
	clientChallenge.Must.HiRespType = 0x01
	copy(clientChallenge.Must.Timestamp[:], timestamp[:8])
	copy(clientChallenge.Must.ChallengeFromClient[:], core.Random(8))
	clientChallenge.Optional.AvPairs = ReadAvPairs(serverInfo)
	return clientChallenge
}

// AddChannelBinding adds a channel binding token to the client challenge
func (c *NTLMv2ClientChallenge) AddChannelBinding(channelBindingToken []byte) {
	if len(channelBindingToken) > 0 {
		// Create channel binding AVPair
		channelBindingPair := CreateChannelBindingAVPair(channelBindingToken)

		// Insert channel binding before the EOL pair
		var newAvPairs AVPairs
		for _, pair := range c.Optional.AvPairs {
			if pair.Must.Id == MsvAvEOL {
				// Insert channel binding before EOL
				newAvPairs = append(newAvPairs, channelBindingPair)
			}
			newAvPairs = append(newAvPairs, pair)
		}
		c.Optional.AvPairs = newAvPairs
	}
}

func (c *NTLMv2ClientChallenge) Serialize() []byte {
	buff := new(bytes.Buffer)
	core.WriteLE(buff, c.Must)
	c.Optional.AvPairs.Write(buff)
	buff.Write([]byte{0x00, 0x00, 0x00, 0x00}) // FIXME: four bytes?
	return buff.Bytes()
}
