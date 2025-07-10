package virtualchannels

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/kdsmith18542/gordp/proto/device"
)

// DeviceHandler handles device redirection
type DeviceHandler struct {
	manager *VirtualChannelManager

	// Device state
	mu        sync.RWMutex
	devices   map[string]*device.DeviceAnnounce
	deviceIDs map[uint32]string // Map device ID to name for reverse lookup
	isEnabled bool
}

// NewDeviceHandler creates a new device handler
func NewDeviceHandler(manager *VirtualChannelManager) *DeviceHandler {
	return &DeviceHandler{
		manager:   manager,
		devices:   make(map[string]*device.DeviceAnnounce),
		deviceIDs: make(map[uint32]string),
		isEnabled: true,
	}
}

// OnDeviceAnnounce is called when a device is announced
func (h *DeviceHandler) OnDeviceAnnounce(device *device.DeviceAnnounce) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if device == nil {
		return fmt.Errorf("device announcement is nil")
	}

	// Validate device data
	if device.PreferredDosName == "" {
		return fmt.Errorf("device preferred DOS name is empty")
	}

	// Store device by both name and ID
	h.devices[device.PreferredDosName] = device
	h.deviceIDs[device.DeviceID] = device.PreferredDosName
	h.manager.SetChannelOpen("device", true)

	fmt.Printf("Device announced: %s (ID: %d, Type: %d, Data: %s)\n",
		device.PreferredDosName, device.DeviceID, device.DeviceType, device.DeviceData)
	return nil
}

// OnDeviceIORequest is called when a device I/O request is received
func (h *DeviceHandler) OnDeviceIORequest(request *device.DeviceIORequest) (*device.DeviceIOCompletion, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if request == nil {
		return nil, fmt.Errorf("device I/O request is nil")
	}

	// Get device name for logging
	deviceName, exists := h.deviceIDs[request.DeviceID]
	if !exists {
		deviceName = fmt.Sprintf("Unknown(%d)", request.DeviceID)
	}

	fmt.Printf("Device I/O request: %s (deviceID=%d, fileID=%d, major=%d, minor=%d, %d bytes)\n",
		deviceName, request.DeviceID, request.FileID, request.MajorFunction, request.MinorFunction, len(request.Data))

	// Handle different major functions
	switch request.MajorFunction {
	case 0x00000000: // IRP_MJ_CREATE
		return h.handleCreateRequest(request)
	case 0x00000001: // IRP_MJ_READ
		return h.handleReadRequest(request)
	case 0x00000002: // IRP_MJ_WRITE
		return h.handleWriteRequest(request)
	case 0x00000003: // IRP_MJ_QUERY_INFORMATION
		return h.handleQueryInformationRequest(request)
	case 0x00000004: // IRP_MJ_SET_INFORMATION
		return h.handleSetInformationRequest(request)
	case 0x00000005: // IRP_MJ_QUERY_EA
		return h.handleQueryEARequest(request)
	case 0x00000006: // IRP_MJ_SET_EA
		return h.handleSetEARequest(request)
	case 0x00000007: // IRP_MJ_FLUSH_BUFFERS
		return h.handleFlushBuffersRequest(request)
	case 0x00000008: // IRP_MJ_QUERY_VOLUME_INFORMATION
		return h.handleQueryVolumeInformationRequest(request)
	case 0x00000009: // IRP_MJ_SET_VOLUME_INFORMATION
		return h.handleSetVolumeInformationRequest(request)
	case 0x0000000A: // IRP_MJ_DIRECTORY_CONTROL
		return h.handleDirectoryControlRequest(request)
	case 0x0000000B: // IRP_MJ_FILE_SYSTEM_CONTROL
		return h.handleFileSystemControlRequest(request)
	case 0x0000000C: // IRP_MJ_DEVICE_CONTROL
		return h.handleDeviceControlRequest(request)
	case 0x0000000D: // IRP_MJ_INTERNAL_DEVICE_CONTROL
		return h.handleInternalDeviceControlRequest(request)
	case 0x0000000E: // IRP_MJ_SHUTDOWN
		return h.handleShutdownRequest(request)
	case 0x0000000F: // IRP_MJ_LOCK_CONTROL
		return h.handleLockControlRequest(request)
	case 0x00000010: // IRP_MJ_CLEANUP
		return h.handleCleanupRequest(request)
	case 0x00000011: // IRP_MJ_CREATE_MAILSLOT
		return h.handleCreateMailslotRequest(request)
	case 0x00000012: // IRP_MJ_QUERY_SECURITY
		return h.handleQuerySecurityRequest(request)
	case 0x00000013: // IRP_MJ_SET_SECURITY
		return h.handleSetSecurityRequest(request)
	case 0x00000014: // IRP_MJ_POWER
		return h.handlePowerRequest(request)
	case 0x00000015: // IRP_MJ_SYSTEM_CONTROL
		return h.handleSystemControlRequest(request)
	case 0x00000016: // IRP_MJ_DEVICE_CHANGE
		return h.handleDeviceChangeRequest(request)
	case 0x00000017: // IRP_MJ_QUERY_QUOTA
		return h.handleQueryQuotaRequest(request)
	case 0x00000018: // IRP_MJ_SET_QUOTA
		return h.handleSetQuotaRequest(request)
	case 0x00000019: // IRP_MJ_PNP
		return h.handlePnPRequest(request)
	default:
		fmt.Printf("Unknown major function: 0x%08X\n", request.MajorFunction)
	}

	// Return default success response
	return &device.DeviceIOCompletion{
		DeviceID:     request.DeviceID,
		CompletionID: request.CompletionID,
		IoStatus:     0, // STATUS_SUCCESS
		Data:         []byte{},
	}, nil
}

