package multimonitor

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Monitor represents a display monitor
type Monitor struct {
	ID      int
	Name    string
	X, Y    int
	Width   int
	Height  int
	Primary bool
	Enabled bool
	// Additional monitor properties
	RefreshRate    int    // Hz
	ColorDepth     int    // bits per pixel
	DPI            int    // dots per inch
	Manufacturer   string // monitor manufacturer
	Model          string // monitor model
	SerialNumber   string // monitor serial number
	ConnectionType string // HDMI, DisplayPort, VGA, etc.
}

// MonitorLayout represents the layout of monitors
type MonitorLayout struct {
	Monitors []*Monitor
	Primary  int
	Spanning bool
}

// MonitorManager manages multi-monitor configurations
type MonitorManager struct {
	mu sync.RWMutex

	// Monitor configuration
	layout    *MonitorLayout
	selected  int // Currently selected monitor
	isEnabled bool

	// Enhanced monitor management
	detectionStats map[string]interface{}
	lastDetection  time.Time
	detectionError error
	platform       string
}

// NewMonitorManager creates a new monitor manager
func NewMonitorManager() *MonitorManager {
	manager := &MonitorManager{
		layout: &MonitorLayout{
			Monitors: []*Monitor{},
			Primary:  0,
			Spanning: false,
		},
		selected:       0,
		isEnabled:      true,
		detectionStats: make(map[string]interface{}),
		platform:       runtime.GOOS,
		lastDetection:  time.Time{},
		detectionError: nil,
	}

	// Initialize statistics
	manager.initializeStats()

	return manager
}

// DetectMonitors detects available monitors using platform-specific APIs
func (m *MonitorManager) DetectMonitors() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.detectionError = nil
	startTime := time.Now()

	var monitors []*Monitor
	var err error

	// Use platform-specific detection
	switch m.platform {
	case "linux":
		monitors, err = m.detectLinuxMonitors()
	case "windows":
		monitors, err = m.detectWindowsMonitors()
	case "darwin":
		monitors, err = m.detectDarwinMonitors()
	default:
		// Fallback to generic detection
		monitors, err = m.detectGenericMonitors()
	}

	if err != nil {
		m.detectionError = err
		m.updateDetectionStats("errors", 1)
		m.detectionStats["last_error"] = err.Error()
		return fmt.Errorf("monitor detection failed: %v", err)
	}

	// Validate detected monitors
	if len(monitors) == 0 {
		// Fallback to basic detection if no monitors found
		monitors = m.detectBasicMonitors()
	}

	// Update layout
	m.layout.Monitors = monitors
	m.layout.Primary = m.findPrimaryMonitor(monitors)
	m.lastDetection = time.Now()

	// Update statistics
	m.updateDetectionStats("detections", 1)
	m.updateDetectionStats("monitors_found", len(monitors))
	m.updateDetectionStats("detection_time_ms", int(time.Since(startTime).Milliseconds()))

	fmt.Printf("Detected %d monitors on %s\n", len(monitors), m.platform)
	return nil
}

// detectLinuxMonitors detects monitors on Linux systems
func (m *MonitorManager) detectLinuxMonitors() ([]*Monitor, error) {
	var monitors []*Monitor

	// Try X11 first (most common)
	if x11Monitors, err := m.detectX11Monitors(); err == nil && len(x11Monitors) > 0 {
		return x11Monitors, nil
	}

	// Try Wayland
	if waylandMonitors, err := m.detectWaylandMonitors(); err == nil && len(waylandMonitors) > 0 {
		return waylandMonitors, nil
	}

	// Try DRM/KMS
	if drmMonitors, err := m.detectDRMMonitors(); err == nil && len(drmMonitors) > 0 {
		return drmMonitors, nil
	}

	// Try xrandr as fallback
	if xrandrMonitors, err := m.detectXrandrMonitors(); err == nil && len(xrandrMonitors) > 0 {
		return xrandrMonitors, nil
	}

	return monitors, fmt.Errorf("no monitor detection method worked")
}

