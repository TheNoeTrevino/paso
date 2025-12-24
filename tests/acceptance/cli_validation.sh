#!/bin/bash
# Paso CLI Acceptance Tests

# Don't exit on error - we want to run all tests
# set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "🧪 Paso CLI Acceptance Tests"
echo "=============================="
echo ""

# Setup: Create test database
export PASO_DB_PATH="/tmp/paso_test_$(date +%s).db"
rm -f "$PASO_DB_PATH"

# Track test results
TESTS_PASSED=0
TESTS_FAILED=0
FAILED_TESTS=()

# Helper function to run a test
run_test() {
    local test_num=$1
    local test_name=$2
    local test_cmd=$3

    echo -n "Test $test_num: $test_name... "

    if eval "$test_cmd" > /dev/null 2>&1; then
        echo -e "${GREEN}✓ PASS${NC}"
        ((TESTS_PASSED++))
        return 0
    else
        echo -e "${RED}✗ FAIL${NC}"
        ((TESTS_FAILED++))
        FAILED_TESTS+=("Test $test_num: $test_name")
        return 1
    fi
}

# Test 1: Help text (default behavior)
run_test 1 "paso shows help text" "paso | grep -q 'Usage:'"

# Test 2: Version flag
run_test 2 "paso --version works" "paso --version | grep -q 'version'"

# Test 3: TUI command exists
run_test 3 "paso tui command exists" "paso tui --help | grep -q 'Launch'"

# Test 4: Project creation (quiet mode)
PROJECT_ID=$(paso project create --title="Test Project" --quiet 2>/dev/null)
if [[ "$PROJECT_ID" =~ ^[0-9]+$ ]]; then
    run_test 4 "Project creation returns ID" "true"
else
    run_test 4 "Project creation returns ID" "false"
fi

# Test 5: Project list JSON output is valid
run_test 5 "JSON output is parseable" "paso project list --json | jq -e '.success' > /dev/null"

# Test 6: Task creation with project
if [ -n "$PROJECT_ID" ]; then
    TASK_ID=$(paso task create --title="Test Task" --project=$PROJECT_ID --quiet 2>/dev/null)
    if [[ "$TASK_ID" =~ ^[0-9]+$ ]]; then
        run_test 6 "Task creation returns ID" "true"
    else
        run_test 6 "Task creation returns ID" "false"
    fi
else
    run_test 6 "Task creation returns ID" "false"
    TASK_ID=""
fi

# Test 7: Task creation with all options
if [ -n "$PROJECT_ID" ]; then
    FEATURE_ID=$(paso task create \
        --title="Feature task" \
        --description="Test description" \
        --type=feature \
        --priority=high \
        --project=$PROJECT_ID \
        --quiet 2>/dev/null)

    if [[ "$FEATURE_ID" =~ ^[0-9]+$ ]]; then
        run_test 7 "Task with options created" "true"
    else
        run_test 7 "Task with options created" "false"
        FEATURE_ID=""
    fi
else
    run_test 7 "Task with options created" "false"
    FEATURE_ID=""
fi

# Test 8: Parent-child relationship
if [ -n "$PROJECT_ID" ] && [ -n "$FEATURE_ID" ]; then
    CHILD_ID=$(paso task create --title="Child task" --project=$PROJECT_ID --parent=$FEATURE_ID --quiet 2>/dev/null)
    if [[ "$CHILD_ID" =~ ^[0-9]+$ ]]; then
        run_test 8 "Subtask creation works" "true"
    else
        run_test 8 "Subtask creation works" "false"
    fi
else
    run_test 8 "Subtask creation works" "false"
fi

# Test 9: Task list
if [ -n "$PROJECT_ID" ]; then
    run_test 9 "Task list works" "paso task list --project=$PROJECT_ID --json | jq -e '.tasks | length > 0'"
else
    run_test 9 "Task list works" "false"
fi

# Test 10: Column creation
if [ -n "$PROJECT_ID" ]; then
    COLUMN_ID=$(paso column create --name="Review" --project=$PROJECT_ID --quiet 2>/dev/null)
    if [[ "$COLUMN_ID" =~ ^[0-9]+$ ]]; then
        run_test 10 "Column creation works" "true"
    else
        run_test 10 "Column creation works" "false"
    fi
