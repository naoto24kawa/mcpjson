package server

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/naoto24kawa/mcpconfig/internal/profile"
	"github.com/naoto24kawa/mcpconfig/internal/utils"
)

type Manager struct {
	serversDir string
}

func NewManager(serversDir string) *Manager {
	return &Manager{
		serversDir: serversDir,
	}
}

func (m *Manager) SaveFromFile(templateName, serverName, mcpConfigPath string, force bool) error {
	if !force && m.exists(templateName) {
		if !utils.ConfirmOverwrite("サーバーテンプレート", templateName) {
			return fmt.Errorf("上書きをキャンセルしました")
		}
	}
	
	mcpConfig := &profile.MCPConfig{}
	if err := utils.LoadJSON(mcpConfigPath, mcpConfig); err != nil {
		return fmt.Errorf("MCP設定ファイルの読み込みに失敗しました: %w", err)
	}
	
	server, exists := mcpConfig.McpServers[serverName]
	if !exists {
		return fmt.Errorf("MCPサーバー '%s' がMCP設定ファイルに見つかりません", serverName)
	}
	
	template := &profile.ServerTemplate{
		Name:        templateName,
		Description: nil,
		CreatedAt:   time.Now(),
		ServerConfig: profile.ServerConfig{
			Command: server.Command,
			Args:    server.Args,
			Env:     server.Env,
		},
	}
	
	if err := m.save(template); err != nil {
		return err
	}
	
	fmt.Printf("サーバーテンプレート '%s' を保存しました\n", templateName)
	fmt.Printf("コマンド: %s\n", server.Command)
	if len(server.Args) > 0 {
		fmt.Printf("引数: %v\n", server.Args)
	}
	
	return nil
}

func (m *Manager) SaveManual(templateName, command string, args []string, env map[string]string, force bool) error {
	existing := m.exists(templateName)
	
	if existing && !force {
		if !utils.ConfirmOverwrite("サーバーテンプレート", templateName) {
			return fmt.Errorf("上書きをキャンセルしました")
		}
	}
	
	var template *profile.ServerTemplate
	
	if existing {
		var err error
		template, err = m.Load(templateName)
		if err != nil {
			return err
		}
		
		if command != "" {
			template.ServerConfig.Command = command
		}
		
		if args != nil {
			if len(args) == 1 && args[0] == "" {
				template.ServerConfig.Args = nil
			} else {
				template.ServerConfig.Args = args
			}
		}
		
		if env != nil {
			if len(env) == 0 {
				template.ServerConfig.Env = nil
			} else {
				if template.ServerConfig.Env == nil {
					template.ServerConfig.Env = make(map[string]string)
				}
				for k, v := range env {
					if v == "" {
						delete(template.ServerConfig.Env, k)
					} else {
						template.ServerConfig.Env[k] = v
					}
				}
			}
		}
		
		fmt.Printf("サーバーテンプレート '%s' を更新しました\n", templateName)
	} else {
		if command == "" {
			return fmt.Errorf("コマンドが指定されていません")
		}
		
		template = &profile.ServerTemplate{
			Name:        templateName,
			Description: nil,
			CreatedAt:   time.Now(),
			ServerConfig: profile.ServerConfig{
				Command: command,
				Args:    args,
				Env:     env,
			},
		}
		
		fmt.Printf("サーバーテンプレート '%s' を作成しました\n", templateName)
	}
	
	if err := m.save(template); err != nil {
		return err
	}
	
	return nil
}

