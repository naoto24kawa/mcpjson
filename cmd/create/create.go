package create

import (
	"github.com/naoto24kawa/mcpjson/cmd/profile"
	"github.com/naoto24kawa/mcpjson/internal/config"
	"github.com/naoto24kawa/mcpjson/internal/utils"
)

func Execute(args []string) {
	profileName, argsOffset := utils.ParseProfileName(args, config.DefaultProfileName)
	var templateName string

	for i := argsOffset; i < len(args); i++ {
		switch args[i] {
		case "--template", "-t":
			var err error
			templateName, i, err = utils.ParseFlag(args, i, "--template")
			utils.HandleArgumentError(err)
		}
	}

	utils.HandleArgumentError(utils.ValidateName(profileName, "プロファイル"))

	cfg, err := config.New()
	utils.HandleEnvironmentError(err)

	utils.HandleGeneralError(profile.Create(cfg, profileName, templateName))
}
