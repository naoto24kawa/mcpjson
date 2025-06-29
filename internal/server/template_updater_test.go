package server

import (
	"testing"
)

const (
	updaterTestTemplateName = "updater-test-template"
	updaterTestCommand      = "node"
	updaterNewCommand       = "python"
)

func createTemplateUpdaterForTest(t *testing.T) (*TemplateUpdater, *TemplateManager, string) {
	t.Helper()

	tempDir := t.TempDir()
	manager := NewTemplateManager(tempDir)
	updater := NewTemplateUpdater(manager)

	return updater, manager, tempDir
}

func TestTemplateUpdater_SaveManual_CreateNew(t *testing.T) {
	// Arrange
	updater, manager, _ := createTemplateUpdaterForTest(t)

	args := []string{"server.js"}
	env := map[string]string{"NODE_ENV": "test"}

	// Act
	err := updater.SaveManual(updaterTestTemplateName, updaterTestCommand, args, env, false)

	// Assert
	if err != nil {
		t.Errorf("SaveManual() failed to create new template: %v", err)
	}

	// テンプレートが作成されているか確認
	template, err := manager.Load(updaterTestTemplateName)
	if err != nil {
		t.Fatalf("Failed to load created template: %v", err)
	}

	if template.Name != updaterTestTemplateName {
		t.Errorf("Template name mismatch: got %s, want %s", template.Name, updaterTestTemplateName)
	}
	if template.ServerConfig.Command != updaterTestCommand {
		t.Errorf("Command mismatch: got %s, want %s", template.ServerConfig.Command, updaterTestCommand)
	}
	if len(template.ServerConfig.Args) != 1 || template.ServerConfig.Args[0] != "server.js" {
		t.Errorf("Args mismatch: got %v, want [server.js]", template.ServerConfig.Args)
	}
	if template.ServerConfig.Env["NODE_ENV"] != "test" {
		t.Errorf("Env mismatch: got %v", template.ServerConfig.Env)
	}
}

func TestTemplateUpdater_SaveManual_UpdateExisting(t *testing.T) {
	// Arrange
	updater, manager, _ := createTemplateUpdaterForTest(t)

	// 既存のテンプレートを作成
	err := manager.SaveFromConfig(updaterTestTemplateName, MCPServer{
		Command: updaterTestCommand,
		Args:    []string{"old.js"},
		Env:     map[string]string{"OLD_ENV": "old_value"},
	})
	if err != nil {
		t.Fatalf("Failed to create existing template: %v", err)
	}

	// Act - テンプレートを更新
	newArgs := []string{"new.py"}
	newEnv := map[string]string{"NEW_ENV": "new_value"}
	err = updater.SaveManual(updaterTestTemplateName, updaterNewCommand, newArgs, newEnv, true)

	// Assert
	if err != nil {
		t.Errorf("SaveManual() failed to update existing template: %v", err)
	}

	// テンプレートが更新されているか確認
	template, err := manager.Load(updaterTestTemplateName)
	if err != nil {
		t.Fatalf("Failed to load updated template: %v", err)
	}

	if template.ServerConfig.Command != updaterNewCommand {
		t.Errorf("Command not updated: got %s, want %s", template.ServerConfig.Command, updaterNewCommand)
	}
	if len(template.ServerConfig.Args) != 1 || template.ServerConfig.Args[0] != "new.py" {
		t.Errorf("Args not updated: got %v, want [new.py]", template.ServerConfig.Args)
	}
	if template.ServerConfig.Env["NEW_ENV"] != "new_value" {
		t.Errorf("Env not updated: got %v", template.ServerConfig.Env)
	}
}

func TestTemplateUpdater_SaveManual_EmptyCommand(t *testing.T) {
	// Arrange
	updater, _, _ := createTemplateUpdaterForTest(t)

	// Act - 空のコマンドで新しいテンプレートを作成
	err := updater.SaveManual(updaterTestTemplateName, "", nil, nil, false)

	// Assert
	if err == nil {
		t.Error("SaveManual() expected error for empty command in new template, got nil")
	}
}

