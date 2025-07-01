package merge

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
	err := profileManager.Create(name, "Test profile for merge testing")
	if err != nil {
		t.Fatalf("Failed to create test profile: %v", err)
	}
}

func createTestProfileWithServers(t *testing.T, cfg *config.Config, name string, serverCount int) {
	t.Helper()

	profileManager := profile.NewManager(cfg.ProfilesDir)
	
	// Create profile with some server references
	testProfile := &profile.Profile{
		Name:        name,
		Description: "Test profile with servers",
		Servers:     make([]profile.ServerRef, serverCount),
	}
	
	// Add mock server references
	for i := 0; i < serverCount; i++ {
		testProfile.Servers[i] = profile.ServerRef{
			Name:     testutil.GenerateUniqueServerName("test-server"),
			Template: "mock-template",
		}
	}

	// Save the profile directly
	if err := profileManager.Create(name, "Test profile"); err != nil {
		t.Fatalf("Failed to create test profile: %v", err)
	}

	// Load and update with servers
	loadedProfile, err := profileManager.Load(name)
	if err != nil {
		t.Fatalf("Failed to load test profile: %v", err)
	}
	
	loadedProfile.Servers = testProfile.Servers
	profilePath := filepath.Join(cfg.ProfilesDir, name+config.FileExtension)
	if err := utils.SaveJSON(profilePath, loadedProfile); err != nil {
		t.Fatalf("Failed to save test profile with servers: %v", err)
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
			name: "複数プロファイルの正常なマージ",
			args: []string{"merged", "source1", "source2"},
			setup: func(cfg *config.Config) {
				createTestProfileWithServers(t, cfg, "source1", 2)
				createTestProfileWithServers(t, cfg, "source2", 1)
			},
			wantErr: false,
		},
		{
			name: "単一プロファイルのマージ",
			args: []string{"merged", "source1"},
			setup: func(cfg *config.Config) {
				createTestProfileWithServers(t, cfg, "source1", 2)
			},
			wantErr: false,
		},
		{
			name: "3つのプロファイルのマージ",
			args: []string{"merged", "source1", "source2", "source3"},
			setup: func(cfg *config.Config) {
				createTestProfileWithServers(t, cfg, "source1", 1)
				createTestProfileWithServers(t, cfg, "source2", 1)
				createTestProfileWithServers(t, cfg, "source3", 1)
			},
			wantErr: false,
		},
		{
			name:    "存在しないソースプロファイル",
			args:    []string{"merged", "nonexistent"},
			setup:   func(cfg *config.Config) {},
			wantErr: true,
		},
		{
			name: "既存の宛先プロファイル（forceなし）",
			args: []string{"existing", "source1"},
			setup: func(cfg *config.Config) {
				createTestProfile(t, cfg, "source1")
				createTestProfile(t, cfg, "existing")
			},
			wantErr: true,
		},
		{
			name: "既存の宛先プロファイル（forceあり）",
			args: []string{"existing", "source1", "--force"},
			setup: func(cfg *config.Config) {
				createTestProfile(t, cfg, "source1")
				createTestProfile(t, cfg, "existing")
			},
			wantErr: false,
		},
		{
			name: "forceフラグ（短縮形）",
			args: []string{"existing", "source1", "-f"},
			setup: func(cfg *config.Config) {
				createTestProfile(t, cfg, "source1")
				createTestProfile(t, cfg, "existing")
			},
			wantErr: false,
		},
		{
			name:    "引数不足（宛先名のみ）",
			args:    []string{"merged"},
			setup:   func(cfg *config.Config) {},
			wantErr: true,
		},
		{
			name:    "引数不足（引数なし）",
			args:    []string{},
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
				testMergeLogic(t, cfg, tt.args, true)
			} else {
				testMergeLogic(t, cfg, tt.args, false)
			}
		})
	}
}