// detectX11Monitors detects monitors using X11
func (m *MonitorManager) detectX11Monitors() ([]*Monitor, error) {
	// Check if X11 is available
	if !m.commandExists("xprop") {
		return nil, fmt.Errorf("xprop not available")
	}

	// Get display info using xprop
	cmd := exec.Command("xprop", "-root", "_NET_DESKTOP_GEOMETRY")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	// Parse output (format: _NET_DESKTOP_GEOMETRY(CARDINAL) = 1920, 1080)
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "_NET_DESKTOP_GEOMETRY") {
			// Extract dimensions
			parts := strings.Split(line, "=")
			if len(parts) >= 2 {
				dimensions := strings.TrimSpace(parts[1])
				dimensions = strings.Trim(dimensions, "()")
				coords := strings.Split(dimensions, ",")
				if len(coords) >= 2 {
					width, _ := strconv.Atoi(strings.TrimSpace(coords[0]))
					height, _ := strconv.Atoi(strings.TrimSpace(coords[1]))

					monitor := &Monitor{
						ID:          0,
						Name:        "X11 Display",
						X:           0,
						Y:           0,
						Width:       width,
						Height:      height,
						Primary:     true,
						Enabled:     true,
						RefreshRate: 60,
						ColorDepth:  24,
						DPI:         96,
					}
					return []*Monitor{monitor}, nil
				}
			}
		}
	}

	return nil, fmt.Errorf("could not parse X11 display info")
}

// detectWaylandMonitors detects monitors using Wayland
func (m *MonitorManager) detectWaylandMonitors() ([]*Monitor, error) {
	// Check if Wayland is available
	if os.Getenv("WAYLAND_DISPLAY") == "" {
		return nil, fmt.Errorf("Wayland not detected")
	}

	// Try using wlr-randr if available
	if m.commandExists("wlr-randr") {
		return m.detectWlrRandrMonitors()
	}

	// Try using swaymsg if available
	if m.commandExists("swaymsg") {
		return m.detectSwayMonitors()
	}

	return nil, fmt.Errorf("no Wayland monitor detection tools available")
}

// detectWlrRandrMonitors detects monitors using wlr-randr
func (m *MonitorManager) detectWlrRandrMonitors() ([]*Monitor, error) {
	cmd := exec.Command("wlr-randr")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var monitors []*Monitor
	lines := strings.Split(string(output), "\n")
	currentMonitor := &Monitor{}

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Check if this is a monitor name line
		if !strings.Contains(line, " ") && !strings.Contains(line, "x") {
			// This is likely a monitor name
			if currentMonitor.Name != "" {
				monitors = append(monitors, currentMonitor)
			}
			currentMonitor = &Monitor{
				ID:      len(monitors),
				Name:    line,
				Enabled: true,
			}
		} else if strings.Contains(line, "x") && strings.Contains(line, "@") {
			// Parse resolution and refresh rate
			parts := strings.Fields(line)
			if len(parts) >= 1 {
				resolution := strings.Split(parts[0], "x")
				if len(resolution) == 2 {
					width, _ := strconv.Atoi(resolution[0])
					height, _ := strconv.Atoi(resolution[1])
					currentMonitor.Width = width
					currentMonitor.Height = height
				}
			}
		}
	}

	// Add the last monitor
	if currentMonitor.Name != "" {
		monitors = append(monitors, currentMonitor)
	}

	if len(monitors) == 0 {
		return nil, fmt.Errorf("no monitors detected")
	}

	// Set first monitor as primary
	if len(monitors) > 0 {
		monitors[0].Primary = true
	}

	return monitors, nil
}

