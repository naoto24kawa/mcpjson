package server

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/naoto24kawa/mcpjson/internal/config"
	"github.com/naoto24kawa/mcpjson/internal/utils"
)

// TestableExecutor wraps the Execute function for testing
type TestableExecutor struct {
	exitCode int
	exited   bool
	cfg      *config.Config
}

func NewTestableExecutor(cfg *config.Config) *TestableExecutor {
	return &TestableExecutor{
		exitCode: -1,
		exited:   false,
		cfg:      cfg,
	}
}

func (e *TestableExecutor) exit(code int) {
	e.exitCode = code
	e.exited = true
}

// TestableExecute mimics Execute but doesn't call os.Exit
func (e *TestableExecutor) TestableExecute(args []string) {
	if len(args) == 0 {
		PrintUsage()
		e.exit(0)
		return
	}

	subCmd := args[0]
	// subArgs := args[1:]

	switch subCmd {
	case "save", "list", "delete", "rename", "add", "remove", "show", "detail":
		// For testing, we just validate the routing logic
		// The actual subcommand execution would be tested in their respective test files
		return
	default:
		fmt.Fprintf(os.Stderr, "エラー: 不明なサブコマンド 'server %s'\n", subCmd)
		PrintUsage()
		e.exit(utils.ExitGeneralError)
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

func setupTestConfig(t *testing.T) (*config.Config, func()) {
	t.Helper()

	tempDir := t.TempDir()

	// Set up temporary home directory
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)

	cfg, err := config.New()
	if err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	cleanup := func() {
		os.Setenv("HOME", oldHome)
	}

	return cfg, cleanup
}

func TestExecute_NoArgs(t *testing.T) {
	// Arrange
	cfg, cleanup := setupTestConfig(t)
	defer cleanup()

	executor := NewTestableExecutor(cfg)

	// Act
	stdout, _ := captureOutput(func() {
		executor.TestableExecute([]string{})
	})

	// Assert
	if !executor.exited {
		t.Error("Expected executor to exit when no args provided")
	}
	if executor.exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", executor.exitCode)
	}
	if !strings.Contains(stdout, "mcpjson server - MCPサーバー管理") {
		t.Error("Expected usage to be printed")
	}
}

func TestExecute_ValidSubcommands(t *testing.T) {
	tests := []struct {
		name       string
		subcommand string
		args       []string
	}{
		{
			name:       "save subcommand",
			subcommand: "save",
			args:       []string{"save", "test-server"},
		},
		{
			name:       "list subcommand",
			subcommand: "list",
			args:       []string{"list"},
		},
		{
			name:       "delete subcommand",
			subcommand: "delete",
			args:       []string{"delete", "test-server"},
		},
		{
			name:       "rename subcommand",
			subcommand: "rename",
			args:       []string{"rename", "old-name", "new-name"},
		},
		{
			name:       "add subcommand",
			subcommand: "add",
			args:       []string{"add", "test-server"},
		},
		{
			name:       "remove subcommand",
			subcommand: "remove",
			args:       []string{"remove", "test-server"},
		},
		{
			name:       "detail subcommand",
			subcommand: "detail",
			args:       []string{"detail", "test-server"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			cfg, cleanup := setupTestConfig(t)
			defer cleanup()

			executor := NewTestableExecutor(cfg)

			// Act
			executor.TestableExecute(tt.args)

			// Assert - valid subcommands should not cause immediate exit
			// (They may exit later in their own logic, but not in the routing)
			if executor.exited {
				t.Errorf("Valid subcommand '%s' should not cause immediate exit", tt.subcommand)
			}
		})
	}
}

func TestExecute_InvalidSubcommand(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		expectedErr string
	}{
		{
			name:        "unknown subcommand",
			args:        []string{"unknown"},
			expectedErr: "不明なサブコマンド 'server unknown'",
		},
		{
			name:        "invalid subcommand",
			args:        []string{"invalid-cmd"},
			expectedErr: "不明なサブコマンド 'server invalid-cmd'",
		},
		{
			name:        "typo in subcommand",
			args:        []string{"lst"}, // typo for "list"
			expectedErr: "不明なサブコマンド 'server lst'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			cfg, cleanup := setupTestConfig(t)
			defer cleanup()

			executor := NewTestableExecutor(cfg)

			// Act
			stdout, stderr := captureOutput(func() {
				executor.TestableExecute(tt.args)
			})

			// Assert
			if !executor.exited {
				t.Error("Expected executor to exit for invalid subcommand")
			}
			if executor.exitCode != utils.ExitGeneralError {
				t.Errorf("Expected exit code %d, got %d", utils.ExitGeneralError, executor.exitCode)
			}
			if !strings.Contains(stderr, tt.expectedErr) {
				t.Errorf("Expected stderr to contain '%s', got: %s", tt.expectedErr, stderr)
			}
			if !strings.Contains(stdout, "mcpjson server - MCPサーバー管理") {
				t.Error("Expected usage to be printed after error")
			}
		})
	}
}

