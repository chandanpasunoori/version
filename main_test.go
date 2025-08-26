package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"sort"
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

// Helper function to create a temporary git repository with test tags for benchmarks
func createTestRepoForBench(b *testing.B, tags []string) (string, func()) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "version-test-*")
	if err != nil {
		b.Fatalf("Failed to create temp directory: %v", err)
	}

	// Change to temp directory and set up git
	originalDir, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get current directory: %v", err)
	}
	
	err = os.Chdir(tempDir)
	if err != nil {
		b.Fatalf("Failed to change directory: %v", err)
	}

	// Initialize git repository with git commands
	if err := exec.Command("git", "init").Run(); err != nil {
		os.Chdir(originalDir)
		b.Fatalf("Failed to git init: %v", err)
	}
	
	if err := exec.Command("git", "config", "user.name", "Test User").Run(); err != nil {
		os.Chdir(originalDir)
		b.Fatalf("Failed to set git user.name: %v", err)
	}
	
	if err := exec.Command("git", "config", "user.email", "test@example.com").Run(); err != nil {
		os.Chdir(originalDir)
		b.Fatalf("Failed to set git user.email: %v", err)
	}

	// Create a test file
	testFile := filepath.Join(tempDir, "test.txt")
	err = os.WriteFile(testFile, []byte("test content"), 0644)
	if err != nil {
		os.Chdir(originalDir)
		b.Fatalf("Failed to create test file: %v", err)
	}

	// Add and commit the test file using git commands
	if err := exec.Command("git", "add", "test.txt").Run(); err != nil {
		os.Chdir(originalDir)
		b.Fatalf("Failed to git add: %v", err)
	}
	
	if err := exec.Command("git", "commit", "-m", "Initial commit").Run(); err != nil {
		os.Chdir(originalDir)
		b.Fatalf("Failed to git commit: %v", err)
	}

	// Create test tags using git commands
	for _, tagName := range tags {
		if err := exec.Command("git", "tag", tagName).Run(); err != nil {
			os.Chdir(originalDir)
			b.Fatalf("Failed to create tag %s: %v", tagName, err)
		}
	}

	// Return to original directory
	os.Chdir(originalDir)

	cleanup := func() {
		os.RemoveAll(tempDir)
	}

	return tempDir, cleanup
}
func createTestRepo(t *testing.T, tags []string) (string, func()) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "version-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}

	// Change to temp directory and set up git
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	
	err = os.Chdir(tempDir)
	if err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	// Initialize git repository with git commands - more reliable for testing
	if err := exec.Command("git", "init").Run(); err != nil {
		os.Chdir(originalDir)
		t.Fatalf("Failed to git init: %v", err)
	}
	
	if err := exec.Command("git", "config", "user.name", "Test User").Run(); err != nil {
		os.Chdir(originalDir)
		t.Fatalf("Failed to set git user.name: %v", err)
	}
	
	if err := exec.Command("git", "config", "user.email", "test@example.com").Run(); err != nil {
		os.Chdir(originalDir)
		t.Fatalf("Failed to set git user.email: %v", err)
	}

	// Create a test file
	testFile := filepath.Join(tempDir, "test.txt")
	err = os.WriteFile(testFile, []byte("test content"), 0644)
	if err != nil {
		os.Chdir(originalDir)
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Add and commit the test file using git commands
	if err := exec.Command("git", "add", "test.txt").Run(); err != nil {
		os.Chdir(originalDir)
		t.Fatalf("Failed to git add: %v", err)
	}
	
	if err := exec.Command("git", "commit", "-m", "Initial commit").Run(); err != nil {
		os.Chdir(originalDir)
		t.Fatalf("Failed to git commit: %v", err)
	}

	// Create test tags using git commands
	for _, tagName := range tags {
		if err := exec.Command("git", "tag", tagName).Run(); err != nil {
			os.Chdir(originalDir)
			t.Fatalf("Failed to create tag %s: %v", tagName, err)
		}
	}

	// Return to original directory
	os.Chdir(originalDir)

	cleanup := func() {
		os.RemoveAll(tempDir)
	}

	return tempDir, cleanup
}

