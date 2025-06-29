package reset

import (
	"os"
	"testing"

	"github.com/naoto24kawa/mcpconfig/internal/config"
	"github.com/naoto24kawa/mcpconfig/internal/profile"
	"github.com/naoto24kawa/mcpconfig/internal/server"
)

func setupTestEnvironment(t *testing.T) (*config.Config, func()) {
	t.Helper()

	tempDir, err := os.MkdirTemp("", "reset_test")
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

	_ = os.MkdirAll(cfg.ProfilesDir, 0755)
	_ = os.MkdirAll(cfg.ServersDir, 0755)

	return cfg, cleanup
}

func TestExecute(t *testing.T) {
	tests := []struct {
		name  string
		args  []string
		setup func(cfg *config.Config)
	}{
		{
			name: "プロファイルリセット（強制）",
			args: []string{"profiles", "--force"},
			setup: func(cfg *config.Config) {
				profileManager := profile.NewManager(cfg.ProfilesDir)
				_ = profileManager.Create("test-profile", "test description")
			},
		},
		{
			name: "サーバーリセット（強制）",
			args: []string{"servers", "--force"},
			setup: func(cfg *config.Config) {
				serverManager := server.NewManager(cfg.ServersDir)
				_ = serverManager.SaveManual("test-server", "python", []string{"test.py"}, nil, false)
			},
		},
		{
			name: "全リセット（強制）",
			args: []string{"all", "--force"},
			setup: func(cfg *config.Config) {
				profileManager := profile.NewManager(cfg.ProfilesDir)
				_ = profileManager.Create("test-profile", "test description")
				serverManager := server.NewManager(cfg.ServersDir)
				_ = serverManager.SaveManual("test-server", "python", []string{"test.py"}, nil, false)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, cleanup := setupTestEnvironment(t)
			defer cleanup()

			tt.setup(cfg)

			defer func() {
				if r := recover(); r != nil {
					t.Errorf("予期しないパニック: %v", r)
				}
			}()

			Execute(cfg, tt.args)
		})
	}
}
