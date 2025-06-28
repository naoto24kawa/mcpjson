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
func (m *Manager) Delete(name string, force bool, profileManager ProfileManager) error {
	return m.templateManager.Delete(name, force, profileManager)
}

// Rename renames a server template
func (m *Manager) Rename(oldName, newName string, force bool) error {
	return m.templateManager.Rename(oldName, newName, force)
}

// AddToMCPConfig adds a server from template to an MCP config file
func (m *Manager) AddToMCPConfig(mcpConfigPath, templateName, serverName string, envOverrides map[string]string) error {
	// テンプレートを読み込む
	template, err := m.templateManager.Load(templateName)
	if err != nil {
		return fmt.Errorf("テンプレート '%s' の読み込みに失敗しました: %w", templateName, err)
	}
	
	// MCP設定ファイルを読み込む（存在しない場合は新規作成）
	mcpConfig := &MCPConfig{McpServers: make(map[string]MCPServer)}
	if utils.FileExists(mcpConfigPath) {
		if err := utils.LoadJSON(mcpConfigPath, mcpConfig); err != nil {
			return fmt.Errorf("MCP設定ファイルの読み込みに失敗しました: %w", err)
		}
	}
	
	if mcpConfig.McpServers == nil {
		mcpConfig.McpServers = make(map[string]MCPServer)
	}
	
	// サーバー名が未指定の場合はテンプレート名を使用
	if serverName == "" {
		serverName = templateName
	}
	
	// 既に存在するかチェック
	if _, exists := mcpConfig.McpServers[serverName]; exists {
		return fmt.Errorf("サーバー '%s' は既にMCP設定ファイルに存在します", serverName)
	}
	
	// 環境変数をマージ
	env := make(map[string]string)
	for k, v := range template.ServerConfig.Env {
		env[k] = v
	}
	for k, v := range envOverrides {
		env[k] = v
	}
	
	// MCPサーバー設定を作成
	mcpServer := MCPServer{
		Command:       template.ServerConfig.Command,
		Args:          template.ServerConfig.Args,
		Env:           env,
		Timeout:       template.ServerConfig.Timeout,
		EnvFile:       template.ServerConfig.EnvFile,
		TransportType: template.ServerConfig.TransportType,
	}
	
	// サーバーを追加
	mcpConfig.McpServers[serverName] = mcpServer
	
	// ファイルに保存
	if err := utils.SaveJSON(mcpConfigPath, mcpConfig); err != nil {
		return fmt.Errorf("MCP設定ファイルの保存に失敗しました: %w", err)
	}
	
	fmt.Printf("サーバー '%s' をMCP設定ファイルに追加しました: %s\n", serverName, mcpConfigPath)
	return nil
}

// RemoveFromMCPConfig removes a server from an MCP config file
func (m *Manager) RemoveFromMCPConfig(mcpConfigPath, serverName string) error {
	// MCP設定ファイルを読み込む
	mcpConfig := &MCPConfig{}
	if err := utils.LoadJSON(mcpConfigPath, mcpConfig); err != nil {
		return fmt.Errorf("MCP設定ファイルの読み込みに失敗しました: %w", err)
	}
	
	// サーバーが存在するかチェック
	if _, exists := mcpConfig.McpServers[serverName]; !exists {
		availableServers := make([]string, 0, len(mcpConfig.McpServers))
		for name := range mcpConfig.McpServers {
			availableServers = append(availableServers, name)
		}
		return fmt.Errorf("サーバー '%s' がMCP設定ファイルに見つかりません\nファイル: %s\n利用可能なサーバー: %v", 
			serverName, mcpConfigPath, availableServers)
	}
	
	// サーバーを削除
	delete(mcpConfig.McpServers, serverName)
	
	// ファイルに保存
	if err := utils.SaveJSON(mcpConfigPath, mcpConfig); err != nil {
		return fmt.Errorf("MCP設定ファイルの保存に失敗しました: %w", err)
	}
	
	fmt.Printf("サーバー '%s' をMCP設定ファイルから削除しました: %s\n", serverName, mcpConfigPath)
	return nil
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

