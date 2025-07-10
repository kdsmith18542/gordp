#!/bin/bash

# GoRDP GUI Build Script
# This script builds the GoRDP GUI applications for multiple platforms

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Build directory
BUILD_DIR="build"
GUI_BUILD_DIR="${BUILD_DIR}/gui"
QT_BUILD_DIR="${BUILD_DIR}/qt-gui"

# Create build directories
mkdir -p "${GUI_BUILD_DIR}"
mkdir -p "${QT_BUILD_DIR}"

echo -e "${GREEN}Building GoRDP GUI Clients...${NC}"

# Function to build Go GUI for a specific platform
build_go_gui_platform() {
    local os=$1
    local arch=$2
    local output_name=$3
    
    echo -e "${YELLOW}Building Go GUI for ${os}/${arch}...${NC}"
    
    # Set environment variables for cross-compilation
    export GOOS=${os}
    export GOARCH=${arch}
    
    # Build the Go GUI application
    go build -ldflags "-X main.Version=$(git describe --tags --always --dirty) -X main.BuildTime=$(date -u '+%Y-%m-%d_%H:%M:%S')" -o "${GUI_BUILD_DIR}/${output_name}" ./gui
    
    echo -e "${GREEN}✓ Built Go GUI ${output_name}${NC}"
}

# Function to build Qt GUI for a specific platform
build_qt_gui_platform() {
    local os=$1
    local arch=$2
    local output_name=$3
    
    echo -e "${YELLOW}Building Qt GUI for ${os}/${arch}...${NC}"
    
    # Check if Qt6 is available
    if ! command -v qmake6 &> /dev/null && ! command -v qmake &> /dev/null; then
        echo -e "${YELLOW}Qt6 not found. Skipping Qt GUI build for ${os}/${arch}${NC}"
        return 0
    fi
    
    # Check if CMake is available
    if ! command -v cmake &> /dev/null; then
        echo -e "${YELLOW}CMake not found. Skipping Qt GUI build for ${os}/${arch}${NC}"
        return 0
    fi
    
    # Create platform-specific build directory
    local platform_build_dir="${QT_BUILD_DIR}/${os}-${arch}"
    mkdir -p "${platform_build_dir}"
    
    # Set environment variables for cross-compilation
    export GOOS=${os}
    export GOARCH=${arch}
    
    # Build the Qt GUI application
    cd qt-gui
    mkdir -p build
    cd build
    
    # Configure with CMake
    if [[ "$os" == "windows" ]]; then
        cmake .. -DCMAKE_BUILD_TYPE=Release -DCMAKE_SYSTEM_NAME=Windows
    elif [[ "$os" == "darwin" ]]; then
        cmake .. -DCMAKE_BUILD_TYPE=Release -DCMAKE_SYSTEM_NAME=Darwin
    else
        cmake .. -DCMAKE_BUILD_TYPE=Release
    fi
    
    # Build
    make -j$(nproc)
    
    # Copy executable to main build directory
    if [[ -f "bin/gordp-gui" ]]; then
        cp "bin/gordp-gui" "../../${platform_build_dir}/${output_name}"
        chmod +x "../../${platform_build_dir}/${output_name}"
        echo -e "${GREEN}✓ Built Qt GUI ${output_name}${NC}"
    else
        echo -e "${RED}✗ Qt GUI build failed for ${os}/${arch}${NC}"
    fi
    
    cd ../..
}

# Build Go GUI for different platforms
echo -e "${BLUE}Building Go GUI for Linux...${NC}"
build_go_gui_platform "linux" "amd64" "gordp-gui-linux-amd64"
build_go_gui_platform "linux" "arm64" "gordp-gui-linux-arm64"

echo -e "${BLUE}Building Go GUI for Windows...${NC}"
build_go_gui_platform "windows" "amd64" "gordp-gui-windows-amd64.exe"
build_go_gui_platform "windows" "arm64" "gordp-gui-windows-arm64.exe"

echo -e "${BLUE}Building Go GUI for macOS...${NC}"
build_go_gui_platform "darwin" "amd64" "gordp-gui-macos-amd64"
build_go_gui_platform "darwin" "arm64" "gordp-gui-macos-arm64"

