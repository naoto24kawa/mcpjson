package create

import (
	"os"
	"testing"

	"github.com/naoto24kawa/mcpconfig/internal/config"
	"github.com/naoto24kawa/mcpconfig/internal/profile"
	"github.com/naoto24kawa/mcpconfig/internal/server"
)

func setupTestEnvironment(t *testing.T) (string, func()) {
	t.Helper()

	tempDir, err := os.MkdirTemp("", "create_test")
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
		name      string
		args      []string
		setup     func(cfg *config.Config)
	}{
		{
			name: "特定プロファイルの作成",
			args: []string{"test-profile"},
			setup: func(cfg *config.Config) {
			},
		},
		{
			name: "テンプレート指定でのプロファイル作成",
			args: []string{"test-profile2", "--template", "test-template"},
			setup: func(cfg *config.Config) {
				serverManager := server.NewManager(cfg.ServersDir)
				_ = serverManager.SaveManual("test-template", "python", []string{"test.py"}, nil, false)
			},
		},
		{
			name: "テンプレート指定（短縮形）でのプロファイル作成",
			args: []string{"test-profile3", "-t", "test-template"},
			setup: func(cfg *config.Config) {
				serverManager := server.NewManager(cfg.ServersDir)
				_ = serverManager.SaveManual("test-template", "python", []string{"test.py"}, nil, false)
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
			if len(tt.args) > 0 && tt.args[0] != "--template" && tt.args[0] != "-t" {
				profileName = tt.args[0]
			}

			if _, err := profileManager.Load(profileName); err != nil {
				t.Errorf("プロファイル '%s' が作成されませんでした: %v", profileName, err)
			}
		})
	}
}

