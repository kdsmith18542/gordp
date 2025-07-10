package plugins

import (
	"fmt"
	"strconv"
)

// PluginDialog manages the plugin configuration dialog
type PluginDialog struct {
	manager *PluginManager
}

// NewPluginDialog creates a new plugin dialog
func NewPluginDialog(manager *PluginManager) *PluginDialog {
	return &PluginDialog{
		manager: manager,
	}
}

// Show displays the plugin management dialog
func (d *PluginDialog) Show() {
	fmt.Println("\n=== Plugin Management ===")

	for {
		fmt.Println("\n1. Plugin List")
		fmt.Println("2. Enable/Disable Plugin")
		fmt.Println("3. Plugin Configuration")
		fmt.Println("4. Start/Stop Plugin")
		fmt.Println("5. Plugin Statistics")
		fmt.Println("6. Start All Plugins")
		fmt.Println("7. Stop All Plugins")
		fmt.Println("8. Back to Main Menu")

		fmt.Print("\nSelect option (1-8): ")
		var choice string
		fmt.Scanln(&choice)

		switch choice {
		case "1":
			d.showPluginList()
		case "2":
			d.enableDisablePlugin()
		case "3":
			d.configurePlugin()
		case "4":
			d.startStopPlugin()
		case "5":
			d.showPluginStats()
		case "6":
			d.startAllPlugins()
		case "7":
			d.stopAllPlugins()
		case "8":
			return
		default:
			fmt.Println("Invalid option. Please try again.")
		}
	}
}

// showPluginList displays the list of plugins
func (d *PluginDialog) showPluginList() {
	plugins := d.manager.GetPlugins()

	if len(plugins) == 0 {
		fmt.Println("\n--- Plugin List ---")
		fmt.Println("No plugins registered.")
		return
	}

	fmt.Println("\n--- Plugin List ---")
	for i, plugin := range plugins {
		status := "Enabled"
		if !plugin.Enabled {
			status = "Disabled"
		}

		// Get plugin status from core manager
		coreStatus, err := d.manager.GetPluginStatus(plugin.Name)
		if err != nil {
			coreStatus = "Unknown"
		}

		fmt.Printf("%d. %s v%s\n", i+1, plugin.Name, plugin.Version)
		fmt.Printf("   Description: %s\n", plugin.Description)
		fmt.Printf("   Status: %s (%s)\n", status, coreStatus)
		fmt.Printf("   Config items: %d\n", len(plugin.Config))
		fmt.Println()
	}
}

// enableDisablePlugin allows enabling/disabling plugins
func (d *PluginDialog) enableDisablePlugin() {
	plugins := d.manager.GetPlugins()

	if len(plugins) == 0 {
		fmt.Println("No plugins available.")
		return
	}

	fmt.Println("\n--- Enable/Disable Plugin ---")
	for i, plugin := range plugins {
		status := "Enabled"
		if !plugin.Enabled {
			status = "Disabled"
		}
		fmt.Printf("%d. %s (%s)\n", i+1, plugin.Name, status)
	}

	fmt.Print("\nSelect plugin (1-" + strconv.Itoa(len(plugins)) + "): ")
	var choice string
	fmt.Scanln(&choice)

	index, err := strconv.Atoi(choice)
	if err != nil || index < 1 || index > len(plugins) {
		fmt.Println("Invalid selection.")
		return
	}

	plugin := plugins[index-1]
	enabled := !plugin.Enabled

	if err := d.manager.EnablePlugin(plugin.Name, enabled); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	status := "enabled"
	if !enabled {
		status = "disabled"
	}
	fmt.Printf("Plugin %s %s\n", plugin.Name, status)
}

// configurePlugin allows configuring plugin settings
func (d *PluginDialog) configurePlugin() {
	plugins := d.manager.GetPlugins()

	if len(plugins) == 0 {
		fmt.Println("No plugins available.")
		return
	}

	fmt.Println("\n--- Plugin Configuration ---")
	for i, plugin := range plugins {
		fmt.Printf("%d. %s\n", i+1, plugin.Name)
	}

	fmt.Print("\nSelect plugin (1-" + strconv.Itoa(len(plugins)) + "): ")
	var choice string
	fmt.Scanln(&choice)

	index, err := strconv.Atoi(choice)
	if err != nil || index < 1 || index > len(plugins) {
		fmt.Println("Invalid selection.")
		return
	}

	plugin := plugins[index-1]
	d.configureSpecificPlugin(plugin)
}

// configureSpecificPlugin configures a specific plugin
func (d *PluginDialog) configureSpecificPlugin(plugin *PluginInfo) {
	fmt.Printf("\n--- Configure %s ---\n", plugin.Name)

	for {
		fmt.Println("\n1. View current configuration")
		fmt.Println("2. Set configuration value")
		fmt.Println("3. Initialize plugin")
		fmt.Println("4. Back")

		fmt.Print("\nSelect option (1-4): ")
		var choice string
		fmt.Scanln(&choice)

		switch choice {
		case "1":
			d.showPluginConfig(plugin)
		case "2":
			d.setPluginConfig(plugin)
		case "3":
			d.initializePlugin(plugin)
		case "4":
			return
		default:
			fmt.Println("Invalid option.")
		}
	}
}

