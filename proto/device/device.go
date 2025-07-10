package device

import (
	"bytes"
	"fmt"
	"io"
	"sync"

	"github.com/kdsmith18542/gordp/core"
	"github.com/kdsmith18542/gordp/glog"
)

// DeviceType represents the type of device being redirected
type DeviceType uint32

const (
	DeviceTypePrinter   DeviceType = 0x00000001
	DeviceTypeDrive     DeviceType = 0x00000002
	DeviceTypePort      DeviceType = 0x00000003
	DeviceTypeSmartCard DeviceType = 0x00000004
	DeviceTypeAudio     DeviceType = 0x00000005
	DeviceTypeVideo     DeviceType = 0x00000006
	DeviceTypeUSB       DeviceType = 0x00000007
)

// DeviceMessageType represents the type of device redirection message
type DeviceMessageType uint16

const (
	RDPDR_CTYP_CORE  DeviceMessageType = 0x4472 // "Dr"
	RDPDR_CTYP_PRN   DeviceMessageType = 0x5052 // "Pr"
	RDPDR_CTYP_SCARD DeviceMessageType = 0x5343 // "Sc"
)

// CoreMessageType represents the type of core device message
type CoreMessageType uint16

const (
	PAKID_CORE_DEVICE_ANNOUNCE                 CoreMessageType = 0x0001
	PAKID_CORE_DEVICE_REPLY_ANNOUNCE           CoreMessageType = 0x0002
	PAKID_CORE_DEVICE_IOREQUEST                CoreMessageType = 0x0003
	PAKID_CORE_DEVICE_IOCOMPLETION             CoreMessageType = 0x0004
	PAKID_CORE_DEVICE_CREATE                   CoreMessageType = 0x0005
	PAKID_CORE_DEVICE_CLOSE                    CoreMessageType = 0x0006
	PAKID_CORE_DEVICE_READ                     CoreMessageType = 0x0007
	PAKID_CORE_DEVICE_WRITE                    CoreMessageType = 0x0008
	PAKID_CORE_DEVICE_QUERY_INFORMATION        CoreMessageType = 0x0009
	PAKID_CORE_DEVICE_SET_INFORMATION          CoreMessageType = 0x000A
	PAKID_CORE_DEVICE_QUERY_VOLUME_INFORMATION CoreMessageType = 0x000B
	PAKID_CORE_DEVICE_SET_VOLUME_INFORMATION   CoreMessageType = 0x000C
	PAKID_CORE_DEVICE_QUERY_DIRECTORY          CoreMessageType = 0x000D
	PAKID_CORE_DEVICE_NOTIFY_CHANGE_DIRECTORY  CoreMessageType = 0x000E
	PAKID_CORE_DEVICE_LOCK_CONTROL             CoreMessageType = 0x000F
	PAKID_CORE_DEVICE_QUERY_EA                 CoreMessageType = 0x0010
	PAKID_CORE_DEVICE_SET_EA                   CoreMessageType = 0x0011
	PAKID_CORE_DEVICE_FLUSH_BUFFERS            CoreMessageType = 0x0012
	PAKID_CORE_DEVICE_QUERY_QUOTA              CoreMessageType = 0x0013
	PAKID_CORE_DEVICE_SET_QUOTA                CoreMessageType = 0x0014
	PAKID_CORE_DEVICE_QUERY_SECURITY           CoreMessageType = 0x0015
	PAKID_CORE_DEVICE_SET_SECURITY             CoreMessageType = 0x0016
)

// PrinterMessageType represents the type of printer message
type PrinterMessageType uint16

const (
	PAKID_PRN_CACHE_DATA     PrinterMessageType = 0x0001
	PAKID_PRN_USING_XPS      PrinterMessageType = 0x0002
	PAKID_PRN_CACHE_DATA_XPS PrinterMessageType = 0x0003
)

// DeviceMessage represents a device redirection message
type DeviceMessage struct {
	ComponentID DeviceMessageType
	PacketID    uint16
	Data        []byte
}

// DeviceAnnounce represents a device announcement
type DeviceAnnounce struct {
	DeviceType       DeviceType
	DeviceID         uint32
	PreferredDosName string
	DeviceData       string
}

// DeviceReplyAnnounce represents a device reply announcement
type DeviceReplyAnnounce struct {
	DeviceID   uint32
	ResultCode uint32
}