# Build Qt GUI for different platforms (if Qt6 is available)
if command -v qmake6 &> /dev/null || command -v qmake &> /dev/null; then
    echo -e "${BLUE}Building Qt GUI for Linux...${NC}"
    build_qt_gui_platform "linux" "amd64" "gordp-qt-gui-linux-amd64"
    build_qt_gui_platform "linux" "arm64" "gordp-qt-gui-linux-arm64"
    
    echo -e "${BLUE}Building Qt GUI for Windows...${NC}"
    build_qt_gui_platform "windows" "amd64" "gordp-qt-gui-windows-amd64.exe"
    build_qt_gui_platform "windows" "arm64" "gordp-qt-gui-windows-arm64.exe"
    
    echo -e "${BLUE}Building Qt GUI for macOS...${NC}"
    build_qt_gui_platform "darwin" "amd64" "gordp-qt-gui-macos-amd64"
    build_qt_gui_platform "darwin" "arm64" "gordp-qt-gui-macos-arm64"
else
    echo -e "${YELLOW}Qt6 not found. Skipping Qt GUI builds.${NC}"
    echo "To build Qt GUI, install Qt6 development tools:"
    echo "  Ubuntu/Debian: sudo apt install qt6-base-dev qt6-tools-dev qt6-websockets-dev qt6-charts-dev"
    echo "  CentOS/RHEL: sudo yum install qt6-qtbase-devel qt6-qttools-devel qt6-qtwebsockets-devel qt6-qtcharts-devel"
    echo "  macOS: brew install qt6"
    echo "  Windows: Download from https://www.qt.io/download"
fi

# Make Linux builds executable
chmod +x "${GUI_BUILD_DIR}/gordp-gui-linux-amd64" 2>/dev/null || true
chmod +x "${GUI_BUILD_DIR}/gordp-gui-linux-arm64" 2>/dev/null || true
chmod +x "${GUI_BUILD_DIR}/gordp-gui-macos-amd64" 2>/dev/null || true
chmod +x "${GUI_BUILD_DIR}/gordp-gui-macos-arm64" 2>/dev/null || true

# Make Qt Linux builds executable
find "${QT_BUILD_DIR}" -name "gordp-qt-gui-linux-*" -exec chmod +x {} \; 2>/dev/null || true
find "${QT_BUILD_DIR}" -name "gordp-qt-gui-macos-*" -exec chmod +x {} \; 2>/dev/null || true

echo -e "${GREEN}✓ All builds completed successfully!${NC}"

# Display build outputs
echo -e "${YELLOW}Go GUI Build outputs:${NC}"
ls -la "${GUI_BUILD_DIR}/" 2>/dev/null || echo "No Go GUI builds found"

echo -e "${YELLOW}Qt GUI Build outputs:${NC}"
find "${QT_BUILD_DIR}" -type f -executable -exec ls -la {} \; 2>/dev/null || echo "No Qt GUI builds found"

# Create test scripts
cat > "${GUI_BUILD_DIR}/test_go_gui.sh" << 'EOF'
#!/bin/bash
echo "Testing GoRDP Go GUI Client..."
echo "Available commands:"
echo "  connect    - Connect to RDP server"
echo "  disconnect - Disconnect from server"
echo "  settings   - Open settings dialog"
echo "  status     - Show connection status"
echo "  quit       - Exit application"
echo ""
./gordp-gui-linux-amd64
EOF

chmod +x "${GUI_BUILD_DIR}/test_go_gui.sh"

# Create Qt GUI test script if Qt builds exist
if find "${QT_BUILD_DIR}" -type f -executable -name "gordp-qt-gui-linux-amd64" | grep -q .; then
    cat > "${QT_BUILD_DIR}/test_qt_gui.sh" << 'EOF'
#!/bin/bash
echo "Testing GoRDP Qt GUI Client..."
echo "Starting Qt GUI application..."
./gordp-qt-gui-linux-amd64
EOF

    chmod +x "${QT_BUILD_DIR}/test_qt_gui.sh"
    echo -e "${GREEN}✓ Qt GUI test script created: ${QT_BUILD_DIR}/test_qt_gui.sh${NC}"
fi

