# GoRDP Windows Installer
# This script installs GoRDP on Windows systems with Qt GUI support

param(
    [string]$Version = "latest",
    [string]$InstallPath = "$env:ProgramFiles\GoRDP",
    [switch]$Force,
    [switch]$Help,
    [switch]$InstallQt,
    [switch]$SkipGui
)

# Configuration
$Repo = "kdsmith18542/gordp"
$BinaryName = "gordp.exe"
$GuiBinaryName = "gordp-gui.exe"
$QtBinaryName = "gordp-qt-gui.exe"
$LatestReleaseUrl = "https://api.github.com/repos/$Repo/releases/latest"

# Functions
function Write-Info {
    param([string]$Message)
    Write-Host "[INFO] $Message" -ForegroundColor Blue
}

function Write-Success {
    param([string]$Message)
    Write-Host "[SUCCESS] $Message" -ForegroundColor Green
}

function Write-Warning {
    param([string]$Message)
    Write-Host "[WARNING] $Message" -ForegroundColor Yellow
}

function Write-Error {
    param([string]$Message)
    Write-Host "[ERROR] $Message" -ForegroundColor Red
}

function Show-Help {
    Write-Host @"
GoRDP Windows Installer

Usage: .\install.ps1 [options]

Options:
    -Version <version>     Version to install (default: latest)
    -InstallPath <path>    Installation directory (default: Program Files\GoRDP)
    -Force                 Force installation even if already installed
    -Help                  Show this help message
    -InstallQt             Install Qt GUI components (requires Qt6)
    -SkipGui               Skip GUI installation (CLI only)

Examples:
    .\install.ps1                          # Install latest version
    .\install.ps1 -Version v1.0.0          # Install specific version
    .\install.ps1 -InstallPath C:\GoRDP    # Install to custom directory
    .\install.ps1 -Force                   # Force reinstall
    .\install.ps1 -InstallQt               # Install with Qt GUI
    .\install.ps1 -SkipGui                 # Install CLI only
"@
}

# Check if running as administrator
function Test-Administrator {
    $currentUser = [Security.Principal.WindowsIdentity]::GetCurrent()
    $principal = New-Object Security.Principal.WindowsPrincipal($currentUser)
    return $principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
}

# Get latest version from GitHub
function Get-LatestVersion {
    Write-Info "Fetching latest version..."
    
    try {
        $response = Invoke-RestMethod -Uri $LatestReleaseUrl -Method Get
        $version = $response.tag_name
        Write-Info "Latest version: $version"
        return $version
    }
    catch {
        Write-Error "Failed to get latest version: $($_.Exception.Message)"
        exit 1
    }
}

# Download binary
function Download-Binary {
    param(
        [string]$Version,
        [string]$BinaryName,
        [string]$Platform = "windows",
        [string]$Arch = "amd64"
    )
    
    $filename = "$BinaryName"
    $downloadUrl = "https://github.com/$Repo/releases/download/$Version/$filename"
    $tempFile = Join-Path $env:TEMP $filename
    
    Write-Info "Downloading $filename..."
    
    try {
        Invoke-WebRequest -Uri $downloadUrl -OutFile $tempFile
        Write-Info "Download completed: $tempFile"
        return $tempFile
    }
    catch {
        Write-Error "Failed to download binary: $($_.Exception.Message)"
        exit 1
    }
}

