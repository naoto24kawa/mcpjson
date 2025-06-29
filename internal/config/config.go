package config

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	ConfigDirName      = ".mcpconfig"
	ProfilesDir        = "profiles"
	ServersDir         = "servers"
	DefaultHomeEnv     = "HOME"
	DefaultMCPConfig   = ".mcp.json"
	DefaultDirPerm     = 0755
	FileExtension      = ".jsonc"
	DefaultProfileName = "default"
)

type Config struct {
	BaseDir     string
	ProfilesDir string
	ServersDir  string
}

func New() (*Config, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("ホームディレクトリの取得に失敗しました: %w", err)
	}

	baseDir := filepath.Join(homeDir, ConfigDirName)

	cfg := &Config{
		BaseDir:     baseDir,
		ProfilesDir: filepath.Join(baseDir, ProfilesDir),
		ServersDir:  filepath.Join(baseDir, ServersDir),
	}

	if err := cfg.ensureDirectories(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) ensureDirectories() error {
	dirs := []string{c.BaseDir, c.ProfilesDir, c.ServersDir}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, DefaultDirPerm); err != nil {
			return fmt.Errorf("ディレクトリの作成に失敗しました %s: %w", dir, err)
		}
	}

	return nil
}

func (c *Config) GetProfilePath(name string) string {
	return filepath.Join(c.ProfilesDir, name+FileExtension)
}

func (c *Config) GetServerPath(name string) string {
	return filepath.Join(c.ServersDir, name+FileExtension)
}

// MCPPathResolver handles MCP configuration file path resolution
type MCPPathResolver struct{}

// NewMCPPathResolver creates a new MCP path resolver
func NewMCPPathResolver() *MCPPathResolver {
	return &MCPPathResolver{}
}

// GetDefaultPath returns the default MCP configuration file path
func (r *MCPPathResolver) GetDefaultPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return DefaultMCPConfig // fallback to current directory
	}
	return filepath.Join(homeDir, DefaultMCPConfig)
}

// FindExistingPath tries to find existing MCP config file in common locations
func (r *MCPPathResolver) FindExistingPath() string {
	locations := r.getSearchLocations()

	for _, path := range locations {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	return "" // not found
}

// GetPreferredPath returns the preferred MCP configuration file path (prefers local)
func (r *MCPPathResolver) GetPreferredPath() string {
	localPath := "./.mcp.json"
	if _, err := os.Stat(localPath); err == nil {
		return localPath
	}
	return localPath // return local path even if it doesn't exist for consistency
}

// getSearchLocations returns common MCP config file locations
func (r *MCPPathResolver) getSearchLocations() []string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return []string{
			".claude/mcp.json", // current directory .claude folder
			"mcp.json",         // current directory
		}
	}

	return []string{
		filepath.Join(homeDir, DefaultMCPConfig),
		filepath.Join(homeDir, ".config", "claude", "mcp.json"),
		filepath.Join(homeDir, ".config", "mcp.json"),
		filepath.Join(homeDir, ".claude", "mcp.json"),
		".claude/mcp.json", // current directory .claude folder
		"mcp.json",         // current directory
	}
}

// Legacy functions for backward compatibility
func GetDefaultMCPConfigPath() string {
	resolver := NewMCPPathResolver()
	return resolver.GetDefaultPath()
}

func FindMCPConfigPath() string {
	resolver := NewMCPPathResolver()
	return resolver.FindExistingPath()
}

func GetDefaultMCPPath() string {
	resolver := NewMCPPathResolver()
	return resolver.GetPreferredPath()
}
