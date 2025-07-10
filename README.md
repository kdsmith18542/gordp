# GoRDP - Production-Grade RDP Client in Go

[![Go Report Card](https://goreportcard.com/badge/github.com/kdsmith18542/gordp)](https://goreportcard.com/report/github.com/kdsmith18542/gordp)
[![Go Version](https://img.shields.io/github/go-mod/go-version/kdsmith18542/gordp)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

> **🎨 GUI coming soon!** A modern graphical user interface is in development to complement the command-line client.

GoRDP is a comprehensive, production-grade implementation of the Remote Desktop Protocol (RDP) client in Go. This library provides full RDP protocol support including input handling, clipboard integration, audio redirection, device redirection, and multi-monitor support.

## 🚀 Features

### ✅ Core Protocol Support
- **Complete RDP Connection Flow** - Full implementation of RDP protocol phases
- **NLA Authentication** - Network Level Authentication support
- **Virtual Channels** - Static and dynamic virtual channel support
- **Bitmap Processing** - Efficient bitmap handling with caching
- **Protocol Negotiation** - Comprehensive capability exchange

### ✅ Input Handling
- **Keyboard Input** - Full keyboard support including:
  - ASCII and Unicode text input
  - Special keys (F1-F24, arrows, navigation keys)
  - Modifier key combinations (Ctrl, Alt, Shift, Meta)
  - Function keys, media keys, browser keys
  - Extended key codes and numpad support
  - Key sequences and timing control
- **Mouse Input** - Complete mouse support including:
  - Movement and positioning
  - Button clicks (left, right, middle, X1, X2)
  - Wheel scrolling (vertical and horizontal)
  - Advanced features (drag, double-click, multi-click)
  - Smooth drag operations

### ✅ Clipboard Integration
- **Format Support** - Multiple clipboard formats (text, HTML, bitmap, etc.)
- **Bidirectional Transfer** - Copy and paste between client and server
- **File Transfer** - Clipboard file list support
- **Event Handling** - Custom clipboard event handlers

### ✅ Audio Redirection
- **Audio Streaming** - Real-time audio from server to client
- **Format Negotiation** - Multiple audio format support
- **Quality Control** - Configurable audio quality settings
- **Event Handling** - Custom audio event handlers

### ✅ Device Redirection
- **Printer Support** - Remote printer redirection
- **Drive Access** - File system redirection
- **Port Access** - Serial/parallel port redirection
- **USB Support** - USB device redirection
- **Smart Card** - Smart card reader support

### ✅ Multi-Monitor Support
- **Monitor Configuration** - Support for multiple monitors
- **High DPI** - High DPI monitor support
- **Orientation** - Portrait and landscape orientation
- **Scaling** - Desktop and device scaling factors
- **Dynamic Layout** - Runtime monitor layout changes

### ✅ Performance & Security
- **Bitmap Caching** - Efficient bitmap caching system
- **Compression** - RDP6 and other compression support
- **Network Optimization** - Optimized network usage
- **Certificate Validation** - Server certificate validation
- **FIPS Compliance** - Federal Information Processing Standards support

### ✅ Developer Experience
- **Comprehensive Testing** - Extensive unit and integration tests
- **Error Handling** - Production-grade error handling
- **Logging** - Structured logging with multiple levels
- **Documentation** - Complete API documentation
- **Examples** - Working examples for all features

## 📦 Installation

### Option 1: Pre-built Binaries (Recommended)

Download the latest release for your platform:

**Linux (x86_64):**
```bash
# Download and install
wget https://github.com/kdsmith18542/gordp/releases/latest/download/gordp-linux-amd64
chmod +x gordp-linux-amd64
sudo mv gordp-linux-amd64 /usr/local/bin/gordp

# Or using curl
curl -L https://github.com/kdsmith18542/gordp/releases/latest/download/gordp-linux-amd64 -o gordp
chmod +x gordp
sudo mv gordp /usr/local/bin/
```

**Windows (x64):**
```powershell
# Download using PowerShell
Invoke-WebRequest -Uri "https://github.com/kdsmith18542/gordp/releases/latest/download/gordp-windows-amd64.exe" -OutFile "gordp.exe"
# Add to PATH or run from current directory
```

**macOS (x86_64/ARM64):**
```bash
# Using Homebrew (recommended)
brew install kdsmith18542/tap/gordp

# Or manual download
curl -L https://github.com/kdsmith18542/gordp/releases/latest/download/gordp-darwin-amd64 -o gordp
chmod +x gordp
sudo mv gordp /usr/local/bin/
```

### Option 2: Using Go (Build from Source)

If you prefer to build from source or need the latest development version:

```bash
# Install Go 1.18+ first, then:
go install github.com/kdsmith18542/gordp@latest

# Or clone and build
git clone https://github.com/kdsmith18542/gordp.git
cd gordp
make build
sudo make install
```

### Option 3: Docker

```bash
# Pull and run the official image
docker pull kdsmith18542/gordp:latest
docker run -it --rm kdsmith18542/gordp:latest

# Or build locally
git clone https://github.com/kdsmith18542/gordp.git
cd gordp
docker build -t gordp .
docker run -it --rm gordp
```

### Option 4: Package Managers

**Ubuntu/Debian:**
```bash
# Add repository (when available)
curl -fsSL https://packages.gordp.dev/gpg | sudo gpg --dearmor -o /usr/share/keyrings/gordp-archive-keyring.gpg
echo "deb [arch=amd64 signed-by=/usr/share/keyrings/gordp-archive-keyring.gpg] https://packages.gordp.dev/ubuntu $(lsb_release -cs) main" | sudo tee /etc/apt/sources.list.d/gordp.list
sudo apt update
sudo apt install gordp
```

**macOS (Homebrew):**
```bash
brew install kdsmith18542/tap/gordp
```

**Windows (Chocolatey):**
```powershell
choco install gordp
```

**Windows (Scoop):**
```powershell
scoop install gordp
```

### Option 5: Universal Installer Scripts

**Linux/macOS:**
```bash
# Download and run the installer
curl -fsSL https://raw.githubusercontent.com/kdsmith18542/gordp/main/install.sh | bash

# Or download first, then run
wget https://raw.githubusercontent.com/kdsmith18542/gordp/main/install.sh
chmod +x install.sh
./install.sh
```

**Windows (PowerShell):**
```powershell
# Download and run the installer
Invoke-Expression (Invoke-WebRequest -Uri "https://raw.githubusercontent.com/kdsmith18542/gordp/main/install.ps1").Content

# Or download first, then run
Invoke-WebRequest -Uri "https://raw.githubusercontent.com/kdsmith18542/gordp/main/install.ps1" -OutFile "install.ps1"
.\install.ps1
```

### Option 6: Development Installation

For developers who want to contribute or use the latest features:

```bash
# Clone the repository
git clone https://github.com/kdsmith18542/gordp.git
cd gordp

# Setup development environment
make dev-setup

# Build and install
make build
make install

# Run tests
make test
```

## 🔧 Prerequisites

- **Go 1.18+** (only required for building from source)
- **Git** (for cloning the repository)
- **Make** (for using the Makefile build system)

### System Requirements

- **Linux**: glibc 2.17+ (CentOS 7+, Ubuntu 16.04+, etc.)
- **Windows**: Windows 7+ (x64)
- **macOS**: macOS 10.12+ (Sierra)
- **Memory**: 128MB RAM minimum, 512MB recommended
- **Network**: TCP/IP connectivity for RDP connections

### Optional Dependencies

- **Docker**: For containerized deployment
- **OpenSSL**: For enhanced security features
- **PulseAudio/ALSA**: For audio redirection (Linux)
- **Core Audio**: For audio redirection (macOS)
- **DirectSound**: For audio redirection (Windows)

## 🚀 Quick Start

### Basic Connection

```go
package main

import (
    "log"
    "time"
    "github.com/kdsmith18542/gordp"
    "github.com/kdsmith18542/gordp/proto/bitmap"
)

type MyProcessor struct{}

func (p *MyProcessor) ProcessBitmap(option *bitmap.Option, bitmap *bitmap.BitMap) {
    log.Printf("Received bitmap: %dx%d at (%d,%d)", 
        option.Width, option.Height, option.Left, option.Top)
}

func main() {
    client := gordp.NewClient(&gordp.Option{
        Addr:           "192.168.1.100:3389",
        UserName:       "username",
        Password:       "password",
        ConnectTimeout: 10 * time.Second,
    })

    err := client.Connect()
    if err != nil {
        log.Fatal("Connection failed:", err)
    }
    defer client.Close()

    processor := &MyProcessor{}
    err = client.Run(processor)
    if err != nil {
        log.Fatal("Session failed:", err)
    }
}
```

### Advanced Features

```go
// Multi-monitor setup
monitors := []mcs.MonitorLayout{
    {
        Left:   0, Top: 0, Right: 1920, Bottom: 1080, Flags: 0x01, // Primary
    },
    {
        Left:   1920, Top: 0, Right: 3840, Bottom: 1080, Flags: 0x00, // Secondary
    },
}

client := gordp.NewClient(&gordp.Option{
    Addr:     "server:3389",
    UserName: "user",
    Password: "pass",
    Monitors: monitors,
})

// Register handlers
client.RegisterClipboardHandler(&MyClipboardHandler{})
client.RegisterDeviceHandler(&MyDeviceHandler{})
client.RegisterDynamicVirtualChannelHandler("MY_CHANNEL", &MyChannelHandler{})

// Input handling
client.SendString("Hello, World!")
client.SendKeyPress(t128.VK_RETURN, t128.ModifierKey{})
client.SendMouseClickEvent(t128.MouseButtonLeft, 100, 200)
```

## 📚 Documentation

- [API Documentation](docs/api.md) - Complete API reference
- [Examples](examples/) - Working examples for all features
- [Protocol Documentation](docs/protocol.md) - RDP protocol details

## 🧪 Testing

Run the comprehensive test suite:

```bash
# Run all tests
go test -v

# Run specific test categories
go test -v -run TestKeyboardInput
go test -v -run TestMouseInput
go test -v -run TestClipboardFunctionality
go test -v -run TestDeviceRedirection

# Run with coverage
go test -cover

# Run benchmarks
go test -bench=.
```

## 📋 Examples

### Interactive Client
```bash
go run examples/interactive_example/interactive_client.go 192.168.1.100:3389 username password
```

### Comprehensive Client
```bash
go run examples/comprehensive_client.go 192.168.1.100:3389 username password
```

### WebRTC Gateway
```bash
go run examples/webrtc_example/webrtc_gateway.go -port 8080
```

### Management Console
```bash
go run examples/management_example/management_console.go -port 8080
```

### Dependency Injection
```bash
go run examples/di_example/di_example.go -host 192.168.1.100 -username admin -password secret
```

### Plugin Demo
```bash
go run examples/plugin_example/plugin_demo.go -host 192.168.1.100 -username admin -password secret
```

### Configuration Example
```bash
go run examples/config_example/config_client.go -config config.json
```

## 🔧 Configuration

### Client Options

```go
type Option struct {
    Addr           string        // Server address (host:port)
    UserName       string        // Username for authentication
    Password       string        // Password for authentication
    ConnectTimeout time.Duration // Connection timeout
    Monitors       []mcs.MonitorLayout // Multi-monitor configuration
}
```

### Multi-Monitor Configuration

```go
monitors := []mcs.MonitorLayout{
    {
        Left:               0,
        Top:                0,
        Right:              1920,
        Bottom:             1080,
        Flags:              0x01, // Primary monitor
        MonitorIndex:       0,
        PhysicalWidthMm:    520,
        PhysicalHeightMm:   290,
        Orientation:        0, // Landscape
        DesktopScaleFactor: 100,
        DeviceScaleFactor:  100,
    },
    // Add more monitors as needed
}
```

## 🎯 Use Cases

- **Remote Desktop Access** - Connect to Windows servers and workstations
- **Automated Testing** - Automated UI testing of remote applications
- **Remote Administration** - Server management and administration
- **Application Streaming** - Stream applications from remote servers
- **Virtual Desktop Infrastructure (VDI)** - Connect to virtual desktops
- **Remote Development** - Development on remote machines
- **Web-Based Access** - Browser-based RDP connections via WebRTC gateway
- **Mobile Access** - RDP connections from mobile devices
- **Enterprise Management** - Centralized RDP session management and monitoring
- **Modern Applications** - Applications using dependency injection and modern patterns

## 🔒 Security Features

- **NLA Authentication** - Network Level Authentication
- **Certificate Validation** - Server certificate verification
- **Encrypted Communication** - All data encrypted in transit
- **Credential Security** - Secure credential handling
- **FIPS Compliance** - Federal Information Processing Standards

## 🚀 Performance

- **Efficient Bitmap Handling** - Optimized bitmap processing and caching
- **Network Optimization** - Minimal network overhead
- **Memory Management** - Efficient memory usage
- **Concurrent Processing** - Multi-threaded operation
- **Compression Support** - Multiple compression algorithms

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guidelines](CONTRIBUTING.md) for details.

### Development Setup

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Run the test suite
6. Submit a pull request

### Code Style

- Follow Go conventions and best practices
- Use meaningful variable and function names
- Add comments for complex logic
- Write comprehensive tests
- Update documentation as needed

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- Microsoft for the RDP protocol specification
- The Go community for excellent tools and libraries
- Contributors and maintainers of this project

## 📞 Support

For support and questions:

- Create an issue on GitHub
- Check the documentation
- Review existing issues and discussions

## 🗺️ Roadmap

### Completed Features ✅
- [x] Basic RDP connection flow
- [x] Keyboard input handling
- [x] Mouse input handling
- [x] Virtual channel support
- [x] Clipboard integration
- [x] Audio redirection
- [x] Device redirection
- [x] Multi-monitor support
- [x] Comprehensive testing
- [x] Performance optimizations
- [x] Security features
- [x] Documentation

### Completed Optional Features ✅
- [x] WebRTC gateway - Web-based RDP client for browser access
- [x] Mobile client support - iOS and Android RDP client framework
- [x] Management console - Enterprise session management and monitoring
- [x] Enterprise features - Load balancing, session recording, audit logging
- [x] Dependency injection system - Modern DI container for applications
- [x] Advanced modernization - Context support, error wrapping, configuration management

### Future Enhancements 🚧
- [ ] Additional codec support (RemoteFX, H.264)
- [ ] Enhanced security features (FIPS compliance, certificate management)
- [ ] Cloud integration (AWS, Azure, GCP)
- [ ] Container support (Docker, Kubernetes)
- [ ] Advanced load balancing algorithms
- [ ] Performance monitoring and analytics

## 📊 Status

This project is **production-ready** and actively maintained. All core RDP features are implemented and thoroughly tested.

---

**GoRDP** - Production-grade RDP client implementation in Go
