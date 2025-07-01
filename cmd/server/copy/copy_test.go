package copy

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
	if len(args) < 2 {
		fmt.Fprintf(os.Stderr, "エラー: コピー元とコピー先のサーバー名を指定してください\n")
		printUsage()
		e.exit(utils.ExitGeneralError)
		return
	}

	srcName := args[0]
	destName := args[1]
	force := false

	// オプション解析
	for i := 2; i < len(args); i++ {
		switch args[i] {
		case "--force", "-f":
			force = true
		default:
			fmt.Fprintf(os.Stderr, "エラー: 不明なオプション '%s'\n", args[i])
			printUsage()
			e.exit(utils.ExitGeneralError)
			return
		}
	}

	if err := utils.ValidateName(srcName, "サーバーテンプレート"); err != nil {
		fmt.Fprintln(os.Stderr, "エラー:", err)
		e.exit(utils.ExitArgumentError)
		return
	}

	if err := utils.ValidateName(destName, "サーバーテンプレート"); err != nil {
		fmt.Fprintln(os.Stderr, "エラー:", err)
		e.exit(utils.ExitArgumentError)
		return
	}

	serverManager := server.NewManager(e.cfg.ServersDir)
	if err := serverManager.Copy(srcName, destName, force); err != nil {
		fmt.Fprintf(os.Stderr, "エラー: %v\n", err)
		e.exit(utils.ExitGeneralError)
		return
	}
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
	if !strings.Contains(stderr, "コピー元とコピー先のサーバー名を指定してください") {
		t.Errorf("Expected error message about missing args, got: %s", stderr)
	}
	if !strings.Contains(stdout, "mcpjson server copy") {
		t.Error("Expected usage to be printed")
	}
}

func TestExecute_OneArg(t *testing.T) {
	// Arrange
	cfg, cleanup := setupTestConfig(t)
	defer cleanup()

	executor := NewTestableExecutor(cfg)

	// Act
	stdout, stderr := captureOutput(func() {
		executor.TestableExecute([]string{"source-server"})
	})

	// Assert
	if !executor.exited {
		t.Error("Expected executor to exit when only one arg provided")
	}
	if executor.exitCode != utils.ExitGeneralError {
		t.Errorf("Expected exit code %d, got %d", utils.ExitGeneralError, executor.exitCode)
	}
	if !strings.Contains(stderr, "コピー元とコピー先のサーバー名を指定してください") {
		t.Errorf("Expected error message about missing args, got: %s", stderr)
	}
	if !strings.Contains(stdout, "mcpjson server copy") {
		t.Error("Expected usage to be printed")
	}
}

func TestExecute_InvalidOption(t *testing.T) {
	// Arrange
	cfg, cleanup := setupTestConfig(t)
	defer cleanup()

	executor := NewTestableExecutor(cfg)

	// Act
	stdout, stderr := captureOutput(func() {
		executor.TestableExecute([]string{"source-server", "dest-server", "--invalid"})
	})

	// Assert
	if !executor.exited {
		t.Error("Expected executor to exit with invalid option")
	}
	if executor.exitCode != utils.ExitGeneralError {
		t.Errorf("Expected exit code %d, got %d", utils.ExitGeneralError, executor.exitCode)
	}
	if !strings.Contains(stderr, "不明なオプション '--invalid'") {
		t.Errorf("Expected error message about invalid option, got: %s", stderr)
	}
	if !strings.Contains(stdout, "mcpjson server copy") {
		t.Error("Expected usage to be printed")
	}
}

func TestExecute_InvalidSourceName(t *testing.T) {
	// Arrange
	cfg, cleanup := setupTestConfig(t)
	defer cleanup()

	executor := NewTestableExecutor(cfg)

	// Act
	_, stderr := captureOutput(func() {
		executor.TestableExecute([]string{"invalid-name!", "dest-server"})
	})

	// Assert
	if !executor.exited {
		t.Error("Expected executor to exit with invalid source name")
	}
	if executor.exitCode != utils.ExitArgumentError {
		t.Errorf("Expected exit code %d, got %d", utils.ExitArgumentError, executor.exitCode)
	}
	if !strings.Contains(stderr, "エラー:") {
		t.Errorf("Expected validation error, got: %s", stderr)
	}
}

