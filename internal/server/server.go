package server

import (
	"fmt"
	"time"

	"github.com/naoto24kawa/mcpconfig/internal/utils"
)

const (
	ListColumnWidth     = 20
	TimestampFormat     = "2006-01-02 15:04:05"
)

type ServerTemplate struct {
	Name         string       `json:"name"`
	Description  *string      `json:"description"`
	CreatedAt    time.Time    `json:"createdAt"`
	ServerConfig ServerConfig `json:"serverConfig"`
}

type ServerConfig struct {
	Command       string            `json:"command"`
	Args          []string          `json:"args,omitempty"`
	Env           map[string]string `json:"env,omitempty"`
	Timeout       *int              `json:"timeout,omitempty"`
	EnvFile       *string           `json:"envFile,omitempty"`
	TransportType *string           `json:"transportType,omitempty"`
}

type MCPServer struct {
	Command       string            `json:"command"`
	Args          []string          `json:"args,omitempty"`
	Env           map[string]string `json:"env,omitempty"`
	Timeout       *int              `json:"timeout,omitempty"`
	EnvFile       *string           `json:"envFile,omitempty"`
	TransportType *string           `json:"transportType,omitempty"`
}

type MCPConfig struct {
	McpServers map[string]MCPServer `json:"mcpServers"`
}

// Manager is a unified manager for server templates
type Manager struct {
	templateManager *TemplateManager
	templateUpdater *TemplateUpdater
	templateDisplay *TemplateDisplay
}

// NewManager creates a new unified Manager instance
func NewManager(serversDir string) *Manager {
	templateManager := NewTemplateManager(serversDir)
	return &Manager{
		templateManager: templateManager,
		templateUpdater: NewTemplateUpdater(templateManager),
		templateDisplay: NewTemplateDisplay(serversDir),
	}
}

// SaveFromFile saves a server template from an MCP config file
func (m *Manager) SaveFromFile(templateName, serverName, mcpConfigPath string, force bool) error {
	return m.templateManager.SaveFromFile(templateName, serverName, mcpConfigPath, force)
}

// SaveManual saves or updates a server template manually
func (m *Manager) SaveManual(templateName, command string, args []string, env map[string]string, force bool) error {
	return m.templateUpdater.SaveManual(templateName, command, args, env, force)
}


// List displays all server templates
func (m *Manager) List(detail bool) error {
	return m.templateDisplay.List(detail)
}

// Delete deletes a server template
func (m *Manager) Delete(name string, force bool) error {
	return m.templateManager.Delete(name, force)
}

// Rename renames a server template
func (m *Manager) Rename(oldName, newName string, force bool) error {
	return m.templateManager.Rename(oldName, newName, force)
}

func (m *Manager) Show(mcpConfigPath string, serverName string) error {
	mcpConfig := &MCPConfig{}
	if err := utils.LoadJSON(mcpConfigPath, mcpConfig); err != nil {
		return fmt.Errorf("MCP設定ファイルの読み込みに失敗しました: %w", err)
	}
	
	if serverName != "" {
		server, exists := mcpConfig.McpServers[serverName]
		if !exists {
			availableServers := make([]string, 0, len(mcpConfig.McpServers))
			for name := range mcpConfig.McpServers {
				availableServers = append(availableServers, name)
			}
			return fmt.Errorf("MCPサーバー '%s' がMCP設定ファイルに見つかりません\nファイル: %s\n利用可能なサーバー: %v", 
				serverName, mcpConfigPath, availableServers)
		}
		
		fmt.Printf("サーバー名: %s\n", serverName)
		fmt.Printf("  コマンド: %s\n", server.Command)
		if len(server.Args) > 0 {
			fmt.Printf("  引数: %v\n", server.Args)
		}
		if len(server.Env) > 0 {
			fmt.Println("  環境変数:")
			for k, v := range server.Env {
				fmt.Printf("    %s: %s\n", k, v)
			}
		}
	} else {
		if len(mcpConfig.McpServers) == 0 {
			fmt.Println("MCPサーバーが設定されていません")
			return nil
		}
		
		for name, server := range mcpConfig.McpServers {
			fmt.Printf("サーバー名: %s\n", name)
			fmt.Printf("  コマンド: %s\n", server.Command)
			if len(server.Args) > 0 {
				fmt.Printf("  引数: %v\n", server.Args)
			}
			if len(server.Env) > 0 {
				fmt.Println("  環境変数:")
				for k, v := range server.Env {
					fmt.Printf("    %s: %s\n", k, v)
				}
			}
			fmt.Println()
		}
	}
	
	return nil
}

// Load loads a server template by name
func (m *Manager) Load(name string) (*ServerTemplate, error) {
	return m.templateManager.Load(name)
}

// Exists checks if a server template exists
func (m *Manager) Exists(name string) (bool, error) {
	return m.templateManager.Exists(name)
}

// SaveFromConfig saves a server template from MCPServer config
func (m *Manager) SaveFromConfig(name string, server MCPServer) error {
	return m.templateManager.SaveFromConfig(name, server)
}

// Reset deletes all server templates
func (m *Manager) Reset(force bool) error {
	return m.templateManager.Reset(force)
}

// GetTemplatePath returns the file path for a server template
func (m *Manager) GetTemplatePath(name string) (string, error) {
	return m.templateManager.GetTemplatePath(name)
}

