package profile

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/naoto24kawa/mcpconfig/internal/config"
	"github.com/naoto24kawa/mcpconfig/internal/interaction"
	"github.com/naoto24kawa/mcpconfig/internal/server"
	"github.com/naoto24kawa/mcpconfig/internal/utils"
)

const (
	ListColumnWidth     = 20
	TimestampFormat     = "2006-01-02 15:04:05"
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

type ServerRef struct {
	Name      string                 `json:"name"`
	Template  string                 `json:"template"`
	Overrides ServerOverrides        `json:"overrides,omitempty"`
}

type ServerOverrides struct {
	Env map[string]string `json:"env,omitempty"`
}


type Manager struct {
	profilesDir string
}

func NewManager(profilesDir string) *Manager {
	return &Manager{
		profilesDir: profilesDir,
	}
}

func (m *Manager) Create(name, description string) error {
	profilePath := filepath.Join(m.profilesDir, name+".json")
	
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
	
	mcpConfig, err := m.loadMCPConfig(mcpConfigPath)
	if err != nil {
		return err
	}
	
	profile := m.createProfileFromMCP(name, mcpConfigPath)
	
	if err := m.processServers(profile, mcpConfig, serverManager); err != nil {
		return err
	}
	
	if err := m.saveProfile(profile); err != nil {
		return err
	}
	
	fmt.Printf("プロファイル '%s' を保存しました (%d個のサーバー)\n", name, len(profile.Servers))
	return nil
}

func (m *Manager) validateProfileCreation(name string, force bool) error {
	profilePath := filepath.Join(m.profilesDir, name+config.FileExtension)
	
	if _, err := os.Stat(profilePath); err == nil && !force {
		return fmt.Errorf("プロファイル '%s' は既に存在します。--force オプションで上書きできます", name)
	}
	
	return nil
}

func (m *Manager) loadMCPConfig(mcpConfigPath string) (*server.MCPConfig, error) {
	mcpConfig := &server.MCPConfig{}
	if err := utils.LoadJSON(mcpConfigPath, mcpConfig); err != nil {
		return nil, fmt.Errorf("MCP設定ファイルの読み込みに失敗しました: %w", err)
	}
	return mcpConfig, nil
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
		templateName := serverName
		
		exists, err := serverManager.Exists(templateName)
		if err != nil {
			return err
		}
		
		if !exists {
			serverConfig := server.MCPServer{
				Command: mcpServer.Command,
				Args:    mcpServer.Args,
				Env:     mcpServer.Env,
			}
			if err := serverManager.SaveFromConfig(templateName, serverConfig); err != nil {
				return err
			}
			fmt.Printf("サーバーテンプレート '%s' を作成しました\n", templateName)
		} else {
			fmt.Printf("サーバーテンプレート '%s' は既に存在するため、既存のものを使用します\n", templateName)
		}
		
		profile.Servers = append(profile.Servers, ServerRef{
			Name:     serverName,
			Template: templateName,
		})
	}
	return nil
}

func (m *Manager) Apply(name string, targetPath string, serverManager *server.Manager) error {
	profile, err := m.Load(name)
	if err != nil {
		return err
	}
	
	mcpConfig, err := m.buildMCPConfig(profile, serverManager)
	if err != nil {
		return err
	}
	
	if err := m.saveMCPConfig(mcpConfig, targetPath); err != nil {
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
		for _, file := range files {
			if strings.HasSuffix(file.Name(), ".json") {
				name := strings.TrimSuffix(file.Name(), ".json")
				profile, err := m.Load(name)
				if err != nil {
					fmt.Printf("エラー: %s の読み込みに失敗しました: %v\n", name, err)
					continue
				}
				
				fmt.Printf("\nプロファイル: %s\n", profile.Name)
				fmt.Printf("  説明: %s\n", profile.Description)
				fmt.Printf("  作成日時: %s\n", profile.CreatedAt.Format("2006-01-02 15:04:05"))
				fmt.Printf("  更新日時: %s\n", profile.UpdatedAt.Format("2006-01-02 15:04:05"))
				fmt.Printf("  サーバー数: %d\n", len(profile.Servers))
				
				if len(profile.Servers) > 0 {
					fmt.Println("  サーバー:")
					for _, server := range profile.Servers {
						fmt.Printf("    - %s (テンプレート: %s)\n", server.Name, server.Template)
					}
				}
			}
		}
	} else {
		fmt.Printf("%-*s %-*s %s\n", ListColumnWidth, "プロファイル名", ListColumnWidth, "作成日時", "サーバー数")
		fmt.Println(strings.Repeat(TableSeparatorChar, TableSeparatorWidth))
		
		for _, file := range files {
			if strings.HasSuffix(file.Name(), ".json") {
				name := strings.TrimSuffix(file.Name(), ".json")
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
	}
	
	return nil
}

func (m *Manager) Delete(name string, force bool) error {
	profilePath := filepath.Join(m.profilesDir, name+".json")
	
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
	oldPath := filepath.Join(m.profilesDir, oldName+".json")
	newPath := filepath.Join(m.profilesDir, newName+".json")
	
	if _, err := os.Stat(oldPath); os.IsNotExist(err) {
		return fmt.Errorf("プロファイル '%s' が見つかりません", oldName)
	}
	
	if _, err := os.Stat(newPath); err == nil && !force {
		return fmt.Errorf("プロファイル '%s' は既に存在します\n別の名前を指定するか、--force オプションで上書きしてください", newName)
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

func (m *Manager) buildMCPConfig(profile *Profile, serverManager *server.Manager) (*server.MCPConfig, error) {
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

func (m *Manager) createMCPServer(template *server.ServerTemplate, serverRef *ServerRef) server.MCPServer {
	mcpServer := server.MCPServer{
		Command: template.ServerConfig.Command,
		Args:    template.ServerConfig.Args,
		Env:     make(map[string]string),
	}
	
	for k, v := range template.ServerConfig.Env {
		mcpServer.Env[k] = v
	}
	
	for k, v := range serverRef.Overrides.Env {
		mcpServer.Env[k] = v
	}
	
	return mcpServer
}

func (m *Manager) saveMCPConfig(mcpConfig *server.MCPConfig, targetPath string) error {
	targetDir := filepath.Dir(targetPath)
	if err := os.MkdirAll(targetDir, config.DefaultDirPerm); err != nil {
		return fmt.Errorf("ディレクトリの作成に失敗しました: %w", err)
	}
	
	if err := utils.SaveJSON(targetPath, mcpConfig); err != nil {
		return fmt.Errorf("MCP設定ファイルの保存に失敗しました: %w", err)
	}
	
	return nil
}

func (m *Manager) Load(name string) (*Profile, error) {
	profilePath := filepath.Join(m.profilesDir, name+".json")
	
	profile := &Profile{}
	if err := utils.LoadJSON(profilePath, profile); err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("プロファイル '%s' が見つかりません", name)
		}
		return nil, fmt.Errorf("プロファイルの読み込みに失敗しました: %w", err)
	}
	
	return profile, nil
}

func (m *Manager) saveProfile(profile *Profile) error {
	profilePath := filepath.Join(m.profilesDir, profile.Name+".json")
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
			return fmt.Errorf("サーバー名 '%s' は既にプロファイル '%s' に存在します\n別の名前を指定してください: --as <新しい名前>", serverName, profileName)
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
	
	found := false
	newServers := []ServerRef{}
	
	for _, server := range profile.Servers {
		if server.Name != serverName {
			newServers = append(newServers, server)
		} else {
			found = true
		}
	}
	
	if !found {
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



