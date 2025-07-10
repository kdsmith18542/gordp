package sec

import (
	"io"

	"github.com/kdsmith18542/gordp/core"
)

// TsAutoReconnectInfo reconnect information
// https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-rdpbcgr/0f9f0375-876b-4c01-8ff9-2c9e5b75b6a8
type TsAutoReconnectInfo struct {
	LogonIdInfo   [16]byte
	ArcRandomBits [16]byte
}

func (i *TsAutoReconnectInfo) Write(w io.Writer) {
	// Write LogonIdInfo (16 bytes)
	core.WriteFull(w, i.LogonIdInfo[:])

	// Write ArcRandomBits (16 bytes)
	core.WriteFull(w, i.ArcRandomBits[:])
}
