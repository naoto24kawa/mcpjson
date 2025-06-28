package apply

import (
	"fmt"

	"github.com/naoto24kawa/mcpconfig/internal/config"
	"github.com/naoto24kawa/mcpconfig/cmd/profile"
	"github.com/naoto24kawa/mcpconfig/internal/utils"
)

func Execute(args []string) {
	profileName, argsOffset := utils.ParseProfileName(args, config.DefaultProfileName)
	var targetPath string

	for i := argsOffset; i < len(args); i++ {
		switch args[i] {
		case "--to", "-t":
			var err error
			targetPath, i, err = utils.ParseFlag(args, i, "--to")
			utils.HandleArgumentError(err)
		}
	}

	if targetPath == "" {
		targetPath = config.GetDefaultMCPConfigPath()
		fmt.Printf("適用先が指定されていないため、デフォルトパスを使用します: %s\n", targetPath)
	}

	utils.HandleArgumentError(utils.ValidateName(profileName, "プロファイル"))

	cfg, err := config.New()
	utils.HandleEnvironmentError(err)

	utils.HandleGeneralError(profile.Apply(cfg, profileName, targetPath))
}