package device

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/kdsmith18542/gordp/core"
	"github.com/kdsmith18542/gordp/glog"
)

func TestDeviceMessageSerialization(t *testing.T) {
	// Test device message serialization and deserialization
	originalMsg := &DeviceMessage{
		ComponentID: RDPDR_CTYP_CORE,
		PacketID:    0x1234,
		Data:        []byte{0x01, 0x02, 0x03, 0x04},
	}

	serialized := originalMsg.Serialize()
	if len(serialized) == 0 {
		t.Fatal("Serialized message is empty")
	}

	// Create a new reader for deserialization
	reader := bytes.NewReader(serialized)
	deserialized, err := ReadDeviceMessage(reader)
	if err != nil {
		t.Fatalf("Failed to deserialize message: %v", err)
	}

	if deserialized.ComponentID != originalMsg.ComponentID {
		t.Errorf("ComponentID mismatch: expected %d, got %d", originalMsg.ComponentID, deserialized.ComponentID)
	}

	if deserialized.PacketID != originalMsg.PacketID {
		t.Errorf("PacketID mismatch: expected %d, got %d", originalMsg.PacketID, deserialized.PacketID)
	}

	if !bytes.Equal(deserialized.Data, originalMsg.Data) {
		t.Errorf("Data mismatch: expected %v, got %v", originalMsg.Data, deserialized.Data)
	}
}

func TestDeviceManagerCreation(t *testing.T) {
	// Test device manager creation with nil handler
	dm := NewDeviceManager(nil)
	if dm == nil {
		t.Fatal("Device manager is nil")
	}

	if dm.GetDeviceCount() != 0 {
		t.Errorf("Expected 0 devices, got %d", dm.GetDeviceCount())
	}

	// Test device manager creation with custom handler
	customHandler := &DefaultDeviceHandler{}
	dm = NewDeviceManager(customHandler)
	if dm == nil {
		t.Fatal("Device manager is nil")
	}
}

func TestDeviceAnnouncement(t *testing.T) {
	dm := NewDeviceManager(nil)

	// Test device announcement message creation
	msg := dm.CreateDeviceAnnounceMessage(DeviceTypePrinter, "PRN1", "Printer data")
	if msg == nil {
		t.Fatal("Device announcement message is nil")
	}

	if msg.ComponentID != RDPDR_CTYP_CORE {
		t.Errorf("Expected component ID %d, got %d", RDPDR_CTYP_CORE, msg.ComponentID)
	}

	if msg.PacketID != uint16(PAKID_CORE_DEVICE_ANNOUNCE) {
		t.Errorf("Expected packet ID %d, got %d", PAKID_CORE_DEVICE_ANNOUNCE, msg.PacketID)
	}

	// Test processing device announcement
	err := dm.ProcessMessage(msg)
	if err != nil {
		t.Fatalf("Failed to process device announcement: %v", err)
	}

	if dm.GetDeviceCount() != 1 {
		t.Errorf("Expected 1 device, got %d", dm.GetDeviceCount())
	}

	devices := dm.ListDevices()
	if len(devices) != 1 {
		t.Errorf("Expected 1 device in list, got %d", len(devices))
	}

	if devices[0].DeviceType != DeviceTypePrinter {
		t.Errorf("Expected device type %d, got %d", DeviceTypePrinter, devices[0].DeviceType)
	}

	if devices[0].PreferredDosName != "PRN1" {
		t.Errorf("Expected DOS name 'PRN1', got '%s'", devices[0].PreferredDosName)
	}
}

func TestPrinterDataMessage(t *testing.T) {
	dm := NewDeviceManager(nil)

	// Test printer data message creation
	printerData := []byte{0x01, 0x02, 0x03, 0x04, 0x05}
	msg := dm.CreatePrinterDataMessage(123, printerData, 0x0001)
	if msg == nil {
		t.Fatal("Printer data message is nil")
	}

	if msg.ComponentID != RDPDR_CTYP_PRN {
		t.Errorf("Expected component ID %d, got %d", RDPDR_CTYP_PRN, msg.ComponentID)
	}

	if msg.PacketID != uint16(PAKID_PRN_CACHE_DATA) {
		t.Errorf("Expected packet ID %d, got %d", PAKID_PRN_CACHE_DATA, msg.PacketID)
	}

	// Test processing printer data message
	err := dm.ProcessMessage(msg)
	if err != nil {
		t.Fatalf("Failed to process printer data message: %v", err)
	}
}