// testMergeLogic tests the merge logic without calling Execute (to avoid os.Exit)
func testMergeLogic(t *testing.T, cfg *config.Config, args []string, expectError bool) {
	t.Helper()

	// Parse arguments similar to Execute function
	if len(args) < 2 {
		if expectError {
			return // Expected error case
		}
		t.Fatal("Expected at least 2 arguments for merge command")
	}

	destName := args[0]
	sourceNames := []string{}
	force := false

	// Parse source names and flags
	for i := 1; i < len(args); i++ {
		if args[i] == "--force" || args[i] == "-f" {
			force = true
		} else {
			sourceNames = append(sourceNames, args[i])
		}
	}

	if len(sourceNames) == 0 {
		if expectError {
			return // Expected error case
		}
		t.Fatal("Expected at least one source profile")
	}

	// Test the merge operation
	profileManager := profile.NewManager(cfg.ProfilesDir)
	err := profileManager.Merge(destName, sourceNames, force)

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

	// Verify all source profiles still exist
	for _, sourceName := range sourceNames {
		sourcePath := filepath.Join(cfg.ProfilesDir, sourceName+config.FileExtension)
		if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
			t.Errorf("Source profile should still exist: %s", sourcePath)
		}
	}
}

func TestExecute_DuplicateServers(t *testing.T) {
	cfg, cleanup := setupTestEnvironment(t)
	defer cleanup()

	// Create profiles with overlapping server names
	profileManager := profile.NewManager(cfg.ProfilesDir)
	
	// Create profile1 with server "common-server"
	err := profileManager.Create("profile1", "Test profile 1")
	if err != nil {
		t.Fatalf("Failed to create profile1: %v", err)
	}
	
	profile1, err := profileManager.Load("profile1")
	if err != nil {
		t.Fatalf("Failed to load profile1: %v", err)
	}
	
	profile1.Servers = []profile.ServerRef{
		{Name: "common-server", Template: "template1"},
		{Name: "unique-server1", Template: "template1"},
	}
	
	profile1Path := filepath.Join(cfg.ProfilesDir, "profile1"+config.FileExtension)
	if err := utils.SaveJSON(profile1Path, profile1); err != nil {
		t.Fatalf("Failed to save profile1: %v", err)
	}

	// Create profile2 with same server "common-server"
	err = profileManager.Create("profile2", "Test profile 2")
	if err != nil {
		t.Fatalf("Failed to create profile2: %v", err)
	}
	
	profile2, err := profileManager.Load("profile2")
	if err != nil {
		t.Fatalf("Failed to load profile2: %v", err)
	}
	
	profile2.Servers = []profile.ServerRef{
		{Name: "common-server", Template: "template2"}, // Same name, different template
		{Name: "unique-server2", Template: "template2"},
	}
	
	profile2Path := filepath.Join(cfg.ProfilesDir, "profile2"+config.FileExtension)
	if err := utils.SaveJSON(profile2Path, profile2); err != nil {
		t.Fatalf("Failed to save profile2: %v", err)
	}

	// Test merge with duplicate servers
	args := []string{"merged", "profile1", "profile2"}
	testMergeLogic(t, cfg, args, false)

	// Verify the merged profile contains servers with first-wins policy
	mergedProfile, err := profileManager.Load("merged")
	if err != nil {
		t.Fatalf("Failed to load merged profile: %v", err)
	}

	expectedServers := 3 // common-server (from profile1), unique-server1, unique-server2
	if len(mergedProfile.Servers) != expectedServers {
		t.Errorf("Expected %d servers in merged profile, got %d", expectedServers, len(mergedProfile.Servers))
	}

	// Verify first-wins policy for common-server
	var commonServer *profile.ServerRef
	for _, server := range mergedProfile.Servers {
		if server.Name == "common-server" {
			commonServer = &server
			break
		}
	}

	if commonServer == nil {
		t.Error("common-server not found in merged profile")
	} else if commonServer.Template != "template1" {
		t.Errorf("Expected common-server to have template1 (first-wins), got %s", commonServer.Template)
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
			name: "無効な文字を含む宛先名",
			args: []string{"invalid@dest", "valid-source"},
		},
		{
			name: "無効な文字を含むソース名",
			args: []string{"dest", "invalid@source"},
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
				// Test dest name validation
				if err := utils.ValidateName(tt.args[0], "プロファイル"); err == nil {
					// Test source name validation
					if err := utils.ValidateName(tt.args[1], "プロファイル"); err == nil {
						t.Error("Expected validation error for invalid profile name")
					}
				}
			}
		})
	}
}