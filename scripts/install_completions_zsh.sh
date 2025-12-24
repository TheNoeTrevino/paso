#!/bin/bash
# Install zsh completions for paso

set -e

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

COMPLETION_DIR="$HOME/.zsh/completion"
COMPLETION_FILE="$COMPLETION_DIR/_paso"

echo "Installing paso zsh completions..."
echo ""

# Create completion directory if it doesn't exist
if [ ! -d "$COMPLETION_DIR" ]; then
    echo -e "${YELLOW}Creating completion directory: $COMPLETION_DIR${NC}"
    mkdir -p "$COMPLETION_DIR"
fi

# Generate completion file
echo "Generating completion script..."
./bin/paso completion zsh > "$COMPLETION_FILE"

if [ -f "$COMPLETION_FILE" ]; then
    echo -e "${GREEN}✓ Completion file created: $COMPLETION_FILE${NC}"
    echo ""
    echo "To activate completions, reload your shell:"
    echo "  exec zsh"
    echo ""
    echo "Or source your zshrc:"
    echo "  source ~/.zshrc"
    echo ""
    echo "Test it with:"
    echo "  paso <TAB>"
else
    echo "Failed to create completion file"
    exit 1
fi
