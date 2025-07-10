package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kdsmith18542/gordp/management"
)

func main() {
	fmt.Println("GoRDP Management Console Example")
	fmt.Println("=================================")

	// Parse command line flags
	var (
		port        = flag.Int("port", 8080, "Port to listen on")
		staticPath  = flag.String("static", "./management/web", "Path to static web files")
		timeout     = flag.Duration("timeout", 30, "Session timeout in minutes")
		maxSessions = flag.Int("max-sessions", 1000, "Maximum number of sessions")
	)
	flag.Parse()

	// Create console configuration
	config := &management.ConsoleConfig{
		Port:                *port,
		AdminPort:           *port + 1,
		SessionTimeout:      *timeout,
		MaxSessions:         *maxSessions,
		EnableLoadBalancing: true,
		EnableRecording:     true,
		EnableAuditLog:      true,
		DatabaseURL:         "sqlite://gordp.db",
		StaticPath:          *staticPath,
	}

	// Create and start management console
	console := management.NewManagementConsole(config)

	// Add some example servers
	addExampleServers(console)

	// Add some example users
	addExampleUsers(console)

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start console in a goroutine
	go func() {
		fmt.Printf("Starting Management Console on port %d\n", *port)
		fmt.Printf("Web interface available at: http://localhost:%d\n", *port)
		fmt.Printf("Static files path: %s\n", *staticPath)
		fmt.Println("Press Ctrl+C to stop the console")

		if err := console.Start(); err != nil {
			log.Fatalf("Failed to start console: %v", err)
		}
	}()

	// Wait for interrupt signal
	<-sigChan
	fmt.Println("\nShutting down Management Console...")

	// Stop the console
	console.Stop()

	fmt.Println("Management Console stopped successfully!")
}

// addExampleServers adds example servers to the console
func addExampleServers(console *management.ManagementConsole) {
	servers := []*management.ServerInfo{
		{
			ID:          "server-1",
			Name:        "Production Server 1",
			Address:     "192.168.1.100",
			Port:        3389,
			Status:      "online",
			Load:        0.25,
			Sessions:    5,
			MaxSessions: 50,
		},
		{
			ID:          "server-2",
			Name:        "Production Server 2",
			Address:     "192.168.1.101",
			Port:        3389,
			Status:      "online",
			Load:        0.15,
			Sessions:    3,
			MaxSessions: 50,
		},
		{
			ID:          "server-3",
			Name:        "Development Server",
			Address:     "192.168.1.102",
			Port:        3389,
			Status:      "online",
			Load:        0.05,
			Sessions:    1,
			MaxSessions: 20,
		},
		{
			ID:          "server-4",
			Name:        "Test Server",
			Address:     "192.168.1.103",
			Port:        3389,
			Status:      "offline",
			Load:        0.0,
			Sessions:    0,
			MaxSessions: 10,
		},
	}

	for _, server := range servers {
		if err := console.AddServer(server); err != nil {
			log.Printf("Failed to add server %s: %v", server.ID, err)
		} else {
			fmt.Printf("Added server: %s (%s)\n", server.Name, server.Address)
		}
	}
}

// addExampleUsers adds example users to the console
func addExampleUsers(console *management.ManagementConsole) {
	users := []*management.UserInfo{
		{
			ID:          "user-1",
			Username:    "admin",
			Email:       "admin@company.com",
			Role:        "administrator",
			Permissions: []string{"connect", "disconnect", "manage_users", "manage_servers", "view_audit"},
			LastLogin:   time.Now(),
		},
		{
			ID:          "user-2",
			Username:    "developer1",
			Email:       "dev1@company.com",
			Role:        "developer",
			Permissions: []string{"connect", "disconnect"},
			LastLogin:   time.Now().Add(-1 * time.Hour),
		},
		{
			ID:          "user-3",
			Username:    "developer2",
			Email:       "dev2@company.com",
			Role:        "developer",
			Permissions: []string{"connect", "disconnect"},
			LastLogin:   time.Now().Add(-2 * time.Hour),
		},
		{
			ID:          "user-4",
			Username:    "tester1",
			Email:       "test1@company.com",
			Role:        "tester",
			Permissions: []string{"connect", "disconnect"},
			LastLogin:   time.Now().Add(-30 * time.Minute),
		},
		{
			ID:          "user-5",
			Username:    "manager1",
			Email:       "manager1@company.com",
			Role:        "manager",
			Permissions: []string{"connect", "disconnect", "view_audit"},
			LastLogin:   time.Now().Add(-15 * time.Minute),
		},
	}

	for _, user := range users {
		if err := console.AddUser(user); err != nil {
			log.Printf("Failed to add user %s: %v", user.ID, err)
		} else {
			fmt.Printf("Added user: %s (%s)\n", user.Username, user.Role)
		}
	}
}

// Example usage:
// go run examples/management_example/management_console.go -port 8080 -static ./management/web
//
// Then open http://localhost:8080 in your web browser to access the management interface.
//
// Features:
// - Server management and monitoring
// - User management with roles and permissions
// - Session tracking and management
// - Load balancing across multiple servers
// - Session recording capabilities
// - Comprehensive audit logging
// - Real-time statistics and monitoring
// - RESTful API for integration
//
// API Endpoints:
// - GET  /api/servers     - List all servers
// - POST /api/servers     - Add a new server
// - GET  /api/users       - List all users
// - POST /api/users       - Add a new user
// - GET  /api/sessions    - List all sessions
// - POST /api/connect     - Create a new session
// - POST /api/disconnect  - Disconnect a session
// - GET  /api/stats       - Get system statistics
// - GET  /api/audit       - Get audit logs
//
// The management interface provides:
// - Dashboard with real-time statistics
// - Server management and monitoring
// - User management and role assignment
// - Session monitoring and control
// - Audit log viewing and filtering
// - Load balancing configuration
// - Recording management
// - System health monitoring
