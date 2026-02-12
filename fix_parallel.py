#!/usr/bin/env python3
"""
Fix duplicate t.Parallel() calls in Go test files.
This script removes duplicate consecutive t.Parallel() calls inside t.Run() subtests.
"""

import re
import sys
from pathlib import Path


def fix_duplicate_parallel(content: str) -> tuple[str, int]:
    """
    Remove duplicate consecutive t.Parallel() calls.
    Returns (fixed_content, num_fixes)
    """
    # Pattern to match duplicate t.Parallel() calls (possibly with different whitespace)
    pattern = r"(\s*t\.Parallel\(\)\s*\n)(\s*t\.Parallel\(\)\s*\n)"

    fixes = 0

    def replace_func(match):
        nonlocal fixes
        fixes += 1
        # Keep only the first t.Parallel() call with its original indentation
        return match.group(1)

    fixed = re.sub(pattern, replace_func, content)
    return fixed, fixes


def process_file(filepath: Path) -> bool:
    """
    Process a single test file.
    Returns True if file was modified.
    """
    try:
        content = filepath.read_text()
        fixed_content, num_fixes = fix_duplicate_parallel(content)

        if num_fixes > 0:
            filepath.write_text(fixed_content)
            print(f"✓ {filepath}: Fixed {num_fixes} duplicate t.Parallel() calls")
            return True
        return False
    except Exception as e:
        print(f"✗ {filepath}: Error - {e}", file=sys.stderr)
        return False


def main():
    # Find all test files
    test_files = list(Path(".").rglob("*_test.go"))

    if not test_files:
        print("No test files found in current directory")
        return 1

    print(f"Found {len(test_files)} test files")
    print("Scanning for duplicate t.Parallel() calls...\n")

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
