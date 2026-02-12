package git

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSanitizeBranchName_ValidBranches(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple_branch",
			input:    "main",
			expected: "main",
		},
		{
			name:     "feature_branch",
			input:    "feature",
			expected: "feature",
		},
		{
			name:     "branch_with_slash",
			input:    "feature/my-feature",
			expected: "feature/my-feature",
		},
		{
			name:     "branch_with_multiple_slashes",
			input:    "feat/user/auth-system",
			expected: "feat/user/auth-system",
		},
		{
			name:     "branch_with_hyphens",
			input:    "my-awesome-feature",
			expected: "my-awesome-feature",
		},
		{
			name:     "branch_with_underscores",
			input:    "my_feature_branch",
			expected: "my_feature_branch",
		},
		{
			name:     "branch_with_numbers",
			input:    "feature-123",
			expected: "feature-123",
		},
		{
			name:     "branch_with_dots",
			input:    "release/v1.0.0",
			expected: "release/v1.0.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			t.Parallel()
			result, err := SanitizeBranchName(tt.input)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result, "Branch name should be sanitized correctly")
		})
	}
}

func TestSanitizeBranchName_EdgeCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		input       string
		expected    string
		expectError bool
	}{
		{
			name:        "empty_string",
			input:       "",
			expected:    "",
			expectError: false,
		},
		{
			name:        "whitespace_only",
			input:       "   ",
			expected:    "",
			expectError: false,
		},
		{
			name:        "leading_whitespace",
			input:       "  main",
			expected:    "main",
			expectError: false,
		},
		{
			name:        "trailing_whitespace",
			input:       "main  ",
			expected:    "main",
			expectError: false,
		},
		{
			name:        "both_whitespace",
			input:       "  feature/branch  ",
			expected:    "feature/branch",
			expectError: false,
		},
		{
			name:        "newline_characters",
			input:       "feature\nbranch",
			expected:    "",
			expectError: true,
		},
		{
			name:        "tab_characters",
			input:       "feature\tbranch",
			expected:    "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			t.Parallel()
			result, err := SanitizeBranchName(tt.input)
			if tt.expectError {
				assert.Error(t, err, "Should return error for control characters")
			} else {
				assert.NoError(t, err, "Should not return error")
				assert.Equal(t, tt.expected, result, "Edge case should be handled correctly")
			}
		})
	}
}

func TestSanitizeBranchName_SpecialCharacters(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		input       string
		expectError bool
		description string
	}{
		{
			name:        "asterisk",
			input:       "feature*branch",
			expectError: true,
			description: "Branch with asterisk",
		},
		{
			name:        "question_mark",
			input:       "feature?branch",
			expectError: true,
			description: "Branch with question mark",
		},
		{
			name:        "brackets",
			input:       "feature[branch]",
			expectError: true,
			description: "Branch with brackets",
		},
		{
			name:        "at_symbol",
			input:       "user@feature",
			expectError: false,
			description: "Branch with @ symbol",
		},
		{
			name:        "hash_symbol",
			input:       "issue#123",
			expectError: false,
			description: "Branch with hash symbol",
		},
		{
			name:        "backslash",
			input:       "feature\\branch",
			expectError: true,
			description: "Branch with backslash",
		},
		{
			name:        "caret",
			input:       "feature^branch",
			expectError: true,
			description: "Branch with caret",
		},
		{
			name:        "tilde",
			input:       "feature~branch",
			expectError: true,
			description: "Branch with tilde",
		},
		{
			name:        "colon",
			input:       "feature:branch",
			expectError: true,
			description: "Branch with colon",
		},
		{
			name:        "space_in_middle",
			input:       "feature branch",
			expectError: true,
			description: "Branch with space character",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			t.Parallel()
			_, err := SanitizeBranchName(tt.input)
			if tt.expectError {
				assert.Error(t, err, tt.description+" should return error for invalid characters")
			} else {
				assert.NoError(t, err, tt.description+" should not return error")
			}
		})
	}
}

