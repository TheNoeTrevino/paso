#!/bin/bash
# Paso Complete Installation Script
# Builds, installs, and configures everything needed to run Paso

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
BOLD='\033[1m'
NC='\033[0m' # No Color

# Configuration
INSTALL_DIR="${INSTALL_DIR:-$HOME/.local/bin}"
DATA_DIR="$HOME/.paso"
CONFIG_DIR="$HOME/.config/paso"
SYSTEMD_USER_DIR="${XDG_CONFIG_HOME:-$HOME/.config}/systemd/user"

echo ""
echo -e "${BOLD}${GREEN}╔════════════════════════════════════════╗${NC}"
echo -e "${BOLD}${GREEN}║   Paso Complete Installation Script    ║${NC}"
echo -e "${BOLD}${GREEN}║   Terminal Kanban Board                ║${NC}"
echo -e "${BOLD}${GREEN}╚════════════════════════════════════════╝${NC}"
echo ""

# ============================================================================
# Step 1: Check Prerequisites
# ============================================================================

echo -e "${BLUE}[1/7] Checking prerequisites...${NC}"
echo ""

# Check if Go is installed
if ! command -v go &>/dev/null; then
  echo -e "${RED}✗ Error: Go is not installed${NC}"
  echo "Please install Go from https://golang.org/dl/"
  exit 1
fi
echo -e "${GREEN}✓ Go found: $(go version)${NC}"

# Check if systemctl is available (for optional systemd setup)
SYSTEMD_AVAILABLE=false
if command -v systemctl &>/dev/null; then
  SYSTEMD_AVAILABLE=true
  echo -e "${GREEN}✓ systemd available${NC}"
else
  echo -e "${YELLOW}⚠ systemd not found (daemon autostart will not be available)${NC}"
fi

echo ""

# ============================================================================
# Step 2: Build Binaries
# ============================================================================

echo -e "${BLUE}[2/7] Building binaries...${NC}"
echo ""

# Get version info from git
VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo 'dev')
COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo 'none')
BUILD_DATE=$(date -u +%Y-%m-%dT%H:%M:%SZ)

echo "Building paso CLI..."
go build -o bin/paso \
  -ldflags "-X main.version=$VERSION -X main.commit=$COMMIT -X main.date=$BUILD_DATE" \
  .

if [ $? -ne 0 ]; then
  echo -e "${RED}✗ Build failed${NC}"
  exit 1
fi
echo -e "${GREEN}✓ Built paso CLI${NC}"

# Build the daemon
if [ -d "cmd/daemon" ]; then
  echo "Building paso-daemon..."
  go build -o bin/paso-daemon \
    -ldflags "-X main.version=$VERSION -X main.commit=$COMMIT -X main.date=$BUILD_DATE" \
    ./cmd/daemon

  if [ $? -ne 0 ]; then
    echo -e "${YELLOW}⚠ Daemon build failed (optional component)${NC}"
    DAEMON_BUILT=false
  else
    echo -e "${GREEN}✓ Built paso-daemon${NC}"
    DAEMON_BUILT=true
  fi
else
  echo -e "${YELLOW}⚠ Daemon source not found (skipping)${NC}"
  DAEMON_BUILT=false
fi

echo ""

# ============================================================================
# Step 3: Install Binaries
# ============================================================================

echo -e "${BLUE}[3/7] Installing binaries to $INSTALL_DIR...${NC}"
echo ""

# Create installation directory
mkdir -p "$INSTALL_DIR"

# Install paso CLI
if [ -f "bin/paso" ]; then
  cp bin/paso "$INSTALL_DIR/"
  chmod +x "$INSTALL_DIR/paso"
  echo -e "${GREEN}✓ Installed paso to $INSTALL_DIR/paso${NC}"
else
  echo -e "${RED}✗ Failed to find bin/paso${NC}"
  exit 1
fi

# Install daemon if it was built
if [ "$DAEMON_BUILT" = true ] && [ -f "bin/paso-daemon" ]; then
  cp bin/paso-daemon "$INSTALL_DIR/"
  chmod +x "$INSTALL_DIR/paso-daemon"
  echo -e "${GREEN}✓ Installed paso-daemon to $INSTALL_DIR/paso-daemon${NC}"