// detectSwayMonitors detects monitors using swaymsg
func (m *MonitorManager) detectSwayMonitors() ([]*Monitor, error) {
	cmd := exec.Command("swaymsg", "-t", "get_outputs")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	// Parse JSON output (simplified)
	var monitors []*Monitor
	lines := strings.Split(string(output), "\n")

	for i, line := range lines {
		if strings.Contains(line, `"name"`) {
			// Extract monitor name
			name := strings.Trim(strings.Split(line, `"name"`)[1], `":, `)
			monitor := &Monitor{
				ID:      i,
				Name:    name,
				Enabled: true,
			}

			// Look for resolution in subsequent lines
			for j := i + 1; j < len(lines) && j < i+10; j++ {
				if strings.Contains(lines[j], `"width"`) {
					widthStr := strings.Trim(strings.Split(lines[j], `"width"`)[1], `":, `)
					width, _ := strconv.Atoi(widthStr)
					monitor.Width = width
				}
				if strings.Contains(lines[j], `"height"`) {
					heightStr := strings.Trim(strings.Split(lines[j], `"height"`)[1], `":, `)
					height, _ := strconv.Atoi(heightStr)
					monitor.Height = height
				}
			}

			monitors = append(monitors, monitor)
		}
	}

	if len(monitors) == 0 {
		return nil, fmt.Errorf("no monitors detected")
	}

	// Set first monitor as primary
	if len(monitors) > 0 {
		monitors[0].Primary = true
	}

	return monitors, nil
}

// detectDRMMonitors detects monitors using DRM/KMS
func (m *MonitorManager) detectDRMMonitors() ([]*Monitor, error) {
	// Try using drm_info if available
	if m.commandExists("drm_info") {
		return m.detectDrmInfoMonitors()
	}

	// Try reading from /sys/class/drm
	return m.detectSysDRMMonitors()
}

// detectDrmInfoMonitors detects monitors using drm_info
func (m *MonitorManager) detectDrmInfoMonitors() ([]*Monitor, error) {
	cmd := exec.Command("drm_info", "--json")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	// Parse JSON output (simplified)
	var monitors []*Monitor
	lines := strings.Split(string(output), "\n")

	for i, line := range lines {
		if strings.Contains(line, `"name"`) && strings.Contains(line, `"card"`) {
			// Extract card name
			parts := strings.Split(line, `"name"`)
			if len(parts) >= 2 {
				name := strings.Trim(strings.Split(parts[1], `"`)[1], `":, `)
				monitor := &Monitor{
					ID:      i,
					Name:    fmt.Sprintf("DRM Card %s", name),
					Enabled: true,
					Width:   1920, // Default values
					Height:  1080,
				}
				monitors = append(monitors, monitor)
			}
		}
	}

	if len(monitors) == 0 {
		return nil, fmt.Errorf("no DRM monitors detected")
	}

	// Set first monitor as primary
	if len(monitors) > 0 {
		monitors[0].Primary = true
	}

	return monitors, nil
}

// detectSysDRMMonitors detects monitors by reading /sys/class/drm
func (m *MonitorManager) detectSysDRMMonitors() ([]*Monitor, error) {
	// Read /sys/class/drm directory
	entries, err := os.ReadDir("/sys/class/drm")
	if err != nil {
		return nil, err
	}

	var monitors []*Monitor
	for i, entry := range entries {
		if strings.HasPrefix(entry.Name(), "card") && entry.IsDir() {
			// Check if this card has a connected monitor
			statusPath := fmt.Sprintf("/sys/class/drm/%s/status", entry.Name())
			statusData, err := os.ReadFile(statusPath)
			if err != nil {
				continue
			}

			status := strings.TrimSpace(string(statusData))
			if status == "connected" {
				monitor := &Monitor{
					ID:      i,
					Name:    fmt.Sprintf("DRM %s", entry.Name()),
					Enabled: true,
					Width:   1920, // Default values
					Height:  1080,
				}
				monitors = append(monitors, monitor)
			}
		}
	}

	if len(monitors) == 0 {
		return nil, fmt.Errorf("no connected DRM monitors detected")
	}

	// Set first monitor as primary
	if len(monitors) > 0 {
		monitors[0].Primary = true
	}

	return monitors, nil
}

// detectXrandrMonitors detects monitors using xrandr
func (m *MonitorManager) detectXrandrMonitors() ([]*Monitor, error) {
	cmd := exec.Command("xrandr", "--listmonitors")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var monitors []*Monitor
	lines := strings.Split(string(output), "\n")

	for i, line := range lines {
		if strings.Contains(line, "x") && strings.Contains(line, "/") {
			// Parse monitor line (format: " 0: +*HDMI-1 1920/509x1080/286+0+0  HDMI-1")
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				resolution := strings.Split(parts[1], "x")
				if len(resolution) == 2 {
					width, _ := strconv.Atoi(strings.Split(resolution[0], "/")[0])
					height, _ := strconv.Atoi(strings.Split(resolution[1], "/")[0])

					name := "Unknown"
					if len(parts) >= 4 {
						name = parts[3]
					}

					monitor := &Monitor{
						ID:      i,
						Name:    name,
						X:       0,
						Y:       0,
						Width:   width,
						Height:  height,
						Primary: i == 0,
						Enabled: true,
					}
					monitors = append(monitors, monitor)
				}
			}
		}
	}

	if len(monitors) == 0 {
		return nil, fmt.Errorf("no xrandr monitors detected")
	}

	return monitors, nil
}

