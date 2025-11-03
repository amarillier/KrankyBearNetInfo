#!/bin/bash
# KrankyBear NetInfo - Linux Compile Script
# Builds Linux binaries for amd64, arm64, and 386 architectures

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${BLUE}=== KrankyBear NetInfo - Linux Compile Script ===${NC}"
echo ""

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo -e "${RED}Error: Go is not installed. Please install Go 1.21 or later.${NC}"
    exit 1
fi

# Create bin directory if it doesn't exist
mkdir -p bin

# Cleanup only Linux binaries (keep other platform binaries)
rm -f bin/netinfo-linux-*

echo -e "${BLUE}Building Linux binaries...${NC}"
echo ""

# Build flags
LDFLAGS="-w -s"

# Build for amd64
echo -e "${BLUE}Building Linux amd64...${NC}"
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="${LDFLAGS}" -trimpath -o bin/netinfo-linux-amd64
if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ Linux amd64 build successful${NC}"
    ls -lh bin/netinfo-linux-amd64
else
    echo -e "${RED}✗ Linux amd64 build failed${NC}"
    exit 1
fi

echo ""

# Build for arm64
echo -e "${BLUE}Building Linux arm64...${NC}"
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags="${LDFLAGS}" -trimpath -o bin/netinfo-linux-arm64
if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ Linux arm64 build successful${NC}"
    ls -lh bin/netinfo-linux-arm64
else
    echo -e "${YELLOW}⚠ Linux arm64 build failed (may require cross-compilation tools)${NC}"
fi

echo ""

# Build for 386
echo -e "${BLUE}Building Linux 386...${NC}"
CGO_ENABLED=0 GOOS=linux GOARCH=386 go build -ldflags="${LDFLAGS}" -trimpath -o bin/netinfo-linux-386
if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ Linux 386 build successful${NC}"
    ls -lh bin/netinfo-linux-386
else
    echo -e "${YELLOW}⚠ Linux 386 build failed (may require cross-compilation tools)${NC}"
fi

echo ""
echo -e "${GREEN}=== Build Complete ===${NC}"
echo ""
echo "Binaries created:"
ls -lh bin/netinfo-linux-* 2>/dev/null || echo "No Linux binaries found"