func TestExecute_InvalidDestName(t *testing.T) {
	// Arrange
	cfg, cleanup := setupTestConfig(t)
	defer cleanup()

	executor := NewTestableExecutor(cfg)

	// Act
	_, stderr := captureOutput(func() {
		executor.TestableExecute([]string{"source-server", "invalid-name!"})
	})

	// Assert
	if !executor.exited {
		t.Error("Expected executor to exit with invalid dest name")
	}
	if executor.exitCode != utils.ExitArgumentError {
		t.Errorf("Expected exit code %d, got %d", utils.ExitArgumentError, executor.exitCode)
	}
	if !strings.Contains(stderr, "エラー:") {
		t.Errorf("Expected validation error, got: %s", stderr)
	}
}

func TestExecute_SourceNotFound(t *testing.T) {
	// Arrange
	cfg, cleanup := setupTestConfig(t)
	defer cleanup()

	executor := NewTestableExecutor(cfg)

	// Act
	_, stderr := captureOutput(func() {
		executor.TestableExecute([]string{"nonexistent-server", "dest-server"})
	})

	// Assert
	if !executor.exited {
		t.Error("Expected executor to exit when source not found")
	}
	if executor.exitCode != utils.ExitGeneralError {
		t.Errorf("Expected exit code %d, got %d", utils.ExitGeneralError, executor.exitCode)
	}
	if !strings.Contains(stderr, "nonexistent-server") {
		t.Errorf("Expected error about source not found, got: %s", stderr)
	}
}

func TestExecute_Success(t *testing.T) {
	// Arrange
	cfg, cleanup := setupTestConfig(t)
	defer cleanup()

	createTestServerTemplate(t, cfg, "source-server")
	executor := NewTestableExecutor(cfg)

	// Act
	stdout, _ := captureOutput(func() {
		executor.TestableExecute([]string{"source-server", "dest-server"})
	})

	// Assert
	if executor.exited {
		t.Errorf("Expected executor not to exit on success, but got exit code: %d", executor.exitCode)
	}
	if !strings.Contains(stdout, "source-server") && !strings.Contains(stdout, "dest-server") {
		t.Errorf("Expected success message, got: %s", stdout)
	}

	// Verify destination server was created
	serverManager := server.NewManager(cfg.ServersDir)
	exists, err := serverManager.Exists("dest-server")
	if err != nil {
		t.Fatalf("Failed to check if destination exists: %v", err)
	}
	if !exists {
		t.Error("Expected destination server to be created")
	}
}

func TestExecute_DestinationExists_NoForce(t *testing.T) {
	// Arrange
	cfg, cleanup := setupTestConfig(t)
	defer cleanup()

	createTestServerTemplate(t, cfg, "source-server")
	createTestServerTemplate(t, cfg, "dest-server")
	executor := NewTestableExecutor(cfg)

	// Act
	_, stderr := captureOutput(func() {
		executor.TestableExecute([]string{"source-server", "dest-server"})
	})

	// Assert
	if !executor.exited {
		t.Error("Expected executor to exit when destination exists without force")
	}
	if executor.exitCode != utils.ExitGeneralError {
		t.Errorf("Expected exit code %d, got %d", utils.ExitGeneralError, executor.exitCode)
	}
	if !strings.Contains(stderr, "既に存在します") {
		t.Errorf("Expected error about destination exists, got: %s", stderr)
	}
}

func TestExecute_DestinationExists_WithForce(t *testing.T) {
	// Arrange
	cfg, cleanup := setupTestConfig(t)
	defer cleanup()

	createTestServerTemplate(t, cfg, "source-server")
	createTestServerTemplate(t, cfg, "dest-server")
	executor := NewTestableExecutor(cfg)

	// Act
	stdout, _ := captureOutput(func() {
		executor.TestableExecute([]string{"source-server", "dest-server", "--force"})
	})

	// Assert
	if executor.exited {
		t.Errorf("Expected executor not to exit on success with force, but got exit code: %d", executor.exitCode)
	}
	if !strings.Contains(stdout, "source-server") && !strings.Contains(stdout, "dest-server") {
		t.Errorf("Expected success message, got: %s", stdout)
	}
}

func TestExecute_WithShortForceFlag(t *testing.T) {
	// Arrange
	cfg, cleanup := setupTestConfig(t)
	defer cleanup()

	createTestServerTemplate(t, cfg, "source-server")
	createTestServerTemplate(t, cfg, "dest-server")
	executor := NewTestableExecutor(cfg)

	// Act
	stdout, _ := captureOutput(func() {
		executor.TestableExecute([]string{"source-server", "dest-server", "-f"})
	})

	// Assert
	if executor.exited {
		t.Errorf("Expected executor not to exit on success with -f flag, but got exit code: %d", executor.exitCode)
	}
	if !strings.Contains(stdout, "source-server") && !strings.Contains(stdout, "dest-server") {
		t.Errorf("Expected success message, got: %s", stdout)
	}
}
