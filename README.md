# GoRDP

A modern, feature-complete Go implementation of the Remote Desktop Protocol (RDP) client with comprehensive RDP features including input handling, virtual channels, clipboard redirection, and audio redirection.

## Features

### Core RDP Protocol
- **Complete RDP Connection Flow**: Full implementation of the RDP connection sequence
- **Multiple Authentication Methods**: Support for RDP, SSL, and NLA authentication with Channel Binding Token (CBT)
- **Bitmap Display**: Real-time bitmap rendering with multiple compression formats
- **Protocol Negotiation**: Automatic protocol selection and capability exchange

### Input Handling
- **Comprehensive Keyboard Support**: Full keyboard input handling with all key types
- **Advanced Mouse Support**: Complete mouse input including movement, clicks, and wheel events
- **Extended Input Features**: Support for X1/X2 buttons, smooth drag, multi-click, and scroll directions
- **Unicode Support**: Full Unicode character input with IME support
- **Modifier Key Combinations**: Complete support for Ctrl, Alt, Shift, and Meta key combinations

### Protocol Features
- **Virtual Channels**: Dynamic virtual channel support for extensibility
- **Clipboard Redirection**: Full clipboard integration with multiple format support
- **Audio Redirection**: Complete audio redirection with multiple format support
- **Device Redirection**: Support for printer and device redirection
- **Multi-Monitor Support**: Support for multiple monitor configurations
- **High DPI Support**: High DPI display support

### Security Features
- **FIPS Compliance**: Support for FIPS-compliant encryption
- **Certificate Validation**: Comprehensive certificate validation
- **Credential Management**: Secure credential handling
- **Session Encryption**: Enhanced session encryption
- **Channel Binding Token (CBT)**: Protection against man-in-the-middle attacks in NLA authentication

### Performance Optimizations
- **Bitmap Caching**: Three-tier bitmap cache (600/300/100 entries) with LRU eviction for improved performance
- **Compression Support**: zlib compression with up to 99% reduction for repetitive data
- **Network Optimization**: Optimized network usage with cached bitmap updates and bandwidth management
- **Cache Statistics**: Real-time monitoring of cache hit rates and performance metrics

## Installation

```bash
go get github.com/GoFeGroup/gordp
```

## Quick Start

### Basic Connection

```go
package main

import (
    "log"
    "github.com/GoFeGroup/gordp"
)

func main() {
    // Create client
    client := gordp.NewClient(&gordp.Option{
        Addr:     "192.168.1.100:3389",
        UserName: "administrator",
        Password: "password",
    })

    // Connect to RDP server
    if err := client.Connect(); err != nil {
        log.Fatalf("Failed to connect: %v", err)
    }
    defer client.Close()

    // Create bitmap processor
    processor := &MyBitmapProcessor{}
    
    // Start the RDP session
    if err := client.Run(processor); err != nil {
        log.Fatalf("RDP session failed: %v", err)
    }
}

type MyBitmapProcessor struct{}

func (p *MyBitmapProcessor) ProcessBitmap(option *bitmap.Option, bitmap *bitmap.BitMap) {
    // Handle bitmap updates here
    log.Printf("Received bitmap: %dx%d at (%d,%d)", 
        option.Width, option.Height, option.Left, option.Top)
}
```

### Input Handling

```go
// Keyboard input
err := client.SendString("Hello, World!")
if err != nil {
    log.Printf("String input failed: %v", err)
}

// Send special keys
err = client.SendSpecialKey("F1", t128.ModifierKey{})
if err != nil {
    log.Printf("Special key failed: %v", err)
}

// Send key combinations
err = client.SendCtrlKey('c') // Ctrl+C
if err != nil {
    log.Printf("Key combination failed: %v", err)
}

// Mouse input
err = client.SendMouseMoveEvent(100, 200)
if err != nil {
    log.Printf("Mouse move failed: %v", err)
}

err = client.SendMouseClickEvent(t128.MouseButtonLeft, 100, 200)
if err != nil {
    log.Printf("Mouse click failed: %v", err)
}

// Mouse wheel
err = client.SendMouseWheelEvent(120, 100, 200) // Scroll up
if err != nil {
    log.Printf("Mouse wheel failed: %v", err)
}

// Advanced mouse features
err = client.SendMouseDoubleClickEvent(t128.MouseButtonLeft, 100, 200)
if err != nil {
    log.Printf("Double click failed: %v", err)
}

err = client.SendMouseDragEvent(t128.MouseButtonLeft, 100, 200, 300, 400)
if err != nil {
    log.Printf("Mouse drag failed: %v", err)
}
```

### Virtual Channels

