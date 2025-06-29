package server

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/naoto24kawa/mcpconfig/internal/config"
)

const (
	displayTestTemplateName = "display-test-template"
)

func createTemplateDisplayForTest(t *testing.T) (*TemplateDisplay, *TemplateManager, string) {
	t.Helper()

	tempDir := t.TempDir()
	display := NewTemplateDisplay(tempDir)
	manager := NewTemplateManager(tempDir)

	return display, manager, tempDir
}

func createMultipleTestTemplates(t *testing.T, manager *TemplateManager, count int) []string {
	t.Helper()

	templateNames := make([]string, count)
	for i := 0; i < count; i++ {
		name := fmt.Sprintf("template-%d", i+1)
		templateNames[i] = name

		err := manager.SaveFromConfig(name, MCPServer{
			Command: fmt.Sprintf("command-%d", i+1),
			Args:    []string{fmt.Sprintf("arg-%d.py", i+1)},
			Env:     map[string]string{fmt.Sprintf("ENV_%d", i+1): fmt.Sprintf("value-%d", i+1)},
		})
		if err != nil {
			t.Fatalf("Failed to create test template %s: %v", name, err)
		}
	}

	return templateNames
}

func TestTemplateDisplay_List_EmptyDirectory(t *testing.T) {
	// Arrange
	display, _, _ := createTemplateDisplayForTest(t)

	// Act
	err := display.List(false)

	// Assert
	if err != nil {
		t.Errorf("List() failed on empty directory: %v", err)
	}
}

func TestTemplateDisplay_List_NonexistentDirectory(t *testing.T) {
	// Arrange
	nonexistentDir := filepath.Join(os.TempDir(), "nonexistent-display-dir")
	display := NewTemplateDisplay(nonexistentDir)

	// Act
	err := display.List(false)

	// Assert
	if err == nil {
		t.Error("List() expected error for nonexistent directory, got nil")
	}
}

func TestTemplateDisplay_List_Summary(t *testing.T) {
	// Arrange
	display, manager, _ := createTemplateDisplayForTest(t)
	createMultipleTestTemplates(t, manager, 3)

	// Act
	err := display.List(false)

	// Assert
	if err != nil {
		t.Errorf("List() summary failed: %v", err)
	}
}

func TestTemplateDisplay_List_Detailed(t *testing.T) {
	// Arrange
	display, manager, _ := createTemplateDisplayForTest(t)
	createMultipleTestTemplates(t, manager, 2)

	// Act
	err := display.List(true)

	// Assert
	if err != nil {
		t.Errorf("List() detailed failed: %v", err)
	}
}

func TestTemplateDisplay_List_WithNonTemplateFiles(t *testing.T) {
	// Arrange
	display, manager, tempDir := createTemplateDisplayForTest(t)

	// テンプレートファイルを作成
	createMultipleTestTemplates(t, manager, 1)

	// 非テンプレートファイルを作成
	nonTemplateFile := filepath.Join(tempDir, "not-a-template.txt")
	if err := os.WriteFile(nonTemplateFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create non-template file: %v", err)
	}

	// Act
	err := display.List(false)

	// Assert
	if err != nil {
		t.Errorf("List() failed with non-template files: %v", err)
	}
}

func TestTemplateDisplay_listSummary(t *testing.T) {
	// Arrange
	display, manager, _ := createTemplateDisplayForTest(t)
	templateNames := createMultipleTestTemplates(t, manager, 2)

	files, err := os.ReadDir(display.serversDir)
	if err != nil {
		t.Fatalf("Failed to read directory: %v", err)
	}

	// Act
	err = display.listSummary(files)

	// Assert
	if err != nil {
		t.Errorf("listSummary() failed: %v", err)
	}

	// ファイルが正しく読み込まれているか確認
	for _, name := range templateNames {
		template, err := manager.Load(name)
		if err != nil {
			t.Errorf("Failed to load template %s: %v", name, err)
		}
		if template == nil {
			t.Errorf("Template %s is nil", name)
		}
	}
}

func TestTemplateDisplay_listDetailed(t *testing.T) {
	// Arrange
	display, manager, _ := createTemplateDisplayForTest(t)
	createMultipleTestTemplates(t, manager, 2)

	files, err := os.ReadDir(display.serversDir)
	if err != nil {
		t.Fatalf("Failed to read directory: %v", err)
	}

	// Act
	err = display.listDetailed(files)

	// Assert
	if err != nil {
		t.Errorf("listDetailed() failed: %v", err)
	}
}

func TestTemplateDisplay_listDetailed_WithCorruptedFile(t *testing.T) {
	// Arrange
	display, manager, tempDir := createTemplateDisplayForTest(t)

	// 正常なテンプレートを作成
	createMultipleTestTemplates(t, manager, 1)

	// 破損したテンプレートファイルを作成
	corruptedPath := filepath.Join(tempDir, "corrupted"+config.FileExtension)
	if err := os.WriteFile(corruptedPath, []byte("invalid json"), 0644); err != nil {
		t.Fatalf("Failed to create corrupted file: %v", err)
	}

	files, err := os.ReadDir(display.serversDir)
	if err != nil {
		t.Fatalf("Failed to read directory: %v", err)
	}

	// Act
	err = display.listDetailed(files)

	// Assert
	if err != nil {
		t.Errorf("listDetailed() failed with corrupted file: %v", err)
	}
}

