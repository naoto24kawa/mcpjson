package path

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/naoto24kawa/mcpconfig/internal/config"
	"github.com/naoto24kawa/mcpconfig/internal/profile"
	"github.com/naoto24kawa/mcpconfig/internal/testutil"
)

func setupTestEnvironment(t *testing.T) (string, *config.Config, func()) {
	t.Helper()
	return testutil.SetupIsolatedTestEnvironment(t)
}

func createTestProfile(t *testing.T, cfg *config.Config, profileName string) string {
	t.Helper()

	// Generate unique profile name if needed
	uniqueName := profileName
	if profileName != config.DefaultProfileName {
		uniqueName = testutil.GenerateUniqueProfileName(profileName)
	}

	profileManager := profile.NewManager(cfg.ProfilesDir)
	err := profileManager.Create(uniqueName, "Test profile for path testing")
	if err != nil {
		t.Fatalf("Failed to create test profile: %v", err)
	}
	
	return uniqueName
}

func TestPathCmd_DefaultProfile(t *testing.T) {
	// Arrange
	_, cfg, cleanup := setupTestEnvironment(t)
	defer cleanup()

	_ = createTestProfile(t, cfg, config.DefaultProfileName)

	// Act
	var output bytes.Buffer
	PathCmd.SetOut(&output)
	PathCmd.SetErr(&output) // Also set error output to capture errors
	PathCmd.SetArgs([]string{})

	err := PathCmd.Execute()

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	result := output.String()
	expectedPath := filepath.Join(cfg.ProfilesDir, config.DefaultProfileName+config.FileExtension)

	if result != expectedPath {
		t.Errorf("Expected path '%s', got '%s'", expectedPath, result)
	}

	// Verify the path exists
	if _, err := os.Stat(result); os.IsNotExist(err) {
		t.Errorf("Expected profile file to exist at '%s'", result)
	}
}

func TestPathCmd_SpecificProfile(t *testing.T) {
	tests := []struct {
		name        string
		profileName string
		createFile  bool
		expectError bool
		errorMsg    string
	}{
		{
			name:        "existing profile",
			profileName: "test-profile",
			createFile:  true,
			expectError: false,
		},
		{
			name:        "non-existent profile",
			profileName: "non-existent",
			createFile:  false,
			expectError: true,
			errorMsg:    "プロファイルパスの取得に失敗しました",
		},
		{
			name:        "profile with special characters",
			profileName: "test-profile-with-dashes",
			createFile:  true,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			_, cfg, cleanup := setupTestEnvironment(t)
			defer cleanup()

			var actualProfileName string
			if tt.createFile {
				actualProfileName = createTestProfile(t, cfg, tt.profileName)
			} else {
				actualProfileName = tt.profileName
			}

			// Act
			var output bytes.Buffer
			var errOutput bytes.Buffer
			PathCmd.SetOut(&output)
			PathCmd.SetErr(&errOutput)
			PathCmd.SetArgs([]string{actualProfileName})

			err := PathCmd.Execute()

			// Assert
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				if tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error to contain '%s', got: %s", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}

				result := output.String()
				expectedPath := filepath.Join(cfg.ProfilesDir, actualProfileName+config.FileExtension)

				if result != expectedPath {
					t.Errorf("Expected path '%s', got '%s'", expectedPath, result)
				}

				// Verify the path exists when we created the file
				if tt.createFile {
					if _, err := os.Stat(result); os.IsNotExist(err) {
						t.Errorf("Expected profile file to exist at '%s'", result)
					}
				}
			}
		})
	}
}

func TestPathCmd_ArgumentValidation(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "no arguments",
			args:        []string{},
			expectError: false,
		},
		{
			name:        "one argument",
			args:        []string{"test-profile"},
			expectError: false,
		},
		{
			name:        "too many arguments",
			args:        []string{"profile1", "profile2"},
			expectError: true,
			errorMsg:    "accepts at most 1 arg",
		},
		{
			name:        "empty string argument",
			args:        []string{""},
			expectError: true, // Empty profile name should cause validation error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			_, cfg, cleanup := setupTestEnvironment(t)
			defer cleanup()

			// Create default profile for cases that might succeed
			var actualArgs []string
			if !tt.expectError {
				profileName := config.DefaultProfileName
				if len(tt.args) > 0 && tt.args[0] != "" {
					actualProfileName := createTestProfile(t, cfg, tt.args[0])
					actualArgs = []string{actualProfileName}
				} else {
					_ = createTestProfile(t, cfg, profileName)
					actualArgs = tt.args
				}
			} else {
				actualArgs = tt.args
			}

			// Act
			var output bytes.Buffer
			var errOutput bytes.Buffer
			PathCmd.SetOut(&output)
			PathCmd.SetErr(&errOutput)
			PathCmd.SetArgs(actualArgs)

			err := PathCmd.Execute()

			// Assert
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				if tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error to contain '%s', got: %s", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
			}
		})
	}
}

