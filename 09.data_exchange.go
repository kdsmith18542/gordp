package gordp

import (
	"fmt"
	"time"

	"github.com/GoFeGroup/gordp/core"
	"github.com/GoFeGroup/gordp/glog"
	"github.com/GoFeGroup/gordp/proto/t128"
)

func (c *Client) readPdu() t128.PDU {
	glog.Debugf("before peek")
	defer func() { glog.Debugf("exit readPDU") }()
	d := c.stream.Peek(1)
	switch d[0] {
	case 3:
		glog.Debugf("read tpkt pdu begin")
		return t128.ReadPDU(c.stream)
	case 0:
		glog.Debugf("read fastpath pdu begin")
		return t128.ReadFastPathPDU(c.stream)
	default:
		core.Throw("invalid package")
	}
	return nil
}

func (c *Client) sendMouseEvent(pointerFlags uint16, xPos, yPos uint16) error {
	pdu := t128.NewFastPathMouseInputPDU(pointerFlags, xPos, yPos)
	data := pdu.Serialize()
	glog.Debugf("send mouse event data: %v - %x:", len(data), data)
	_, err := c.stream.Write(data)
	return err
}

// SendMouseMoveEvent sends a mouse movement event
func (c *Client) SendMouseMoveEvent(xPos, yPos uint16) error {
	return c.sendMouseEvent(t128.PTRFLAGS_MOVE, xPos, yPos)
}

// SendMouseMoveRelative sends a relative mouse movement event
func (c *Client) SendMouseMoveRelative(deltaX, deltaY int16) error {
	// For relative movement, we need to calculate the new position
	// This is a simplified implementation - in a full implementation,
	// you would track the current mouse position
	return c.sendMouseEvent(t128.PTRFLAGS_MOVE, uint16(deltaX), uint16(deltaY))
}

// SendMouseLeftDownEvent sends a left mouse button down event
func (c *Client) SendMouseLeftDownEvent(xPos, yPos uint16) error {
	return c.sendMouseEvent(t128.PTRFLAGS_DOWN|t128.PTRFLAGS_BUTTON1, xPos, yPos)
}

// SendMouseLeftUpEvent sends a left mouse button up event
func (c *Client) SendMouseLeftUpEvent(xPos, yPos uint16) error {
	return c.sendMouseEvent(t128.PTRFLAGS_BUTTON1, xPos, yPos)
}

// SendMouseRightDownEvent sends a right mouse button down event
func (c *Client) SendMouseRightDownEvent(xPos, yPos uint16) error {
	return c.sendMouseEvent(t128.PTRFLAGS_DOWN|t128.PTRFLAGS_BUTTON2, xPos, yPos)
}

// SendMouseRightUpEvent sends a right mouse button up event
func (c *Client) SendMouseRightUpEvent(xPos, yPos uint16) error {
	return c.sendMouseEvent(t128.PTRFLAGS_BUTTON2, xPos, yPos)
}

// SendMouseMiddleDownEvent sends a middle mouse button down event
func (c *Client) SendMouseMiddleDownEvent(xPos, yPos uint16) error {
	return c.sendMouseEvent(t128.PTRFLAGS_DOWN|t128.PTRFLAGS_BUTTON3, xPos, yPos)
}

// SendMouseMiddleUpEvent sends a middle mouse button up event
func (c *Client) SendMouseMiddleUpEvent(xPos, yPos uint16) error {
	return c.sendMouseEvent(t128.PTRFLAGS_BUTTON3, xPos, yPos)
}

// SendMouseX1DownEvent sends an X1 mouse button down event
func (c *Client) SendMouseX1DownEvent(xPos, yPos uint16) error {
	return c.sendMouseEvent(t128.PTRFLAGS_DOWN|t128.PTRFLAGS_BUTTON4, xPos, yPos)
}

// SendMouseX1UpEvent sends an X1 mouse button up event
func (c *Client) SendMouseX1UpEvent(xPos, yPos uint16) error {
	return c.sendMouseEvent(t128.PTRFLAGS_BUTTON4, xPos, yPos)
}

// SendMouseX2DownEvent sends an X2 mouse button down event
func (c *Client) SendMouseX2DownEvent(xPos, yPos uint16) error {
	return c.sendMouseEvent(t128.PTRFLAGS_DOWN|t128.PTRFLAGS_BUTTON5, xPos, yPos)
}

// SendMouseX2UpEvent sends an X2 mouse button up event
func (c *Client) SendMouseX2UpEvent(xPos, yPos uint16) error {
	return c.sendMouseEvent(t128.PTRFLAGS_BUTTON5, xPos, yPos)
}

// SendMouseClickEvent sends a complete mouse click (down + up) for the specified button
func (c *Client) SendMouseClickEvent(button t128.MouseButton, xPos, yPos uint16) error {
	// Send button down
	if err := c.SendMouseButtonEvent(button, true, xPos, yPos); err != nil {
		return err
	}
	// Send button up
	return c.SendMouseButtonEvent(button, false, xPos, yPos)
}

