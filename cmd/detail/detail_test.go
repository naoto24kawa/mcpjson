package detail

import (
	"os"
	"strings"
	"testing"

	"github.com/naoto24kawa/mcpconfig/internal/config"
	"github.com/naoto24kawa/mcpconfig/internal/profile"
)

func setupTestEnvironment(t *testing.T) func() {
	t.Helper()

	tempDir, err := os.MkdirTemp("", "detail_test")
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
		name    string
		args    []string
		setup   func(cfg *config.Config)
		wantErr bool
	}{
		{
			name: "引数なし",
			args: []string{},
			setup: func(cfg *config.Config) {
			},
			wantErr: true,
		},
		{
			name: "存在しないプロファイル",
			args: []string{"non-existent"},
			setup: func(cfg *config.Config) {
			},
			wantErr: true,
		},
		{
			name: "正常なプロファイル詳細表示",
			args: []string{"test-profile"},
			setup: func(cfg *config.Config) {
				profileManager := profile.NewManager(cfg.ProfilesDir)
				profileManager.Create("test-profile", "テスト用プロファイル")
			},
			wantErr: false,
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

			err = Execute(tt.args)

			if tt.wantErr && err == nil {
				t.Error("エラーが期待されましたが発生しませんでした")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("予期しないエラー: %v", err)
			}
		})
	}
}

func TestShowProfileDetail(t *testing.T) {
	cleanup := setupTestEnvironment(t)
	defer cleanup()

	cfg, err := config.New()
	if err != nil {
		t.Fatalf("設定の作成に失敗: %v", err)
	}

	profileManager := profile.NewManager(cfg.ProfilesDir)
	profileManager.Create("test-detail", "詳細テスト用")

	err = showProfileDetail("test-detail")
	if err != nil {
		t.Errorf("showProfileDetail() error = %v", err)
	}

	err = showProfileDetail("non-existent")
	if err == nil {
		t.Error("存在しないプロファイルでエラーが発生しませんでした")
	}
	if !strings.Contains(err.Error(), "見つかりません") {
		t.Errorf("期待されるエラーメッセージではありません: %v", err)
	}
}