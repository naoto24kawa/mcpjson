package list

import (
	"fmt"
	"os"

	"github.com/naoto24kawa/mcpjson/internal/config"
	"github.com/naoto24kawa/mcpjson/internal/server"
	"github.com/naoto24kawa/mcpjson/internal/utils"
)

func Execute(cfg *config.Config, args []string) {
	detail := false

	for _, arg := range args {
		switch arg {
		case "--detail", "-d":
			detail = true
		}
	}

	serverManager := server.NewManager(cfg.ServersDir)
	if err := serverManager.List(detail); err != nil {
		fmt.Fprintln(os.Stderr, "エラー:", err)
		os.Exit(utils.ExitGeneralError)
	}
}