// detectWindowsMonitors detects monitors on Windows systems
func (m *MonitorManager) detectWindowsMonitors() ([]*Monitor, error) {
	// Use PowerShell to get monitor information
	script := `
	Add-Type -AssemblyName System.Windows.Forms
	$screens = [System.Windows.Forms.Screen]::AllScreens
	$monitors = @()
	
	for ($i = 0; $i -lt $screens.Length; $i++) {
		$screen = $screens[$i]
		$monitor = @{
			ID = $i
			Name = "Monitor $($i + 1)"
			X = $screen.Bounds.X
			Y = $screen.Bounds.Y
			Width = $screen.Bounds.Width
			Height = $screen.Bounds.Height
			Primary = $screen.Primary
			Enabled = $true
		}
		$monitors += $monitor
	}
	
	$monitors | ConvertTo-Json
	`

	cmd := exec.Command("powershell", "-Command", script)
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	// Parse JSON output (simplified)
	var monitors []*Monitor
	lines := strings.Split(string(output), "\n")

	for _, line := range lines {
		if strings.Contains(line, `"ID"`) {
			// Extract monitor info
			monitor := &Monitor{
				ID:      len(monitors),
				Name:    fmt.Sprintf("Windows Monitor %d", len(monitors)+1),
				Enabled: true,
			}

			// Parse other fields (simplified)
			if strings.Contains(line, `"Primary"`) && strings.Contains(line, `true`) {
				monitor.Primary = true
			}

			monitors = append(monitors, monitor)
		}
	}

	if len(monitors) == 0 {
		return nil, fmt.Errorf("no Windows monitors detected")
	}

	return monitors, nil
}

// detectDarwinMonitors detects monitors on macOS systems
func (m *MonitorManager) detectDarwinMonitors() ([]*Monitor, error) {
	// Use system_profiler to get display information
	cmd := exec.Command("system_profiler", "SPDisplaysDataType")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var monitors []*Monitor
	lines := strings.Split(string(output), "\n")
	currentMonitor := &Monitor{}

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "Display:") {
			// Start of a new monitor
			if currentMonitor.Name != "" {
				monitors = append(monitors, currentMonitor)
			}
			currentMonitor = &Monitor{
				ID:      len(monitors),
				Name:    strings.TrimPrefix(line, "Display: "),
				Enabled: true,
			}
		} else if strings.Contains(line, "Resolution:") {
			// Parse resolution
			resolution := strings.TrimPrefix(line, "Resolution: ")
			parts := strings.Fields(resolution)
			if len(parts) >= 1 {
				res := strings.Split(parts[0], "x")
				if len(res) == 2 {
					width, _ := strconv.Atoi(res[0])
					height, _ := strconv.Atoi(res[1])
					currentMonitor.Width = width
					currentMonitor.Height = height
				}
			}
		}
	}

	// Add the last monitor
	if currentMonitor.Name != "" {
		monitors = append(monitors, currentMonitor)
	}

	if len(monitors) == 0 {
		return nil, fmt.Errorf("no macOS monitors detected")
	}

	// Set first monitor as primary
	if len(monitors) > 0 {
		monitors[0].Primary = true
	}

	return monitors, nil
}