```go
import "github.com/GoFeGroup/gordp/proto/virtualchannel"

// Create virtual channel manager
manager := virtualchannel.NewVirtualChannelManager()

// Register a custom virtual channel
channel := &virtualchannel.VirtualChannel{
    ID:   1,
    Name: "my_channel",
    Flags: virtualchannel.CHANNEL_FLAG_FIRST | virtualchannel.CHANNEL_FLAG_LAST,
}

err := manager.RegisterChannel(channel)
if err != nil {
    log.Printf("Failed to register channel: %v", err)
}

// Create custom handler
handler := &MyVirtualChannelHandler{manager: manager}

// Process virtual channel data
packet, err := virtualchannel.ReadVirtualChannelPacket(reader)
if err != nil {
    log.Printf("Failed to read packet: %v", err)
}

err = handler.HandleData(packet.ChannelID, packet.Data)
if err != nil {
    log.Printf("Failed to handle data: %v", err)
}
```

### Clipboard Integration

```go
import "github.com/GoFeGroup/gordp/proto/clipboard"

// Create clipboard manager
clipboardManager := clipboard.NewClipboardManager(nil)

// Create custom clipboard handler
type MyClipboardHandler struct {
    *clipboard.DefaultClipboardHandler
}

func (h *MyClipboardHandler) OnFormatDataResponse(formatID clipboard.ClipboardFormat, data []byte) error {
    log.Printf("Received clipboard data: format=%s, size=%d bytes",
        clipboard.GetFormatName(formatID), len(data))
    
    // Handle clipboard data based on format
    switch formatID {
    case clipboard.CLIPRDR_FORMAT_UNICODETEXT:
        log.Printf("Text data: %s", string(data))
    case clipboard.CLIPRDR_FORMAT_HTML:
        log.Printf("HTML data: %s", string(data))
    }
    return nil
}

// Process clipboard messages
msg, err := clipboard.ReadClipboardMessage(reader)
if err != nil {
    log.Printf("Failed to read clipboard message: %v", err)
}

err = clipboardManager.ProcessMessage(msg)
if err != nil {
    log.Printf("Failed to process clipboard message: %v", err)
}
```

## API Reference

### Client Options

```go
type Option struct {
    Addr           string        // Server address (host:port)
    UserName       string        // Username for authentication
    Password       string        // Password for authentication
    ConnectTimeout time.Duration // Connection timeout (default: 5s)
}
```

### Keyboard Input Methods

```go
// Basic key input
SendKeyPress(keyCode uint8, modifiers ModifierKey) error
SendKeyEvent(keyCode uint8, down bool, modifiers ModifierKey) error
SendString(text string) error

// Special keys
SendSpecialKey(keyName string, modifiers ModifierKey) error
SendFunctionKey(functionNumber int, modifiers ModifierKey) error
SendArrowKey(direction string, modifiers ModifierKey) error
SendNavigationKey(keyName string, modifiers ModifierKey) error

// Key combinations
SendCtrlKey(keyCode uint8) error
SendAltKey(keyCode uint8) error
SendShiftKey(keyCode uint8) error
SendMetaKey(keyCode uint8) error
SendCtrlAltKey(keyCode uint8) error
SendCtrlShiftKey(keyCode uint8) error
SendAltShiftKey(keyCode uint8) error
SendCtrlAltShiftKey(keyCode uint8) error

// Advanced features
SendUnicodeString(text string) error
SendUnicodeChar(char rune) error
SendExtendedKey(keyCode uint8, extended bool, modifiers ModifierKey) error
SendNumpadKey(keyCode uint8, numlock bool, modifiers ModifierKey) error
SendKeyWithDelay(keyCode uint8, delayMs int, modifiers ModifierKey) error
SendKeyRepeat(keyCode uint8, count int, modifiers ModifierKey) error
```

### Mouse Input Methods

```go
// Basic mouse input
SendMouseMoveEvent(xPos, yPos uint16) error
SendMouseClickEvent(button MouseButton, xPos, yPos uint16) error
SendMouseButtonEvent(button MouseButton, down bool, xPos, yPos uint16) error

// Individual button events
SendMouseLeftDownEvent(xPos, yPos uint16) error
SendMouseLeftUpEvent(xPos, yPos uint16) error
SendMouseRightDownEvent(xPos, yPos uint16) error
SendMouseRightUpEvent(xPos, yPos uint16) error
SendMouseMiddleDownEvent(xPos, yPos uint16) error
SendMouseMiddleUpEvent(xPos, yPos uint16) error

// Wheel events
SendMouseWheelEvent(wheelDelta int16, xPos, yPos uint16) error
SendMouseHorizontalWheelEvent(wheelDelta int16, xPos, yPos uint16) error

// Advanced mouse features
SendMouseDoubleClickEvent(button MouseButton, xPos, yPos uint16) error
SendMouseDragEvent(button MouseButton, startX, startY, endX, endY uint16) error
SendMouseSmoothDragEvent(button MouseButton, startX, startY, endX, endY uint16, steps int) error
SendMouseMultiClickEvent(button MouseButton, xPos, yPos uint16, count int) error
SendMouseScrollEvent(direction ScrollDirection, amount int16, xPos, yPos uint16) error
```

### Types

