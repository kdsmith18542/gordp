# GoRDP Qt GUI - User Manual

## Table of Contents
1. [Introduction](#introduction)
2. [Installation](#installation)
3. [Getting Started](#getting-started)
4. [User Interface](#user-interface)
5. [Connections](#connections)
6. [Advanced Features](#advanced-features)
7. [Settings](#settings)
8. [Troubleshooting](#troubleshooting)
9. [FAQ](#faq)

## Introduction

GoRDP Qt GUI is a modern, high-performance Remote Desktop Protocol (RDP) client built with Qt C++ and Go. It provides a user-friendly interface for connecting to Windows servers and workstations with advanced features like multi-monitor support, virtual channels, and performance monitoring.

### Key Features

- **Modern Qt C++ Interface**: Professional, responsive GUI with dark theme
- **High Performance**: Hardware-accelerated rendering and optimized networking
- **Multi-Monitor Support**: Connect to multiple displays simultaneously
- **Virtual Channels**: Clipboard, audio, and device redirection
- **Plugin System**: Extensible architecture with custom plugins
- **Performance Monitoring**: Real-time connection statistics
- **Cross-Platform**: Windows, macOS, and Linux support
- **Connection History**: Track and manage previous connections
- **Favorites Management**: Save and organize server connections

### System Requirements

- **Operating System**: Windows 10/11, macOS 11+, or Linux (Ubuntu 20.04+)
- **Memory**: 2 GB RAM minimum, 4 GB recommended
- **Storage**: 100 MB available disk space
- **Network**: Internet connection for RDP connections
- **Display**: 1024x768 minimum resolution

## Installation

### Windows

1. Download the GoRDP Qt GUI installer from the official website
2. Run the installer as Administrator
3. Follow the installation wizard
4. Launch GoRDP Qt GUI from the Start menu

### macOS

1. Download the GoRDP Qt GUI .dmg file
2. Open the .dmg file and drag GoRDP to Applications
3. Launch GoRDP from Applications folder
4. Grant necessary permissions when prompted

### Linux

#### Ubuntu/Debian
```bash
# Install dependencies
sudo apt update
sudo apt install qt6-base-dev qt6-websockets-dev

# Download and install
wget https://github.com/gordp/gordp/releases/latest/download/gordp-gui-linux-amd64.deb
sudo dpkg -i gordp-gui-linux-amd64.deb
```

#### CentOS/RHEL/Fedora
```bash
# Install dependencies
sudo dnf install qt6-qtbase-devel qt6-qtwebsockets-devel

# Download and install
wget https://github.com/gordp/gordp/releases/latest/download/gordp-gui-linux-amd64.rpm
sudo rpm -i gordp-gui-linux-amd64.rpm
```

## Getting Started

### First Launch

1. **Launch GoRDP Qt GUI**
   - Windows: Start menu → GoRDP Qt GUI
   - macOS: Applications → GoRDP Qt GUI
   - Linux: Applications menu → GoRDP Qt GUI

2. **Main Window Overview**
   - Menu Bar: File, Edit, View, Tools, Help menus
   - Toolbar: Quick access to common functions
   - Display Area: Shows remote desktop when connected
   - Status Bar: Connection status and information

3. **Create Your First Connection**
   - Click "Connect" in the toolbar or File → New Connection
   - Enter server address (e.g., `192.168.1.100` or `server.domain.com`)
   - Enter username and password
   - Click "Connect"

### Connection Dialog

The connection dialog provides comprehensive options for RDP connections:

#### Basic Settings
- **Server**: IP address or hostname of the remote computer
- **Port**: RDP port (default: 3389)
- **Username**: Your username on the remote computer
- **Password**: Your password (saved securely)

#### Display Settings
- **Resolution**: Choose from predefined resolutions or custom
- **Color Depth**: 8-bit, 16-bit, 24-bit, or 32-bit color
- **Full Screen**: Enable for immersive experience
- **Multi-Monitor**: Use multiple displays

#### Performance Settings
- **Connection Speed**: Auto, Modem, Broadband, or LAN
- **Desktop Composition**: Enable for Aero effects
- **Menu Animations**: Enable for smooth transitions
- **Themes**: Enable for visual themes
- **Font Smoothing**: Enable for better text rendering

#### Advanced Settings
- **Security**: Choose authentication method
- **Gateway**: Configure RD Gateway settings
- **Certificates**: Manage SSL certificates
- **Virtual Channels**: Configure clipboard, audio, devices

## User Interface

### Menu Bar

#### File Menu
- **New Connection**: Open connection dialog
- **Quick Connect**: Connect to last server
- **Open Connection File**: Load saved connection
- **Save Connection As**: Save current connection
- **Recent Connections**: List of recent servers
- **Exit**: Close application

#### Edit Menu
- **Copy**: Copy selected content
- **Paste**: Paste from clipboard
- **Select All**: Select all content
- **Find**: Search in remote desktop
- **Preferences**: Open settings dialog

#### View Menu
- **Full Screen**: Toggle full screen mode
- **Zoom In/Out**: Adjust display scaling
- **Fit to Window**: Auto-scale to window size
- **Actual Size**: Show at 100% scale
- **Refresh**: Refresh display
- **Toolbar**: Show/hide toolbar
- **Status Bar**: Show/hide status bar

#### Tools Menu
- **Performance Monitor**: Open performance dialog
- **Connection History**: View connection history
- **Favorites**: Manage favorite servers
- **Plugin Manager**: Manage plugins
- **Virtual Channels**: Configure virtual channels
- **Multi-Monitor**: Configure multiple displays

#### Help Menu
- **User Manual**: Open this manual
- **About**: Application information
- **Check for Updates**: Check for new versions
- **Report Bug**: Submit bug report

### Toolbar

The toolbar provides quick access to common functions:

- **Connect**: Open connection dialog
- **Disconnect**: End current connection
- **Full Screen**: Toggle full screen mode
- **Zoom In**: Increase display scale
- **Zoom Out**: Decrease display scale
- **Fit to Window**: Auto-scale display
- **Performance**: Open performance monitor
- **Settings**: Open settings dialog

### Status Bar

The status bar shows:
- Connection status (Connected, Disconnected, Connecting)
- Server information
- Performance indicators
- Current resolution and color depth

## Connections

### Creating Connections

1. **Basic Connection**
   - Click "Connect" or File → New Connection
   - Enter server address and credentials
   - Click "Connect"

2. **Advanced Connection**
   - Use the "Advanced" tab in connection dialog
   - Configure display, performance, and security settings
   - Save connection for future use

3. **Quick Connect**
   - Use File → Quick Connect for last server
   - Or use Ctrl+Q keyboard shortcut

### Managing Connections

#### Connection History
- View all previous connections
- Filter by success/failure
- View connection statistics
- Reconnect to previous servers

#### Favorites
- Save frequently used servers
- Organize favorites in folders
- Quick access from favorites menu
- Import/export favorites

#### Connection Profiles
- Save connection settings as profiles
- Apply profiles to new connections
- Share profiles with other users
- Backup and restore profiles

### Connection Types

#### Standard RDP
- Direct connection to Windows servers
- No additional infrastructure required
- Suitable for local networks

#### RD Gateway
- Connect through Remote Desktop Gateway
- Secure connections over internet
- Enterprise deployment support

#### NLA (Network Level Authentication)
- Enhanced security with pre-authentication
- Recommended for production environments
- Requires server-side configuration

## Advanced Features

### Multi-Monitor Support

1. **Enable Multi-Monitor**
   - Check "Use all monitors" in connection dialog
   - Or use Tools → Multi-Monitor

2. **Monitor Configuration**
   - Select which monitors to use
   - Arrange monitor layout
   - Set primary monitor

3. **Spanning Options**
   - Span across all monitors
   - Use individual monitor windows
   - Custom monitor arrangement

### Virtual Channels

#### Clipboard
- **Enable**: Check "Enable clipboard sharing"
- **Text**: Copy/paste text between local and remote
- **Images**: Copy/paste images
- **Files**: Copy/paste files (if enabled)

#### Audio
- **Playback**: Hear remote computer audio
- **Recording**: Use local microphone on remote
- **Quality**: Adjust audio quality settings

#### Device Redirection
- **Drives**: Access local drives on remote computer
- **Printers**: Use local printers on remote
- **Ports**: Redirect serial/parallel ports

### Plugin System

#### Installing Plugins
1. Download plugin file (.so, .dll, .dylib)
2. Open Tools → Plugin Manager
3. Click "Install Plugin"
4. Select plugin file
5. Enable plugin if needed

#### Managing Plugins
- Enable/disable plugins
- Configure plugin settings
- View plugin information
- Remove plugins

#### Available Plugins
- **Performance Monitor**: Enhanced performance tracking
- **Logger**: Connection logging
- **Security**: Enhanced security features
- **Custom**: User-developed plugins

### Performance Monitoring

#### Real-Time Statistics
- **Bandwidth**: Network usage
- **Latency**: Connection delay
- **FPS**: Frame rate
- **CPU**: Processor usage
- **Memory**: Memory usage

#### Performance Graphs
- Historical performance data
- Trend analysis
- Performance alerts
- Export performance data

#### Optimization Tips
- Adjust color depth for better performance
- Use hardware acceleration
- Optimize network settings
- Monitor resource usage

## Settings

### General Settings

#### Application
- **Language**: Choose interface language
- **Theme**: Light or dark theme
- **Startup**: Auto-start with system
- **Updates**: Automatic update settings

#### Display
- **Default Resolution**: Set default display resolution
- **Color Depth**: Default color depth
- **Scaling**: Display scaling behavior
- **Hardware Acceleration**: Enable/disable

#### Network
- **Connection Timeout**: Set timeout values
- **Retry Attempts**: Number of retry attempts
- **Proxy Settings**: Configure proxy if needed
- **Bandwidth Limit**: Limit bandwidth usage

### Security Settings

#### Authentication
- **Default Security Level**: Choose security method
- **Certificate Validation**: SSL certificate settings
- **Credential Storage**: How to store passwords
- **Session Security**: Session-specific settings

#### Privacy
- **Connection History**: How long to keep history
- **Favorites**: Password storage in favorites
- **Logging**: What to log
- **Analytics**: Usage analytics settings

### Advanced Settings

#### Performance
- **Hardware Acceleration**: GPU acceleration settings
- **Memory Management**: Memory usage limits
- **Caching**: Bitmap and data caching
- **Compression**: Data compression settings

#### Compatibility
- **Legacy Support**: Support for older RDP versions
- **Protocol Options**: RDP protocol settings
- **Virtual Channels**: Default virtual channel settings
- **Multi-Monitor**: Default multi-monitor settings

## Troubleshooting

### Common Issues

#### Connection Problems

**Cannot Connect to Server**
- Verify server address and port
- Check network connectivity
- Ensure RDP is enabled on server
- Verify firewall settings

**Authentication Failed**
- Check username and password
- Verify account permissions
- Check domain settings
- Try different authentication method

**Slow Performance**
- Reduce color depth
- Disable visual effects
- Check network bandwidth
- Enable hardware acceleration

#### Display Issues

**Screen Not Displaying**
- Check display settings
- Try different resolution
- Restart application
- Update graphics drivers

**Poor Image Quality**
- Increase color depth
- Disable compression
- Check network quality
- Adjust scaling settings

**Multi-Monitor Not Working**
- Verify monitor detection
- Check monitor arrangement
- Update display drivers
- Restart application

#### Audio Problems

**No Audio**
- Check audio settings
- Verify virtual channel enabled
- Check system audio
- Update audio drivers

**Poor Audio Quality**
- Adjust audio quality settings
- Check network bandwidth
- Disable other audio applications
- Update audio drivers

### Error Messages

#### Common Error Codes

**0x00000000**: Success
**0x00000001**: General error
**0x00000002**: Invalid parameter
**0x00000003**: Access denied
**0x00000004**: Not enough memory
**0x00000005**: Network error
**0x00000006**: Timeout
**0x00000007**: Connection lost

#### Troubleshooting Steps

1. **Check Error Details**
   - Note exact error message
   - Check error code
   - Review application logs

2. **Basic Troubleshooting**
   - Restart application
   - Check system resources
   - Verify network connectivity
   - Update application

3. **Advanced Troubleshooting**
   - Check system logs
   - Test with different server
   - Verify server configuration
   - Contact support

### Getting Help

#### Self-Help Resources
- **User Manual**: This document
- **Online Documentation**: Official website
- **Video Tutorials**: YouTube channel
- **Community Forum**: User community

#### Support Options
- **Email Support**: support@gordp.com
- **Live Chat**: Available on website
- **Phone Support**: Business hours
- **Remote Support**: Screen sharing assistance

## FAQ

### General Questions

**Q: Is GoRDP Qt GUI free?**
A: Yes, GoRDP Qt GUI is open-source and free to use.

**Q: What operating systems are supported?**
A: Windows 10/11, macOS 11+, and Linux (Ubuntu 20.04+).

**Q: Can I connect to any Windows computer?**
A: Yes, as long as Remote Desktop is enabled on the target computer.

**Q: Is it secure to use?**
A: Yes, it supports all standard RDP security features including NLA and SSL.

### Technical Questions

**Q: What is the difference between RDP and VNC?**
A: RDP is Microsoft's protocol optimized for Windows, while VNC is platform-independent but less efficient.

**Q: Can I use multiple monitors?**
A: Yes, GoRDP Qt GUI supports multi-monitor connections.

**Q: Does it support audio?**
A: Yes, audio playback and recording are supported through virtual channels.

**Q: Can I transfer files?**
A: Yes, through drive redirection virtual channels.

### Performance Questions

**Q: How much bandwidth does it use?**
A: Typically 100KB/s to 1MB/s depending on activity and settings.

**Q: What affects performance most?**
A: Network latency, color depth, and visual effects settings.

**Q: Can I optimize for slow connections?**
A: Yes, reduce color depth and disable visual effects.

**Q: Does hardware acceleration help?**
A: Yes, especially for high-resolution displays.

### Troubleshooting Questions

**Q: Why is my connection slow?**
A: Check network quality, reduce color depth, and disable visual effects.

**Q: Why can't I connect?**
A: Verify server address, check firewall settings, and ensure RDP is enabled.

**Q: Why is the display blurry?**
A: Increase color depth and check scaling settings.

**Q: Why is audio not working?**
A: Check virtual channel settings and system audio configuration.

---

**Version**: 1.0.0  
**Last Updated**: December 2024  
**Support**: support@gordp.com  
**Website**: https://gordp.com 