# Install binary
function Install-Binary {
    param(
        [string]$TempFile,
        [string]$BinaryName
    )
    
    Write-Info "Installing to $InstallPath..."
    
    # Create installation directory
    if (!(Test-Path $InstallPath)) {
        New-Item -ItemType Directory -Path $InstallPath -Force | Out-Null
    }
    
    # Copy binary to installation directory
    $targetFile = Join-Path $InstallPath $BinaryName
    Copy-Item -Path $TempFile -Destination $targetFile -Force
    
    # Create start menu shortcut
    $startMenuPath = "$env:APPDATA\Microsoft\Windows\Start Menu\Programs\GoRDP"
    if (!(Test-Path $startMenuPath)) {
        New-Item -ItemType Directory -Path $startMenuPath -Force | Out-Null
    }
    
    $WshShell = New-Object -comObject WScript.Shell
    $Shortcut = $WshShell.CreateShortcut("$startMenuPath\$($BinaryName.Replace('.exe', '')).lnk")
    $Shortcut.TargetPath = $targetFile
    $Shortcut.WorkingDirectory = $InstallPath
    $Shortcut.Description = "GoRDP - Production-grade RDP client"
    $Shortcut.Save()
    
    # Add to PATH
    $currentPath = [Environment]::GetEnvironmentVariable("PATH", "Machine")
    if ($currentPath -notlike "*$InstallPath*") {
        $newPath = "$currentPath;$InstallPath"
        [Environment]::SetEnvironmentVariable("PATH", $newPath, "Machine")
        Write-Info "Added $InstallPath to system PATH"
    }
    
    # Cleanup
    Remove-Item -Path $TempFile -Force -ErrorAction SilentlyContinue
    
    Write-Success "$BinaryName installed successfully!"
}

# Install using Chocolatey
function Install-WithChocolatey {
    Write-Info "Attempting to install with Chocolatey..."
    
    if (Get-Command choco -ErrorAction SilentlyContinue) {
        Write-Info "Installing with Chocolatey..."
        choco install gordp -y
        if ($InstallQt) {
            choco install gordp-gui -y
        }
        return $true
    }
    
    return $false
}

# Install using Scoop
function Install-WithScoop {
    Write-Info "Attempting to install with Scoop..."
    
    if (Get-Command scoop -ErrorAction SilentlyContinue) {
        Write-Info "Installing with Scoop..."
        scoop install gordp
        if ($InstallQt) {
            scoop install gordp-gui
        }
        return $true
    }
    
    return $false
}

# Install using Go
function Install-WithGo {
    Write-Info "Installing with Go..."
    
    if (!(Get-Command go -ErrorAction SilentlyContinue)) {
        Write-Error "Go is not installed. Please install Go 1.18+ first."
        exit 1
    }
    
    go install "github.com/$Repo@latest"
    if (!$SkipGui) {
        go install "github.com/$Repo/gui@latest"
    }
    Write-Success "GoRDP installed with Go!"
}

# Build from source
function Build-FromSource {
    Write-Info "Building GoRDP from source..."
    
    if (!(Get-Command go -ErrorAction SilentlyContinue)) {
        Write-Error "Go is not installed. Please install Go 1.18+ first."
        exit 1
    }
    
    if (!(Get-Command git -ErrorAction SilentlyContinue)) {
        Write-Error "Git is not installed. Please install Git first."
        exit 1
    }
    
    # Clone repository
    $tempDir = Join-Path $env:TEMP "gordp-build"
    if (Test-Path $tempDir) {
        Remove-Item -Path $tempDir -Recurse -Force
    }
    
    git clone "https://github.com/$Repo.git" $tempDir
    Set-Location $tempDir
    
    # Build core binary
    Write-Info "Building core binary..."
    go build -o (Join-Path $InstallPath $BinaryName) .
    
    # Build GUI binary if not skipped
    if (!$SkipGui) {
        Write-Info "Building GUI binary..."
        go build -o (Join-Path $InstallPath $GuiBinaryName) ./gui
    }
    
    # Build Qt GUI if requested
    if ($InstallQt) {
        Write-Info "Building Qt GUI..."
        if (Test-Path "qt-gui") {
            Set-Location qt-gui
            if (Get-Command cmake -ErrorAction SilentlyContinue) {
                New-Item -ItemType Directory -Path "build" -Force | Out-Null
                Set-Location build
                cmake .. -DCMAKE_BUILD_TYPE=Release
                cmake --build . --config Release
                if (Test-Path "bin\gordp-gui.exe") {
                    Copy-Item "bin\gordp-gui.exe" (Join-Path $InstallPath $QtBinaryName)
                    Write-Success "Qt GUI built successfully!"
                }
            }
            else {
                Write-Warning "CMake not found. Skipping Qt GUI build."
            }
        }
        else {
            Write-Warning "Qt GUI source not found. Skipping Qt GUI build."
        }
    }
    
    # Cleanup
    Set-Location $PSScriptRoot
    Remove-Item -Path $tempDir -Recurse -Force
    
    Write-Success "GoRDP built from source successfully!"
}

