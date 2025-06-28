package profile

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/naoto24kawa/mcpconfig/internal/server"
)

const (
	testProfileName    = "test-profile"
	testTemplateName   = "test-template"
	testServerName     = "test-server"
	testDescription    = "テスト用プロファイル"
	oldProfileName     = "old-profile"
	newProfileName     = "new-profile"
	duplicateServerName = "duplicate-server"
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