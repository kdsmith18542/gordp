package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/kdsmith18542/gordp"
	"github.com/kdsmith18542/gordp/proto/bitmap"
	"github.com/kdsmith18542/gordp/proto/t128"
)

type InteractiveProcessor struct {
	client *gordp.Client
}

func (p *InteractiveProcessor) ProcessBitmap(option *bitmap.Option, bitmap *bitmap.BitMap) {
	log.Printf("Display update: %dx%d at (%d,%d)",
		option.Width, option.Height, option.Left, option.Top)

	// Save bitmap to file for debugging (optional)
	// filename := fmt.Sprintf("screenshot_%d.png", time.Now().Unix())
	// if err := bitmap.SaveToFile(filename); err != nil {
	//     log.Printf("Failed to save bitmap: %v", err)
	// }
}

func main() {
	if len(os.Args) < 4 {
		fmt.Println("Usage: interactive_client <host:port> <username> <password>")
		fmt.Println("Example: interactive_client 192.168.1.100:3389 administrator password")
		os.Exit(1)
	}

	addr := os.Args[1]
	username := os.Args[2]
	password := os.Args[3]

	// Create client
	client := gordp.NewClient(&gordp.Option{
		Addr:           addr,
		UserName:       username,
		Password:       password,
		ConnectTimeout: 10 * time.Second,
	})

	processor := &InteractiveProcessor{client: client}

	// Connect to RDP server
	fmt.Printf("Connecting to %s...\n", addr)
	if err := client.Connect(); err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer client.Close()

	fmt.Println("Connected successfully!")
	fmt.Println("Starting interactive session...")

	// Start input handling in a goroutine
	go handleUserInput(client)

	// Start the RDP session
	if err := client.Run(processor); err != nil {
		log.Fatalf("RDP session failed: %v", err)
	}
}

