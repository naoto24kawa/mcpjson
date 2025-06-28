package profile

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/naoto24kawa/mcpconfig/internal/config"
	"github.com/naoto24kawa/mcpconfig/internal/interaction"
	"github.com/naoto24kawa/mcpconfig/internal/mcpconfig"
	"github.com/naoto24kawa/mcpconfig/internal/server"
	"github.com/naoto24kawa/mcpconfig/internal/utils"
)

const (
	ListColumnWidth     = 20
	TimestampFormat     = time.RFC3339
	TableSeparatorChar  = "-"
	TableSeparatorWidth = 60
)

type Profile struct {
	Name        string           `json:"name"`
	Description string           `json:"description"`
	CreatedAt   time.Time        `json:"createdAt"`
	UpdatedAt   time.Time        `json:"updatedAt"`
	Servers     []ServerRef      `json:"servers"`
}

type ServerRef = mcpconfig.ServerRef
type ServerOverrides = mcpconfig.ServerOverrides


type Manager struct {
	profilesDir string
}

func NewManager(profilesDir string) *Manager {
	return &Manager{
		profilesDir: profilesDir,
	}
}

func (m *Manager) Create(name, description string) error {
	profilePath := m.getProfilePath(name)
	
	if _, err := os.Stat(profilePath); err == nil {
		return fmt.Errorf("プロファイル '%s' は既に存在します", name)
	}
	
	profile := &Profile{
		Name:        name,
		Description: description,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Servers:     []ServerRef{},
	}
	
	return m.saveProfile(profile)
}

func (m *Manager) Save(name string, mcpConfigPath string, serverManager *server.Manager, force bool) error {
	if err := m.validateProfileCreation(name, force); err != nil {
		return err
	}
	
	profile, err := m.buildProfileFromMCP(name, mcpConfigPath, serverManager)
	if err != nil {
		return err
	}
	
	if err := m.saveProfile(profile); err != nil {
		return err
	}
	
	m.printSaveSuccess(name, len(profile.Servers))
	return nil
}

func (m *Manager) printSaveSuccess(name string, serverCount int) {
	fmt.Printf("プロファイル '%s' を保存しました (%d個のサーバー)\n", name, serverCount)
}

func (m *Manager) buildProfileFromMCP(name, mcpConfigPath string, serverManager *server.Manager) (*Profile, error) {
	mcpConfig, err := m.loadMCPConfig(mcpConfigPath)
	if err != nil {
		return nil, fmt.Errorf("MCP設定の読み込みに失敗: %w", err)
	}
	
	profile := m.createProfileFromMCP(name, mcpConfigPath)
	
	if err := m.processServers(profile, mcpConfig, serverManager); err != nil {
		return nil, fmt.Errorf("サーバー処理に失敗: %w", err)
	}
	
	return profile, nil
}

func (m *Manager) validateProfileCreation(name string, force bool) error {
	profilePath := filepath.Join(m.profilesDir, name+config.FileExtension)
	
	if _, err := os.Stat(profilePath); err == nil && !force {
		return fmt.Errorf("プロファイル '%s' は既に存在します。--force オプションで上書きしてください", name)
	}
	
	return nil
}

func (m *Manager) loadMCPConfig(mcpConfigPath string) (*server.MCPConfig, error) {
	mcpManager := mcpconfig.NewMCPConfigManager()
	return mcpManager.Load(mcpConfigPath)
}

func (m *Manager) createProfileFromMCP(name, mcpConfigPath string) *Profile {
	return &Profile{
		Name:        name,
		Description: fmt.Sprintf("%s から保存", mcpConfigPath),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Servers:     []ServerRef{},
	}
}

func (m *Manager) processServers(profile *Profile, mcpConfig *server.MCPConfig, serverManager *server.Manager) error {
	for serverName, mcpServer := range mcpConfig.McpServers {
		if err := m.processServer(profile, serverName, mcpServer, serverManager); err != nil {
			return err
		}
	}
	return nil
}

func (m *Manager) processServer(profile *Profile, serverName string, mcpServer server.MCPServer, serverManager *server.Manager) error {
	templateName := serverName
	
	if err := m.ensureServerTemplate(templateName, mcpServer, serverManager); err != nil {
		return err
	}
	
	profile.Servers = append(profile.Servers, ServerRef{
		Name:     serverName,
		Template: templateName,
	})
	return nil
}