func TestTemplateUpdater_SaveManual_UpdateExistingEmptyCommand(t *testing.T) {
	// Arrange
	updater, manager, _ := createTemplateUpdaterForTest(t)

	// 既存のテンプレートを作成
	err := manager.SaveFromConfig(updaterTestTemplateName, MCPServer{
		Command: updaterTestCommand,
		Args:    []string{"test.js"},
	})
	if err != nil {
		t.Fatalf("Failed to create existing template: %v", err)
	}

	// Act - 空のコマンドで既存テンプレートを更新（コマンドは変更されないはず）
	err = updater.SaveManual(updaterTestTemplateName, "", []string{"new.js"}, nil, true)

	// Assert
	if err != nil {
		t.Errorf("SaveManual() failed to update existing template with empty command: %v", err)
	}

	// コマンドが変更されていないことを確認
	template, err := manager.Load(updaterTestTemplateName)
	if err != nil {
		t.Fatalf("Failed to load updated template: %v", err)
	}

	if template.ServerConfig.Command != updaterTestCommand {
		t.Errorf("Command unexpectedly changed: got %s, want %s", template.ServerConfig.Command, updaterTestCommand)
	}
	if len(template.ServerConfig.Args) != 1 || template.ServerConfig.Args[0] != "new.js" {
		t.Errorf("Args not updated: got %v, want [new.js]", template.ServerConfig.Args)
	}
}

func TestTemplateUpdater_UpdateTemplateArgs_Normal(t *testing.T) {
	// Arrange
	updater, manager, _ := createTemplateUpdaterForTest(t)
	err := manager.SaveFromConfig(updaterTestTemplateName, MCPServer{
		Command: updaterTestCommand,
		Args:    []string{"old.js"},
	})
	if err != nil {
		t.Fatalf("Failed to create template: %v", err)
	}

	template, err := manager.Load(updaterTestTemplateName)
	if err != nil {
		t.Fatalf("Failed to load template: %v", err)
	}

	// Act
	newArgs := []string{"new.js", "--verbose"}
	updater.UpdateTemplateArgs(template, newArgs)

	// Assert
	if len(template.ServerConfig.Args) != 2 {
		t.Errorf("Args length mismatch: got %d, want 2", len(template.ServerConfig.Args))
	}
	if template.ServerConfig.Args[0] != "new.js" || template.ServerConfig.Args[1] != "--verbose" {
		t.Errorf("Args content mismatch: got %v, want [new.js --verbose]", template.ServerConfig.Args)
	}
}

func TestTemplateUpdater_UpdateTemplateArgs_Nil(t *testing.T) {
	// Arrange
	updater, manager, _ := createTemplateUpdaterForTest(t)
	err := manager.SaveFromConfig(updaterTestTemplateName, MCPServer{
		Command: updaterTestCommand,
		Args:    []string{"old.js"},
	})
	if err != nil {
		t.Fatalf("Failed to create template: %v", err)
	}

	template, err := manager.Load(updaterTestTemplateName)
	if err != nil {
		t.Fatalf("Failed to load template: %v", err)
	}

	originalArgs := template.ServerConfig.Args

	// Act - nilを渡すと引数は変更されない
	updater.UpdateTemplateArgs(template, nil)

	// Assert
	if len(template.ServerConfig.Args) != len(originalArgs) {
		t.Errorf("Args should not change when nil is passed: got %v, want %v", template.ServerConfig.Args, originalArgs)
	}
}

func TestTemplateUpdater_UpdateTemplateArgs_EmptyString(t *testing.T) {
	// Arrange
	updater, manager, _ := createTemplateUpdaterForTest(t)
	err := manager.SaveFromConfig(updaterTestTemplateName, MCPServer{
		Command: updaterTestCommand,
		Args:    []string{"old.js"},
	})
	if err != nil {
		t.Fatalf("Failed to create template: %v", err)
	}

	template, err := manager.Load(updaterTestTemplateName)
	if err != nil {
		t.Fatalf("Failed to load template: %v", err)
	}

	// Act - 空文字列を含む配列を渡すと引数がクリアされる
	updater.UpdateTemplateArgs(template, []string{""})

	// Assert
	if template.ServerConfig.Args != nil {
		t.Errorf("Args should be cleared when empty string is passed: got %v, want nil", template.ServerConfig.Args)
	}
}