fi

# Check if installation directory is in PATH
if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
  echo ""
  echo -e "${YELLOW}⚠ Note: $INSTALL_DIR is not in your PATH${NC}"
  echo ""
  echo "Add this line to your ~/.bashrc or ~/.zshrc:"
  echo -e "${BOLD}    export PATH=\"$INSTALL_DIR:\$PATH\"${NC}"
  echo ""
  echo "Then run: source ~/.bashrc (or source ~/.zshrc)"
  echo ""

  # Ask if user wants to add to PATH automatically
  read -p "Would you like to add it to your PATH now? (y/N) " -n 1 -r
  echo
  if [[ $REPLY =~ ^[Yy]$ ]]; then
    # Detect shell and add to appropriate rc file
    if [ -n "$ZSH_VERSION" ]; then
      RC_FILE="$HOME/.zshrc"
    elif [ -n "$BASH_VERSION" ]; then
      RC_FILE="$HOME/.bashrc"
    else
      RC_FILE="$HOME/.profile"
    fi

    echo "" >>"$RC_FILE"
    echo "# Paso installation" >>"$RC_FILE"
    echo "export PATH=\"$INSTALL_DIR:\$PATH\"" >>"$RC_FILE"
    echo -e "${GREEN}✓ Added to $RC_FILE${NC}"
    echo "Run: source $RC_FILE"

    # Export for current session
    export PATH="$INSTALL_DIR:$PATH"
  fi
fi

echo ""

# ============================================================================
# Step 4: Create Directories
# ============================================================================

echo -e "${BLUE}[4/7] Creating directories...${NC}"
echo ""

mkdir -p "$DATA_DIR"
echo -e "${GREEN}✓ Created data directory: $DATA_DIR${NC}"

mkdir -p "$CONFIG_DIR"
echo -e "${GREEN}✓ Created config directory: $CONFIG_DIR${NC}"

# Copy example config if it exists and user doesn't have one
if [ -f "config.example.yaml" ] && [ ! -f "$CONFIG_DIR/config.yaml" ]; then
  cp config.example.yaml "$CONFIG_DIR/config.yaml"
  echo -e "${GREEN}✓ Created config file: $CONFIG_DIR/config.yaml${NC}"
fi

echo ""

# ============================================================================
# Step 5: Install Zsh Completions (Optional)
# ============================================================================

echo -e "${BLUE}[5/7] Shell completions...${NC}"
echo ""

# Check if zsh is available
if command -v zsh &>/dev/null; then
  read -p "Install zsh completions? (Y/n) " -n 1 -r
  echo
  if [[ ! $REPLY =~ ^[Nn]$ ]]; then
    COMPLETION_DIR="$HOME/.zsh/completion"
    COMPLETION_FILE="$COMPLETION_DIR/_paso"

    mkdir -p "$COMPLETION_DIR"
    # Use the freshly built binary, not an installed one
    ./bin/paso completion zsh >"$COMPLETION_FILE"

    if [ -f "$COMPLETION_FILE" ]; then
      echo -e "${GREEN}✓ Installed zsh completions to $COMPLETION_FILE${NC}"
      echo "  Reload your shell to activate: exec zsh"
    else
      echo -e "${YELLOW}⚠ Failed to create completion file${NC}"
    fi
  else
    echo "Skipping zsh completions"
  fi
else
  echo "Zsh not found, skipping completions"
  echo ""
  echo "For bash completions, run:"
  echo "  paso completion bash >> ~/.bashrc"
fi

echo ""

# ============================================================================
# Step 6: Setup Systemd Service (Optional)
# ============================================================================

echo -e "${BLUE}[6/7] Systemd service setup...${NC}"
echo ""

