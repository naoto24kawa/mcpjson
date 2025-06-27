package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNew(t *testing.T) {
	// テスト用の一時ホームディレクトリを設定
	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	cfg, err := New()
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	// ディレクトリが作成されているか確認
	expectedDirs := []string{
		filepath.Join(tempDir, ConfigDirName),
		filepath.Join(tempDir, ConfigDirName, ProfilesDir),
		filepath.Join(tempDir, ConfigDirName, ServersDir),
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
}

func TestGetProfilePath(t *testing.T) {
	cfg := &Config{
		ProfilesDir: "/home/user/.mcpconfig/profiles",
	}

	got := cfg.GetProfilePath("test-profile")
	want := "/home/user/.mcpconfig/profiles/test-profile.json"

	if got != want {
		t.Errorf("GetProfilePath() = %v, want %v", got, want)
	}
}

func TestGetServerPath(t *testing.T) {
	cfg := &Config{
		ServersDir: "/home/user/.mcpconfig/servers",
	}

	got := cfg.GetServerPath("test-server")
	want := "/home/user/.mcpconfig/servers/test-server.json"

	if got != want {
		t.Errorf("GetServerPath() = %v, want %v", got, want)
	}
}