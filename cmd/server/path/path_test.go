package path

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/naoto24kawa/mcpjson/internal/config"
	"github.com/naoto24kawa/mcpjson/internal/server"
	"github.com/naoto24kawa/mcpjson/internal/utils"
)

// TestableExecutor wraps Execute for testing
type TestableExecutor struct {
	cfg      *config.Config
	exitCode int
	exited   bool
}

func NewTestableExecutor(cfg *config.Config) *TestableExecutor {
	return &TestableExecutor{
		cfg:      cfg,
		exitCode: -1,
		exited:   false,
	}
}

func (e *TestableExecutor) exit(code int) {
	e.exitCode = code
	e.exited = true
}

func (e *TestableExecutor) TestableExecute(args []string) {
	if len(args) != 1 {
		fmt.Fprintf(os.Stderr, "エラー: サーバーテンプレート名を指定してください\n")
		printUsage()
		e.exit(utils.ExitGeneralError)
		return
	}

	templateName := args[0]
	serverManager := server.NewManager(e.cfg.ServersDir)
	templatePath, err := serverManager.GetTemplatePath(templateName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "エラー: サーバーテンプレートパスの取得に失敗しました: %v\n", err)
		e.exit(utils.ExitGeneralError)
		return
	}
	fmt.Print(templatePath)
}

func setupTestConfig(t *testing.T) (*config.Config, func()) {
	tempDir := t.TempDir()
	cfg := &config.Config{
		BaseDir:     tempDir,
		ProfilesDir: filepath.Join(tempDir, "profiles"),
		ServersDir:  filepath.Join(tempDir, "servers"),
		GroupsDir:   filepath.Join(tempDir, "groups"),
	}

	// Create necessary directories
	if err := os.MkdirAll(cfg.ServersDir, 0755); err != nil {
		t.Fatalf("Failed to create servers dir: %v", err)
	}

	return cfg, func() {
		// t.TempDir() handles cleanup automatically
	}
}

func captureOutput(fn func()) (string, string) {
	oldStdout := os.Stdout
	oldStderr := os.Stderr

	rOut, wOut, _ := os.Pipe()
	rErr, wErr, _ := os.Pipe()

	os.Stdout = wOut
	os.Stderr = wErr

	fn()

	wOut.Close()
	wErr.Close()

	os.Stdout = oldStdout
	os.Stderr = oldStderr

	var bufOut, bufErr bytes.Buffer
	io.Copy(&bufOut, rOut)
	io.Copy(&bufErr, rErr)

	return bufOut.String(), bufErr.String()
}

func createTestServerTemplate(t *testing.T, cfg *config.Config, name string) {
	serverManager := server.NewManager(cfg.ServersDir)
	testServer := server.MCPServer{
		Command: "test-command",
		Args:    []string{"test-arg"},
	}
	if err := serverManager.SaveFromConfig(name, testServer); err != nil {
		t.Fatalf("Failed to create test server template: %v", err)
	}
}

func TestExecute_NoArgs(t *testing.T) {
	// Arrange
	cfg, cleanup := setupTestConfig(t)
	defer cleanup()

	executor := NewTestableExecutor(cfg)

	// Act
	stdout, stderr := captureOutput(func() {
		executor.TestableExecute([]string{})
	})

	// Assert
	if !executor.exited {
		t.Error("Expected executor to exit when no args provided")
	}
	if executor.exitCode != utils.ExitGeneralError {
		t.Errorf("Expected exit code %d, got %d", utils.ExitGeneralError, executor.exitCode)
	}
	if !strings.Contains(stderr, "サーバーテンプレート名を指定してください") {
		t.Errorf("Expected error message about missing template name, got: %s", stderr)
	}
	if !strings.Contains(stdout, "mcpjson server path") {
		t.Error("Expected usage to be printed")
	}
}