func handleUserInput(client *gordp.Client) {
	// Wait a moment for connection to stabilize
	time.Sleep(2 * time.Second)

	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("\n=== GoRDP Interactive Client ===")
	fmt.Println("Available commands:")
	fmt.Println("  help                           - Show this help")
	fmt.Println("  quit                           - Quit the client")
	fmt.Println("  type <text>                    - Type text")
	fmt.Println("  key <keyname>                  - Press a key")
	fmt.Println("  ctrl+<key>                     - Ctrl+key combination")
	fmt.Println("  alt+<key>                      - Alt+key combination")
	fmt.Println("  shift+<key>                    - Shift+key combination")
	fmt.Println("  mouse <x> <y>                  - Move mouse to position")
	fmt.Println("  click <button> <x> <y>         - Click mouse button")
	fmt.Println("  double <button> <x> <y>        - Double-click mouse button")
	fmt.Println("  drag <button> <x1> <y1> <x2> <y2> - Drag mouse")
	fmt.Println("  wheel <delta> <x> <y>          - Scroll mouse wheel")
	fmt.Println("  smooth <button> <x1> <y1> <x2> <y2> <steps> - Smooth drag")
	fmt.Println("  multi <button> <x> <y> <count> - Multiple clicks")
	fmt.Println("  scroll <direction> <amount> <x> <y> - Scroll in direction")
	fmt.Println("  function <number>              - Press function key")
	fmt.Println("  arrow <direction>              - Press arrow key")
	fmt.Println("  nav <keyname>                  - Press navigation key")
	fmt.Println("  media <keyname>                - Press media key")
	fmt.Println("  browser <keyname>              - Press browser key")
	fmt.Println("  unicode <text>                 - Send Unicode text")
	fmt.Println("  repeat <key> <count>           - Repeat key press")
	fmt.Println("  delay <key> <ms>               - Press key with delay")
	fmt.Println("  numpad <key> <numlock>         - Press numpad key")
	fmt.Println("  demo                           - Run demo sequence")
	fmt.Println()

	for {
		fmt.Print("gordp> ")
		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}

		parts := strings.Fields(input)
		if len(parts) == 0 {
			continue
		}

		command := strings.ToLower(parts[0])

		switch command {
		case "help":
			showHelp()
		case "quit", "exit":
			fmt.Println("Goodbye!")
			os.Exit(0)
		case "type":
			if len(parts) < 2 {
				fmt.Println("Usage: type <text>")
				continue
			}
			text := strings.Join(parts[1:], " ")
			if err := client.SendString(text); err != nil {
				fmt.Printf("Error typing text: %v\n", err)
			} else {
				fmt.Printf("Typed: %s\n", text)
			}
		case "key":
			if len(parts) < 2 {
				fmt.Println("Usage: key <keyname>")
				continue
			}
			if err := client.SendSpecialKey(parts[1], t128.ModifierKey{}); err != nil {
				fmt.Printf("Error pressing key: %v\n", err)
			} else {
				fmt.Printf("Pressed key: %s\n", parts[1])
			}
		case "ctrl":
			if len(parts) < 2 {
				fmt.Println("Usage: ctrl+<key>")
				continue
			}
			key := parts[1]
			if keyCode, ok := t128.KeyMap[rune(key[0])]; ok {
				if err := client.SendCtrlKey(keyCode); err != nil {
					fmt.Printf("Error pressing Ctrl+%s: %v\n", key, err)
				} else {
					fmt.Printf("Pressed Ctrl+%s\n", key)
				}
			} else {
				fmt.Printf("Unknown key: %s\n", key)
			}
		case "alt":
			if len(parts) < 2 {
				fmt.Println("Usage: alt+<key>")
				continue
			}
			key := parts[1]
			if keyCode, ok := t128.KeyMap[rune(key[0])]; ok {
				if err := client.SendAltKey(keyCode); err != nil {
					fmt.Printf("Error pressing Alt+%s: %v\n", key, err)
				} else {
					fmt.Printf("Pressed Alt+%s\n", key)
				}
			} else {
				fmt.Printf("Unknown key: %s\n", key)
			}
		case "shift":
			if len(parts) < 2 {
				fmt.Println("Usage: shift+<key>")
				continue
			}
			key := parts[1]
			if keyCode, ok := t128.KeyMap[rune(key[0])]; ok {
				if err := client.SendShiftKey(keyCode); err != nil {
					fmt.Printf("Error pressing Shift+%s: %v\n", key, err)
				} else {
					fmt.Printf("Pressed Shift+%s\n", key)
				}
			} else {
				fmt.Printf("Unknown key: %s\n", key)
			}
		case "mouse":
			if len(parts) < 3 {
				fmt.Println("Usage: mouse <x> <y>")
				continue
			}
			x, err := strconv.ParseUint(parts[1], 10, 16)
			if err != nil {
				fmt.Printf("Invalid x coordinate: %s\n", parts[1])
				continue
			}
			y, err := strconv.ParseUint(parts[2], 10, 16)
			if err != nil {
				fmt.Printf("Invalid y coordinate: %s\n", parts[2])
				continue
			}
			if err := client.SendMouseMoveEvent(uint16(x), uint16(y)); err != nil {
				fmt.Printf("Error moving mouse: %v\n", err)
			} else {
				fmt.Printf("Moved mouse to (%d, %d)\n", x, y)
			}
		case "click":
			if len(parts) < 4 {
				fmt.Println("Usage: click <button> <x> <y>")
				continue
			}
			button := parseMouseButton(parts[1])
			x, err := strconv.ParseUint(parts[2], 10, 16)
			if err != nil {
				fmt.Printf("Invalid x coordinate: %s\n", parts[2])
				continue
			}
			y, err := strconv.ParseUint(parts[3], 10, 16)
			if err != nil {
				fmt.Printf("Invalid y coordinate: %s\n", parts[3])
				continue
			}
			if err := client.SendMouseClickEvent(button, uint16(x), uint16(y)); err != nil {
				fmt.Printf("Error clicking mouse: %v\n", err)
			} else {
				fmt.Printf("Clicked %s button at (%d, %d)\n", parts[1], x, y)
			}
		case "double":
			if len(parts) < 4 {
				fmt.Println("Usage: double <button> <x> <y>")
				continue
			}
			button := parseMouseButton(parts[1])
			x, err := strconv.ParseUint(parts[2], 10, 16)
			if err != nil {
				fmt.Printf("Invalid x coordinate: %s\n", parts[2])
				continue
			}
			y, err := strconv.ParseUint(parts[3], 10, 16)
			if err != nil {
				fmt.Printf("Invalid y coordinate: %s\n", parts[3])
				continue
			}
			if err := client.SendMouseDoubleClickEvent(button, uint16(x), uint16(y)); err != nil {
				fmt.Printf("Error double-clicking mouse: %v\n", err)
			} else {
				fmt.Printf("Double-clicked %s button at (%d, %d)\n", parts[1], x, y)
			}
		case "drag":
			if len(parts) < 6 {
				fmt.Println("Usage: drag <button> <x1> <y1> <x2> <y2>")
				continue
			}
			button := parseMouseButton(parts[1])
			x1, err := strconv.ParseUint(parts[2], 10, 16)
			if err != nil {
				fmt.Printf("Invalid x1 coordinate: %s\n", parts[2])
				continue
			}
			y1, err := strconv.ParseUint(parts[3], 10, 16)
			if err != nil {
				fmt.Printf("Invalid y1 coordinate: %s\n", parts[3])
				continue
			}
			x2, err := strconv.ParseUint(parts[4], 10, 16)
			if err != nil {
				fmt.Printf("Invalid x2 coordinate: %s\n", parts[4])
				continue
			}
			y2, err := strconv.ParseUint(parts[5], 10, 16)
			if err != nil {
				fmt.Printf("Invalid y2 coordinate: %s\n", parts[5])
				continue
			}
			if err := client.SendMouseDragEvent(button, uint16(x1), uint16(y1), uint16(x2), uint16(y2)); err != nil {
				fmt.Printf("Error dragging mouse: %v\n", err)
			} else {
				fmt.Printf("Dragged %s button from (%d, %d) to (%d, %d)\n", parts[1], x1, y1, x2, y2)
			}
		case "wheel":
			if len(parts) < 4 {
				fmt.Println("Usage: wheel <delta> <x> <y>")
				continue
			}
			delta, err := strconv.ParseInt(parts[1], 10, 16)
			if err != nil {
				fmt.Printf("Invalid delta: %s\n", parts[1])
				continue
			}
			x, err := strconv.ParseUint(parts[2], 10, 16)
			if err != nil {
				fmt.Printf("Invalid x coordinate: %s\n", parts[2])
				continue
			}
			y, err := strconv.ParseUint(parts[3], 10, 16)
			if err != nil {
				fmt.Printf("Invalid y coordinate: %s\n", parts[3])
				continue
			}
			if err := client.SendMouseWheelEvent(int16(delta), uint16(x), uint16(y)); err != nil {
				fmt.Printf("Error scrolling mouse wheel: %v\n", err)
			} else {
				fmt.Printf("Scrolled mouse wheel by %d at (%d, %d)\n", delta, x, y)
			}
		case "smooth":
			if len(parts) < 7 {
				fmt.Println("Usage: smooth <button> <x1> <y1> <x2> <y2> <steps>")
				continue
			}
			button := parseMouseButton(parts[1])
			x1, err := strconv.ParseUint(parts[2], 10, 16)
			if err != nil {
				fmt.Printf("Invalid x1 coordinate: %s\n", parts[2])
				continue
			}
			y1, err := strconv.ParseUint(parts[3], 10, 16)
			if err != nil {
				fmt.Printf("Invalid y1 coordinate: %s\n", parts[3])
				continue
			}
			x2, err := strconv.ParseUint(parts[4], 10, 16)
			if err != nil {
				fmt.Printf("Invalid x2 coordinate: %s\n", parts[4])
				continue
			}
			y2, err := strconv.ParseUint(parts[5], 10, 16)
			if err != nil {
				fmt.Printf("Invalid y2 coordinate: %s\n", parts[5])
				continue
			}
			steps, err := strconv.Atoi(parts[6])
			if err != nil {
				fmt.Printf("Invalid steps: %s\n", parts[6])
				continue
			}
			if err := client.SendMouseSmoothDragEvent(button, uint16(x1), uint16(y1), uint16(x2), uint16(y2), steps); err != nil {
				fmt.Printf("Error smooth dragging mouse: %v\n", err)
			} else {
				fmt.Printf("Smooth dragged %s button from (%d, %d) to (%d, %d) in %d steps\n", parts[1], x1, y1, x2, y2, steps)
			}
		case "multi":
			if len(parts) < 5 {
				fmt.Println("Usage: multi <button> <x> <y> <count>")
				continue
			}
			button := parseMouseButton(parts[1])
			x, err := strconv.ParseUint(parts[2], 10, 16)
			if err != nil {
				fmt.Printf("Invalid x coordinate: %s\n", parts[2])
				continue
			}
			y, err := strconv.ParseUint(parts[3], 10, 16)
			if err != nil {
				fmt.Printf("Invalid y coordinate: %s\n", parts[3])
				continue
			}
			count, err := strconv.Atoi(parts[4])
			if err != nil {
				fmt.Printf("Invalid count: %s\n", parts[4])
				continue
			}
			if err := client.SendMouseMultiClickEvent(button, uint16(x), uint16(y), count); err != nil {
				fmt.Printf("Error multi-clicking mouse: %v\n", err)
			} else {
				fmt.Printf("Multi-clicked %s button %d times at (%d, %d)\n", parts[1], count, x, y)
			}
		case "scroll":
			if len(parts) < 5 {
				fmt.Println("Usage: scroll <direction> <amount> <x> <y>")
				continue
			}
			direction := parseScrollDirection(parts[1])
			amount, err := strconv.ParseInt(parts[2], 10, 16)
			if err != nil {
				fmt.Printf("Invalid amount: %s\n", parts[2])
				continue
			}
			x, err := strconv.ParseUint(parts[3], 10, 16)
			if err != nil {
				fmt.Printf("Invalid x coordinate: %s\n", parts[3])
				continue
			}
			y, err := strconv.ParseUint(parts[4], 10, 16)
			if err != nil {
				fmt.Printf("Invalid y coordinate: %s\n", parts[4])
				continue
			}
			if err := client.SendMouseScrollEvent(direction, int16(amount), uint16(x), uint16(y)); err != nil {
				fmt.Printf("Error scrolling: %v\n", err)
			} else {
				fmt.Printf("Scrolled %s by %d at (%d, %d)\n", parts[1], amount, x, y)
			}
		case "function":
			if len(parts) < 2 {
				fmt.Println("Usage: function <number>")
				continue
			}
			number, err := strconv.Atoi(parts[1])
			if err != nil {
				fmt.Printf("Invalid function number: %s\n", parts[1])
				continue
			}
			if err := client.SendFunctionKey(number, t128.ModifierKey{}); err != nil {
				fmt.Printf("Error pressing function key: %v\n", err)
			} else {
				fmt.Printf("Pressed F%d\n", number)
			}
		case "arrow":
			if len(parts) < 2 {
				fmt.Println("Usage: arrow <direction>")
				continue
			}
			if err := client.SendArrowKey(parts[1], t128.ModifierKey{}); err != nil {
				fmt.Printf("Error pressing arrow key: %v\n", err)
			} else {
				fmt.Printf("Pressed %s arrow key\n", parts[1])
			}
		case "nav":
			if len(parts) < 2 {
				fmt.Println("Usage: nav <keyname>")
				continue
			}
			if err := client.SendNavigationKey(parts[1], t128.ModifierKey{}); err != nil {
				fmt.Printf("Error pressing navigation key: %v\n", err)
			} else {
				fmt.Printf("Pressed %s navigation key\n", parts[1])
			}
		case "media":
			if len(parts) < 2 {
				fmt.Println("Usage: media <keyname>")
				continue
			}
			if err := client.SendMediaKey(parts[1]); err != nil {
				fmt.Printf("Error pressing media key: %v\n", err)
			} else {
				fmt.Printf("Pressed %s media key\n", parts[1])
			}
		case "browser":
			if len(parts) < 2 {
				fmt.Println("Usage: browser <keyname>")
				continue
			}
			if err := client.SendBrowserKey(parts[1]); err != nil {
				fmt.Printf("Error pressing browser key: %v\n", err)
			} else {
				fmt.Printf("Pressed %s browser key\n", parts[1])
			}
		case "unicode":
			if len(parts) < 2 {
				fmt.Println("Usage: unicode <text>")
				continue
			}
			text := strings.Join(parts[1:], " ")
			if err := client.SendUnicodeString(text); err != nil {
				fmt.Printf("Error sending Unicode text: %v\n", err)
			} else {
				fmt.Printf("Sent Unicode text: %s\n", text)
			}
		case "repeat":
			if len(parts) < 3 {
				fmt.Println("Usage: repeat <key> <count>")
				continue
			}
			key := parts[1]
			count, err := strconv.Atoi(parts[2])
			if err != nil {
				fmt.Printf("Invalid count: %s\n", parts[2])
				continue
			}
			if keyCode, ok := t128.KeyMap[rune(key[0])]; ok {
				if err := client.SendKeyRepeat(keyCode, count, t128.ModifierKey{}); err != nil {
					fmt.Printf("Error repeating key: %v\n", err)
				} else {
					fmt.Printf("Repeated key %s %d times\n", key, count)
				}
			} else {
				fmt.Printf("Unknown key: %s\n", key)
			}
		case "delay":
			if len(parts) < 3 {
				fmt.Println("Usage: delay <key> <ms>")
				continue
			}
			key := parts[1]
			ms, err := strconv.Atoi(parts[2])
			if err != nil {
				fmt.Printf("Invalid delay: %s\n", parts[2])
				continue
			}
			if keyCode, ok := t128.KeyMap[rune(key[0])]; ok {
				if err := client.SendKeyWithDelay(keyCode, ms, t128.ModifierKey{}); err != nil {
					fmt.Printf("Error pressing key with delay: %v\n", err)
				} else {
					fmt.Printf("Pressed key %s with %dms delay\n", key, ms)
				}
			} else {
				fmt.Printf("Unknown key: %s\n", key)
			}
		case "numpad":
			if len(parts) < 3 {
				fmt.Println("Usage: numpad <key> <numlock>")
				continue
			}
			key := parts[1]
			numlock := parts[2] == "true" || parts[2] == "1"
			if keyCode, ok := t128.KeyMap[rune(key[0])]; ok {
				if err := client.SendNumpadKey(keyCode, numlock, t128.ModifierKey{}); err != nil {
					fmt.Printf("Error pressing numpad key: %v\n", err)
				} else {
					fmt.Printf("Pressed numpad key %s (numlock: %v)\n", key, numlock)
				}
			} else {
				fmt.Printf("Unknown key: %s\n", key)
			}
		case "demo":
			runDemo(client)
		default:
			fmt.Printf("Unknown command: %s\n", command)
			fmt.Println("Type 'help' for available commands")
		}
	}
}

