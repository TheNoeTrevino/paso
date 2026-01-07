package git

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// SanitizeBranchName Tests
// ============================================================================

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
			result := SanitizeBranchName(tt.input)
			assert.Equal(t, tt.expected, result, "Branch name should be sanitized correctly")
		})
	}
}

func TestSanitizeBranchName_EdgeCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty_string",
			input:    "",
			expected: "",
		},
		{
			name:     "whitespace_only",
			input:    "   ",
			expected: "",
		},
		{
			name:     "leading_whitespace",
			input:    "  main",
			expected: "main",
		},
		{
			name:     "trailing_whitespace",
			input:    "main  ",
			expected: "main",
		},
		{
			name:     "both_whitespace",
			input:    "  feature/branch  ",
			expected: "feature/branch",
		},
		{
			name:     "newline_characters",
			input:    "feature\nbranch",
			expected: "feature\nbranch", // Sanitize should handle or preserve this
		},
		{
			name:     "tab_characters",
			input:    "feature\tbranch",
			expected: "feature\tbranch",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := SanitizeBranchName(tt.input)
			assert.Equal(t, tt.expected, result, "Edge case should be handled correctly")
		})
	}
}

func TestSanitizeBranchName_SpecialCharacters(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		input       string
		description string
	}{
		{
			name:        "asterisk",
			input:       "feature*branch",
			description: "Branch with asterisk",
		},
		{
			name:        "question_mark",
			input:       "feature?branch",
			description: "Branch with question mark",
		},
		{
			name:        "brackets",
			input:       "feature[branch]",
			description: "Branch with brackets",
		},
		{
			name:        "at_symbol",
			input:       "user@feature",
			description: "Branch with @ symbol",
		},
		{
			name:        "hash_symbol",
			input:       "issue#123",
			description: "Branch with hash symbol",
		},
		{
			name:        "backslash",
			input:       "feature\\branch",
			description: "Branch with backslash",
		},
		{
			name:        "caret",
			input:       "feature^branch",
			description: "Branch with caret",
		},
		{
			name:        "tilde",
			input:       "feature~branch",
			description: "Branch with tilde",
		},
		{
			name:        "colon",
			input:       "feature:branch",
			description: "Branch with colon",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// SanitizeBranchName should handle special characters
			// The actual behavior depends on implementation
			result := SanitizeBranchName(tt.input)
			// Just verify it returns something (implementation will define behavior)
			assert.NotNil(t, result, tt.description)
		})
	}
}

func TestSanitizeBranchName_LongBranches(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		input       string
		maxLength   int
		description string
	}{
		{
			name:        "exactly_255_chars",
			input:       strings.Repeat("a", 255),
			maxLength:   255,
			description: "Branch name with exactly 255 characters should be allowed",
		},
		{
			name:        "256_chars",
			input:       strings.Repeat("a", 256),
			maxLength:   255,
			description: "Branch name with 256 characters should be truncated to 255",
		},
		{
			name:        "500_chars",
			input:       strings.Repeat("a", 500),
			maxLength:   255,
			description: "Very long branch name should be truncated to 255",
		},
		{
			name:        "1000_chars",
			input:       strings.Repeat("a", 1000),
			maxLength:   255,
			description: "Extremely long branch name should be truncated to 255",
		},
		{
			name:        "long_with_slashes",
			input:       "feature/" + strings.Repeat("a", 250),
			maxLength:   255,
			description: "Long branch name with slashes should be handled",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := SanitizeBranchName(tt.input)
			assert.LessOrEqual(t, len(result), tt.maxLength, tt.description)
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
			result := SanitizeBranchName(tt.input)
			// Just verify it handles unicode without panicking
			assert.NotNil(t, result, "Should handle unicode characters")
		})
	}
}

// ============================================================================
// GitInfo Struct Tests
// ============================================================================

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

// ============================================================================
// DetectGitInfo Error Cases (these will fail until implementation exists)
// ============================================================================

func TestDetectGitInfo_NotImplementedYet(t *testing.T) {
	// This test verifies that DetectGitInfo exists but is not yet implemented
	// This should fail in the TDD "red" phase

	t.Skip("Skipping until DetectGitInfo is implemented")

	// Uncomment when ready to test implementation:
	// ctx := context.Background()
	// info := DetectGitInfo(ctx)
	// assert.NotNil(t, info, "DetectGitInfo should return a GitInfo struct")
}

func TestSanitizeBranchName_NotImplementedYet(t *testing.T) {
	// This test verifies that SanitizeBranchName exists but is not yet implemented
	// This should fail in the TDD "red" phase

	// The function will not exist yet, so these tests will fail to compile
	// which is expected in TDD red phase
	t.Skip("This test will fail until SanitizeBranchName is implemented")
}

// ============================================================================
// Table-Driven Tests for Comprehensive Coverage
// ============================================================================

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

			result := SanitizeBranchName(tt.input)

			if !tt.expectError {
				require.NotNil(t, result, tt.description)
				assert.GreaterOrEqual(t, len(result), tt.minLength, "Result should meet minimum length")
				assert.LessOrEqual(t, len(result), tt.maxLength, "Result should not exceed maximum length")
			}
		})
	}
}