func (m *Manager) ensureServerTemplate(templateName string, mcpServer server.MCPServer, serverManager *server.Manager) error {
	exists, err := serverManager.Exists(templateName)
	if err != nil {
		return err
	}
	
	if !exists {
		if err := m.createServerTemplate(templateName, mcpServer, serverManager); err != nil {
			return err
		}
		fmt.Printf("サーバーテンプレート '%s' を作成しました\n", templateName)
	} else {
		fmt.Printf("サーバーテンプレート '%s' は既に存在するため、既存のものを使用します\n", templateName)
	}
	return nil
}

func (m *Manager) createServerTemplate(templateName string, mcpServer server.MCPServer, serverManager *server.Manager) error {
	serverConfig := server.MCPServer{
		Command:       mcpServer.Command,
		Args:          mcpServer.Args,
		Env:           mcpServer.Env,
		Timeout:       mcpServer.Timeout,
		EnvFile:       mcpServer.EnvFile,
		TransportType: mcpServer.TransportType,
	}
	return serverManager.SaveFromConfig(templateName, serverConfig)
}

func (m *Manager) Apply(name string, targetPath string, serverManager *server.Manager) error {
	profile, err := m.Load(name)
	if err != nil {
		return err
	}
	
	mcpManager := mcpconfig.NewMCPConfigManager()
	mcpConfig, err := mcpManager.BuildFromProfile((*mcpconfig.ProfileData)(profile), serverManager)
	if err != nil {
		return err
	}
	
	if err := mcpManager.Save(mcpConfig, targetPath); err != nil {
		return err
	}
	
	fmt.Printf("プロファイル '%s' を適用しました\n", name)
	fmt.Printf("%d個のサーバー設定を '%s' に保存\n", len(profile.Servers), targetPath)
	return nil
}

func (m *Manager) List(detail bool) error {
	files, err := os.ReadDir(m.profilesDir)
	if err != nil {
		return fmt.Errorf("プロファイルディレクトリの読み込みに失敗しました: %w", err)
	}
	
	if len(files) == 0 {
		fmt.Println("プロファイルが存在しません")
		return nil
	}
	
	if detail {
		return m.listDetailed(files)
	}
	return m.listSummary(files)
}

func (m *Manager) listDetailed(files []os.DirEntry) error {
	for _, file := range files {
		if strings.HasSuffix(file.Name(), config.FileExtension) {
			name := strings.TrimSuffix(file.Name(), config.FileExtension)
			profile, err := m.Load(name)
			if err != nil {
				fmt.Printf("エラー: %s の読み込みに失敗しました: %v\n", name, err)
				continue
			}
			
			m.printProfileDetails(profile)
		}
	}
	return nil
}

func (m *Manager) listSummary(files []os.DirEntry) error {
	fmt.Printf("%-*s %-*s %s\n", ListColumnWidth, "プロファイル名", ListColumnWidth, "作成日時", "サーバー数")
	fmt.Println(strings.Repeat(TableSeparatorChar, TableSeparatorWidth))
	
	for _, file := range files {
		if strings.HasSuffix(file.Name(), config.FileExtension) {
			name := strings.TrimSuffix(file.Name(), config.FileExtension)
			profile, err := m.Load(name)
			if err != nil {
				continue
			}
			
			fmt.Printf("%-*s %-*s %d\n",
				ListColumnWidth, profile.Name,
				ListColumnWidth, profile.CreatedAt.Format(TimestampFormat),
				len(profile.Servers))
		}
	}
	return nil
}

func (m *Manager) printProfileDetails(profile *Profile) {
	fmt.Printf("\nプロファイル: %s\n", profile.Name)
	fmt.Printf("  説明: %s\n", profile.Description)
	fmt.Printf("  作成日時: %s\n", profile.CreatedAt.Format(TimestampFormat))
	fmt.Printf("  更新日時: %s\n", profile.UpdatedAt.Format(TimestampFormat))
	fmt.Printf("  サーバー数: %d\n", len(profile.Servers))
	
	if len(profile.Servers) > 0 {
		fmt.Println("  サーバー:")
		for _, server := range profile.Servers {
			fmt.Printf("    - %s (テンプレート: %s)\n", server.Name, server.Template)
		}
	}
}

func (m *Manager) Delete(name string, force bool) error {
	profilePath := m.getProfilePath(name)
	
	if _, err := os.Stat(profilePath); os.IsNotExist(err) {
		return fmt.Errorf("プロファイル '%s' が見つかりません", name)
	}
	
	if !force {
		if !interaction.Confirm(fmt.Sprintf("プロファイル '%s' を削除しますか？", name)) {
			fmt.Println("削除をキャンセルしました")
			return nil
		}
	}
	
	if err := os.Remove(profilePath); err != nil {
		return fmt.Errorf("プロファイルの削除に失敗しました: %w", err)
	}
	
	fmt.Printf("プロファイル '%s' を削除しました\n", name)
	return nil
}

