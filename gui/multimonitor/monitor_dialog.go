package multimonitor

import (
	"fmt"
	"strconv"
)

// MonitorDialog manages the multi-monitor configuration dialog
type MonitorDialog struct {
	manager *MonitorManager
}

// NewMonitorDialog creates a new monitor dialog
func NewMonitorDialog(manager *MonitorManager) *MonitorDialog {
	return &MonitorDialog{
		manager: manager,
	}
}

// Show displays the multi-monitor configuration dialog
func (d *MonitorDialog) Show() {
	fmt.Println("\n=== Multi-Monitor Configuration ===")

	for {
		fmt.Println("\n1. Detect Monitors")
		fmt.Println("2. Monitor List")
		fmt.Println("3. Set Primary Monitor")
		fmt.Println("4. Enable/Disable Monitors")
		fmt.Println("5. Display Spanning")
		fmt.Println("6. Monitor Layout")
		fmt.Println("7. Back to Main Menu")

		fmt.Print("\nSelect option (1-7): ")
		var choice string
		fmt.Scanln(&choice)

		switch choice {
		case "1":
			d.detectMonitors()
		case "2":
			d.showMonitorList()
		case "3":
			d.setPrimaryMonitor()
		case "4":
			d.enableDisableMonitors()
		case "5":
			d.configureSpanning()
		case "6":
			d.showMonitorLayout()
		case "7":
			return
		default:
			fmt.Println("Invalid option. Please try again.")
		}
	}
}

// detectMonitors detects available monitors
func (d *MonitorDialog) detectMonitors() {
	fmt.Println("\n--- Detecting Monitors ---")

	if err := d.manager.DetectMonitors(); err != nil {
		fmt.Printf("Error detecting monitors: %v\n", err)
		return
	}

	monitors := d.manager.GetMonitors()
	fmt.Printf("Successfully detected %d monitors\n", len(monitors))

	for _, monitor := range monitors {
		fmt.Printf("  - %s (%dx%d at %d,%d)\n",
			monitor.Name, monitor.Width, monitor.Height, monitor.X, monitor.Y)
	}
}

// showMonitorList displays the list of monitors
func (d *MonitorDialog) showMonitorList() {
	monitors := d.manager.GetMonitors()

	if len(monitors) == 0 {
		fmt.Println("\n--- Monitor List ---")
		fmt.Println("No monitors detected. Run 'Detect Monitors' first.")
		return
	}

	fmt.Println("\n--- Monitor List ---")
	for i, monitor := range monitors {
		status := "Enabled"
		if !monitor.Enabled {
			status = "Disabled"
		}

		primary := ""
		if monitor.Primary {
			primary = " (Primary)"
		}

		selected := ""
		if d.manager.GetSelectedMonitor() != nil && d.manager.GetSelectedMonitor().ID == monitor.ID {
			selected = " [Selected]"
		}

		fmt.Printf("%d. %s%s%s\n", i+1, monitor.Name, primary, selected)
		fmt.Printf("   Resolution: %dx%d\n", monitor.Width, monitor.Height)
		fmt.Printf("   Position: (%d, %d)\n", monitor.X, monitor.Y)
		fmt.Printf("   Status: %s\n", status)
		fmt.Println()
	}
}

// setPrimaryMonitor allows setting the primary monitor
func (d *MonitorDialog) setPrimaryMonitor() {
	monitors := d.manager.GetMonitors()

	if len(monitors) == 0 {
		fmt.Println("No monitors available. Run 'Detect Monitors' first.")
		return
	}

	fmt.Println("\n--- Set Primary Monitor ---")
	for i, monitor := range monitors {
		primary := ""
		if monitor.Primary {
			primary = " (Current Primary)"
		}
		fmt.Printf("%d. %s%s\n", i+1, monitor.Name, primary)
	}

	fmt.Print("\nSelect monitor (1-" + strconv.Itoa(len(monitors)) + "): ")
	var choice string
	fmt.Scanln(&choice)

	index, err := strconv.Atoi(choice)
	if err != nil || index < 1 || index > len(monitors) {
		fmt.Println("Invalid selection.")
		return
	}

	monitor := monitors[index-1]
	if err := d.manager.SetPrimaryMonitor(monitor.ID); err != nil {
		fmt.Printf("Error setting primary monitor: %v\n", err)
		return
	}

	fmt.Printf("Primary monitor set to: %s\n", monitor.Name)
}

