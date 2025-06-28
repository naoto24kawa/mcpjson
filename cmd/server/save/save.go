package save

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
	var serverName, fromPath, command, argsStr, envStr, envFile string
	force := false

	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "--server", "-s":
			if i+1 >= len(args) {
				fmt.Fprintln(os.Stderr, "エラー: --server オプションに値が指定されていません")
				os.Exit(utils.ExitArgumentError)
			}
			serverName = args[i+1]
			i++
		case "--from", "-f":
			if i+1 >= len(args) {
				fmt.Fprintln(os.Stderr, "エラー: --from オプションに値が指定されていません")
				os.Exit(utils.ExitArgumentError)
			}
			fromPath = args[i+1]
			i++
		case "--command", "-c":
			if i+1 >= len(args) {
				fmt.Fprintln(os.Stderr, "エラー: --command オプションに値が指定されていません")
				os.Exit(utils.ExitArgumentError)
			}
			command = args[i+1]
			i++
		case "--args", "-a":
			if i+1 >= len(args) {
				fmt.Fprintln(os.Stderr, "エラー: --args オプションに値が指定されていません")
				os.Exit(utils.ExitArgumentError)
			}
			argsStr = args[i+1]
			i++
		case "--env", "-e":
			if i+1 >= len(args) {
				fmt.Fprintln(os.Stderr, "エラー: --env オプションに値が指定されていません")
				os.Exit(utils.ExitArgumentError)
			}
			envStr = args[i+1]
			i++
		case "--env-file":
			if i+1 >= len(args) {
				fmt.Fprintln(os.Stderr, "エラー: --env-file オプションに値が指定されていません")
				os.Exit(utils.ExitArgumentError)
			}
			envFile = args[i+1]
			i++
		case "--force", "-F":
			force = true
		}
	}

	if err := utils.ValidateName(templateName, "サーバーテンプレート"); err != nil {
		fmt.Fprintln(os.Stderr, "エラー:", err)
		os.Exit(utils.ExitArgumentError)
	}

	// --from が未指定の場合、デフォルトで ./.mcp.json を使用
	if fromPath == "" && serverName != "" {
		fromPath = "./.mcp.json"
	}

	serverManager := server.NewManager(cfg.ServersDir)

	if fromPath != "" && serverName != "" {
		if err := serverManager.SaveFromFile(templateName, serverName, fromPath, force); err != nil {
			fmt.Fprintln(os.Stderr, "エラー:", err)
			os.Exit(utils.ExitGeneralError)
		}
	} else if command != "" || argsStr != "" || envStr != "" || envFile != "" {
		env := make(map[string]string)
		
		if envFile != "" {
			fileEnv, err := utils.LoadEnvFile(envFile)
			if err != nil {
				fmt.Fprintln(os.Stderr, "エラー:", err)
				os.Exit(utils.ExitFileError)
			}
			for k, v := range fileEnv {
				env[k] = v
			}
		}
		
		if envStr != "" {
			parsedEnv, err := utils.ParseEnvVars(envStr)
			if err != nil {
				fmt.Fprintln(os.Stderr, "エラー:", err)
				os.Exit(utils.ExitArgumentError)
			}
			for k, v := range parsedEnv {
				env[k] = v
			}
		}
		
		parsedArgs := utils.ParseArgs(argsStr)
		
		if err := serverManager.SaveManual(templateName, command, parsedArgs, env, force); err != nil {
			fmt.Fprintln(os.Stderr, "エラー:", err)
			os.Exit(utils.ExitGeneralError)
		}
	} else {
		fmt.Fprintln(os.Stderr, "エラー: 設定ファイルからの保存には --server と --from が必要です")
		fmt.Fprintln(os.Stderr, "手動作成には --command が必要です")
		os.Exit(utils.ExitArgumentError)
	}
}