if [ "$SYSTEMD_AVAILABLE" = true ] && [ "$DAEMON_BUILT" = true ]; then
  echo "The systemd service will:"
  echo "  • Start paso-daemon automatically on login"
  echo "  • Enable real-time updates across terminal sessions"
  echo "  • Restart the daemon automatically if it crashes"
  echo ""

  read -p "Install and enable systemd service? (Y/n) " -n 1 -r
  echo

  if [[ ! $REPLY =~ ^[Nn]$ ]]; then
    echo ""
    echo -e "${YELLOW}Setting up systemd user service...${NC}"
    echo "This will:"
    echo "  1. Create service file in $SYSTEMD_USER_DIR"
    echo "  2. Enable the service to start on login"
    echo "  3. Start the service immediately"
    echo ""

    # Create systemd user directory
    mkdir -p "$SYSTEMD_USER_DIR"

    # Write the service file with dynamic ExecStart path
    SERVICE_FILE="$SYSTEMD_USER_DIR/paso.service"
    cat >"$SERVICE_FILE" <<EOF
[Unit]
Description=Paso Daemon - Terminal Kanban Board Real-time Sync
After=network.target

[Service]
Type=simple
ExecStart=$INSTALL_DIR/paso-daemon
Restart=on-failure
RestartSec=5

# Security hardening
PrivateTmp=yes
NoNewPrivileges=yes
ProtectSystem=strict
ReadWritePaths=$DATA_DIR

# Logging
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=default.target
EOF

    echo -e "${GREEN}✓ Created service file: $SERVICE_FILE${NC}"

    # Reload systemd daemon
    systemctl --user daemon-reload
    echo -e "${GREEN}✓ Reloaded systemd daemon${NC}"

    # Enable the service
    systemctl --user enable paso.service
    echo -e "${GREEN}✓ Enabled paso.service (will start on login)${NC}"

    # Start the service
    systemctl --user start paso.service
    echo -e "${GREEN}✓ Started paso.service${NC}"

    echo ""
    echo -e "${GREEN}Systemd service installed successfully!${NC}"
    echo ""
    echo "Useful commands:"
    echo "  systemctl --user status paso      # View status"
    echo "  journalctl --user -u paso -f      # View logs"
    echo "  systemctl --user restart paso     # Restart service"
    echo "  systemctl --user stop paso        # Stop service"
    echo "  systemctl --user disable paso     # Disable autostart"
  else
    echo "Skipping systemd service"
    echo ""
    echo "To start the daemon manually, run:"
    echo "  paso-daemon &"
  fi
elif [ "$DAEMON_BUILT" = false ]; then
  echo "Daemon not available, skipping systemd setup"
else
  echo "systemd not available, skipping service setup"
  echo ""
  echo "To start the daemon manually, run:"
  echo "  paso-daemon &"
fi

echo ""

# ============================================================================
# Step 7: Verify Installation
# ============================================================================

echo -e "${BLUE}[7/7] Verifying installation...${NC}"
echo ""

if command -v paso &>/dev/null; then
  echo -e "${GREEN}✓ paso command is available${NC}"
  paso --version
else
  echo -e "${YELLOW}⚠ paso command not found in PATH${NC}"
  echo "You may need to reload your shell or add $INSTALL_DIR to PATH"
  echo ""
  echo "Installed version (from build):"
  ./bin/paso --version
fi

echo ""

# ============================================================================
# Installation Complete
# ============================================================================

echo ""
echo -e "${BOLD}${GREEN}╔════════════════════════════════════════╗${NC}"
echo -e "${BOLD}${GREEN}║     Installation Complete! 🎉          ║${NC}"
echo -e "${BOLD}${GREEN}╚════════════════════════════════════════╝${NC}"
echo ""
echo -e "${BOLD}Quick Start:${NC}"
echo "  paso                              # Show help"
echo "  paso tui                          # Launch interactive TUI"
echo "  paso project create --title=\"My Project\""
echo ""
echo -e "${BOLD}Configuration:${NC}"
echo "  Data:   $DATA_DIR"
echo "  Config: $CONFIG_DIR/config.yaml"
echo ""
echo -e "${BOLD}Documentation:${NC}"
echo "  paso --help                       # Command help"
echo "  paso task --help                  # Task commands"
echo "  paso completion --help            # Shell completion"
echo ""

if [ "$SYSTEMD_AVAILABLE" = true ] && systemctl --user is-enabled paso.service &>/dev/null; then
  echo -e "${GREEN}✓ Daemon is running and will start automatically on login${NC}"
  echo ""
fi

echo "Happy tasking! 📝"
echo ""
