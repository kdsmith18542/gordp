package gordp

import (
	"context"
	"image"
	"testing"
	"time"

	"github.com/kdsmith18542/gordp/proto/bitmap"
	"github.com/kdsmith18542/gordp/proto/clipboard"
	"github.com/kdsmith18542/gordp/proto/device"
	"github.com/kdsmith18542/gordp/proto/mcs"
	"github.com/kdsmith18542/gordp/proto/t128"
	"github.com/stretchr/testify/assert"
)

// TestClientCreationAndConfiguration tests client creation and configuration
func TestClientCreationAndConfiguration(t *testing.T) {
	t.Run("BasicClientCreation", func(t *testing.T) {
		option := &Option{
			Addr:           "localhost:3389",
			UserName:       "testuser",
			Password:       "testpass",
			ConnectTimeout: 10 * time.Second,
		}

		client := NewClient(option)
		assert.NotNil(t, client)
		assert.Equal(t, "localhost:3389", client.option.Addr)
		assert.Equal(t, "testuser", client.option.UserName)
		assert.Equal(t, "testpass", client.option.Password)
		assert.Equal(t, 10*time.Second, client.option.ConnectTimeout)
	})

	t.Run("ClientWithContext", func(t *testing.T) {
		option := &Option{
			Addr:     "localhost:3389",
			UserName: "testuser",
			Password: "testpass",
		}

		ctx := context.Background()
		client := NewClientWithContext(ctx, option)
		assert.NotNil(t, client)
		assert.NotNil(t, client.ctx)
		assert.NotNil(t, client.cancel)
	})

	t.Run("ClientWithMonitors", func(t *testing.T) {
		monitors := []mcs.MonitorLayout{
			{Left: 0, Top: 0, Right: 1920, Bottom: 1080, Flags: 0x01},
			{Left: 1920, Top: 0, Right: 3840, Bottom: 1080, Flags: 0x00},
		}

		option := &Option{
			Addr:     "localhost:3389",
			UserName: "testuser",
			Password: "testpass",
			Monitors: monitors,
		}

		client := NewClient(option)
		assert.NotNil(t, client)
		assert.Len(t, client.monitors, 2)
		assert.Equal(t, int32(1920), client.monitors[0].Right)
		assert.Equal(t, int32(1080), client.monitors[0].Bottom)
	})

	t.Run("DefaultTimeout", func(t *testing.T) {
		option := &Option{
			Addr:     "localhost:3389",
			UserName: "testuser",
			Password: "testpass",
		}

		client := NewClient(option)
		assert.Equal(t, 5*time.Second, client.option.ConnectTimeout)
	})
}

// TestVirtualChannelManagement tests virtual channel management
func TestVirtualChannelManagement(t *testing.T) {
	client := NewClient(&Option{
		Addr:     "localhost:3389",
		UserName: "test",
		Password: "test",
	})

	t.Run("DefaultVirtualChannels", func(t *testing.T) {
		// Check that default virtual channels are registered
		channels := client.vcManager.ListChannels()
		assert.Len(t, channels, 4)

		// Check specific channels
		cliprdr, exists := client.vcManager.GetChannelByName("cliprdr")
		assert.True(t, exists)
		assert.NotNil(t, cliprdr)
		assert.Equal(t, uint16(1), cliprdr.ID)

		rdpsnd, exists := client.vcManager.GetChannelByName("rdpsnd")
		assert.True(t, exists)
		assert.NotNil(t, rdpsnd)
		assert.Equal(t, uint16(2), rdpsnd.ID)

		drdynvc, exists := client.vcManager.GetChannelByName("drdynvc")
		assert.True(t, exists)
		assert.NotNil(t, drdynvc)
		assert.Equal(t, uint16(3), drdynvc.ID)

		rdpdr, exists := client.vcManager.GetChannelByName("rdpdr")
		assert.True(t, exists)
		assert.NotNil(t, rdpdr)
		assert.Equal(t, uint16(4), rdpdr.ID)
	})

	t.Run("DynamicVirtualChannelRegistration", func(t *testing.T) {
		handler := &testDVCHandler{}
		err := client.RegisterDynamicVirtualChannelHandler("TEST_CHANNEL", handler)
		assert.NoError(t, err)

		channels := client.ListDynamicVirtualChannels()
		assert.Contains(t, channels, "TEST_CHANNEL")
	})

	t.Run("VirtualChannelDataSending", func(t *testing.T) {
		// This should fail without connection, but test the method exists
		err := client.SendVirtualChannelData("cliprdr", []byte("test"), 0)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no active connection")
	})
}

