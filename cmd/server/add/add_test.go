package add

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
	"github.com/naoto24kawa/mcpjson/internal/testutil"
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
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "エラー: テンプレート名が指定されていません")
		e.exit(utils.ExitArgumentError)
		return
	}

	templateName := args[0]
	var mcpConfigPath, serverName, envStr string

	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "--to", "-t":
			if i+1 >= len(args) {
				fmt.Fprintln(os.Stderr, "エラー: --to オプションに値が指定されていません")
				e.exit(utils.ExitArgumentError)
				return
			}
			mcpConfigPath = args[i+1]
			i++
		case "--as", "-a":
			if i+1 >= len(args) {
				fmt.Fprintln(os.Stderr, "エラー: --as オプションに値が指定されていません")
				e.exit(utils.ExitArgumentError)
				return
			}
			serverName = args[i+1]
			i++
		case "--env", "-e":
			if i+1 >= len(args) {
				fmt.Fprintln(os.Stderr, "エラー: --env オプションに値が指定されていません")
				e.exit(utils.ExitArgumentError)
				return
			}
			envStr = args[i+1]
			i++
		}
	}

	if mcpConfigPath == "" {
		mcpConfigPath = "./.mcp.json"
	}

	if err := utils.ValidateName(templateName, "サーバーテンプレート"); err != nil {
		fmt.Fprintln(os.Stderr, "エラー:", err)
		e.exit(utils.ExitArgumentError)
		return
	}

	envOverrides := make(map[string]string)
	if envStr != "" {
		parsedEnv, err := utils.ParseEnvVars(envStr)
		if err != nil {
			fmt.Fprintln(os.Stderr, "エラー:", err)
			e.exit(utils.ExitArgumentError)
			return
		}
		envOverrides = parsedEnv
	}

	serverManager := server.NewManager(e.cfg.ServersDir)
	if err := serverManager.AddToMCPConfig(mcpConfigPath, templateName, serverName, envOverrides); err != nil {
		fmt.Fprintln(os.Stderr, "エラー:", err)
		e.exit(utils.ExitGeneralError)
		return
	}
}

func captureOutput(f func()) (stdout, stderr string) {
	oldStdout := os.Stdout
	oldStderr := os.Stderr

	stdoutR, stdoutW, _ := os.Pipe()
	stderrR, stderrW, _ := os.Pipe()

	os.Stdout = stdoutW
	os.Stderr = stderrW

	stdoutCh := make(chan string)
	stderrCh := make(chan string)

	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, stdoutR)
		stdoutCh <- buf.String()
	}()

	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, stderrR)
		stderrCh <- buf.String()
	}()

	f()

	stdoutW.Close()
	stderrW.Close()

	stdout = <-stdoutCh
	stderr = <-stderrCh

	os.Stdout = oldStdout
	os.Stderr = oldStderr

	return stdout, stderr
}

func setupTestEnvironment(t *testing.T) (*config.Config, string, func()) {
	t.Helper()
	tempDir, cfg, cleanup := testutil.SetupIsolatedTestEnvironment(t)
	return cfg, tempDir, cleanup
}

func createTestTemplate(t *testing.T, cfg *config.Config, templateName string) string {
	t.Helper()

	// Generate unique template name
	uniqueName := testutil.GenerateUniqueServerName(templateName)

	serverManager := server.NewManager(cfg.ServersDir)
	err := serverManager.SaveManual(uniqueName, "python", []string{"-m", "test"},
		map[string]string{"TEST_ENV": "value"}, false)
	if err != nil {
		t.Fatalf("Failed to create test template: %v", err)
	}
	
	return uniqueName
}

func TestExecute_NoArgs(t *testing.T) {
	// Arrange
	cfg, _, cleanup := setupTestEnvironment(t)
	defer cleanup()

	executor := NewTestableExecutor(cfg)

	// Act
	_, stderr := captureOutput(func() {
		executor.TestableExecute([]string{})
	})

	// Assert
	if !executor.exited {
		t.Error("Expected executor to exit when no args provided")
	}
	if executor.exitCode != utils.ExitArgumentError {
		t.Errorf("Expected exit code %d, got %d", utils.ExitArgumentError, executor.exitCode)
	}
	if !strings.Contains(stderr, "テンプレート名が指定されていません") {
		t.Errorf("Expected template name error in stderr, got: %s", stderr)
	}
}

