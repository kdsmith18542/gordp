package integration

import (
	"context"
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

// MockRDPServer represents a mock RDP server for integration testing
type MockRDPServer struct {
	listener    net.Listener
	addr        string
	clients     map[string]*MockRDPClient
	clientsMux  sync.RWMutex
	ctx         context.Context
	cancel      context.CancelFunc
	config      *MockServerConfig
	bitmapData  []byte
	eventLog    []ServerEvent
	eventLogMux sync.RWMutex
}

// MockServerConfig contains configuration for the mock server
type MockServerConfig struct {
	Port            int
	MaxConnections  int
	ResponseDelay   time.Duration
	EnableSecurity  bool
	EnableBitmap    bool
	EnableAudio     bool
	EnableClipboard bool
	EnableDevice    bool
	ScreenWidth     int
	ScreenHeight    int
	ColorDepth      int
}

// DefaultMockServerConfig returns default mock server configuration
func DefaultMockServerConfig() *MockServerConfig {
	return &MockServerConfig{
		Port:            3389,
		MaxConnections:  10,
		ResponseDelay:   10 * time.Millisecond,
		EnableSecurity:  true,
		EnableBitmap:    true,
		EnableAudio:     true,
		EnableClipboard: true,
		EnableDevice:    true,
		ScreenWidth:     1024,
		ScreenHeight:    768,
		ColorDepth:      24,
	}
}

// ServerEvent represents an event that occurred on the server
type ServerEvent struct {
	Type      string
	Timestamp time.Time
	Data      interface{}
}

// MockRDPClient represents a client connected to the mock server
type MockRDPClient struct {
	conn       net.Conn
	server     *MockRDPServer
	clientID   string
	connected  bool
	ctx        context.Context
	cancel     context.CancelFunc
	bitmapData []byte
	events     []ClientEvent
	eventsMux  sync.RWMutex
}

// ClientEvent represents an event from a client
type ClientEvent struct {
	Type      string
	Timestamp time.Time
	Data      interface{}
}

// NewMockRDPServer creates a new mock RDP server
func NewMockRDPServer(cfg *MockServerConfig) *MockRDPServer {
	if cfg == nil {
		cfg = DefaultMockServerConfig()
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &MockRDPServer{
		addr:       fmt.Sprintf(":%d", cfg.Port),
		clients:    make(map[string]*MockRDPClient),
		ctx:        ctx,
		cancel:     cancel,
		config:     cfg,
		bitmapData: generateTestBitmap(cfg.ScreenWidth, cfg.ScreenHeight, cfg.ColorDepth),
		eventLog:   make([]ServerEvent, 0),
	}
}

// Start starts the mock RDP server
func (s *MockRDPServer) Start() error {
	var err error
	s.listener, err = net.Listen("tcp", s.addr)
	if err != nil {
		return fmt.Errorf("failed to start mock server: %w", err)
	}

	s.logEvent("server_started", map[string]interface{}{
		"addr":   s.addr,
		"config": s.config,
	})

	go s.acceptLoop()
	return nil
}

// Stop stops the mock RDP server
func (s *MockRDPServer) Stop() {
	s.cancel()

	if s.listener != nil {
		s.listener.Close()
	}

	// Close all client connections
	s.clientsMux.Lock()
	for _, client := range s.clients {
		client.Close()
	}
	s.clientsMux.Unlock()

	s.logEvent("server_stopped", nil)
}

// GetAddr returns the server address
func (s *MockRDPServer) GetAddr() string {
	return s.addr
}

// GetConnectedClients returns the number of connected clients
func (s *MockRDPServer) GetConnectedClients() int {
	s.clientsMux.RLock()
	defer s.clientsMux.RUnlock()
	return len(s.clients)
}

// GetEvents returns all server events
func (s *MockRDPServer) GetEvents() []ServerEvent {
	s.eventLogMux.RLock()
	defer s.eventLogMux.RUnlock()
	return append([]ServerEvent{}, s.eventLog...)
}

// acceptLoop accepts incoming connections
func (s *MockRDPServer) acceptLoop() {
	for {
		select {
		case <-s.ctx.Done():
			return
		default:
			conn, err := s.listener.Accept()
			if err != nil {
				if s.ctx.Err() != nil {
					return // Server is shutting down
				}
				s.logEvent("accept_error", map[string]interface{}{
					"error": err.Error(),
				})
				continue
			}

			// Check connection limit
			if s.GetConnectedClients() >= s.config.MaxConnections {
				conn.Close()
				s.logEvent("connection_rejected", map[string]interface{}{
					"reason": "max_connections_reached",
					"client": conn.RemoteAddr().String(),
				})
				continue
			}

			// Handle connection in goroutine
			go s.handleConnection(conn)
		}
	}
}

// handleConnection handles a new client connection
func (s *MockRDPServer) handleConnection(conn net.Conn) {
	clientID := fmt.Sprintf("%s-%d", conn.RemoteAddr().String(), time.Now().UnixNano())
	ctx, cancel := context.WithCancel(s.ctx)

	client := &MockRDPClient{
		conn:      conn,
		server:    s,
		clientID:  clientID,
		connected: false,
		ctx:       ctx,
		cancel:    cancel,
		events:    make([]ClientEvent, 0),
	}

	// Add client to server
	s.clientsMux.Lock()
	s.clients[clientID] = client
	s.clientsMux.Unlock()

	s.logEvent("client_connected", map[string]interface{}{
		"client_id": clientID,
		"addr":      conn.RemoteAddr().String(),
	})

	// Handle the connection
	s.handleSimpleProtocol(client)

	// Remove client from server
	s.clientsMux.Lock()
	delete(s.clients, clientID)
	s.clientsMux.Unlock()

	s.logEvent("client_disconnected", map[string]interface{}{
		"client_id": clientID,
	})
}

// handleSimpleProtocol handles a simplified protocol that just responds to keep the client happy
func (s *MockRDPServer) handleSimpleProtocol(client *MockRDPClient) {
	defer client.Close()

	// Set a timeout for the connection
	client.conn.SetReadDeadline(time.Now().Add(30 * time.Second))

	// Mark as connected immediately since we're a mock server
	client.connected = true
	s.logEvent("client_authenticated", map[string]interface{}{
		"client_id": client.clientID,
	})

	// Handle RDP handshake
	if err := s.handleRDPHandshake(client); err != nil {
		client.logEvent("handshake_error", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	// Handle MCS Connect Initial
	if err := s.handleMCSConnectInitial(client); err != nil {
		client.logEvent("mcs_connect_error", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	// Simple echo server that responds to any data
	buffer := make([]byte, 1024)
	for {
		select {
		case <-client.ctx.Done():
			return
		default:
			// Set a shorter read timeout for each read
			client.conn.SetReadDeadline(time.Now().Add(5 * time.Second))

			n, err := client.conn.Read(buffer)
			if err != nil {
				if err == io.EOF {
					client.logEvent("client_disconnected", nil)
				} else {
					// Check if it's a timeout error
					if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
						// Send a keep-alive response on timeout
						keepAlive := s.createKeepAlivePacket()
						_, writeErr := client.conn.Write(keepAlive)
						if writeErr != nil {
							client.logEvent("keepalive_write_error", map[string]interface{}{
								"error": writeErr.Error(),
							})
							return
						}
						client.logEvent("keepalive_sent", nil)
						continue
					}

					client.logEvent("read_error", map[string]interface{}{
						"error": err.Error(),
					})
				}
				return
			}

			if n > 0 {
				client.logEvent("data_received", map[string]interface{}{
					"data_size": n,
					"data":      buffer[:n],
				})

				// Simulate response delay
				time.Sleep(s.config.ResponseDelay)

				// Send a valid TPKT packet
				response := s.createDataPacket()
				_, err = client.conn.Write(response)
				if err != nil {
					client.logEvent("write_error", map[string]interface{}{
						"error": err.Error(),
					})
					return
				}

				client.logEvent("response_sent", map[string]interface{}{
					"response_size": len(response),
				})
			}
		}
	}
}

// handleRDPHandshake handles the RDP connection handshake
func (s *MockRDPServer) handleRDPHandshake(client *MockRDPClient) error {
	// Read the client connection request
	buffer := make([]byte, 1024)
	n, err := client.conn.Read(buffer)
	if err != nil {
		return fmt.Errorf("failed to read connection request: %w", err)
	}

	client.logEvent("connection_request_received", map[string]interface{}{
		"data_size": n,
		"data":      buffer[:n],
	})

	// Parse TPKT header
	if n < 4 {
		return fmt.Errorf("invalid TPKT header length: %d", n)
	}

	// TPKT header: version(1) + reserved(1) + length(2)
	version := buffer[0]
	reserved := buffer[1]
	length := uint16(buffer[2])<<8 | uint16(buffer[3])

	client.logEvent("tpkt_header_parsed", map[string]interface{}{
		"version":  version,
		"reserved": reserved,
		"length":   length,
	})

	// Validate TPKT header
	if version != 3 {
		return fmt.Errorf("invalid TPKT version: %d", version)
	}

	// Parse X.224 header (starts at byte 4)
	if n < 7 {
		return fmt.Errorf("invalid X.224 header length: %d", n-4)
	}

	x224Length := buffer[4]
	x224Type := buffer[5]
	x224DstRef := uint16(buffer[6])<<8 | uint16(buffer[7])
	x224SrcRef := uint16(buffer[8])<<8 | uint16(buffer[9])
	x224Flags := buffer[10]

	client.logEvent("x224_header_parsed", map[string]interface{}{
		"length":  x224Length,
		"type":    x224Type,
		"dst_ref": x224DstRef,
		"src_ref": x224SrcRef,
		"flags":   x224Flags,
	})

	// Validate X.224 header - should be connection request (0xE0)
	if x224Type != 0xE0 {
		return fmt.Errorf("invalid X.224 type: %x, expected 0xE0", x224Type)
	}

	// Simulate response delay
	time.Sleep(s.config.ResponseDelay)

	// Send server connection confirm
	confirmData := s.createConnectionConfirm()
	_, err = client.conn.Write(confirmData)
	if err != nil {
		return fmt.Errorf("failed to send connection confirm: %w", err)
	}

	client.logEvent("connection_confirm_sent", map[string]interface{}{
		"response_size": len(confirmData),
	})

	return nil
}

// createConnectionConfirm creates a valid RDP server connection confirm packet
func (s *MockRDPServer) createConnectionConfirm() []byte {
	// RDP Negotiation Response
	// Type: 0x02 (TYPE_RDP_NEG_RSP)
	// Flags: 0x00
	// Length: 0x08 0x00 (8 bytes)
	// Result: 0x00000000 (PROTOCOL_RDP)
	negotiation := []byte{
		0x02,       // Type: TYPE_RDP_NEG_RSP
		0x00,       // Flags
		0x08, 0x00, // Length: 8 bytes
		0x00, 0x00, 0x00, 0x00, // Result: PROTOCOL_RDP
	}

	// X.224 Connection Confirm header
	// Length: 6 + len(negotiation) = 14
	// Type: 0xD0 (TPDU_CONNECTION_CONFIRM)
	// DstRef: 0x0000
	// SrcRef: 0x0000
	// Flags: 0x00
	x224Header := []byte{
		14,         // Length: 6 + 8 = 14
		0xD0,       // Type: TPDU_CONNECTION_CONFIRM
		0x00, 0x00, // DstRef
		0x00, 0x00, // SrcRef
		0x00, // Flags
	}

	// TPKT header
	// Version: 3
	// Reserved: 0
	// Length: 4 + len(x224Header) + len(negotiation) = 4 + 6 + 8 = 18
	totalLength := 4 + len(x224Header) + len(negotiation)
	tpktHeader := []byte{
		0x03,                     // Version: 3
		0x00,                     // Reserved: 0
		byte(totalLength >> 8),   // Length high byte
		byte(totalLength & 0xFF), // Length low byte
	}

	// Combine all parts
	result := make([]byte, 0, totalLength)
	result = append(result, tpktHeader...)
	result = append(result, x224Header...)
	result = append(result, negotiation...)

	return result
}

// createDataPacket creates a valid TPKT data packet
func (s *MockRDPServer) createDataPacket() []byte {
	// Simple data payload
	data := []byte{0x02, 0xF0, 0x80}

	// X.224 Data header
	// Length: 3 + len(data) = 6
	// Type: 0xF0 (TPDU_DATA)
	// Flags: 0x80
	x224Header := []byte{
		6,    // Length: 3 + 3 = 6
		0xF0, // Type: TPDU_DATA
		0x80, // Flags
	}

	// TPKT header
	// Version: 3
	// Reserved: 0
	// Length: 4 + len(x224Header) + len(data) = 4 + 3 + 3 = 10
	totalLength := 4 + len(x224Header) + len(data)
	tpktHeader := []byte{
		0x03,                     // Version: 3
		0x00,                     // Reserved: 0
		byte(totalLength >> 8),   // Length high byte
		byte(totalLength & 0xFF), // Length low byte
	}

	// Combine all parts
	result := make([]byte, 0, totalLength)
	result = append(result, tpktHeader...)
	result = append(result, x224Header...)
	result = append(result, data...)

	return result
}

// createKeepAlivePacket creates a keep-alive packet
func (s *MockRDPServer) createKeepAlivePacket() []byte {
	// Same as data packet for simplicity
	return s.createDataPacket()
}

// handleMCSConnectInitial handles the MCS Connect Initial PDU
func (s *MockRDPServer) handleMCSConnectInitial(client *MockRDPClient) error {
	// Read the MCS Connect Initial PDU
	buffer := make([]byte, 4096) // Larger buffer for MCS data
	n, err := client.conn.Read(buffer)
	if err != nil {
		return fmt.Errorf("failed to read MCS Connect Initial: %w", err)
	}

	client.logEvent("mcs_connect_initial_received", map[string]interface{}{
		"data_size": n,
		"data":      buffer[:n],
	})

	// Parse TPKT header
	if n < 4 {
		return fmt.Errorf("invalid TPKT header length: %d", n)
	}

	// TPKT header: version(1) + reserved(1) + length(2)
	version := buffer[0]
	reserved := buffer[1]
	length := uint16(buffer[2])<<8 | uint16(buffer[3])

	client.logEvent("mcs_tpkt_header_parsed", map[string]interface{}{
		"version":  version,
		"reserved": reserved,
		"length":   length,
	})

	// Validate TPKT header
	if version != 3 {
		return fmt.Errorf("invalid TPKT version: %d", version)
	}

	// Parse X.224 header (starts at byte 4)
	if n < 7 {
		return fmt.Errorf("invalid X.224 header length: %d", n-4)
	}

	x224Length := buffer[4]
	x224Type := buffer[5]
	x224Flags := buffer[6]

	client.logEvent("mcs_x224_header_parsed", map[string]interface{}{
		"length": x224Length,
		"type":   x224Type,
		"flags":  x224Flags,
	})

	// Validate X.224 header - should be data (0xF0)
	if x224Type != 0xF0 {
		return fmt.Errorf("invalid X.224 type: %x, expected 0xF0", x224Type)
	}

	// Simulate response delay
	time.Sleep(s.config.ResponseDelay)

	// Send MCS Connect Response
	responseData := s.createMCSConnectResponse()
	_, err = client.conn.Write(responseData)
	if err != nil {
		return fmt.Errorf("failed to send MCS Connect Response: %w", err)
	}

	client.logEvent("mcs_connect_response_sent", map[string]interface{}{
		"response_size": len(responseData),
	})

	return nil
}

// createMCSConnectResponse creates a valid MCS Connect Response PDU
func (s *MockRDPServer) createMCSConnectResponse() []byte {
	// MCS Connect Response structure:
	// - MCS Connect Response header
	// - GCC Conference Create Response
	// - Server Core Data
	// - Server Security Data
	// - Server Network Data

	// Server Core Data (simplified)
	serverCoreData := []byte{
		0x01, 0x00, 0x00, 0x00, // Version: RDP 8.1
		0x00, 0x00, 0x00, 0x00, // Client Requested Protocol
		0x00, 0x00, 0x00, 0x00, // Early Capability Flags
	}

	// Server Security Data (simplified)
	serverSecurityData := []byte{
		0x00, 0x00, 0x00, 0x00, // Encryption Method
		0x00, 0x00, 0x00, 0x00, // Encryption Level
		0x00, 0x00, 0x00, 0x00, // Server Random
		0x00, 0x00, 0x00, 0x00, // Server Certificate
	}

	// Server Network Data (simplified)
	serverNetworkData := []byte{
		0x00, 0x00, 0x00, 0x00, // MCS Channel Count
		0x00, 0x00, 0x00, 0x00, // Channel ID Array
	}

	// GCC Conference Create Response
	gccResponse := []byte{
		0x00, 0x00, 0x00, 0x00, // Result: success
		0x00, 0x00, 0x00, 0x00, // Conference ID
	}

	// Combine user data
	userData := make([]byte, 0)
	userData = append(userData, gccResponse...)
	userData = append(userData, serverCoreData...)
	userData = append(userData, serverSecurityData...)
	userData = append(userData, serverNetworkData...)

	// MCS Connect Response header
	mcsHeader := []byte{
		0x02, 0x00, 0x00, 0x00, // Type: Connect Response
		0x00, 0x00, 0x00, 0x00, // Called Connect ID
		0x00, 0x00, 0x00, 0x00, // Result: successful
	}

	// Combine MCS data
	mcsData := make([]byte, 0)
	mcsData = append(mcsData, mcsHeader...)
	mcsData = append(mcsData, userData...)

	// X.224 Data header
	x224Header := []byte{
		byte(3 + len(mcsData)), // Length: 3 + len(mcsData)
		0xF0,                   // Type: TPDU_DATA
		0x80,                   // Flags
	}

	// TPKT header
	totalLength := 4 + len(x224Header) + len(mcsData)
	tpktHeader := []byte{
		0x03,                     // Version: 3
		0x00,                     // Reserved: 0
		byte(totalLength >> 8),   // Length high byte
		byte(totalLength & 0xFF), // Length low byte
	}

	// Combine all parts
	result := make([]byte, 0, totalLength)
	result = append(result, tpktHeader...)
	result = append(result, x224Header...)
	result = append(result, mcsData...)

	return result
}

// Close closes the client connection
func (c *MockRDPClient) Close() {
	c.connected = false
	c.cancel()
	if c.conn != nil {
		c.conn.Close()
	}
}

// logEvent logs a client event
func (c *MockRDPClient) logEvent(eventType string, data interface{}) {
	c.eventsMux.Lock()
	defer c.eventsMux.Unlock()

	c.events = append(c.events, ClientEvent{
		Type:      eventType,
		Timestamp: time.Now(),
		Data:      data,
	})
}

// GetEvents returns all client events
func (c *MockRDPClient) GetEvents() []ClientEvent {
	c.eventsMux.RLock()
	defer c.eventsMux.RUnlock()
	return append([]ClientEvent{}, c.events...)
}

// logEvent logs a server event
func (s *MockRDPServer) logEvent(eventType string, data interface{}) {
	s.eventLogMux.Lock()
	defer s.eventLogMux.Unlock()

	s.eventLog = append(s.eventLog, ServerEvent{
		Type:      eventType,
		Timestamp: time.Now(),
		Data:      data,
	})
}

// generateTestBitmap generates test bitmap data
func generateTestBitmap(width, height, colorDepth int) []byte {
	// Create a simple test bitmap
	bytesPerPixel := colorDepth / 8
	size := width * height * bytesPerPixel
	data := make([]byte, size)

	// Fill with a gradient pattern
	for i := 0; i < size; i += bytesPerPixel {
		x := (i / bytesPerPixel) % width
		y := (i / bytesPerPixel) / width

		// Create a simple gradient
		r := byte((x * 255) / width)
		g := byte((y * 255) / height)
		b := byte(128)

		if bytesPerPixel >= 3 {
			data[i] = r
			data[i+1] = g
			data[i+2] = b
		}
	}

	return data
}
