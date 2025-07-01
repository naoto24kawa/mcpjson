package delete

import (
	"os"
	"testing"

	"github.com/naoto24kawa/mcpjson/internal/config"
	"github.com/naoto24kawa/mcpjson/internal/profile"
)

func setupTestEnvironment(t *testing.T) (string, func()) {
	t.Helper()

	tempDir, err := os.MkdirTemp("", "delete_test")
	if err != nil {
		t.Fatalf("一時ディレクトリの作成に失敗: %v", err)
	}

	origXDGConfigHome := os.Getenv("XDG_CONFIG_HOME")
	os.Setenv("XDG_CONFIG_HOME", tempDir)

	cleanup := func() {
		os.Setenv("XDG_CONFIG_HOME", origXDGConfigHome)
		os.RemoveAll(tempDir)
	}

	cfg, err := config.New()
	if err != nil {
		cleanup()
		t.Fatalf("設定の作成に失敗: %v", err)
	}

	os.MkdirAll(cfg.ProfilesDir, 0755)
	os.MkdirAll(cfg.ServersDir, 0755)

	return tempDir, cleanup
}

func TestExecute(t *testing.T) {
	tests := []struct {
		name  string
		args  []string
		setup func(cfg *config.Config)
	}{
		{
			name: "デフォルトプロファイルの強制削除",
			args: []string{config.DefaultProfileName, "--force"},
			setup: func(cfg *config.Config) {
				profileManager := profile.NewManager(cfg.ProfilesDir)
				_ = profileManager.Create(config.DefaultProfileName, "")
			},
		},
		{
			name: "デフォルトプロファイルの強制削除（短縮形）",
			args: []string{config.DefaultProfileName, "-f"},
			setup: func(cfg *config.Config) {
				profileManager := profile.NewManager(cfg.ProfilesDir)
				_ = profileManager.Create(config.DefaultProfileName, "")
			},
		},
		{
			name: "特定プロファイルの強制削除",
			args: []string{"test-profile", "--force"},
			setup: func(cfg *config.Config) {
				profileManager := profile.NewManager(cfg.ProfilesDir)
				profileManager.Create("test-profile", "")
			},
		},
		{
			name: "特定プロファイルの強制削除（短縮形）",
			args: []string{"test-profile2", "-f"},
			setup: func(cfg *config.Config) {
				profileManager := profile.NewManager(cfg.ProfilesDir)
				profileManager.Create("test-profile2", "")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, cleanup := setupTestEnvironment(t)
			defer cleanup()

			cfg, err := config.New()
			if err != nil {
				t.Fatalf("設定の作成に失敗: %v", err)
			}

			tt.setup(cfg)

			Execute(tt.args)

			profileManager := profile.NewManager(cfg.ProfilesDir)
			profileName := config.DefaultProfileName
			if len(tt.args) > 0 && tt.args[0] != "--force" && tt.args[0] != "-f" {
				profileName = tt.args[0]
			}

			if _, err := profileManager.Load(profileName); err == nil {
				t.Errorf("プロファイル '%s' が削除されませんでした", profileName)
			}
		})
	}
}

func TestExecuteWithMultipleFlags(t *testing.T) {
	_, cleanup := setupTestEnvironment(t)
	defer cleanup()

	cfg, err := config.New()
	if err != nil {
		t.Fatalf("設定の作成に失敗: %v", err)
	}

	profileManager := profile.NewManager(cfg.ProfilesDir)
	profileManager.Create("test-profile3", "")

	Execute([]string{"test-profile3", "--force", "-f"})

	if _, err := profileManager.Load("test-profile3"); err == nil {
		t.Error("プロファイル 'test-profile3' が削除されませんでした")
	}
}