func TestExecute_ValidTemplate(t *testing.T) {
	// Arrange
	cfg, tempDir, cleanup := setupTestEnvironment(t)
	defer cleanup()

	templateName := createTestTemplate(t, cfg, "test-template")

	mcpConfigPath := filepath.Join(tempDir, "test-mcp.json")
	executor := NewTestableExecutor(cfg)

	// Act
	stdout, stderr := captureOutput(func() {
		executor.TestableExecute([]string{templateName, "--to", mcpConfigPath})
	})

	// Assert
	if executor.exited {
		t.Errorf("Expected no exit for valid template, but exited with code %d. Stderr: %s", executor.exitCode, stderr)
	}

	// Verify MCP config file was created
	if _, err := os.Stat(mcpConfigPath); os.IsNotExist(err) {
		t.Error("Expected MCP config file to be created")
	}

	// Verify no error output
	if stderr != "" && !strings.Contains(stdout, "追加しました") {
		t.Errorf("Unexpected stderr output: %s", stderr)
	}
}

func TestExecute_MissingOptionValues(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		expectedErr string
		exitCode    int
	}{
		{
			name:        "missing --to value",
			args:        []string{"template", "--to"},
			expectedErr: "--to オプションに値が指定されていません",
			exitCode:    utils.ExitArgumentError,
		},
		{
			name:        "missing --as value",
			args:        []string{"template", "--as"},
			expectedErr: "--as オプションに値が指定されていません",
			exitCode:    utils.ExitArgumentError,
		},
		{
			name:        "missing --env value",
			args:        []string{"template", "--env"},
			expectedErr: "--env オプションに値が指定されていません",
			exitCode:    utils.ExitArgumentError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			cfg, _, cleanup := setupTestEnvironment(t)
			defer cleanup()

			executor := NewTestableExecutor(cfg)

			// Act
			_, stderr := captureOutput(func() {
				executor.TestableExecute(tt.args)
			})

			// Assert
			if !executor.exited {
				t.Error("Expected executor to exit for missing option value")
			}
			if executor.exitCode != tt.exitCode {
				t.Errorf("Expected exit code %d, got %d", tt.exitCode, executor.exitCode)
			}
			if !strings.Contains(stderr, tt.expectedErr) {
				t.Errorf("Expected error '%s' in stderr, got: %s", tt.expectedErr, stderr)
			}
		})
	}
}

func TestExecute_DefaultMCPPath(t *testing.T) {
	// Arrange
	cfg, tempDir, cleanup := setupTestEnvironment(t)
	defer cleanup()

	templateName := createTestTemplate(t, cfg, "default-path-template")

	// Change to temp directory so default ./.mcp.json is created there
	oldWd, _ := os.Getwd()
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}
	defer func() {
		if err := os.Chdir(oldWd); err != nil {
			t.Errorf("Failed to restore directory: %v", err)
		}
	}()

	executor := NewTestableExecutor(cfg)

	// Act
	captureOutput(func() {
		executor.TestableExecute([]string{templateName})
	})

	// Assert
	if executor.exited {
		t.Errorf("Expected no exit, but exited with code %d", executor.exitCode)
	}

	// Verify default MCP config file was created
	defaultPath := "./.mcp.json"
	if _, err := os.Stat(defaultPath); os.IsNotExist(err) {
		t.Error("Expected default MCP config file to be created at ./.mcp.json")
	}
}

