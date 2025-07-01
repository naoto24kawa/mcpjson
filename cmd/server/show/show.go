package show

import (
	"fmt"
	"os"
	"strings"

	"github.com/naoto24kawa/mcpjson/internal/config"
	"github.com/naoto24kawa/mcpjson/internal/server"
	"github.com/naoto24kawa/mcpjson/internal/utils"
)

func Execute(cfg *config.Config, args []string) {
	var fromPath, serverName string

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--from", "-f":
			if i+1 >= len(args) {
				fmt.Fprintln(os.Stderr, "エラー: --from オプションに値が指定されていません")
				os.Exit(utils.ExitArgumentError)
			}
			fromPath = args[i+1]
			i++
		case "--server", "-s":
			if i+1 >= len(args) {
				fmt.Fprintln(os.Stderr, "エラー: --server オプションに値が指定されていません")
				os.Exit(utils.ExitArgumentError)
			}
			serverName = args[i+1]
			i++
		}
	}

	// --from が未指定の場合、デフォルトで ./.mcp.json を使用
	if fromPath == "" {
		fromPath = "./.mcp.json"
	}

	if !utils.FileExists(fromPath) {
		fmt.Fprintln(os.Stderr, "エラー: MCP設定ファイルが見つかりません:", fromPath)
		os.Exit(utils.ExitFileError)
	}

	serverManager := server.NewManager(cfg.ServersDir)
	if err := serverManager.Show(fromPath, serverName); err != nil {
		fmt.Fprintln(os.Stderr, "エラー:", err)
		if strings.Contains(err.Error(), "MCPサーバー") && strings.Contains(err.Error(), "見つかりません") {
			os.Exit(utils.ExitServerError)
		}
		os.Exit(utils.ExitGeneralError)
	}
}