func TestVersion_SemVerList_Sorting(t *testing.T) {
	tests := []struct {
		name     string
		input    SemVerList
		expected SemVerList
	}{
		{
			name: "sort ascending order",
			input: SemVerList{
				{Major: 1, Minor: 2, Patch: 3},
				{Major: 1, Minor: 1, Patch: 0},
				{Major: 2, Minor: 0, Patch: 0},
				{Major: 1, Minor: 2, Patch: 2},
			},
			expected: SemVerList{
				{Major: 1, Minor: 1, Patch: 0},
				{Major: 1, Minor: 2, Patch: 2},
				{Major: 1, Minor: 2, Patch: 3},
				{Major: 2, Minor: 0, Patch: 0},
			},
		},
		{
			name: "already sorted",
			input: SemVerList{
				{Major: 1, Minor: 0, Patch: 0},
				{Major: 1, Minor: 1, Patch: 0},
				{Major: 2, Minor: 0, Patch: 0},
			},
			expected: SemVerList{
				{Major: 1, Minor: 0, Patch: 0},
				{Major: 1, Minor: 1, Patch: 0},
				{Major: 2, Minor: 0, Patch: 0},
			},
		},
		{
			name: "reverse order",
			input: SemVerList{
				{Major: 3, Minor: 0, Patch: 0},
				{Major: 2, Minor: 0, Patch: 0},
				{Major: 1, Minor: 0, Patch: 0},
			},
			expected: SemVerList{
				{Major: 1, Minor: 0, Patch: 0},
				{Major: 2, Minor: 0, Patch: 0},
				{Major: 3, Minor: 0, Patch: 0},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sort.Sort(tt.input)
			if !reflect.DeepEqual(tt.input, tt.expected) {
				t.Errorf("Sort() = %v, expected %v", tt.input, tt.expected)
			}
		})
	}
}

func TestGetCurrentModules(t *testing.T) {
	tests := []struct {
		name            string
		tags            []string
		expectedModules []string
		expectedReleases []string
		expectError     bool
	}{
		{
			name: "valid tags with multiple modules and releases",
			tags: []string{
				"myapp/dev/v1.0.0",
				"myapp/prod/v2.1.5",
				"backend/dev/v0.1.0",
				"frontend/prod/v1.5.2",
			},
			expectedModules: []string{"backend", "frontend", "myapp"}, // sorted
			expectedReleases: []string{"dev", "prod"}, // sorted
			expectError:     false,
		},
		{
			name: "single module single release",
			tags: []string{
				"api/staging/v1.0.0",
			},
			expectedModules: []string{"api"},
			expectedReleases: []string{"staging"},
			expectError:     false,
		},
		{
			name: "invalid tag formats mixed with valid ones",
			tags: []string{
				"myapp/dev/v1.0.0", // valid
				"invalid-tag",      // invalid
				"also/invalid",     // invalid
				"backend/prod/v2.0.0", // valid
			},
			expectedModules: []string{"backend", "myapp"},
			expectedReleases: []string{"dev", "prod"},
			expectError:     false,
		},
		{
			name:            "no valid tags",
			tags:            []string{"invalid-tag", "another-invalid"},
			expectedModules: []string{},
			expectedReleases: []string{},
			expectError:     false,
		},
		{
			name:            "empty repository",
			tags:            []string{},
			expectedModules: []string{},
			expectedReleases: []string{},
			expectError:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir, cleanup := createTestRepo(t, tt.tags)
			defer cleanup()

			// Change to the test repository directory
			originalDir, err := os.Getwd()
			if err != nil {
				t.Fatalf("Failed to get current directory: %v", err)
			}
			defer os.Chdir(originalDir)

			err = os.Chdir(tempDir)
			if err != nil {
				t.Fatalf("Failed to change directory: %v", err)
			}

			modules, releases, err := getCurrentModules()

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			// Sort the results for consistent comparison
			sort.Strings(modules)
			sort.Strings(releases)
			sort.Strings(tt.expectedModules)
			sort.Strings(tt.expectedReleases)

			// Handle nil vs empty slice comparison
			if len(modules) == 0 && len(tt.expectedModules) == 0 {
				// Both empty, they match
			} else if !reflect.DeepEqual(modules, tt.expectedModules) {
				t.Errorf("modules = %v, expected %v", modules, tt.expectedModules)
			}

			if len(releases) == 0 && len(tt.expectedReleases) == 0 {
				// Both empty, they match
			} else if !reflect.DeepEqual(releases, tt.expectedReleases) {
				t.Errorf("releases = %v, expected %v", releases, tt.expectedReleases)
			}
		})
	}
}

