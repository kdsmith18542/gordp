#!/bin/bash

# GoRDP Universal Installer
# This script automatically detects your platform and installs GoRDP

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
REPO="kdsmith18542/gordp"
LATEST_RELEASE_URL="https://api.github.com/repos/$REPO/releases/latest"
INSTALL_DIR="/usr/local/bin"
BINARY_NAME="gordp"
GUI_BINARY_NAME="gordp-gui"

# Functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Detect platform
detect_platform() {
    log_info "Detecting platform..."
    
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)
    
    case $ARCH in
        x86_64)
            ARCH="amd64"
            ;;
        aarch64|arm64)
            ARCH="arm64"
            ;;
        armv7l)
            ARCH="arm"
            ;;
        *)
            log_error "Unsupported architecture: $ARCH"
            exit 1
            ;;
    esac
    
    case $OS in
        linux)
            PLATFORM="linux"
            ;;
        darwin)
            PLATFORM="darwin"
            ;;
        msys*|cygwin*|mingw*)
            PLATFORM="windows"
            ;;
        *)
            log_error "Unsupported operating system: $OS"
            exit 1
            ;;
    esac
    
    log_info "Detected platform: $PLATFORM-$ARCH"
}

# Check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Check if running as root
check_root() {
    if [[ $EUID -eq 0 ]]; then
        log_warning "Running as root. This is not recommended for security reasons."
        read -p "Continue anyway? (y/N): " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            exit 1
        fi
    fi
}

# Install Qt6 dependencies
install_qt6_dependencies() {
    log_info "Installing Qt6 dependencies..."
    
    if command_exists apt; then
        log_info "Installing Qt6 dependencies with apt..."
        sudo apt update
        sudo apt install -y \
            qt6-base-dev \
            qt6-tools-dev \
            qt6-websockets-dev \
            qt6-charts-dev \
            qt6-declarative-dev \
            cmake \
            build-essential \
            pkg-config
    elif command_exists yum; then
        log_info "Installing Qt6 dependencies with yum..."
        sudo yum install -y \
            qt6-qtbase-devel \
            qt6-qttools-devel \
            qt6-qtwebsockets-devel \
            qt6-qtcharts-devel \
            qt6-qtdeclarative-devel \
            cmake \
            gcc-c++ \
            pkgconfig
    elif command_exists dnf; then
        log_info "Installing Qt6 dependencies with dnf..."
        sudo dnf install -y \
            qt6-qtbase-devel \
            qt6-qttools-devel \
            qt6-qtwebsockets-devel \
            qt6-qtcharts-devel \
            qt6-qtdeclarative-devel \
            cmake \
            gcc-c++ \
            pkgconfig
    elif command_exists brew; then
        log_info "Installing Qt6 dependencies with Homebrew..."
        brew install qt6 cmake pkg-config
    else
        log_warning "Package manager not detected. Please install Qt6 manually:"
        echo "  - Ubuntu/Debian: sudo apt install qt6-base-dev qt6-tools-dev qt6-websockets-dev qt6-charts-dev"
        echo "  - CentOS/RHEL: sudo yum install qt6-qtbase-devel qt6-qttools-devel qt6-qtwebsockets-devel qt6-qtcharts-devel"
        echo "  - macOS: brew install qt6"
        echo "  - Windows: Download from https://www.qt.io/download"
    fi
}