func TestExecute_MultipleArgs(t *testing.T) {
	// Arrange
	cfg, cleanup := setupTestConfig(t)
	defer cleanup()

	executor := NewTestableExecutor(cfg)

	// Act
	stdout, stderr := captureOutput(func() {
		executor.TestableExecute([]string{"template1", "template2"})
	})

	// Assert
	if !executor.exited {
		t.Error("Expected executor to exit when multiple args provided")
	}
	if executor.exitCode != utils.ExitGeneralError {
		t.Errorf("Expected exit code %d, got %d", utils.ExitGeneralError, executor.exitCode)
	}
	if !strings.Contains(stderr, "サーバーテンプレート名を指定してください") {
		t.Errorf("Expected error message about args, got: %s", stderr)
	}
	if !strings.Contains(stdout, "mcpjson server path") {
		t.Error("Expected usage to be printed")
	}
}

func TestExecute_NonExistentTemplate(t *testing.T) {
	// Arrange
	cfg, cleanup := setupTestConfig(t)
	defer cleanup()

	executor := NewTestableExecutor(cfg)

	// Act
	_, stderr := captureOutput(func() {
		executor.TestableExecute([]string{"nonexistent-template"})
	})

	// Assert
	if !executor.exited {
		t.Error("Expected executor to exit when template not found")
	}
	if executor.exitCode != utils.ExitGeneralError {
		t.Errorf("Expected exit code %d, got %d", utils.ExitGeneralError, executor.exitCode)
	}
	if !strings.Contains(stderr, "サーバーテンプレートパスの取得に失敗しました") {
		t.Errorf("Expected error about template not found, got: %s", stderr)
	}
}

func TestExecute_ValidTemplate(t *testing.T) {
	// Arrange
	cfg, cleanup := setupTestConfig(t)
	defer cleanup()

	templateName := "test-template"
	createTestServerTemplate(t, cfg, templateName)
	executor := NewTestableExecutor(cfg)

	// Act
	stdout, stderr := captureOutput(func() {
		executor.TestableExecute([]string{templateName})
	})

	// Assert
	if executor.exited {
		t.Errorf("Expected executor not to exit on success, but got exit code: %d", executor.exitCode)
	}
	if stderr != "" {
		t.Errorf("Expected no error output, got: %s", stderr)
	}

	// Verify the path is printed
	expectedPath := filepath.Join(cfg.ServersDir, templateName+".jsonc")
	if stdout != expectedPath {
		t.Errorf("Expected path '%s', got: %s", expectedPath, stdout)
	}

	// Verify the template file actually exists at the returned path
	if _, err := os.Stat(stdout); os.IsNotExist(err) {
		t.Errorf("Template file does not exist at returned path: %s", stdout)
	}
}

func TestExecute_EmptyTemplateName(t *testing.T) {
	// Arrange
	cfg, cleanup := setupTestConfig(t)
	defer cleanup()

	executor := NewTestableExecutor(cfg)

	// Act
	_, stderr := captureOutput(func() {
		executor.TestableExecute([]string{""})
	})

	// Assert
	if !executor.exited {
		t.Error("Expected executor to exit with empty template name")
	}
	if executor.exitCode != utils.ExitGeneralError {
		t.Errorf("Expected exit code %d, got %d", utils.ExitGeneralError, executor.exitCode)
	}
	// Empty template name will be handled by GetTemplatePath as "template not found"
	if !strings.Contains(stderr, "サーバーテンプレートパスの取得に失敗しました") {
		t.Errorf("Expected error message about template path retrieval failure, got: %s", stderr)
	}
}

func TestExecute_WhitespaceTemplateName(t *testing.T) {
	// Arrange
	cfg, cleanup := setupTestConfig(t)
	defer cleanup()

	executor := NewTestableExecutor(cfg)

	// Act
	_, stderr := captureOutput(func() {
		executor.TestableExecute([]string{"   "})
	})

	// Assert
	if !executor.exited {
		t.Error("Expected executor to exit with whitespace template name")
	}
	if executor.exitCode != utils.ExitGeneralError {
		t.Errorf("Expected exit code %d, got %d", utils.ExitGeneralError, executor.exitCode)
	}
	if !strings.Contains(stderr, "サーバーテンプレートパスの取得に失敗しました") {
		t.Errorf("Expected error about template not found, got: %s", stderr)
	}
}