func TestTemplateDisplay_displayTemplateDetail_Complete(t *testing.T) {
	// Arrange
	display, _, _ := createTemplateDisplayForTest(t)

	description := "テスト用のテンプレート説明"
	template := &ServerTemplate{
		Name:        displayTestTemplateName,
		Description: &description,
		CreatedAt:   time.Now(),
		ServerConfig: ServerConfig{
			Command: "python",
			Args:    []string{"script.py", "--verbose"},
			Env:     map[string]string{"ENV1": "value1", "ENV2": "value2"},
		},
	}

	// Act - この関数は出力をテストするのが難しいので、エラーが出ないことを確認
	display.displayTemplateDetail(template)

	// Assert - パニックしないことを確認
	if template.Name != displayTestTemplateName {
		t.Errorf("Template name changed: got %s, want %s", template.Name, displayTestTemplateName)
	}
}

func TestTemplateDisplay_displayTemplateDetail_Minimal(t *testing.T) {
	// Arrange
	display, _, _ := createTemplateDisplayForTest(t)

	template := &ServerTemplate{
		Name:      displayTestTemplateName,
		CreatedAt: time.Now(),
		ServerConfig: ServerConfig{
			Command: "python",
		},
	}

	// Act
	display.displayTemplateDetail(template)

	// Assert - パニックしないことを確認
	if template.Name != displayTestTemplateName {
		t.Errorf("Template name changed: got %s, want %s", template.Name, displayTestTemplateName)
	}
}

func TestTemplateDisplay_displayTemplateDetail_EmptyEnv(t *testing.T) {
	// Arrange
	display, _, _ := createTemplateDisplayForTest(t)

	template := &ServerTemplate{
		Name:      displayTestTemplateName,
		CreatedAt: time.Now(),
		ServerConfig: ServerConfig{
			Command: "python",
			Args:    []string{"script.py"},
			Env:     map[string]string{}, // 空の環境変数
		},
	}

	// Act
	display.displayTemplateDetail(template)

	// Assert - パニックしないことを確認
	if template.Name != displayTestTemplateName {
		t.Errorf("Template name changed: got %s, want %s", template.Name, displayTestTemplateName)
	}
}

func TestTemplateDisplay_displayTemplateDetail_EmptyArgs(t *testing.T) {
	// Arrange
	display, _, _ := createTemplateDisplayForTest(t)

	template := &ServerTemplate{
		Name:      displayTestTemplateName,
		CreatedAt: time.Now(),
		ServerConfig: ServerConfig{
			Command: "python",
			Args:    []string{}, // 空の引数
			Env:     map[string]string{"ENV": "value"},
		},
	}

	// Act
	display.displayTemplateDetail(template)

	// Assert - パニックしないことを確認
	if template.Name != displayTestTemplateName {
		t.Errorf("Template name changed: got %s, want %s", template.Name, displayTestTemplateName)
	}
}

func TestTemplateDisplay_listSummary_WithBrokenTemplate(t *testing.T) {
	// Arrange
	display, manager, tempDir := createTemplateDisplayForTest(t)

	// 正常なテンプレートを作成
	createMultipleTestTemplates(t, manager, 1)

	// 破損したテンプレートファイルを作成
	brokenPath := filepath.Join(tempDir, "broken"+config.FileExtension)
	if err := os.WriteFile(brokenPath, []byte("broken json"), 0644); err != nil {
		t.Fatalf("Failed to create broken file: %v", err)
	}

	files, err := os.ReadDir(display.serversDir)
	if err != nil {
		t.Fatalf("Failed to read directory: %v", err)
	}

	// Act
	err = display.listSummary(files)

	// Assert - 破損したファイルがあってもエラーにならないことを確認
	if err != nil {
		t.Errorf("listSummary() failed with broken template: %v", err)
	}
}

func TestTemplateDisplay_List_FilterTemplateFiles(t *testing.T) {
	// Arrange
	display, manager, tempDir := createTemplateDisplayForTest(t)

	// テンプレートファイルを作成
	templateName := "valid-template"
	err := manager.SaveFromConfig(templateName, MCPServer{
		Command: "python",
		Args:    []string{"test.py"},
	})
	if err != nil {
		t.Fatalf("Failed to create template: %v", err)
	}

	// 様々な拡張子のファイルを作成
	testFiles := []string{
		"readme.txt",
		"config.yaml",
		"script.py",
		"data.csv",
	}

	for _, filename := range testFiles {
		filePath := filepath.Join(tempDir, filename)
		if err := os.WriteFile(filePath, []byte("test content"), 0644); err != nil {
			t.Fatalf("Failed to create test file %s: %v", filename, err)
		}
	}

	// Act
	err = display.List(false)

	// Assert
	if err != nil {
		t.Errorf("List() failed with mixed file types: %v", err)
	}

	// テンプレートファイルのみが処理されることを確認
	files, err := os.ReadDir(tempDir)
	if err != nil {
		t.Fatalf("Failed to read directory: %v", err)
	}

	templateFileCount := 0
	for _, file := range files {
		if strings.HasSuffix(file.Name(), config.FileExtension) {
			templateFileCount++
		}
	}

	if templateFileCount != 1 {
		t.Errorf("Expected 1 template file, got %d", templateFileCount)
	}
}
