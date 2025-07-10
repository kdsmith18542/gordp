# GoRDP Dependencies

This document outlines all dependencies required to build, run, and develop GoRDP with its various components.

## Table of Contents

1. [Core Dependencies](#core-dependencies)
2. [GUI Dependencies](#gui-dependencies)
3. [Qt GUI Dependencies](#qt-gui-dependencies)
4. [Development Dependencies](#development-dependencies)
5. [Optional Dependencies](#optional-dependencies)
6. [Platform-Specific Dependencies](#platform-specific-dependencies)
7. [Docker Dependencies](#docker-dependencies)
8. [Installation Commands](#installation-commands)

## Core Dependencies

### Required
- **Go 1.18+** - Primary programming language
- **Git** - Version control system
- **Make** - Build system (optional but recommended)

### Runtime
- **OpenSSL** - Cryptographic functions and SSL/TLS support
- **CA Certificates** - SSL certificate validation
- **Timezone Data** - Time zone support

## GUI Dependencies

### Go GUI
- **Go 1.18+** - Same as core dependencies
- **Standard Go libraries** - Included with Go installation

### Qt GUI
- **Qt6** - GUI framework
  - Qt6 Core
  - Qt6 Widgets
  - Qt6 Network
  - Qt6 WebSockets
  - Qt6 Charts
  - Qt6 Declarative
- **CMake 3.16+** - Build system
- **C++ Compiler** - GCC or Clang with C++17 support
- **pkg-config** - Package configuration tool

### Additional Qt Dependencies
- **OpenSSL Development Libraries** - Enhanced security features
- **Audio Libraries** - Audio redirection support
  - ALSA (Linux)
  - PulseAudio (Linux)
  - Core Audio (macOS)
  - DirectSound (Windows)

## Development Dependencies

### Code Quality Tools
- **golangci-lint** - Go linter
- **goimports** - Import formatting
- **gocyclo** - Cyclomatic complexity checker
- **gosec** - Security scanner

### Testing Tools
- **Go test framework** - Included with Go
- **Test coverage tools** - Included with Go
- **Benchmark tools** - Included with Go

### Documentation Tools
- **Go doc** - Documentation generator (included with Go)
- **Markdown tools** - Documentation formatting

## Optional Dependencies

### Performance Monitoring
- **InfluxDB** - Time series database for metrics
- **Grafana** - Metrics visualization
- **Redis** - Session management and caching

### Database Support
- **PostgreSQL** - Management console database
- **SQLite** - Local database (included with Go)

### Audio/Video
- **FFmpeg** - Media processing
- **GStreamer** - Multimedia framework

### Security
- **Smart Card Libraries** - Smart card authentication
- **Biometric Libraries** - Biometric authentication
- **Certificate Management** - Advanced certificate handling

## Platform-Specific Dependencies

### Linux (Ubuntu/Debian)

#### Core Dependencies
```bash
sudo apt update
sudo apt install -y \
    golang-go \
    git \
    make \
    openssl \
    ca-certificates \
    tzdata
```

#### Qt GUI Dependencies
```bash
sudo apt install -y \
    qt6-base-dev \
    qt6-tools-dev \
    qt6-websockets-dev \
    qt6-charts-dev \
    qt6-declarative-dev \
    cmake \
    build-essential \
    pkg-config \
    libssl-dev \
    libasound2-dev \
    libpulse-dev
```

#### Development Dependencies
```bash
sudo apt install -y \
    golangci-lint \
    gosec
```

### Linux (CentOS/RHEL/Fedora)

#### Core Dependencies
```bash
sudo yum install -y \
    golang \
    git \
    make \
    openssl \
    ca-certificates \
    tzdata
```

#### Qt GUI Dependencies
```bash
sudo yum install -y \
    qt6-qtbase-devel \
    qt6-qttools-devel \
    qt6-qtwebsockets-devel \
    qt6-qtcharts-devel \
    qt6-qtdeclarative-devel \
    cmake \
    gcc-c++ \
    pkgconfig \
    openssl-devel \
    alsa-lib-devel \
    pulseaudio-libs-devel
```

### macOS

#### Core Dependencies
```bash
# Using Homebrew
brew install go git make openssl
```

#### Qt GUI Dependencies
```bash
brew install qt6 cmake pkg-config
```

#### Development Dependencies
```bash
brew install golangci-lint gosec
```

### Windows

#### Core Dependencies
- **Go** - Download from https://golang.org/dl/
- **Git** - Download from https://git-scm.com/
- **Make** - Install via Chocolatey: `choco install make`

#### Qt GUI Dependencies
- **Qt6** - Download from https://www.qt.io/download
- **CMake** - Download from https://cmake.org/download/
- **Visual Studio Build Tools** - For C++ compilation
- **pkg-config** - Install via Chocolatey: `choco install pkg-config`

#### Development Dependencies
```powershell
# Using Chocolatey
choco install golangci-lint gosec
```

## Docker Dependencies

### Build Dependencies
- **Docker** - Container runtime
- **Docker Compose** - Multi-container orchestration

### Runtime Dependencies (included in Docker image)
- **Alpine Linux** - Base image
- **Qt6 Runtime Libraries** - GUI support
- **OpenSSL** - Security
- **Audio Libraries** - Audio redirection

## Installation Commands

### Quick Installation

#### Linux (Ubuntu/Debian)
```bash
# Install all dependencies
curl -fsSL https://raw.githubusercontent.com/kdsmith18542/gordp/main/install.sh | bash

# Or install manually
sudo apt update
sudo apt install -y golang-go git make qt6-base-dev qt6-tools-dev qt6-websockets-dev qt6-charts-dev cmake build-essential pkg-config libssl-dev
```

#### macOS
```bash
# Install all dependencies
brew install go git make qt6 cmake pkg-config golangci-lint gosec
```

#### Windows
```powershell
# Install using Chocolatey
choco install go git make cmake pkg-config golangci-lint gosec

# Install Qt6 manually from https://www.qt.io/download
```

### Development Setup

#### Full Development Environment
```bash
# Clone repository
git clone https://github.com/kdsmith18542/gordp.git
cd gordp

# Setup development environment
make dev-setup-full

# Or setup manually
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install golang.org/x/tools/cmd/goimports@latest
go install github.com/fzipp/gocyclo/cmd/gocyclo@latest
go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
```

#### Qt Development Setup
```bash
# Install Qt6 dependencies
make deps-qt

# Check Qt6 installation
make check-qt

# Build Qt GUI
make build-qt
```

## Version Requirements

### Minimum Versions
- **Go**: 1.18.0
- **Git**: 2.0.0
- **CMake**: 3.16.0
- **Qt6**: 6.0.0
- **GCC**: 7.0.0
- **Clang**: 6.0.0

### Recommended Versions
- **Go**: 1.21.0+
- **Git**: 2.30.0+
- **CMake**: 3.20.0+
- **Qt6**: 6.5.0+
- **GCC**: 9.0.0+
- **Clang**: 12.0.0+

## Dependency Verification

### Check Core Dependencies
```bash
# Check Go installation
go version

# Check Git installation
git --version

# Check Make installation
make --version
```

### Check Qt Dependencies
```bash
# Check Qt6 installation
qmake6 --version

# Check CMake installation
cmake --version

# Check C++ compiler
g++ --version
clang++ --version
```

### Check Development Dependencies
```bash
# Check linter
golangci-lint --version

# Check security scanner
gosec --version

# Check import formatter
goimports --version
```

## Troubleshooting

### Common Issues

#### Qt6 Not Found
```bash
# Ubuntu/Debian
sudo apt install qt6-base-dev qt6-tools-dev

# CentOS/RHEL
sudo yum install qt6-qtbase-devel qt6-qttools-devel

# macOS
brew install qt6

# Windows
# Download from https://www.qt.io/download
```

#### CMake Not Found
```bash
# Ubuntu/Debian
sudo apt install cmake

# CentOS/RHEL
sudo yum install cmake

# macOS
brew install cmake

# Windows
# Download from https://cmake.org/download/
```

#### OpenSSL Development Libraries
```bash
# Ubuntu/Debian
sudo apt install libssl-dev

# CentOS/RHEL
sudo yum install openssl-devel

# macOS
brew install openssl

# Windows
# Usually included with Qt6
```

#### Audio Libraries
```bash
# Ubuntu/Debian
sudo apt install libasound2-dev libpulse-dev

# CentOS/RHEL
sudo yum install alsa-lib-devel pulseaudio-libs-devel

# macOS
# Core Audio is included with macOS

# Windows
# DirectSound is included with Windows
```

### Platform-Specific Issues

#### Linux
- **X11 Display**: Set `DISPLAY` environment variable for GUI
- **Audio**: Ensure PulseAudio or ALSA is running
- **Permissions**: May need to run with `sudo` for system-wide installation

#### macOS
- **Xcode**: Install Xcode Command Line Tools for C++ compilation
- **Homebrew**: Use Homebrew for package management
- **Permissions**: Grant necessary permissions for audio and accessibility

#### Windows
- **Visual Studio**: Install Visual Studio Build Tools for C++ compilation
- **PATH**: Ensure all tools are in system PATH
- **Antivirus**: May need to exclude build directories

## Docker Dependencies

### Build Requirements
- **Docker**: 20.10.0+
- **Docker Compose**: 2.0.0+

### Runtime Requirements
- **Linux**: Kernel 4.0+
- **Windows**: Windows 10/11 with WSL2
- **macOS**: macOS 10.15+ with Docker Desktop

### Docker Installation
```bash
# Linux
curl -fsSL https://get.docker.com | sh

# macOS
# Download Docker Desktop from https://www.docker.com/products/docker-desktop

# Windows
# Download Docker Desktop from https://www.docker.com/products/docker-desktop
```

## Performance Considerations

### Build Performance
- **Parallel Builds**: Use `make -j$(nproc)` for parallel compilation
- **Caching**: Enable Go module cache and Docker layer caching
- **SSD Storage**: Use SSD for faster build times

### Runtime Performance
- **Memory**: 2GB RAM minimum, 4GB recommended
- **CPU**: Multi-core CPU recommended for GUI applications
- **Network**: Stable network connection for RDP connections
- **Storage**: 100MB available disk space minimum

## Security Considerations

### Dependencies
- **Regular Updates**: Keep all dependencies updated
- **Security Scanning**: Use `gosec` for security analysis
- **Vulnerability Monitoring**: Monitor for known vulnerabilities

### Build Security
- **Source Verification**: Verify source code integrity
- **Dependency Scanning**: Scan for vulnerable dependencies
- **Secure Build**: Use secure build environments

## Support

### Getting Help
- **Documentation**: Check the main README.md
- **Issues**: Report issues on GitHub
- **Discussions**: Use GitHub Discussions for questions
- **Wiki**: Check the project wiki for additional information

### Contributing
- **Development Setup**: Follow the development setup guide
- **Code Standards**: Follow Go and Qt coding standards
- **Testing**: Write tests for new features
- **Documentation**: Update documentation for changes 