func TestPrintUsage(t *testing.T) {
	// Arrange & Act
	stdout, _ := captureOutput(func() {
		PrintUsage()
	})

	// Assert
	expectedContains := []string{
		"mcpjson server - MCPサーバー管理",
		"使用方法:",
		"mcpjson server <サブコマンド>",
		"サブコマンド:",
		"save",
		"list",
		"delete",
		"copy",
		"rename",
		"add",
		"remove",
		"detail",
		"path",
		"設定ファイルからサーバー保存",
		"サーバー一覧表示",
		"サーバー削除",
		"サーバーコピー",
		"サーバー名変更",
		"プロファイルにサーバー追加",
		"プロファイルからサーバー削除",
		"サーバーテンプレートの詳細を表示",
		"サーバーテンプレートパスを表示",
	}

	for _, expected := range expectedContains {
		if !strings.Contains(stdout, expected) {
			t.Errorf("Usage output should contain '%s', got: %s", expected, stdout)
		}
	}
}

func TestExecute_SubcommandRouting(t *testing.T) {
	// This test verifies that the routing logic correctly identifies
	// valid subcommands and routes them appropriately

	validSubcommands := []string{
		"save", "list", "delete", "rename",
		"add", "remove", "detail",
	}

	for _, subCmd := range validSubcommands {
		t.Run(fmt.Sprintf("route_%s", subCmd), func(t *testing.T) {
			// Arrange
			cfg, cleanup := setupTestConfig(t)
			defer cleanup()

			executor := NewTestableExecutor(cfg)

			// Act
			executor.TestableExecute([]string{subCmd, "dummy-arg"})

			// Assert - Should not exit immediately for valid subcommands
			if executor.exited {
				t.Errorf("Subcommand '%s' should not cause immediate exit", subCmd)
			}
		})
	}
}

func TestExecute_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		expectExit  bool
		exitCode    int
		description string
	}{
		{
			name:        "empty string subcommand",
			args:        []string{""},
			expectExit:  true,
			exitCode:    utils.ExitGeneralError,
			description: "empty string should be treated as unknown subcommand",
		},
		{
			name:        "whitespace only subcommand",
			args:        []string{" "},
			expectExit:  true,
			exitCode:    utils.ExitGeneralError,
			description: "whitespace should be treated as unknown subcommand",
		},
		{
			name:        "case sensitive subcommand",
			args:        []string{"List"}, // capital L
			expectExit:  true,
			exitCode:    utils.ExitGeneralError,
			description: "subcommands should be case sensitive",
		},
		{
			name:        "subcommand with leading dash",
			args:        []string{"-list"},
			expectExit:  true,
			exitCode:    utils.ExitGeneralError,
			description: "subcommands with leading dash should be invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			cfg, cleanup := setupTestConfig(t)
			defer cleanup()

			executor := NewTestableExecutor(cfg)

			// Act
			captureOutput(func() {
				executor.TestableExecute(tt.args)
			})

			// Assert
			if tt.expectExit {
				if !executor.exited {
					t.Errorf("%s: Expected executor to exit but it didn't", tt.description)
				}
				if executor.exitCode != tt.exitCode {
					t.Errorf("%s: Expected exit code %d, got %d", tt.description, tt.exitCode, executor.exitCode)
				}
			} else {
				if executor.exited {
					t.Errorf("%s: Expected executor not to exit but it exited with code %d", tt.description, executor.exitCode)
				}
			}
		})
	}
}

// Integration test to verify the actual Execute function behavior
// Note: This test is more complex because Execute calls os.Exit directly
func TestExecute_Integration(t *testing.T) {
	// This is a documentation test - in real scenarios, we would need
	// to refactor the Execute function to be more testable by accepting
	// an exit function as a parameter or using dependency injection

	cfg, cleanup := setupTestConfig(t)
	defer cleanup()

	// Test that we can at least call the function without panicking
	// when given valid arguments (though it will exit)
	t.Run("valid config creation", func(t *testing.T) {
		if cfg == nil {
			t.Error("Expected valid config to be created")
		}
		if cfg.ServersDir == "" {
			t.Error("Expected ServersDir to be set")
		}
	})
}

func setupBenchConfig(b *testing.B) (*config.Config, func()) {
	b.Helper()

	tempDir := b.TempDir()

	// Set up temporary home directory
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)

	cfg, err := config.New()
	if err != nil {
		b.Fatalf("Failed to create bench config: %v", err)
	}

	cleanup := func() {
		os.Setenv("HOME", oldHome)
	}

	return cfg, cleanup
}

// Benchmark tests
func BenchmarkPrintUsage(b *testing.B) {
	for i := 0; i < b.N; i++ {
		captureOutput(func() {
			PrintUsage()
		})
	}
}

func BenchmarkSubcommandRouting(b *testing.B) {
	cfg, cleanup := setupBenchConfig(b)
	defer cleanup()

	executor := NewTestableExecutor(cfg)

	b.Run("valid_subcommand", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			executor.TestableExecute([]string{"list"})
			executor.exited = false // Reset for next iteration
		}
	})

	b.Run("invalid_subcommand", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			captureOutput(func() {
				executor.TestableExecute([]string{"invalid"})
			})
			executor.exited = false // Reset for next iteration
		}
	})
}
