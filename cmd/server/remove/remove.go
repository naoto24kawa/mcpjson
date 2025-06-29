package remove

import (
	"fmt"
	"os"

	"github.com/naoto24kawa/mcpconfig/internal/config"
	"github.com/naoto24kawa/mcpconfig/internal/server"
	"github.com/naoto24kawa/mcpconfig/internal/utils"
)

func Execute(cfg *config.Config, args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "エラー: サーバー名が指定されていません")
		os.Exit(utils.ExitArgumentError)
	}

	serverName := args[0]
	var mcpConfigPath string

	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "--from", "-f":
			if i+1 >= len(args) {
				fmt.Fprintln(os.Stderr, "エラー: --from オプションに値が指定されていません")
				os.Exit(utils.ExitArgumentError)
			}
			mcpConfigPath = args[i+1]
			i++
		}
	}

	// --from が未指定の場合、デフォルトで ./.mcp.json を使用
	if mcpConfigPath == "" {
		mcpConfigPath = "./.mcp.json"
	}

	serverManager := server.NewManager(cfg.ServersDir)
	if err := serverManager.RemoveFromMCPConfig(mcpConfigPath, serverName); err != nil {
		fmt.Fprintln(os.Stderr, "エラー:", err)
		os.Exit(utils.ExitGeneralError)
	}
}
