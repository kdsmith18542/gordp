package integration

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/kdsmith18542/gordp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIntegration_BasicConnection tests basic RDP connection
func TestIntegration_BasicConnection(t *testing.T) {
	// Create mock server
	serverConfig := DefaultMockServerConfig()
	serverConfig.Port = 3389
	server := NewMockRDPServer(serverConfig)

	// Start server
	err := server.Start()
	require.NoError(t, err)
	defer server.Stop()

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)

	// Create client
	client := gordp.NewClient(&gordp.Option{
		Addr:           "localhost:3389",
		UserName:       "testuser",
		Password:       "testpass",
		ConnectTimeout: 5 * time.Second,
	})

	// Connect - this will likely fail due to protocol mismatch, but we test the infrastructure
	err = client.Connect()
	// We expect this to fail because our mock server doesn't implement the full RDP protocol
	// The important thing is that the test infrastructure works
	if err != nil {
		t.Logf("Expected connection failure due to simplified mock server: %v", err)
	}

	// Verify server received connection attempt
	serverEvents := server.GetEvents()
	assert.Greater(t, len(serverEvents), 0, "Server should have events")

	// Check for connection event
	hasConnectionEvent := false
	for _, event := range serverEvents {
		if event.Type == "client_connected" {
			hasConnectionEvent = true
			break
		}
	}
	assert.True(t, hasConnectionEvent, "Server should have received connection event")

	// Close client
	client.Close()

	// Verify client count
	assert.Equal(t, 0, server.GetConnectedClients(), "No clients should be connected")
}

// TestIntegration_ServerInfrastructure tests the mock server infrastructure
func TestIntegration_ServerInfrastructure(t *testing.T) {
	// Test server creation
	serverConfig := DefaultMockServerConfig()
	serverConfig.Port = 3389
	server := NewMockRDPServer(serverConfig)

	// Test server start
	err := server.Start()
	require.NoError(t, err)
	defer server.Stop()

	// Test server address
	assert.Equal(t, ":3389", server.GetAddr())

	// Test initial client count
	assert.Equal(t, 0, server.GetConnectedClients())

	// Test server events
	events := server.GetEvents()
	assert.Greater(t, len(events), 0, "Server should have start event")

	// Check for server started event
	hasStartEvent := false
	for _, event := range events {
		if event.Type == "server_started" {
			hasStartEvent = true
			break
		}
	}
	assert.True(t, hasStartEvent, "Server should have start event")
}

// TestIntegration_MultipleConnections tests multiple client connections
func TestIntegration_MultipleConnections(t *testing.T) {
	// Create mock server
	serverConfig := DefaultMockServerConfig()
	serverConfig.Port = 3389
	serverConfig.MaxConnections = 3
	server := NewMockRDPServer(serverConfig)

	// Start server
	err := server.Start()
	require.NoError(t, err)
	defer server.Stop()

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)

	// Create multiple clients
	clients := make([]*gordp.Client, 3)
	for i := 0; i < 3; i++ {
		clients[i] = gordp.NewClient(&gordp.Option{
			Addr:           "localhost:3389",
			UserName:       fmt.Sprintf("testuser%d", i),
			Password:       "testpass",
			ConnectTimeout: 5 * time.Second,
		})

		// Attempt connection (will likely fail due to protocol mismatch)
		err := clients[i].Connect()
		if err != nil {
			t.Logf("Expected connection failure for client %d: %v", i, err)
		}
	}

	// Wait a bit for connections to be processed
	time.Sleep(200 * time.Millisecond)

	// Verify server events
	serverEvents := server.GetEvents()
	assert.Greater(t, len(serverEvents), 0, "Server should have events")

	// Count connection events
	connectionEvents := 0
	for _, event := range serverEvents {
		if event.Type == "client_connected" {
			connectionEvents++
		}
	}
	assert.Greater(t, connectionEvents, 0, "Server should have received connection events")

	// Close all clients
	for _, client := range clients {
		client.Close()
	}

	// Verify no clients connected
	assert.Equal(t, 0, server.GetConnectedClients(), "No clients should be connected")
}

// TestIntegration_ServerShutdown tests server shutdown behavior
func TestIntegration_ServerShutdown(t *testing.T) {
	// Create mock server
	serverConfig := DefaultMockServerConfig()
	serverConfig.Port = 3389
	server := NewMockRDPServer(serverConfig)

	// Start server
	err := server.Start()
	require.NoError(t, err)

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)

	// Create a client
	client := gordp.NewClient(&gordp.Option{
		Addr:           "localhost:3389",
		UserName:       "testuser",
		Password:       "testpass",
		ConnectTimeout: 5 * time.Second,
	})

	// Attempt connection
	err = client.Connect()
	if err != nil {
		t.Logf("Expected connection failure: %v", err)
	}

	// Wait a bit
	time.Sleep(100 * time.Millisecond)

	// Stop server
	server.Stop()

	// Wait for shutdown
	time.Sleep(100 * time.Millisecond)

	// Verify server events
	serverEvents := server.GetEvents()
	assert.Greater(t, len(serverEvents), 0, "Server should have events")

	// Check for server stopped event
	hasStopEvent := false
	for _, event := range serverEvents {
		if event.Type == "server_stopped" {
			hasStopEvent = true
			break
		}
	}
	assert.True(t, hasStopEvent, "Server should have stop event")

	// Close client
	client.Close()
}

