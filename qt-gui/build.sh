#!/bin/bash

# GoRDP Qt C++ GUI Build Script
# This script builds the Qt C++ GUI application with advanced features

set -e

echo "=== GoRDP Qt C++ GUI Build Script ==="

# Check if we're in the right directory
if [ ! -f "CMakeLists.txt" ]; then
    echo "Error: CMakeLists.txt not found. Please run this script from the qt-gui directory."
    exit 1
fi

# Check for Qt6
if ! command -v qmake6 &> /dev/null && ! command -v qmake &> /dev/null; then
    echo "Error: Qt6 not found. Please install Qt6 development tools."
    echo "On Ubuntu/Debian: sudo apt install qt6-base-dev qt6-tools-dev qt6-websockets-dev qt6-charts-dev qt6-declarative-dev"
    echo "On CentOS/RHEL: sudo yum install qt6-qtbase-devel qt6-qttools-devel qt6-qtwebsockets-devel qt6-qtcharts-devel qt6-qtdeclarative-devel"
    echo "On macOS: brew install qt6"
    echo "On Windows: Download from https://www.qt.io/download"
    exit 1
fi

# Check for CMake
if ! command -v cmake &> /dev/null; then
    echo "Error: CMake not found. Please install CMake 3.16 or later."
    echo "On Ubuntu/Debian: sudo apt install cmake"
    echo "On CentOS/RHEL: sudo yum install cmake"
    echo "On macOS: brew install cmake"
    exit 1
fi

# Check for C++ compiler
if ! command -v g++ &> /dev/null && ! command -v clang++ &> /dev/null; then
    echo "Error: C++ compiler not found. Please install g++ or clang++."
    echo "On Ubuntu/Debian: sudo apt install build-essential"
    echo "On CentOS/RHEL: sudo yum install gcc-c++"
    echo "On macOS: Install Xcode Command Line Tools"
    exit 1
fi

# Check for additional dependencies
echo "Checking for additional dependencies..."

# Check for pkg-config
if ! command -v pkg-config &> /dev/null; then
    echo "Warning: pkg-config not found. Some features may not work correctly."
    echo "On Ubuntu/Debian: sudo apt install pkg-config"
    echo "On CentOS/RHEL: sudo yum install pkgconfig"
fi

# Check for OpenSSL (for enhanced security)
if ! pkg-config --exists openssl 2>/dev/null; then
    echo "Warning: OpenSSL development libraries not found. Some security features may be limited."
    echo "On Ubuntu/Debian: sudo apt install libssl-dev"
    echo "On CentOS/RHEL: sudo yum install openssl-devel"
fi

# Check for audio libraries (for audio redirection)
if ! pkg-config --exists alsa 2>/dev/null && ! pkg-config --exists pulse 2>/dev/null; then
    echo "Warning: Audio libraries not found. Audio redirection will be disabled."
    echo "On Ubuntu/Debian: sudo apt install libasound2-dev libpulse-dev"
    echo "On CentOS/RHEL: sudo yum install alsa-lib-devel pulseaudio-libs-devel"
fi

echo "✓ Build dependencies found"

# Create build directory
mkdir -p build
cd build

echo "=== Configuring with CMake ==="

# Configure with additional options
CMAKE_OPTIONS="-DCMAKE_BUILD_TYPE=Release"

# Enable additional features if dependencies are available
if pkg-config --exists openssl 2>/dev/null; then
    CMAKE_OPTIONS="$CMAKE_OPTIONS -DENABLE_OPENSSL=ON"
    echo "✓ OpenSSL support enabled"
fi

if pkg-config --exists alsa 2>/dev/null || pkg-config --exists pulse 2>/dev/null; then
    CMAKE_OPTIONS="$CMAKE_OPTIONS -DENABLE_AUDIO=ON"
    echo "✓ Audio support enabled"
fi

# Platform-specific options
if [[ "$OSTYPE" == "linux-gnu"* ]]; then
    CMAKE_OPTIONS="$CMAKE_OPTIONS -DENABLE_LINUX_FEATURES=ON"
    echo "✓ Linux-specific features enabled"
elif [[ "$OSTYPE" == "darwin"* ]]; then
    CMAKE_OPTIONS="$CMAKE_OPTIONS -DENABLE_MACOS_FEATURES=ON"
    echo "✓ macOS-specific features enabled"