// enableDisableMonitors allows enabling/disabling monitors
func (d *MonitorDialog) enableDisableMonitors() {
	monitors := d.manager.GetMonitors()

	if len(monitors) == 0 {
		fmt.Println("No monitors available. Run 'Detect Monitors' first.")
		return
	}

	fmt.Println("\n--- Enable/Disable Monitors ---")
	for i, monitor := range monitors {
		status := "Enabled"
		if !monitor.Enabled {
			status = "Disabled"
		}
		fmt.Printf("%d. %s (%s)\n", i+1, monitor.Name, status)
	}

	fmt.Print("\nSelect monitor (1-" + strconv.Itoa(len(monitors)) + "): ")
	var choice string
	fmt.Scanln(&choice)

	index, err := strconv.Atoi(choice)
	if err != nil || index < 1 || index > len(monitors) {
		fmt.Println("Invalid selection.")
		return
	}

	monitor := monitors[index-1]
	enabled := !monitor.Enabled

	if err := d.manager.EnableMonitor(monitor.ID, enabled); err != nil {
		fmt.Printf("Error updating monitor: %v\n", err)
		return
	}

	status := "enabled"
	if !enabled {
		status = "disabled"
	}
	fmt.Printf("Monitor %s %s\n", monitor.Name, status)
}

// configureSpanning configures display spanning
func (d *MonitorDialog) configureSpanning() {
	fmt.Println("\n--- Display Spanning Configuration ---")

	current := d.manager.IsSpanning()
	fmt.Printf("Current spanning: %s\n", d.getEnabledStatus(current))

	fmt.Println("\n1. Enable spanning")
	fmt.Println("2. Disable spanning")
	fmt.Println("3. Show spanning info")
	fmt.Println("4. Back")

	fmt.Print("\nSelect option (1-4): ")
	var choice string
	fmt.Scanln(&choice)

	switch choice {
	case "1":
		d.manager.SetSpanning(true)
		fmt.Println("Display spanning enabled")
	case "2":
		d.manager.SetSpanning(false)
		fmt.Println("Display spanning disabled")
	case "3":
		d.showSpanningInfo()
	case "4":
		return
	default:
		fmt.Println("Invalid option.")
	}
}

// showSpanningInfo displays information about spanning
func (d *MonitorDialog) showSpanningInfo() {
	spanning := d.manager.IsSpanning()
	monitors := d.manager.GetMonitors()

	fmt.Println("\n--- Spanning Information ---")
	fmt.Printf("Spanning: %s\n", d.getEnabledStatus(spanning))

	if spanning {
		width, height := d.manager.GetTotalResolution()
		fmt.Printf("Total resolution: %dx%d\n", width, height)

		fmt.Println("\nMonitor layout:")
		for _, monitor := range monitors {
			if !monitor.Enabled {
				continue
			}
			fmt.Printf("  - %s: %dx%d at (%d,%d)\n",
				monitor.Name, monitor.Width, monitor.Height, monitor.X, monitor.Y)
		}
	} else {
		selected := d.manager.GetSelectedMonitor()
		if selected != nil {
			fmt.Printf("Selected monitor: %s (%dx%d)\n",
				selected.Name, selected.Width, selected.Height)
		}
	}
}

// showMonitorLayout displays the current monitor layout
func (d *MonitorDialog) showMonitorLayout() {
	layout := d.manager.GetLayout()

	fmt.Println("\n--- Monitor Layout ---")
	fmt.Printf("Total monitors: %d\n", len(layout.Monitors))
	fmt.Printf("Spanning: %s\n", d.getEnabledStatus(layout.Spanning))

	if len(layout.Monitors) > 0 {
		primary := layout.Monitors[layout.Primary]
		fmt.Printf("Primary monitor: %s\n", primary.Name)

		fmt.Println("\nMonitor details:")
		for i, monitor := range layout.Monitors {
			primary := ""
			if monitor.Primary {
				primary = " (Primary)"
			}

			status := "Enabled"
			if !monitor.Enabled {
				status = "Disabled"
			}

			fmt.Printf("%d. %s%s\n", i+1, monitor.Name, primary)
			fmt.Printf("   Resolution: %dx%d\n", monitor.Width, monitor.Height)
			fmt.Printf("   Position: (%d, %d)\n", monitor.X, monitor.Y)
			fmt.Printf("   Status: %s\n", status)
			fmt.Println()
		}

		if layout.Spanning {
			width, height := d.manager.GetTotalResolution()
			fmt.Printf("Total spanning resolution: %dx%d\n", width, height)
		}
	}

	fmt.Println("Press Enter to continue...")
	var input string
	fmt.Scanln(&input)
}

// getEnabledStatus returns a string representation of enabled/disabled status
func (d *MonitorDialog) getEnabledStatus(enabled bool) string {
	if enabled {
		return "Enabled"
	}
	return "Disabled"
}
