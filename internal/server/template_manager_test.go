package server

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/naoto24kawa/mcpjson/internal/config"
)

const (
	testTemplateName      = "test-template"
	testTemplateNameOld   = "old-template"
	testTemplateNameNew   = "new-template"
	testServerName        = "test-server"
	testCommand           = "python"
	testServerDescription = "テスト用サーバーテンプレート"
)

func createTestMCPConfig(t *testing.T, filePath string, config *MCPConfig) {
	t.Helper()

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

func createTestTemplate(t *testing.T, manager *TemplateManager, name string) {
	t.Helper()

	err := manager.SaveFromConfig(name, MCPServer{
		Command: testCommand,
		Args:    []string{"test.py"},
		Env:     map[string]string{"TEST_ENV": "value"},
	})
	if err != nil {
		t.Fatalf("Failed to create test template %s: %v", name, err)
	}
}

func TestTemplateManager_SaveFromFile_Success(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	manager := NewTemplateManager(tempDir)

	mcpConfigPath := filepath.Join(tempDir, "test_config.json")
	testConfig := &MCPConfig{
		McpServers: map[string]MCPServer{
			testServerName: {
				Command: testCommand,
				Args:    []string{"test.py"},
				Env:     map[string]string{"ENV_VAR": "test_value"},
			},
		},
	}
	createTestMCPConfig(t, mcpConfigPath, testConfig)

	// Act
	err := manager.SaveFromFile(testTemplateName, testServerName, mcpConfigPath, false)

	// Assert
	if err != nil {
		t.Errorf("SaveFromFile() failed: %v", err)
	}

	// テンプレートが正しく保存されているか確認
	template, err := manager.Load(testTemplateName)
	if err != nil {
		t.Fatalf("Failed to load saved template: %v", err)
	}

	if template.Name != testTemplateName {
		t.Errorf("Template name mismatch: got %s, want %s", template.Name, testTemplateName)
	}
	if template.ServerConfig.Command != testCommand {
		t.Errorf("Command mismatch: got %s, want %s", template.ServerConfig.Command, testCommand)
	}
}

func TestTemplateManager_SaveFromFile_ServerNotFound(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	manager := NewTemplateManager(tempDir)

	mcpConfigPath := filepath.Join(tempDir, "test_config.json")
	testConfig := &MCPConfig{
		McpServers: map[string]MCPServer{},
	}
	createTestMCPConfig(t, mcpConfigPath, testConfig)

	// Act
	err := manager.SaveFromFile(testTemplateName, "nonexistent-server", mcpConfigPath, false)

	// Assert
	if err == nil {
		t.Error("SaveFromFile() expected error for nonexistent server, got nil")
	}
}

func TestTemplateManager_SaveFromFile_InvalidMCPConfig(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	manager := NewTemplateManager(tempDir)
	nonexistentPath := filepath.Join(tempDir, "nonexistent.json")

	// Act
	err := manager.SaveFromFile(testTemplateName, testServerName, nonexistentPath, false)

	// Assert
	if err == nil {
		t.Error("SaveFromFile() expected error for invalid MCP config path, got nil")
	}
}

func TestTemplateManager_SaveFromConfig_Success(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	manager := NewTemplateManager(tempDir)

	testServer := MCPServer{
		Command: testCommand,
		Args:    []string{"server.py"},
		Env:     map[string]string{"TEST": "value"},
	}

	// Act
	err := manager.SaveFromConfig(testTemplateName, testServer)

	// Assert
	if err != nil {
		t.Errorf("SaveFromConfig() failed: %v", err)
	}

	// テンプレートが正しく保存されているか確認
	template, err := manager.Load(testTemplateName)
	if err != nil {
		t.Fatalf("Failed to load saved template: %v", err)
	}

	if template.Name != testTemplateName {
		t.Errorf("Template name mismatch: got %s, want %s", template.Name, testTemplateName)
	}
	if template.ServerConfig.Command != testCommand {
		t.Errorf("Command mismatch: got %s, want %s", template.ServerConfig.Command, testCommand)
	}
}

func TestTemplateManager_Load_Success(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	manager := NewTemplateManager(tempDir)
	createTestTemplate(t, manager, testTemplateName)

	// Act
	template, err := manager.Load(testTemplateName)

	// Assert
	if err != nil {
		t.Errorf("Load() failed: %v", err)
	}
	if template == nil {
		t.Fatal("Load() returned nil template")
	}
	if template.Name != testTemplateName {
		t.Errorf("Template name mismatch: got %s, want %s", template.Name, testTemplateName)
	}
}

func TestTemplateManager_Load_NotFound(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	manager := NewTemplateManager(tempDir)

	// Act
	template, err := manager.Load("nonexistent-template")

	// Assert
	if err == nil {
		t.Error("Load() expected error for nonexistent template, got nil")
	}
	if template != nil {
		t.Error("Load() expected nil template for nonexistent template")
	}
}

func TestTemplateManager_Exists_True(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	manager := NewTemplateManager(tempDir)
	createTestTemplate(t, manager, testTemplateName)

	// Act
	exists, err := manager.Exists(testTemplateName)

	// Assert
	if err != nil {
		t.Errorf("Exists() failed: %v", err)
	}
	if !exists {
		t.Error("Exists() returned false for existing template")
	}
}

func TestTemplateManager_Exists_False(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	manager := NewTemplateManager(tempDir)

	// Act
	exists, err := manager.Exists("nonexistent-template")

	// Assert
	if err != nil {
		t.Errorf("Exists() failed: %v", err)
	}
	if exists {
		t.Error("Exists() returned true for nonexistent template")
	}
}

func TestTemplateManager_Delete_Success(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	manager := NewTemplateManager(tempDir)
	createTestTemplate(t, manager, testTemplateName)

	// Act
	err := manager.Delete(testTemplateName, true, nil) // force=true to skip confirmation

	// Assert
	if err != nil {
		t.Errorf("Delete() failed: %v", err)
	}

	// テンプレートが削除されているか確認
	exists, _ := manager.Exists(testTemplateName)
	if exists {
		t.Error("Template still exists after deletion")
	}
}

func TestTemplateManager_Delete_NotFound(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	manager := NewTemplateManager(tempDir)

	// Act
	err := manager.Delete("nonexistent-template", true, nil)

	// Assert
	if err == nil {
		t.Error("Delete() expected error for nonexistent template, got nil")
	}
}

func TestTemplateManager_Rename_Success(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	manager := NewTemplateManager(tempDir)
	createTestTemplate(t, manager, testTemplateNameOld)

	// Act
	err := manager.Rename(testTemplateNameOld, testTemplateNameNew, false)

	// Assert
	if err != nil {
		t.Errorf("Rename() failed: %v", err)
	}

	// 古いテンプレートが存在しないことを確認
	exists, _ := manager.Exists(testTemplateNameOld)
	if exists {
		t.Error("Old template still exists after rename")
	}

	// 新しいテンプレートが存在することを確認
	template, err := manager.Load(testTemplateNameNew)
	if err != nil {
		t.Fatalf("Failed to load renamed template: %v", err)
	}
	if template.Name != testTemplateNameNew {
		t.Errorf("Template name not updated: got %s, want %s", template.Name, testTemplateNameNew)
	}
}

func TestTemplateManager_Rename_SourceNotFound(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	manager := NewTemplateManager(tempDir)

	// Act
	err := manager.Rename("nonexistent-template", testTemplateNameNew, false)

	// Assert
	if err == nil {
		t.Error("Rename() expected error for nonexistent source template, got nil")
	}
}

func TestTemplateManager_Rename_TargetExists(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	manager := NewTemplateManager(tempDir)
	createTestTemplate(t, manager, testTemplateNameOld)
	createTestTemplate(t, manager, testTemplateNameNew)

	// Act
	err := manager.Rename(testTemplateNameOld, testTemplateNameNew, false)

	// Assert
	if err == nil {
		t.Error("Rename() expected error when target exists without force, got nil")
	}
}

func TestTemplateManager_Rename_ForceOverwrite(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	manager := NewTemplateManager(tempDir)
	createTestTemplate(t, manager, testTemplateNameOld)
	createTestTemplate(t, manager, testTemplateNameNew)

	// Act
	err := manager.Rename(testTemplateNameOld, testTemplateNameNew, true)

	// Assert
	if err != nil {
		t.Errorf("Rename() with force failed: %v", err)
	}

	// 古いテンプレートが存在しないことを確認
	exists, _ := manager.Exists(testTemplateNameOld)
	if exists {
		t.Error("Old template still exists after forced rename")
	}

	// 新しいテンプレートが存在することを確認
	template, err := manager.Load(testTemplateNameNew)
	if err != nil {
		t.Fatalf("Failed to load renamed template: %v", err)
	}
	if template.Name != testTemplateNameNew {
		t.Errorf("Template name not updated: got %s, want %s", template.Name, testTemplateNameNew)
	}
}

func TestTemplateManager_Reset_EmptyDirectory(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	manager := NewTemplateManager(tempDir)

	// Act
	err := manager.Reset(true)

	// Assert
	if err != nil {
		t.Errorf("Reset() on empty directory failed: %v", err)
	}
}

func TestTemplateManager_Reset_NonexistentDirectory(t *testing.T) {
	// Arrange
	nonexistentDir := filepath.Join(os.TempDir(), "nonexistent-dir")
	manager := NewTemplateManager(nonexistentDir)

	// Act
	err := manager.Reset(true)

	// Assert
	if err != nil {
		t.Errorf("Reset() on nonexistent directory failed: %v", err)
	}
}

func TestTemplateManager_Reset_MultipleTemplates(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	manager := NewTemplateManager(tempDir)

	templateNames := []string{"template1", "template2", "template3"}
	for _, name := range templateNames {
		createTestTemplate(t, manager, name)
	}

	// Act
	err := manager.Reset(true)

	// Assert
	if err != nil {
		t.Errorf("Reset() with multiple templates failed: %v", err)
	}

	// すべてのテンプレートが削除されているか確認
	for _, name := range templateNames {
		exists, _ := manager.Exists(name)
		if exists {
			t.Errorf("Template %s still exists after reset", name)
		}
	}
}

func TestTemplateManager_Reset_WithNonTemplateFiles(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	manager := NewTemplateManager(tempDir)

	createTestTemplate(t, manager, testTemplateName)

	// テンプレートファイル以外のファイルを作成
	nonTemplateFile := filepath.Join(tempDir, "not-a-template.txt")
	if err := os.WriteFile(nonTemplateFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create non-template file: %v", err)
	}

	// Act
	err := manager.Reset(true)

	// Assert
	if err != nil {
		t.Errorf("Reset() failed: %v", err)
	}

	// テンプレートファイルが削除されているか確認
	exists, _ := manager.Exists(testTemplateName)
	if exists {
		t.Error("Template still exists after reset")
	}

	// 非テンプレートファイルが残っているか確認
	if _, err := os.Stat(nonTemplateFile); os.IsNotExist(err) {
		t.Error("Non-template file was unexpectedly deleted")
	}
}

func TestTemplateManager_SaveFromFile_WithForce(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	manager := NewTemplateManager(tempDir)

	// 既存のテンプレートを作成
	createTestTemplate(t, manager, testTemplateName)

	mcpConfigPath := filepath.Join(tempDir, "test_config.json")
	testConfig := &MCPConfig{
		McpServers: map[string]MCPServer{
			testServerName: {
				Command: "node",
				Args:    []string{"server.js"},
				Env:     map[string]string{"NODE_ENV": "test"},
			},
		},
	}
	createTestMCPConfig(t, mcpConfigPath, testConfig)

	// Act
	err := manager.SaveFromFile(testTemplateName, testServerName, mcpConfigPath, true)

	// Assert
	if err != nil {
		t.Errorf("SaveFromFile() with force failed: %v", err)
	}

	// テンプレートが上書きされているか確認
	template, err := manager.Load(testTemplateName)
	if err != nil {
		t.Fatalf("Failed to load overwritten template: %v", err)
	}

	if template.ServerConfig.Command != "node" {
		t.Errorf("Template was not overwritten: got command %s, want node", template.ServerConfig.Command)
	}
}

// エラーハンドリングのテスト
func TestTemplateManager_SaveFromFile_InvalidJSON(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	manager := NewTemplateManager(tempDir)

	invalidJSONPath := filepath.Join(tempDir, "invalid.json")
	if err := os.WriteFile(invalidJSONPath, []byte("invalid json"), 0644); err != nil {
		t.Fatalf("Failed to create invalid JSON file: %v", err)
	}

	// Act
	err := manager.SaveFromFile(testTemplateName, testServerName, invalidJSONPath, false)

	// Assert
	if err == nil {
		t.Error("SaveFromFile() expected error for invalid JSON, got nil")
	}
}

func TestTemplateManager_Load_CorruptedFile(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	manager := NewTemplateManager(tempDir)

	// 破損したファイルを作成
	corruptedPath := filepath.Join(tempDir, testTemplateName+config.FileExtension)
	if err := os.WriteFile(corruptedPath, []byte("corrupted json"), 0644); err != nil {
		t.Fatalf("Failed to create corrupted file: %v", err)
	}

	// Act
	template, err := manager.Load(testTemplateName)

	// Assert
	if err == nil {
		t.Error("Load() expected error for corrupted file, got nil")
	}
	if template != nil {
		t.Error("Load() expected nil template for corrupted file")
	}
}

// ヘルパー関数のテスト
func TestTemplateManager_exists(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	manager := NewTemplateManager(tempDir)

	// Act & Assert for non-existent template
	if manager.exists("nonexistent") {
		t.Error("exists() returned true for nonexistent template")
	}

	// Create template and test again
	createTestTemplate(t, manager, testTemplateName)
	if !manager.exists(testTemplateName) {
		t.Error("exists() returned false for existing template")
	}
}

func TestTemplateManager_save(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	manager := NewTemplateManager(tempDir)

	template := &ServerTemplate{
		Name:        testTemplateName,
		Description: nil,
		CreatedAt:   time.Now(),
		ServerConfig: ServerConfig{
			Command: testCommand,
			Args:    []string{"test.py"},
			Env:     map[string]string{"TEST": "value"},
		},
	}

	// Act
	err := manager.save(template)

	// Assert
	if err != nil {
		t.Errorf("save() failed: %v", err)
	}

	// ファイルが作成されているか確認
	templatePath := filepath.Join(tempDir, testTemplateName+config.FileExtension)
	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		t.Error("Template file was not created")
	}
}