// SendMouseButtonEvent sends a mouse button event (down or up)
func (c *Client) SendMouseButtonEvent(button t128.MouseButton, down bool, xPos, yPos uint16) error {
	switch button {
	case t128.MouseButtonLeft:
		if down {
			return c.SendMouseLeftDownEvent(xPos, yPos)
		}
		return c.SendMouseLeftUpEvent(xPos, yPos)
	case t128.MouseButtonRight:
		if down {
			return c.SendMouseRightDownEvent(xPos, yPos)
		}
		return c.SendMouseRightUpEvent(xPos, yPos)
	case t128.MouseButtonMiddle:
		if down {
			return c.SendMouseMiddleDownEvent(xPos, yPos)
		}
		return c.SendMouseMiddleUpEvent(xPos, yPos)
	case t128.MouseButtonX1:
		if down {
			return c.SendMouseX1DownEvent(xPos, yPos)
		}
		return c.SendMouseX1UpEvent(xPos, yPos)
	case t128.MouseButtonX2:
		if down {
			return c.SendMouseX2DownEvent(xPos, yPos)
		}
		return c.SendMouseX2UpEvent(xPos, yPos)
	default:
		return fmt.Errorf("unsupported mouse button")
	}
}

// SendMouseWheelEvent sends a vertical mouse wheel event
func (c *Client) SendMouseWheelEvent(wheelDelta int16, xPos, yPos uint16) error {
	event := t128.NewFastPathMouseWheelEvent(wheelDelta, xPos, yPos)
	data := event.Serialize()
	_, err := c.stream.Write(data)
	return err
}

// SendMouseHorizontalWheelEvent sends a horizontal mouse wheel event
func (c *Client) SendMouseHorizontalWheelEvent(wheelDelta int16, xPos, yPos uint16) error {
	event := t128.NewFastPathMouseHorizontalWheelEvent(wheelDelta, xPos, yPos)
	data := event.Serialize()
	_, err := c.stream.Write(data)
	return err
}

// SendMouseDoubleClickEvent sends a double-click event for the specified button
func (c *Client) SendMouseDoubleClickEvent(button t128.MouseButton, xPos, yPos uint16) error {
	// Send first click
	if err := c.SendMouseClickEvent(button, xPos, yPos); err != nil {
		return err
	}

	// Small delay to simulate double-click timing
	time.Sleep(50 * time.Millisecond)

	// Send second click
	return c.SendMouseClickEvent(button, xPos, yPos)
}

// SendMouseDragEvent sends a mouse drag event (move with button pressed)
func (c *Client) SendMouseDragEvent(button t128.MouseButton, startX, startY, endX, endY uint16) error {
	// Press button at start position
	if err := c.SendMouseButtonEvent(button, true, startX, startY); err != nil {
		return err
	}

	// Move to end position
	if err := c.SendMouseMoveEvent(endX, endY); err != nil {
		return err
	}

	// Release button at end position
	return c.SendMouseButtonEvent(button, false, endX, endY)
}

// SendMouseSmoothDragEvent sends a smooth mouse drag event with intermediate points
func (c *Client) SendMouseSmoothDragEvent(button t128.MouseButton, startX, startY, endX, endY uint16, steps int) error {
	if steps < 2 {
		steps = 2
	}

	// Press button at start position
	if err := c.SendMouseButtonEvent(button, true, startX, startY); err != nil {
		return err
	}

	// Calculate step sizes
	stepX := float64(endX-startX) / float64(steps-1)
	stepY := float64(endY-startY) / float64(steps-1)

	// Move through intermediate points
	for i := 1; i < steps; i++ {
		x := uint16(float64(startX) + stepX*float64(i))
		y := uint16(float64(startY) + stepY*float64(i))

		if err := c.SendMouseMoveEvent(x, y); err != nil {
			return err
		}

		// Small delay for smooth movement
		time.Sleep(10 * time.Millisecond)
	}

	// Release button at end position
	return c.SendMouseButtonEvent(button, false, endX, endY)
}

// SendMouseMultiClickEvent sends multiple clicks for the specified button
func (c *Client) SendMouseMultiClickEvent(button t128.MouseButton, xPos, yPos uint16, count int) error {
	for i := 0; i < count; i++ {
		if err := c.SendMouseClickEvent(button, xPos, yPos); err != nil {
			return err
		}

		// Small delay between clicks
		if i < count-1 {
			time.Sleep(100 * time.Millisecond)
		}
	}
	return nil
}

// SendMouseScrollEvent sends a scroll event with specified direction and amount
func (c *Client) SendMouseScrollEvent(direction t128.ScrollDirection, amount int16, xPos, yPos uint16) error {
	var wheelDelta int16

	switch direction {
	case t128.ScrollUp:
		wheelDelta = amount
	case t128.ScrollDown:
		wheelDelta = -amount
	case t128.ScrollLeft:
		return c.SendMouseHorizontalWheelEvent(-amount, xPos, yPos)
	case t128.ScrollRight:
		return c.SendMouseHorizontalWheelEvent(amount, xPos, yPos)
	default:
		return fmt.Errorf("unsupported scroll direction")
	}

	return c.SendMouseWheelEvent(wheelDelta, xPos, yPos)
}
