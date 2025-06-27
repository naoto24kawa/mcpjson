package config

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	ConfigDirName  = ".mcpconfig"
	ProfilesDir    = "profiles"
	ServersDir     = "servers"
	DefaultHomeEnv = "HOME"
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
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("ディレクトリの作成に失敗しました %s: %w", dir, err)
		}
	}
	
	return nil
}

func (c *Config) GetProfilePath(name string) string {
	return filepath.Join(c.ProfilesDir, name+".json")
}

func (c *Config) GetServerPath(name string) string {
	return filepath.Join(c.ServersDir, name+".json")
}