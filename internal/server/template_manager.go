package server

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/naoto24kawa/mcpconfig/internal/config"
	"github.com/naoto24kawa/mcpconfig/internal/interaction"
	"github.com/naoto24kawa/mcpconfig/internal/utils"
)

// TemplateManager handles server template CRUD operations
type TemplateManager struct {
	serversDir string
}

// NewTemplateManager creates a new TemplateManager instance
func NewTemplateManager(serversDir string) *TemplateManager {
	return &TemplateManager{
		serversDir: serversDir,
	}
}

// SaveFromFile saves a server template from an MCP config file
func (tm *TemplateManager) SaveFromFile(templateName, serverName, mcpConfigPath string, force bool) error {
	if !force && tm.exists(templateName) {
		if !interaction.ConfirmOverwrite("サーバーテンプレート", templateName) {
			return fmt.Errorf("上書きをキャンセルしました")
		}
	}

	mcpConfig := &MCPConfig{}
	if err := utils.LoadJSON(mcpConfigPath, mcpConfig); err != nil {
		return fmt.Errorf("MCP設定ファイルの読み込みに失敗しました: %w", err)
	}

	server, exists := mcpConfig.McpServers[serverName]
	if !exists {
		return fmt.Errorf("MCPサーバー '%s' がMCP設定ファイルに見つかりません", serverName)
	}

	template := &ServerTemplate{
		Name:         templateName,
		Description:  nil,
		CreatedAt:    time.Now(),
		ServerConfig: ServerConfig(server),
	}

	if err := tm.save(template); err != nil {
		return err
	}

	fmt.Printf("サーバーテンプレート '%s' を保存しました\n", templateName)
	fmt.Printf("コマンド: %s\n", server.Command)
	if len(server.Args) > 0 {
		fmt.Printf("引数: %v\n", server.Args)
	}

	return nil
}

// SaveFromConfig saves a server template from MCPServer config
func (tm *TemplateManager) SaveFromConfig(name string, server MCPServer) error {
	template := &ServerTemplate{
		Name:         name,
		Description:  nil,
		CreatedAt:    time.Now(),
		ServerConfig: ServerConfig(server),
	}

	return tm.save(template)
}

// Load loads a server template by name
func (tm *TemplateManager) Load(name string) (*ServerTemplate, error) {
	templatePath := filepath.Join(tm.serversDir, name+config.FileExtension)

	template := &ServerTemplate{}
	if err := utils.LoadJSON(templatePath, template); err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("サーバーテンプレート '%s' が見つかりません", name)
		}
		return nil, fmt.Errorf("サーバーテンプレートの読み込みに失敗しました: %w", err)
	}

	return template, nil
}

// Exists checks if a server template exists
func (tm *TemplateManager) Exists(name string) (bool, error) {
	return tm.exists(name), nil
}

// Delete deletes a server template
func (tm *TemplateManager) Delete(name string, force bool) error {
	templatePath := filepath.Join(tm.serversDir, name+config.FileExtension)

	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		return fmt.Errorf("サーバーテンプレート '%s' が見つかりません", name)
	}

	if !force {
		if !interaction.Confirm(fmt.Sprintf("サーバーテンプレート '%s' を削除しますか？", name)) {
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

// Rename renames a server template
func (tm *TemplateManager) Rename(oldName, newName string, force bool) error {
	if err := tm.validateRename(oldName, newName, force); err != nil {
		return err
	}

	template, err := tm.Load(oldName)
	if err != nil {
		return err
	}

	if err := tm.performRename(template, oldName, newName); err != nil {
		return err
	}

	fmt.Printf("サーバーテンプレート '%s' を '%s' に変更しました\n", oldName, newName)
	return nil
}

func (tm *TemplateManager) validateRename(oldName, newName string, force bool) error {
	oldPath := filepath.Join(tm.serversDir, oldName+config.FileExtension)
	newPath := filepath.Join(tm.serversDir, newName+config.FileExtension)

	if _, err := os.Stat(oldPath); os.IsNotExist(err) {
		return fmt.Errorf("サーバーテンプレート '%s' が見つかりません", oldName)
	}

	if _, err := os.Stat(newPath); err == nil && !force {
		return fmt.Errorf("サーバーテンプレート '%s' は既に存在します\n別の名前を指定するか、--force オプションで上書きしてください", newName)
	}

	return nil
}

func (tm *TemplateManager) performRename(template *ServerTemplate, oldName, newName string) error {
	template.Name = newName

	if err := tm.save(template); err != nil {
		return err
	}

	oldPath := filepath.Join(tm.serversDir, oldName+config.FileExtension)
	if err := os.Remove(oldPath); err != nil {
		return fmt.Errorf("古いサーバーテンプレートの削除に失敗しました: %w", err)
	}

	return nil
}

func (tm *TemplateManager) exists(name string) bool {
	templatePath := filepath.Join(tm.serversDir, name+config.FileExtension)
	_, err := os.Stat(templatePath)
	return err == nil
}

func (tm *TemplateManager) save(template *ServerTemplate) error {
	templatePath := filepath.Join(tm.serversDir, template.Name+config.FileExtension)
	return utils.SaveJSON(templatePath, template)
}

// Reset deletes all server templates
func (tm *TemplateManager) Reset(force bool) error {
	if _, err := os.Stat(tm.serversDir); os.IsNotExist(err) {
		fmt.Println("サーバーテンプレートディレクトリが存在しません")
		return nil
	}
	
	files, err := os.ReadDir(tm.serversDir)
	if err != nil {
		return fmt.Errorf("サーバーテンプレートディレクトリの読み込みに失敗しました: %w", err)
	}
	
	templateFiles := []string{}
	for _, file := range files {
		if strings.HasSuffix(file.Name(), config.FileExtension) {
			templateFiles = append(templateFiles, file.Name())
		}
	}
	
	if len(templateFiles) == 0 {
		fmt.Println("削除するサーバーテンプレートが存在しません")
		return nil
	}
	
	if !force {
		fmt.Printf("以下の%d個のサーバーテンプレートを削除します:\n", len(templateFiles))
		for _, file := range templateFiles {
			name := strings.TrimSuffix(file, config.FileExtension)
			fmt.Printf("  - %s\n", name)
		}
		fmt.Println()
		
		if !interaction.Confirm("すべてのサーバーテンプレートを削除しますか？") {
			fmt.Println("リセットをキャンセルしました")
			return nil
		}
	}
	
	deletedCount := 0
	for _, file := range templateFiles {
		templatePath := filepath.Join(tm.serversDir, file)
		if err := os.Remove(templatePath); err != nil {
			fmt.Printf("警告: %s の削除に失敗しました: %v\n", file, err)
		} else {
			deletedCount++
		}
	}
	
	fmt.Printf("サーバーテンプレートを%d個削除しました\n", deletedCount)
	return nil
}