func TestTemplateUpdater_UpdateTemplateEnv_Normal(t *testing.T) {
	// Arrange
	updater, manager, _ := createTemplateUpdaterForTest(t)
	err := manager.SaveFromConfig(updaterTestTemplateName, MCPServer{
		Command: updaterTestCommand,
		Env:     map[string]string{"OLD_VAR": "old_value"},
	})
	if err != nil {
		t.Fatalf("Failed to create template: %v", err)
	}

	template, err := manager.Load(updaterTestTemplateName)
	if err != nil {
		t.Fatalf("Failed to load template: %v", err)
	}

	// Act
	newEnv := map[string]string{"NEW_VAR": "new_value", "ANOTHER_VAR": "another_value"}
	updater.UpdateTemplateEnv(template, newEnv)

	// Assert
	if template.ServerConfig.Env["NEW_VAR"] != "new_value" {
		t.Errorf("NEW_VAR not set: got %s", template.ServerConfig.Env["NEW_VAR"])
	}
	if template.ServerConfig.Env["ANOTHER_VAR"] != "another_value" {
		t.Errorf("ANOTHER_VAR not set: got %s", template.ServerConfig.Env["ANOTHER_VAR"])
	}
	if template.ServerConfig.Env["OLD_VAR"] != "old_value" {
		t.Errorf("OLD_VAR should be preserved: got %s", template.ServerConfig.Env["OLD_VAR"])
	}
}

func TestTemplateUpdater_UpdateTemplateEnv_Nil(t *testing.T) {
	// Arrange
	updater, manager, _ := createTemplateUpdaterForTest(t)
	err := manager.SaveFromConfig(updaterTestTemplateName, MCPServer{
		Command: updaterTestCommand,
		Env:     map[string]string{"OLD_VAR": "old_value"},
	})
	if err != nil {
		t.Fatalf("Failed to create template: %v", err)
	}

	template, err := manager.Load(updaterTestTemplateName)
	if err != nil {
		t.Fatalf("Failed to load template: %v", err)
	}

	originalEnv := template.ServerConfig.Env

	// Act - nilを渡すと環境変数は変更されない
	updater.UpdateTemplateEnv(template, nil)

	// Assert
	if len(template.ServerConfig.Env) != len(originalEnv) {
		t.Errorf("Env should not change when nil is passed: got %v, want %v", template.ServerConfig.Env, originalEnv)
	}
}

func TestTemplateUpdater_UpdateTemplateEnv_EmptyMap(t *testing.T) {
	// Arrange
	updater, manager, _ := createTemplateUpdaterForTest(t)
	err := manager.SaveFromConfig(updaterTestTemplateName, MCPServer{
		Command: updaterTestCommand,
		Env:     map[string]string{"OLD_VAR": "old_value"},
	})
	if err != nil {
		t.Fatalf("Failed to create template: %v", err)
	}

	template, err := manager.Load(updaterTestTemplateName)
	if err != nil {
		t.Fatalf("Failed to load template: %v", err)
	}

	// Act - 空のマップを渡すと環境変数がクリアされる
	updater.UpdateTemplateEnv(template, map[string]string{})

	// Assert
	if template.ServerConfig.Env != nil {
		t.Errorf("Env should be cleared when empty map is passed: got %v, want nil", template.ServerConfig.Env)
	}
}

func TestTemplateUpdater_UpdateTemplateEnv_DeleteVariable(t *testing.T) {
	// Arrange
	updater, manager, _ := createTemplateUpdaterForTest(t)
	err := manager.SaveFromConfig(updaterTestTemplateName, MCPServer{
		Command: updaterTestCommand,
		Env:     map[string]string{"VAR1": "value1", "VAR2": "value2"},
	})
	if err != nil {
		t.Fatalf("Failed to create template: %v", err)
	}

	template, err := manager.Load(updaterTestTemplateName)
	if err != nil {
		t.Fatalf("Failed to load template: %v", err)
	}

	// Act - 空文字列を値に設定すると環境変数が削除される
	updater.UpdateTemplateEnv(template, map[string]string{"VAR1": ""})

	// Assert
	if _, exists := template.ServerConfig.Env["VAR1"]; exists {
		t.Error("VAR1 should be deleted when empty string is set")
	}
	if template.ServerConfig.Env["VAR2"] != "value2" {
		t.Errorf("VAR2 should be preserved: got %s", template.ServerConfig.Env["VAR2"])
	}
}

