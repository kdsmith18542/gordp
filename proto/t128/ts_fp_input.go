package t128

import (
	"bytes"
	"io"

	"github.com/kdsmith18542/gordp/core"
	"github.com/kdsmith18542/gordp/glog"
)

// TsFpInputPdu
// https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-rdpbcgr/b8e7c588-51cb-455b-bb73-92d480903133
type TsFpInputPdu struct {
	Header          FpInputHeader
	Length          uint16
	FipsInformation uint32           // Optional: when Server Security Data (TS_UD_SC_SEC1) is set
	DataSignature   [8]byte          // Optional: existed if (Header.Flag & FASTPATH_INPUT_SECURE_CHECKSUM)
	NumEvents       uint8            // Optional: if (header.NumEvent != 0)
	FpInputEvents   []TsFpInputEvent // An array of Fast-Path Input Event (section 2.2.8.1.2.2)
}

func (pdu *TsFpInputPdu) iDataPDU() {}

// Read implements the DataPDU interface for reading FastPath input PDUs
func (pdu *TsFpInputPdu) Read(r io.Reader) DataPDU {
	// Read the FastPath input header
	pdu.Header.Read(r)

	// Read the length field (2 bytes, big-endian)
	core.ReadBE(r, &pdu.Length)

	// Check if FIPS information is present (when Server Security Data is set)
	if pdu.Header.Flags&FASTPATH_INPUT_SECURE_CHECKSUM != 0 {
		core.ReadLE(r, &pdu.FipsInformation)
	}

	// Check if data signature is present
	if pdu.Header.Flags&FASTPATH_INPUT_SECURE_CHECKSUM != 0 {
		core.ReadFull(r, pdu.DataSignature[:])
	}

	// Read number of events if present
	if pdu.Header.NumEvents != 0 {
		core.ReadLE(r, &pdu.NumEvents)
	} else {
		pdu.NumEvents = pdu.Header.NumEvents
	}

	// Read input events
	pdu.FpInputEvents = make([]TsFpInputEvent, pdu.NumEvents)
	for i := uint8(0); i < pdu.NumEvents; i++ {
		pdu.FpInputEvents[i] = readFastPathInputEvent(r)
	}

	glog.Debugf("FastPath Input PDU read: %d events, length: %d", pdu.NumEvents, pdu.Length)
	return pdu
}

func (pdu *TsFpInputPdu) Type2() uint8 {
	return PDUTYPE2_INPUT
}

// readFastPathInputEvent reads a single FastPath input event from the reader
func readFastPathInputEvent(r io.Reader) TsFpInputEvent {
	// Read event header (1 byte)
	var eventHeader uint8
	core.ReadLE(r, &eventHeader)

	// Extract event code and flags
	eventCode := (eventHeader >> 5) & 0x07
	eventFlags := eventHeader & 0x1F

	glog.Debugf("Reading FastPath input event: code=%d, flags=%d", eventCode, eventFlags)

	switch eventCode {
	case FASTPATH_INPUT_EVENT_SCANCODE:
		return readFastPathKeyboardEvent(r, eventFlags)
	case FASTPATH_INPUT_EVENT_MOUSE:
		return readFastPathPointerEvent(r)
	case FASTPATH_INPUT_EVENT_MOUSEX:
		return readFastPathPointerEvent(r) // Extended mouse events use same format
	case FASTPATH_INPUT_EVENT_SYNC:
		return readFastPathSyncEvent(r)
	case FASTPATH_INPUT_EVENT_UNICODE:
		return readFastPathUnicodeEvent(r, eventFlags)
	default:
		glog.Errorf("Unknown FastPath input event code: %d", eventCode)
		return nil
	}
}

// readFastPathKeyboardEvent reads a keyboard event
func readFastPathKeyboardEvent(r io.Reader, eventFlags uint8) TsFpInputEvent {
	var keyCode uint8
	core.ReadLE(r, &keyCode)

	down := (eventFlags & 0x01) != 0

	return &TsFpKeyboardEvent{
		EventHeader: eventFlags,
		KeyCode:     keyCode,
	}
}

// readFastPathPointerEvent reads a pointer/mouse event
func readFastPathPointerEvent(r io.Reader) TsFpInputEvent {
	event := &TsFpPointerEvent{}
	core.ReadLE(r, event)
	return event
}