// handleCreateRequest handles IRP_MJ_CREATE requests
func (h *DeviceHandler) handleCreateRequest(request *device.DeviceIORequest) (*device.DeviceIOCompletion, error) {
	fmt.Printf("Create request: fileID=%d, minor=%d\n", request.FileID, request.MinorFunction)
	return &device.DeviceIOCompletion{
		DeviceID:     request.DeviceID,
		CompletionID: request.CompletionID,
		IoStatus:     0, // STATUS_SUCCESS
		Data:         []byte{},
	}, nil
}

// handleReadRequest handles IRP_MJ_READ requests
func (h *DeviceHandler) handleReadRequest(request *device.DeviceIORequest) (*device.DeviceIOCompletion, error) {
	fmt.Printf("Read request: fileID=%d, %d bytes\n", request.FileID, len(request.Data))
	return &device.DeviceIOCompletion{
		DeviceID:     request.DeviceID,
		CompletionID: request.CompletionID,
		IoStatus:     0,        // STATUS_SUCCESS
		Data:         []byte{}, // Empty data for now
	}, nil
}

// handleWriteRequest handles IRP_MJ_WRITE requests
func (h *DeviceHandler) handleWriteRequest(request *device.DeviceIORequest) (*device.DeviceIOCompletion, error) {
	fmt.Printf("Write request: fileID=%d, %d bytes\n", request.FileID, len(request.Data))
	return &device.DeviceIOCompletion{
		DeviceID:     request.DeviceID,
		CompletionID: request.CompletionID,
		IoStatus:     0, // STATUS_SUCCESS
		Data:         []byte{},
	}, nil
}

// handleQueryInformationRequest handles IRP_MJ_QUERY_INFORMATION requests
func (h *DeviceHandler) handleQueryInformationRequest(request *device.DeviceIORequest) (*device.DeviceIOCompletion, error) {
	fmt.Printf("Query information request: fileID=%d, minor=%d\n", request.FileID, request.MinorFunction)
	return &device.DeviceIOCompletion{
		DeviceID:     request.DeviceID,
		CompletionID: request.CompletionID,
		IoStatus:     0, // STATUS_SUCCESS
		Data:         []byte{},
	}, nil
}

// handleSetInformationRequest handles IRP_MJ_SET_INFORMATION requests
func (h *DeviceHandler) handleSetInformationRequest(request *device.DeviceIORequest) (*device.DeviceIOCompletion, error) {
	fmt.Printf("Set information request: fileID=%d, minor=%d\n", request.FileID, request.MinorFunction)
	return &device.DeviceIOCompletion{
		DeviceID:     request.DeviceID,
		CompletionID: request.CompletionID,
		IoStatus:     0, // STATUS_SUCCESS
		Data:         []byte{},
	}, nil
}

