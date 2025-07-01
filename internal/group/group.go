package group

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/naoto24kawa/mcpjson/internal/config"
	"github.com/naoto24kawa/mcpjson/internal/interaction"
	"github.com/naoto24kawa/mcpjson/internal/server"
	"github.com/naoto24kawa/mcpjson/internal/utils"
)

// Group represents a server group
type Group struct {
	Name        string    `json:"name"`
	Description *string   `json:"description"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
	Servers     []string  `json:"servers"`
}

// Manager handles server group operations
type Manager struct {
	groupsDir string
}

// NewManager creates a new Group Manager instance
func NewManager(groupsDir string) *Manager {
	return &Manager{
		groupsDir: groupsDir,
	}
}

// Create creates a new group
func (gm *Manager) Create(name, description string, force bool) error {
	if !force && gm.exists(name) {
		if !interaction.ConfirmOverwrite("グループ", name) {
			return fmt.Errorf("上書きをキャンセルしました")
		}
	}

	group := &Group{
		Name:        name,
		Description: nil,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Servers:     []string{},
	}

	if description != "" {
		group.Description = &description
	}

	if err := gm.save(group); err != nil {
		return err
	}

	fmt.Printf("グループ '%s' を作成しました\n", name)
	return nil
}

// List displays all groups
func (gm *Manager) List(detail bool) error {
	if _, err := os.Stat(gm.groupsDir); os.IsNotExist(err) {
		fmt.Println("グループが存在しません")
		return nil
	}

	files, err := os.ReadDir(gm.groupsDir)
	if err != nil {
		return fmt.Errorf("グループディレクトリの読み込みに失敗しました: %w", err)
	}

	groupFiles := []string{}
	for _, file := range files {
		if strings.HasSuffix(file.Name(), config.FileExtension) {
			groupFiles = append(groupFiles, file.Name())
		}
	}

	if len(groupFiles) == 0 {
		fmt.Println("グループが存在しません")
		return nil
	}

	if detail {
		return gm.printDetailedList(groupFiles)
	}

	return gm.printSimpleList(groupFiles)
}

func (gm *Manager) printSimpleList(groupFiles []string) error {
	fmt.Printf("%-20s %-20s %s\n", "グループ名", "作成日時", "サーバー数")
	fmt.Printf("%-20s %-20s %s\n", strings.Repeat("-", 20), strings.Repeat("-", 20), strings.Repeat("-", 8))

	for _, file := range groupFiles {
		group := &Group{}
		if err := utils.LoadJSON(filepath.Join(gm.groupsDir, file), group); err != nil {
			fmt.Printf("%-20s %-20s %s\n", "エラー", "-", "-")
			continue
		}

		serverCount := len(group.Servers)
		fmt.Printf("%-20s %-20s %d\n",
			group.Name,
			group.CreatedAt.Format(server.TimestampFormat),
			serverCount,
		)
	}

	return nil
}

func (gm *Manager) printDetailedList(groupFiles []string) error {
	for i, file := range groupFiles {
		if i > 0 {
			fmt.Println()
		}

		group := &Group{}
		if err := utils.LoadJSON(filepath.Join(gm.groupsDir, file), group); err != nil {
			fmt.Printf("グループ: %s (読み込みエラー)\n", strings.TrimSuffix(file, config.FileExtension))
			continue
		}

		fmt.Printf("グループ名: %s\n", group.Name)
		if group.Description != nil {
			fmt.Printf("説明: %s\n", *group.Description)
		}
		fmt.Printf("作成日時: %s\n", group.CreatedAt.Format(server.TimestampFormat))
		fmt.Printf("更新日時: %s\n", group.UpdatedAt.Format(server.TimestampFormat))
		fmt.Printf("サーバー数: %d\n", len(group.Servers))
		if len(group.Servers) > 0 {
			fmt.Println("サーバー:")
			for _, serverName := range group.Servers {
				fmt.Printf("  - %s\n", serverName)
			}
		}
	}

	return nil
}

// Delete deletes a group
func (gm *Manager) Delete(name string, force bool) error {
	groupPath := gm.getGroupPath(name)

	if _, err := os.Stat(groupPath); os.IsNotExist(err) {
		return fmt.Errorf("グループ '%s' が見つかりません", name)
	}

	if !force {
		if !interaction.Confirm(fmt.Sprintf("グループ '%s' を削除しますか？", name)) {
			fmt.Println("削除をキャンセルしました")
			return nil
		}
	}

	if err := os.Remove(groupPath); err != nil {
		return fmt.Errorf("グループの削除に失敗しました: %w", err)
	}

	fmt.Printf("グループ '%s' を削除しました\n", name)
	return nil
}

// Rename renames a group
func (gm *Manager) Rename(oldName, newName string, force bool) error {
	if err := gm.validateRename(oldName, newName, force); err != nil {
		return err
	}

	group, err := gm.Load(oldName)
	if err != nil {
		return err
	}

	group.Name = newName
	group.UpdatedAt = time.Now()

	if err := gm.save(group); err != nil {
		return err
	}

	if err := os.Remove(gm.getGroupPath(oldName)); err != nil {
		return fmt.Errorf("古いグループファイルの削除に失敗しました: %w", err)
	}

	fmt.Printf("グループ '%s' を '%s' に変更しました\n", oldName, newName)
	return nil
}

func (gm *Manager) validateRename(oldName, newName string, force bool) error {
	oldPath := gm.getGroupPath(oldName)
	newPath := gm.getGroupPath(newName)

	if _, err := os.Stat(oldPath); os.IsNotExist(err) {
		return fmt.Errorf("グループ '%s' が見つかりません", oldName)
	}

	if _, err := os.Stat(newPath); err == nil && !force {
		return fmt.Errorf("グループ '%s' は既に存在します\n別の名前を指定するか、--force オプションで上書きしてください", newName)
	}

	return nil
}

// AddServer adds a server to a group
func (gm *Manager) AddServer(groupName, serverName string, serverManager *server.Manager) error {
	// サーバーが存在するかチェック
	if exists, err := serverManager.Exists(serverName); err != nil {
		return fmt.Errorf("サーバー存在確認に失敗しました: %w", err)
	} else if !exists {
		return fmt.Errorf("サーバーテンプレート '%s' が見つかりません", serverName)
	}

	group, err := gm.Load(groupName)
	if err != nil {
		return err
	}

	// 既にグループに含まれているかチェック
	for _, server := range group.Servers {
		if server == serverName {
			return fmt.Errorf("サーバー '%s' は既にグループ '%s' に含まれています", serverName, groupName)
		}
	}

	group.Servers = append(group.Servers, serverName)
	group.UpdatedAt = time.Now()

	if err := gm.save(group); err != nil {
		return err
	}

	fmt.Printf("サーバー '%s' をグループ '%s' に追加しました\n", serverName, groupName)
	return nil
}

// RemoveServer removes a server from a group
func (gm *Manager) RemoveServer(groupName, serverName string) error {
	group, err := gm.Load(groupName)
	if err != nil {
		return err
	}

	// サーバーがグループに含まれているかチェック
	serverIndex := -1
	for i, server := range group.Servers {
		if server == serverName {
			serverIndex = i
			break
		}
	}

	if serverIndex == -1 {
		return fmt.Errorf("サーバー '%s' はグループ '%s' に含まれていません", serverName, groupName)
	}

	// サーバーを削除
	group.Servers = append(group.Servers[:serverIndex], group.Servers[serverIndex+1:]...)
	group.UpdatedAt = time.Now()

	if err := gm.save(group); err != nil {
		return err
	}

	fmt.Printf("サーバー '%s' をグループ '%s' から削除しました\n", serverName, groupName)
	return nil
}

// Show displays group details
func (gm *Manager) Show(name string) error {
	group, err := gm.Load(name)
	if err != nil {
		return err
	}

	fmt.Printf("グループ名: %s\n", group.Name)
	if group.Description != nil {
		fmt.Printf("説明: %s\n", *group.Description)
	}
	fmt.Printf("作成日時: %s\n", group.CreatedAt.Format(server.TimestampFormat))
	fmt.Printf("更新日時: %s\n", group.UpdatedAt.Format(server.TimestampFormat))
	fmt.Printf("サーバー数: %d\n", len(group.Servers))

	if len(group.Servers) > 0 {
		fmt.Println("サーバー:")
		for _, serverName := range group.Servers {
			fmt.Printf("  - %s\n", serverName)
		}
	}

	return nil
}

// Load loads a group by name
func (gm *Manager) Load(name string) (*Group, error) {
	group := &Group{}
	if err := utils.LoadJSON(gm.getGroupPath(name), group); err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("グループ '%s' が見つかりません", name)
		}
		return nil, fmt.Errorf("グループの読み込みに失敗しました: %w", err)
	}

	return group, nil
}

// Apply applies all servers in a group to an MCP config file
func (gm *Manager) Apply(groupName, mcpConfigPath string, serverManager *server.Manager) error {
	group, err := gm.Load(groupName)
	if err != nil {
		return err
	}

	if len(group.Servers) == 0 {
		fmt.Printf("グループ '%s' にはサーバーが含まれていません\n", groupName)
		return nil
	}

	successCount := 0
	for _, serverName := range group.Servers {
		err := serverManager.AddToMCPConfig(mcpConfigPath, serverName, "", nil)
		if err != nil {
			fmt.Printf("警告: サーバー '%s' の追加に失敗しました: %v\n", serverName, err)
		} else {
			successCount++
		}
	}

	fmt.Printf("グループ '%s' から %d/%d のサーバーをMCP設定ファイルに追加しました: %s\n",
		groupName, successCount, len(group.Servers), mcpConfigPath)
	return nil
}

// RemoveFromMCP removes all servers in a group from an MCP config file
func (gm *Manager) RemoveFromMCP(groupName, mcpConfigPath string, serverManager *server.Manager) error {
	group, err := gm.Load(groupName)
	if err != nil {
		return err
	}

	if len(group.Servers) == 0 {
		fmt.Printf("グループ '%s' にはサーバーが含まれていません\n", groupName)
		return nil
	}

	successCount := 0
	for _, serverName := range group.Servers {
		err := serverManager.RemoveFromMCPConfig(mcpConfigPath, serverName)
		if err != nil {
			fmt.Printf("警告: サーバー '%s' の削除に失敗しました: %v\n", serverName, err)
		} else {
			successCount++
		}
	}

	fmt.Printf("グループ '%s' から %d/%d のサーバーをMCP設定ファイルから削除しました: %s\n",
		groupName, successCount, len(group.Servers), mcpConfigPath)
	return nil
}

func (gm *Manager) getGroupPath(name string) string {
	return filepath.Join(gm.groupsDir, name+config.FileExtension)
}

func (gm *Manager) exists(name string) bool {
	_, err := os.Stat(gm.getGroupPath(name))
	return err == nil
}

func (gm *Manager) save(group *Group) error {
	if err := os.MkdirAll(gm.groupsDir, 0755); err != nil {
		return fmt.Errorf("グループディレクトリの作成に失敗しました: %w", err)
	}
	return utils.SaveJSON(gm.getGroupPath(group.Name), group)
}

// Reset deletes all groups
func (gm *Manager) Reset(force bool) error {
	if _, err := os.Stat(gm.groupsDir); os.IsNotExist(err) {
		fmt.Println("グループディレクトリが存在しません")
		return nil
	}

	files, err := os.ReadDir(gm.groupsDir)
	if err != nil {
		return fmt.Errorf("グループディレクトリの読み込みに失敗しました: %w", err)
	}

	groupFiles := []string{}
	for _, file := range files {
		if strings.HasSuffix(file.Name(), config.FileExtension) {
			groupFiles = append(groupFiles, file.Name())
		}
	}

	if len(groupFiles) == 0 {
		fmt.Println("削除するグループが存在しません")
		return nil
	}

	if !force {
		fmt.Printf("以下の%d個のグループを削除します:\n", len(groupFiles))
		for _, file := range groupFiles {
			name := strings.TrimSuffix(file, config.FileExtension)
			fmt.Printf("  - %s\n", name)
		}
		fmt.Println()

		if !interaction.Confirm("すべてのグループを削除しますか？") {
			fmt.Println("リセットをキャンセルしました")
			return nil
		}
	}

	deletedCount := 0
	for _, file := range groupFiles {
		groupPath := filepath.Join(gm.groupsDir, file)
		if err := os.Remove(groupPath); err != nil {
			fmt.Printf("警告: %s の削除に失敗しました: %v\n", file, err)
		} else {
			deletedCount++
		}
	}

	fmt.Printf("グループを%d個削除しました\n", deletedCount)
	return nil
}