// readFastPathSyncEvent reads a sync event
func readFastPathSyncEvent(r io.Reader) TsFpInputEvent {
	// Sync events are just the event header, no additional data
	return &TsFpSyncEvent{}
}

// readFastPathUnicodeEvent reads a Unicode event
func readFastPathUnicodeEvent(r io.Reader, eventFlags uint8) TsFpInputEvent {
	var unicodeCode uint16
	core.ReadLE(r, &unicodeCode)

	down := (eventFlags & 0x01) != 0

	return &TsFpUnicodeEvent{
		EventHeader: eventFlags,
		UnicodeCode: unicodeCode,
	}
}

// TsFpSyncEvent represents a sync event (no additional data)
type TsFpSyncEvent struct{}

func (e *TsFpSyncEvent) iInputEvent() {}

func (e *TsFpSyncEvent) Serialize() []byte {
	return []byte{FASTPATH_INPUT_EVENT_SYNC << 5}
}

// TsFpUnicodeEvent represents a Unicode input event
type TsFpUnicodeEvent struct {
	EventHeader uint8
	UnicodeCode uint16
}

func (e *TsFpUnicodeEvent) iInputEvent() {}

func (e *TsFpUnicodeEvent) Serialize() []byte {
	b := make([]byte, 3)
	b[0] = (FASTPATH_INPUT_EVENT_UNICODE << 5) | (e.EventHeader & 0x1F)
	core.WriteLE(bytes.NewBuffer(b[1:]), e.UnicodeCode)
	return b
}

// SlowPath Input PDU
// input_send_mouse_event
//  - rdp_client_input_pdu_init
//    - rdp_data_pdu_init
//      - rdp_send_stream_pdu_init
//        - rdp_send_stream_init
//          - RDP_PACKET_HEADER_MAX_LENGTH
//            = TPDU_DATA_LENGTH + MCS_SEND_DATA_HEADER_MAX_LENGTH
//            = (TPKT_HEADER_LENGTH + TPDU_DATA_HEADER_LENGTH) + 8
//            = (4 + 3) + 8
//          - security == can be 0
//        - RDP_SHARE_CONTROL_HEADER_LENGTH = 6
//     - RDP_SHARE_DATA_HEADER_LENGTH = 12
//    - rdp_write_client_input_pdu_header   // TS_INPUT_PDU_DATA <- SlowPath
//      - numberEvents = 2
//      - pad2Octets = 2
//    - rdp_write_input_event_header
//      - eventTime = 4
//      - messageType = 2
//  - input_write_mouse_event
//    - flags = 2
//    - xPos = 2
//    - yPos = 2

// FastPath Input PDU
// input_send_fastpath_mouse_event
//  - fastpath_input_pdu_init
//    - fastpath_input_pdu_init_header
//      - transport_send_stream_init = 0
//      - fpInputHeader, length1 and length2 = 3
//      - fastpath_get_sec_bytes = 0
//    - eventHeader = (eventFlags | (eventCode << 5)) = 1
//  - input_write_mouse_event
//    - flags = 2
//    - xPos = 2
//    - yPos = 2

func (pdu *TsFpInputPdu) Serialize() []byte {
	var events [][]byte
	for _, v := range pdu.FpInputEvents {
		events = append(events, v.Serialize())
	}
	eventsData := bytes.Join(events, nil)
	pdu.Length = uint16(len(eventsData))

	pdu.Header.Action = FASTPATH_INPUT_ACTION_FASTPATH
	pdu.Header.NumEvents = uint8(len(pdu.FpInputEvents))

	buff := new(bytes.Buffer)
	pdu.Header.Write(buff)

	core.WriteBE(buff, (pdu.Length+3)|0x8000) // copy from FreeRDP
	//per.WriteLength(buff, int(pdu.Length))
	buff.Write(eventsData)

	return buff.Bytes()
}

func NewFastPathMouseInputPDU(pointerFlags uint16, xPos, yPos uint16) *TsFpInputPdu {
	return &TsFpInputPdu{
		FpInputEvents: []TsFpInputEvent{&TsFpPointerEvent{
			PointerFlags: pointerFlags,
			XPos:         xPos,
			YPos:         yPos,
		}},
	}
}