func TestTemplateManager_Copy_Success(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	manager := NewTemplateManager(tempDir)
	createTestTemplate(t, manager, testTemplateNameOld)

	// Act
	err := manager.Copy(testTemplateNameOld, testTemplateNameNew, false)

	// Assert
	if err != nil {
		t.Errorf("Copy() failed: %v", err)
	}

	// 元のテンプレートが存在することを確認
	exists, _ := manager.Exists(testTemplateNameOld)
	if !exists {
		t.Error("Source template should still exist after copy")
	}

	// コピー先のテンプレートが存在することを確認
	template, err := manager.Load(testTemplateNameNew)
	if err != nil {
		t.Fatalf("Failed to load copied template: %v", err)
	}
	if template.Name != testTemplateNameNew {
		t.Errorf("Copied template name not updated: got %s, want %s", template.Name, testTemplateNameNew)
	}
	if template.ServerConfig.Command != testCommand {
		t.Errorf("Copied template command mismatch: got %s, want %s", template.ServerConfig.Command, testCommand)
	}
}

func TestTemplateManager_Copy_SourceNotFound(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	manager := NewTemplateManager(tempDir)

	// Act
	err := manager.Copy("nonexistent-template", testTemplateNameNew, false)

	// Assert
	if err == nil {
		t.Error("Copy() expected error for nonexistent source template, got nil")
	}
}