// TestClipboardManagement tests clipboard management
func TestClipboardManagement(t *testing.T) {
	client := NewClient(&Option{
		Addr:     "localhost:3389",
		UserName: "test",
		Password: "test",
	})

	t.Run("ClipboardHandlerRegistration", func(t *testing.T) {
		handler := &testClipboardHandler{}
		err := client.RegisterClipboardHandler(handler)
		assert.NoError(t, err)
	})

	t.Run("ClipboardChannelStatus", func(t *testing.T) {
		// Should be false without connection
		assert.False(t, client.IsClipboardChannelOpen())
	})
}

// TestDeviceManagement tests device management
func TestDeviceManagement(t *testing.T) {
	client := NewClient(&Option{
		Addr:     "localhost:3389",
		UserName: "test",
		Password: "test",
	})

	t.Run("DeviceChannelStatus", func(t *testing.T) {
		// Should be false without connection
		assert.False(t, client.IsDeviceChannelOpen())
	})

	t.Run("DeviceCount", func(t *testing.T) {
		// Should be 0 without connection
		assert.Equal(t, 0, client.GetDeviceCount())
	})

	t.Run("DeviceList", func(t *testing.T) {
		// Should be empty without connection
		devices := client.ListDevices()
		assert.Empty(t, devices)
	})

	t.Run("DeviceStats", func(t *testing.T) {
		// Should return empty stats without connection
		stats := client.GetDeviceStats()
		assert.NotNil(t, stats)
		assert.IsType(t, map[string]interface{}{}, stats)
	})
}

// TestMonitorManagement tests monitor management
func TestMonitorManagement(t *testing.T) {
	client := NewClient(&Option{
		Addr:     "localhost:3389",
		UserName: "test",
		Password: "test",
	})

	t.Run("MonitorConfiguration", func(t *testing.T) {
		monitors := []mcs.MonitorLayout{
			{Left: 0, Top: 0, Right: 1920, Bottom: 1080, Flags: 0x01},
			{Left: 1920, Top: 0, Right: 3840, Bottom: 1080, Flags: 0x00},
		}

		client.SetMonitors(monitors)
		retrieved := client.GetMonitors()
		assert.Len(t, retrieved, 2)
		assert.Equal(t, int32(1920), retrieved[0].Right)
		assert.Equal(t, int32(1080), retrieved[0].Bottom)
		assert.Equal(t, uint32(0x01), retrieved[0].Flags)
	})

	t.Run("EmptyMonitorConfiguration", func(t *testing.T) {
		client.SetMonitors(nil)
		retrieved := client.GetMonitors()
		assert.Empty(t, retrieved)
	})
}

// TestContextManagement tests context management
func TestContextManagement(t *testing.T) {
	client := NewClient(&Option{
		Addr:     "localhost:3389",
		UserName: "test",
		Password: "test",
	})

	t.Run("ContextAccess", func(t *testing.T) {
		ctx := client.Context()
		assert.NotNil(t, ctx)

		done := client.Done()
		assert.NotNil(t, done)
	})

	t.Run("ContextCancellation", func(t *testing.T) {
		// Test that cancel doesn't panic
		client.Cancel()
		// Should not panic
	})
}

// TestBitmapCacheManagement tests bitmap cache management
func TestBitmapCacheManagement(t *testing.T) {
	client := NewClient(&Option{
		Addr:     "localhost:3389",
		UserName: "test",
		Password: "test",
	})

	t.Run("BitmapCacheStats", func(t *testing.T) {
		stats := client.GetBitmapCacheStats()
		assert.NotNil(t, stats)
		assert.IsType(t, map[string]interface{}{}, stats)
	})

	t.Run("BitmapCacheClear", func(t *testing.T) {
		// Should not panic
		client.ClearBitmapCache()
	})
}

