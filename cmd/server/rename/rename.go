package rename

import (
	"fmt"
	"os"

	"github.com/naoto24kawa/mcpconfig/internal/config"
	"github.com/naoto24kawa/mcpconfig/internal/server"
	"github.com/naoto24kawa/mcpconfig/internal/utils"
)

func Execute(cfg *config.Config, args []string) {
	if len(args) < 2 {
		fmt.Fprintln(os.Stderr, "エラー: テンプレート名が指定されていません")
		os.Exit(utils.ExitArgumentError)
	}

	oldName := args[0]
	newName := args[1]
	force := false

	for i := 2; i < len(args); i++ {
		switch args[i] {
		case "--force", "-f":
			force = true
		}
	}

	if err := utils.ValidateName(oldName, "サーバーテンプレート"); err != nil {
		fmt.Fprintln(os.Stderr, "エラー:", err)
		os.Exit(utils.ExitArgumentError)
	}

	if err := utils.ValidateName(newName, "サーバーテンプレート"); err != nil {
		fmt.Fprintln(os.Stderr, "エラー:", err)
		os.Exit(utils.ExitArgumentError)
	}

	serverManager := server.NewManager(cfg.ServersDir)
	if err := serverManager.Rename(oldName, newName, force); err != nil {
		fmt.Fprintln(os.Stderr, "エラー:", err)
		os.Exit(utils.ExitGeneralError)
	}
}
