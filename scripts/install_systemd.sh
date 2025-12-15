#!/usr/bin/env bash
#
# install_systemd.sh - Install paso daemon as a systemd user service
#
# This script sets up the paso daemon to run as a systemd user service,
# enabling automatic startup and management through systemctl.

set -e

# Trap errors and print useful message
trap 'echo "Error: Installation failed at line $LINENO. Exit code: $?" >&2' ERR

# Color output for better readability
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo "Installing paso daemon as systemd user service..."

# Create systemd user directory if it doesn't exist
SYSTEMD_USER_DIR="${XDG_CONFIG_HOME:-$HOME/.config}/systemd/user"
mkdir -p "$SYSTEMD_USER_DIR"
echo "✓ Created systemd user directory: $SYSTEMD_USER_DIR"

# Write the service file
SERVICE_FILE="$SYSTEMD_USER_DIR/paso.service"
cat > "$SERVICE_FILE" << 'EOF'
[Unit]
Description=Paso Daemon - Terminal Kanban Board Real-time Sync
After=network.target

[Service]
Type=simple
ExecStart=/usr/local/bin/paso-daemon
Restart=on-failure
RestartSec=5

# Security hardening
PrivateTmp=yes
NoNewPrivileges=yes
ProtectSystem=strict
ReadWritePaths=%h/.paso

# Logging
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=default.target
EOF

echo "✓ Created service file: $SERVICE_FILE"

# Reload systemd daemon
systemctl --user daemon-reload
echo "✓ Reloaded systemd daemon"

# Enable the service (start on login)
systemctl --user enable paso.service
echo "✓ Enabled paso.service (will start on login)"

# Start the service
systemctl --user start paso.service
echo "✓ Started paso.service"

# Print success message with usage instructions
echo ""
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}Paso daemon installed successfully!${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo -e "${BLUE}Useful commands:${NC}"
echo ""
echo "  View service status:"
echo "    systemctl --user status paso"
echo ""
echo "  View logs (follow mode):"
echo "    journalctl --user -u paso -f"
echo ""
echo "  View recent logs:"
echo "    journalctl --user -u paso -n 50"
echo ""
echo "  Stop service:"
echo "    systemctl --user stop paso"
echo ""
echo "  Restart service:"
echo "    systemctl --user restart paso"
echo ""
echo "  Disable service (prevent autostart):"
echo "    systemctl --user disable paso"
echo ""
echo -e "${GREEN}The service is now running and will start automatically on login.${NC}"
