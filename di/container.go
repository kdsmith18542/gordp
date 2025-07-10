package di

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/kdsmith18542/gordp"
	"github.com/kdsmith18542/gordp/config"
	"github.com/kdsmith18542/gordp/gateway/webrtc"
	"github.com/kdsmith18542/gordp/management"
	"github.com/kdsmith18542/gordp/mobile"
)

// Container represents a dependency injection container
type Container struct {
	services  map[string]interface{}
	factories map[string]Factory
	mu        sync.RWMutex
	ctx       context.Context
	cancel    context.CancelFunc
}

// Factory represents a factory function for creating services
type Factory func(container *Container) (interface{}, error)

// NewContainer creates a new dependency injection container
func NewContainer() *Container {
	ctx, cancel := context.WithCancel(context.Background())
	return &Container{
		services:  make(map[string]interface{}),
		factories: make(map[string]Factory),
		ctx:       ctx,
		cancel:    cancel,
	}
}

// Register registers a service with the container
func (c *Container) Register(name string, service interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.services[name] = service
}

// RegisterFactory registers a factory function for creating services
func (c *Container) RegisterFactory(name string, factory Factory) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.factories[name] = factory
}

// Get retrieves a service from the container
func (c *Container) Get(name string) (interface{}, error) {
	c.mu.RLock()

	// Check if service already exists
	if service, exists := c.services[name]; exists {
		c.mu.RUnlock()
		return service, nil
	}

	// Check if factory exists
	factory, exists := c.factories[name]
	c.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("service '%s' not found", name)
	}

	// Create service using factory
	c.mu.Lock()
	defer c.mu.Unlock()

	// Double-check in case another goroutine created it
	if service, exists := c.services[name]; exists {
		return service, nil
	}

	service, err := factory(c)
	if err != nil {
		return nil, fmt.Errorf("failed to create service '%s': %w", name, err)
	}

	c.services[name] = service
	return service, nil
}

// GetTyped retrieves a service with type assertion
func (c *Container) GetTyped(name string, target interface{}) error {
	service, err := c.Get(name)
	if err != nil {
		return err
	}

	// Use reflection to set the target
	targetValue := reflect.ValueOf(target)
	if targetValue.Kind() != reflect.Ptr {
		return fmt.Errorf("target must be a pointer")
	}

	serviceValue := reflect.ValueOf(service)
	if !serviceValue.Type().AssignableTo(targetValue.Elem().Type()) {
		return fmt.Errorf("service type %s is not assignable to target type %s",
			serviceValue.Type(), targetValue.Elem().Type())
	}

	targetValue.Elem().Set(serviceValue)
	return nil
}

// Has checks if a service exists in the container
func (c *Container) Has(name string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	_, exists := c.services[name]
	if !exists {
		_, exists = c.factories[name]
	}
	return exists
}

// Remove removes a service from the container
func (c *Container) Remove(name string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.services, name)
	delete(c.factories, name)
}

// Clear removes all services from the container
func (c *Container) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.services = make(map[string]interface{})
	c.factories = make(map[string]Factory)
}

// List returns all registered service names
func (c *Container) List() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	names := make([]string, 0, len(c.services)+len(c.factories))

	for name := range c.services {
		names = append(names, name)
	}

	for name := range c.factories {
		// Only add if not already in services
		if _, exists := c.services[name]; !exists {
			names = append(names, name)
		}
	}

	return names
}

// Close closes the container and cleans up resources
func (c *Container) Close() error {
	c.cancel()

	// Call Close() method on services that implement it
	c.mu.RLock()
	services := make([]interface{}, 0, len(c.services))
	for _, service := range c.services {
		services = append(services, service)
	}
	c.mu.RUnlock()

	for _, service := range services {
		if closer, ok := service.(interface{ Close() error }); ok {
			if err := closer.Close(); err != nil {
				return fmt.Errorf("failed to close service: %w", err)
			}
		}
	}

	return nil
}

// Context returns the container's context
func (c *Container) Context() context.Context {
	return c.ctx
}

// ServiceProvider represents a service provider interface
type ServiceProvider interface {
	Register(container *Container) error
}

// Module represents a module that can register multiple services
type Module struct {
	name     string
	provider ServiceProvider
}

// NewModule creates a new module
func NewModule(name string, provider ServiceProvider) *Module {
	return &Module{
		name:     name,
		provider: provider,
	}
}

// Register registers the module with a container
func (m *Module) Register(container *Container) error {
	return m.provider.Register(container)
}

// Name returns the module name
func (c *Module) Name() string {
	return c.name
}

// CoreModule provides core GoRDP services
type CoreModule struct{}

// Register registers core services
func (cm *CoreModule) Register(container *Container) error {
	// Register configuration service
	container.RegisterFactory("config", func(c *Container) (interface{}, error) {
		return &config.Config{}, nil
	})

	// Register RDP client factory
	container.RegisterFactory("rdp_client_factory", func(c *Container) (interface{}, error) {
		return &RDPClientFactory{container: c}, nil
	})

	return nil
}

// RDPClientFactory creates RDP clients with dependencies
type RDPClientFactory struct {
	container *Container
}

