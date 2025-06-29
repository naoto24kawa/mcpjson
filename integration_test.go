package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/naoto24kawa/mcpconfig/internal/config"
	"github.com/naoto24kawa/mcpconfig/internal/profile"
	"github.com/naoto24kawa/mcpconfig/internal/server"
)

func TestIntegration_ProfileAndServerWorkflow(t *testing.T) {
	// テスト用の一時ディレクトリを作成
	tempDir := t.TempDir()
	profilesDir := filepath.Join(tempDir, "profiles")
	serversDir := filepath.Join(tempDir, "servers")

	err := os.MkdirAll(profilesDir, 0755)
	if err != nil {
		t.Fatalf("プロファイルディレクトリ作成に失敗: %v", err)
	}

	err = os.MkdirAll(serversDir, 0755)
	if err != nil {
		t.Fatalf("サーバーディレクトリ作成に失敗: %v", err)
	}

	// 設定を作成（現在は直接使用しないが、将来の拡張のために残す）
	_ = &config.Config{
		ProfilesDir: profilesDir,
		ServersDir:  serversDir,
	}

	// サーバーマネージャーとプロファイルマネージャーを作成
	serverManager := server.NewManager(serversDir)
	profileManager := profile.NewManager(profilesDir)

	// 1. サーバーテンプレートを作成
	t.Run("サーバーテンプレート作成", func(t *testing.T) {
		err := serverManager.SaveManual("test-server", "python", []string{"-m", "test"}, map[string]string{"TEST_ENV": "value"}, false)
		if err != nil {
			t.Errorf("サーバーテンプレート作成に失敗: %v", err)
		}

		// テンプレートが存在することを確認
		exists, err := serverManager.Exists("test-server")
		if err != nil {
			t.Errorf("存在確認に失敗: %v", err)
		}
		if !exists {
			t.Errorf("サーバーテンプレートが作成されていません")
		}
	})

	// 2. プロファイルを作成
	t.Run("プロファイル作成", func(t *testing.T) {
		err := profileManager.Create("test-profile", "統合テスト用プロファイル")
		if err != nil {
			t.Errorf("プロファイル作成に失敗: %v", err)
		}

		// プロファイルが作成されていることを確認
		profile, err := profileManager.Load("test-profile")
		if err != nil {
			t.Errorf("プロファイル読み込みに失敗: %v", err)
		}
		if profile.Name != "test-profile" {
			t.Errorf("プロファイル名が不正: got %s, want test-profile", profile.Name)
		}
	})

	// 3. プロファイルにサーバーを追加
	t.Run("プロファイルにサーバー追加", func(t *testing.T) {
		err := profileManager.AddServer("test-profile", "test-server", "my-server", map[string]string{"OVERRIDE_ENV": "override_value"})
		if err != nil {
			t.Errorf("サーバー追加に失敗: %v", err)
		}

		// サーバーが追加されていることを確認
		profile, err := profileManager.Load("test-profile")
		if err != nil {
			t.Errorf("プロファイル読み込みに失敗: %v", err)
		}
		if len(profile.Servers) != 1 {
			t.Errorf("サーバー数が不正: got %d, want 1", len(profile.Servers))
		}
		if profile.Servers[0].Name != "my-server" {
			t.Errorf("サーバー名が不正: got %s, want my-server", profile.Servers[0].Name)
		}
	})

	// 4. MCP設定ファイルからプロファイルを作成するテスト
	t.Run("MCP設定ファイルからプロファイル作成", func(t *testing.T) {
		// テスト用のMCP設定ファイルを作成
		mcpConfig := &server.MCPConfig{
			McpServers: map[string]server.MCPServer{
				"file-server": {
					Command: "python",
					Args:    []string{"-m", "file_server"},
					Env:     map[string]string{"FILE_PATH": "/data"},
				},
				"api-server": {
					Command: "node",
					Args:    []string{"api-server.js"},
					Env:     map[string]string{"API_KEY": "secret"},
				},
			},
		}

		mcpPath := filepath.Join(tempDir, "test-mcp-config.json")
		file, err := os.Create(mcpPath)
		if err != nil {
			t.Fatalf("MCP設定ファイル作成に失敗: %v", err)
		}
		defer file.Close()

		encoder := json.NewEncoder(file)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(mcpConfig); err != nil {
			t.Fatalf("MCP設定ファイル書き込みに失敗: %v", err)
		}

		// プロファイルを保存
		err = profileManager.Save("from-mcp", mcpPath, serverManager, false)
		if err != nil {
			t.Errorf("MCP設定からプロファイル保存に失敗: %v", err)
		}

		// プロファイルが正しく作成されていることを確認
		profile, err := profileManager.Load("from-mcp")
		if err != nil {
			t.Errorf("プロファイル読み込みに失敗: %v", err)
		}
		if len(profile.Servers) != 2 {
			t.Errorf("サーバー数が不正: got %d, want 2", len(profile.Servers))
		}
	})

	// 5. プロファイルの適用テスト
	t.Run("プロファイル適用", func(t *testing.T) {
		outputPath := filepath.Join(tempDir, "output-mcp-config.json")

		err := profileManager.Apply("test-profile", outputPath, serverManager)
		if err != nil {
			t.Errorf("プロファイル適用に失敗: %v", err)
		}

		// 出力ファイルが作成されていることを確認
		if _, err := os.Stat(outputPath); os.IsNotExist(err) {
			t.Errorf("出力ファイルが作成されていません: %s", outputPath)
		}

		// 出力ファイルの内容を確認
		var outputConfig server.MCPConfig
		file, err := os.Open(outputPath)
		if err != nil {
			t.Errorf("出力ファイル読み込みに失敗: %v", err)
		}
		defer file.Close()

		decoder := json.NewDecoder(file)
		if err := decoder.Decode(&outputConfig); err != nil {
			t.Errorf("出力ファイルデコードに失敗: %v", err)
		}

		if len(outputConfig.McpServers) != 1 {
			t.Errorf("出力サーバー数が不正: got %d, want 1", len(outputConfig.McpServers))
		}

		server, exists := outputConfig.McpServers["my-server"]
		if !exists {
			t.Errorf("サーバー 'my-server' が出力に含まれていません")
		}

		if server.Command != "python" {
			t.Errorf("サーバーコマンドが不正: got %s, want python", server.Command)
		}

		// 環境変数のオーバーライドが適用されていることを確認
		if server.Env["OVERRIDE_ENV"] != "override_value" {
			t.Errorf("環境変数オーバーライドが適用されていません: got %s, want override_value", server.Env["OVERRIDE_ENV"])
		}
	})

	// 6. サーバーの削除とプロファイルからの削除
	t.Run("サーバー削除", func(t *testing.T) {
		// プロファイルからサーバーを削除
		err := profileManager.RemoveServer("test-profile", "my-server")
		if err != nil {
			t.Errorf("プロファイルからのサーバー削除に失敗: %v", err)
		}

		// サーバーが削除されていることを確認
		profile, err := profileManager.Load("test-profile")
		if err != nil {
			t.Errorf("プロファイル読み込みに失敗: %v", err)
		}
		if len(profile.Servers) != 0 {
			t.Errorf("サーバーが削除されていません: got %d servers, want 0", len(profile.Servers))
		}

		// サーバーテンプレートを削除
		err = serverManager.Delete("test-server", true, nil)
		if err != nil {
			t.Errorf("サーバーテンプレート削除に失敗: %v", err)
		}

		// サーバーテンプレートが削除されていることを確認
		exists, err := serverManager.Exists("test-server")
		if err != nil {
			t.Errorf("存在確認に失敗: %v", err)
		}
		if exists {
			t.Errorf("サーバーテンプレートが削除されていません")
		}
	})

	// 7. プロファイルの削除
	t.Run("プロファイル削除", func(t *testing.T) {
		err := profileManager.Delete("test-profile", true)
		if err != nil {
			t.Errorf("プロファイル削除に失敗: %v", err)
		}

		// プロファイルが削除されていることを確認
		_, err = profileManager.Load("test-profile")
		if err == nil {
			t.Errorf("プロファイルが削除されていません")
		}
	})
}
