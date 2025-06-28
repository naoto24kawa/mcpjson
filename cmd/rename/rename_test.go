package rename

import (
	"os"
	"testing"

	"github.com/naoto24kawa/mcpconfig/internal/config"
	"github.com/naoto24kawa/mcpconfig/internal/profile"
)

func setupTestEnvironment(t *testing.T) func() {
	t.Helper()

	tempDir, err := os.MkdirTemp("", "rename_test")
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

	return cleanup
}

func TestExecute(t *testing.T) {
	tests := []struct {
		name  string
		args  []string
		setup func(cfg *config.Config)
	}{
		{
			name: "プロファイルのリネーム",
			args: []string{"old-name1", "new-name1"},
			setup: func(cfg *config.Config) {
				profileManager := profile.NewManager(cfg.ProfilesDir)
				profileManager.Create("old-name1", "test description")
			},
		},
		{
			name: "プロファイルのリネーム（強制）",
			args: []string{"old-name2", "new-name2", "--force"},
			setup: func(cfg *config.Config) {
				profileManager := profile.NewManager(cfg.ProfilesDir)
				profileManager.Create("old-name2", "test description")
			},
		},
		{
			name: "デフォルトプロファイルのリネーム",
			args: []string{"new-name3"},
			setup: func(cfg *config.Config) {
				profileManager := profile.NewManager(cfg.ProfilesDir)
				profileManager.Create(config.DefaultProfileName, "test description")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := setupTestEnvironment(t)
			defer cleanup()

			cfg, err := config.New()
			if err != nil {
				t.Fatalf("設定の作成に失敗: %v", err)
			}

			tt.setup(cfg)

			defer func() {
				if r := recover(); r != nil {
					t.Errorf("予期しないパニック: %v", r)
				}
			}()

			Execute(tt.args)
		})
	}
}