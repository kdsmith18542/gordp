package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/kdsmith18542/gordp/config"
	"github.com/kdsmith18542/gordp/di"
	"github.com/kdsmith18542/gordp/proto/bitmap"
)

func main() {
	fmt.Println("GoRDP Dependency Injection Example")
	fmt.Println("==================================")

	// Parse command line flags
	var (
		host     = flag.String("host", "localhost", "RDP server host")
		port     = flag.Int("port", 3389, "RDP server port")
		username = flag.String("username", "administrator", "Username")
		password = flag.String("password", "", "Password")
	)
	flag.Parse()

	if *password == "" {
		log.Fatal("Password is required")
	}

	// Create application with dependency injection
	app := di.NewApplication()
	if err := app.Initialize(); err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
	}
	defer app.Close()

	// Get RDP client factory from DI container
	var clientFactory *di.RDPClientFactory
	if err := app.GetTyped("rdp_client_factory", &clientFactory); err != nil {
		log.Fatalf("Failed to get RDP client factory: %v", err)
	}

	// Create configuration
	cfg := &config.Config{
		Connection: config.ConnectionConfig{
			Address: *host,
			Port:    *port,
		},
		Authentication: config.AuthConfig{
			Username: *username,
			Password: *password,
		},
		Display: config.DisplayConfig{
			Width:      1024,
			Height:     768,
			ColorDepth: 24,
		},
		Performance: config.PerformanceConfig{
			BitmapCache: true,
			Compression: true,
		},
	}

	// Create RDP client using factory
	client, err := clientFactory.CreateWithConfig(cfg)
	if err != nil {
		log.Fatalf("Failed to create RDP client: %v", err)
	}

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Connect to RDP server
	fmt.Printf("Connecting to %s:%d...\n", *host, *port)
	if err := client.Connect(); err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer client.Close()

	fmt.Println("Connected successfully!")

	// Create bitmap processor
	processor := &ExampleBitmapProcessor{}

	// Start RDP session
	fmt.Println("Starting RDP session...")
	if err := client.Run(processor); err != nil {
		log.Fatalf("RDP session failed: %v", err)
	}

	// Wait for interrupt signal
	<-sigChan
	fmt.Println("\nShutting down...")
}

// ExampleBitmapProcessor processes bitmap data
type ExampleBitmapProcessor struct{}

// ProcessBitmap processes bitmap data
func (p *ExampleBitmapProcessor) ProcessBitmap(option *bitmap.Option, bitmap *bitmap.BitMap) {
	fmt.Printf("Received bitmap: %dx%d at (%d,%d)\n",
		option.Width, option.Height, option.Left, option.Top)
}

// Example of using DI with custom modules
func exampleWithCustomModules() {
	// Create application
	app := di.NewApplication()

	// Register custom module
	app.RegisterModule(di.NewModule("custom", &CustomModule{}))

	// Initialize
	if err := app.Initialize(); err != nil {
		log.Fatal(err)
	}
	defer app.Close()

	// Get custom service
	var customService *CustomService
	if err := app.GetTyped("custom_service", &customService); err != nil {
		log.Fatal(err)
	}

	// Use custom service
	customService.DoSomething()
}

// CustomModule provides custom services
type CustomModule struct{}

// Register registers custom services
func (cm *CustomModule) Register(container *di.Container) error {
	container.RegisterFactory("custom_service", func(c *di.Container) (interface{}, error) {
		return &CustomService{}, nil
	})
	return nil
}

// CustomService represents a custom service
type CustomService struct{}

// DoSomething does something
func (cs *CustomService) DoSomething() {
	fmt.Println("Custom service doing something...")
}

// Example of using DI with configuration
func exampleWithConfiguration() {
	// Create application
	app := di.NewApplication()

	// Register configuration module
	app.RegisterModule(di.NewModule("config", &ConfigModule{}))

	// Initialize
	if err := app.Initialize(); err != nil {
		log.Fatal(err)
	}
	defer app.Close()

	// Get configuration
	var cfg *config.Config
	if err := app.GetTyped("app_config", &cfg); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Configuration loaded: %s:%d\n", cfg.Connection.Address, cfg.Connection.Port)
}

// ConfigModule provides configuration services
type ConfigModule struct{}

// Register registers configuration services
func (cm *ConfigModule) Register(container *di.Container) error {
	container.RegisterFactory("app_config", func(c *di.Container) (interface{}, error) {
		cfg := &config.Config{
			Connection: config.ConnectionConfig{
				Address: "localhost",
				Port:    3389,
			},
			Authentication: config.AuthConfig{
				Username: "admin",
				Password: "password",
			},
			Display: config.DisplayConfig{
				Width:      1024,
				Height:     768,
				ColorDepth: 24,
			},
		}
		return cfg, nil
	})
	return nil
}

// Example usage:
// go run examples/di_example/di_example.go -host 192.168.1.100 -username admin -password secret
//
// Features demonstrated:
// - Dependency injection container
// - Service registration and retrieval
// - Factory pattern for service creation
// - Module-based architecture
// - Type-safe service resolution
// - Application lifecycle management
// - Graceful shutdown
//
// Benefits:
// - Loose coupling between components
// - Easy testing with mock services
// - Centralized service management
// - Modular architecture
// - Dependency resolution
// - Service lifecycle management
