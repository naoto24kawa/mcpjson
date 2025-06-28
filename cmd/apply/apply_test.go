package apply

import (
	"os"
	"testing"

	"github.com/naoto24kawa/mcpconfig/internal/config"
	"github.com/naoto24kawa/mcpconfig/internal/profile"
	"github.com/naoto24kawa/mcpconfig/internal/server"
)

func setupTestEnvironment(t *testing.T) (string, func()) {
	t.Helper()

	tempDir, err := os.MkdirTemp("", "apply_test")
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
		name     string
		args     []string
		setup    func(cfg *config.Config)
		wantPanic bool
	}{
		{
			name: "デフォルトプロファイルの適用",
			args: []string{},
			setup: func(cfg *config.Config) {
				profileManager := profile.NewManager(cfg.ProfilesDir)
				serverManager := server.NewManager(cfg.ServersDir)

				serverManager.SaveManual("test-server", "python", []string{"test.py"}, nil, false)
				profileManager.Create("default", "")
				profileManager.AddServer("default", "test-server", "my-test-server", nil)
			},
			wantPanic: false,
		},
		{
			name: "特定プロファイルの適用",
			args: []string{"test-profile"},
			setup: func(cfg *config.Config) {
				profileManager := profile.NewManager(cfg.ProfilesDir)
				serverManager := server.NewManager(cfg.ServersDir)

				serverManager.SaveManual("test-server", "python", []string{"test.py"}, nil, false)
				profileManager.Create("test-profile", "")
				profileManager.AddServer("test-profile", "test-server", "my-test-server", nil)
			},
			wantPanic: false,
		},
		{
			name: "カスタムパスでの適用",
			args: []string{"test-profile", "--to", "/tmp/custom-config.json"},
			setup: func(cfg *config.Config) {
				profileManager := profile.NewManager(cfg.ProfilesDir)
				serverManager := server.NewManager(cfg.ServersDir)

				serverManager.SaveManual("test-server", "python", []string{"test.py"}, nil, false)
				profileManager.Create("test-profile", "")
				profileManager.AddServer("test-profile", "test-server", "my-test-server", nil)
			},
			wantPanic: false,
		},
		{
			name: "存在しないプロファイルの適用",
			args: []string{"non-existent"},
			setup: func(cfg *config.Config) {
			},
			wantPanic: true,
		},
		{
			name: "無効なプロファイル名",
			args: []string{"invalid-name!"},
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
				var outputPath string
				if len(tt.args) >= 3 && (tt.args[len(tt.args)-2] == "--to" || tt.args[len(tt.args)-2] == "-t") {
					outputPath = tt.args[len(tt.args)-1]
				} else {
					outputPath = config.GetDefaultMCPPath()
				}

				if outputPath != config.GetDefaultMCPPath() {
					if _, err := os.Stat(outputPath); os.IsNotExist(err) {
						t.Errorf("出力ファイルが作成されませんでした: %s", outputPath)
					}
				}
			}
		})
	}
}

func TestExecuteInvalidFlags(t *testing.T) {
	_, cleanup := setupTestEnvironment(t)
	defer cleanup()

	tests := []struct {
		name string
		args []string
	}{
		{
			name: "--toフラグに値がない",
			args: []string{"test-profile", "--to"},
		},
		{
			name: "-tフラグに値がない",
			args: []string{"test-profile", "-t"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r == nil {
					t.Error("パニックが期待されましたが発生しませんでした")
				}
			}()


			Execute(tt.args)
		})
	}
}