func showHelp() {
	fmt.Println("=== GoRDP Interactive Client Help ===")
	fmt.Println("Available commands:")
	fmt.Println("  help                           - Show this help")
	fmt.Println("  quit                           - Quit the client")
	fmt.Println("  type <text>                    - Type text")
	fmt.Println("  key <keyname>                  - Press a key")
	fmt.Println("  ctrl+<key>                     - Ctrl+key combination")
	fmt.Println("  alt+<key>                      - Alt+key combination")
	fmt.Println("  shift+<key>                    - Shift+key combination")
	fmt.Println("  mouse <x> <y>                  - Move mouse to position")
	fmt.Println("  click <button> <x> <y>         - Click mouse button")
	fmt.Println("  double <button> <x> <y>        - Double-click mouse button")
	fmt.Println("  drag <button> <x1> <y1> <x2> <y2> - Drag mouse")
	fmt.Println("  wheel <delta> <x> <y>          - Scroll mouse wheel")
	fmt.Println("  smooth <button> <x1> <y1> <x2> <y2> <steps> - Smooth drag")
	fmt.Println("  multi <button> <x> <y> <count> - Multiple clicks")
	fmt.Println("  scroll <direction> <amount> <x> <y> - Scroll in direction")
	fmt.Println("  function <number>              - Press function key")
	fmt.Println("  arrow <direction>              - Press arrow key")
	fmt.Println("  nav <keyname>                  - Press navigation key")
	fmt.Println("  media <keyname>                - Press media key")
	fmt.Println("  browser <keyname>              - Press browser key")
	fmt.Println("  unicode <text>                 - Send Unicode text")
	fmt.Println("  repeat <key> <count>           - Repeat key press")
	fmt.Println("  delay <key> <ms>               - Press key with delay")
	fmt.Println("  numpad <key> <numlock>         - Press numpad key")
	fmt.Println("  demo                           - Run demo sequence")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  type Hello, World!")
	fmt.Println("  key F1")
	fmt.Println("  ctrl+c")
	fmt.Println("  mouse 500 300")
	fmt.Println("  click left 500 300")
	fmt.Println("  double left 500 300")
	fmt.Println("  drag left 100 100 200 200")
	fmt.Println("  wheel 120 500 300")
	fmt.Println("  function 1")
	fmt.Println("  arrow UP")
	fmt.Println("  nav HOME")
	fmt.Println("  media PLAY")
	fmt.Println("  browser BACK")
	fmt.Println("  unicode Hello, 世界!")
	fmt.Println("  repeat a 5")
	fmt.Println("  delay a 1000")
	fmt.Println("  numpad 5 true")
	fmt.Println()
}