func TestDeviceIORequest(t *testing.T) {
	dm := NewDeviceManager(nil)

	// Create a device first
	announceMsg := dm.CreateDeviceAnnounceMessage(DeviceTypeDrive, "C:", "Drive data")
	err := dm.ProcessMessage(announceMsg)
	if err != nil {
		t.Fatalf("Failed to announce device: %v", err)
	}

	// Create I/O request message
	buf := new(bytes.Buffer)
	core.WriteLE(buf, uint16(PAKID_CORE_DEVICE_IOREQUEST))
	core.WriteLE(buf, uint32(1))        // Device ID
	core.WriteLE(buf, uint32(1))        // File ID
	core.WriteLE(buf, uint32(1))        // Completion ID
	core.WriteLE(buf, uint32(0))        // Major function
	core.WriteLE(buf, uint32(0))        // Minor function
	buf.Write([]byte{0x01, 0x02, 0x03}) // Request data

	ioMsg := &DeviceMessage{
		ComponentID: RDPDR_CTYP_CORE,
		PacketID:    uint16(PAKID_CORE_DEVICE_IOREQUEST),
		Data:        buf.Bytes(),
	}

	// Test processing I/O request
	err = dm.ProcessMessage(ioMsg)
	if err != nil {
		t.Fatalf("Failed to process I/O request: %v", err)
	}
}

func TestCustomDeviceHandler(t *testing.T) {
	// Create a custom device handler for testing
	customHandler := &TestDeviceHandler{
		announceCount:    0,
		ioRequestCount:   0,
		printerDataCount: 0,
	}

	dm := NewDeviceManager(customHandler)

	// Test device announcement
	announceMsg := dm.CreateDeviceAnnounceMessage(DeviceTypePrinter, "TEST", "Test data")
	err := dm.ProcessMessage(announceMsg)
	if err != nil {
		t.Fatalf("Failed to process device announcement: %v", err)
	}

	if customHandler.announceCount != 1 {
		t.Errorf("Expected 1 device announcement, got %d", customHandler.announceCount)
	}

	// Test I/O request
	buf := new(bytes.Buffer)
	core.WriteLE(buf, uint16(PAKID_CORE_DEVICE_IOREQUEST))
	core.WriteLE(buf, uint32(1))        // Device ID
	core.WriteLE(buf, uint32(1))        // File ID
	core.WriteLE(buf, uint32(1))        // Completion ID
	core.WriteLE(buf, uint32(0))        // Major function
	core.WriteLE(buf, uint32(0))        // Minor function
	buf.Write([]byte{0x01, 0x02, 0x03}) // Request data

	ioMsg := &DeviceMessage{
		ComponentID: RDPDR_CTYP_CORE,
		PacketID:    uint16(PAKID_CORE_DEVICE_IOREQUEST),
		Data:        buf.Bytes(),
	}

	err = dm.ProcessMessage(ioMsg)
	if err != nil {
		t.Fatalf("Failed to process I/O request: %v", err)
	}

	if customHandler.ioRequestCount != 1 {
		t.Errorf("Expected 1 I/O request, got %d", customHandler.ioRequestCount)
	}

	// Test printer data
	printerMsg := dm.CreatePrinterDataMessage(456, []byte{0x01, 0x02}, 0x0002)
	err = dm.ProcessMessage(printerMsg)
	if err != nil {
		t.Fatalf("Failed to process printer data: %v", err)
	}

	if customHandler.printerDataCount != 1 {
		t.Errorf("Expected 1 printer data, got %d", customHandler.printerDataCount)
	}
}

func TestDeviceManagerStats(t *testing.T) {
	dm := NewDeviceManager(nil)

	// Add multiple devices of different types
	devices := []struct {
		deviceType DeviceType
		dosName    string
		data       string
	}{
		{DeviceTypePrinter, "PRN1", "Printer 1"},
		{DeviceTypePrinter, "PRN2", "Printer 2"},
		{DeviceTypeDrive, "C:", "C Drive"},
		{DeviceTypePort, "COM1", "Serial Port"},
	}

	for _, dev := range devices {
		msg := dm.CreateDeviceAnnounceMessage(dev.deviceType, dev.dosName, dev.data)
		err := dm.ProcessMessage(msg)
		if err != nil {
			t.Fatalf("Failed to announce device %s: %v", dev.dosName, err)
		}
	}

	// Test device count
	if dm.GetDeviceCount() != 4 {
		t.Errorf("Expected 4 devices, got %d", dm.GetDeviceCount())
	}

	// Test device stats
	stats := dm.GetDeviceStats()
	if stats["total_devices"] != 4 {
		t.Errorf("Expected total_devices to be 4, got %v", stats["total_devices"])
	}

	devicesByType, ok := stats["devices_by_type"].(map[DeviceType]int)
	if !ok {
		t.Fatal("devices_by_type is not a map")
	}

	if devicesByType[DeviceTypePrinter] != 2 {
		t.Errorf("Expected 2 printers, got %d", devicesByType[DeviceTypePrinter])
	}

	if devicesByType[DeviceTypeDrive] != 1 {
		t.Errorf("Expected 1 drive, got %d", devicesByType[DeviceTypeDrive])
	}

	if devicesByType[DeviceTypePort] != 1 {
		t.Errorf("Expected 1 port, got %d", devicesByType[DeviceTypePort])
	}
}

