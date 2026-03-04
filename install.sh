#!/bin/bash
#
# PicoClaw Installer Script
# Installs the picoclaw binary with Kimi 2.5 support to ~/.local/bin
#

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🦞 PicoClaw Installer${NC}"
echo ""

# Determine install directory
INSTALL_DIR="${INSTALL_PREFIX:-$HOME/.local/bin}"
PICOCLAW_HOME="${PICOCLAW_HOME:-$HOME/.picoclaw}"

# Check if binary exists
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BINARY_PATH="$SCRIPT_DIR/picoclaw"
LAUNCHER_PATH="$SCRIPT_DIR/picoclaw-launcher"

if [ ! -f "$BINARY_PATH" ]; then
    echo -e "${RED}Error: picoclaw binary not found at $BINARY_PATH${NC}"
    echo "Please run 'make build' first or place the compiled binary in the current directory."
    exit 1
fi

# Create install directory if needed
if [ ! -d "$INSTALL_DIR" ]; then
    echo -e "${YELLOW}Creating install directory: $INSTALL_DIR${NC}"
    mkdir -p "$INSTALL_DIR"
fi

# Create picoclaw home directory
if [ ! -d "$PICOCLAW_HOME" ]; then
    echo -e "${YELLOW}Creating picoclaw home directory: $PICOCLAW_HOME${NC}"
    mkdir -p "$PICOCLAW_HOME"
fi

# Copy main binary
echo -e "${BLUE}Installing picoclaw to $INSTALL_DIR...${NC}"
cp "$BINARY_PATH" "$INSTALL_DIR/picoclaw"
chmod +x "$INSTALL_DIR/picoclaw"

# Copy launcher if it exists
if [ -f "$LAUNCHER_PATH" ]; then
    echo -e "${BLUE}Installing picoclaw-launcher to $INSTALL_DIR...${NC}"
    cp "$LAUNCHER_PATH" "$INSTALL_DIR/picoclaw-launcher"
    chmod +x "$INSTALL_DIR/picoclaw-launcher"
    INSTALL_LAUNCHER=true
else
    echo -e "${YELLOW}Note: picoclaw-launcher not found, skipping...${NC}"
    INSTALL_LAUNCHER=false
fi

# Check if install dir is in PATH
if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
    echo ""
    echo -e "${YELLOW}⚠️  Warning: $INSTALL_DIR is not in your PATH${NC}"
    echo ""
    echo "Add the following line to your ~/.bashrc or ~/.zshrc:"
    echo "    export PATH=\"$INSTALL_DIR:\$PATH\""
    echo ""
    echo "Then run: source ~/.bashrc (or source ~/.zshrc)"
fi

# Show success message
echo ""
echo -e "${GREEN}✅ PicoClaw installed successfully!${NC}"
echo ""
echo -e "${BLUE}Installation details:${NC}"
echo "  Binary: $INSTALL_DIR/picoclaw"
if [ "$INSTALL_LAUNCHER" = true ]; then
    echo "  Launcher: $INSTALL_DIR/picoclaw-launcher"
fi
echo "  Config: $PICOCLAW_HOME/config.json"
echo "  Workspace: $PICOCLAW_HOME/workspace"
echo ""

# Verify installation
if command -v picoclaw &> /dev/null; then
    echo -e "${GREEN}Verification:${NC}"
    picoclaw version
elif [ -f "$INSTALL_DIR/picoclaw" ]; then
    echo -e "${YELLOW}Note: picoclaw is installed but not in your current PATH.${NC}"
    echo "Run: $INSTALL_DIR/picoclaw version"
fi

echo ""
echo -e "${BLUE}Next steps:${NC}"
echo "  1. Ensure $PICOCLAW_HOME/config.json has your Kimi API key:"
echo "     'api_key': 'sk-kimi-...'"
echo "  2. Run: picoclaw agent -m 'Hello'"
if [ "$INSTALL_LAUNCHER" = true ]; then
    echo "  3. Run: picoclaw-launcher (for web-based config editor)"
fi
echo ""