func (m *Manager) Rename(oldName, newName string, force bool) error {
	oldPath := m.getProfilePath(oldName)
	newPath := m.getProfilePath(newName)
	
	if _, err := os.Stat(oldPath); os.IsNotExist(err) {
		return fmt.Errorf("プロファイル '%s' が見つかりません", oldName)
	}
	
	if _, err := os.Stat(newPath); err == nil && !force {
		return fmt.Errorf("プロファイル '%s' は既に存在します。別の名前を指定するか、--force オプションで上書きしてください", newName)
	}
	
	profile, err := m.Load(oldName)
	if err != nil {
		return err
	}
	
	profile.Name = newName
	profile.UpdatedAt = time.Now()
	
	if err := m.saveProfile(profile); err != nil {
		return err
	}
	
	if err := os.Remove(oldPath); err != nil {
		return fmt.Errorf("古いプロファイルの削除に失敗しました: %w", err)
	}
	
	fmt.Printf("プロファイル '%s' を '%s' に変更しました\n", oldName, newName)
	return nil
}


func (m *Manager) Load(name string) (*Profile, error) {
	profilePath := m.getProfilePath(name)
	
	profile := &Profile{}
	if err := utils.LoadJSON(profilePath, profile); err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("プロファイル '%s' が見つかりません", name)
		}
		return nil, fmt.Errorf("プロファイルの読み込みに失敗しました: %w", err)
	}
	
	return profile, nil
}

func (m *Manager) getProfilePath(name string) string {
	return filepath.Join(m.profilesDir, name+config.FileExtension)
}

func (m *Manager) saveProfile(profile *Profile) error {
	profilePath := m.getProfilePath(profile.Name)
	return utils.SaveJSON(profilePath, profile)
}

func (m *Manager) AddServer(profileName, templateName, serverName string, envOverrides map[string]string) error {
	profile, err := m.Load(profileName)
	if err != nil {
		return err
	}
	
	if serverName == "" {
		serverName = templateName
	}
	
	for _, server := range profile.Servers {
		if server.Name == serverName {
			return fmt.Errorf("サーバー名 '%s' は既にプロファイル '%s' に存在します。別の名前を指定してください（--as オプションを使用）", serverName, profileName)
		}
	}
	
	serverRef := ServerRef{
		Name:     serverName,
		Template: templateName,
	}
	
	if len(envOverrides) > 0 {
		serverRef.Overrides.Env = envOverrides
	}
	
	profile.Servers = append(profile.Servers, serverRef)
	profile.UpdatedAt = time.Now()
	
	if err := m.saveProfile(profile); err != nil {
		return err
	}
	
	fmt.Printf("サーバー '%s' をプロファイル '%s' に追加しました\n", serverName, profileName)
	return nil
}

func (m *Manager) RemoveServer(profileName, serverName string) error {
	profile, err := m.Load(profileName)
	if err != nil {
		return err
	}
	
	newServers, removedCount := m.filterServersByName(profile.Servers, serverName)
	
	if removedCount == 0 {
		return fmt.Errorf("サーバー '%s' がプロファイル '%s' に見つかりません", serverName, profileName)
	}
	
	profile.Servers = newServers
	profile.UpdatedAt = time.Now()
	
	if err := m.saveProfile(profile); err != nil {
		return err
	}
	
	fmt.Printf("サーバー '%s' をプロファイル '%s' から削除しました\n", serverName, profileName)
	return nil
}

func (m *Manager) filterServersByName(servers []ServerRef, serverName string) ([]ServerRef, int) {
	newServers := []ServerRef{}
	removedCount := 0
	
	for _, server := range servers {
		if server.Name != serverName {
			newServers = append(newServers, server)
		} else {
			removedCount++
		}
	}
	
	return newServers, removedCount
}

func (m *Manager) GetProfilePath(name string) (string, error) {
	profilePath := m.getProfilePath(name)
	
	if _, err := os.Stat(profilePath); os.IsNotExist(err) {
		return "", fmt.Errorf("プロファイル '%s' が見つかりません", name)
	}
	
	return profilePath, nil
}

