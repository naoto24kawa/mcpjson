package rename

import (
	"fmt"
	"os"

	"github.com/naoto24kawa/mcpjson/cmd/profile"
	"github.com/naoto24kawa/mcpjson/internal/config"
	"github.com/naoto24kawa/mcpjson/internal/utils"
)

func Execute(args []string) {
	oldName, newName, argsOffset, err := utils.ParseRenameArgs(args, config.DefaultProfileName)
	if err != nil {
		fmt.Fprintln(os.Stderr, "エラー:", err)
		os.Exit(utils.ExitArgumentError)
	}
	force := false

	for i := argsOffset; i < len(args); i++ {
		switch args[i] {
		case "--force", "-f":
			force = true
		}
	}

	utils.HandleArgumentError(utils.ValidateName(oldName, "プロファイル"))

	utils.HandleArgumentError(utils.ValidateName(newName, "プロファイル"))

	cfg, err := config.New()
	utils.HandleEnvironmentError(err)

	utils.HandleGeneralError(profile.Rename(cfg, oldName, newName, force))
}