// handleQueryEARequest handles IRP_MJ_QUERY_EA requests
func (h *DeviceHandler) handleQueryEARequest(request *device.DeviceIORequest) (*device.DeviceIOCompletion, error) {
	fmt.Printf("Query EA request: fileID=%d\n", request.FileID)
	return &device.DeviceIOCompletion{
		DeviceID:     request.DeviceID,
		CompletionID: request.CompletionID,
		IoStatus:     0, // STATUS_SUCCESS
		Data:         []byte{},
	}, nil
}

// handleSetEARequest handles IRP_MJ_SET_EA requests
func (h *DeviceHandler) handleSetEARequest(request *device.DeviceIORequest) (*device.DeviceIOCompletion, error) {
	fmt.Printf("Set EA request: fileID=%d\n", request.FileID)
	return &device.DeviceIOCompletion{
		DeviceID:     request.DeviceID,
		CompletionID: request.CompletionID,
		IoStatus:     0, // STATUS_SUCCESS
		Data:         []byte{},
	}, nil
}

// handleFlushBuffersRequest handles IRP_MJ_FLUSH_BUFFERS requests
func (h *DeviceHandler) handleFlushBuffersRequest(request *device.DeviceIORequest) (*device.DeviceIOCompletion, error) {
	fmt.Printf("Flush buffers request: fileID=%d\n", request.FileID)
	return &device.DeviceIOCompletion{
		DeviceID:     request.DeviceID,
		CompletionID: request.CompletionID,
		IoStatus:     0, // STATUS_SUCCESS
		Data:         []byte{},
	}, nil
}

// handleQueryVolumeInformationRequest handles IRP_MJ_QUERY_VOLUME_INFORMATION requests
func (h *DeviceHandler) handleQueryVolumeInformationRequest(request *device.DeviceIORequest) (*device.DeviceIOCompletion, error) {
	fmt.Printf("Query volume information request: fileID=%d\n", request.FileID)
	return &device.DeviceIOCompletion{
		DeviceID:     request.DeviceID,
		CompletionID: request.CompletionID,
		IoStatus:     0, // STATUS_SUCCESS
		Data:         []byte{},
	}, nil
}

// handleSetVolumeInformationRequest handles IRP_MJ_SET_VOLUME_INFORMATION requests
func (h *DeviceHandler) handleSetVolumeInformationRequest(request *device.DeviceIORequest) (*device.DeviceIOCompletion, error) {
	fmt.Printf("Set volume information request: fileID=%d\n", request.FileID)
	return &device.DeviceIOCompletion{
		DeviceID:     request.DeviceID,
		CompletionID: request.CompletionID,
		IoStatus:     0, // STATUS_SUCCESS
		Data:         []byte{},
	}, nil
}

// handleDirectoryControlRequest handles IRP_MJ_DIRECTORY_CONTROL requests
func (h *DeviceHandler) handleDirectoryControlRequest(request *device.DeviceIORequest) (*device.DeviceIOCompletion, error) {
	fmt.Printf("Directory control request: fileID=%d, minor=%d\n", request.FileID, request.MinorFunction)
	return &device.DeviceIOCompletion{
		DeviceID:     request.DeviceID,
		CompletionID: request.CompletionID,
		IoStatus:     0, // STATUS_SUCCESS
		Data:         []byte{},
	}, nil
}

// handleFileSystemControlRequest handles IRP_MJ_FILE_SYSTEM_CONTROL requests
func (h *DeviceHandler) handleFileSystemControlRequest(request *device.DeviceIORequest) (*device.DeviceIOCompletion, error) {
	fmt.Printf("File system control request: fileID=%d\n", request.FileID)
	return &device.DeviceIOCompletion{
		DeviceID:     request.DeviceID,
		CompletionID: request.CompletionID,
		IoStatus:     0, // STATUS_SUCCESS
		Data:         []byte{},
	}, nil
}

// handleDeviceControlRequest handles IRP_MJ_DEVICE_CONTROL requests
func (h *DeviceHandler) handleDeviceControlRequest(request *device.DeviceIORequest) (*device.DeviceIOCompletion, error) {
	fmt.Printf("Device control request: fileID=%d, minor=%d\n", request.FileID, request.MinorFunction)
	return &device.DeviceIOCompletion{
		DeviceID:     request.DeviceID,
		CompletionID: request.CompletionID,
		IoStatus:     0, // STATUS_SUCCESS
		Data:         []byte{},
	}, nil
}

