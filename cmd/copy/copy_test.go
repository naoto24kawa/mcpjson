package copy

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/naoto24kawa/mcpjson/internal/config"
	"github.com/naoto24kawa/mcpjson/internal/profile"
	"github.com/naoto24kawa/mcpjson/internal/testutil"
	"github.com/naoto24kawa/mcpjson/internal/utils"
)

func setupTestEnvironment(t *testing.T) (*config.Config, func()) {
	t.Helper()

	tempDir, cfg, cleanup := testutil.SetupIsolatedTestEnvironment(t)
	_ = tempDir

	return cfg, cleanup
}

func createTestProfile(t *testing.T, cfg *config.Config, name string) {
	t.Helper()

	profileManager := profile.NewManager(cfg.ProfilesDir)
	err := profileManager.Create(name, "Test profile for copy testing")
	if err != nil {
		t.Fatalf("Failed to create test profile: %v", err)
	}
}

func TestExecute(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		setup   func(cfg *config.Config)
		wantErr bool
	}{
		{
			name: "正常なプロファイルコピー",
			args: []string{"source", "dest"},
			setup: func(cfg *config.Config) {
				createTestProfile(t, cfg, "source")
			},
			wantErr: false,
		},
		{
			name: "デフォルトプロファイルから別名でコピー",
			args: []string{"copied"},
			setup: func(cfg *config.Config) {
				createTestProfile(t, cfg, "default")
			},
			wantErr: false,
		},
		{
			name:    "存在しないソースプロファイル",
			args:    []string{"nonexistent", "dest"},
			setup:   func(cfg *config.Config) {},
			wantErr: true,
		},
		{
			name: "既存の宛先プロファイル（forceなし）",
			args: []string{"source", "dest"},
			setup: func(cfg *config.Config) {
				createTestProfile(t, cfg, "source")
				createTestProfile(t, cfg, "dest")
			},
			wantErr: true,
		},
		{
			name: "既存の宛先プロファイル（forceあり）",
			args: []string{"source", "dest", "--force"},
			setup: func(cfg *config.Config) {
				createTestProfile(t, cfg, "source")
				createTestProfile(t, cfg, "dest")
			},
			wantErr: false,
		},
		{
			name: "forceフラグ（短縮形）",
			args: []string{"source", "dest", "-f"},
			setup: func(cfg *config.Config) {
				createTestProfile(t, cfg, "source")
				createTestProfile(t, cfg, "dest")
			},
			wantErr: false,
		},
		{
			name:    "引数不足",
			args:    []string{"source"},
			setup:   func(cfg *config.Config) {},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			cfg, cleanup := setupTestEnvironment(t)
			defer cleanup()

			// Setup test data
			tt.setup(cfg)

			// Capture stderr for error testing
			oldStderr := os.Stderr
			if tt.wantErr {
				// Redirect stderr to suppress error output during testing
				os.Stderr, _ = os.Open(os.DevNull)
				defer func() { os.Stderr = oldStderr }()
			}

			// Act & Assert
			if tt.wantErr {
				// Test should call os.Exit, so we run it in a subprocess-like manner
				// For now, we'll test the underlying logic without Execute
				testCopyLogic(t, cfg, tt.args, true)
			} else {
				testCopyLogic(t, cfg, tt.args, false)
			}
		})
	}
}

// testCopyLogic tests the copy logic without calling Execute (to avoid os.Exit)
func testCopyLogic(t *testing.T, cfg *config.Config, args []string, expectError bool) {
	t.Helper()

	// Parse arguments similar to Execute function
	if len(args) < 1 {
		if expectError {
			return // Expected error case
		}
		t.Fatal("Expected at least 1 argument for copy command")
	}

	var sourceName, destName string
	var force bool

	if len(args) == 1 {
		sourceName = "default"
		destName = args[0]
	} else {
		sourceName = args[0]
		destName = args[1]
	}

	// Check for force flag
	for i := 2; i < len(args); i++ {
		if args[i] == "--force" || args[i] == "-f" {
			force = true
		}
	}

	// Test the copy operation
	profileManager := profile.NewManager(cfg.ProfilesDir)
	err := profileManager.Copy(sourceName, destName, force)

	if expectError {
		if err == nil {
			t.Error("Expected error but got none")
		}
		return
	}

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
		return
	}

	// Verify the destination profile was created
	destPath := filepath.Join(cfg.ProfilesDir, destName+config.FileExtension)
	if _, err := os.Stat(destPath); os.IsNotExist(err) {
		t.Errorf("Destination profile was not created: %s", destPath)
	}

	// Verify the source profile still exists
	sourcePath := filepath.Join(cfg.ProfilesDir, sourceName+config.FileExtension)
	if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
		t.Errorf("Source profile should still exist: %s", sourcePath)
	}
}

func TestExecute_InvalidProfileName(t *testing.T) {
	cfg, cleanup := setupTestEnvironment(t)
	defer cleanup()

	createTestProfile(t, cfg, "valid-source")

	tests := []struct {
		name string
		args []string
	}{
		{
			name: "無効な文字を含むソース名",
			args: []string{"invalid@source", "dest"},
		},
		{
			name: "無効な文字を含む宛先名",
			args: []string{"valid-source", "invalid@dest"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Redirect stderr to suppress error output
			oldStderr := os.Stderr
			os.Stderr, _ = os.Open(os.DevNull)
			defer func() { os.Stderr = oldStderr }()

			// Test validation by calling utils.ValidateName directly
			if len(tt.args) >= 2 {
				// For copy, test both source and dest validation
				if err := utils.ValidateName(tt.args[0], "プロファイル"); err == nil {
					if err := utils.ValidateName(tt.args[1], "プロファイル"); err == nil {
						t.Error("Expected validation error for invalid profile name")
					}
				}
			}
		})
	}
}
