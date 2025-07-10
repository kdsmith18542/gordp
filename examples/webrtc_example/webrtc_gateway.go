package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/kdsmith18542/gordp/gateway/webrtc"
)

func main() {
	fmt.Println("GoRDP WebRTC Gateway Example")
	fmt.Println("============================")

	// Parse command line flags
	var (
		port       = flag.Int("port", 8080, "Port to listen on")
		staticPath = flag.String("static", "./gateway/webrtc/web", "Path to static web files")
		timeout    = flag.Duration("timeout", 30, "Session timeout in minutes")
	)
	flag.Parse()

	// Create gateway configuration
	config := &webrtc.GatewayConfig{
		Port:           *port,
		MaxConnections: 100,
		SessionTimeout: *timeout,
		EnableCORS:     true,
		StaticPath:     *staticPath,
	}

	// Create and start gateway
	gateway := webrtc.NewWebRTCGateway(config)

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start gateway in a goroutine
	go func() {
		fmt.Printf("Starting WebRTC Gateway on port %d\n", *port)
		fmt.Printf("Web interface available at: http://localhost:%d\n", *port)
		fmt.Printf("Static files path: %s\n", *staticPath)
		fmt.Println("Press Ctrl+C to stop the gateway")

		if err := gateway.Start(); err != nil {
			log.Fatalf("Failed to start gateway: %v", err)
		}
	}()

	// Wait for interrupt signal
	<-sigChan
	fmt.Println("\nShutting down WebRTC Gateway...")

	// Stop the gateway
	gateway.Stop()

	fmt.Println("WebRTC Gateway stopped successfully!")
}

// Example usage:
// go run examples/webrtc_example/webrtc_gateway.go -port 8080 -static ./gateway/webrtc/web
//
// Then open http://localhost:8080 in your web browser to access the RDP client.
//
// Features:
// - Web-based RDP client interface
// - Session management
// - Real-time status updates
// - Multiple concurrent sessions
// - Automatic session cleanup
// - RESTful API endpoints
//
// API Endpoints:
// - GET  /api/sessions     - List all active sessions
// - POST /api/connect      - Create a new RDP connection
// - POST /api/disconnect   - Disconnect a session
// - GET  /api/status       - Get session status
//
// The web interface provides:
// - Connection form for RDP server details
// - Real-time session status
// - Session management
// - Connection logs
// - Fullscreen and screenshot capabilities