func TestExecute_SpecialCharacterTemplate(t *testing.T) {
	// Arrange
	cfg, cleanup := setupTestConfig(t)
	defer cleanup()

	executor := NewTestableExecutor(cfg)

	// Act
	_, stderr := captureOutput(func() {
		executor.TestableExecute([]string{"template@#$"})
	})

	// Assert
	if !executor.exited {
		t.Error("Expected executor to exit with special character template name")
	}
	if executor.exitCode != utils.ExitGeneralError {
		t.Errorf("Expected exit code %d, got %d", utils.ExitGeneralError, executor.exitCode)
	}
	if !strings.Contains(stderr, "サーバーテンプレートパスの取得に失敗しました") {
		t.Errorf("Expected error about template not found, got: %s", stderr)
	}
}

func TestExecute_PathOutput(t *testing.T) {
	// Arrange
	cfg, cleanup := setupTestConfig(t)
	defer cleanup()

	templateName := "path-output-test"
	createTestServerTemplate(t, cfg, templateName)
	executor := NewTestableExecutor(cfg)

	// Act
	stdout, _ := captureOutput(func() {
		executor.TestableExecute([]string{templateName})
	})

	// Assert
	expectedPath := filepath.Join(cfg.ServersDir, templateName+".jsonc")
	if stdout != expectedPath {
		t.Errorf("Expected exact path '%s', got: '%s'", expectedPath, stdout)
	}

	// Verify no extra newlines or whitespace
	if strings.Contains(stdout, "\n") {
		t.Error("Expected no newlines in path output")
	}
	if stdout != strings.TrimSpace(stdout) {
		t.Error("Expected no leading/trailing whitespace in path output")
	}
}

func TestExecute_MultipleTemplates(t *testing.T) {
	// Test with multiple templates to ensure correct path resolution
	// Arrange
	cfg, cleanup := setupTestConfig(t)
	defer cleanup()

	templates := []string{"template-a", "template-b", "template-c"}
	for _, name := range templates {
		createTestServerTemplate(t, cfg, name)
	}

	executor := NewTestableExecutor(cfg)

	// Act & Assert for each template
	for _, templateName := range templates {
		t.Run(templateName, func(t *testing.T) {
			executor.exited = false // Reset for each test
			executor.exitCode = -1

			stdout, stderr := captureOutput(func() {
				executor.TestableExecute([]string{templateName})
			})

			if executor.exited {
				t.Errorf("Expected executor not to exit for template %s, but got exit code: %d", templateName, executor.exitCode)
			}
			if stderr != "" {
				t.Errorf("Expected no error output for template %s, got: %s", templateName, stderr)
			}

			expectedPath := filepath.Join(cfg.ServersDir, templateName+".jsonc")
			if stdout != expectedPath {
				t.Errorf("Expected path '%s' for template %s, got: %s", expectedPath, templateName, stdout)
			}
		})
	}
}

// Edge case tests for path construction
func TestExecute_PathConstructionEdgeCases(t *testing.T) {
	tests := []struct {
		name         string
		templateName string
		shouldCreate bool
	}{
		{
			name:         "template with numbers",
			templateName: "template123",
			shouldCreate: true,
		},
		{
			name:         "template with hyphens",
			templateName: "template-with-hyphens",
			shouldCreate: true,
		},
		{
			name:         "template with underscores",
			templateName: "template_with_underscores",
			shouldCreate: true,
		},
		{
			name:         "single character template",
			templateName: "a",
			shouldCreate: true,
		},
		{
			name:         "long template name",
			templateName: "very-long-template-name-with-many-characters",
			shouldCreate: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			cfg, cleanup := setupTestConfig(t)
			defer cleanup()

			if tt.shouldCreate {
				createTestServerTemplate(t, cfg, tt.templateName)
			}

			executor := NewTestableExecutor(cfg)

			// Act
			stdout, stderr := captureOutput(func() {
				executor.TestableExecute([]string{tt.templateName})
			})

			// Assert
			if tt.shouldCreate {
				if executor.exited {
					t.Errorf("Expected executor not to exit for valid template, but got exit code: %d. Stderr: %s", executor.exitCode, stderr)
				}
				expectedPath := filepath.Join(cfg.ServersDir, tt.templateName+".jsonc")
				if stdout != expectedPath {
					t.Errorf("Expected path '%s', got: %s", expectedPath, stdout)
				}
			} else {
				if !executor.exited {
					t.Error("Expected executor to exit for invalid template")
				}
			}
		})
	}
}