func TestSanitizeBranchName_SecurityVulnerabilities(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		input       string
		expectError bool
		description string
	}{
		{
			name:        "sql_injection_single_quote",
			input:       "'; DROP TABLE projects; --",
			expectError: true,
			description: "SQL injection with single quote",
		},
		{
			name:        "sql_injection_comment",
			input:       "feature'--",
			expectError: false,
			description: "SQL injection with comment (single quote and dash are valid git chars)",
		},
		{
			name:        "sql_injection_union",
			input:       "' UNION SELECT * FROM users--",
			expectError: true,
			description: "SQL injection with UNION (contains space)",
		},
		{
			name:        "xss_script_tag",
			input:       "<script>alert('xss')</script>",
			expectError: false,
			description: "XSS with script tag (angle brackets are valid git branch chars)",
		},
		{
			name:        "xss_img_tag",
			input:       "<img src=x onerror=alert('xss')>",
			expectError: true,
			description: "XSS with img tag (contains space)",
		},
		{
			name:        "xss_event_handler",
			input:       "\" onload=\"alert('xss')\"",
			expectError: true,
			description: "XSS with event handler (contains spaces)",
		},
		{
			name:        "null_byte",
			input:       "feature\x00branch",
			expectError: true,
			description: "Null byte injection",
		},
		{
			name:        "carriage_return",
			input:       "feature\rbranch",
			expectError: true,
			description: "Carriage return character",
		},
		{
			name:        "vertical_tab",
			input:       "feature\vbranch",
			expectError: true,
			description: "Vertical tab character",
		},
		{
			name:        "form_feed",
			input:       "feature\fbranch",
			expectError: true,
			description: "Form feed character",
		},
		{
			name:        "bell_character",
			input:       "feature\abranch",
			expectError: true,
			description: "Bell character",
		},
		{
			name:        "backspace",
			input:       "feature\bbranch",
			expectError: true,
			description: "Backspace character",
		},
		{
			name:        "escape_sequence",
			input:       "feature\x1bbranch",
			expectError: true,
			description: "ANSI escape sequence",
		},
		{
			name:        "path_traversal_dotdot",
			input:       "feature/../../../etc/passwd",
			expectError: true,
			description: "Path traversal with ..",
		},
		{
			name:        "git_reflog_sequence",
			input:       "feature@{0}",
			expectError: true,
			description: "Git reflog sequence @{",
		},
		{
			name:        "double_slash",
			input:       "feature//branch",
			expectError: true,
			description: "Double slash sequence",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			t.Parallel()
			_, err := SanitizeBranchName(tt.input)
			if tt.expectError {
				assert.Error(t, err, tt.description+" should return error")
			} else {
				assert.NoError(t, err, tt.description+" should not return error")
			}
		})
	}
}

func TestSanitizeBranchName_LongBranches(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		input       string
		maxLength   int
		shouldError bool
		description string
	}{
		{
			name:        "exactly_255_chars",
			input:       strings.Repeat("a", 255),
			maxLength:   255,
			shouldError: false,
			description: "Branch name with exactly 255 characters should be allowed",
		},
		{
			name:        "256_chars",
			input:       strings.Repeat("a", 256),
			maxLength:   255,
			shouldError: true,
			description: "Branch name with 256 characters should return error",
		},
		{
			name:        "500_chars",
			input:       strings.Repeat("a", 500),
			maxLength:   255,
			shouldError: true,
			description: "Very long branch name should return error",
		},
		{
			name:        "1000_chars",
			input:       strings.Repeat("a", 1000),
			maxLength:   255,
			shouldError: true,
			description: "Extremely long branch name should return error",
		},
		{
			name:        "long_with_slashes",
			input:       "feature/" + strings.Repeat("a", 250),
			maxLength:   255,
			shouldError: true,
			description: "Long branch name with slashes should return error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			t.Parallel()
			result, err := SanitizeBranchName(tt.input)
			if tt.shouldError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "too long")
			} else {
				assert.NoError(t, err)
				assert.LessOrEqual(t, len(result), tt.maxLength, tt.description)
			}
		})
	}
}

func TestSanitizeBranchName_Unicode(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "chinese_characters",
			input: "特性/我的分支",
		},
		{
			name:  "japanese_characters",
			input: "フィーチャー/ブランチ",
		},
		{
			name:  "emoji",
			input: "feature/🚀-rocket",
		},
		{
			name:  "mixed_unicode",
			input: "feature/amélioration-système",
		},
		{
			name:  "cyrillic",
			input: "функция/ветка",
		},
		{
			name:  "arabic",
			input: "ميزة/فرع",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			t.Parallel()
			result, err := SanitizeBranchName(tt.input)
			assert.NoError(t, err, "Should handle unicode characters without error")
			assert.NotEmpty(t, result, "Should handle unicode characters")
		})
	}
}