// DeviceIORequest represents a device I/O request
type DeviceIORequest struct {
	DeviceID      uint32
	FileID        uint32
	CompletionID  uint32
	MajorFunction uint32
	MinorFunction uint32
	Data          []byte
}

// DeviceIOCompletion represents a device I/O completion
type DeviceIOCompletion struct {
	DeviceID     uint32
	CompletionID uint32
	IoStatus     uint32
	Data         []byte
}

// PrinterData represents printer data
type PrinterData struct {
	JobID uint32
	Data  []byte
	Flags uint32
}

// DeviceHandler handles device events
type DeviceHandler interface {
	OnDeviceAnnounce(device *DeviceAnnounce) error
	OnDeviceIORequest(request *DeviceIORequest) (*DeviceIOCompletion, error)
	OnPrinterData(data *PrinterData) error
	OnDriveAccess(path string, operation string) error
	OnPortAccess(portName string, operation string) error
}

// DefaultDeviceHandler provides a default implementation
type DefaultDeviceHandler struct{}

// NewDefaultDeviceHandler creates a new default device handler
func NewDefaultDeviceHandler() *DefaultDeviceHandler {
	return &DefaultDeviceHandler{}
}

// OnDeviceAnnounce handles device announcement events
func (h *DefaultDeviceHandler) OnDeviceAnnounce(device *DeviceAnnounce) error {
	glog.Debugf("Device announced: type=%d, id=%d, name=%s",
		device.DeviceType, device.DeviceID, device.PreferredDosName)
	return nil
}

// OnDeviceIORequest handles device I/O request events
func (h *DefaultDeviceHandler) OnDeviceIORequest(request *DeviceIORequest) (*DeviceIOCompletion, error) {
	glog.Debugf("Device I/O request: device=%d, file=%d, major=%d, minor=%d",
		request.DeviceID, request.FileID, request.MajorFunction, request.MinorFunction)

	// Default success response
	return &DeviceIOCompletion{
		DeviceID:     request.DeviceID,
		CompletionID: request.CompletionID,
		IoStatus:     0, // STATUS_SUCCESS
		Data:         []byte{},
	}, nil
}

// OnPrinterData handles printer data events
func (h *DefaultDeviceHandler) OnPrinterData(data *PrinterData) error {
	glog.Debugf("Printer data: job=%d, size=%d bytes, flags=0x%08X",
		data.JobID, len(data.Data), data.Flags)
	return nil
}

// OnDriveAccess handles drive access events
func (h *DefaultDeviceHandler) OnDriveAccess(path string, operation string) error {
	glog.Debugf("Drive access: %s %s", operation, path)
	return nil
}

// OnPortAccess handles port access events
func (h *DefaultDeviceHandler) OnPortAccess(portName string, operation string) error {
	glog.Debugf("Port access: %s %s", operation, portName)
	return nil
}

// DeviceManager manages device redirection
type DeviceManager struct {
	devices map[uint32]*DeviceAnnounce
	handler DeviceHandler
	mutex   sync.RWMutex
	nextID  uint32
}

// NewDeviceManager creates a new device manager
func NewDeviceManager(handler DeviceHandler) *DeviceManager {
	if handler == nil {
		handler = NewDefaultDeviceHandler()
	}

	return &DeviceManager{
		devices: make(map[uint32]*DeviceAnnounce),
		handler: handler,
		nextID:  1,
	}
}

// ReadDeviceMessage reads a device message from the stream
func ReadDeviceMessage(r io.Reader) (*DeviceMessage, error) {
	msg := &DeviceMessage{}

	// Read message header
	core.ReadLE(r, &msg.ComponentID)
	core.ReadLE(r, &msg.PacketID)

	// Read remaining data
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read message data: %w", err)
	}
	msg.Data = data

	return msg, nil
}

// Serialize serializes the device message
func (m *DeviceMessage) Serialize() []byte {
	buf := new(bytes.Buffer)
	core.WriteLE(buf, m.ComponentID)
	core.WriteLE(buf, m.PacketID)
	if len(m.Data) > 0 {
		buf.Write(m.Data)
	}
	return buf.Bytes()
}

// ProcessMessage processes a device message
func (dm *DeviceManager) ProcessMessage(msg *DeviceMessage) error {
	switch msg.ComponentID {
	case RDPDR_CTYP_CORE:
		return dm.handleCoreMessage(msg)
	case RDPDR_CTYP_PRN:
		return dm.handlePrinterMessage(msg)
	default:
		glog.Debugf("Unhandled device message component: 0x%04X", msg.ComponentID)
		return nil
	}
}