func TestExecute_WithServerName(t *testing.T) {
	// Arrange
	cfg, tempDir, cleanup := setupTestEnvironment(t)
	defer cleanup()

	templateName := createTestTemplate(t, cfg, "test-template")
	serverName := testutil.GenerateUniqueServerName("custom-server-name")

	mcpConfigPath := filepath.Join(tempDir, "custom-mcp.json")
	executor := NewTestableExecutor(cfg)

	// Act
	captureOutput(func() {
		executor.TestableExecute([]string{templateName, "--to", mcpConfigPath, "--as", serverName})
	})

	// Assert
	if executor.exited {
		t.Errorf("Expected no exit, but exited with code %d", executor.exitCode)
	}

	// Verify MCP config file contains the custom server name
	if _, err := os.Stat(mcpConfigPath); os.IsNotExist(err) {
		t.Error("Expected MCP config file to be created")
	}

	// Load and verify content
	var mcpConfig server.MCPConfig
	if err := utils.LoadJSON(mcpConfigPath, &mcpConfig); err != nil {
		t.Fatalf("Failed to load MCP config: %v", err)
	}

	if _, exists := mcpConfig.McpServers[serverName]; !exists {
		t.Errorf("Expected server '%s' to exist in MCP config", serverName)
	}
}

func TestExecute_WithEnvironmentOverrides(t *testing.T) {
	// Arrange
	cfg, tempDir, cleanup := setupTestEnvironment(t)
	defer cleanup()

	templateName := createTestTemplate(t, cfg, "env-template")

	mcpConfigPath := filepath.Join(tempDir, "env-mcp.json")
	executor := NewTestableExecutor(cfg)

	// Act
	captureOutput(func() {
		executor.TestableExecute([]string{
			templateName,
			"--to", mcpConfigPath,
			"--env", "CUSTOM_ENV=custom_value,ANOTHER_ENV=another_value",
		})
	})

	// Assert
	if executor.exited {
		t.Errorf("Expected no exit, but exited with code %d", executor.exitCode)
	}

	// Verify environment variables in MCP config
	var mcpConfig server.MCPConfig
	if err := utils.LoadJSON(mcpConfigPath, &mcpConfig); err != nil {
		t.Fatalf("Failed to load MCP config: %v", err)
	}

	serverConfig, exists := mcpConfig.McpServers[templateName]
	if !exists {
		t.Error("Expected server to exist in MCP config")
	}

	if serverConfig.Env["CUSTOM_ENV"] != "custom_value" {
		t.Errorf("Expected CUSTOM_ENV=custom_value, got: %s", serverConfig.Env["CUSTOM_ENV"])
	}
	if serverConfig.Env["ANOTHER_ENV"] != "another_value" {
		t.Errorf("Expected ANOTHER_ENV=another_value, got: %s", serverConfig.Env["ANOTHER_ENV"])
	}
}

func TestExecute_InvalidTemplateName(t *testing.T) {
	tests := []struct {
		name         string
		templateName string
		expectedErr  string
	}{
		{
			name:         "empty template name",
			templateName: "",
			expectedErr:  "サーバーテンプレート名が指定されていません",
		},
		{
			name:         "template name with spaces",
			templateName: "template with spaces",
			expectedErr:  "使用できない文字が含まれています",
		},
		{
			name:         "template name with special chars",
			templateName: "template@#$",
			expectedErr:  "使用できない文字が含まれています",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			cfg, _, cleanup := setupTestEnvironment(t)
			defer cleanup()

			executor := NewTestableExecutor(cfg)

			// Act
			_, stderr := captureOutput(func() {
				executor.TestableExecute([]string{tt.templateName})
			})

			// Assert
			if !executor.exited {
				t.Error("Expected executor to exit for invalid template name")
			}
			if executor.exitCode != utils.ExitArgumentError {
				t.Errorf("Expected exit code %d, got %d", utils.ExitArgumentError, executor.exitCode)
			}
			if !strings.Contains(stderr, tt.expectedErr) {
				t.Errorf("Expected error '%s' in stderr, got: %s", tt.expectedErr, stderr)
			}
		})
	}
}

func TestExecute_InvalidEnvironmentFormat(t *testing.T) {
	// Arrange
	cfg, _, cleanup := setupTestEnvironment(t)
	defer cleanup()

	templateName := createTestTemplate(t, cfg, "test-template")

	executor := NewTestableExecutor(cfg)

	// Act
	_, stderr := captureOutput(func() {
		executor.TestableExecute([]string{templateName, "--env", "invalid_format"})
	})

	// Assert
	if !executor.exited {
		t.Error("Expected executor to exit for invalid env format")
	}
	if executor.exitCode != utils.ExitArgumentError {
		t.Errorf("Expected exit code %d, got %d", utils.ExitArgumentError, executor.exitCode)
	}
	if !strings.Contains(stderr, "環境変数の形式が不正です") {
		t.Errorf("Expected env format error in stderr, got: %s", stderr)
	}
}