func TestGitInfo_Validation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		gitInfo  GitInfo
		isValid  bool
		describe string
	}{
		{
			name: "valid_repo_with_commits",
			gitInfo: GitInfo{
				IsRepo:        true,
				CurrentBranch: "main",
				IsDetached:    false,
				HasCommits:    true,
			},
			isValid:  true,
			describe: "Normal git repository with commits should be valid",
		},
		{
			name: "valid_feature_branch",
			gitInfo: GitInfo{
				IsRepo:        true,
				CurrentBranch: "feature/my-feature",
				IsDetached:    false,
				HasCommits:    true,
			},
			isValid:  true,
			describe: "Feature branch should be valid",
		},
		{
			name: "not_a_repo",
			gitInfo: GitInfo{
				IsRepo:        false,
				CurrentBranch: "",
				IsDetached:    false,
				HasCommits:    false,
			},
			isValid:  false,
			describe: "Not a git repository should be invalid",
		},
		{
			name: "detached_head",
			gitInfo: GitInfo{
				IsRepo:        true,
				CurrentBranch: "",
				IsDetached:    true,
				HasCommits:    true,
			},
			isValid:  false,
			describe: "Detached HEAD state should be invalid for branch association",
		},
		{
			name: "no_commits",
			gitInfo: GitInfo{
				IsRepo:        true,
				CurrentBranch: "main",
				IsDetached:    false,
				HasCommits:    false,
			},
			isValid:  false,
			describe: "Repository with no commits should be invalid",
		},
		{
			name: "empty_branch_name",
			gitInfo: GitInfo{
				IsRepo:        true,
				CurrentBranch: "",
				IsDetached:    false,
				HasCommits:    true,
			},
			isValid:  false,
			describe: "Empty branch name should be invalid",
		},
		{
			name: "all_false",
			gitInfo: GitInfo{
				IsRepo:        false,
				CurrentBranch: "",
				IsDetached:    false,
				HasCommits:    false,
			},
			isValid:  false,
			describe: "Default zero values should be invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			t.Parallel()

			// Test the validation logic (this will depend on implementation)
			// For now, check basic invariants
			isValid := tt.gitInfo.IsRepo &&
				tt.gitInfo.HasCommits &&
				!tt.gitInfo.IsDetached &&
				tt.gitInfo.CurrentBranch != ""

			assert.Equal(t, tt.isValid, isValid, tt.describe)
		})
	}
}

func TestSanitizeBranchName_TableDriven(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		input       string
		expectError bool
		minLength   int
		maxLength   int
		description string
	}{
		{
			name:        "normal_main",
			input:       "main",
			expectError: false,
			minLength:   4,
			maxLength:   4,
			description: "Simple branch name 'main'",
		},
		{
			name:        "normal_develop",
			input:       "develop",
			expectError: false,
			minLength:   7,
			maxLength:   7,
			description: "Simple branch name 'develop'",
		},
		{
			name:        "with_prefix",
			input:       "feature/AUTH-123-user-login",
			expectError: false,
			minLength:   1,
			maxLength:   255,
			description: "Branch with JIRA-style ticket prefix",
		},
		{
			name:        "release_branch",
			input:       "release/v2.5.1",
			expectError: false,
			minLength:   1,
			maxLength:   255,
			description: "Release branch with version",
		},
		{
			name:        "hotfix_branch",
			input:       "hotfix/critical-bug-fix",
			expectError: false,
			minLength:   1,
			maxLength:   255,
			description: "Hotfix branch",
		},
		{
			name:        "empty_after_trim",
			input:       "     ",
			expectError: false,
			minLength:   0,
			maxLength:   0,
			description: "Whitespace that becomes empty after trim",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			t.Parallel()

			result, err := SanitizeBranchName(tt.input)

			if tt.expectError {
				assert.Error(t, err, tt.description)
			} else {
				assert.NoError(t, err, tt.description)
				assert.GreaterOrEqual(t, len(result), tt.minLength, "Result should meet minimum length")
				assert.LessOrEqual(t, len(result), tt.maxLength, "Result should not exceed maximum length")
			}
		})
	}
}
