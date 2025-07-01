package mcpconfig

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/naoto24kawa/mcpjson/internal/server"
	"github.com/naoto24kawa/mcpjson/internal/utils"
)

func setupTestEnvironment(t *testing.T) (string, func()) {
	t.Helper()

	tempDir := t.TempDir()
	return tempDir, func() {}
}

func createTestMCPConfig() *server.MCPConfig {
	return &server.MCPConfig{
		McpServers: map[string]server.MCPServer{
			"test-server": {
				Command: "python",
				Args:    []string{"-m", "test"},
				Env:     map[string]string{"TEST_ENV": "value"},
			},
			"another-server": {
				Command: "node",
				Args:    []string{"server.js"},
				Env:     map[string]string{"NODE_ENV": "production"},
			},
		},
	}
}

func createTestServerTemplate() *server.ServerTemplate {
	return &server.ServerTemplate{
		Name:        "test-template",
		Description: stringPtr("Test template"),
		CreatedAt:   time.Now(),
		ServerConfig: server.ServerConfig{
			Command: "python",
			Args:    []string{"-m", "test"},
			Env:     map[string]string{"DEFAULT_ENV": "default"},
		},
	}
}

func createTestProfile() *ProfileData {
	return &ProfileData{
		Name:        "test-profile",
		Description: "Test profile",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Servers: []ServerRef{
			{
				Name:     "test-server",
				Template: "test-template",
				Overrides: ServerOverrides{
					Env: map[string]string{"OVERRIDE_ENV": "override"},
				},
			},
		},
	}
}

func stringPtr(s string) *string {
	return &s
}

func TestNewMCPConfigManager(t *testing.T) {
	// Arrange & Act
	manager := NewMCPConfigManager()

	// Assert
	if manager == nil {
		t.Error("NewMCPConfigManager() should return non-nil manager")
	}
}

func TestMCPConfigManager_Load(t *testing.T) {
	tests := []struct {
		name        string
		setupFile   func(filePath string) error
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid MCP config file",
			setupFile: func(filePath string) error {
				mcpConfig := createTestMCPConfig()
				return utils.SaveJSON(filePath, mcpConfig)
			},
			expectError: false,
		},
		{
			name: "non-existent file",
			setupFile: func(filePath string) error {
				return nil // Don't create file
			},
			expectError: true,
			errorMsg:    "MCP設定ファイルの読み込みに失敗しました",
		},
		{
			name: "invalid JSON file",
			setupFile: func(filePath string) error {
				return os.WriteFile(filePath, []byte("invalid json"), 0644)
			},
			expectError: true,
			errorMsg:    "MCP設定ファイルの読み込みに失敗しました",
		},
		{
			name: "empty file",
			setupFile: func(filePath string) error {
				return os.WriteFile(filePath, []byte(""), 0644)
			},
			expectError: true,
			errorMsg:    "MCP設定ファイルの読み込みに失敗しました",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			tempDir, cleanup := setupTestEnvironment(t)
			defer cleanup()

			manager := NewMCPConfigManager()
			filePath := filepath.Join(tempDir, "mcp-config.json")

			if tt.setupFile != nil {
				err := tt.setupFile(filePath)
				if err != nil && !tt.expectError {
					t.Fatalf("Failed to setup test file: %v", err)
				}
			}

			// Act
			mcpConfig, err := manager.Load(filePath)

			// Assert
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				if tt.errorMsg != "" && err != nil {
					if !contains(err.Error(), tt.errorMsg) {
						t.Errorf("Expected error to contain '%s', got: %s", tt.errorMsg, err.Error())
					}
				}
				if mcpConfig != nil {
					t.Error("Expected nil config when error occurs")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if mcpConfig == nil {
					t.Error("Expected non-nil config")
				}
				if mcpConfig.McpServers == nil {
					t.Error("Expected McpServers to be initialized")
				}
			}
		})
	}
}

func TestMCPConfigManager_Save(t *testing.T) {
	tests := []struct {
		name        string
		mcpConfig   *server.MCPConfig
		targetPath  string
		expectError bool
		errorMsg    string
	}{
		{
			name:       "valid save to new file",
			mcpConfig:  createTestMCPConfig(),
			targetPath: "mcp-config.json",
		},
		{
			name:       "save to nested directory",
			mcpConfig:  createTestMCPConfig(),
			targetPath: "nested/dir/mcp-config.json",
		},
		{
			name:        "nil config",
			mcpConfig:   nil,
			targetPath:  "mcp-config.json",
			expectError: false, // utils.SaveJSON handles nil gracefully by saving "null"
		},
		{
			name:       "empty config",
			mcpConfig:  &server.MCPConfig{McpServers: make(map[string]server.MCPServer)},
			targetPath: "empty-config.json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			tempDir, cleanup := setupTestEnvironment(t)
			defer cleanup()

			manager := NewMCPConfigManager()
			fullPath := filepath.Join(tempDir, tt.targetPath)

			// Act
			err := manager.Save(tt.mcpConfig, fullPath)

			// Assert
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				if tt.errorMsg != "" && err != nil {
					if !contains(err.Error(), tt.errorMsg) {
						t.Errorf("Expected error to contain '%s', got: %s", tt.errorMsg, err.Error())
					}
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}

				// Verify file was created
				if _, err := os.Stat(fullPath); os.IsNotExist(err) {
					t.Error("Expected file to be created")
				}

				// Verify directory was created if nested
				if dir := filepath.Dir(fullPath); dir != tempDir {
					if _, err := os.Stat(dir); os.IsNotExist(err) {
						t.Error("Expected nested directory to be created")
					}
				}

				// Verify content is valid JSON
				if tt.mcpConfig != nil {
					data, err := os.ReadFile(fullPath)
					if err != nil {
						t.Errorf("Failed to read saved file: %v", err)
					}

					var loadedConfig server.MCPConfig
					if err := json.Unmarshal(data, &loadedConfig); err != nil {
						t.Errorf("Saved file is not valid JSON: %v", err)
					}
				}
			}
		})
	}
}

