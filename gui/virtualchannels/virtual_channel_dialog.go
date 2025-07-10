package virtualchannels

import (
	"fmt"
)

// VirtualChannelDialog manages the virtual channel settings dialog
type VirtualChannelDialog struct {
	manager *VirtualChannelManager
}

// NewVirtualChannelDialog creates a new virtual channel dialog
func NewVirtualChannelDialog(manager *VirtualChannelManager) *VirtualChannelDialog {
	return &VirtualChannelDialog{
		manager: manager,
	}
}

// Show displays the virtual channel settings dialog
func (d *VirtualChannelDialog) Show() {
	fmt.Println("\n=== Virtual Channel Settings ===")

	for {
		fmt.Println("\n1. Clipboard Settings")
		fmt.Println("2. Audio Settings")
		fmt.Println("3. Device Redirection Settings")
		fmt.Println("4. Channel Status")
		fmt.Println("5. Back to Main Menu")

		fmt.Print("\nSelect option (1-5): ")
		var choice string
		fmt.Scanln(&choice)

		switch choice {
		case "1":
			d.showClipboardSettings()
		case "2":
			d.showAudioSettings()
		case "3":
			d.showDeviceSettings()
		case "4":
			d.showChannelStatus()
		case "5":
			return
		default:
			fmt.Println("Invalid option. Please try again.")
		}
	}
}

// showClipboardSettings displays clipboard synchronization settings
func (d *VirtualChannelDialog) showClipboardSettings() {
	clipboardHandler := d.manager.GetClipboardHandler()

	fmt.Println("\n--- Clipboard Settings ---")
	fmt.Printf("Synchronization: %s\n", d.getEnabledStatus(clipboardHandler.IsEnabled()))
	fmt.Printf("Local clipboard: %d bytes\n", len(clipboardHandler.GetLocalClipboard()))
	fmt.Printf("Remote clipboard: %d bytes\n", len(clipboardHandler.GetRemoteClipboard()))

	fmt.Println("\n1. Enable/Disable synchronization")
	fmt.Println("2. Sync clipboard now")
	fmt.Println("3. Set local clipboard content")
	fmt.Println("4. Back")

	fmt.Print("\nSelect option (1-4): ")
	var choice string
	fmt.Scanln(&choice)

	switch choice {
	case "1":
		enabled := !clipboardHandler.IsEnabled()
		clipboardHandler.SetEnabled(enabled)
		fmt.Printf("Clipboard synchronization %s\n", d.getEnabledStatus(enabled))
	case "2":
		clipboardHandler.SyncClipboard()
		fmt.Println("Clipboard synchronization triggered")
	case "3":
		fmt.Print("Enter clipboard content: ")
		var content string
		fmt.Scanln(&content)
		clipboardHandler.SetLocalClipboard(content)
		fmt.Println("Local clipboard updated")
	case "4":
		return
	default:
		fmt.Println("Invalid option.")
	}
}

// showAudioSettings displays audio redirection settings
func (d *VirtualChannelDialog) showAudioSettings() {
	audioHandler := d.manager.GetAudioHandler()

	fmt.Println("\n--- Audio Settings ---")
	fmt.Printf("Redirection: %s\n", d.getEnabledStatus(audioHandler.IsEnabled()))
	fmt.Printf("Playing: %s\n", d.getEnabledStatus(audioHandler.IsPlaying()))
	fmt.Printf("Volume: %.0f%%\n", audioHandler.GetVolume()*100)

	fmt.Println("\n1. Enable/Disable audio redirection")
	fmt.Println("2. Adjust volume")
	fmt.Println("3. Play/Pause audio")
	fmt.Println("4. Audio statistics")
	fmt.Println("5. Back")

	fmt.Print("\nSelect option (1-5): ")
	var choice string
	fmt.Scanln(&choice)

	switch choice {
	case "1":
		enabled := !audioHandler.IsEnabled()
		audioHandler.SetEnabled(enabled)
		fmt.Printf("Audio redirection %s\n", d.getEnabledStatus(enabled))
	case "2":
		fmt.Print("Enter volume (0-100): ")
		var volume int
		fmt.Scanln(&volume)
		if volume >= 0 && volume <= 100 {
			audioHandler.SetVolume(float64(volume) / 100.0)
		} else {
			fmt.Println("Invalid volume level.")
		}
	case "3":
		if audioHandler.IsPlaying() {
			audioHandler.PauseAudio()
		} else {
			audioHandler.ResumeAudio()
		}
	case "4":
		stats := audioHandler.GetAudioStats()
		fmt.Println("\n--- Audio Statistics ---")
		for key, value := range stats {
			fmt.Printf("%s: %v\n", key, value)
		}
	case "5":
		return
	default:
		fmt.Println("Invalid option.")
	}
}