// handleCoreMessage handles core device messages
func (dm *DeviceManager) handleCoreMessage(msg *DeviceMessage) error {
	if len(msg.Data) < 2 {
		return fmt.Errorf("invalid core message size")
	}

	reader := bytes.NewReader(msg.Data)
	var messageType CoreMessageType
	core.ReadLE(reader, &messageType)

	switch messageType {
	case PAKID_CORE_DEVICE_ANNOUNCE:
		return dm.handleDeviceAnnounce(reader)
	case PAKID_CORE_DEVICE_IOREQUEST:
		return dm.handleDeviceIORequest(reader)
	default:
		glog.Debugf("Unhandled core message type: 0x%04X", messageType)
		return nil
	}
}

// handlePrinterMessage handles printer messages
func (dm *DeviceManager) handlePrinterMessage(msg *DeviceMessage) error {
	if len(msg.Data) < 2 {
		return fmt.Errorf("invalid printer message size")
	}

	reader := bytes.NewReader(msg.Data)
	var messageType PrinterMessageType
	core.ReadLE(reader, &messageType)

	switch messageType {
	case PAKID_PRN_CACHE_DATA:
		return dm.handlePrinterData(reader)
	default:
		glog.Debugf("Unhandled printer message type: 0x%04X", messageType)
		return nil
	}
}

// handleDeviceAnnounce handles device announcement
func (dm *DeviceManager) handleDeviceAnnounce(r io.Reader) error {
	if len(dm.devices) >= 10 { // Limit number of devices
		return fmt.Errorf("too many devices")
	}

	device := &DeviceAnnounce{}
	core.ReadLE(r, &device.DeviceType)
	core.ReadLE(r, &device.DeviceID)

	// Read preferred DOS name (8 characters)
	dosNameBytes := make([]byte, 8)
	r.Read(dosNameBytes)
	device.PreferredDosName = string(bytes.TrimRight(dosNameBytes, "\x00"))

	// Read device data
	deviceData, err := io.ReadAll(r)
	if err != nil {
		return fmt.Errorf("failed to read device data: %w", err)
	}
	device.DeviceData = string(deviceData)

	dm.mutex.Lock()
	dm.devices[device.DeviceID] = device
	dm.mutex.Unlock()

	glog.GetStructuredLogger().InfoStructured("Device announced", map[string]interface{}{
		"device_type": device.DeviceType,
		"device_id":   device.DeviceID,
		"dos_name":    device.PreferredDosName,
	})

	return dm.handler.OnDeviceAnnounce(device)
}

// handleDeviceIORequest handles device I/O request
func (dm *DeviceManager) handleDeviceIORequest(r io.Reader) error {
	request := &DeviceIORequest{}
	core.ReadLE(r, &request.DeviceID)
	core.ReadLE(r, &request.FileID)
	core.ReadLE(r, &request.CompletionID)
	core.ReadLE(r, &request.MajorFunction)
	core.ReadLE(r, &request.MinorFunction)

	// Read request data
	data, err := io.ReadAll(r)
	if err != nil {
		return fmt.Errorf("failed to read request data: %w", err)
	}
	request.Data = data

	glog.GetStructuredLogger().DebugStructured("Device I/O request", map[string]interface{}{
		"device_id":      request.DeviceID,
		"file_id":        request.FileID,
		"completion_id":  request.CompletionID,
		"major_function": request.MajorFunction,
		"minor_function": request.MinorFunction,
		"data_size":      len(request.Data),
	})

	completion, err := dm.handler.OnDeviceIORequest(request)
	if err != nil {
		glog.GetStructuredLogger().ErrorStructured("Device I/O request failed", err, map[string]interface{}{
			"device_id": request.DeviceID,
			"file_id":   request.FileID,
		})
		return err
	}

	// Send completion response
	return dm.sendIOCompletion(completion)
}

// handlePrinterData handles printer data
func (dm *DeviceManager) handlePrinterData(r io.Reader) error {
	data := &PrinterData{}
	core.ReadLE(r, &data.JobID)
	core.ReadLE(r, &data.Flags)

	// Read printer data
	printerData, err := io.ReadAll(r)
	if err != nil {
		return fmt.Errorf("failed to read printer data: %w", err)
	}
	data.Data = printerData

	glog.GetStructuredLogger().InfoStructured("Printer data received", map[string]interface{}{
		"job_id": data.JobID,
		"size":   len(data.Data),
		"flags":  data.Flags,
	})

	return dm.handler.OnPrinterData(data)
}