func TestTemplateUpdater_UpdateTemplateEnv_InitializeEnv(t *testing.T) {
	// Arrange
	updater, manager, _ := createTemplateUpdaterForTest(t)
	err := manager.SaveFromConfig(updaterTestTemplateName, MCPServer{
		Command: updaterTestCommand,
		// Envはnil
	})
	if err != nil {
		t.Fatalf("Failed to create template: %v", err)
	}

	template, err := manager.Load(updaterTestTemplateName)
	if err != nil {
		t.Fatalf("Failed to load template: %v", err)
	}

	// Act - nilの環境変数に新しい値を設定
	updater.UpdateTemplateEnv(template, map[string]string{"NEW_VAR": "new_value"})

	// Assert
	if template.ServerConfig.Env == nil {
		t.Error("Env should be initialized")
	}
	if template.ServerConfig.Env["NEW_VAR"] != "new_value" {
		t.Errorf("NEW_VAR not set: got %s", template.ServerConfig.Env["NEW_VAR"])
	}
}

func TestTemplateUpdater_templateExists(t *testing.T) {
	// Arrange
	updater, manager, _ := createTemplateUpdaterForTest(t)

	// Act & Assert - 存在しないテンプレート
	if updater.templateExists("nonexistent") {
		t.Error("templateExists() returned true for nonexistent template")
	}

	// テンプレートを作成
	err := manager.SaveFromConfig(updaterTestTemplateName, MCPServer{
		Command: updaterTestCommand,
	})
	if err != nil {
		t.Fatalf("Failed to create template: %v", err)
	}

	// Act & Assert - 存在するテンプレート
	if !updater.templateExists(updaterTestTemplateName) {
		t.Error("templateExists() returned false for existing template")
	}
}

func TestTemplateUpdater_createNewTemplate(t *testing.T) {
	// Arrange
	updater, _, _ := createTemplateUpdaterForTest(t)

	command := "python"
	args := []string{"script.py"}
	env := map[string]string{"PYTHON_ENV": "test"}

	// Act
	template, err := updater.createNewTemplate(updaterTestTemplateName, command, args, env)

	// Assert
	if err != nil {
		t.Errorf("createNewTemplate() failed: %v", err)
	}
	if template == nil {
		t.Fatal("createNewTemplate() returned nil template")
	}
	if template.Name != updaterTestTemplateName {
		t.Errorf("Template name mismatch: got %s, want %s", template.Name, updaterTestTemplateName)
	}
	if template.ServerConfig.Command != command {
		t.Errorf("Command mismatch: got %s, want %s", template.ServerConfig.Command, command)
	}
}

func TestTemplateUpdater_createNewTemplate_EmptyCommand(t *testing.T) {
	// Arrange
	updater, _, _ := createTemplateUpdaterForTest(t)

	// Act
	template, err := updater.createNewTemplate(updaterTestTemplateName, "", nil, nil)

	// Assert
	if err == nil {
		t.Error("createNewTemplate() expected error for empty command, got nil")
	}
	if template != nil {
		t.Error("createNewTemplate() expected nil template for empty command")
	}
}

func TestTemplateUpdater_updateExistingTemplate(t *testing.T) {
	// Arrange
	updater, manager, _ := createTemplateUpdaterForTest(t)

	// 既存のテンプレートを作成
	err := manager.SaveFromConfig(updaterTestTemplateName, MCPServer{
		Command: "old-command",
		Args:    []string{"old.py"},
		Env:     map[string]string{"OLD_ENV": "old_value"},
	})
	if err != nil {
		t.Fatalf("Failed to create existing template: %v", err)
	}

	// Act
	newCommand := "new-command"
	newArgs := []string{"new.py"}
	newEnv := map[string]string{"NEW_ENV": "new_value"}
	template, err := updater.updateExistingTemplate(updaterTestTemplateName, newCommand, newArgs, newEnv)

	// Assert
	if err != nil {
		t.Errorf("updateExistingTemplate() failed: %v", err)
	}
	if template == nil {
		t.Error("updateExistingTemplate() returned nil template")
	}
	if template.ServerConfig.Command != newCommand {
		t.Errorf("Command not updated: got %s, want %s", template.ServerConfig.Command, newCommand)
	}
}
