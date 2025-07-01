package server

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/naoto24kawa/mcpjson/internal/config"
	"github.com/naoto24kawa/mcpjson/internal/interaction"
	"github.com/naoto24kawa/mcpjson/internal/utils"
)

// ProfileManager インターフェースはプロファイル管理機能を抽象化します
type ProfileManager interface {
	FindProfilesUsingTemplate(templateName string) ([]string, error)
	RemoveTemplateReferencesFromAllProfiles(templateName string) error
}

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
	template := &ServerTemplate{}
	if err := utils.LoadJSON(tm.getTemplatePath(name), template); err != nil {
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
func (tm *TemplateManager) Delete(name string, force bool, profileManager ProfileManager) error {
	templatePath := tm.getTemplatePath(name)

	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		return fmt.Errorf("サーバーテンプレート '%s' が見つかりません", name)
	}

	// プロファイルでの使用状況をチェック
	var usingProfiles []string
	if profileManager != nil {
		var err error
		usingProfiles, err = profileManager.FindProfilesUsingTemplate(name)
		if err != nil {
			return fmt.Errorf("プロファイルでの使用状況確認に失敗しました: %w", err)
		}
	}

	// 使用中のプロファイルがある場合の警告と削除確認
	if len(usingProfiles) > 0 {
		fmt.Printf("警告: サーバーテンプレート '%s' は以下のプロファイルで使用されています:\n", name)
		for _, profileName := range usingProfiles {
			fmt.Printf("  - %s\n", profileName)
		}
		fmt.Println()
	}

	// 削除確認
	if !force {
		if !interaction.Confirm(fmt.Sprintf("サーバーテンプレート '%s' を削除しますか？", name)) {
			fmt.Println("削除をキャンセルしました")
			return nil
		}
	}

	// プロファイルからの参照削除処理
	if len(usingProfiles) > 0 && profileManager != nil {
		if !force {
			// 通常削除の場合、プロファイルからの参照も削除するか確認
			if interaction.Confirm("プロファイルからの参照も削除しますか？") {
				if err := profileManager.RemoveTemplateReferencesFromAllProfiles(name); err != nil {
					fmt.Printf("警告: プロファイルからの参照削除に失敗しました: %v\n", err)
				}
			}
		} else {
			// 強制削除の場合、プロファイルからの参照も自動削除
			fmt.Println("強制削除: プロファイルからの参照も削除します")
			if err := profileManager.RemoveTemplateReferencesFromAllProfiles(name); err != nil {
				fmt.Printf("警告: プロファイルからの参照削除に失敗しました: %v\n", err)
			}
		}
	}

	if err := os.Remove(templatePath); err != nil {
		return fmt.Errorf("サーバーテンプレートの削除に失敗しました: %w", err)
	}

	fmt.Printf("サーバーテンプレート '%s' を削除しました\n", name)
	return nil
}

// Copy copies a server template
func (tm *TemplateManager) Copy(srcName, destName string, force bool) error {
	if err := tm.validateCopy(srcName, destName, force); err != nil {
		return err
	}

	template, err := tm.Load(srcName)
	if err != nil {
		return err
	}

	if err := tm.performCopy(template, destName); err != nil {
		return err
	}

	fmt.Printf("サーバーテンプレート '%s' を '%s' にコピーしました\n", srcName, destName)
	return nil
}

func (tm *TemplateManager) validateCopy(srcName, destName string, force bool) error {
	// 空の名前をチェック
	if srcName == "" {
		return fmt.Errorf("コピー元のサーバーテンプレート名が指定されていません")
	}
	if destName == "" {
		return fmt.Errorf("コピー先のサーバーテンプレート名が指定されていません")
	}

	// 同じ名前をチェック
	if srcName == destName {
		return fmt.Errorf("コピー元とコピー先が同じ名前です")
	}

	srcPath := tm.getTemplatePath(srcName)
	destPath := tm.getTemplatePath(destName)

	if _, err := os.Stat(srcPath); os.IsNotExist(err) {
		return fmt.Errorf("サーバーテンプレート '%s' が見つかりません", srcName)
	}

	if _, err := os.Stat(destPath); err == nil && !force {
		return fmt.Errorf("サーバーテンプレート '%s' は既に存在します\n別の名前を指定するか、--force オプションで上書きしてください", destName)
	}

	return nil
}

func (tm *TemplateManager) performCopy(template *ServerTemplate, destName string) error {
	// 新しいテンプレートを作成（名前と作成日時を更新）
	newTemplate := &ServerTemplate{
		Name:         destName,
		Description:  template.Description,
		CreatedAt:    time.Now(),
		ServerConfig: template.ServerConfig,
	}

	return tm.save(newTemplate)
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
	oldPath := tm.getTemplatePath(oldName)
	newPath := tm.getTemplatePath(newName)

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

	if err := os.Remove(tm.getTemplatePath(oldName)); err != nil {
		return fmt.Errorf("古いサーバーテンプレートの削除に失敗しました: %w", err)
	}

	return nil
}

// GetTemplatePath returns the file path for a server template
func (tm *TemplateManager) GetTemplatePath(name string) (string, error) {
	templatePath := tm.getTemplatePath(name)

	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		return "", fmt.Errorf("サーバーテンプレート '%s' が見つかりません", name)
	}

	return templatePath, nil
}

func (tm *TemplateManager) getTemplatePath(name string) string {
	return filepath.Join(tm.serversDir, name+config.FileExtension)
}

func (tm *TemplateManager) exists(name string) bool {
	_, err := os.Stat(tm.getTemplatePath(name))
	return err == nil
}

func (tm *TemplateManager) save(template *ServerTemplate) error {
	return utils.SaveJSON(tm.getTemplatePath(template.Name), template)
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