func TestParseCurrentVersion(t *testing.T) {
	tests := []struct {
		name            string
		tags            []string
		moduleName      string
		releaseChannels []string
		expectedVersion Version
		expectError     bool
	}{
		{
			name: "single module single release with multiple versions",
			tags: []string{
				"myapp/dev/v1.0.0",
				"myapp/dev/v1.2.3",
				"myapp/dev/v2.1.0",
				"myapp/dev/v1.5.2", // should return the highest version (2.1.0)
			},
			moduleName:      "myapp",
			releaseChannels: []string{"dev"},
			expectedVersion: Version{Major: 2, Minor: 1, Patch: 0},
			expectError:     false,
		},
		{
			name: "multiple release channels",
			tags: []string{
				"api/dev/v1.0.0",
				"api/staging/v2.0.0",
				"api/prod/v1.5.0",
			},
			moduleName:      "api",
			releaseChannels: []string{"dev", "staging", "prod"},
			expectedVersion: Version{Major: 2, Minor: 0, Patch: 0}, // highest across all channels
			expectError:     false,
		},
		{
			name: "module exists but different release channel",
			tags: []string{
				"myapp/dev/v1.0.0",
				"myapp/staging/v2.0.0",
			},
			moduleName:      "myapp",
			releaseChannels: []string{"prod"}, // prod doesn't exist
			expectedVersion: Version{Major: 0, Minor: 0, Patch: 0}, // default
			expectError:     false,
		},
		{
			name: "non-existent module",
			tags: []string{
				"myapp/dev/v1.0.0",
			},
			moduleName:      "nonexistent",
			releaseChannels: []string{"dev"},
			expectedVersion: Version{Major: 0, Minor: 0, Patch: 0},
			expectError:     false,
		},
		{
			name:            "empty repository",
			tags:            []string{},
			moduleName:      "anymodule",
			releaseChannels: []string{"dev"},
			expectedVersion: Version{Major: 0, Minor: 0, Patch: 0},
			expectError:     false,
		},
		{
			name: "mixed valid and invalid tags",
			tags: []string{
				"myapp/dev/v1.0.0",
				"invalid-tag",
				"myapp/dev/v2.1.5",
				"another-invalid",
			},
			moduleName:      "myapp",
			releaseChannels: []string{"dev"},
			expectedVersion: Version{Major: 2, Minor: 1, Patch: 5},
			expectError:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir, cleanup := createTestRepo(t, tt.tags)
			defer cleanup()

			// Change to the test repository directory
			originalDir, err := os.Getwd()
			if err != nil {
				t.Fatalf("Failed to get current directory: %v", err)
			}
			defer os.Chdir(originalDir)

			err = os.Chdir(tempDir)
			if err != nil {
				t.Fatalf("Failed to change directory: %v", err)
			}

			version, err := parseCurrentVersion(tt.moduleName, tt.releaseChannels)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if !reflect.DeepEqual(version, tt.expectedVersion) {
				t.Errorf("parseCurrentVersion() = %v, expected %v", version, tt.expectedVersion)
			}
		})
	}
}

