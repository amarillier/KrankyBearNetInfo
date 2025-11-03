# PowerShell script to compile Windows binaries for KrankyBear NetInfo
# Usage: .\compile.ps1 [-Arch <arch>] [-Clean]

param(
    [string]$Arch = "amd64",
    [switch]$Clean
)

$ErrorActionPreference = "Stop"

# Colors for output
function Write-ColorOutput($ForegroundColor, $Message) {
    $fc = $host.UI.RawUI.ForegroundColor
    $host.UI.RawUI.ForegroundColor = $ForegroundColor
    Write-Output $Message
    $host.UI.RawUI.ForegroundColor = $fc
}

Write-ColorOutput Cyan "=== Compiling KrankyBear netinfo for Windows ==="
Write-Output ""

# Clean build artifacts if requested
if ($Clean) {
    Write-ColorOutput Yellow "Cleaning build artifacts..."
    if (Test-Path "bin") {
        Remove-Item -Recurse -Force "bin"
    }
    if (Test-Path "netinfo.exe") {
        Remove-Item -Force "netinfo*.exe"
    }
    Write-Output ""
}

# Create bin directory if it doesn't exist
if (-not (Test-Path "bin")) {
    New-Item -ItemType Directory -Path "bin" | Out-Null
}

# Build flags for optimization
$ldflags = "-w -s"
$output = "bin\netinfo-$Arch.exe"

Write-ColorOutput Green "Building Windows $Arch binary..."
Write-Output "  Output: $output"
Write-Output "  Flags: $ldflags -trimpath"
Write-Output ""

$env:CGO_ENABLED = "1"
$env:GOOS = "windows"
$env:GOARCH = $Arch

try {
    go build -ldflags $ldflags -trimpath -o $output
    
    if ($LASTEXITCODE -eq 0) {
        Write-Output ""
        Write-ColorOutput Green "✓ Build successful!"
        Write-Output ""
        
        # Show file info
        $fileInfo = Get-Item $output
        $sizeKB = [math]::Round($fileInfo.Length / 1KB, 2)
        Write-Output "  File: $output"
        Write-Output "  Size: $sizeKB KB"
        Write-Output "  Date: $($fileInfo.LastWriteTime)"
        Write-Output ""
        
        Write-ColorOutput Cyan "=== Build Complete ==="
    } else {
        Write-ColorOutput Red "✗ Build failed with exit code $LASTEXITCODE"
        exit $LASTEXITCODE
    }
} catch {
    Write-ColorOutput Red "✗ Build error: $_"
    exit 1
} finally {
    # Clean up environment variables
    Remove-Item Env:\CGO_ENABLED -ErrorAction SilentlyContinue
    Remove-Item Env:\GOOS -ErrorAction SilentlyContinue
    Remove-Item Env:\GOARCH -ErrorAction SilentlyContinue
}