func TestMCPConfigManager_BuildFromProfile(t *testing.T) {
	tests := []struct {
		name           string
		profile        *ProfileData
		setupManager   func() *server.Manager
		expectError    bool
		errorMsg       string
		expectedCount  int
		validateResult func(*server.MCPConfig) bool
	}{
		{
			name:    "valid profile with one server",
			profile: createTestProfile(),
			setupManager: func() *server.Manager {
				tempDir := t.TempDir()
				manager := server.NewManager(tempDir)

				// Create test template
				template := createTestServerTemplate()
				if err := utils.SaveJSON(filepath.Join(tempDir, "test-template.jsonc"), template); err != nil {
					t.Fatalf("Failed to save test template: %v", err)
				}

				return manager
			},
			expectedCount: 1,
			validateResult: func(config *server.MCPConfig) bool {
				if len(config.McpServers) != 1 {
					return false
				}
				srv, exists := config.McpServers["test-server"]
				if !exists {
					return false
				}
				return srv.Command == "python" && srv.Env["OVERRIDE_ENV"] == "override"
			},
		},
		{
			name: "profile with multiple servers",
			profile: &ProfileData{
				Name:        "multi-profile",
				Description: "Multi server profile",
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
				Servers: []ServerRef{
					{
						Name:     "server1",
						Template: "template1",
						Overrides: ServerOverrides{
							Env: map[string]string{"ENV1": "value1"},
						},
					},
					{
						Name:     "server2",
						Template: "template2",
						Overrides: ServerOverrides{
							Env: map[string]string{"ENV2": "value2"},
						},
					},
				},
			},
			setupManager: func() *server.Manager {
				tempDir := t.TempDir()
				manager := server.NewManager(tempDir)

				// Create multiple templates
				template1 := &server.ServerTemplate{
					Name:      "template1",
					CreatedAt: time.Now(),
					ServerConfig: server.ServerConfig{
						Command: "python",
						Args:    []string{"script1.py"},
						Env:     map[string]string{"BASE_ENV": "base1"},
					},
				}
				template2 := &server.ServerTemplate{
					Name:      "template2",
					CreatedAt: time.Now(),
					ServerConfig: server.ServerConfig{
						Command: "node",
						Args:    []string{"script2.js"},
						Env:     map[string]string{"BASE_ENV": "base2"},
					},
				}

				if err := utils.SaveJSON(filepath.Join(tempDir, "template1.jsonc"), template1); err != nil {
					t.Fatalf("Failed to save template1: %v", err)
				}
				if err := utils.SaveJSON(filepath.Join(tempDir, "template2.jsonc"), template2); err != nil {
					t.Fatalf("Failed to save template2: %v", err)
				}

				return manager
			},
			expectedCount: 2,
			validateResult: func(config *server.MCPConfig) bool {
				return len(config.McpServers) == 2 &&
					config.McpServers["server1"].Command == "python" &&
					config.McpServers["server2"].Command == "node"
			},
		},
		{
			name:    "profile with non-existent template",
			profile: createTestProfile(),
			setupManager: func() *server.Manager {
				tempDir := t.TempDir()
				return server.NewManager(tempDir) // No templates created
			},
			expectError: true,
			errorMsg:    "サーバーテンプレート 'test-template' の読み込みに失敗しました",
		},
		{
			name: "empty profile",
			profile: &ProfileData{
				Name:        "empty-profile",
				Description: "Empty profile",
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
				Servers:     []ServerRef{},
			},
			setupManager: func() *server.Manager {
				tempDir := t.TempDir()
				return server.NewManager(tempDir)
			},
			expectedCount: 0,
			validateResult: func(config *server.MCPConfig) bool {
				return len(config.McpServers) == 0
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			manager := NewMCPConfigManager()
			serverManager := tt.setupManager()

			// Act
			mcpConfig, err := manager.BuildFromProfile(tt.profile, serverManager)

			// Assert
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				if tt.errorMsg != "" && err != nil {
					if !contains(err.Error(), tt.errorMsg) {
						t.Errorf("Expected error to contain '%s', got: %s", tt.errorMsg, err.Error())
					}
				}
				if mcpConfig != nil {
					t.Error("Expected nil config when error occurs")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if mcpConfig == nil {
					t.Error("Expected non-nil config")
				}
				if len(mcpConfig.McpServers) != tt.expectedCount {
					t.Errorf("Expected %d servers, got %d", tt.expectedCount, len(mcpConfig.McpServers))
				}
				if tt.validateResult != nil && !tt.validateResult(mcpConfig) {
					t.Error("Result validation failed")
				}
			}
		})
	}
}

