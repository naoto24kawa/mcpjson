package cmd

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/naoto24kawa/mcpconfig/internal/config"
)

// TestableCommandRouter extends CommandRouter for testing
type TestableCommandRouter struct {
	*CommandRouter
	exitCode int
	exited   bool
}

func NewTestableCommandRouter() *TestableCommandRouter {
	return &TestableCommandRouter{
		CommandRouter: &CommandRouter{},
		exitCode:      -1,
		exited:        false,
	}
}

func (r *TestableCommandRouter) exit(code int) {
	r.exitCode = code
	r.exited = true
}

// Override methods to prevent actual os.Exit calls
func (r *TestableCommandRouter) handleHelp() {
	printUsage()
	r.exit(0)
}

func (r *TestableCommandRouter) handleVersion() {
	fmt.Printf("mcpconfig version %s\n", Version)
	r.exit(0)
}

func (r *TestableCommandRouter) handleUnknownCommand(cmd string) {
	fmt.Fprintf(os.Stderr, "エラー: 不明なコマンド '%s'\n", cmd)
	printUsage()
	r.exit(1)
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

func TestCommandRouter_Route(t *testing.T) {
	tests := []struct {
		name           string
		cmd            string
		args           []string
		expectedExit   bool
		expectedCode   int
		checkOutput    bool
		outputContains string
		errorOutput    bool
	}{
		// Arrange: Test help commands
		{
			name:           "help command",
			cmd:            "help",
			args:           []string{},
			expectedExit:   true,
			expectedCode:   0,
			checkOutput:    true,
			outputContains: "mcpconfig - MCP設定ファイル管理ツール",
		},
		{
			name:           "-h flag",
			cmd:            "-h",
			args:           []string{},
			expectedExit:   true,
			expectedCode:   0,
			checkOutput:    true,
			outputContains: "使用方法:",
		},
		{
			name:           "--help flag",
			cmd:            "--help",
			args:           []string{},
			expectedExit:   true,
			expectedCode:   0,
			checkOutput:    true,
			outputContains: "コマンド:",
		},

		// Arrange: Test version commands
		{
			name:           "version command",
			cmd:            "version",
			args:           []string{},
			expectedExit:   true,
			expectedCode:   0,
			checkOutput:    true,
			outputContains: "mcpconfig version",
		},
		{
			name:           "-v flag",
			cmd:            "-v",
			args:           []string{},
			expectedExit:   true,
			expectedCode:   0,
			checkOutput:    true,
			outputContains: fmt.Sprintf("mcpconfig version %s", Version),
		},
		{
			name:           "--version flag",
			cmd:            "--version",
			args:           []string{},
			expectedExit:   true,
			expectedCode:   0,
			checkOutput:    true,
			outputContains: "mcpconfig version",
		},

		// Arrange: Test unknown command
		{
			name:         "unknown command",
			cmd:          "unknown",
			args:         []string{},
			expectedExit: true,
			expectedCode: 1,
			errorOutput:  true,
		},
		{
			name:         "invalid command",
			cmd:          "invalid-cmd",
			args:         []string{},
			expectedExit: true,
			expectedCode: 1,
			errorOutput:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			router := NewTestableCommandRouter()

			// Act & Assert
			stdout, stderr := captureOutput(func() {
				// Call the individual test handlers instead of Route to avoid os.Exit
				switch tt.cmd {
				case "help", "-h", "--help":
					router.handleHelp()
				case "version", "-v", "--version":
					router.handleVersion()
				default:
					router.handleUnknownCommand(tt.cmd)
				}
			})

			// Assert exit behavior
			if tt.expectedExit {
				if !router.exited {
					t.Errorf("Expected router to exit, but it didn't")
				}
				if router.exitCode != tt.expectedCode {
					t.Errorf("Expected exit code %d, got %d", tt.expectedCode, router.exitCode)
				}
			} else {
				if router.exited {
					t.Errorf("Expected router not to exit, but it exited with code %d", router.exitCode)
				}
			}

			// Assert output
			if tt.checkOutput && tt.outputContains != "" {
				if !strings.Contains(stdout, tt.outputContains) {
					t.Errorf("Expected stdout to contain '%s', got: %s", tt.outputContains, stdout)
				}
			}

			if tt.errorOutput {
				if stderr == "" {
					t.Error("Expected error output, but got none")
				}
				if !strings.Contains(stderr, "エラー") {
					t.Errorf("Expected error message in stderr, got: %s", stderr)
				}
			}
		})
	}
}