// handleInternalDeviceControlRequest handles IRP_MJ_INTERNAL_DEVICE_CONTROL requests
func (h *DeviceHandler) handleInternalDeviceControlRequest(request *device.DeviceIORequest) (*device.DeviceIOCompletion, error) {
	fmt.Printf("Internal device control request: fileID=%d, minor=%d\n", request.FileID, request.MinorFunction)
	return &device.DeviceIOCompletion{
		DeviceID:     request.DeviceID,
		CompletionID: request.CompletionID,
		IoStatus:     0, // STATUS_SUCCESS
		Data:         []byte{},
	}, nil
}

// handleShutdownRequest handles IRP_MJ_SHUTDOWN requests
func (h *DeviceHandler) handleShutdownRequest(request *device.DeviceIORequest) (*device.DeviceIOCompletion, error) {
	fmt.Printf("Shutdown request: fileID=%d\n", request.FileID)
	return &device.DeviceIOCompletion{
		DeviceID:     request.DeviceID,
		CompletionID: request.CompletionID,
		IoStatus:     0, // STATUS_SUCCESS
		Data:         []byte{},
	}, nil
}

// handleLockControlRequest handles IRP_MJ_LOCK_CONTROL requests
func (h *DeviceHandler) handleLockControlRequest(request *device.DeviceIORequest) (*device.DeviceIOCompletion, error) {
	fmt.Printf("Lock control request: fileID=%d\n", request.FileID)
	return &device.DeviceIOCompletion{
		DeviceID:     request.DeviceID,
		CompletionID: request.CompletionID,
		IoStatus:     0, // STATUS_SUCCESS
		Data:         []byte{},
	}, nil
}

// handleCleanupRequest handles IRP_MJ_CLEANUP requests
func (h *DeviceHandler) handleCleanupRequest(request *device.DeviceIORequest) (*device.DeviceIOCompletion, error) {
	fmt.Printf("Cleanup request: fileID=%d\n", request.FileID)
	return &device.DeviceIOCompletion{
		DeviceID:     request.DeviceID,
		CompletionID: request.CompletionID,
		IoStatus:     0, // STATUS_SUCCESS
		Data:         []byte{},
	}, nil
}

// handleCreateMailslotRequest handles IRP_MJ_CREATE_MAILSLOT requests
func (h *DeviceHandler) handleCreateMailslotRequest(request *device.DeviceIORequest) (*device.DeviceIOCompletion, error) {
	fmt.Printf("Create mailslot request: fileID=%d\n", request.FileID)
	return &device.DeviceIOCompletion{
		DeviceID:     request.DeviceID,
		CompletionID: request.CompletionID,
		IoStatus:     0, // STATUS_SUCCESS
		Data:         []byte{},
	}, nil
}

// handleQuerySecurityRequest handles IRP_MJ_QUERY_SECURITY requests
func (h *DeviceHandler) handleQuerySecurityRequest(request *device.DeviceIORequest) (*device.DeviceIOCompletion, error) {
	fmt.Printf("Query security request: fileID=%d\n", request.FileID)
	return &device.DeviceIOCompletion{
		DeviceID:     request.DeviceID,
		CompletionID: request.CompletionID,
		IoStatus:     0, // STATUS_SUCCESS
		Data:         []byte{},
	}, nil
}

// handleSetSecurityRequest handles IRP_MJ_SET_SECURITY requests
func (h *DeviceHandler) handleSetSecurityRequest(request *device.DeviceIORequest) (*device.DeviceIOCompletion, error) {
	fmt.Printf("Set security request: fileID=%d\n", request.FileID)
	return &device.DeviceIOCompletion{
		DeviceID:     request.DeviceID,
		CompletionID: request.CompletionID,
		IoStatus:     0, // STATUS_SUCCESS
		Data:         []byte{},
	}, nil
}