# Main installation function
function Main {
    Write-Host @"
╔══════════════════════════════════════════════════════════════╗
║                    GoRDP Windows Installer                   ║
║              Production-grade RDP client in Go               ║
║              Now with Qt GUI and advanced features           ║
╚══════════════════════════════════════════════════════════════╝
"@ -ForegroundColor Blue
    
    # Show help if requested
    if ($Help) {
        Show-Help
        exit 0
    }
    
    # Check if already installed
    $existingInstall = Get-Command $BinaryName -ErrorAction SilentlyContinue
    if ($existingInstall -and !$Force) {
        Write-Warning "$BinaryName is already installed at: $($existingInstall.Source)"
        $response = Read-Host "Reinstall? (y/N)"
        if ($response -notmatch '^[Yy]$') {
            exit 0
        }
    }
    
    # Check if running as administrator
    if (!(Test-Administrator)) {
        Write-Error "This script requires administrator privileges."
        Write-Info "Please run PowerShell as Administrator and try again."
        exit 1
    }
    
    # Ask about GUI installation if not specified
    if (!$SkipGui -and !$InstallQt) {
        Write-Host ""
        Write-Info "GoRDP now includes multiple GUI options:"
        Write-Host "  1. Command-line interface (CLI)"
        Write-Host "  2. Go-based GUI (recommended)"
        Write-Host "  3. Qt C++ GUI (advanced features)"
        Write-Host ""
        $response = Read-Host "Install GUI components? (Y/n)"
        if ($response -match '^[Nn]$') {
            $SkipGui = $true
        }
        else {
            $response = Read-Host "Install Qt GUI? (y/N)"
            if ($response -match '^[Yy]$') {
                $InstallQt = $true
            }
        }
    }
    
    # Try package managers first
    if (Install-WithChocolatey) {
        Write-Success "Installation completed using Chocolatey!"
        exit 0
    }
    
    if (Install-WithScoop) {
        Write-Success "Installation completed using Scoop!"
        exit 0
    }
    
    # Try Go installation
    if (Get-Command go -ErrorAction SilentlyContinue) {
        Write-Info "Go is available. Would you like to install using Go?"
        $response = Read-Host "Install with Go? (Y/n)"
        if ($response -notmatch '^[Nn]$') {
            Install-WithGo
            exit 0
        }
    }
    
    # Ask about building from source
    if (Get-Command go -ErrorAction SilentlyContinue -and Get-Command git -ErrorAction SilentlyContinue) {
        Write-Info "Go and Git are available. Would you like to build from source?"
        $response = Read-Host "Build from source? (y/N)"
        if ($response -match '^[Yy]$') {
            Build-FromSource
            exit 0
        }
    }
    
    # Download and install binary
    if ($Version -eq "latest") {
        $Version = Get-LatestVersion
    }
    
    $tempFile = Download-Binary -Version $Version -BinaryName $BinaryName
    Install-Binary -TempFile $tempFile -BinaryName $BinaryName
    
    # Download and install GUI binary if not skipped
    if (!$SkipGui) {
        Write-Info "Downloading GUI binary..."
        $guiTempFile = Download-Binary -Version $Version -BinaryName $GuiBinaryName
        Install-Binary -TempFile $guiTempFile -BinaryName $GuiBinaryName
    }
    
    Write-Host ""
    Write-Success "Installation completed successfully!"
    Write-Host ""
    Write-Host "Next steps:"
    Write-Host "1. Run '$BinaryName --help' to see available options"
    if (!$SkipGui) {
        Write-Host "2. Run '$GuiBinaryName' to start the GUI"
    }
    Write-Host "3. Check the documentation at https://github.com/$Repo"
    Write-Host "4. Try the examples in the repository"
    Write-Host ""
    if (!$SkipGui) {
        Write-Host "GUI Features:"
        Write-Host "- Multi-monitor support"
        Write-Host "- Virtual channels (clipboard, audio, device redirection)"
        Write-Host "- Performance monitoring"
        Write-Host "- Connection history and favorites"
        Write-Host "- Plugin system"
        Write-Host "- Advanced security features"
    }
    Write-Host ""
    Write-Host "Note: You may need to restart your terminal or computer for PATH changes to take effect."
}

# Run main function
Main 