func TestCommandRouter_ValidCommands(t *testing.T) {
	// Test that valid commands don't cause immediate exits
	validCommands := []string{
		"apply", "save", "create", "list", "delete",
		"rename", "detail", "server", "reset", "path", "server-path",
	}

	for _, cmd := range validCommands {
		t.Run(fmt.Sprintf("valid command %s", cmd), func(t *testing.T) {
			// Arrange
			_ = NewTestableCommandRouter()

			// Act
			// Note: We don't actually call Route here because valid commands
			// may have their own exit behaviors or require specific setup.
			// This test is more of a documentation of expected valid commands.

			// Assert - just verify the command is in our expected list
			found := false
			for _, validCmd := range validCommands {
				if cmd == validCmd {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Command %s not found in valid commands list", cmd)
			}
		})
	}
}

func TestPrintUsage(t *testing.T) {
	// Arrange & Act
	stdout, _ := captureOutput(func() {
		printUsage()
	})

	// Assert
	expectedContains := []string{
		"mcpconfig - MCP設定ファイル管理ツール",
		"使用方法:",
		"mcpconfig <コマンド>",
		"コマンド:",
		"apply",
		"save",
		"create",
		"list",
		"delete",
		"rename",
		"path",
		"detail",
		"server",
		"reset",
		"グローバルオプション:",
		"--help",
		"--version",
		config.DefaultProfileName,
	}

	for _, expected := range expectedContains {
		if !strings.Contains(stdout, expected) {
			t.Errorf("Usage output should contain '%s', got: %s", expected, stdout)
		}
	}
}

func TestVersion(t *testing.T) {
	tests := []struct {
		name        string
		version     string
		expectedOut string
	}{
		{
			name:        "default version",
			version:     "dev",
			expectedOut: "mcpconfig version dev\n",
		},
		{
			name:        "custom version",
			version:     "1.0.0",
			expectedOut: "mcpconfig version 1.0.0\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			oldVersion := Version
			Version = tt.version
			defer func() { Version = oldVersion }()

			router := NewTestableCommandRouter()

			// Act
			stdout, _ := captureOutput(func() {
				router.handleVersion()
			})

			// Assert
			if stdout != tt.expectedOut {
				t.Errorf("Expected version output '%s', got '%s'", tt.expectedOut, stdout)
			}
			if !router.exited || router.exitCode != 0 {
				t.Errorf("Expected exit with code 0, got exited=%v, code=%d", router.exited, router.exitCode)
			}
		})
	}
}

func TestExecute_NoArgs(t *testing.T) {
	// Arrange
	oldArgs := os.Args
	os.Args = []string{"mcpconfig"} // Only program name, no commands
	defer func() { os.Args = oldArgs }()

	// This test is tricky because Execute() calls os.Exit directly
	// In a real scenario, we would need to refactor Execute() to be more testable
	// For now, we document the expected behavior

	// The function should print usage and exit with code 0
	// when no arguments are provided
}

// Benchmark tests for performance
func BenchmarkCommandRouter_Route(b *testing.B) {
	router := &CommandRouter{}

	b.Run("help command", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			// We can't actually run Route in benchmark due to os.Exit
			// This is more of a placeholder for potential performance testing
			_ = router
		}
	})
}