func TestExecute_NonExistentTemplate(t *testing.T) {
	// Arrange
	cfg, tempDir, cleanup := setupTestEnvironment(t)
	defer cleanup()

	templateName := "non-existent-template"
	mcpConfigPath := filepath.Join(tempDir, "test-mcp.json")
	executor := NewTestableExecutor(cfg)

	// Act
	_, stderr := captureOutput(func() {
		executor.TestableExecute([]string{templateName, "--to", mcpConfigPath})
	})

	// Assert
	if !executor.exited {
		t.Error("Expected executor to exit for non-existent template")
	}
	if executor.exitCode != utils.ExitGeneralError {
		t.Errorf("Expected exit code %d, got %d", utils.ExitGeneralError, executor.exitCode)
	}
	if !strings.Contains(stderr, "エラー:") {
		t.Errorf("Expected error in stderr, got: %s", stderr)
	}
}

func TestExecute_ShortFormOptions(t *testing.T) {
	// Arrange
	cfg, tempDir, cleanup := setupTestEnvironment(t)
	defer cleanup()

	templateName := createTestTemplate(t, cfg, "short-options-template")

	mcpConfigPath := filepath.Join(tempDir, "short-mcp.json")
	serverName := testutil.GenerateUniqueServerName("short-server")
	executor := NewTestableExecutor(cfg)

	// Act
	captureOutput(func() {
		executor.TestableExecute([]string{
			templateName,
			"-t", mcpConfigPath,
			"-a", serverName,
			"-e", "SHORT_ENV=short_value",
		})
	})

	// Assert
	if executor.exited {
		t.Errorf("Expected no exit, but exited with code %d", executor.exitCode)
	}

	// Verify results
	var mcpConfig server.MCPConfig
	if err := utils.LoadJSON(mcpConfigPath, &mcpConfig); err != nil {
		t.Fatalf("Failed to load MCP config: %v", err)
	}

	serverConfig, exists := mcpConfig.McpServers[serverName]
	if !exists {
		t.Error("Expected server to exist in MCP config")
	}

	if serverConfig.Env["SHORT_ENV"] != "short_value" {
		t.Errorf("Expected SHORT_ENV=short_value, got: %s", serverConfig.Env["SHORT_ENV"])
	}
}

// Benchmark helper functions
func setupBenchEnvironment(b *testing.B) (*config.Config, string, func()) {
	b.Helper()
	tempDir, cfg, cleanup := testutil.SetupIsolatedTestEnvironment(&testing.T{})
	return cfg, tempDir, cleanup
}

func createTestBenchTemplate(b *testing.B, cfg *config.Config, templateName string) string {
	b.Helper()

	// Generate unique template name
	uniqueName := testutil.GenerateUniqueServerName(templateName)

	serverManager := server.NewManager(cfg.ServersDir)
	err := serverManager.SaveManual(uniqueName, "python", []string{"-m", "test"},
		map[string]string{"TEST_ENV": "value"}, false)
	if err != nil {
		b.Fatalf("Failed to create bench template: %v", err)
	}
	
	return uniqueName
}

// Benchmark tests
func BenchmarkExecute_ValidTemplate(b *testing.B) {
	cfg, tempDir, cleanup := setupBenchEnvironment(b)
	defer cleanup()

	templateName := createTestBenchTemplate(b, cfg, "bench-template")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mcpConfigPath := filepath.Join(tempDir, fmt.Sprintf("bench-mcp-%d.json", i))
		executor := NewTestableExecutor(cfg)

		captureOutput(func() {
			executor.TestableExecute([]string{templateName, "--to", mcpConfigPath})
		})

		if executor.exited {
			b.Fatalf("Benchmark failed: executor exited with code %d", executor.exitCode)
		}
	}
}