func TestGenerateNextVersion(t *testing.T) {
	tests := []struct {
		name           string
		moduleName     string
		releaseChannel string
		currentVersion Version
		expected       string
	}{
		{
			name:           "simple patch increment",
			moduleName:     "myapp",
			releaseChannel: "dev",
			currentVersion: Version{Major: 1, Minor: 2, Patch: 3},
			expected:       "myapp/dev/v1.2.4",
		},
		{
			name:           "patch overflow to minor",
			moduleName:     "api",
			releaseChannel: "prod",
			currentVersion: Version{Major: 1, Minor: 5, Patch: 9},
			expected:       "api/prod/v1.6.0",
		},
		{
			name:           "minor overflow to major",
			moduleName:     "backend",
			releaseChannel: "staging",
			currentVersion: Version{Major: 2, Minor: 9, Patch: 9},
			expected:       "backend/staging/v3.0.0",
		},
		{
			name:           "from zero version",
			moduleName:     "newapp",
			releaseChannel: "dev",
			currentVersion: Version{Major: 0, Minor: 0, Patch: 0},
			expected:       "newapp/dev/v0.0.1",
		},
		{
			name:           "large version numbers",
			moduleName:     "enterprise",
			releaseChannel: "release",
			currentVersion: Version{Major: 15, Minor: 7, Patch: 8},
			expected:       "enterprise/release/v15.7.9",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generateNextVersion(tt.moduleName, tt.releaseChannel, tt.currentVersion)
			if result != tt.expected {
				t.Errorf("generateNextVersion() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestCreateGitTag(t *testing.T) {
	tests := []struct {
		name        string
		tagName     string
		expectError bool
	}{
		{
			name:        "valid tag creation",
			tagName:     "myapp/dev/v1.0.0",
			expectError: false,
		},
		{
			name:        "another valid tag",
			tagName:     "backend/prod/v2.1.5",
			expectError: false,
		},
		{
			name:        "tag with special characters",
			tagName:     "my-app_2/dev-test/v1.0.0",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir, cleanup := createTestRepo(t, []string{}) // Start with empty repo
			defer cleanup()

			// Change to the test repository directory
			originalDir, err := os.Getwd()
			if err != nil {
				t.Fatalf("Failed to get current directory: %v", err)
			}
			defer os.Chdir(originalDir)

			err = os.Chdir(tempDir)
			if err != nil {
				t.Fatalf("Failed to change directory: %v", err)
			}

			err = createGitTag(tt.tagName)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			// Verify the tag was actually created
			repo, err := git.PlainOpen(".")
			if err != nil {
				t.Fatalf("Failed to open repository for verification: %v", err)
			}

			tags, err := repo.Tags()
			if err != nil {
				t.Fatalf("Failed to get tags for verification: %v", err)
			}

			tagFound := false
			err = tags.ForEach(func(ref *plumbing.Reference) error {
				if ref.Name().Short() == tt.tagName {
					tagFound = true
				}
				return nil
			})
			if err != nil {
				t.Fatalf("Error iterating tags: %v", err)
			}

			if !tagFound {
				t.Errorf("Tag %s was not found after creation", tt.tagName)
			}
		})
	}
}

func TestCreateGitTag_ErrorConditions(t *testing.T) {
	t.Run("duplicate tag creation", func(t *testing.T) {
		tempDir, cleanup := createTestRepo(t, []string{"existing/tag/v1.0.0"})
		defer cleanup()

		// Change to the test repository directory
		originalDir, err := os.Getwd()
		if err != nil {
			t.Fatalf("Failed to get current directory: %v", err)
		}
		defer os.Chdir(originalDir)

		err = os.Chdir(tempDir)
		if err != nil {
			t.Fatalf("Failed to change directory: %v", err)
		}

		// Try to create the same tag again - should fail
		err = createGitTag("existing/tag/v1.0.0")
		if err == nil {
			t.Errorf("Expected error when creating duplicate tag, but got none")
		}
	})

	t.Run("non-git repository", func(t *testing.T) {
		// Create temporary directory without git
		tempDir, err := os.MkdirTemp("", "non-git-*")
		if err != nil {
			t.Fatalf("Failed to create temp directory: %v", err)
		}
		defer os.RemoveAll(tempDir)

		originalDir, err := os.Getwd()
		if err != nil {
			t.Fatalf("Failed to get current directory: %v", err)
		}
		defer os.Chdir(originalDir)

		err = os.Chdir(tempDir)
		if err != nil {
			t.Fatalf("Failed to change directory: %v", err)
		}

		err = createGitTag("test/tag/v1.0.0")
		if err == nil {
			t.Errorf("Expected error when creating tag in non-git directory, but got none")
		}
	})
}

// Test the integration of getCurrentModules and parseCurrentVersion
func TestIntegration_ModulesAndVersions(t *testing.T) {
	tags := []string{
		"frontend/dev/v1.0.0",
		"frontend/dev/v1.2.0",
		"frontend/prod/v1.1.0",
		"backend/dev/v2.0.0",
		"backend/dev/v2.1.5",
		"backend/staging/v2.1.0",
		"api/prod/v0.5.0",
	}

	tempDir, cleanup := createTestRepo(t, tags)
	defer cleanup()

	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(originalDir)

	err = os.Chdir(tempDir)
	if err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	// Test getCurrentModules
	modules, releases, err := getCurrentModules()
	if err != nil {
		t.Fatalf("getCurrentModules() error: %v", err)
	}

	expectedModules := []string{"api", "backend", "frontend"}
	expectedReleases := []string{"dev", "prod", "staging"}
	sort.Strings(modules)
	sort.Strings(releases)

	if !reflect.DeepEqual(modules, expectedModules) {
		t.Errorf("modules = %v, expected %v", modules, expectedModules)
	}
	if !reflect.DeepEqual(releases, expectedReleases) {
		t.Errorf("releases = %v, expected %v", releases, expectedReleases)
	}

	// Test parseCurrentVersion for each module
	testCases := []struct {
		module   string
		channels []string
		expected Version
	}{
		{
			module:   "frontend",
			channels: []string{"dev"},
			expected: Version{Major: 1, Minor: 2, Patch: 0}, // highest dev version
		},
		{
			module:   "backend",
			channels: []string{"dev", "staging"},
			expected: Version{Major: 2, Minor: 1, Patch: 5}, // highest across both channels
		},
		{
			module:   "api",
			channels: []string{"prod"},
			expected: Version{Major: 0, Minor: 5, Patch: 0},
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("parse_%s", tc.module), func(t *testing.T) {
			version, err := parseCurrentVersion(tc.module, tc.channels)
			if err != nil {
				t.Errorf("parseCurrentVersion() error: %v", err)
			}
			if !reflect.DeepEqual(version, tc.expected) {
				t.Errorf("parseCurrentVersion(%s, %v) = %v, expected %v",
					tc.module, tc.channels, version, tc.expected)
			}
		})
	}
}

// Benchmark tests for performance validation
func BenchmarkGetCurrentModules(b *testing.B) {
	// Create a repository with many tags
	tags := make([]string, 1000)
	for i := 0; i < 1000; i++ {
		tags[i] = fmt.Sprintf("module%d/dev/v1.%d.%d", i%10, i/10, i%10)
	}

	tempDir, cleanup := createTestRepoForBench(b, tags)
	defer cleanup()

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tempDir)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		getCurrentModules()
	}
}

func BenchmarkParseCurrentVersion(b *testing.B) {
	tags := make([]string, 100)
	for i := 0; i < 100; i++ {
		tags[i] = fmt.Sprintf("testmodule/dev/v%d.%d.%d", i/25, (i%25)/5, i%5)
	}

	tempDir, cleanup := createTestRepoForBench(b, tags)
	defer cleanup()

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tempDir)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parseCurrentVersion("testmodule", []string{"dev"})
	}
}