else
    run_test 10 "Column creation works" "false"
fi

# Test 11: Column list
if [ -n "$PROJECT_ID" ]; then
    run_test 11 "Column list works" "paso column list --project=$PROJECT_ID --json | jq -e '.columns | length >= 4'"
else
    run_test 11 "Column list works" "false"
fi

# Test 12: Label creation
if [ -n "$PROJECT_ID" ]; then
    LABEL_ID=$(paso label create --name="bug" --color="#FF0000" --project=$PROJECT_ID --quiet 2>/dev/null)
    if [[ "$LABEL_ID" =~ ^[0-9]+$ ]]; then
        run_test 12 "Label creation works" "true"
    else
        run_test 12 "Label creation works" "false"
    fi
else
    run_test 12 "Label creation works" "false"
    LABEL_ID=""
fi

# Test 13: Label attach
if [ -n "$FEATURE_ID" ] && [ -n "$LABEL_ID" ]; then
    if paso label attach --task-id=$FEATURE_ID --label-id=$LABEL_ID > /dev/null 2>&1; then
        run_test 13 "Label attach works" "true"
    else
        run_test 13 "Label attach works" "false"
    fi
else
    run_test 13 "Label attach works" "false"
fi

# Test 14: Error handling - non-existent project
paso task create --title="Test" --project=999 --quiet > /dev/null 2>&1
EXIT_CODE=$?
if [ $EXIT_CODE -eq 3 ]; then
    run_test 14 "Exit code 3 for not found" "true"
else
    run_test 14 "Exit code 3 for not found" "false"
fi

# Test 15: Error handling - invalid type
paso task create --title="Test" --type=invalid --project=1 --quiet > /dev/null 2>&1
EXIT_CODE=$?
if [ $EXIT_CODE -eq 5 ]; then
    run_test 15 "Exit code 5 for validation error" "true"
else
    run_test 15 "Exit code 5 for validation error" "false"
fi

# Test 16: Human-readable output
if [ -n "$PROJECT_ID" ]; then
    run_test 16 "Human-readable output works" "paso project list | grep -q 'Test Project'"
else
    run_test 16 "Human-readable output works" "false"
fi

# Test 17: Completion command
run_test 17 "Bash completion generation works" "paso completion bash | grep -q 'bash completion'"

# Test 18: Task update
if [ -n "$TASK_ID" ]; then
    if paso task update --id=$TASK_ID --title="Updated Task" > /dev/null 2>&1; then
        run_test 18 "Task update works" "true"
    else
        run_test 18 "Task update works" "false"
    fi
else
    run_test 18 "Task update works" "false"
fi

# Test 19: Task delete
if [ -n "$CHILD_ID" ]; then
    if paso task delete --id=$CHILD_ID --force > /dev/null 2>&1; then
        run_test 19 "Task delete works" "true"
    else
        run_test 19 "Task delete works" "false"
    fi
else
    run_test 19 "Task delete works" "false"
fi

# Test 20: Project delete
if [ -n "$PROJECT_ID" ]; then
    if paso project delete --id=$PROJECT_ID --force > /dev/null 2>&1; then
        run_test 20 "Project delete works" "true"
    else
        run_test 20 "Project delete works" "false"
    fi
else
    run_test 20 "Project delete works" "false"
fi

# Cleanup
rm -f "$PASO_DB_PATH"

echo ""
echo "=============================="
echo "Test Summary"
echo "=============================="
echo -e "${GREEN}Passed: $TESTS_PASSED${NC}"
echo -e "${RED}Failed: $TESTS_FAILED${NC}"
echo ""

if [ $TESTS_FAILED -gt 0 ]; then
    echo "Failed tests:"
    for test in "${FAILED_TESTS[@]}"; do
        echo -e "  ${RED}✗ $test${NC}"
    done
    echo ""
    exit 1
else
    echo -e "${GREEN}✅ All acceptance tests passed!${NC}"
    echo "CLI implementation is complete and agent-ready."
    exit 0
fi
