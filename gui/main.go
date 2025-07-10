package main

import (
	"fmt"

	"github.com/kdsmith18542/gordp/gui/mainwindow"
)

func main() {
	fmt.Println("GoRDP GUI Client - Starting...")

	// Create and show main window
	window := mainwindow.NewMainWindow()
	window.Show()

	fmt.Println("GoRDP GUI Client - Running...")

	// Keep the application running
	select {}
}