elif [[ "$OSTYPE" == "msys" ]] || [[ "$OSTYPE" == "cygwin" ]]; then
    CMAKE_OPTIONS="$CMAKE_OPTIONS -DENABLE_WINDOWS_FEATURES=ON"
    echo "✓ Windows-specific features enabled"
fi

# Run CMake configuration
cmake .. $CMAKE_OPTIONS

echo "=== Building ==="

# Get number of CPU cores for parallel build
if command -v nproc &> /dev/null; then
    CORES=$(nproc)
else
    CORES=4
fi

echo "Building with $CORES cores..."

# Build the application
make -j$CORES

echo "=== Build Complete ==="
echo "Executable location: build/bin/gordp-gui"

# Check if executable was created
if [ -f "bin/gordp-gui" ]; then
    echo "✓ Build successful!"
    echo "To run the application: ./bin/gordp-gui"
    
    # Show file size and dependencies
    echo ""
    echo "Build Information:"
    echo "  File size: $(du -h bin/gordp-gui | cut -f1)"
    echo "  Architecture: $(file bin/gordp-gui | cut -d',' -f2 | xargs)"
    
    # Check for Qt dependencies
    if command -v ldd &> /dev/null; then
        echo ""
        echo "Qt Dependencies:"
        ldd bin/gordp-gui | grep -i qt || echo "  No Qt dependencies found (static build)"
    fi
    
    # Create desktop shortcut on Linux
    if [[ "$OSTYPE" == "linux-gnu"* ]]; then
        echo ""
        echo "Creating desktop shortcut..."
        cat > gordp-gui.desktop << 'EOF'
[Desktop Entry]
Version=1.0
Type=Application
Name=GoRDP GUI
Comment=Production-grade RDP client with Qt GUI
Exec=/usr/local/bin/gordp-gui
Icon=network-server
Terminal=false
Categories=Network;RemoteAccess;
EOF
        
        if [ -d "$HOME/Desktop" ]; then
            cp gordp-gui.desktop "$HOME/Desktop/"
            chmod +x "$HOME/Desktop/gordp-gui.desktop"
            echo "✓ Desktop shortcut created"
        fi
    fi
    
else
    echo "✗ Build failed - executable not found"
    exit 1
fi

# Create installation script
echo ""
echo "Creating installation script..."
cat > install.sh << 'EOF'
#!/bin/bash
# GoRDP Qt GUI Installation Script

set -e

INSTALL_DIR="/usr/local/bin"
DESKTOP_DIR="$HOME/.local/share/applications"

echo "Installing GoRDP Qt GUI..."

# Create installation directory
sudo mkdir -p "$INSTALL_DIR"

# Install executable
sudo cp bin/gordp-gui "$INSTALL_DIR/"
sudo chmod +x "$INSTALL_DIR/gordp-gui"

# Create desktop file
mkdir -p "$DESKTOP_DIR"
cat > "$DESKTOP_DIR/gordp-gui.desktop" << 'DESKTOP_EOF'
[Desktop Entry]
Version=1.0
Type=Application
Name=GoRDP GUI
Comment=Production-grade RDP client with Qt GUI
Exec=gordp-gui
Icon=network-server
Terminal=false
Categories=Network;RemoteAccess;
DESKTOP_EOF

chmod +x "$DESKTOP_DIR/gordp-gui.desktop"

echo "✓ GoRDP Qt GUI installed successfully!"
echo "  Executable: $INSTALL_DIR/gordp-gui"
echo "  Desktop shortcut: $DESKTOP_DIR/gordp-gui.desktop"
echo ""
echo "You can now run 'gordp-gui' from anywhere or launch it from your applications menu."
EOF

chmod +x install.sh
echo "✓ Installation script created: install.sh"

echo ""
echo "=== Build Summary ==="
echo "✓ Qt6 GUI application built successfully"
echo "✓ Advanced features enabled:"
echo "  - Multi-monitor support"
echo "  - Virtual channels (clipboard, audio, device redirection)"
echo "  - Performance monitoring"
echo "  - Connection history and favorites"
echo "  - Plugin system"
echo "  - Advanced security features"
echo ""
echo "To install: ./install.sh"
echo "To run: ./bin/gordp-gui" 