func TestTemplateManager_Copy_DestinationExists(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	manager := NewTemplateManager(tempDir)
	createTestTemplate(t, manager, testTemplateNameOld)
	createTestTemplate(t, manager, testTemplateNameNew)

	// Act
	err := manager.Copy(testTemplateNameOld, testTemplateNameNew, false)

	// Assert
	if err == nil {
		t.Error("Copy() expected error when destination exists without force, got nil")
	}

	// 元のテンプレートが変更されていないことを確認
	template, err := manager.Load(testTemplateNameOld)
	if err != nil {
		t.Fatalf("Failed to load source template after failed copy: %v", err)
	}
	if template.Name != testTemplateNameOld {
		t.Error("Source template was unexpectedly modified")
	}
}

func TestTemplateManager_Copy_ForceOverwrite(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	manager := NewTemplateManager(tempDir)
	createTestTemplate(t, manager, testTemplateNameOld)

	// コピー先に異なる設定のテンプレートを作成
	err := manager.SaveFromConfig(testTemplateNameNew, MCPServer{
		Command: "node",
		Args:    []string{"different.js"},
		Env:     map[string]string{"DIFFERENT": "value"},
	})
	if err != nil {
		t.Fatalf("Failed to create destination template: %v", err)
	}

	// Act
	err = manager.Copy(testTemplateNameOld, testTemplateNameNew, true)

	// Assert
	if err != nil {
		t.Errorf("Copy() with force failed: %v", err)
	}

	// 元のテンプレートが存在することを確認
	exists, _ := manager.Exists(testTemplateNameOld)
	if !exists {
		t.Error("Source template should still exist after copy")
	}

	// コピー先のテンプレートが上書きされているか確認
	template, err := manager.Load(testTemplateNameNew)
	if err != nil {
		t.Fatalf("Failed to load copied template: %v", err)
	}
	if template.Name != testTemplateNameNew {
		t.Errorf("Copied template name not updated: got %s, want %s", template.Name, testTemplateNameNew)
	}
	if template.ServerConfig.Command != testCommand {
		t.Errorf("Template was not overwritten: got command %s, want %s", template.ServerConfig.Command, testCommand)
	}
}

