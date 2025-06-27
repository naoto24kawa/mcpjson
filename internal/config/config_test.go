package config

import (
	"os"
	"path/filepath"
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
	// filepath.Joinを使用してOSに依存しないパスを作成
	baseDir := filepath.Join("home", "user", ".mcpconfig")
	profilesDir := filepath.Join(baseDir, "profiles")
	
	cfg := &Config{
		ProfilesDir: profilesDir,
	}

	got := cfg.GetProfilePath("test-profile")
	want := filepath.Join(profilesDir, "test-profile.json")

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
	want := filepath.Join(serversDir, "test-server.json")

	if got != want {
		t.Errorf("GetServerPath() = %v, want %v", got, want)
	}
}