# Get latest version
get_latest_version() {
    log_info "Fetching latest version..."
    
    if command_exists curl; then
        VERSION=$(curl -s "$LATEST_RELEASE_URL" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
    elif command_exists wget; then
        VERSION=$(wget -qO- "$LATEST_RELEASE_URL" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
    else
        log_error "Neither curl nor wget is installed. Please install one of them."
        exit 1
    fi
    
    if [[ -z "$VERSION" ]]; then
        log_error "Failed to get latest version"
        exit 1
    fi
    
    log_info "Latest version: $VERSION"
}

# Download binary
download_binary() {
    local version=$1
    local platform=$2
    local arch=$3
    local binary_name=$4
    
    local filename="${binary_name}-${platform}-${arch}"
    if [[ "$platform" == "windows" ]]; then
        filename="${filename}.exe"
    fi
    
    local download_url="https://github.com/$REPO/releases/download/$version/$filename"
    local temp_file="/tmp/$filename"
    
    log_info "Downloading $filename..."
    
    if command_exists curl; then
        curl -L -o "$temp_file" "$download_url"
    elif command_exists wget; then
        wget -O "$temp_file" "$download_url"
    fi
    
    if [[ ! -f "$temp_file" ]]; then
        log_error "Failed to download binary"
        exit 1
    fi
    
    chmod +x "$temp_file"
    echo "$temp_file"
}

# Install binary
install_binary() {
    local temp_file=$1
    local binary_name=$2
    
    log_info "Installing to $INSTALL_DIR..."
    
    # Create install directory if it doesn't exist
    if [[ ! -d "$INSTALL_DIR" ]]; then
        sudo mkdir -p "$INSTALL_DIR"
    fi
    
    # Move binary to install directory
    sudo mv "$temp_file" "$INSTALL_DIR/$binary_name"
    
    # Verify installation
    if command_exists "$binary_name"; then
        log_success "$binary_name installed successfully!"
        log_info "You can now run: $binary_name --help"
    else
        log_error "Installation failed"
        exit 1
    fi
}

# Install using package manager
install_with_package_manager() {
    log_info "Attempting to install using package manager..."
    
    if command_exists brew; then
        log_info "Installing with Homebrew..."
        brew install "$REPO/tap/$BINARY_NAME"
        return 0
    elif command_exists apt; then
        log_info "Installing with apt..."
        # Add repository and install
        curl -fsSL https://packages.gordp.dev/gpg | sudo gpg --dearmor -o /usr/share/keyrings/gordp-archive-keyring.gpg
        echo "deb [arch=amd64 signed-by=/usr/share/keyrings/gordp-archive-keyring.gpg] https://packages.gordp.dev/ubuntu $(lsb_release -cs) main" | sudo tee /etc/apt/sources.list.d/gordp.list
        sudo apt update
        sudo apt install -y "$BINARY_NAME" "$GUI_BINARY_NAME"
        return 0
    elif command_exists yum; then
        log_info "Installing with yum..."
        # Add repository and install
        sudo yum-config-manager --add-repo https://packages.gordp.dev/rpm/gordp.repo
        sudo yum install -y "$BINARY_NAME" "$GUI_BINARY_NAME"
        return 0
    elif command_exists dnf; then
        log_info "Installing with dnf..."
        # Add repository and install
        sudo dnf config-manager --add-repo https://packages.gordp.dev/rpm/gordp.repo
        sudo dnf install -y "$BINARY_NAME" "$GUI_BINARY_NAME"
        return 0
    fi
    
    return 1
}

# Install using Go
install_with_go() {
    log_info "Installing with Go..."
    
    if ! command_exists go; then
        log_error "Go is not installed. Please install Go 1.18+ first."
        exit 1
    fi
    
    go install "github.com/$REPO@latest"
    
    # Add GOPATH/bin to PATH if not already there
    if [[ ":$PATH:" != *":$GOPATH/bin:"* ]]; then
        log_warning "Please add $GOPATH/bin to your PATH"
        echo "export PATH=\$PATH:\$GOPATH/bin" >> ~/.bashrc
        echo "export PATH=\$PATH:\$GOPATH/bin" >> ~/.zshrc
    fi
    
    log_success "GoRDP installed with Go!"
}

# Build from source
build_from_source() {
    log_info "Building GoRDP from source..."
    
    if ! command_exists go; then
        log_error "Go is not installed. Please install Go 1.18+ first."
        exit 1
    fi
    
    if ! command_exists git; then
        log_error "Git is not installed. Please install Git first."
        exit 1
    fi
    
    # Clone repository
    local temp_dir="/tmp/gordp-build"
    rm -rf "$temp_dir"
    git clone "https://github.com/$REPO.git" "$temp_dir"
    cd "$temp_dir"
    
    # Build core binary
    log_info "Building core binary..."
    go build -o "$INSTALL_DIR/$BINARY_NAME" .
    
    # Build GUI binary
    log_info "Building GUI binary..."
    go build -o "$INSTALL_DIR/$GUI_BINARY_NAME" ./gui
    
    # Build Qt GUI if Qt6 is available
    if command_exists qmake6 || command_exists qmake; then
        log_info "Building Qt GUI..."
        cd qt-gui
        if command_exists cmake; then
            mkdir -p build
            cd build
            cmake .. -DCMAKE_BUILD_TYPE=Release
            make -j$(nproc)
            sudo cp bin/gordp-gui "$INSTALL_DIR/gordp-qt-gui"
            log_success "Qt GUI built successfully!"
        else
            log_warning "CMake not found. Skipping Qt GUI build."
        fi
    else
        log_warning "Qt6 not found. Skipping Qt GUI build."
    fi
    
    # Cleanup
    rm -rf "$temp_dir"
    
    log_success "GoRDP built from source successfully!"
}

# Main installation function
main() {
    echo -e "${BLUE}"
    echo "╔══════════════════════════════════════════════════════════════╗"
    echo "║                    GoRDP Universal Installer                 ║"
    echo "║              Production-grade RDP client in Go               ║"
    echo "║              Now with Qt GUI and advanced features           ║"
    echo "╚══════════════════════════════════════════════════════════════╝"
    echo -e "${NC}"
    
    # Check if already installed
    if command_exists "$BINARY_NAME"; then
        log_warning "$BINARY_NAME is already installed"
        read -p "Reinstall? (y/N): " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            exit 0
        fi
    fi
    
    # Detect platform
    detect_platform
    
    # Check root
    check_root
    
    # Ask about GUI installation
    echo
    log_info "GoRDP now includes multiple GUI options:"
    echo "  1. Command-line interface (CLI)"
    echo "  2. Go-based GUI (recommended)"
    echo "  3. Qt C++ GUI (advanced features)"
    echo
    read -p "Install GUI components? (Y/n): " -n 1 -r
    echo
    INSTALL_GUI=true
    if [[ $REPLY =~ ^[Nn]$ ]]; then
        INSTALL_GUI=false
    fi
    
    # Install Qt6 dependencies if GUI is requested
    if [[ "$INSTALL_GUI" == true ]]; then
        install_qt6_dependencies
    fi
    
    # Try package manager first
    if install_with_package_manager; then
        log_success "Installation completed using package manager!"
        exit 0
    fi
    
    # Try Go installation
    if command_exists go; then
        log_info "Go is available. Would you like to install using Go?"
        read -p "Install with Go? (Y/n): " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Nn]$ ]]; then
            # Continue with binary download
        else
            install_with_go
            exit 0
        fi
    fi
    
    # Ask about building from source
    if command_exists go && command_exists git; then
        log_info "Go and Git are available. Would you like to build from source?"
        read -p "Build from source? (y/N): " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            build_from_source
            exit 0
        fi
    fi
    
    # Download and install binary
    get_latest_version
    temp_file=$(download_binary "$VERSION" "$PLATFORM" "$ARCH" "$BINARY_NAME")
    install_binary "$temp_file" "$BINARY_NAME"
    
    # Download and install GUI binary if requested
    if [[ "$INSTALL_GUI" == true ]]; then
        log_info "Downloading GUI binary..."
        gui_temp_file=$(download_binary "$VERSION" "$PLATFORM" "$ARCH" "$GUI_BINARY_NAME")
        install_binary "$gui_temp_file" "$GUI_BINARY_NAME"
    fi
    
    # Cleanup
    rm -f "$temp_file"
    if [[ "$INSTALL_GUI" == true ]]; then
        rm -f "$gui_temp_file"
    fi
    
    echo
    log_success "Installation completed successfully!"
    echo
    echo "Next steps:"
    echo "1. Run '$BINARY_NAME --help' to see available options"
    if [[ "$INSTALL_GUI" == true ]]; then
        echo "2. Run '$GUI_BINARY_NAME' to start the GUI"
    fi
    echo "3. Check the documentation at https://github.com/$REPO"
    echo "4. Try the examples in the repository"
    echo
    if [[ "$INSTALL_GUI" == true ]]; then
        echo "GUI Features:"
        echo "- Multi-monitor support"
        echo "- Virtual channels (clipboard, audio, device redirection)"
        echo "- Performance monitoring"
        echo "- Connection history and favorites"
        echo "- Plugin system"
        echo "- Advanced security features"
    fi
}

# Run main function
main "$@" 