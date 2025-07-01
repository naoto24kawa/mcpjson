package profile

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/naoto24kawa/mcpjson/internal/config"
)

func TestCreate(t *testing.T) {
	tempDir := t.TempDir()
	cfg := &config.Config{
		ProfilesDir: tempDir,
		ServersDir:  tempDir,
	}

	tests := []struct {
		name         string
		profileName  string
		templateName string
		wantErr      bool
	}{
		{
			name:         "正常なプロファイル作成",
			profileName:  "test-profile",
			templateName: "",
			wantErr:      false,
		},
		{
			name:         "テンプレート指定でのプロファイル作成",
			profileName:  "test-profile-with-template",
			templateName: "test-template",
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Create(cfg, tt.profileName, tt.templateName)
			if (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
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

func TestList(t *testing.T) {
	tempDir := t.TempDir()
	cfg := &config.Config{
		ProfilesDir: tempDir,
		ServersDir:  tempDir,
	}

	// テスト用プロファイルを作成
	err := Create(cfg, "test-profile", "")
	if err != nil {
		t.Fatalf("テストプロファイル作成に失敗: %v", err)
	}

	tests := []struct {
		name    string
		detail  bool
		wantErr bool
	}{
		{
			name:    "簡単な一覧表示",
			detail:  false,
			wantErr: false,
		},
		{
			name:    "詳細な一覧表示",
			detail:  true,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := List(cfg, tt.detail)
			if (err != nil) != tt.wantErr {
				t.Errorf("List() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDelete(t *testing.T) {
	tempDir := t.TempDir()
	cfg := &config.Config{
		ProfilesDir: tempDir,
		ServersDir:  tempDir,
	}

	// テスト用プロファイルを作成
	err := Create(cfg, "test-profile", "")
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
			err := Delete(cfg, tt.profileName, tt.force)
			if (err != nil) != tt.wantErr {
				t.Errorf("Delete() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRename(t *testing.T) {
	tempDir := t.TempDir()
	cfg := &config.Config{
		ProfilesDir: tempDir,
		ServersDir:  tempDir,
	}

	// テスト用プロファイルを作成
	err := Create(cfg, "old-profile", "")
	if err != nil {
		t.Fatalf("テストプロファイル作成に失敗: %v", err)
	}

	tests := []struct {
		name    string
		oldName string
		newName string
		force   bool
		wantErr bool
	}{
		{
			name:    "正常なリネーム",
			oldName: "old-profile",
			newName: "new-profile",
			force:   false,
			wantErr: false,
		},
		{
			name:    "存在しないプロファイルのリネーム",
			oldName: "nonexistent",
			newName: "new-name",
			force:   false,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Rename(cfg, tt.oldName, tt.newName, tt.force)
			if (err != nil) != tt.wantErr {
				t.Errorf("Rename() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
