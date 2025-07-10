package t128

import (
	"bytes"

	"github.com/kdsmith18542/gordp/core"
	"github.com/kdsmith18542/gordp/glog"
)

const (
	// Mouse wheel event:
	PTRFLAGS_HWHEEL         = 0x0400
	PTRFLAGS_WHEEL          = 0x0200
	PTRFLAGS_WHEEL_NEGATIVE = 0x0100
	WheelRotationMask       = 0x01FF

	// Mouse movement event:
	PTRFLAGS_MOVE = 0x0800

	// Mouse button events:
	PTRFLAGS_DOWN    = 0x8000
	PTRFLAGS_BUTTON1 = 0x1000 // Left button
	PTRFLAGS_BUTTON2 = 0x2000 // Right button
	PTRFLAGS_BUTTON3 = 0x4000 // Middle button
	PTRFLAGS_BUTTON4 = 0x0100 // X1 button (back)
	PTRFLAGS_BUTTON5 = 0x0200 // X2 button (forward)
)

// MouseButton represents different mouse buttons
type MouseButton int

const (
	MouseButtonLeft MouseButton = iota
	MouseButtonRight
	MouseButtonMiddle
	MouseButtonX1
	MouseButtonX2
)

// ScrollDirection represents scroll directions
type ScrollDirection int

const (
	ScrollUp ScrollDirection = iota
	ScrollDown
	ScrollLeft
	ScrollRight
)

// TsFpPointerEvent
// https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-rdpbcgr/16a96ded-b3d3-4468-b993-9c7a51297510
// https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-rdpbcgr/2c1ced34-340a-46cd-be6e-fc8cab7c3b17
type TsFpPointerEvent struct {
	PointerFlags uint16
	XPos, YPos   uint16
}

func (e *TsFpPointerEvent) iInputEvent() {}

func (e *TsFpPointerEvent) Serialize() []byte {
	buff := new(bytes.Buffer)
	core.WriteLE(buff, uint8(FASTPATH_INPUT_EVENT_MOUSE<<5)) // eventHeader: eventFlags=0, eventCode=FASTPATH_INPUT_EVENT_MOUSE
	core.WriteLE(buff, e)

	glog.Debugf("mouse event: %v - %x", buff.Len(), buff.Bytes())
	return buff.Bytes()
}

// NewFastPathPointerEvent creates a new pointer event
func NewFastPathPointerEvent(pointerFlags uint16, xPos, yPos uint16) *TsFpPointerEvent {
	return &TsFpPointerEvent{
		PointerFlags: pointerFlags,
		XPos:         xPos,
		YPos:         yPos,
	}
}

// NewFastPathMouseMoveEvent creates a mouse movement event
func NewFastPathMouseMoveEvent(xPos, yPos uint16) *TsFpPointerEvent {
	return NewFastPathPointerEvent(PTRFLAGS_MOVE, xPos, yPos)
}

// NewFastPathMouseButtonEvent creates a mouse button event
func NewFastPathMouseButtonEvent(button MouseButton, down bool, xPos, yPos uint16) *TsFpPointerEvent {
	var flags uint16
	switch button {
	case MouseButtonLeft:
		flags = PTRFLAGS_BUTTON1
	case MouseButtonRight:
		flags = PTRFLAGS_BUTTON2
	case MouseButtonMiddle:
		flags = PTRFLAGS_BUTTON3
	case MouseButtonX1:
		flags = PTRFLAGS_BUTTON4
	case MouseButtonX2:
		flags = PTRFLAGS_BUTTON5
	}

	if down {
		flags |= PTRFLAGS_DOWN
	}

	return NewFastPathPointerEvent(flags, xPos, yPos)
}

// NewFastPathMouseWheelEvent creates a mouse wheel event
func NewFastPathMouseWheelEvent(wheelDelta int16, xPos, yPos uint16) *TsFpPointerEvent {
	flags := uint16(PTRFLAGS_WHEEL)
	if wheelDelta < 0 {
		flags |= PTRFLAGS_WHEEL_NEGATIVE
	}

	// Convert wheel delta to rotation value (0-255)
	rotation := uint16(wheelDelta) & WheelRotationMask

	return NewFastPathPointerEvent(flags|rotation, xPos, yPos)
}

// NewFastPathMouseHorizontalWheelEvent creates a horizontal mouse wheel event
func NewFastPathMouseHorizontalWheelEvent(wheelDelta int16, xPos, yPos uint16) *TsFpPointerEvent {
	flags := uint16(PTRFLAGS_HWHEEL)
	if wheelDelta < 0 {
		flags |= PTRFLAGS_WHEEL_NEGATIVE
	}

	// Convert wheel delta to rotation value (0-255)
	rotation := uint16(wheelDelta) & WheelRotationMask

	return NewFastPathPointerEvent(flags|rotation, xPos, yPos)
}
