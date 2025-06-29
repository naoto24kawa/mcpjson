package delete

import (
	"github.com/naoto24kawa/mcpconfig/cmd/profile"
	"github.com/naoto24kawa/mcpconfig/internal/config"
	"github.com/naoto24kawa/mcpconfig/internal/utils"
)

func Execute(args []string) {
	profileName, argsOffset := utils.ParseProfileName(args, config.DefaultProfileName)
	force := false

	for i := argsOffset; i < len(args); i++ {
		switch args[i] {
		case "--force", "-f":
			force = true
		}
	}

	utils.HandleArgumentError(utils.ValidateName(profileName, "プロファイル"))

	cfg, err := config.New()
	utils.HandleEnvironmentError(err)

	utils.HandleGeneralError(profile.Delete(cfg, profileName, force))
}