func TestTemplateManager_Copy_SameName(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	manager := NewTemplateManager(tempDir)
	createTestTemplate(t, manager, testTemplateName)

	// Act
	err := manager.Copy(testTemplateName, testTemplateName, false)

	// Assert
	if err == nil {
		t.Error("Copy() expected error when source and destination are the same, got nil")
	}
}

func TestTemplateManager_Copy_CreatedAtUpdate(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	manager := NewTemplateManager(tempDir)
	createTestTemplate(t, manager, testTemplateNameOld)

	// 元のテンプレートのCreatedAtを取得
	originalTemplate, err := manager.Load(testTemplateNameOld)
	if err != nil {
		t.Fatalf("Failed to load original template: %v", err)
	}

	// 少し待機してからコピー
	time.Sleep(time.Millisecond * 10)

	// Act
	err = manager.Copy(testTemplateNameOld, testTemplateNameNew, false)

	// Assert
	if err != nil {
		t.Errorf("Copy() failed: %v", err)
	}

	// コピーされたテンプレートのCreatedAtが更新されているか確認
	copiedTemplate, err := manager.Load(testTemplateNameNew)
	if err != nil {
		t.Fatalf("Failed to load copied template: %v", err)
	}

	if !copiedTemplate.CreatedAt.After(originalTemplate.CreatedAt) {
		t.Error("Copied template CreatedAt should be newer than original")
	}
}