// detectGenericMonitors provides a generic fallback detection
func (m *MonitorManager) detectGenericMonitors() ([]*Monitor, error) {
	// Try to get screen size using environment variables
	width := 1920
	height := 1080

	if w := os.Getenv("DISPLAY_WIDTH"); w != "" {
		if wInt, err := strconv.Atoi(w); err == nil {
			width = wInt
		}
	}
	if h := os.Getenv("DISPLAY_HEIGHT"); h != "" {
		if hInt, err := strconv.Atoi(h); err == nil {
			height = hInt
		}
	}

	monitor := &Monitor{
		ID:          0,
		Name:        fmt.Sprintf("Generic %s Display", m.platform),
		X:           0,
		Y:           0,
		Width:       width,
		Height:      height,
		Primary:     true,
		Enabled:     true,
		RefreshRate: 60,
		ColorDepth:  24,
		DPI:         96,
	}

	return []*Monitor{monitor}, nil
}

// detectBasicMonitors provides basic fallback detection
func (m *MonitorManager) detectBasicMonitors() []*Monitor {
	return []*Monitor{
		{
			ID:          0,
			Name:        fmt.Sprintf("Fallback %s Display", m.platform),
			X:           0,
			Y:           0,
			Width:       1920,
			Height:      1080,
			Primary:     true,
			Enabled:     true,
			RefreshRate: 60,
			ColorDepth:  24,
			DPI:         96,
		},
	}
}

// findPrimaryMonitor finds the primary monitor in the list
func (m *MonitorManager) findPrimaryMonitor(monitors []*Monitor) int {
	for i, monitor := range monitors {
		if monitor.Primary {
			return i
		}
	}
	return 0 // Default to first monitor
}

// commandExists checks if a command exists
func (m *MonitorManager) commandExists(command string) bool {
	_, err := exec.LookPath(command)
	return err == nil
}

// initializeStats initializes detection statistics
func (m *MonitorManager) initializeStats() {
	m.detectionStats["total_detections"] = 0
	m.detectionStats["total_errors"] = 0
	m.detectionStats["total_monitors_found"] = 0
	m.detectionStats["average_detection_time_ms"] = 0
	m.detectionStats["platform"] = m.platform
	m.detectionStats["start_time"] = time.Now()
}

// updateDetectionStats updates detection statistics
func (m *MonitorManager) updateDetectionStats(key string, value int) {
	if current, exists := m.detectionStats[key]; exists {
		if intValue, ok := current.(int); ok {
			m.detectionStats[key] = intValue + value
		}
	} else {
		m.detectionStats[key] = value
	}
}

// GetDetectionStats returns detection statistics
func (m *MonitorManager) GetDetectionStats() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := make(map[string]interface{})
	for k, v := range m.detectionStats {
		stats[k] = v
	}

	// Add current state
	stats["last_detection"] = m.lastDetection
	stats["last_error"] = m.detectionError
	stats["monitors_count"] = len(m.layout.Monitors)
	stats["platform"] = m.platform

	return stats
}

// GetMonitors returns all detected monitors
func (m *MonitorManager) GetMonitors() []*Monitor {
	m.mu.RLock()
	defer m.mu.RUnlock()

	monitors := make([]*Monitor, len(m.layout.Monitors))
	copy(monitors, m.layout.Monitors)
	return monitors
}

// GetMonitor returns a specific monitor by ID
func (m *MonitorManager) GetMonitor(id int) *Monitor {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, monitor := range m.layout.Monitors {
		if monitor.ID == id {
			return monitor
		}
	}
	return nil
}

// GetPrimaryMonitor returns the primary monitor
func (m *MonitorManager) GetPrimaryMonitor() *Monitor {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(m.layout.Monitors) == 0 {
		return nil
	}

	return m.layout.Monitors[m.layout.Primary]
}

// GetSelectedMonitor returns the currently selected monitor
func (m *MonitorManager) GetSelectedMonitor() *Monitor {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(m.layout.Monitors) == 0 {
		return nil
	}

	return m.layout.Monitors[m.selected]
}

