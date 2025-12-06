#!/bin/bash

set -e

echo "Building paso..."
go build -o paso .

echo "Installing paso to ~/.local/bin..."
mkdir -p ~/.local/bin
cp paso ~/.local/bin/

echo "Cleaning up build artifacts..."
rm paso

echo ""
echo "✓ paso installed successfully!"
echo ""

# Check if ~/.local/bin is in PATH
if [[ ":$PATH:" != *":$HOME/.local/bin:"* ]]; then
    echo "⚠ Note: ~/.local/bin is not in your PATH"
    echo ""
    echo "Add this line to your ~/.bashrc or ~/.zshrc:"
    echo ""
    echo "    export PATH=\"\$HOME/.local/bin:\$PATH\""
    echo ""
    echo "Then run: source ~/.bashrc (or source ~/.zshrc)"
else
    echo "You can now run 'paso' from anywhere!"
fi