func TestPathCmd_OutputFormat(t *testing.T) {
	// Arrange
	_, cfg, cleanup := setupTestEnvironment(t)
	defer cleanup()

	profileName := createTestProfile(t, cfg, "format-test-profile")

	// Act
	var output bytes.Buffer
	PathCmd.SetOut(&output)
	PathCmd.SetArgs([]string{profileName})

	err := PathCmd.Execute()

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	result := output.String()

	// Verify output format
	if strings.Contains(result, "\n") {
		t.Error("Expected output without newlines")
	}

	if !strings.HasSuffix(result, config.FileExtension) {
		t.Errorf("Expected output to end with '%s', got: %s", config.FileExtension, result)
	}

	if !filepath.IsAbs(result) {
		t.Errorf("Expected absolute path, got: %s", result)
	}
}

func TestPathCmd_ConfigError(t *testing.T) {
	// Arrange - Set invalid HOME to cause config error
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", "/invalid/nonexistent/path")
	defer func() { os.Setenv("HOME", oldHome) }()

	// Act
	var output bytes.Buffer
	var errOutput bytes.Buffer
	PathCmd.SetOut(&output)
	PathCmd.SetErr(&errOutput)
	PathCmd.SetArgs([]string{"test-profile"})

	err := PathCmd.Execute()

	// Assert
	if err == nil {
		t.Error("Expected error due to invalid config, but got none")
	} else {
		if !strings.Contains(err.Error(), "設定の読み込みに失敗しました") {
			t.Errorf("Expected config error message, got: %s", err.Error())
		}
	}
}

func TestPathCmd_ValidationEdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		profileName string
		expectError bool
		description string
	}{
		{
			name:        "profile name with spaces",
			profileName: "profile with spaces",
			expectError: true,
			description: "profile names with spaces should be invalid",
		},
		{
			name:        "very long profile name",
			profileName: strings.Repeat("a", 100),
			expectError: true,
			description: "very long profile names should be invalid",
		},
		{
			name:        "profile name with special chars",
			profileName: "profile@#$%",
			expectError: true,
			description: "profile names with special characters should be invalid",
		},
		{
			name:        "profile name starting with dash",
			profileName: "-profile",
			expectError: true,
			description: "profile names starting with dash should be invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			_, cfg, cleanup := setupTestEnvironment(t)
			defer cleanup()

			// Don't create profile for invalid names
			var actualProfileName string
			if !tt.expectError {
				actualProfileName = createTestProfile(t, cfg, tt.profileName)
			} else {
				actualProfileName = tt.profileName
			}

			// Act
			var output bytes.Buffer
			PathCmd.SetOut(&output)
			PathCmd.SetArgs([]string{actualProfileName})

			err := PathCmd.Execute()

			// Assert
			if tt.expectError {
				if err == nil {
					t.Errorf("%s: Expected error but got none", tt.description)
				}
			} else {
				if err != nil {
					t.Errorf("%s: Expected no error, got: %v", tt.description, err)
				}
			}
		})
	}
}

// Integration test
func TestPathCmd_Integration(t *testing.T) {
	// Arrange
	tempDir, cfg, cleanup := setupTestEnvironment(t)
	defer cleanup()

	// Create multiple profiles
	profileBasenames := []string{"profile1", "profile2", "profile3"}
	actualProfiles := make([]string, len(profileBasenames))
	for i, profileName := range profileBasenames {
		actualProfiles[i] = createTestProfile(t, cfg, profileName)
	}

	// Act & Assert - Test each profile
	for _, profileName := range actualProfiles {
		t.Run("profile_"+profileName, func(t *testing.T) {
			var output bytes.Buffer
			PathCmd.SetOut(&output)
			PathCmd.SetArgs([]string{profileName})

			err := PathCmd.Execute()
			if err != nil {
				t.Errorf("Failed to get path for profile '%s': %v", profileName, err)
			}

			result := output.String()
			expectedPath := filepath.Join(tempDir, ".mcpconfig", "profiles", profileName+".jsonc")

			if result != expectedPath {
				t.Errorf("Profile '%s': Expected path '%s', got '%s'", profileName, expectedPath, result)
			}
		})
	}
}

// Benchmark tests
func BenchmarkPathCmd_Execute(b *testing.B) {
	// Setup
	tempDir := b.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer func() { os.Setenv("HOME", oldHome) }()

	cfg, _ := config.New()
	profileManager := profile.NewManager(cfg.ProfilesDir)
	if err := profileManager.Create("bench-profile", "Benchmark profile"); err != nil {
		b.Fatalf("Failed to create benchmark profile: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var output bytes.Buffer
		PathCmd.SetOut(&output)
		PathCmd.SetArgs([]string{"bench-profile"})

		err := PathCmd.Execute()
		if err != nil {
			b.Fatalf("Benchmark failed: %v", err)
		}
	}
}

func BenchmarkPathCmd_DefaultProfile(b *testing.B) {
	// Setup
	tempDir := b.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer func() { os.Setenv("HOME", oldHome) }()

	cfg, _ := config.New()
	profileManager := profile.NewManager(cfg.ProfilesDir)
	if err := profileManager.Create(config.DefaultProfileName, "Default profile"); err != nil {
		b.Fatalf("Failed to create default profile: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var output bytes.Buffer
		PathCmd.SetOut(&output)
		PathCmd.SetArgs([]string{})

		err := PathCmd.Execute()
		if err != nil {
			b.Fatalf("Benchmark failed: %v", err)
		}
	}
}
