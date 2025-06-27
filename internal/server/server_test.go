package server

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestManager_SaveManual(t *testing.T) {
	tempDir := t.TempDir()
	manager := NewManager(tempDir)

	tests := []struct {
		name         string
		templateName string
		command      string
		args         []string
		env          map[string]string
		force        bool
		wantErr      bool
	}{
		{
			name:         "新規テンプレート作成",
			templateName: "test-template",
			command:      "test-command",
			args:         []string{"arg1", "arg2"},
			env:          map[string]string{"ENV1": "value1"},
			force:        false,
			wantErr:      false,
		},
		{
			name:         "コマンドなしでの作成",
			templateName: "empty-command",
			command:      "",
			args:         nil,
			env:          nil,
			force:        false,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := manager.SaveManual(tt.templateName, tt.command, tt.args, tt.env, tt.force)
			if (err != nil) != tt.wantErr {
				t.Errorf("Manager.SaveManual() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				// テンプレートファイルが作成されているか確認
				templatePath := filepath.Join(tempDir, tt.templateName+".json")
				if _, err := os.Stat(templatePath); os.IsNotExist(err) {
					t.Errorf("テンプレートファイルが作成されていません: %s", templatePath)
				}
			}
		})
	}
}

func TestManager_Load(t *testing.T) {
	tempDir := t.TempDir()
	manager := NewManager(tempDir)

	// テスト用テンプレートを作成
	testTemplate := &ServerTemplate{
		Name:        "test-template",
		Description: nil,
		CreatedAt:   time.Now(),
		ServerConfig: ServerConfig{
			Command: "test-command",
			Args:    []string{"arg1"},
			Env:     map[string]string{"ENV1": "value1"},
		},
	}

	templatePath := filepath.Join(tempDir, "test-template.json")
	file, err := os.Create(templatePath)
	if err != nil {
		t.Fatalf("テストファイル作成に失敗: %v", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(testTemplate); err != nil {
		t.Fatalf("テストデータ書き込みに失敗: %v", err)
	}

	tests := []struct {
		name         string
		templateName string
		wantErr      bool
	}{
		{
			name:         "既存テンプレートの読み込み",
			templateName: "test-template",
			wantErr:      false,
		},
		{
			name:         "存在しないテンプレートの読み込み",
			templateName: "nonexistent",
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			template, err := manager.Load(tt.templateName)
			if (err != nil) != tt.wantErr {
				t.Errorf("Manager.Load() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && template == nil {
				t.Errorf("Manager.Load() = nil, want template")
			}
		})
	}
}

func TestManager_Delete(t *testing.T) {
	tempDir := t.TempDir()
	manager := NewManager(tempDir)

	// テスト用テンプレートを作成
	err := manager.SaveManual("test-template", "test-command", nil, nil, false)
	if err != nil {
		t.Fatalf("テストテンプレート作成に失敗: %v", err)
	}

	tests := []struct {
		name         string
		templateName string
		force        bool
		wantErr      bool
	}{
		{
			name:         "強制削除",
			templateName: "test-template",
			force:        true,
			wantErr:      false,
		},
		{
			name:         "存在しないテンプレートの削除",
			templateName: "nonexistent",
			force:        true,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := manager.Delete(tt.templateName, tt.force)
			if (err != nil) != tt.wantErr {
				t.Errorf("Manager.Delete() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestManager_Exists(t *testing.T) {
	tempDir := t.TempDir()
	manager := NewManager(tempDir)

	// テスト用テンプレートを作成
	err := manager.SaveManual("test-template", "test-command", nil, nil, false)
	if err != nil {
		t.Fatalf("テストテンプレート作成に失敗: %v", err)
	}

	tests := []struct {
		name         string
		templateName string
		want         bool
	}{
		{
			name:         "既存テンプレートの存在確認",
			templateName: "test-template",
			want:         true,
		},
		{
			name:         "存在しないテンプレートの存在確認",
			templateName: "nonexistent",
			want:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := manager.Exists(tt.templateName)
			if err != nil {
				t.Errorf("Manager.Exists() error = %v", err)
				return
			}
			if got != tt.want {
				t.Errorf("Manager.Exists() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestManager_updateTemplateArgs(t *testing.T) {
	manager := &Manager{}
	template := &ServerTemplate{
		ServerConfig: ServerConfig{
			Args: []string{"old-arg"},
		},
	}

	tests := []struct {
		name     string
		args     []string
		expected []string
	}{
		{
			name:     "引数の更新",
			args:     []string{"new-arg1", "new-arg2"},
			expected: []string{"new-arg1", "new-arg2"},
		},
		{
			name:     "空文字列での引数クリア",
			args:     []string{""},
			expected: nil,
		},
		{
			name:     "nilでの引数変更なし",
			args:     nil,
			expected: []string{"old-arg"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// テンプレートを元の状態にリセット
			template.ServerConfig.Args = []string{"old-arg"}
			
			// TemplateUpdaterを使用してテスト
			updater := NewTemplateUpdater(manager.templateManager)
			updater.UpdateTemplateArgs(template, tt.args)
			
			if len(template.ServerConfig.Args) != len(tt.expected) {
				t.Errorf("引数の長さが期待値と異なります: got %d, want %d", len(template.ServerConfig.Args), len(tt.expected))
				return
			}
			
			for i, arg := range template.ServerConfig.Args {
				if arg != tt.expected[i] {
					t.Errorf("引数[%d] = %v, want %v", i, arg, tt.expected[i])
				}
			}
		})
	}
}

func TestManager_updateTemplateEnv(t *testing.T) {
	manager := &Manager{}
	template := &ServerTemplate{
		ServerConfig: ServerConfig{
			Env: map[string]string{"OLD_ENV": "old_value"},
		},
	}

	tests := []struct {
		name     string
		env      map[string]string
		expected map[string]string
	}{
		{
			name:     "環境変数の追加",
			env:      map[string]string{"NEW_ENV": "new_value"},
			expected: map[string]string{"OLD_ENV": "old_value", "NEW_ENV": "new_value"},
		},
		{
			name:     "環境変数の削除（空文字）",
			env:      map[string]string{"OLD_ENV": ""},
			expected: map[string]string{},
		},
		{
			name:     "環境変数のクリア",
			env:      map[string]string{},
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// テンプレートを元の状態にリセット
			template.ServerConfig.Env = map[string]string{"OLD_ENV": "old_value"}
			
			// TemplateUpdaterを使用してテスト
			updater := NewTemplateUpdater(manager.templateManager)
			updater.UpdateTemplateEnv(template, tt.env)
			
			if len(template.ServerConfig.Env) != len(tt.expected) {
				t.Errorf("環境変数の数が期待値と異なります: got %d, want %d", len(template.ServerConfig.Env), len(tt.expected))
				return
			}
			
			for key, value := range tt.expected {
				if template.ServerConfig.Env[key] != value {
					t.Errorf("環境変数[%s] = %v, want %v", key, template.ServerConfig.Env[key], value)
				}
			}
		})
	}
}