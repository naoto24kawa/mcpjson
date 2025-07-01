package save

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/naoto24kawa/mcpjson/internal/config"
	"github.com/naoto24kawa/mcpjson/internal/server"
)

func setupTestEnvironment(t *testing.T) (string, func()) {
	t.Helper()

	tempDir, err := os.MkdirTemp("", "save_test")
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

	return tempDir, cleanup
}

func createTestMCPConfig(path string) error {
	mcpConfig := server.MCPConfig{
		McpServers: map[string]server.MCPServer{
			"test-server": {
				Command: "python",
				Args:    []string{"test.py"},
			},
		},
	}

	data, err := json.MarshalIndent(mcpConfig, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

func TestExecute(t *testing.T) {
	tests := []struct {
		name  string
		args  func(tempDir string) []string
		setup func(tempDir string, cfg *config.Config) error
	}{
		{
			name: "MCPファイルからプロファイル保存",
			args: func(tempDir string) []string {
				return []string{"test-profile", "--from", filepath.Join(tempDir, "test-mcp.json")}
			},
			setup: func(tempDir string, cfg *config.Config) error {
				mcpPath := filepath.Join(tempDir, "test-mcp.json")
				return createTestMCPConfig(mcpPath)
			},
		},
		{
			name: "強制保存",
			args: func(tempDir string) []string {
				return []string{"test-profile", "--from", filepath.Join(tempDir, "test-mcp.json"), "--force"}
			},
			setup: func(tempDir string, cfg *config.Config) error {
				mcpPath := filepath.Join(tempDir, "test-mcp.json")
				return createTestMCPConfig(mcpPath)
			},
		},
		{
			name: "プロファイル名未指定でデフォルト使用",
			args: func(tempDir string) []string {
				return []string{"--from", filepath.Join(tempDir, "test-mcp.json")}
			},
			setup: func(tempDir string, cfg *config.Config) error {
				mcpPath := filepath.Join(tempDir, "test-mcp.json")
				return createTestMCPConfig(mcpPath)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir, cleanup := setupTestEnvironment(t)
			defer cleanup()

			cfg, err := config.New()
			if err != nil {
				t.Fatalf("設定の作成に失敗: %v", err)
			}

			if err := tt.setup(tempDir, cfg); err != nil {
				t.Fatalf("セットアップに失敗: %v", err)
			}

			defer func() {
				if r := recover(); r != nil {
					t.Errorf("予期しないパニック: %v", r)
				}
			}()

			Execute(tt.args(tempDir))
		})
	}
}