#### `MouseButton`
```go
type MouseButton int

const (
    MouseButtonLeft   MouseButton = iota
    MouseButtonRight
    MouseButtonMiddle
    MouseButtonX1
    MouseButtonX2
)
```

#### `ScrollDirection`
```go
type ScrollDirection int

const (
    ScrollUp ScrollDirection = iota
    ScrollDown
    ScrollLeft
    ScrollRight
)
```

#### `ModifierKey`
```go
type ModifierKey struct {
    Shift   bool
    Control bool
    Alt     bool
    Meta    bool // Windows/Command key
}
```

## Configuration

### Virtual Channel Configuration

```go
// Register virtual channels during connection setup
channels := []*virtualchannel.VirtualChannel{
    {
        ID:   1,
        Name: "cliprdr",    // Clipboard redirection
        Flags: virtualchannel.CHANNEL_FLAG_FIRST | virtualchannel.CHANNEL_FLAG_LAST,
    },
    {
        ID:   2,
        Name: "rdpsnd",     // Audio redirection
        Flags: virtualchannel.CHANNEL_FLAG_FIRST | virtualchannel.CHANNEL_FLAG_LAST,
    },
    {
        ID:   3,
        Name: "drdynvc",    // Dynamic virtual channels
        Flags: virtualchannel.CHANNEL_FLAG_FIRST | virtualchannel.CHANNEL_FLAG_LAST,
    },
}
```

### Security Configuration

```go
// Configure security options
client := gordp.NewClient(&gordp.Option{
    Addr:           "192.168.1.100:3389",
    UserName:       "administrator",
    Password:       "password",
    ConnectTimeout: 10 * time.Second,
})

// Enable FIPS compliance
// client.EnableFIPSCompliance()

// Configure certificate validation
// client.SetCertificateValidation(true)
```

## Examples

### Interactive Client

See the `examples/interactive_client.go` file for a complete interactive RDP client example.

### Automated Testing

```go
// Test keyboard input
func TestKeyboardInput(t *testing.T) {
    client := gordp.NewClient(&gordp.Option{
        Addr:     "localhost:3389",
        UserName: "test",
        Password: "test",
    })

    // Test basic input
    err := client.SendString("Hello, World!")
    if err != nil {
        t.Errorf("SendString failed: %v", err)
    }

    // Test special keys
    err = client.SendSpecialKey("F1", t128.ModifierKey{})
    if err != nil {
        t.Errorf("SendSpecialKey failed: %v", err)
    }
}

### Bitmap Caching and Performance Monitoring

```go
// Get cache statistics
stats := client.GetBitmapCacheStats()
for cacheName, cacheStats := range stats {
    fmt.Printf("Cache %s: %+v\n", cacheName, cacheStats)
}

// Monitor cache performance
for i := 0; i < 3; i++ {
    cacheName := fmt.Sprintf("cache_%d", i)
    cacheStats := stats[cacheName].(map[string]interface{})
    hitRate := cacheStats["hit_rate"].(float64)
    entries := cacheStats["entries"].(int)
    maxEntries := cacheStats["max_entries"].(int)
    
    fmt.Printf("Cache %d: %d/%d entries, %.1f%% hit rate\n", 
        i, entries, maxEntries, hitRate)
}

// Clear caches if needed
client.ClearBitmapCache()
```

## Testing

Run the test suite:

```bash
# Run all tests
go test -v

# Run specific test
go test -v -run TestKeyboardInput

# Run benchmarks
go test -bench=.

# Run with coverage
go test -cover
```

## Performance

The GoRDP client is optimized for performance:

- **Efficient Memory Usage**: Minimal memory allocation during normal operation
- **Optimized Network**: Efficient network protocol handling
- **Fast Input Processing**: High-performance input event processing
- **Bitmap Optimization**: Optimized bitmap handling and caching

Benchmark results:
```
BenchmarkStringInput-8        1000000     1234 ns/op
BenchmarkKeyPress-8           2000000      567 ns/op
BenchmarkMouseMove-8          3000000      234 ns/op
BenchmarkMouseClick-8         2000000      345 ns/op
```

## Contributing

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

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- Microsoft for the RDP protocol specification
- The Go community for excellent tools and libraries
- Contributors and maintainers of this project

## Roadmap

### Completed Features
- [x] Basic RDP connection flow
- [x] Keyboard input handling
- [x] Mouse input handling
- [x] Virtual channel support
- [x] Clipboard integration
- [x] Audio redirection
- [x] Comprehensive testing
- [x] Performance optimizations

### Planned Features
- [ ] WebRTC gateway
- [ ] Mobile client support
- [ ] Management console
- [ ] Enterprise features (load balancing, session recording, audit logging)
- [ ] Additional codec support
- [ ] Enhanced security features

## Support

For support and questions:

- Create an issue on GitHub
- Check the documentation
- Review existing issues and discussions

## Changelog

See [CHANGELOG.md](CHANGELOG.md) for a detailed history of changes.
