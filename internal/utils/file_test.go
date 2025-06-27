package utils

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestFileExists(t *testing.T) {
	tempDir := t.TempDir()
	existingFile := filepath.Join(tempDir, "existing.txt")
	
	// ファイルを作成
	file, err := os.Create(existingFile)
	if err != nil {
		t.Fatalf("テストファイル作成に失敗: %v", err)
	}
	file.Close()

	tests := []struct {
		name string
		path string
		want bool
	}{
		{
			name: "存在するファイル",
			path: existingFile,
			want: true,
		},
		{
			name: "存在しないファイル",
			path: filepath.Join(tempDir, "nonexistent.txt"),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FileExists(tt.path)
			if got != tt.want {
				t.Errorf("FileExists() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLoadJSON(t *testing.T) {
	tempDir := t.TempDir()
	
	// テスト用JSONファイルを作成
	testData := map[string]interface{}{
		"name": "test",
		"value": 123,
	}
	
	jsonFile := filepath.Join(tempDir, "test.json")
	file, err := os.Create(jsonFile)
	if err != nil {
		t.Fatalf("テストファイル作成に失敗: %v", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	if err := encoder.Encode(testData); err != nil {
		t.Fatalf("テストデータ書き込みに失敗: %v", err)
	}

	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "正常なJSON読み込み",
			path:    jsonFile,
			wantErr: false,
		},
		{
			name:    "存在しないファイル",
			path:    filepath.Join(tempDir, "nonexistent.json"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result map[string]interface{}
			err := LoadJSON(tt.path, &result)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if result["name"] != "test" {
					t.Errorf("LoadJSON() name = %v, want test", result["name"])
				}
				if result["value"].(float64) != 123 {
					t.Errorf("LoadJSON() value = %v, want 123", result["value"])
				}
			}
		})
	}
}

func TestSaveJSON(t *testing.T) {
	tempDir := t.TempDir()
	
	testData := map[string]interface{}{
		"name": "test",
		"value": 456,
	}

	tests := []struct {
		name    string
		path    string
		data    interface{}
		wantErr bool
	}{
		{
			name:    "正常なJSON保存",
			path:    filepath.Join(tempDir, "output.json"),
			data:    testData,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := SaveJSON(tt.path, tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("SaveJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// 保存されたファイルを読み込んで確認
				var result map[string]interface{}
				err := LoadJSON(tt.path, &result)
				if err != nil {
					t.Errorf("保存されたファイルの読み込みに失敗: %v", err)
				}
				
				if result["name"] != "test" {
					t.Errorf("保存されたデータのname = %v, want test", result["name"])
				}
				if result["value"].(float64) != 456 {
					t.Errorf("保存されたデータのvalue = %v, want 456", result["value"])
				}
			}
		})
	}
}

func TestLoadEnvFile(t *testing.T) {
	tempDir := t.TempDir()
	
	// テスト用の環境変数ファイルを作成
	envContent := `# テストコメント
KEY1=value1
KEY2="quoted value"
KEY3='single quoted'
# 別のコメント
KEY4=value with spaces
`
	
	envFile := filepath.Join(tempDir, "test.env")
	if err := os.WriteFile(envFile, []byte(envContent), 0644); err != nil {
		t.Fatalf("テスト環境ファイル作成に失敗: %v", err)
	}

	tests := []struct {
		name    string
		path    string
		want    map[string]string
		wantErr bool
	}{
		{
			name: "正常な環境ファイル読み込み",
			path: envFile,
			want: map[string]string{
				"KEY1": "value1",
				"KEY2": "quoted value",
				"KEY3": "single quoted",
				"KEY4": "value with spaces",
			},
			wantErr: false,
		},
		{
			name:    "存在しないファイル",
			path:    filepath.Join(tempDir, "nonexistent.env"),
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := LoadEnvFile(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadEnvFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				for key, expectedValue := range tt.want {
					if got[key] != expectedValue {
						t.Errorf("LoadEnvFile() key %s = %v, want %v", key, got[key], expectedValue)
					}
				}
				
				// 無効な行が除外されていることを確認
				if len(got) != len(tt.want) {
					t.Errorf("LoadEnvFile() loaded %d keys, want %d", len(got), len(tt.want))
				}
			}
		})
	}
	
	// 不正な形式のファイルのテスト
	t.Run("不正な形式のファイル", func(t *testing.T) {
		invalidFile := filepath.Join(tempDir, "invalid.env")
		invalidContent := "INVALID_LINE_WITHOUT_EQUALS_AND_NO_COMMENT"
		
		if err := os.WriteFile(invalidFile, []byte(invalidContent), 0644); err != nil {
			t.Fatalf("不正なテストファイル作成に失敗: %v", err)
		}
		
		_, err := LoadEnvFile(invalidFile)
		if err == nil {
			t.Errorf("不正な形式のファイルでエラーが発生しませんでした")
		}
	})
}