// handlePowerRequest handles IRP_MJ_POWER requests
func (h *DeviceHandler) handlePowerRequest(request *device.DeviceIORequest) (*device.DeviceIOCompletion, error) {
	fmt.Printf("Power request: fileID=%d\n", request.FileID)
	return &device.DeviceIOCompletion{
		DeviceID:     request.DeviceID,
		CompletionID: request.CompletionID,
		IoStatus:     0, // STATUS_SUCCESS
		Data:         []byte{},
	}, nil
}

// handleSystemControlRequest handles IRP_MJ_SYSTEM_CONTROL requests
func (h *DeviceHandler) handleSystemControlRequest(request *device.DeviceIORequest) (*device.DeviceIOCompletion, error) {
	fmt.Printf("System control request: fileID=%d\n", request.FileID)
	return &device.DeviceIOCompletion{
		DeviceID:     request.DeviceID,
		CompletionID: request.CompletionID,
		IoStatus:     0, // STATUS_SUCCESS
		Data:         []byte{},
	}, nil
}

// handleDeviceChangeRequest handles IRP_MJ_DEVICE_CHANGE requests
func (h *DeviceHandler) handleDeviceChangeRequest(request *device.DeviceIORequest) (*device.DeviceIOCompletion, error) {
	fmt.Printf("Device change request: fileID=%d\n", request.FileID)
	return &device.DeviceIOCompletion{
		DeviceID:     request.DeviceID,
		CompletionID: request.CompletionID,
		IoStatus:     0, // STATUS_SUCCESS
		Data:         []byte{},
	}, nil
}

// handleQueryQuotaRequest handles IRP_MJ_QUERY_QUOTA requests
func (h *DeviceHandler) handleQueryQuotaRequest(request *device.DeviceIORequest) (*device.DeviceIOCompletion, error) {
	fmt.Printf("Query quota request: fileID=%d\n", request.FileID)
	return &device.DeviceIOCompletion{
		DeviceID:     request.DeviceID,
		CompletionID: request.CompletionID,
		IoStatus:     0, // STATUS_SUCCESS
		Data:         []byte{},
	}, nil
}

// handleSetQuotaRequest handles IRP_MJ_SET_QUOTA requests
func (h *DeviceHandler) handleSetQuotaRequest(request *device.DeviceIORequest) (*device.DeviceIOCompletion, error) {
	fmt.Printf("Set quota request: fileID=%d\n", request.FileID)
	return &device.DeviceIOCompletion{
		DeviceID:     request.DeviceID,
		CompletionID: request.CompletionID,
		IoStatus:     0, // STATUS_SUCCESS
		Data:         []byte{},
	}, nil
}

// handlePnPRequest handles IRP_MJ_PNP requests
func (h *DeviceHandler) handlePnPRequest(request *device.DeviceIORequest) (*device.DeviceIOCompletion, error) {
	fmt.Printf("PnP request: fileID=%d, minor=%d\n", request.FileID, request.MinorFunction)
	return &device.DeviceIOCompletion{
		DeviceID:     request.DeviceID,
		CompletionID: request.CompletionID,
		IoStatus:     0, // STATUS_SUCCESS
		Data:         []byte{},
	}, nil
}

// OnPrinterData is called when printer data is received
func (h *DeviceHandler) OnPrinterData(data *device.PrinterData) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if data == nil {
		return fmt.Errorf("printer data is nil")
	}

	fmt.Printf("Printer data received: jobID=%d, %d bytes, flags=0x%08X\n",
		data.JobID, len(data.Data), data.Flags)
	return nil
}

// OnDriveAccess is called when drive access occurs
func (h *DeviceHandler) OnDriveAccess(path string, operation string) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	fmt.Printf("Drive access: %s %s\n", operation, path)
	return nil
}

// OnPortAccess is called when port access occurs
func (h *DeviceHandler) OnPortAccess(portName string, operation string) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	fmt.Printf("Port access: %s %s\n", operation, portName)
	return nil
}

// GetDevices returns all announced devices
func (h *DeviceHandler) GetDevices() []*device.DeviceAnnounce {
	h.mu.RLock()
	defer h.mu.RUnlock()

	var devices []*device.DeviceAnnounce
	for _, device := range h.devices {
		devices = append(devices, device)
	}
	return devices
}

// GetDevice returns a specific device by name
func (h *DeviceHandler) GetDevice(name string) *device.DeviceAnnounce {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.devices[name]
}

