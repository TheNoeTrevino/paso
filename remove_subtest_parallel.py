#!/usr/bin/env python3
"""
Remove t.Parallel() from subtests that share parent test state.

This script identifies test functions that:
1. Have SetupCLITest or similar setup in the parent
2. Have subtests (t.Run) that call t.Parallel()

And removes the t.Parallel() from those subtests to prevent race conditions.
"""

import re
import sys
from pathlib import Path
from typing import List, Tuple


def has_shared_setup(content: str, func_start: int, func_end: int) -> bool:
    """Check if a test function has shared setup before subtests."""
    func_content = content[func_start:func_end]

    # Look for setup patterns before the first t.Run
    first_t_run = func_content.find("t.Run(")
    if first_t_run == -1:
        return False

    before_subtests = func_content[:first_t_run]

    # Patterns that indicate shared state
    setup_patterns = [
        r"SetupCLITest\(t\)",
        r":=.*Create.*\(t,.*db",  # CreateTestProject, CreateTestColumn, etc.
        r"projectID\s*:=",
        r"columnID\s*:=",
        r"labelID\s*:=",
        r"taskID\s*:=",
        r"db,\s*app\s*:=",
    ]

    for pattern in setup_patterns:
        if re.search(pattern, before_subtests):
            return True

    return False


def remove_parallel_from_subtests(
    content: str, func_start: int, func_end: int
) -> Tuple[str, int]:
    """
    Remove t.Parallel() from subtests within a function.
    Returns (modified_content, num_removals)
    """
    func_content = content[func_start:func_end]

    # Pattern to match t.Run with t.Parallel() inside
    # We need to be careful to only match the t.Parallel() that's directly inside t.Run
    pattern = (
        r"(t\.Run\([^,]+,\s*func\(t \*testing\.T\)\s*\{)\s*\n\s*t\.Parallel\(\)\s*\n"
    )

    removals = 0

    def replace_func(match):
        nonlocal removals
        removals += 1
        # Keep the t.Run line but remove t.Parallel()
        return match.group(1) + "\n"

    modified_func = re.sub(pattern, replace_func, func_content)

    if removals > 0:
        new_content = content[:func_start] + modified_func + content[func_end:]
        return new_content, removals

    return content, 0


def find_test_functions(content: str) -> List[Tuple[int, int, str]]:
    """Find all test functions and their positions. Returns [(start, end, name)]."""
    functions = []

    # Find all test function definitions
    pattern = r"^func (Test\w+)\(t \*testing\.T\) \{"

    for match in re.finditer(pattern, content, re.MULTILINE):
        func_name = match.group(1)
        func_start = match.start()

        # Find the end of the function (matching braces)
        brace_count = 0
        i = match.end() - 1  # Start at the opening brace

        while i < len(content):
            if content[i] == "{":
                brace_count += 1
            elif content[i] == "}":
                brace_count -= 1
                if brace_count == 0:
                    functions.append((func_start, i + 1, func_name))
                    break
            i += 1

    return functions


def process_file(filepath: Path) -> bool:
    """Process a single test file. Returns True if modified."""
    try:
        content = filepath.read_text()
        original_content = content

        functions = find_test_functions(content)
        total_removals = 0

        # Process functions in reverse order to maintain positions
        for func_start, func_end, func_name in reversed(functions):
            if has_shared_setup(content, func_start, func_end):
                content, removals = remove_parallel_from_subtests(
                    content, func_start, func_end
                )
                if removals > 0:
                    total_removals += removals
                    print(
                        f"  {func_name}: removed {removals} subtest t.Parallel() calls"
                    )

        if total_removals > 0:
            filepath.write_text(content)
            print(f"✓ {filepath}: Fixed {total_removals} subtests")
            return True

        return False

    except Exception as e:
        print(f"✗ {filepath}: Error - {e}", file=sys.stderr)
        return False


def main():
    test_files = list(Path(".").rglob("*_test.go"))

    if not test_files:
        print("No test files found")
        return 1

    print(f"Found {len(test_files)} test files")
    print("Looking for tests with shared state and parallel subtests...\n")

    modified = 0
    for test_file in sorted(test_files):
        if process_file(test_file):
            modified += 1

    print(f"\n{'=' * 60}")
    print(f"Modified {modified} file(s)")
    print(f"{'=' * 60}")

    return 0


if __name__ == "__main__":
    sys.exit(main())
