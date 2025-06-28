package delete

import (
	"os"
	"testing"

	"github.com/naoto24kawa/mcpconfig/internal/config"
	"github.com/naoto24kawa/mcpconfig/internal/profile"
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
		name      string
		args      []string
		setup     func(cfg *config.Config)
		wantPanic bool
	}{
		{
			name: "デフォルトプロファイルの強制削除",
			args: []string{config.DefaultProfileName, "--force"},
			setup: func(cfg *config.Config) {
				profileManager := profile.NewManager(cfg.ProfilesDir)
				profileManager.Create(config.DefaultProfileName, "")
			},
			wantPanic: false,
		},
		{
			name: "デフォルトプロファイルの強制削除（短縮形）",
			args: []string{config.DefaultProfileName, "-f"},
			setup: func(cfg *config.Config) {
				profileManager := profile.NewManager(cfg.ProfilesDir)
				profileManager.Create(config.DefaultProfileName, "")
			},
			wantPanic: false,
		},
		{
			name: "特定プロファイルの強制削除",
			args: []string{"test-profile", "--force"},
			setup: func(cfg *config.Config) {
				profileManager := profile.NewManager(cfg.ProfilesDir)
				profileManager.Create("test-profile", "")
			},
			wantPanic: false,
		},
		{
			name: "特定プロファイルの強制削除（短縮形）",
			args: []string{"test-profile", "-f"},
			setup: func(cfg *config.Config) {
				profileManager := profile.NewManager(cfg.ProfilesDir)
				profileManager.Create("test-profile", "")
			},
			wantPanic: false,
		},
		{
			name: "存在しないプロファイルの削除",
			args: []string{"non-existent", "--force"},
			setup: func(cfg *config.Config) {
			},
			wantPanic: true,
		},
		{
			name: "無効なプロファイル名",
			args: []string{"invalid-name!", "--force"},
			setup: func(cfg *config.Config) {
			},
			wantPanic: true,
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

			defer func() {
				r := recover()
				if tt.wantPanic && r == nil {
					t.Error("パニックが期待されましたが発生しませんでした")
				}
				if !tt.wantPanic && r != nil {
					t.Errorf("予期しないパニック: %v", r)
				}
			}()


			Execute(tt.args)

			if !tt.wantPanic {
				profileManager := profile.NewManager(cfg.ProfilesDir)
				profileName := config.DefaultProfileName
				if len(tt.args) > 0 && tt.args[0] != "--force" && tt.args[0] != "-f" {
					profileName = tt.args[0]
				}

				if _, err := profileManager.Load(profileName); err == nil {
					t.Errorf("プロファイル '%s' が削除されませんでした", profileName)
				}
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
	profileManager.Create("test-profile", "")

	defer func() {
		r := recover()
		if r != nil {
			t.Errorf("予期しないパニック: %v", r)
		}
	}()

	Execute([]string{"test-profile", "--force", "-f"})

	if _, err := profileManager.Load("test-profile"); err == nil {
		t.Error("プロファイル 'test-profile' が削除されませんでした")
	}
}