// GetDeviceByID returns a specific device by ID
func (h *DeviceHandler) GetDeviceByID(deviceID uint32) *device.DeviceAnnounce {
	h.mu.RLock()
	defer h.mu.RUnlock()

	deviceName, exists := h.deviceIDs[deviceID]
	if !exists {
		return nil
	}
	return h.devices[deviceName]
}

// IsEnabled returns whether device redirection is enabled
func (h *DeviceHandler) IsEnabled() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.isEnabled
}

// SetEnabled enables or disables device redirection
func (h *DeviceHandler) SetEnabled(enabled bool) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.isEnabled = enabled
	if enabled {
		fmt.Println("Device redirection enabled")
	} else {
		fmt.Println("Device redirection disabled")
	}
}

// AnnouncePrinter announces a printer device
func (h *DeviceHandler) AnnouncePrinter(printerName string) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if !h.isEnabled {
		return fmt.Errorf("device redirection is disabled")
	}

	// Validate printer name
	if printerName == "" {
		return fmt.Errorf("printer name cannot be empty")
	}

	// Check if printer is already announced
	if _, exists := h.devices[printerName]; exists {
		return fmt.Errorf("printer %s is already announced", printerName)
	}

	client := h.manager.client
	if client == nil {
		return fmt.Errorf("RDP client is not initialized")
	}

	// Announce printer device
	err := client.AnnounceDevice(device.DeviceTypePrinter, printerName, "Printer device")
	if err != nil {
		return fmt.Errorf("failed to announce printer: %w", err)
	}

	fmt.Printf("Announced printer: %s\n", printerName)
	return nil
}

// AnnounceDrive announces a drive device
func (h *DeviceHandler) AnnounceDrive(driveLetter string, path string) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if !h.isEnabled {
		return fmt.Errorf("device redirection is disabled")
	}

	// Validate drive letter
	if driveLetter == "" || len(driveLetter) != 1 {
		return fmt.Errorf("drive letter must be a single character")
	}

	// Validate path
	if path == "" {
		return fmt.Errorf("path cannot be empty")
	}

	// Check if path exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("path does not exist: %s", path)
	}

	// Check if drive is already announced
	if _, exists := h.devices[driveLetter+":"]; exists {
		return fmt.Errorf("drive %s: is already announced", driveLetter)
	}

	client := h.manager.client
	if client == nil {
		return fmt.Errorf("RDP client is not initialized")
	}

	// Convert path to absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Announce drive device
	err = client.AnnounceDevice(device.DeviceTypeDrive, driveLetter+":", absPath)
	if err != nil {
		return fmt.Errorf("failed to announce drive: %w", err)
	}

	fmt.Printf("Announced drive: %s: -> %s\n", driveLetter, absPath)
	return nil
}

// AnnouncePort announces a port device
func (h *DeviceHandler) AnnouncePort(portName string, portType string) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if !h.isEnabled {
		return fmt.Errorf("device redirection is disabled")
	}

	// Validate port name
	if portName == "" {
		return fmt.Errorf("port name cannot be empty")
	}

	// Check if port is already announced
	if _, exists := h.devices[portName]; exists {
		return fmt.Errorf("port %s is already announced", portName)
	}

	client := h.manager.client
	if client == nil {
		return fmt.Errorf("RDP client is not initialized")
	}

	// Announce port device
	err := client.AnnounceDevice(device.DeviceTypePort, portName, portType)
	if err != nil {
		return fmt.Errorf("failed to announce port: %w", err)
	}

	fmt.Printf("Announced port: %s (%s)\n", portName, portType)
	return nil
}

// AnnounceSmartCard announces a smart card device
func (h *DeviceHandler) AnnounceSmartCard(cardName string, readerName string) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if !h.isEnabled {
		return fmt.Errorf("device redirection is disabled")
	}

	// Validate card name
	if cardName == "" {
		return fmt.Errorf("card name cannot be empty")
	}

	// Check if smart card is already announced
	if _, exists := h.devices[cardName]; exists {
		return fmt.Errorf("smart card %s is already announced", cardName)
	}

	client := h.manager.client
	if client == nil {
		return fmt.Errorf("RDP client is not initialized")
	}

	// Announce smart card device
	err := client.AnnounceDevice(device.DeviceTypeSmartCard, cardName, readerName)
	if err != nil {
		return fmt.Errorf("failed to announce smart card: %w", err)
	}

	fmt.Printf("Announced smart card: %s (%s)\n", cardName, readerName)
	return nil
}