// showDeviceSettings displays device redirection settings
func (d *VirtualChannelDialog) showDeviceSettings() {
	deviceHandler := d.manager.GetDeviceHandler()

	fmt.Println("\n--- Device Redirection Settings ---")
	fmt.Printf("Redirection: %s\n", d.getEnabledStatus(deviceHandler.IsEnabled()))

	devices := deviceHandler.GetDevices()
	if len(devices) > 0 {
		fmt.Println("\nAnnounced devices:")
		for _, device := range devices {
			fmt.Printf("  - %s (%s)\n", device.PreferredDosName, device.DeviceData)
		}
	} else {
		fmt.Println("\nNo devices announced")
	}

	fmt.Println("\n1. Enable/Disable device redirection")
	fmt.Println("2. Announce printer")
	fmt.Println("3. Announce drive")
	fmt.Println("4. List devices")
	fmt.Println("5. Back")

	fmt.Print("\nSelect option (1-5): ")
	var choice string
	fmt.Scanln(&choice)

	switch choice {
	case "1":
		enabled := !deviceHandler.IsEnabled()
		deviceHandler.SetEnabled(enabled)
		fmt.Printf("Device redirection %s\n", d.getEnabledStatus(enabled))
	case "2":
		fmt.Print("Enter printer name: ")
		var printerName string
		fmt.Scanln(&printerName)
		if err := deviceHandler.AnnouncePrinter(printerName); err != nil {
			fmt.Printf("Error: %v\n", err)
		}
	case "3":
		fmt.Print("Enter drive letter (e.g., D): ")
		var driveLetter string
		fmt.Scanln(&driveLetter)
		fmt.Print("Enter local path: ")
		var path string
		fmt.Scanln(&path)
		if err := deviceHandler.AnnounceDrive(driveLetter, path); err != nil {
			fmt.Printf("Error: %v\n", err)
		}
	case "4":
		devices := deviceHandler.GetDevices()
		if len(devices) > 0 {
			fmt.Println("\n--- Device List ---")
			for _, device := range devices {
				fmt.Printf("ID: %d, Type: %d, Name: %s, Data: %s\n",
					device.DeviceID, device.DeviceType, device.PreferredDosName, device.DeviceData)
			}
		} else {
			fmt.Println("No devices available")
		}
	case "5":
		return
	default:
		fmt.Println("Invalid option.")
	}
}

// showChannelStatus displays the status of all virtual channels
func (d *VirtualChannelDialog) showChannelStatus() {
	fmt.Println("\n--- Virtual Channel Status ---")

	openChannels := d.manager.GetOpenChannels()
	if len(openChannels) > 0 {
		fmt.Println("Open channels:")
		for _, channel := range openChannels {
			fmt.Printf("  - %s\n", channel)
		}
	} else {
		fmt.Println("No channels are currently open")
	}

	// Show individual channel status
	fmt.Printf("\nClipboard channel: %s\n", d.getChannelStatus("clipboard"))
	fmt.Printf("Audio channel: %s\n", d.getChannelStatus("audio"))
	fmt.Printf("Device channel: %s\n", d.getChannelStatus("device"))

	fmt.Println("\nPress Enter to continue...")
	var input string
	fmt.Scanln(&input)
}

// getEnabledStatus returns a string representation of enabled/disabled status
func (d *VirtualChannelDialog) getEnabledStatus(enabled bool) string {
	if enabled {
		return "Enabled"
	}
	return "Disabled"
}

// getChannelStatus returns the status of a specific channel
func (d *VirtualChannelDialog) getChannelStatus(channelName string) string {
	if d.manager.IsChannelOpen(channelName) {
		return "Open"
	}
	return "Closed"
}
