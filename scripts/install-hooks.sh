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
cat >"$HOOK_PATH" <<'EOF'
#!/bin/bash
# Pre-commit hook for formatting staged Go files

set -e

# Get staged Go files
STAGED_GO_FILES=$(git diff --cached --name-only --diff-filter=ACM | grep '\.go$' || true)

if [ -z "$STAGED_GO_FILES" ]; then
    exit 0
fi

echo "Formatting staged Go files..."

# Format each file and re-stage it
for file in $STAGED_GO_FILES; do
    if [ -f "$file" ]; then
        gofmt -w "$file"
        git add "$file"
    fi
done

echo "✓ All staged Go files formatted"
EOF

# Make the hook executable
chmod +x "$HOOK_PATH"

echo -e "${GREEN}✓ Pre-commit hook installed successfully${NC}"
echo ""
echo "The hook will:"
echo "  • Format staged Go files with gofmt"
echo "  • Automatically re-stage formatted files"
echo ""
echo "Usage:"
echo "  • Normal commit: ${YELLOW}git commit${NC}"
echo "  • Bypass hook: ${YELLOW}git commit --no-verify${NC}"
echo "  • Manual format: ${YELLOW}gofmt -w <file>${NC}"