func (m *Manager) List(detail bool) error {
	files, err := os.ReadDir(m.serversDir)
	if err != nil {
		return fmt.Errorf("サーバーディレクトリの読み込みに失敗しました: %w", err)
	}
	
	if len(files) == 0 {
		fmt.Println("サーバーテンプレートが存在しません")
		return nil
	}
	
	if detail {
		for _, file := range files {
			if strings.HasSuffix(file.Name(), ".json") {
				name := strings.TrimSuffix(file.Name(), ".json")
				template, err := m.Load(name)
				if err != nil {
					fmt.Printf("エラー: %s の読み込みに失敗しました: %v\n", name, err)
					continue
				}
				
				fmt.Printf("\nテンプレート: %s\n", template.Name)
				if template.Description != nil {
					fmt.Printf("  説明: %s\n", *template.Description)
				}
				fmt.Printf("  作成日時: %s\n", template.CreatedAt.Format("2006-01-02 15:04:05"))
				fmt.Printf("  コマンド: %s\n", template.ServerConfig.Command)
				if len(template.ServerConfig.Args) > 0 {
					fmt.Printf("  引数: %v\n", template.ServerConfig.Args)
				}
				if len(template.ServerConfig.Env) > 0 {
					fmt.Println("  環境変数:")
					for k, v := range template.ServerConfig.Env {
						fmt.Printf("    %s=%s\n", k, v)
					}
				}
			}
		}
	} else {
		fmt.Printf("%-20s %-20s %s\n", "テンプレート名", "作成日時", "コマンド")
		fmt.Println(strings.Repeat("-", 60))
		
		for _, file := range files {
			if strings.HasSuffix(file.Name(), ".json") {
				name := strings.TrimSuffix(file.Name(), ".json")
				template, err := m.Load(name)
				if err != nil {
					continue
				}
				
				fmt.Printf("%-20s %-20s %s\n",
					template.Name,
					template.CreatedAt.Format("2006-01-02 15:04:05"),
					template.ServerConfig.Command)
			}
		}
	}
	
	return nil
}

func (m *Manager) Delete(name string, force bool) error {
	templatePath := filepath.Join(m.serversDir, name+".json")
	
	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		return fmt.Errorf("サーバーテンプレート '%s' が見つかりません", name)
	}
	
	if !force {
		fmt.Printf("サーバーテンプレート '%s' を削除しますか？ (y/N): ", name)
		var response string
		fmt.Scanln(&response)
		if strings.ToLower(response) != "y" {
			fmt.Println("削除をキャンセルしました")
			return nil
		}
	}
	
	if err := os.Remove(templatePath); err != nil {
		return fmt.Errorf("サーバーテンプレートの削除に失敗しました: %w", err)
	}
	
	fmt.Printf("サーバーテンプレート '%s' を削除しました\n", name)
	return nil
}

func (m *Manager) Rename(oldName, newName string, force bool) error {
	oldPath := filepath.Join(m.serversDir, oldName+".json")
	newPath := filepath.Join(m.serversDir, newName+".json")
	
	if _, err := os.Stat(oldPath); os.IsNotExist(err) {
		return fmt.Errorf("サーバーテンプレート '%s' が見つかりません", oldName)
	}
	
	if _, err := os.Stat(newPath); err == nil && !force {
		return fmt.Errorf("サーバーテンプレート '%s' は既に存在します\n別の名前を指定するか、--force オプションで上書きしてください", newName)
	}
	
	template, err := m.Load(oldName)
	if err != nil {
		return err
	}
	
	template.Name = newName
	
	if err := m.save(template); err != nil {
		return err
	}
	
	if err := os.Remove(oldPath); err != nil {
		return fmt.Errorf("古いサーバーテンプレートの削除に失敗しました: %w", err)
	}
	
	fmt.Printf("サーバーテンプレート '%s' を '%s' に変更しました\n", oldName, newName)
	return nil
}

func (m *Manager) Show(mcpConfigPath string, serverName string) error {
	mcpConfig := &profile.MCPConfig{}
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

func (m *Manager) Load(name string) (*profile.ServerTemplate, error) {
	templatePath := filepath.Join(m.serversDir, name+".json")
	
	template := &profile.ServerTemplate{}
	if err := utils.LoadJSON(templatePath, template); err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("サーバーテンプレート '%s' が見つかりません", name)
		}
		return nil, fmt.Errorf("サーバーテンプレートの読み込みに失敗しました: %w", err)
	}
	
	return template, nil
}

func (m *Manager) Exists(name string) (bool, error) {
	return m.exists(name), nil
}

func (m *Manager) SaveFromConfig(name string, server profile.MCPServer) error {
	template := &profile.ServerTemplate{
		Name:        name,
		Description: nil,
		CreatedAt:   time.Now(),
		ServerConfig: profile.ServerConfig{
			Command: server.Command,
			Args:    server.Args,
			Env:     server.Env,
		},
	}
	
	return m.save(template)
}

func (m *Manager) exists(name string) bool {
	templatePath := filepath.Join(m.serversDir, name+".json")
	_, err := os.Stat(templatePath)
	return err == nil
}

func (m *Manager) save(template *profile.ServerTemplate) error {
	templatePath := filepath.Join(m.serversDir, template.Name+".json")
	return utils.SaveJSON(templatePath, template)
}