func TestDeviceRemoval(t *testing.T) {
	dm := NewDeviceManager(nil)

	// Add a device
	msg := dm.CreateDeviceAnnounceMessage(DeviceTypePrinter, "PRN1", "Printer data")
	err := dm.ProcessMessage(msg)
	if err != nil {
		t.Fatalf("Failed to announce device: %v", err)
	}

	if dm.GetDeviceCount() != 1 {
		t.Errorf("Expected 1 device, got %d", dm.GetDeviceCount())
	}

	// Get the device ID
	devices := dm.ListDevices()
	if len(devices) == 0 {
		t.Fatal("No devices found")
	}

	deviceID := devices[0].DeviceID

	// Remove the device
	dm.RemoveDevice(deviceID)

	if dm.GetDeviceCount() != 0 {
		t.Errorf("Expected 0 devices after removal, got %d", dm.GetDeviceCount())
	}

	// Verify device is not found
	_, exists := dm.GetDevice(deviceID)
	if exists {
		t.Error("Device still exists after removal")
	}
}

func TestDeviceLimit(t *testing.T) {
	dm := NewDeviceManager(nil)

	// Try to add more than 10 devices (the limit)
	for i := 0; i < 12; i++ {
		msg := dm.CreateDeviceAnnounceMessage(DeviceTypePrinter, fmt.Sprintf("PRN%d", i), "Printer data")
		err := dm.ProcessMessage(msg)
		if i < 10 {
			if err != nil {
				t.Fatalf("Failed to announce device %d: %v", i, err)
			}
		} else {
			if err == nil {
				t.Errorf("Expected error when adding device %d, but got none", i)
			}
		}
	}

	if dm.GetDeviceCount() != 10 {
		t.Errorf("Expected 10 devices (limit), got %d", dm.GetDeviceCount())
	}
}

func TestInvalidMessageHandling(t *testing.T) {
	dm := NewDeviceManager(nil)

	// Test invalid message with insufficient data
	invalidMsg := &DeviceMessage{
		ComponentID: RDPDR_CTYP_CORE,
		PacketID:    uint16(PAKID_CORE_DEVICE_ANNOUNCE),
		Data:        []byte{0x01}, // Too short
	}

	err := dm.ProcessMessage(invalidMsg)
	if err == nil {
		t.Error("Expected error for invalid message, but got none")
	}

	// Test unknown component ID
	unknownMsg := &DeviceMessage{
		ComponentID: 0x9999, // Unknown component
		PacketID:    0x0001,
		Data:        []byte{},
	}

	err = dm.ProcessMessage(unknownMsg)
	if err != nil {
		t.Errorf("Unexpected error for unknown component: %v", err)
	}
}

// TestDeviceHandler is a test implementation of DeviceHandler
type TestDeviceHandler struct {
	announceCount    int
	ioRequestCount   int
	printerDataCount int
}

func (h *TestDeviceHandler) OnDeviceAnnounce(device *DeviceAnnounce) error {
	h.announceCount++
	return nil
}

func (h *TestDeviceHandler) OnDeviceIORequest(request *DeviceIORequest) (*DeviceIOCompletion, error) {
	h.ioRequestCount++
	return &DeviceIOCompletion{
		DeviceID:     request.DeviceID,
		CompletionID: request.CompletionID,
		IoStatus:     0, // STATUS_SUCCESS
		Data:         []byte{},
	}, nil
}

func (h *TestDeviceHandler) OnPrinterData(data *PrinterData) error {
	h.printerDataCount++
	return nil
}

func (h *TestDeviceHandler) OnDriveAccess(path string, operation string) error {
	return nil
}

func (h *TestDeviceHandler) OnPortAccess(portName string, operation string) error {
	return nil
}

func TestMain(m *testing.M) {
	// Set log level to suppress debug output during tests
	glog.SetLevel(glog.ERROR)
	m.Run()
}