// SetPrimaryMonitor sets the primary monitor
func (m *MonitorManager) SetPrimaryMonitor(id int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Find monitor
	var found bool
	for i, monitor := range m.layout.Monitors {
		if monitor.ID == id {
			// Update primary flags
			for _, m := range m.layout.Monitors {
				m.Primary = false
			}
			monitor.Primary = true
			m.layout.Primary = i
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("monitor with ID %d not found", id)
	}

	fmt.Printf("Primary monitor set to: %s\n", m.layout.Monitors[m.layout.Primary].Name)
	return nil
}

// SetSelectedMonitor sets the currently selected monitor
func (m *MonitorManager) SetSelectedMonitor(id int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i, monitor := range m.layout.Monitors {
		if monitor.ID == id {
			m.selected = i
			fmt.Printf("Selected monitor: %s\n", monitor.Name)
			return nil
		}
	}

	return fmt.Errorf("monitor with ID %d not found", id)
}

// EnableMonitor enables or disables a monitor
func (m *MonitorManager) EnableMonitor(id int, enabled bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, monitor := range m.layout.Monitors {
		if monitor.ID == id {
			monitor.Enabled = enabled
			status := "enabled"
			if !enabled {
				status = "disabled"
			}
			fmt.Printf("Monitor %s %s\n", monitor.Name, status)
			return nil
		}
	}

	return fmt.Errorf("monitor with ID %d not found", id)
}

// SetSpanning enables or disables display spanning
func (m *MonitorManager) SetSpanning(spanning bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.layout.Spanning = spanning
	status := "enabled"
	if !spanning {
		status = "disabled"
	}
	fmt.Printf("Display spanning %s\n", status)
}

// IsSpanning returns whether display spanning is enabled
func (m *MonitorManager) IsSpanning() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.layout.Spanning
}

// IsEnabled returns whether multi-monitor support is enabled
func (m *MonitorManager) IsEnabled() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.isEnabled
}

// SetEnabled enables or disables multi-monitor support
func (m *MonitorManager) SetEnabled(enabled bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.isEnabled = enabled
	status := "enabled"
	if !enabled {
		status = "disabled"
	}
	fmt.Printf("Multi-monitor support %s\n", status)
}

// GetTotalResolution returns the total resolution when spanning
func (m *MonitorManager) GetTotalResolution() (width, height int) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if !m.layout.Spanning || len(m.layout.Monitors) == 0 {
		if len(m.layout.Monitors) > 0 {
			monitor := m.layout.Monitors[m.selected]
			return monitor.Width, monitor.Height
		}
		return 0, 0
	}

	// Calculate total resolution
	minX, maxX := 0, 0
	minY, maxY := 0, 0

	for _, monitor := range m.layout.Monitors {
		if !monitor.Enabled {
			continue
		}

		if monitor.X < minX {
			minX = monitor.X
		}
		if monitor.X+monitor.Width > maxX {
			maxX = monitor.X + monitor.Width
		}
		if monitor.Y < minY {
			minY = monitor.Y
		}
		if monitor.Y+monitor.Height > maxY {
			maxY = monitor.Y + monitor.Height
		}
	}

	return maxX - minX, maxY - minY
}

// GetMonitorAtPosition returns the monitor at the specified position
func (m *MonitorManager) GetMonitorAtPosition(x, y int) *Monitor {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, monitor := range m.layout.Monitors {
		if !monitor.Enabled {
			continue
		}

		if x >= monitor.X && x < monitor.X+monitor.Width &&
			y >= monitor.Y && y < monitor.Y+monitor.Height {
			return monitor
		}
	}

	return nil
}

// GetLayout returns the current monitor layout
func (m *MonitorManager) GetLayout() *MonitorLayout {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Return a copy to avoid race conditions
	layout := &MonitorLayout{
		Monitors: make([]*Monitor, len(m.layout.Monitors)),
		Primary:  m.layout.Primary,
		Spanning: m.layout.Spanning,
	}

	for i, monitor := range m.layout.Monitors {
		layout.Monitors[i] = &Monitor{
			ID:             monitor.ID,
			Name:           monitor.Name,
			X:              monitor.X,
			Y:              monitor.Y,
			Width:          monitor.Width,
			Height:         monitor.Height,
			Primary:        monitor.Primary,
			Enabled:        monitor.Enabled,
			RefreshRate:    monitor.RefreshRate,
			ColorDepth:     monitor.ColorDepth,
			DPI:            monitor.DPI,
			Manufacturer:   monitor.Manufacturer,
			Model:          monitor.Model,
			SerialNumber:   monitor.SerialNumber,
			ConnectionType: monitor.ConnectionType,
		}
	}

	return layout
}
