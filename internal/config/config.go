package config

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	ConfigDirName       = ".mcpconfig"
	ProfilesDir         = "profiles"
	ServersDir          = "servers"
	DefaultHomeEnv      = "HOME"
	DefaultMCPConfig    = ".mcp.json"
	DefaultDirPerm      = 0755
	FileExtension       = ".json"
	DefaultProfileName  = "default"
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

// GetDefaultMCPConfigPath returns the default MCP configuration file path
func GetDefaultMCPConfigPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return DefaultMCPConfig // fallback to current directory
	}
	return filepath.Join(homeDir, DefaultMCPConfig)
}

// FindMCPConfigPath tries to find MCP config file in common locations
func FindMCPConfigPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	
	// Try common locations
	locations := []string{
		filepath.Join(homeDir, DefaultMCPConfig),
		filepath.Join(homeDir, ".config", "claude", "mcp.json"),
		filepath.Join(homeDir, ".config", "mcp.json"),
		filepath.Join(homeDir, ".claude", "mcp.json"),
		".claude/mcp.json", // current directory .claude folder
		"mcp.json", // current directory
	}
	
	for _, path := range locations {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}
	
	return "" // not found
}

// GetDefaultMCPPath returns the default MCP configuration file path (prefers local)
func GetDefaultMCPPath() string {
	localPath := "./.mcp.json"
	if _, err := os.Stat(localPath); err == nil {
		return localPath
	}
	return localPath // return local path even if it doesn't exist for consistency
}