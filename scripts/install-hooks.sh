#!/bin/bash
# Install pre-commit hook for paso project

set -e

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Get the git directory (handles worktrees)
GIT_DIR=$(git rev-parse --git-common-dir)

if [ ! -d "$GIT_DIR" ]; then
    echo "Error: Not in a git repository"
    exit 1
fi

HOOK_PATH="$GIT_DIR/hooks/pre-commit"

# Create the pre-commit hook
cat > "$HOOK_PATH" << 'EOF'
#!/bin/sh
# Pre-commit hook for formatting staged files

# Get the root directory of the git repository
GIT_ROOT=$(git rev-parse --show-toplevel)

# Path to the pre-commit binary
PRECOMMIT_BIN="$GIT_ROOT/bin/pre-commit"

# Build the pre-commit binary if it doesn't exist
if [ ! -f "$PRECOMMIT_BIN" ]; then
    echo "Building pre-commit hook binary..."
    if ! go build -o "$PRECOMMIT_BIN" "$GIT_ROOT/cmd/pre-commit"; then
        echo "Error: Failed to build pre-commit hook" >&2
        echo "Run 'go build -o bin/pre-commit ./cmd/pre-commit' to diagnose" >&2
        exit 1
    fi
    echo "Pre-commit hook binary built successfully"
fi

# Run the pre-commit binary
exec "$PRECOMMIT_BIN"
EOF

# Make the hook executable
chmod +x "$HOOK_PATH"

echo -e "${GREEN}✓ Pre-commit hook installed successfully${NC}"
echo ""
echo "The hook will:"
echo "  • Format staged Go files with gofmt"
echo "  • Automatically re-stage formatted files"
echo "  • Build on first commit (lazy build)"
echo ""
echo "Usage:"
echo "  • Normal commit: ${YELLOW}git commit${NC}"
echo "  • Bypass hook: ${YELLOW}git commit --no-verify${NC}"
echo "  • Manual format: ${YELLOW}gofmt -w <file>${NC}"
echo ""
echo "The hook binary will be built automatically on your first commit."
