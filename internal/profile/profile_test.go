package profile

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

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
				profilePath := filepath.Join(tempDir, tt.profileName+".json")
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

	profilePath := filepath.Join(tempDir, "test-profile.json")
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