// TestInputErrorHandling tests input error handling scenarios
func TestInputErrorHandling(t *testing.T) {
	t.Run("InvalidFunctionKey", func(t *testing.T) {
		client := NewClient(&Option{
			Addr:     "localhost:3389",
			UserName: "test",
			Password: "test",
		})

		err := client.SendFunctionKey(0, t128.ModifierKey{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "function key number must be between 1 and 24")

		err = client.SendFunctionKey(25, t128.ModifierKey{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "function key number must be between 1 and 24")
	})

	t.Run("InvalidArrowDirection", func(t *testing.T) {
		client := NewClient(&Option{
			Addr:     "localhost:3389",
			UserName: "test",
			Password: "test",
		})

		err := client.SendArrowKey("INVALID", t128.ModifierKey{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported arrow direction")
	})

	t.Run("InvalidNavigationKey", func(t *testing.T) {
		client := NewClient(&Option{
			Addr:     "localhost:3389",
			UserName: "test",
			Password: "test",
		})

		err := client.SendNavigationKey("INVALID", t128.ModifierKey{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported navigation key")
	})

	t.Run("InvalidMediaKey", func(t *testing.T) {
		client := NewClient(&Option{
			Addr:     "localhost:3389",
			UserName: "test",
			Password: "test",
		})

		err := client.SendMediaKey("INVALID")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported media key")
	})

	t.Run("InvalidBrowserKey", func(t *testing.T) {
		client := NewClient(&Option{
			Addr:     "localhost:3389",
			UserName: "test",
			Password: "test",
		})

		err := client.SendBrowserKey("INVALID")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported browser key")
	})

	t.Run("InvalidSpecialKey", func(t *testing.T) {
		client := NewClient(&Option{
			Addr:     "localhost:3389",
			UserName: "test",
			Password: "test",
		})

		err := client.SendSpecialKey("INVALID", t128.ModifierKey{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported special key")
	})

	t.Run("InvalidMouseButton", func(t *testing.T) {
		client := NewClient(&Option{
			Addr:     "localhost:3389",
			UserName: "test",
			Password: "test",
		})

		err := client.SendMouseButtonEvent(t128.MouseButton(999), true, 100, 100)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported mouse button")
	})
}

// TestDeviceOperations tests device operations
func TestDeviceOperations(t *testing.T) {
	client := NewClient(&Option{
		Addr:     "localhost:3389",
		UserName: "test",
		Password: "test",
	})

	t.Run("DeviceAnnouncement", func(t *testing.T) {
		err := client.AnnounceDevice(device.DeviceTypePrinter, "TEST_PRINTER", "test data")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no active connection")
	})

	t.Run("PrinterDataSending", func(t *testing.T) {
		err := client.SendPrinterData(1, []byte("test data"), 0)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no active connection")
	})
}

// Test helper structs
type testDVCHandler struct{}

func (h *testDVCHandler) OnChannelCreated(channelId uint32, channelName string) error {
	return nil
}

func (h *testDVCHandler) OnChannelOpened(channelId uint32) error {
	return nil
}

func (h *testDVCHandler) OnChannelClosed(channelId uint32) error {
	return nil
}

func (h *testDVCHandler) OnDataReceived(channelId uint32, data []byte) error {
	return nil
}

type testClipboardHandler struct{}

func (h *testClipboardHandler) OnFormatList(formats []clipboard.ClipboardFormat) error {
	return nil
}

func (h *testClipboardHandler) OnFormatDataRequest(formatID clipboard.ClipboardFormat) error {
	return nil
}

func (h *testClipboardHandler) OnFormatDataResponse(formatID clipboard.ClipboardFormat, data []byte) error {
	return nil
}

func (h *testClipboardHandler) OnFileContentsRequest(streamID uint32, listIndex uint32, dwFlags uint32, nPositionLow uint32, nPositionHigh uint32, cbRequested uint32, clipDataID uint32) error {
	return nil
}

type testDeviceHandler struct{}

func (h *testDeviceHandler) OnDeviceAnnounce(device *device.DeviceAnnounce) error {
	return nil
}

func (h *testDeviceHandler) OnDeviceIORequest(request *device.DeviceIORequest) (*device.DeviceIOCompletion, error) {
	return nil, nil
}

func (h *testDeviceHandler) OnDriveAccess(request *device.DeviceIORequest) (*device.DeviceIOCompletion, error) {
	return nil, nil
}

// TestProcessor tests the processor interface
func TestProcessor(t *testing.T) {
	t.Run("ProcessorInterface", func(t *testing.T) {
		processor := &testProcessor{}

		// Test that processor implements the interface
		var _ Processor = processor

		// Test bitmap processing
		option := &bitmap.Option{
			Width:  100,
			Height: 100,
			Left:   0,
			Top:    0,
		}

		// Create a simple test image
		img := image.NewRGBA(image.Rect(0, 0, 100, 100))
		bitmap := &bitmap.BitMap{
			Image: img,
		}

		// Should not panic
		processor.ProcessBitmap(option, bitmap)
		assert.Equal(t, 1, processor.processCount)
	})
}

type testProcessor struct {
	processCount int
}

func (p *testProcessor) ProcessBitmap(option *bitmap.Option, bitmap *bitmap.BitMap) {
	p.processCount++
}