func TestMCPConfigManager_createMCPServer(t *testing.T) {
	tests := []struct {
		name        string
		template    *server.ServerTemplate
		serverRef   *ServerRef
		validateFn  func(server.MCPServer) bool
		description string
	}{
		{
			name:     "basic server creation",
			template: createTestServerTemplate(),
			serverRef: &ServerRef{
				Name:      "test-server",
				Template:  "test-template",
				Overrides: ServerOverrides{},
			},
			validateFn: func(srv server.MCPServer) bool {
				return srv.Command == "python" &&
					len(srv.Args) == 2 &&
					srv.Args[0] == "-m" &&
					srv.Args[1] == "test" &&
					srv.Env["DEFAULT_ENV"] == "default"
			},
			description: "should copy all template properties",
		},
		{
			name:     "server with environment overrides",
			template: createTestServerTemplate(),
			serverRef: &ServerRef{
				Name:     "override-server",
				Template: "test-template",
				Overrides: ServerOverrides{
					Env: map[string]string{
						"DEFAULT_ENV": "overridden",
						"NEW_ENV":     "new_value",
					},
				},
			},
			validateFn: func(srv server.MCPServer) bool {
				return srv.Env["DEFAULT_ENV"] == "overridden" &&
					srv.Env["NEW_ENV"] == "new_value"
			},
			description: "should apply environment overrides correctly",
		},
		{
			name: "server with complex configuration",
			template: &server.ServerTemplate{
				Name:      "complex-template",
				CreatedAt: time.Now(),
				ServerConfig: server.ServerConfig{
					Command:       "node",
					Args:          []string{"server.js", "--port", "3000"},
					Env:           map[string]string{"NODE_ENV": "development", "PORT": "3000"},
					Timeout:       intPtr(30),
					EnvFile:       stringPtr(".env"),
					TransportType: stringPtr("stdio"),
				},
			},
			serverRef: &ServerRef{
				Name:     "complex-server",
				Template: "complex-template",
				Overrides: ServerOverrides{
					Env: map[string]string{"NODE_ENV": "production"},
				},
			},
			validateFn: func(srv server.MCPServer) bool {
				return srv.Command == "node" &&
					len(srv.Args) == 3 &&
					srv.Env["NODE_ENV"] == "production" &&
					srv.Env["PORT"] == "3000" &&
					*srv.Timeout == 30 &&
					*srv.EnvFile == ".env" &&
					*srv.TransportType == "stdio"
			},
			description: "should handle complex configurations with all fields",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			manager := NewMCPConfigManager()

			// Act
			mcpServer := manager.createMCPServer(tt.template, tt.serverRef)

			// Assert
			if !tt.validateFn(mcpServer) {
				t.Errorf("Validation failed: %s", tt.description)
				t.Logf("Resulting server: %+v", mcpServer)
			}

			// Verify environment map is properly initialized
			if mcpServer.Env == nil {
				t.Error("Environment map should be initialized")
			}
		})
	}
}

func intPtr(i int) *int {
	return &i
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr ||
		len(s) > len(substr) && s[len(s)-len(substr):] == substr ||
		len(s) > len(substr) &&
			func() bool {
				for i := 0; i <= len(s)-len(substr); i++ {
					if s[i:i+len(substr)] == substr {
						return true
					}
				}
				return false
			}()
}

// Benchmark tests
func BenchmarkMCPConfigManager_BuildFromProfile(b *testing.B) {
	tempDir := b.TempDir()
	manager := NewMCPConfigManager()
	serverManager := server.NewManager(tempDir)

	// Setup test template
	template := createTestServerTemplate()
	if err := utils.SaveJSON(filepath.Join(tempDir, "test-template.jsonc"), template); err != nil {
		b.Fatalf("Failed to save test template: %v", err)
	}

	profile := createTestProfile()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := manager.BuildFromProfile(profile, serverManager)
		if err != nil {
			b.Fatalf("Benchmark failed: %v", err)
		}
	}
}

func BenchmarkMCPConfigManager_createMCPServer(b *testing.B) {
	manager := NewMCPConfigManager()
	template := createTestServerTemplate()
	serverRef := &ServerRef{
		Name:     "bench-server",
		Template: "bench-template",
		Overrides: ServerOverrides{
			Env: map[string]string{"BENCH_ENV": "value"},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = manager.createMCPServer(template, serverRef)
	}
}
