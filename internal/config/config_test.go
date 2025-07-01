package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNew(t *testing.T) {
	// テスト用の一時ホームディレクトリを設定
	tempDir := t.TempDir()

	// WindowsとUnixで環境変数を適切に設定
	originalHome := os.Getenv("HOME")
	originalUserProfile := os.Getenv("USERPROFILE")

	os.Setenv("HOME", tempDir)
	os.Setenv("USERPROFILE", tempDir)

	defer func() {
		os.Setenv("HOME", originalHome)
		os.Setenv("USERPROFILE", originalUserProfile)
	}()

	cfg, err := New()
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	// ディレクトリが作成されているか確認
	expectedDirs := []string{
		filepath.Join(tempDir, ConfigDirName),
		filepath.Join(tempDir, ConfigDirName, ProfilesDir),
		filepath.Join(tempDir, ConfigDirName, ServersDir),
		filepath.Join(tempDir, ConfigDirName, GroupsDir),
	}

	for _, dir := range expectedDirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			t.Errorf("Directory not created: %s", dir)
		}
	}

	// パスが正しく設定されているか確認
	if cfg.BaseDir != filepath.Join(tempDir, ConfigDirName) {
		t.Errorf("BaseDir = %v, want %v", cfg.BaseDir, filepath.Join(tempDir, ConfigDirName))
	}

	// GroupsDirが正しく設定されているか確認
	expectedGroupsDir := filepath.Join(tempDir, ConfigDirName, GroupsDir)
	if cfg.GroupsDir != expectedGroupsDir {
		t.Errorf("GroupsDir = %v, want %v", cfg.GroupsDir, expectedGroupsDir)
	}
}

func TestGetProfilePath(t *testing.T) {
	// filepath.Joinを使用してOSに依存しないパスを作成
	baseDir := filepath.Join("home", "user", ".mcpconfig")
	profilesDir := filepath.Join(baseDir, "profiles")

	cfg := &Config{
		ProfilesDir: profilesDir,
	}

	got := cfg.GetProfilePath("test-profile")
	want := filepath.Join(profilesDir, "test-profile.jsonc")

	if got != want {
		t.Errorf("GetProfilePath() = %v, want %v", got, want)
	}
}

func TestGetServerPath(t *testing.T) {
	// filepath.Joinを使用してOSに依存しないパスを作成
	baseDir := filepath.Join("home", "user", ".mcpconfig")
	serversDir := filepath.Join(baseDir, "servers")

	cfg := &Config{
		ServersDir: serversDir,
	}

	got := cfg.GetServerPath("test-server")
	want := filepath.Join(serversDir, "test-server.jsonc")

	if got != want {
		t.Errorf("GetServerPath() = %v, want %v", got, want)
	}
}

func TestGetGroupPath(t *testing.T) {
	// filepath.Joinを使用してOSに依存しないパスを作成
	baseDir := filepath.Join("home", "user", ".mcpconfig")
	groupsDir := filepath.Join(baseDir, "groups")

	cfg := &Config{
		GroupsDir: groupsDir,
	}

	got := cfg.GetGroupPath("test-group")
	want := filepath.Join(groupsDir, "test-group.jsonc")

	if got != want {
		t.Errorf("GetGroupPath() = %v, want %v", got, want)
	}
}

func TestConfig_AllPaths(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	baseDir := filepath.Join(tempDir, ConfigDirName)

	cfg := &Config{
		BaseDir:     baseDir,
		ProfilesDir: filepath.Join(baseDir, ProfilesDir),
		ServersDir:  filepath.Join(baseDir, ServersDir),
		GroupsDir:   filepath.Join(baseDir, GroupsDir),
	}

	// Act & Assert for profiles
	profilePath := cfg.GetProfilePath("test-profile")
	expectedProfilePath := filepath.Join(cfg.ProfilesDir, "test-profile.jsonc")
	if profilePath != expectedProfilePath {
		t.Errorf("GetProfilePath() = %v, want %v", profilePath, expectedProfilePath)
	}

	// Act & Assert for servers
	serverPath := cfg.GetServerPath("test-server")
	expectedServerPath := filepath.Join(cfg.ServersDir, "test-server.jsonc")
	if serverPath != expectedServerPath {
		t.Errorf("GetServerPath() = %v, want %v", serverPath, expectedServerPath)
	}

	// Act & Assert for groups
	groupPath := cfg.GetGroupPath("test-group")
	expectedGroupPath := filepath.Join(cfg.GroupsDir, "test-group.jsonc")
	if groupPath != expectedGroupPath {
		t.Errorf("GetGroupPath() = %v, want %v", groupPath, expectedGroupPath)
	}
}

