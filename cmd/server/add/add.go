package add

import (
	"fmt"
	"os"

	"github.com/naoto24kawa/mcpconfig/internal/config"
	"github.com/naoto24kawa/mcpconfig/internal/server"
	"github.com/naoto24kawa/mcpconfig/internal/utils"
)

func Execute(cfg *config.Config, args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "エラー: テンプレート名が指定されていません")
		os.Exit(utils.ExitArgumentError)
	}

	templateName := args[0]
	var mcpConfigPath, serverName, envStr string

	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "--to", "-t":
			if i+1 >= len(args) {
				fmt.Fprintln(os.Stderr, "エラー: --to オプションに値が指定されていません")
				os.Exit(utils.ExitArgumentError)
			}
			mcpConfigPath = args[i+1]
			i++
		case "--as", "-a":
			if i+1 >= len(args) {
				fmt.Fprintln(os.Stderr, "エラー: --as オプションに値が指定されていません")
				os.Exit(utils.ExitArgumentError)
			}
			serverName = args[i+1]
			i++
		case "--env", "-e":
			if i+1 >= len(args) {
				fmt.Fprintln(os.Stderr, "エラー: --env オプションに値が指定されていません")
				os.Exit(utils.ExitArgumentError)
			}
			envStr = args[i+1]
			i++
		}
	}

	// --to が未指定の場合、デフォルトで ./.mcp.json を使用
	if mcpConfigPath == "" {
		mcpConfigPath = "./.mcp.json"
	}

	if err := utils.ValidateName(templateName, "サーバーテンプレート"); err != nil {
		fmt.Fprintln(os.Stderr, "エラー:", err)
		os.Exit(utils.ExitArgumentError)
	}

	envOverrides := make(map[string]string)
	if envStr != "" {
		parsedEnv, err := utils.ParseEnvVars(envStr)
		if err != nil {
			fmt.Fprintln(os.Stderr, "エラー:", err)
			os.Exit(utils.ExitArgumentError)
		}
		envOverrides = parsedEnv
	}

	serverManager := server.NewManager(cfg.ServersDir)
	if err := serverManager.AddToMCPConfig(mcpConfigPath, templateName, serverName, envOverrides); err != nil {
		fmt.Fprintln(os.Stderr, "エラー:", err)
		os.Exit(utils.ExitGeneralError)
	}
}