// RemoveDevice removes a device
func (h *DeviceHandler) RemoveDevice(deviceID uint32) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Check if device exists
	deviceName, exists := h.deviceIDs[deviceID]
	if !exists {
		return fmt.Errorf("device with ID %d not found", deviceID)
	}

	client := h.manager.client
	if client == nil {
		return fmt.Errorf("RDP client is not initialized")
	}

	// Construct a device removal message (using IOCTL or close message)
	// For now, send a DEVICE_CLOSE message
	msg := &device.DeviceMessage{
		ComponentID: device.RDPDR_CTYP_CORE,
		PacketID:    uint16(device.PAKID_CORE_DEVICE_CLOSE),
		Data:        make([]byte, 4), // DeviceID
	}
	// Write DeviceID (little endian)
	msg.Data[0] = byte(deviceID)
	msg.Data[1] = byte(deviceID >> 8)
	msg.Data[2] = byte(deviceID >> 16)
	msg.Data[3] = byte(deviceID >> 24)

	err := client.SendDeviceMessage(msg)
	if err != nil {
		return fmt.Errorf("failed to remove device: %w", err)
	}

	// Remove from our tracking
	delete(h.devices, deviceName)
	delete(h.deviceIDs, deviceID)

	fmt.Printf("Removed device: %s (ID: %d)\n", deviceName, deviceID)
	return nil
}

// RemoveDeviceByName removes a device by name
func (h *DeviceHandler) RemoveDeviceByName(deviceName string) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Check if device exists
	device, exists := h.devices[deviceName]
	if !exists {
		return fmt.Errorf("device %s not found", deviceName)
	}

	return h.RemoveDevice(device.DeviceID)
}

// SendDeviceMessage sends a message to a device
func (h *DeviceHandler) SendDeviceMessage(componentID device.DeviceMessageType, packetID uint16, data []byte) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if !h.isEnabled {
		return fmt.Errorf("device redirection is disabled")
	}

	client := h.manager.client
	if client == nil {
		return fmt.Errorf("RDP client is not initialized")
	}

	msg := &device.DeviceMessage{
		ComponentID: componentID,
		PacketID:    packetID,
		Data:        data,
	}

	err := client.SendDeviceMessage(msg)
	if err != nil {
		return fmt.Errorf("failed to send device message: %w", err)
	}

	fmt.Printf("Sent device message: componentID=%d, packetID=%d, %d bytes\n", componentID, packetID, len(data))
	return nil
}

// GetDeviceStats returns statistics about device usage
func (h *DeviceHandler) GetDeviceStats() map[string]interface{} {
	h.mu.RLock()
	defer h.mu.RUnlock()

	stats := make(map[string]interface{})
	stats["total_devices"] = len(h.devices)
	stats["enabled"] = h.isEnabled

	// Count devices by type
	typeCounts := make(map[device.DeviceType]int)
	for _, device := range h.devices {
		typeCounts[device.DeviceType]++
	}

	stats["device_types"] = typeCounts
	return stats
}

// ListDeviceTypes returns all supported device types
func (h *DeviceHandler) ListDeviceTypes() []device.DeviceType {
	return []device.DeviceType{
		device.DeviceTypePrinter,
		device.DeviceTypeDrive,
		device.DeviceTypePort,
		device.DeviceTypeSmartCard,
		device.DeviceTypeAudio,
		device.DeviceTypeVideo,
		device.DeviceTypeUSB,
	}
}

// GetDeviceTypeName returns the human-readable name for a device type
func (h *DeviceHandler) GetDeviceTypeName(deviceType device.DeviceType) string {
	switch deviceType {
	case device.DeviceTypePrinter:
		return "Printer"
	case device.DeviceTypeDrive:
		return "Drive"
	case device.DeviceTypePort:
		return "Port"
	case device.DeviceTypeSmartCard:
		return "Smart Card"
	case device.DeviceTypeAudio:
		return "Audio"
	case device.DeviceTypeVideo:
		return "Video"
	case device.DeviceTypeUSB:
		return "USB"
	default:
		return "Unknown"
	}
}