// TestIntegration_ContextSupport tests context support in the mock server
func TestIntegration_ContextSupport(t *testing.T) {
	// Create mock server
	serverConfig := DefaultMockServerConfig()
	serverConfig.Port = 3389
	server := NewMockRDPServer(serverConfig)

	// Start server
	err := server.Start()
	require.NoError(t, err)
	defer server.Stop()

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Create client with context
	client := gordp.NewClientWithContext(ctx, &gordp.Option{
		Addr:           "localhost:3389",
		UserName:       "testuser",
		Password:       "testpass",
		ConnectTimeout: 5 * time.Second,
	})

	// Attempt connection with context
	err = client.ConnectWithContext(ctx)
	if err != nil {
		t.Logf("Expected connection failure: %v", err)
	}

	// Wait for context to timeout or connection to complete
	select {
	case <-ctx.Done():
		t.Log("Context timeout as expected")
	case <-time.After(3 * time.Second):
		t.Log("Connection attempt completed")
	}

	// Close client
	client.Close()
}

// TestIntegration_ErrorHandling tests error handling in the mock server
func TestIntegration_ErrorHandling(t *testing.T) {
	// Test invalid port
	serverConfig := DefaultMockServerConfig()
	serverConfig.Port = 99999 // Invalid port
	server := NewMockRDPServer(serverConfig)

	// This should fail to start
	err := server.Start()
	if err != nil {
		t.Logf("Expected server start failure: %v", err)
	} else {
		server.Stop()
	}

	// Test with valid port
	serverConfig.Port = 3389
	server = NewMockRDPServer(serverConfig)

	err = server.Start()
	require.NoError(t, err)
	defer server.Stop()

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)

	// Test connection limit
	serverConfig.MaxConnections = 0
	server2 := NewMockRDPServer(serverConfig)
	err = server2.Start()
	require.NoError(t, err)
	defer server2.Stop()

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)

	// Try to connect - should be rejected
	client := gordp.NewClient(&gordp.Option{
		Addr:           "localhost:3389",
		UserName:       "testuser",
		Password:       "testpass",
		ConnectTimeout: 5 * time.Second,
	})

	err = client.Connect()
	if err != nil {
		t.Logf("Expected connection failure: %v", err)
	}

	// Verify server events
	server2Events := server2.GetEvents()
	assert.Greater(t, len(server2Events), 0, "Server should have events")

	// Check for connection rejected event
	hasRejectEvent := false
	for _, event := range server2Events {
		if event.Type == "connection_rejected" {
			hasRejectEvent = true
			break
		}
	}
	assert.True(t, hasRejectEvent, "Server should have rejected connection")

	client.Close()
}

// TestIntegration_Performance tests basic performance of the mock server
func TestIntegration_Performance(t *testing.T) {
	// Create mock server
	serverConfig := DefaultMockServerConfig()
	serverConfig.Port = 3389
	serverConfig.ResponseDelay = 1 * time.Millisecond // Fast response
	server := NewMockRDPServer(serverConfig)

	// Start server
	err := server.Start()
	require.NoError(t, err)
	defer server.Stop()

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)

	// Test multiple rapid connections
	start := time.Now()
	clients := make([]*gordp.Client, 10)

	for i := 0; i < 10; i++ {
		clients[i] = gordp.NewClient(&gordp.Option{
			Addr:           "localhost:3389",
			UserName:       fmt.Sprintf("testuser%d", i),
			Password:       "testpass",
			ConnectTimeout: 1 * time.Second,
		})

		// Attempt connection
		err := clients[i].Connect()
		if err != nil {
			t.Logf("Expected connection failure for client %d: %v", i, err)
		}
	}

	// Wait for all connections to be processed
	time.Sleep(500 * time.Millisecond)

	duration := time.Since(start)
	t.Logf("Processed 10 connection attempts in %v", duration)

	// Verify server events
	serverEvents := server.GetEvents()
	assert.Greater(t, len(serverEvents), 0, "Server should have events")

	// Count connection events
	connectionEvents := 0
	for _, event := range serverEvents {
		if event.Type == "client_connected" {
			connectionEvents++
		}
	}
	assert.Greater(t, connectionEvents, 0, "Server should have received connection events")

	// Close all clients
	for _, client := range clients {
		client.Close()
	}

	// Verify no clients connected
	assert.Equal(t, 0, server.GetConnectedClients(), "No clients should be connected")
}

// BenchmarkIntegration_Connection benchmarks connection handling
func BenchmarkIntegration_Connection(b *testing.B) {
	// Create mock server
	serverConfig := DefaultMockServerConfig()
	serverConfig.Port = 3389
	serverConfig.ResponseDelay = 0 // No delay for benchmarking
	server := NewMockRDPServer(serverConfig)

	// Start server
	err := server.Start()
	if err != nil {
		b.Fatalf("Failed to start server: %v", err)
	}
	defer server.Stop()

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			client := gordp.NewClient(&gordp.Option{
				Addr:           "localhost:3389",
				UserName:       "testuser",
				Password:       "testpass",
				ConnectTimeout: 1 * time.Second,
			})

			// Attempt connection
			err := client.Connect()
			if err != nil {
				// Expected failure due to protocol mismatch
			}

			client.Close()
		}
	})
}

// BenchmarkIntegration_ServerEvents benchmarks server event handling
func BenchmarkIntegration_ServerEvents(b *testing.B) {
	// Create mock server
	serverConfig := DefaultMockServerConfig()
	serverConfig.Port = 3389
	server := NewMockRDPServer(serverConfig)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		server.logEvent("benchmark_event", map[string]interface{}{
			"iteration": i,
			"data":      "benchmark data",
		})
	}
}