// sendIOCompletion sends an I/O completion response
func (dm *DeviceManager) sendIOCompletion(completion *DeviceIOCompletion) error {
	buf := new(bytes.Buffer)
	core.WriteLE(buf, uint16(PAKID_CORE_DEVICE_IOCOMPLETION))
	core.WriteLE(buf, completion.DeviceID)
	core.WriteLE(buf, completion.CompletionID)
	core.WriteLE(buf, completion.IoStatus)
	if len(completion.Data) > 0 {
		buf.Write(completion.Data)
	}

	// In a real implementation, this would be sent through the virtual channel
	glog.Debugf("I/O completion: device=%d, completion=%d, status=0x%08X",
		completion.DeviceID, completion.CompletionID, completion.IoStatus)

	return nil
}

// CreateDeviceAnnounceMessage creates a device announcement message
func (dm *DeviceManager) CreateDeviceAnnounceMessage(deviceType DeviceType, preferredDosName, deviceData string) *DeviceMessage {
	dm.mutex.Lock()
	deviceID := dm.nextID
	dm.nextID++
	dm.mutex.Unlock()

	buf := new(bytes.Buffer)
	core.WriteLE(buf, uint16(PAKID_CORE_DEVICE_ANNOUNCE))
	core.WriteLE(buf, deviceType)
	core.WriteLE(buf, deviceID)

	// Write DOS name (8 characters, padded with nulls)
	dosNameBytes := make([]byte, 8)
	copy(dosNameBytes, []byte(preferredDosName))
	buf.Write(dosNameBytes)

	// Write device data
	if len(deviceData) > 0 {
		buf.Write([]byte(deviceData))
	}

	return &DeviceMessage{
		ComponentID: RDPDR_CTYP_CORE,
		PacketID:    uint16(PAKID_CORE_DEVICE_ANNOUNCE),
		Data:        buf.Bytes(),
	}
}

// CreatePrinterDataMessage creates a printer data message
func (dm *DeviceManager) CreatePrinterDataMessage(jobID uint32, data []byte, flags uint32) *DeviceMessage {
	buf := new(bytes.Buffer)
	core.WriteLE(buf, uint16(PAKID_PRN_CACHE_DATA))
	core.WriteLE(buf, jobID)
	core.WriteLE(buf, flags)
	buf.Write(data)

	return &DeviceMessage{
		ComponentID: RDPDR_CTYP_PRN,
		PacketID:    uint16(PAKID_PRN_CACHE_DATA),
		Data:        buf.Bytes(),
	}
}

// GetDevice returns a device by ID
func (dm *DeviceManager) GetDevice(deviceID uint32) (*DeviceAnnounce, bool) {
	dm.mutex.RLock()
	defer dm.mutex.RUnlock()
	device, exists := dm.devices[deviceID]
	return device, exists
}

// ListDevices returns all registered devices
func (dm *DeviceManager) ListDevices() []*DeviceAnnounce {
	dm.mutex.RLock()
	defer dm.mutex.RUnlock()

	devices := make([]*DeviceAnnounce, 0, len(dm.devices))
	for _, device := range dm.devices {
		devices = append(devices, device)
	}
	return devices
}

// RemoveDevice removes a device
func (dm *DeviceManager) RemoveDevice(deviceID uint32) {
	dm.mutex.Lock()
	defer dm.mutex.Unlock()
	delete(dm.devices, deviceID)
}

// GetDeviceCount returns the number of registered devices
func (dm *DeviceManager) GetDeviceCount() int {
	dm.mutex.RLock()
	defer dm.mutex.RUnlock()
	return len(dm.devices)
}

// GetDeviceStats returns statistics about device usage
func (dm *DeviceManager) GetDeviceStats() map[string]interface{} {
	dm.mutex.RLock()
	defer dm.mutex.RUnlock()

	stats := make(map[string]interface{})
	stats["total_devices"] = len(dm.devices)

	// Count by device type
	typeCount := make(map[DeviceType]int)
	for _, device := range dm.devices {
		typeCount[device.DeviceType]++
	}
	stats["devices_by_type"] = typeCount

	return stats
}
