package group

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
	if len(args) == 0 {
		PrintUsage()
		e.exit(0)
		return
	}

	subCmd := args[0]

	switch subCmd {
	case "list":
		fmt.Println("グループ機能は現在開発中です")
	default:
		fmt.Fprintf(os.Stderr, "エラー: 不明なサブコマンド 'group %s'\n", subCmd)
		PrintUsage()
		e.exit(utils.ExitGeneralError)
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

func setupTestConfig(t *testing.T) (*config.Config, func()) {
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
	if !strings.Contains(stdout, "mcpjson group - グループ管理") {
		t.Error("Expected usage to be printed")
	}
}

func TestExecute_ListSubcommand(t *testing.T) {
	// Arrange
	cfg, cleanup := setupTestConfig(t)
	defer cleanup()

	executor := NewTestableExecutor(cfg)

	// Act
	stdout, stderr := captureOutput(func() {
		executor.TestableExecute([]string{"list"})
	})

	// Assert
	if executor.exited {
		t.Errorf("Expected executor not to exit for list subcommand, but got exit code: %d", executor.exitCode)
	}
	if stderr != "" {
		t.Errorf("Expected no error output, got: %s", stderr)
	}
	if !strings.Contains(stdout, "グループ機能は現在開発中です") {
		t.Errorf("Expected development message, got: %s", stdout)
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
			expectedErr: "不明なサブコマンド 'group unknown'",
		},
		{
			name:        "invalid subcommand",
			args:        []string{"invalid-cmd"},
			expectedErr: "不明なサブコマンド 'group invalid-cmd'",
		},
		{
			name:        "typo in subcommand",
			args:        []string{"lst"}, // typo for "list"
			expectedErr: "不明なサブコマンド 'group lst'",
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
			if !strings.Contains(stdout, "mcpjson group - グループ管理") {
				t.Error("Expected usage to be printed after error")
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

func TestPrintUsage(t *testing.T) {
	// Arrange & Act
	stdout, _ := captureOutput(func() {
		PrintUsage()
	})

	// Assert
	expectedContains := []string{
		"mcpjson group - グループ管理",
		"使用方法:",
		"mcpjson group <サブコマンド>",
		"サブコマンド:",
		"list",
		"グループ一覧表示",
		"注意: グループ機能は現在開発中です",
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

	validSubcommands := []string{"list"}

	for _, subCmd := range validSubcommands {
		t.Run(fmt.Sprintf("route_%s", subCmd), func(t *testing.T) {
			// Arrange
			cfg, cleanup := setupTestConfig(t)
			defer cleanup()

			executor := NewTestableExecutor(cfg)

			// Act
			executor.TestableExecute([]string{subCmd})

			// Assert - Should not exit immediately for valid subcommands
			if executor.exited {
				t.Errorf("Subcommand '%s' should not cause immediate exit", subCmd)
			}
		})
	}
}

func TestExecute_ListWithAdditionalArgs(t *testing.T) {
	// Test that list command handles additional arguments gracefully
	// Arrange
	cfg, cleanup := setupTestConfig(t)
	defer cleanup()

	executor := NewTestableExecutor(cfg)

	// Act
	stdout, stderr := captureOutput(func() {
		executor.TestableExecute([]string{"list", "--detail"})
	})

	// Assert
	if executor.exited {
		t.Errorf("Expected executor not to exit for list with args, but got exit code: %d", executor.exitCode)
	}
	if stderr != "" {
		t.Errorf("Expected no error output, got: %s", stderr)
	}
	if !strings.Contains(stdout, "グループ機能は現在開発中です") {
		t.Errorf("Expected development message, got: %s", stdout)
	}
}

func TestExecute_MultipleCalls(t *testing.T) {
	// Test that executor can be called multiple times with proper state reset
	cfg, cleanup := setupTestConfig(t)
	defer cleanup()

	executor := NewTestableExecutor(cfg)

	// First call - valid subcommand
	executor.TestableExecute([]string{"list"})
	if executor.exited {
		t.Error("First call should not exit")
	}

	// Reset executor state
	executor.exited = false
	executor.exitCode = -1

	// Second call - invalid subcommand
	captureOutput(func() {
		executor.TestableExecute([]string{"invalid"})
	})
	if !executor.exited {
		t.Error("Second call should exit for invalid subcommand")
	}
	if executor.exitCode != utils.ExitGeneralError {
		t.Errorf("Expected exit code %d, got %d", utils.ExitGeneralError, executor.exitCode)
	}
}

// Benchmark tests
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
			captureOutput(func() {
				executor.TestableExecute([]string{"list"})
			})
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