// Create creates a new RDP client
func (rf *RDPClientFactory) Create(options *gordp.Option) (*gordp.Client, error) {
	client := gordp.NewClient(options)
	return client, nil
}

// CreateWithConfig creates a new RDP client with configuration
func (rf *RDPClientFactory) CreateWithConfig(cfg *config.Config) (*gordp.Client, error) {
	options := &gordp.Option{
		Addr:           fmt.Sprintf("%s:%d", cfg.Connection.Address, cfg.Connection.Port),
		UserName:       cfg.Authentication.Username,
		Password:       cfg.Authentication.Password,
		ConnectTimeout: 10 * time.Second,
	}

	return rf.Create(options)
}

// PluginModule provides plugin services
type PluginModule struct{}

// Register registers plugin services
func (pm *PluginModule) Register(container *Container) error {
	// Plugin services will be implemented later
	return nil
}

// GatewayModule provides gateway services
type GatewayModule struct{}

// Register registers gateway services
func (gm *GatewayModule) Register(container *Container) error {
	// Register WebRTC gateway
	container.RegisterFactory("webrtc_gateway", func(c *Container) (interface{}, error) {
		gatewayConfig := &webrtc.GatewayConfig{
			Port:           8080,
			MaxConnections: 100,
			SessionTimeout: 30 * time.Minute,
			EnableCORS:     true,
			StaticPath:     "./gateway/webrtc/web",
		}

		return webrtc.NewWebRTCGateway(gatewayConfig), nil
	})

	return nil
}

// ManagementModule provides management services
type ManagementModule struct{}

// Register registers management services
func (mm *ManagementModule) Register(container *Container) error {
	// Register management console
	container.RegisterFactory("management_console", func(c *Container) (interface{}, error) {
		cfg := &management.ConsoleConfig{
			Port:                8080,
			AdminPort:           8081,
			SessionTimeout:      30 * time.Minute,
			MaxSessions:         1000,
			EnableLoadBalancing: true,
			EnableRecording:     true,
			EnableAuditLog:      true,
			DatabaseURL:         "sqlite://gordp.db",
			StaticPath:          "./management/web",
		}

		return management.NewManagementConsole(cfg), nil
	})

	return nil
}

// MobileModule provides mobile services
type MobileModule struct{}

// Register registers mobile services
func (mm *MobileModule) Register(container *Container) error {
	// Register mobile client factory
	container.RegisterFactory("mobile_client_factory", func(c *Container) (interface{}, error) {
		return &MobileClientFactory{container: c}, nil
	})

	return nil
}

// MobileClientFactory creates mobile clients
type MobileClientFactory struct {
	container *Container
}

// Create creates a new mobile client
func (mcf *MobileClientFactory) Create() *mobile.MobileClient {
	return mobile.NewMobileClient()
}

// CreateWithConfig creates a new mobile client with configuration
func (mcf *MobileClientFactory) CreateWithConfig(cfg *mobile.MobileConfig) *mobile.MobileClient {
	client := mobile.NewMobileClient()
	// Apply configuration here
	return client
}

// Application represents a GoRDP application with dependency injection
type Application struct {
	container *Container
	modules   []*Module
}

// NewApplication creates a new application
func NewApplication() *Application {
	container := NewContainer()

	app := &Application{
		container: container,
		modules:   make([]*Module, 0),
	}

	// Register default modules
	app.RegisterModule(NewModule("core", &CoreModule{}))
	app.RegisterModule(NewModule("plugin", &PluginModule{}))
	app.RegisterModule(NewModule("gateway", &GatewayModule{}))
	app.RegisterModule(NewModule("management", &ManagementModule{}))
	app.RegisterModule(NewModule("mobile", &MobileModule{}))

	return app
}

// RegisterModule registers a module with the application
func (app *Application) RegisterModule(module *Module) {
	app.modules = append(app.modules, module)
}

// Initialize initializes the application and all modules
func (app *Application) Initialize() error {
	for _, module := range app.modules {
		if err := module.Register(app.container); err != nil {
			return fmt.Errorf("failed to register module '%s': %w", module.Name(), err)
		}
	}
	return nil
}

// Get retrieves a service from the application
func (app *Application) Get(name string) (interface{}, error) {
	return app.container.Get(name)
}

// GetTyped retrieves a service with type assertion
func (app *Application) GetTyped(name string, target interface{}) error {
	return app.container.GetTyped(name, target)
}

// Container returns the underlying container
func (app *Application) Container() *Container {
	return app.container
}

// Close closes the application
func (app *Application) Close() error {
	return app.container.Close()
}

// Example usage:
//
// func main() {
//     app := di.NewApplication()
//     if err := app.Initialize(); err != nil {
//         log.Fatal(err)
//     }
//     defer app.Close()
//
//     // Get services
//     var clientFactory *di.RDPClientFactory
//     if err := app.GetTyped("rdp_client_factory", &clientFactory); err != nil {
//         log.Fatal(err)
//     }
//
//     // Use services
//     client, err := clientFactory.CreateWithConfig(config)
//     if err != nil {
//         log.Fatal(err)
//     }
// }
