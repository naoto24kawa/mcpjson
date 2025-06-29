package mcpconfig

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/naoto24kawa/mcpconfig/internal/config"
	"github.com/naoto24kawa/mcpconfig/internal/server"
	"github.com/naoto24kawa/mcpconfig/internal/utils"
)

// MCPConfigManager handles MCP configuration file operations
type MCPConfigManager struct{}

// NewMCPConfigManager creates a new MCPConfigManager instance
func NewMCPConfigManager() *MCPConfigManager {
	return &MCPConfigManager{}
}

// Load loads an MCP configuration from file
func (m *MCPConfigManager) Load(mcpConfigPath string) (*server.MCPConfig, error) {
	mcpConfig := &server.MCPConfig{}
	if err := utils.LoadJSON(mcpConfigPath, mcpConfig); err != nil {
		return nil, fmt.Errorf("MCP設定ファイルの読み込みに失敗しました: %w", err)
	}
	return mcpConfig, nil
}

// Save saves an MCP configuration to file
func (m *MCPConfigManager) Save(mcpConfig *server.MCPConfig, targetPath string) error {
	targetDir := filepath.Dir(targetPath)
	if err := os.MkdirAll(targetDir, config.DefaultDirPerm); err != nil {
		return fmt.Errorf("ディレクトリの作成に失敗しました: %w", err)
	}

	if err := utils.SaveJSON(targetPath, mcpConfig); err != nil {
		return fmt.Errorf("MCP設定ファイルの保存に失敗しました: %w", err)
	}

	return nil
}

// BuildFromProfile builds an MCP configuration from profile and server templates
func (m *MCPConfigManager) BuildFromProfile(profile *ProfileData, serverManager *server.Manager) (*server.MCPConfig, error) {
	mcpConfig := &server.MCPConfig{
		McpServers: make(map[string]server.MCPServer),
	}

	for _, serverRef := range profile.Servers {
		serverTemplate, err := serverManager.Load(serverRef.Template)
		if err != nil {
			return nil, fmt.Errorf("サーバーテンプレート '%s' の読み込みに失敗しました: %w", serverRef.Template, err)
		}

		mcpServer := m.createMCPServer(serverTemplate, &serverRef)
		mcpConfig.McpServers[serverRef.Name] = mcpServer
	}

	return mcpConfig, nil
}

func (m *MCPConfigManager) createMCPServer(template *server.ServerTemplate, serverRef *ServerRef) server.MCPServer {
	mcpServer := server.MCPServer{
		Command:       template.ServerConfig.Command,
		Args:          template.ServerConfig.Args,
		Env:           make(map[string]string),
		Timeout:       template.ServerConfig.Timeout,
		EnvFile:       template.ServerConfig.EnvFile,
		TransportType: template.ServerConfig.TransportType,
	}

	// Copy template environment variables
	for k, v := range template.ServerConfig.Env {
		mcpServer.Env[k] = v
	}

	// Apply overrides
	for k, v := range serverRef.Overrides.Env {
		mcpServer.Env[k] = v
	}

	return mcpServer
}

// ProfileData represents profile data structure
type ProfileData struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	CreatedAt   time.Time   `json:"createdAt"`
	UpdatedAt   time.Time   `json:"updatedAt"`
	Servers     []ServerRef `json:"servers"`
}

// ServerRef represents a server reference in a profile
type ServerRef struct {
	Name      string           `json:"name"`
	Template  string           `json:"template"`
	Overrides ServerOverrides  `json:"overrides,omitempty"`
}

// ServerOverrides represents environment variable overrides
type ServerOverrides struct {
	Env map[string]string `json:"env,omitempty"`
}