func TestConfig_GroupsDirIntegration(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()

	originalHome := os.Getenv("HOME")
	originalUserProfile := os.Getenv("USERPROFILE")

	os.Setenv("HOME", tempDir)
	os.Setenv("USERPROFILE", tempDir)

	defer func() {
		os.Setenv("HOME", originalHome)
		os.Setenv("USERPROFILE", originalUserProfile)
	}()

	// Act
	cfg, err := New()
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	// Assert - GroupsDir should be properly initialized
	if cfg.GroupsDir == "" {
		t.Error("GroupsDir should not be empty after New()")
	}

	expectedGroupsDir := filepath.Join(tempDir, ConfigDirName, GroupsDir)
	if cfg.GroupsDir != expectedGroupsDir {
		t.Errorf("GroupsDir = %v, want %v", cfg.GroupsDir, expectedGroupsDir)
	}

	// Assert - GroupsDir directory should exist
	if _, err := os.Stat(cfg.GroupsDir); os.IsNotExist(err) {
		t.Errorf("GroupsDir directory should exist: %s", cfg.GroupsDir)
	}

	// Assert - GetGroupPath should work correctly
	testGroupName := "integration-test-group"
	groupPath := cfg.GetGroupPath(testGroupName)
	expectedGroupPath := filepath.Join(cfg.GroupsDir, testGroupName+".jsonc")
	if groupPath != expectedGroupPath {
		t.Errorf("GetGroupPath() = %v, want %v", groupPath, expectedGroupPath)
	}
}

func TestConfig_PathHandling(t *testing.T) {
	tests := []struct {
		name       string
		methodName string
		inputName  string
		getPath    func(*Config, string) string
		baseDir    string
	}{
		{
			name:       "profile path with spaces",
			methodName: "GetProfilePath",
			inputName:  "profile with spaces",
			getPath:    (*Config).GetProfilePath,
			baseDir:    "profiles",
		},
		{
			name:       "server path with hyphens",
			methodName: "GetServerPath",
			inputName:  "server-with-hyphens",
			getPath:    (*Config).GetServerPath,
			baseDir:    "servers",
		},
		{
			name:       "group path with underscores",
			methodName: "GetGroupPath",
			inputName:  "group_with_underscores",
			getPath:    (*Config).GetGroupPath,
			baseDir:    "groups",
		},
		{
			name:       "empty name handling",
			methodName: "GetGroupPath",
			inputName:  "",
			getPath:    (*Config).GetGroupPath,
			baseDir:    "groups",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			baseDir := filepath.Join("test", "config")
			targetDir := filepath.Join(baseDir, tt.baseDir)

			cfg := &Config{
				BaseDir:     baseDir,
				ProfilesDir: filepath.Join(baseDir, "profiles"),
				ServersDir:  filepath.Join(baseDir, "servers"),
				GroupsDir:   filepath.Join(baseDir, "groups"),
			}

			// Act
			result := tt.getPath(cfg, tt.inputName)

			// Assert
			expected := filepath.Join(targetDir, tt.inputName+".jsonc")
			if result != expected {
				t.Errorf("%s(%q) = %v, want %v", tt.methodName, tt.inputName, result, expected)
			}

			// Verify the path contains the correct extension
			if !strings.HasSuffix(result, ".jsonc") {
				t.Errorf("%s should return path with .jsonc extension, got: %s", tt.methodName, result)
			}
		})
	}
}
