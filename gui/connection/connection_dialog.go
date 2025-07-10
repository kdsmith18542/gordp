package connection

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// ConnectionConfig holds the connection parameters
type ConnectionConfig struct {
	Address  string
	Port     int
	Username string
	Password string
	Domain   string
}

// ConnectionDialog represents the connection settings dialog
type ConnectionDialog struct {
	// Current configuration
	config *ConnectionConfig
}

// NewConnectionDialog creates a new connection dialog
func NewConnectionDialog() *ConnectionDialog {
	dialog := &ConnectionDialog{
		config: &ConnectionConfig{
			Port: 3389,
		},
	}

	return dialog
}

// Show displays the connection dialog and returns the configuration
func (d *ConnectionDialog) Show() *ConnectionConfig {
	fmt.Println("=== Connection Settings ===")

	// Get server address
	fmt.Print("Server Address: ")
	reader := bufio.NewReader(os.Stdin)
	address, _ := reader.ReadString('\n')
	d.config.Address = strings.TrimSpace(address)

	// Get port
	fmt.Print("Port (default 3389): ")
	portStr, _ := reader.ReadString('\n')
	portStr = strings.TrimSpace(portStr)
	if portStr != "" {
		if port, err := strconv.Atoi(portStr); err == nil {
			d.config.Port = port
		}
	}

	// Get username
	fmt.Print("Username: ")
	username, _ := reader.ReadString('\n')
	d.config.Username = strings.TrimSpace(username)

	// Get password
	fmt.Print("Password: ")
	password, _ := reader.ReadString('\n')
	d.config.Password = strings.TrimSpace(password)

	// Get domain
	fmt.Print("Domain (optional): ")
	domain, _ := reader.ReadString('\n')
	d.config.Domain = strings.TrimSpace(domain)

	// Validate required fields
	if d.config.Address == "" || d.config.Username == "" || d.config.Password == "" {
		fmt.Println("Error: Address, username, and password are required")
		return nil
	}

	fmt.Printf("Connecting to %s:%d as %s\n", d.config.Address, d.config.Port, d.config.Username)
	return d.config
}

// GetConnectionConfig returns the current connection configuration
func (d *ConnectionDialog) GetConnectionConfig() *ConnectionConfig {
	return d.config
}

// SetConnectionConfig sets the connection configuration
func (d *ConnectionDialog) SetConnectionConfig(config *ConnectionConfig) {
	if config != nil {
		d.config = config
	}
}