func TestTemplateManager_Copy_EmptySourceName(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	manager := NewTemplateManager(tempDir)

	// Act
	err := manager.Copy("", testTemplateNameNew, false)

	// Assert
	if err == nil {
		t.Error("Copy() expected error for empty source name, got nil")
	}
}

func TestTemplateManager_Copy_EmptyDestinationName(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	manager := NewTemplateManager(tempDir)
	createTestTemplate(t, manager, testTemplateName)

	// Act
	err := manager.Copy(testTemplateName, "", false)

	// Assert
	if err == nil {
		t.Error("Copy() expected error for empty destination name, got nil")
	}
}

func TestTemplateManager_Copy_PreservesAllFields(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	manager := NewTemplateManager(tempDir)

	// より複雑なテンプレートを作成
	testDescription := "Test template description"
	complexServer := MCPServer{
		Command: testCommand,
		Args:    []string{"test.py", "--verbose", "--config", "test.json"},
		Env: map[string]string{
			"TEST_ENV":     "value",
			"ANOTHER_ENV":  "another_value",
			"COMPLEX_PATH": "/path/to/complex/dir",
		},
	}

	err := manager.SaveFromConfig(testTemplateNameOld, complexServer)
	if err != nil {
		t.Fatalf("Failed to create complex template: %v", err)
	}

	// 手動でdescriptionを追加（SaveFromConfigでは追加されないため）
	originalTemplate, err := manager.Load(testTemplateNameOld)
	if err != nil {
		t.Fatalf("Failed to load template for description update: %v", err)
	}
	originalTemplate.Description = &testDescription
	err = manager.save(originalTemplate)
	if err != nil {
		t.Fatalf("Failed to save template with description: %v", err)
	}

	// Act
	err = manager.Copy(testTemplateNameOld, testTemplateNameNew, false)

	// Assert
	if err != nil {
		t.Errorf("Copy() failed: %v", err)
	}

	// コピーされたテンプレートのすべてのフィールドを確認
	copiedTemplate, err := manager.Load(testTemplateNameNew)
	if err != nil {
		t.Fatalf("Failed to load copied template: %v", err)
	}

	// Name以外のすべてのフィールドが正しくコピーされているか確認
	if copiedTemplate.Name != testTemplateNameNew {
		t.Errorf("Name not updated: got %s, want %s", copiedTemplate.Name, testTemplateNameNew)
	}

	if copiedTemplate.Description == nil || *copiedTemplate.Description != testDescription {
		t.Errorf("Description not copied correctly: got %v, want %s", copiedTemplate.Description, testDescription)
	}

	if copiedTemplate.ServerConfig.Command != complexServer.Command {
		t.Errorf("Command not copied: got %s, want %s", copiedTemplate.ServerConfig.Command, complexServer.Command)
	}

	if len(copiedTemplate.ServerConfig.Args) != len(complexServer.Args) {
		t.Errorf("Args length mismatch: got %d, want %d", len(copiedTemplate.ServerConfig.Args), len(complexServer.Args))
	}

	for i, arg := range complexServer.Args {
		if copiedTemplate.ServerConfig.Args[i] != arg {
			t.Errorf("Args[%d] mismatch: got %s, want %s", i, copiedTemplate.ServerConfig.Args[i], arg)
		}
	}

	if len(copiedTemplate.ServerConfig.Env) != len(complexServer.Env) {
		t.Errorf("Env length mismatch: got %d, want %d", len(copiedTemplate.ServerConfig.Env), len(complexServer.Env))
	}

	for key, value := range complexServer.Env {
		if copiedTemplate.ServerConfig.Env[key] != value {
			t.Errorf("Env[%s] mismatch: got %s, want %s", key, copiedTemplate.ServerConfig.Env[key], value)
		}
	}
}
