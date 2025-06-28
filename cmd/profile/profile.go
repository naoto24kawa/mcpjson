package profile

import (
	"github.com/naoto24kawa/mcpconfig/internal/config"
	"github.com/naoto24kawa/mcpconfig/internal/profile"
	"github.com/naoto24kawa/mcpconfig/internal/server"
)

func Apply(cfg *config.Config, profileName, targetPath string) error {
	profileManager := profile.NewManager(cfg.ProfilesDir)
	serverManager := server.NewManager(cfg.ServersDir)
	
	return profileManager.Apply(profileName, targetPath, serverManager)
}

func Save(cfg *config.Config, profileName, fromPath string, force bool) error {
	profileManager := profile.NewManager(cfg.ProfilesDir)
	serverManager := server.NewManager(cfg.ServersDir)
	
	return profileManager.Save(profileName, fromPath, serverManager, force)
}

func Create(cfg *config.Config, profileName, templateName string) error {
	profileManager := profile.NewManager(cfg.ProfilesDir)
	
	description := ""
	if templateName != "" {
		description = "テンプレート " + templateName + " から作成"
	}
	
	return profileManager.Create(profileName, description)
}

func List(cfg *config.Config, detail bool) error {
	profileManager := profile.NewManager(cfg.ProfilesDir)
	return profileManager.List(detail)
}

func Delete(cfg *config.Config, profileName string, force bool) error {
	profileManager := profile.NewManager(cfg.ProfilesDir)
	return profileManager.Delete(profileName, force)
}

func Rename(cfg *config.Config, oldName, newName string, force bool) error {
	profileManager := profile.NewManager(cfg.ProfilesDir)
	return profileManager.Rename(oldName, newName, force)
}

func GetProfilePath(cfg *config.Config, profileName string) (string, error) {
	profileManager := profile.NewManager(cfg.ProfilesDir)
	return profileManager.GetProfilePath(profileName)
}