func parseMouseButton(buttonStr string) t128.MouseButton {
	switch strings.ToLower(buttonStr) {
	case "left":
		return t128.MouseButtonLeft
	case "right":
		return t128.MouseButtonRight
	case "middle":
		return t128.MouseButtonMiddle
	case "x1":
		return t128.MouseButtonX1
	case "x2":
		return t128.MouseButtonX2
	default:
		return t128.MouseButtonLeft
	}
}

func parseScrollDirection(directionStr string) t128.ScrollDirection {
	switch strings.ToLower(directionStr) {
	case "up":
		return t128.ScrollUp
	case "down":
		return t128.ScrollDown
	case "left":
		return t128.ScrollLeft
	case "right":
		return t128.ScrollRight
	default:
		return t128.ScrollUp
	}
}

func runDemo(client *gordp.Client) {
	fmt.Println("Running demo sequence...")

	// Wait a moment
	time.Sleep(1 * time.Second)

	// Type some text
	fmt.Println("Typing text...")
	if err := client.SendString("Hello from GoRDP!"); err != nil {
		fmt.Printf("Demo typing failed: %v\n", err)
	}

	time.Sleep(500 * time.Millisecond)

	// Press Enter
	if err := client.SendKeyPress(t128.VK_RETURN, t128.ModifierKey{}); err != nil {
		fmt.Printf("Demo Enter key failed: %v\n", err)
	}

	time.Sleep(500 * time.Millisecond)

	// Move mouse and click
	fmt.Println("Moving mouse and clicking...")
	if err := client.SendMouseMoveEvent(500, 300); err != nil {
		fmt.Printf("Demo mouse move failed: %v\n", err)
	}

	time.Sleep(200 * time.Millisecond)

	if err := client.SendMouseClickEvent(t128.MouseButtonLeft, 500, 300); err != nil {
		fmt.Printf("Demo mouse click failed: %v\n", err)
	}

	time.Sleep(500 * time.Millisecond)

	// Press some special keys
	fmt.Println("Pressing special keys...")
	if err := client.SendSpecialKey("F1", t128.ModifierKey{}); err != nil {
		fmt.Printf("Demo F1 key failed: %v\n", err)
	}

	time.Sleep(200 * time.Millisecond)

	if err := client.SendSpecialKey("HOME", t128.ModifierKey{}); err != nil {
		fmt.Printf("Demo HOME key failed: %v\n", err)
	}

	time.Sleep(200 * time.Millisecond)

	if err := client.SendSpecialKey("END", t128.ModifierKey{}); err != nil {
		fmt.Printf("Demo END key failed: %v\n", err)
	}

	time.Sleep(500 * time.Millisecond)

	// Scroll wheel
	fmt.Println("Scrolling...")
	if err := client.SendMouseWheelEvent(120, 500, 300); err != nil {
		fmt.Printf("Demo wheel scroll failed: %v\n", err)
	}

	time.Sleep(200 * time.Millisecond)

	if err := client.SendMouseWheelEvent(-120, 500, 300); err != nil {
		fmt.Printf("Demo wheel scroll failed: %v\n", err)
	}

	time.Sleep(500 * time.Millisecond)

	// Key combinations
	fmt.Println("Testing key combinations...")
	if err := client.SendCtrlKey('a'); err != nil {
		fmt.Printf("Demo Ctrl+A failed: %v\n", err)
	}

	time.Sleep(200 * time.Millisecond)

	if err := client.SendCtrlKey('c'); err != nil {
		fmt.Printf("Demo Ctrl+C failed: %v\n", err)
	}

	time.Sleep(200 * time.Millisecond)

	if err := client.SendCtrlKey('v'); err != nil {
		fmt.Printf("Demo Ctrl+V failed: %v\n", err)
	}

	fmt.Println("Demo sequence completed!")
}