// showPluginConfig displays the current configuration of a plugin
func (d *PluginDialog) showPluginConfig(plugin *PluginInfo) {
	fmt.Printf("\n--- %s Configuration ---\n", plugin.Name)

	if len(plugin.Config) == 0 {
		fmt.Println("No configuration set.")
		return
	}

	for key, value := range plugin.Config {
		fmt.Printf("%s: %v\n", key, value)
	}
}

// setPluginConfig sets a configuration value for a plugin
func (d *PluginDialog) setPluginConfig(plugin *PluginInfo) {
	fmt.Print("Enter configuration key: ")
	var key string
	fmt.Scanln(&key)

	fmt.Print("Enter configuration value: ")
	var value string
	fmt.Scanln(&value)

	if err := d.manager.SetPluginConfig(plugin.Name, key, value); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Configuration updated: %s = %s\n", key, value)
}

// initializePlugin initializes a plugin with its configuration
func (d *PluginDialog) initializePlugin(plugin *PluginInfo) {
	if err := d.manager.InitializePlugin(plugin.Name, plugin.Config); err != nil {
		fmt.Printf("Error initializing plugin: %v\n", err)
		return
	}

	fmt.Printf("Plugin %s initialized successfully\n", plugin.Name)
}

// startStopPlugin allows starting/stopping plugins
func (d *PluginDialog) startStopPlugin() {
	plugins := d.manager.GetPlugins()

	if len(plugins) == 0 {
		fmt.Println("No plugins available.")
		return
	}

	fmt.Println("\n--- Start/Stop Plugin ---")
	for i, plugin := range plugins {
		status, err := d.manager.GetPluginStatus(plugin.Name)
		if err != nil {
			status = "Unknown"
		}
		fmt.Printf("%d. %s (%s)\n", i+1, plugin.Name, status)
	}

	fmt.Print("\nSelect plugin (1-" + strconv.Itoa(len(plugins)) + "): ")
	var choice string
	fmt.Scanln(&choice)

	index, err := strconv.Atoi(choice)
	if err != nil || index < 1 || index > len(plugins) {
		fmt.Println("Invalid selection.")
		return
	}

	plugin := plugins[index-1]
	_, err = d.manager.GetPluginStatus(plugin.Name)
	if err != nil {
		fmt.Printf("Error getting plugin status: %v\n", err)
		return
	}

	fmt.Println("\n1. Start plugin")
	fmt.Println("2. Stop plugin")
	fmt.Print("\nSelect option (1-2): ")

	var action string
	fmt.Scanln(&action)

	switch action {
	case "1":
		if err := d.manager.StartPlugin(plugin.Name); err != nil {
			fmt.Printf("Error starting plugin: %v\n", err)
		} else {
			fmt.Printf("Plugin %s started\n", plugin.Name)
		}
	case "2":
		if err := d.manager.StopPlugin(plugin.Name); err != nil {
			fmt.Printf("Error stopping plugin: %v\n", err)
		} else {
			fmt.Printf("Plugin %s stopped\n", plugin.Name)
		}
	default:
		fmt.Println("Invalid option.")
	}
}

// showPluginStats displays plugin statistics
func (d *PluginDialog) showPluginStats() {
	stats := d.manager.GetPluginStats()

	fmt.Println("\n--- Plugin Statistics ---")
	fmt.Printf("Total plugins: %v\n", stats["total_plugins"])
	fmt.Printf("Running plugins: %v\n", stats["running_plugins"])
	fmt.Printf("Error plugins: %v\n", stats["error_plugins"])
	fmt.Printf("System enabled: %v\n", stats["system_enabled"])

	plugins := d.manager.GetPlugins()
	if len(plugins) > 0 {
		fmt.Println("\nPlugin details:")
		for _, plugin := range plugins {
			status, err := d.manager.GetPluginStatus(plugin.Name)
			if err != nil {
				status = "Unknown"
			}
			fmt.Printf("  - %s: %s\n", plugin.Name, status)
		}
	}

	fmt.Println("\nPress Enter to continue...")
	var input string
	fmt.Scanln(&input)
}

// startAllPlugins starts all plugins
func (d *PluginDialog) startAllPlugins() {
	fmt.Println("\n--- Start All Plugins ---")

	if err := d.manager.StartAllPlugins(); err != nil {
		fmt.Printf("Error starting all plugins: %v\n", err)
		return
	}

	fmt.Println("All plugins started successfully")
}

// stopAllPlugins stops all plugins
func (d *PluginDialog) stopAllPlugins() {
	fmt.Println("\n--- Stop All Plugins ---")

	if err := d.manager.StopAllPlugins(); err != nil {
		fmt.Printf("Error stopping all plugins: %v\n", err)
		return
	}

	fmt.Println("All plugins stopped successfully")
}