echo -e "${GREEN}✓ Go GUI test script created: ${GUI_BUILD_DIR}/test_go_gui.sh${NC}"
echo -e "${YELLOW}To test the Go GUI, run: cd ${GUI_BUILD_DIR} && ./test_go_gui.sh${NC}"
if find "${QT_BUILD_DIR}" -type f -executable -name "gordp-qt-gui-linux-amd64" | grep -q .; then
    echo -e "${YELLOW}To test the Qt GUI, run: cd ${QT_BUILD_DIR} && ./test_qt_gui.sh${NC}"
fi

# Create installation script
cat > "${BUILD_DIR}/install.sh" << 'EOF'
#!/bin/bash
# GoRDP GUI Installation Script
# This script installs the built GUI applications

set -e

INSTALL_DIR="/usr/local/bin"

echo "Installing GoRDP GUI applications..."

# Install Go GUI
if [ -f "gui/gordp-gui-linux-amd64" ]; then
    sudo cp gui/gordp-gui-linux-amd64 "$INSTALL_DIR/gordp-gui"
    sudo chmod +x "$INSTALL_DIR/gordp-gui"
    echo "✓ Go GUI installed as gordp-gui"
fi

# Install Qt GUI
if [ -f "qt-gui/gordp-qt-gui-linux-amd64" ]; then
    sudo cp qt-gui/gordp-qt-gui-linux-amd64 "$INSTALL_DIR/gordp-qt-gui"
    sudo chmod +x "$INSTALL_DIR/gordp-qt-gui"
    echo "✓ Qt GUI installed as gordp-qt-gui"
fi

echo "Installation completed!"
echo "Run 'gordp-gui' for Go GUI or 'gordp-qt-gui' for Qt GUI"
EOF

chmod +x "${BUILD_DIR}/install.sh"
echo -e "${GREEN}✓ Installation script created: ${BUILD_DIR}/install.sh${NC}"

# Create package script
cat > "${BUILD_DIR}/package.sh" << 'EOF'
#!/bin/bash
# GoRDP GUI Packaging Script
# This script creates distribution packages

set -e

VERSION=$(git describe --tags --always --dirty)
PACKAGE_DIR="gordp-gui-${VERSION}"

echo "Creating GoRDP GUI package ${VERSION}..."

# Create package directory
mkdir -p "$PACKAGE_DIR"

# Copy Go GUI builds
if [ -d "gui" ]; then
    cp -r gui "$PACKAGE_DIR/"
fi

# Copy Qt GUI builds
if [ -d "qt-gui" ]; then
    cp -r qt-gui "$PACKAGE_DIR/"
fi

# Copy installation script
cp install.sh "$PACKAGE_DIR/"

# Create README
cat > "$PACKAGE_DIR/README.md" << 'README_EOF'
# GoRDP GUI ${VERSION}

This package contains GoRDP GUI applications for multiple platforms.

## Contents

- **Go GUI**: Command-line interface with GUI capabilities
- **Qt GUI**: Full-featured Qt C++ GUI with advanced features

## Installation

Run the installation script:
```bash
./install.sh
```

## Usage

- Go GUI: `gordp-gui`
- Qt GUI: `gordp-qt-gui`

## Features

- Multi-monitor support
- Virtual channels (clipboard, audio, device redirection)
- Performance monitoring
- Connection history and favorites
- Plugin system
- Advanced security features

## Supported Platforms

- Linux (amd64, arm64)
- Windows (amd64, arm64)
- macOS (amd64, arm64)
README_EOF

# Create archives
tar -czf "${PACKAGE_DIR}.tar.gz" "$PACKAGE_DIR"
zip -r "${PACKAGE_DIR}.zip" "$PACKAGE_DIR"

echo "✓ Package created: ${PACKAGE_DIR}.tar.gz and ${PACKAGE_DIR}.zip"
EOF

chmod +x "${BUILD_DIR}/package.sh"
echo -e "${GREEN}✓ Package script created: ${BUILD_DIR}/package.sh${NC}"

echo -e "${GREEN}✓ All build scripts completed!${NC}"
echo -e "${BLUE}Next steps:${NC}"
echo "1. Test the applications: cd ${GUI_BUILD_DIR} && ./test_go_gui.sh"
echo "2. Install the applications: cd ${BUILD_DIR} && ./install.sh"
echo "3. Create distribution packages: cd ${BUILD_DIR} && ./package.sh" 