#!/bin/bash
# Paso installation script

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Default installation directory
INSTALL_DIR="${INSTALL_DIR:-$HOME/.local/bin}"

echo -e "${GREEN}Installing Paso - Terminal Kanban Board${NC}"
echo "==========================================="
echo ""

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo -e "${RED}Error: Go is not installed${NC}"
    echo "Please install Go from https://golang.org/dl/"
    exit 1
fi

echo "✓ Go found: $(go version)"
echo ""

# Build the main CLI
echo "Building paso CLI..."
go build -o bin/paso \
    -ldflags "-X main.version=$(git describe --tags --always --dirty 2>/dev/null || echo 'dev') \
              -X main.commit=$(git rev-parse --short HEAD 2>/dev/null || echo 'none') \
              -X main.date=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
    .

if [ $? -ne 0 ]; then
    echo -e "${RED}✗ Build failed${NC}"
    exit 1
fi
echo -e "${GREEN}✓ Built paso CLI${NC}"

# Build the daemon (optional)
if [ -d "daemon" ]; then
    echo "Building paso-daemon..."
    go build -o bin/paso-daemon \
        -ldflags "-X main.version=$(git describe --tags --always --dirty 2>/dev/null || echo 'dev') \
                  -X main.commit=$(git rev-parse --short HEAD 2>/dev/null || echo 'none') \
                  -X main.date=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
        ./daemon

    if [ $? -ne 0 ]; then
        echo -e "${YELLOW}⚠ Daemon build failed (optional)${NC}"
    else
        echo -e "${GREEN}✓ Built paso-daemon${NC}"
    fi
fi

echo ""

# Create installation directory if it doesn't exist
mkdir -p "$INSTALL_DIR"

# Install binaries
echo "Installing to $INSTALL_DIR..."

# Install paso
if [ -f "bin/paso" ]; then
    cp bin/paso "$INSTALL_DIR/"
    chmod +x "$INSTALL_DIR/paso"
    echo -e "${GREEN}✓ Installed paso to $INSTALL_DIR/paso${NC}"
else
    echo -e "${RED}✗ Failed to find bin/paso${NC}"
    exit 1
fi

# Install daemon if it was built
if [ -f "bin/paso-daemon" ]; then
    cp bin/paso-daemon "$INSTALL_DIR/"
    chmod +x "$INSTALL_DIR/paso-daemon"
    echo -e "${GREEN}✓ Installed paso-daemon to $INSTALL_DIR/paso-daemon${NC}"
fi

echo ""

# Check if installation directory is in PATH
if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
    echo -e "${YELLOW}⚠ Note: $INSTALL_DIR is not in your PATH${NC}"
    echo ""
    echo "Add this line to your ~/.bashrc or ~/.zshrc:"
    echo ""
    echo "    export PATH=\"$INSTALL_DIR:\$PATH\""
    echo ""
    echo "Then run: source ~/.bashrc (or source ~/.zshrc)"
    echo ""
else
    # Verify installation
    if command -v paso &> /dev/null; then
        echo -e "${GREEN}✓ Installation successful!${NC}"
        echo ""
        echo "Installed version:"
        paso --version
        echo ""
    fi
fi

# Create data directory
mkdir -p ~/.paso
echo -e "${GREEN}✓ Created data directory at ~/.paso${NC}"

# Optionally copy example config
if [ -f "config.example.yaml" ]; then
    if [ ! -f ~/.config/paso/config.yaml ]; then
        mkdir -p ~/.config/paso
        cp config.example.yaml ~/.config/paso/config.yaml
        echo -e "${GREEN}✓ Created config file at ~/.config/paso/config.yaml${NC}"
    fi
fi

echo ""
echo "Quick start:"
echo "  paso               # Show help"
echo "  paso tui           # Launch interactive TUI"
echo "  paso project create --title=\"My Project\""
echo ""
echo "Set up shell completion (optional):"
echo "  paso completion bash >> ~/.bashrc                        # Bash"
echo "  paso completion zsh > \"\${fpath[1]}/_paso\"                # Zsh"
echo "  paso completion fish > ~/.config/fish/completions/paso.fish  # Fish"
echo ""
echo -e "${GREEN}Installation complete!${NC}"