func (m *Manager) Reset(force bool) error {
	if _, err := os.Stat(m.profilesDir); os.IsNotExist(err) {
		fmt.Println("プロファイルディレクトリが存在しません")
		return nil
	}
	
	files, err := os.ReadDir(m.profilesDir)
	if err != nil {
		return fmt.Errorf("プロファイルディレクトリの読み込みに失敗しました: %w", err)
	}
	
	profileFiles := []string{}
	for _, file := range files {
		if strings.HasSuffix(file.Name(), config.FileExtension) {
			profileFiles = append(profileFiles, file.Name())
		}
	}
	
	if len(profileFiles) == 0 {
		fmt.Println("削除するプロファイルが存在しません")
		return nil
	}
	
	if !force {
		fmt.Printf("以下の%d個のプロファイルを削除します:\n", len(profileFiles))
		for _, file := range profileFiles {
			name := strings.TrimSuffix(file, config.FileExtension)
			fmt.Printf("  - %s\n", name)
		}
		fmt.Println()
		
		if !interaction.Confirm("すべてのプロファイルを削除しますか？") {
			fmt.Println("リセットをキャンセルしました")
			return nil
		}
	}
	
	deletedCount := 0
	for _, file := range profileFiles {
		profilePath := filepath.Join(m.profilesDir, file)
		if err := os.Remove(profilePath); err != nil {
			fmt.Printf("警告: %s の削除に失敗しました: %v\n", file, err)
		} else {
			deletedCount++
		}
	}
	
	fmt.Printf("プロファイルを%d個削除しました\n", deletedCount)
	return nil
}

// FindProfilesUsingTemplate は指定されたサーバーテンプレートを使用しているプロファイルを検索します
func (m *Manager) FindProfilesUsingTemplate(templateName string) ([]string, error) {
	files, err := os.ReadDir(m.profilesDir)
	if err != nil {
		return nil, fmt.Errorf("プロファイルディレクトリの読み込みに失敗しました: %w", err)
	}
	
	var usingProfiles []string
	
	for _, file := range files {
		if !strings.HasSuffix(file.Name(), config.FileExtension) {
			continue
		}
		
		profileName := strings.TrimSuffix(file.Name(), config.FileExtension)
		profile, err := m.Load(profileName)
		if err != nil {
			continue // 読み込みに失敗したプロファイルはスキップ
		}
		
		// プロファイル内のサーバーで指定されたテンプレートを使用しているかチェック
		for _, server := range profile.Servers {
			if server.Template == templateName {
				usingProfiles = append(usingProfiles, profileName)
				break // 同じプロファイル内で複数使用されていても一度だけ追加
			}
		}
	}
	
	return usingProfiles, nil
}

// RemoveTemplateReferencesFromProfile は指定されたプロファイルから特定のサーバーテンプレート参照を削除します
func (m *Manager) RemoveTemplateReferencesFromProfile(profileName, templateName string) error {
	profile, err := m.Load(profileName)
	if err != nil {
		return err
	}
	
	newServers, removedCount := m.filterServersByTemplate(profile.Servers, templateName)
	
	if removedCount == 0 {
		return nil
	}
	
	profile.Servers = newServers
	profile.UpdatedAt = time.Now()
	
	if err := m.saveProfile(profile); err != nil {
		return err
	}
	
	fmt.Printf("プロファイル '%s' からサーバーテンプレート '%s' の参照を%d個削除しました\n", profileName, templateName, removedCount)
	return nil
}

func (m *Manager) filterServersByTemplate(servers []ServerRef, templateName string) ([]ServerRef, int) {
	newServers := []ServerRef{}
	removedCount := 0
	
	for _, server := range servers {
		if server.Template != templateName {
			newServers = append(newServers, server)
		} else {
			removedCount++
		}
	}
	
	return newServers, removedCount
}

// RemoveTemplateReferencesFromAllProfiles はすべてのプロファイルから特定のサーバーテンプレート参照を削除します
func (m *Manager) RemoveTemplateReferencesFromAllProfiles(templateName string) error {
	usingProfiles, err := m.FindProfilesUsingTemplate(templateName)
	if err != nil {
		return err
	}
	
	totalRemoved := 0
	for _, profileName := range usingProfiles {
		if err := m.RemoveTemplateReferencesFromProfile(profileName, templateName); err != nil {
			fmt.Printf("警告: プロファイル '%s' からの参照削除に失敗しました: %v\n", profileName, err)
		} else {
			totalRemoved++
		}
	}
	
	if totalRemoved > 0 {
		fmt.Printf("合計%d個のプロファイルからサーバーテンプレート '%s' の参照を削除しました\n", totalRemoved, templateName)
	}
	
	return nil
}



