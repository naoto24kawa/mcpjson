package profile

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/naoto24kawa/mcpjson/internal/server"
)

const (
	testProfileName      = "test-profile"
	testTemplateName     = "test-template"
	testServerName       = "test-server"
	testDescription      = "テスト用プロファイル"
	oldProfileName       = "old-profile"
	newProfileName       = "new-profile"
	duplicateServerName  = "duplicate-server"
	templateNameForEmpty = "template-name"
)

// createTestMCPConfig creates a test MCP configuration file
func createTestMCPConfig(t *testing.T, filePath string, config *server.MCPConfig) {
	t.Helper()

	// Create directory if it doesn't exist
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("Failed to create directory %s: %v", dir, err)
	}

	file, err := os.Create(filePath)
	if err != nil {
		t.Fatalf("Failed to create test MCP config file: %v", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	if err := encoder.Encode(config); err != nil {
		t.Fatalf("Failed to write test MCP config: %v", err)
	}
}

func TestManager_Create(t *testing.T) {
	tempDir := t.TempDir()
	manager := NewManager(tempDir)

	tests := []struct {
		name        string
		profileName string
		description string
		wantErr     bool
	}{
		{
			name:        "正常なプロファイル作成",
			profileName: "test-profile",
			description: "テスト用プロファイル",
			wantErr:     false,
		},
		{
			name:        "既存プロファイルの重複作成",
			profileName: "test-profile",
			description: "重複テスト",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := manager.Create(tt.profileName, tt.description)
			if (err != nil) != tt.wantErr {
				t.Errorf("Manager.Create() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// プロファイルファイルが作成されているか確認
				profilePath := filepath.Join(tempDir, tt.profileName+".jsonc")
				if _, err := os.Stat(profilePath); os.IsNotExist(err) {
					t.Errorf("プロファイルファイルが作成されていません: %s", profilePath)
				}
			}
		})
	}
}

func TestManager_Load(t *testing.T) {
	tempDir := t.TempDir()
	manager := NewManager(tempDir)

	// テスト用プロファイルを作成
	testProfile := &Profile{
		Name:        "test-profile",
		Description: "テスト用",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Servers:     []ServerRef{},
	}

	profilePath := filepath.Join(tempDir, "test-profile.jsonc")
	file, err := os.Create(profilePath)
	if err != nil {
		t.Fatalf("テストファイル作成に失敗: %v", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	if err := encoder.Encode(testProfile); err != nil {
		t.Fatalf("テストデータ書き込みに失敗: %v", err)
	}

	tests := []struct {
		name        string
		profileName string
		wantErr     bool
	}{
		{
			name:        "既存プロファイルの読み込み",
			profileName: "test-profile",
			wantErr:     false,
		},
		{
			name:        "存在しないプロファイルの読み込み",
			profileName: "nonexistent",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			profile, err := manager.Load(tt.profileName)
			if (err != nil) != tt.wantErr {
				t.Errorf("Manager.Load() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && profile == nil {
				t.Errorf("Manager.Load() = nil, want profile")
			}
		})
	}
}

func TestManager_Delete(t *testing.T) {
	tempDir := t.TempDir()
	manager := NewManager(tempDir)

	// テスト用プロファイルを作成
	err := manager.Create("test-profile", "テスト用")
	if err != nil {
		t.Fatalf("テストプロファイル作成に失敗: %v", err)
	}

	tests := []struct {
		name        string
		profileName string
		force       bool
		wantErr     bool
	}{
		{
			name:        "強制削除",
			profileName: "test-profile",
			force:       true,
			wantErr:     false,
		},
		{
			name:        "存在しないプロファイルの削除",
			profileName: "nonexistent",
			force:       true,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := manager.Delete(tt.profileName, tt.force)
			if (err != nil) != tt.wantErr {
				t.Errorf("Manager.Delete() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestManager_AddServer(t *testing.T) {
	tempDir := t.TempDir()
	manager := NewManager(tempDir)

	// テスト用プロファイルを作成
	err := manager.Create("test-profile", "テスト用")
	if err != nil {
		t.Fatalf("テストプロファイル作成に失敗: %v", err)
	}

	tests := []struct {
		name         string
		profileName  string
		templateName string
		serverName   string
		envOverrides map[string]string
		wantErr      bool
	}{
		{
			name:         "サーバー追加",
			profileName:  "test-profile",
			templateName: "test-template",
			serverName:   "test-server",
			envOverrides: map[string]string{"ENV_VAR": "value"},
			wantErr:      false,
		},
		{
			name:         "存在しないプロファイルへの追加",
			profileName:  "nonexistent",
			templateName: "test-template",
			serverName:   "test-server",
			envOverrides: nil,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := manager.AddServer(tt.profileName, tt.templateName, tt.serverName, tt.envOverrides)
			if (err != nil) != tt.wantErr {
				t.Errorf("Manager.AddServer() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestManager_RemoveServer(t *testing.T) {
	tempDir := t.TempDir()
	manager := NewManager(tempDir)

	// テスト用プロファイルを作成
	err := manager.Create("test-profile", "テスト用")
	if err != nil {
		t.Fatalf("テストプロファイル作成に失敗: %v", err)
	}

	// サーバーを追加
	err = manager.AddServer("test-profile", "test-template", "test-server", nil)
	if err != nil {
		t.Fatalf("テストサーバー追加に失敗: %v", err)
	}

	tests := []struct {
		name        string
		profileName string
		serverName  string
		wantErr     bool
	}{
		{
			name:        "サーバー削除",
			profileName: "test-profile",
			serverName:  "test-server",
			wantErr:     false,
		},
		{
			name:        "存在しないサーバーの削除",
			profileName: "test-profile",
			serverName:  "nonexistent-server",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := manager.RemoveServer(tt.profileName, tt.serverName)
			if (err != nil) != tt.wantErr {
				t.Errorf("Manager.RemoveServer() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// テスト用ヘルパー関数
func createTestProfiles(t *testing.T, manager *Manager, profileNames []string) {
	t.Helper()
	for _, name := range profileNames {
		err := manager.Create(name, testDescription)
		if err != nil {
			t.Fatalf("テストプロファイル '%s' の作成に失敗: %v", name, err)
		}
	}
}

func verifyProfilesDeleted(t *testing.T, tempDir string, profileNames []string) {
	t.Helper()
	for _, name := range profileNames {
		profilePath := filepath.Join(tempDir, name+".jsonc")
		if _, err := os.Stat(profilePath); !os.IsNotExist(err) {
			t.Errorf("プロファイルファイル '%s' が削除されていません", profilePath)
		}
	}
}

func TestManager_Reset_EmptyDirectory(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	manager := NewManager(tempDir)

	// Act
	err := manager.Reset(true)

	// Assert
	if err != nil {
		t.Errorf("Manager.Reset() on empty directory failed: %v", err)
	}
}

func TestManager_Reset_MultipleProfiles(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	manager := NewManager(tempDir)
	profileNames := []string{"profile1", "profile2", "profile3"}
	createTestProfiles(t, manager, profileNames)

	// Act
	err := manager.Reset(true)

	// Assert
	if err != nil {
		t.Errorf("Manager.Reset() with multiple profiles failed: %v", err)
	}
	verifyProfilesDeleted(t, tempDir, profileNames)
}

func TestManager_Reset_SingleProfile(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	manager := NewManager(tempDir)
	profileNames := []string{"single-profile"}
	createTestProfiles(t, manager, profileNames)

	// Act
	err := manager.Reset(true)

	// Assert
	if err != nil {
		t.Errorf("Manager.Reset() with single profile failed: %v", err)
	}
	verifyProfilesDeleted(t, tempDir, profileNames)
}

func TestManager_Reset_NonexistentDirectory(t *testing.T) {
	// Arrange
	tempDir := filepath.Join(os.TempDir(), "nonexistent-dir")
	manager := NewManager(tempDir)

	// Act
	err := manager.Reset(true)

	// Assert
	if err != nil {
		t.Errorf("Manager.Reset() with nonexistent directory failed: %v", err)
	}
}

// リネーム検証用ヘルパー関数
func verifyRenameSuccess(t *testing.T, manager *Manager, tempDir, oldName, newName string) {
	t.Helper()

	// 新しい名前のファイルが存在することを確認
	newPath := filepath.Join(tempDir, newName+".jsonc")
	if _, err := os.Stat(newPath); os.IsNotExist(err) {
		t.Errorf("リネーム後のプロファイルファイル '%s' が存在しません", newPath)
	}

	// 古い名前のファイルが削除されていることを確認
	oldPath := filepath.Join(tempDir, oldName+".jsonc")
	if _, err := os.Stat(oldPath); !os.IsNotExist(err) {
		t.Errorf("リネーム前のプロファイルファイル '%s' が削除されていません", oldPath)
	}

	// プロファイルの内容が正しく更新されているかを確認
	profile, err := manager.Load(newName)
	if err != nil {
		t.Fatalf("リネーム後のプロファイルの読み込みに失敗: %v", err)
	}
	if profile.Name != newName {
		t.Errorf("プロファイル名が更新されていません: got %s, want %s", profile.Name, newName)
	}
}

func TestManager_Rename_Success(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	manager := NewManager(tempDir)
	err := manager.Create(oldProfileName, "古いプロファイル")
	if err != nil {
		t.Fatalf("テストプロファイル作成に失敗: %v", err)
	}

	// Act
	err = manager.Rename(oldProfileName, newProfileName, false)

	// Assert
	if err != nil {
		t.Errorf("Manager.Rename() failed: %v", err)
	}
	verifyRenameSuccess(t, manager, tempDir, oldProfileName, newProfileName)
}

func TestManager_Rename_NonexistentProfile(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	manager := NewManager(tempDir)

	// Act
	err := manager.Rename("nonexistent-profile", newProfileName, false)

	// Assert
	if err == nil {
		t.Error("Manager.Rename() expected error for nonexistent profile, got nil")
	}
}

func TestManager_Rename_ForceOverwrite(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	manager := NewManager(tempDir)

	// 古いプロファイルと既存の新しいプロファイルを作成
	err := manager.Create(oldProfileName, "古いプロファイル")
	if err != nil {
		t.Fatalf("古いプロファイル作成に失敗: %v", err)
	}
	err = manager.Create("existing-new", "既存の新しいプロファイル")
	if err != nil {
		t.Fatalf("既存プロファイル作成に失敗: %v", err)
	}

	// Act
	err = manager.Rename(oldProfileName, "existing-new", true)

	// Assert
	if err != nil {
		t.Errorf("Manager.Rename() with force failed: %v", err)
	}
	verifyRenameSuccess(t, manager, tempDir, oldProfileName, "existing-new")
}

func TestManager_Rename_ConflictWithoutForce(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	manager := NewManager(tempDir)

	// 古いプロファイルと既存の新しいプロファイルを作成
	err := manager.Create(oldProfileName, "古いプロファイル")
	if err != nil {
		t.Fatalf("古いプロファイル作成に失敗: %v", err)
	}
	err = manager.Create("existing-new", "既存の新しいプロファイル")
	if err != nil {
		t.Fatalf("既存プロファイル作成に失敗: %v", err)
	}

	// Act
	err = manager.Rename(oldProfileName, "existing-new", false)

	// Assert
	if err == nil {
		t.Error("Manager.Rename() expected error for existing profile without force, got nil")
	}
}

func TestManager_List(t *testing.T) {
	// Arrange
	tests := []struct {
		name          string
		setupProfiles []string
		detail        bool
		wantErr       bool
	}{
		{
			name:          "空ディレクトリの一覧表示",
			setupProfiles: []string{},
			detail:        false,
			wantErr:       false,
		},
		{
			name:          "複数プロファイルの簡易一覧表示",
			setupProfiles: []string{"profile1", "profile2"},
			detail:        false,
			wantErr:       false,
		},
		{
			name:          "単一プロファイルの詳細一覧表示",
			setupProfiles: []string{"detailed-profile"},
			detail:        true,
			wantErr:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			tempDir := t.TempDir()
			manager := NewManager(tempDir)

			// テスト用プロファイルを作成
			for _, profileName := range tt.setupProfiles {
				err := manager.Create(profileName, "テスト用プロファイル")
				if err != nil {
					t.Fatalf("テストプロファイル '%s' の作成に失敗: %v", profileName, err)
				}
			}

			// Act
			err := manager.List(tt.detail)

			// Assert
			if (err != nil) != tt.wantErr {
				t.Errorf("Manager.List() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestManager_List_NonexistentDirectory(t *testing.T) {
	// Arrange: 存在しないディレクトリ
	manager := NewManager("/nonexistent/directory")

	// Act
	err := manager.List(false)

	// Assert
	if err == nil {
		t.Error("Manager.List() expected error for nonexistent directory, got nil")
	}
}

func TestManager_AddServer_Duplicate(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	manager := NewManager(tempDir)

	// テスト用プロファイルを作成
	err := manager.Create("test-profile", "テスト用")
	if err != nil {
		t.Fatalf("テストプロファイル作成に失敗: %v", err)
	}

	// 最初のサーバーを追加
	err = manager.AddServer("test-profile", "test-template", "duplicate-server", nil)
	if err != nil {
		t.Fatalf("初回サーバー追加に失敗: %v", err)
	}

	// Act: 同じ名前のサーバーを再度追加
	err = manager.AddServer("test-profile", "test-template", "duplicate-server", nil)

	// Assert
	if err == nil {
		t.Error("Manager.AddServer() expected error for duplicate server name, got nil")
	}
}

func TestManager_AddServer_EmptyName(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	manager := NewManager(tempDir)

	// テスト用プロファイルを作成
	err := manager.Create("test-profile", "テスト用")
	if err != nil {
		t.Fatalf("テストプロファイル作成に失敗: %v", err)
	}

	// Act: 空のサーバー名でサーバーを追加（テンプレート名が使用されるべき）
	err = manager.AddServer("test-profile", "template-name", "", nil)

	// Assert
	if err != nil {
		t.Errorf("Manager.AddServer() with empty server name failed: %v", err)
	}

	// サーバーがテンプレート名で追加されているか確認
	profile, err := manager.Load("test-profile")
	if err != nil {
		t.Fatalf("プロファイル読み込みに失敗: %v", err)
	}

	if len(profile.Servers) != 1 || profile.Servers[0].Name != "template-name" {
		t.Errorf("サーバー名がテンプレート名に設定されていません: got %v", profile.Servers)
	}
}

func TestManager_Save_BasicFlow(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	profilesDir := filepath.Join(tempDir, "profiles")
	serversDir := filepath.Join(tempDir, "servers")

	// ディレクトリを作成
	if err := os.MkdirAll(profilesDir, 0755); err != nil {
		t.Fatalf("Failed to create profiles directory: %v", err)
	}
	if err := os.MkdirAll(serversDir, 0755); err != nil {
		t.Fatalf("Failed to create servers directory: %v", err)
	}

	manager := NewManager(profilesDir)
	serverManager := server.NewManager(serversDir)

	mcpConfigPath := filepath.Join(tempDir, "mcp_config.json")
	testConfig := &server.MCPConfig{
		McpServers: map[string]server.MCPServer{
			"test-server": {
				Command: "python",
				Args:    []string{"test.py"},
				Env:     map[string]string{"ENV_VAR": "test_value"},
			},
		},
	}
	createTestMCPConfig(t, mcpConfigPath, testConfig)

	// Act
	err := manager.Save("test-profile", mcpConfigPath, serverManager, false)

	// Assert
	if err != nil {
		t.Errorf("Manager.Save() failed: %v", err)
	}

	// プロファイルが正しく保存されているか確認
	profile, err := manager.Load("test-profile")
	if err != nil {
		t.Fatalf("Failed to load saved profile: %v", err)
	}

	if profile.Name != "test-profile" {
		t.Errorf("Profile name mismatch: got %s, want test-profile", profile.Name)
	}

	if len(profile.Servers) != 1 {
		t.Errorf("Expected 1 server, got %d", len(profile.Servers))
	}

	if profile.Servers[0].Name != "test-server" {
		t.Errorf("Server name mismatch: got %s, want test-server", profile.Servers[0].Name)
	}

	// サーバーテンプレートが作成されているかも確認
	exists, err := serverManager.Exists("test-server")
	if err != nil {
		t.Fatalf("Failed to check server template existence: %v", err)
	}
	if !exists {
		t.Error("Server template was not created")
	}
}

func TestManager_Save_WithForceOverwrite(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	profilesDir := filepath.Join(tempDir, "profiles")
	serversDir := filepath.Join(tempDir, "servers")

	if err := os.MkdirAll(profilesDir, 0755); err != nil {
		t.Fatalf("Failed to create profiles directory: %v", err)
	}
	if err := os.MkdirAll(serversDir, 0755); err != nil {
		t.Fatalf("Failed to create servers directory: %v", err)
	}

	manager := NewManager(profilesDir)
	serverManager := server.NewManager(serversDir)

	// 既存のプロファイルを作成
	err := manager.Create("existing-profile", "既存のプロファイル")
	if err != nil {
		t.Fatalf("Failed to create existing profile: %v", err)
	}

	mcpConfigPath := filepath.Join(tempDir, "mcp_config.json")
	testConfig := &server.MCPConfig{
		McpServers: map[string]server.MCPServer{
			"new-server": {
				Command: "node",
				Args:    []string{"server.js"},
			},
		},
	}
	createTestMCPConfig(t, mcpConfigPath, testConfig)

	// Act
	err = manager.Save("existing-profile", mcpConfigPath, serverManager, true)

	// Assert
	if err != nil {
		t.Errorf("Manager.Save() with force failed: %v", err)
	}

	// プロファイルが上書きされているか確認
	profile, err := manager.Load("existing-profile")
	if err != nil {
		t.Fatalf("Failed to load overwritten profile: %v", err)
	}

	if len(profile.Servers) != 1 || profile.Servers[0].Name != "new-server" {
		t.Errorf("Profile was not overwritten correctly: %v", profile.Servers)
	}
}

func TestManager_Save_MCPConfigLoadError(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	profilesDir := filepath.Join(tempDir, "profiles")
	serversDir := filepath.Join(tempDir, "servers")

	if err := os.MkdirAll(profilesDir, 0755); err != nil {
		t.Fatalf("Failed to create profiles directory: %v", err)
	}
	if err := os.MkdirAll(serversDir, 0755); err != nil {
		t.Fatalf("Failed to create servers directory: %v", err)
	}

	manager := NewManager(profilesDir)
	serverManager := server.NewManager(serversDir)

	nonexistentPath := filepath.Join(tempDir, "nonexistent.json")

	// Act
	err := manager.Save("test-profile", nonexistentPath, serverManager, false)

	// Assert
	if err == nil {
		t.Error("Manager.Save() expected error for nonexistent MCP config, got nil")
	}
}

func TestManager_Save_ProfileExistsWithoutForce(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	profilesDir := filepath.Join(tempDir, "profiles")
	serversDir := filepath.Join(tempDir, "servers")

	if err := os.MkdirAll(profilesDir, 0755); err != nil {
		t.Fatalf("Failed to create profiles directory: %v", err)
	}
	if err := os.MkdirAll(serversDir, 0755); err != nil {
		t.Fatalf("Failed to create servers directory: %v", err)
	}

	manager := NewManager(profilesDir)
	serverManager := server.NewManager(serversDir)

	// 既存のプロファイルを作成
	err := manager.Create("existing-profile", "既存のプロファイル")
	if err != nil {
		t.Fatalf("Failed to create existing profile: %v", err)
	}

	mcpConfigPath := filepath.Join(tempDir, "mcp_config.json")
	testConfig := &server.MCPConfig{
		McpServers: map[string]server.MCPServer{},
	}
	createTestMCPConfig(t, mcpConfigPath, testConfig)

	// Act
	err = manager.Save("existing-profile", mcpConfigPath, serverManager, false)

	// Assert
	if err == nil {
		t.Error("Manager.Save() expected error for existing profile without force, got nil")
	}
}

func TestManager_Apply_BasicFlow(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	profilesDir := filepath.Join(tempDir, "profiles")
	serversDir := filepath.Join(tempDir, "servers")

	if err := os.MkdirAll(profilesDir, 0755); err != nil {
		t.Fatalf("Failed to create profiles directory: %v", err)
	}
	if err := os.MkdirAll(serversDir, 0755); err != nil {
		t.Fatalf("Failed to create servers directory: %v", err)
	}

	manager := NewManager(profilesDir)
	serverManager := server.NewManager(serversDir)

	// サーバーテンプレートを先に作成
	err := serverManager.SaveFromConfig("test-template", server.MCPServer{
		Command: "python",
		Args:    []string{"test.py"},
		Env:     map[string]string{"TEST_ENV": "value"},
	})
	if err != nil {
		t.Fatalf("Failed to create server template: %v", err)
	}

	// テスト用プロファイルを作成
	profile := &Profile{
		Name:        "test-profile",
		Description: "テスト用プロファイル",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Servers: []ServerRef{
			{
				Name:     "test-server",
				Template: "test-template",
			},
		},
	}

	// プロファイルを保存
	if err := manager.saveProfile(profile); err != nil {
		t.Fatalf("Failed to save test profile: %v", err)
	}

	targetPath := filepath.Join(tempDir, "output_config.json")

	// Act
	err = manager.Apply("test-profile", targetPath, serverManager)

	// Assert
	if err != nil {
		t.Errorf("Manager.Apply() failed: %v", err)
	}

	// 出力されたMCP設定ファイルを確認
	if _, err := os.Stat(targetPath); os.IsNotExist(err) {
		t.Errorf("Output MCP config file was not created: %s", targetPath)
	}

	// ファイルの内容を確認
	var mcpConfig server.MCPConfig
	file, err := os.Open(targetPath)
	if err != nil {
		t.Fatalf("Failed to open output file: %v", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&mcpConfig); err != nil {
		t.Fatalf("Failed to decode output file: %v", err)
	}

	// MCPConfigの内容を確認
	if len(mcpConfig.McpServers) != 1 {
		t.Errorf("Expected 1 server in MCP config, got %d", len(mcpConfig.McpServers))
	}

	testServer, exists := mcpConfig.McpServers["test-server"]
	if !exists {
		t.Error("test-server not found in MCP config")
	} else {
		if testServer.Command != "python" {
			t.Errorf("Expected command 'python', got '%s'", testServer.Command)
		}
		if len(testServer.Args) != 1 || testServer.Args[0] != "test.py" {
			t.Errorf("Expected args ['test.py'], got %v", testServer.Args)
		}
		if testServer.Env["TEST_ENV"] != "value" {
			t.Errorf("Expected env TEST_ENV='value', got %v", testServer.Env)
		}
	}
}

func TestManager_Apply_ProfileNotFound(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	profilesDir := filepath.Join(tempDir, "profiles")
	serversDir := filepath.Join(tempDir, "servers")

	if err := os.MkdirAll(profilesDir, 0755); err != nil {
		t.Fatalf("Failed to create profiles directory: %v", err)
	}
	if err := os.MkdirAll(serversDir, 0755); err != nil {
		t.Fatalf("Failed to create servers directory: %v", err)
	}

	manager := NewManager(profilesDir)
	serverManager := server.NewManager(serversDir)

	targetPath := filepath.Join(tempDir, "output_config.json")

	// Act
	err := manager.Apply("nonexistent-profile", targetPath, serverManager)

	// Assert
	if err == nil {
		t.Error("Manager.Apply() expected error for nonexistent profile, got nil")
	}
}

func TestManager_Copy(t *testing.T) {
	tests := []struct {
		name          string
		sourceName    string
		destName      string
		force         bool
		setupSource   bool
		setupDest     bool
		expectError   bool
		errorContains string
	}{
		{
			name:        "正常なプロファイルコピー",
			sourceName:  "source-profile",
			destName:    "dest-profile",
			force:       false,
			setupSource: true,
			setupDest:   false,
			expectError: false,
		},
		{
			name:          "存在しないソースプロファイル",
			sourceName:    "nonexistent",
			destName:      "dest-profile",
			force:         false,
			setupSource:   false,
			setupDest:     false,
			expectError:   true,
			errorContains: "が見つかりません",
		},
		{
			name:          "既存の宛先プロファイル（forceなし）",
			sourceName:    "source-profile",
			destName:      "existing-dest",
			force:         false,
			setupSource:   true,
			setupDest:     true,
			expectError:   true,
			errorContains: "は既に存在します",
		},
		{
			name:        "既存の宛先プロファイル（forceあり）",
			sourceName:  "source-profile",
			destName:    "existing-dest",
			force:       true,
			setupSource: true,
			setupDest:   true,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			tempDir := t.TempDir()
			manager := NewManager(tempDir)

			// Setup source profile if needed
			if tt.setupSource {
				sourceProfile := &Profile{
					Name:        tt.sourceName,
					Description: "Test source profile",
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
					Servers: []ServerRef{
						{Name: "test-server", Template: "test-template"},
					},
				}
				sourcePath := filepath.Join(tempDir, tt.sourceName+".jsonc")
				if err := createTestProfile(t, sourcePath, sourceProfile); err != nil {
					t.Fatalf("Failed to create source profile: %v", err)
				}
			}

			// Setup destination profile if needed
			if tt.setupDest {
				destProfile := &Profile{
					Name:        tt.destName,
					Description: "Test dest profile",
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
					Servers:     []ServerRef{},
				}
				destPath := filepath.Join(tempDir, tt.destName+".jsonc")
				if err := createTestProfile(t, destPath, destProfile); err != nil {
					t.Fatalf("Failed to create dest profile: %v", err)
				}
			}

			// Act
			err := manager.Copy(tt.sourceName, tt.destName, tt.force)

			// Assert
			if tt.expectError {
				if err == nil {
					t.Error("Manager.Copy() expected error, got nil")
					return
				}
				if tt.errorContains != "" && len(tt.errorContains) > 0 {
					if !containsString(err.Error(), tt.errorContains) {
						t.Errorf("Manager.Copy() error = %q, expected to contain %q", err.Error(), tt.errorContains)
					}
				}
			} else {
				if err != nil {
					t.Errorf("Manager.Copy() unexpected error = %v", err)
					return
				}

				// Verify destination profile was created
				destPath := filepath.Join(tempDir, tt.destName+".jsonc")
				if _, err := os.Stat(destPath); os.IsNotExist(err) {
					t.Error("Destination profile was not created")
				}

				// Verify source profile still exists
				sourcePath := filepath.Join(tempDir, tt.sourceName+".jsonc")
				if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
					t.Error("Source profile should still exist after copy")
				}

				// Verify content was copied correctly
				destProfile, err := manager.Load(tt.destName)
				if err != nil {
					t.Fatalf("Failed to load copied profile: %v", err)
				}

				if destProfile.Name != tt.destName {
					t.Errorf("Copied profile name = %q, expected %q", destProfile.Name, tt.destName)
				}

				// Verify timestamps were updated
				if destProfile.CreatedAt.IsZero() {
					t.Error("Copied profile CreatedAt should be set")
				}
				if destProfile.UpdatedAt.IsZero() {
					t.Error("Copied profile UpdatedAt should be set")
				}
			}
		})
	}
}

func TestManager_Merge(t *testing.T) {
	tests := []struct {
		name            string
		destName        string
		sourceNames     []string
		force           bool
		setupSources    []bool
		setupDest       bool
		expectError     bool
		errorContains   string
		expectedServers int
	}{
		{
			name:            "2つのプロファイルのマージ",
			destName:        "merged-profile",
			sourceNames:     []string{"source1", "source2"},
			force:           false,
			setupSources:    []bool{true, true},
			setupDest:       false,
			expectError:     false,
			expectedServers: 4, // source1-server1, source1-server2, source2-server1, source2-server2
		},
		{
			name:            "単一プロファイルのマージ",
			destName:        "merged-profile",
			sourceNames:     []string{"source1"},
			force:           false,
			setupSources:    []bool{true},
			setupDest:       false,
			expectError:     false,
			expectedServers: 2, // source1-server1, source1-server2
		},
		{
			name:          "存在しないソースプロファイル",
			destName:      "merged-profile",
			sourceNames:   []string{"nonexistent"},
			force:         false,
			setupSources:  []bool{false},
			setupDest:     false,
			expectError:   true,
			errorContains: "の読み込みに失敗しました",
		},
		{
			name:          "既存の宛先プロファイル（forceなし）",
			destName:      "existing-dest",
			sourceNames:   []string{"source1"},
			force:         false,
			setupSources:  []bool{true},
			setupDest:     true,
			expectError:   true,
			errorContains: "は既に存在します",
		},
		{
			name:            "既存の宛先プロファイル（forceあり）",
			destName:        "existing-dest",
			sourceNames:     []string{"source1"},
			force:           true,
			setupSources:    []bool{true},
			setupDest:       true,
			expectError:     false,
			expectedServers: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			tempDir := t.TempDir()
			manager := NewManager(tempDir)

			// Setup source profiles
			for i, sourceName := range tt.sourceNames {
				if i < len(tt.setupSources) && tt.setupSources[i] {
					sourceProfile := &Profile{
						Name:        sourceName,
						Description: "Test source profile " + sourceName,
						CreatedAt:   time.Now(),
						UpdatedAt:   time.Now(),
						Servers: []ServerRef{
							{Name: sourceName + "-server1", Template: "template1"}, // unique server per source
							{Name: sourceName + "-server2", Template: "template2"}, // unique server per source
						},
					}
					sourcePath := filepath.Join(tempDir, sourceName+".jsonc")
					if err := createTestProfile(t, sourcePath, sourceProfile); err != nil {
						t.Fatalf("Failed to create source profile %s: %v", sourceName, err)
					}
				}
			}

			// Setup destination profile if needed
			if tt.setupDest {
				destProfile := &Profile{
					Name:        tt.destName,
					Description: "Test dest profile",
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
					Servers:     []ServerRef{},
				}
				destPath := filepath.Join(tempDir, tt.destName+".jsonc")
				if err := createTestProfile(t, destPath, destProfile); err != nil {
					t.Fatalf("Failed to create dest profile: %v", err)
				}
			}

			// Act
			err := manager.Merge(tt.destName, tt.sourceNames, tt.force)

			// Assert
			if tt.expectError {
				if err == nil {
					t.Error("Manager.Merge() expected error, got nil")
					return
				}
				if tt.errorContains != "" && len(tt.errorContains) > 0 {
					if !containsString(err.Error(), tt.errorContains) {
						t.Errorf("Manager.Merge() error = %q, expected to contain %q", err.Error(), tt.errorContains)
					}
				}
			} else {
				if err != nil {
					t.Errorf("Manager.Merge() unexpected error = %v", err)
					return
				}

				// Verify destination profile was created
				destPath := filepath.Join(tempDir, tt.destName+".jsonc")
				if _, err := os.Stat(destPath); os.IsNotExist(err) {
					t.Error("Destination profile was not created")
				}

				// Verify merged profile content
				mergedProfile, err := manager.Load(tt.destName)
				if err != nil {
					t.Fatalf("Failed to load merged profile: %v", err)
				}

				if mergedProfile.Name != tt.destName {
					t.Errorf("Merged profile name = %q, expected %q", mergedProfile.Name, tt.destName)
				}

				if len(mergedProfile.Servers) != tt.expectedServers {
					t.Errorf("Merged profile servers count = %d, expected %d", len(mergedProfile.Servers), tt.expectedServers)
				}

				// Verify timestamps were set
				if mergedProfile.CreatedAt.IsZero() {
					t.Error("Merged profile CreatedAt should be set")
				}
				if mergedProfile.UpdatedAt.IsZero() {
					t.Error("Merged profile UpdatedAt should be set")
				}
			}
		})
	}
}

func TestManager_Merge_DuplicateServers(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	manager := NewManager(tempDir)

	// Create profiles with overlapping server names
	profile1 := &Profile{
		Name:        "profile1",
		Description: "Profile with common server",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Servers: []ServerRef{
			{Name: "common-server", Template: "template1"},
			{Name: "unique-server1", Template: "template1"},
		},
	}
	profile1Path := filepath.Join(tempDir, "profile1.jsonc")
	if err := createTestProfile(t, profile1Path, profile1); err != nil {
		t.Fatalf("Failed to create profile1: %v", err)
	}

	profile2 := &Profile{
		Name:        "profile2",
		Description: "Profile with same common server",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Servers: []ServerRef{
			{Name: "common-server", Template: "template2"}, // Same name, different template
			{Name: "unique-server2", Template: "template2"},
		},
	}
	profile2Path := filepath.Join(tempDir, "profile2.jsonc")
	if err := createTestProfile(t, profile2Path, profile2); err != nil {
		t.Fatalf("Failed to create profile2: %v", err)
	}

	// Act
	err := manager.Merge("merged", []string{"profile1", "profile2"}, false)

	// Assert
	if err != nil {
		t.Errorf("Manager.Merge() unexpected error = %v", err)
		return
	}

	mergedProfile, err := manager.Load("merged")
	if err != nil {
		t.Fatalf("Failed to load merged profile: %v", err)
	}

	// Should have 3 servers: common-server (from profile1), unique-server1, unique-server2
	expectedServers := 3
	if len(mergedProfile.Servers) != expectedServers {
		t.Errorf("Merged profile servers count = %d, expected %d", len(mergedProfile.Servers), expectedServers)
	}

	// Verify first-wins policy for common-server
	var commonServer *ServerRef
	for _, server := range mergedProfile.Servers {
		if server.Name == "common-server" {
			commonServer = &server
			break
		}
	}

	if commonServer == nil {
		t.Error("common-server not found in merged profile")
	} else if commonServer.Template != "template1" {
		t.Errorf("Expected common-server to have template1 (first-wins), got %s", commonServer.Template)
	}
}

// Helper function to create a test profile file
func createTestProfile(t *testing.T, path string, profile *Profile) error {
	t.Helper()

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(profile)
}

// Helper function to check if a string contains a substring
func containsString(s, substr string) bool {
	return len(substr) == 0 || (len(s) >= len(substr) && findInString(s, substr))
}

func findInString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
