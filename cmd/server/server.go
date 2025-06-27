package server

import (
	"fmt"
	"os"
	"strings"

	"github.com/naoto24kawa/mcpconfig/internal/config"
	"github.com/naoto24kawa/mcpconfig/internal/profile"
	"github.com/naoto24kawa/mcpconfig/internal/server"
	"github.com/naoto24kawa/mcpconfig/internal/utils"
)

func Execute(cfg *config.Config, args []string) {
	if len(args) == 0 {
		PrintUsage()
		os.Exit(0)
	}

	subCmd := args[0]
	subArgs := args[1:]

	switch subCmd {
	case "save":
		handleSave(cfg, subArgs)
	case "list":
		handleList(cfg, subArgs)
	case "delete":
		handleDelete(cfg, subArgs)
	case "rename":
		handleRename(cfg, subArgs)
	case "add":
		handleAdd(cfg, subArgs)
	case "remove":
		handleRemove(cfg, subArgs)
	case "show":
		handleShow(cfg, subArgs)
	default:
		fmt.Fprintf(os.Stderr, "エラー: 不明なサブコマンド 'server %s'\n", subCmd)
		PrintUsage()
		os.Exit(utils.ExitGeneralError)
	}
}

func PrintUsage() {
	fmt.Println(`mcpconfig server - MCPサーバー管理

使用方法:
  mcpconfig server <サブコマンド> [オプション]

サブコマンド:
  save <名前> --server <サーバー名> --from <パス>    設定ファイルからテンプレート保存
  save <名前> --command <コマンド> [オプション]      手動でテンプレート作成
  list [--detail]                                    テンプレート一覧表示
  delete <名前>                                      テンプレート削除
  rename <現在の名前> <新しい名前>                    テンプレート名変更
  add <テンプレート名> --to <プロファイル名>          プロファイルにテンプレート追加
  remove <サーバー名> --from <プロファイル名>         プロファイルからサーバー削除
  show --from <パス> [--server <名前>]               設定ファイルからサーバー情報表示`)
}

func handleSave(cfg *config.Config, args []string) {
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

func handleList(cfg *config.Config, args []string) {
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

func handleDelete(cfg *config.Config, args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "エラー: テンプレート名が指定されていません")
		os.Exit(utils.ExitArgumentError)
	}

	templateName := args[0]
	force := false

	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "--force", "-f":
			force = true
		}
	}

	if err := utils.ValidateName(templateName, "サーバーテンプレート"); err != nil {
		fmt.Fprintln(os.Stderr, "エラー:", err)
		os.Exit(utils.ExitArgumentError)
	}

	serverManager := server.NewManager(cfg.ServersDir)
	if err := serverManager.Delete(templateName, force); err != nil {
		fmt.Fprintln(os.Stderr, "エラー:", err)
		os.Exit(utils.ExitGeneralError)
	}
}

func handleRename(cfg *config.Config, args []string) {
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

func handleAdd(cfg *config.Config, args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "エラー: テンプレート名が指定されていません")
		os.Exit(utils.ExitArgumentError)
	}

	templateName := args[0]
	var profileName, serverName, envStr string

	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "--to", "-t":
			if i+1 >= len(args) {
				fmt.Fprintln(os.Stderr, "エラー: --to オプションに値が指定されていません")
				os.Exit(utils.ExitArgumentError)
			}
			profileName = args[i+1]
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

	if profileName == "" {
		fmt.Fprintln(os.Stderr, "エラー: 追加先のプロファイル名が指定されていません")
		fmt.Fprintln(os.Stderr, "使用方法: mcpconfig server add <テンプレート名> --to <プロファイル名>")
		os.Exit(utils.ExitArgumentError)
	}

	if err := utils.ValidateName(templateName, "サーバーテンプレート"); err != nil {
		fmt.Fprintln(os.Stderr, "エラー:", err)
		os.Exit(utils.ExitArgumentError)
	}

	if err := utils.ValidateName(profileName, "プロファイル"); err != nil {
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

	profileManager := profile.NewManager(cfg.ProfilesDir)
	if err := profileManager.AddServer(profileName, templateName, serverName, envOverrides); err != nil {
		fmt.Fprintln(os.Stderr, "エラー:", err)
		os.Exit(utils.ExitGeneralError)
	}
}

func handleRemove(cfg *config.Config, args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "エラー: サーバー名が指定されていません")
		os.Exit(utils.ExitArgumentError)
	}

	serverName := args[0]
	var profileName string

	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "--from", "-f":
			if i+1 >= len(args) {
				fmt.Fprintln(os.Stderr, "エラー: --from オプションに値が指定されていません")
				os.Exit(utils.ExitArgumentError)
			}
			profileName = args[i+1]
			i++
		}
	}

	if profileName == "" {
		fmt.Fprintln(os.Stderr, "エラー: 削除元のプロファイル名が指定されていません")
		fmt.Fprintln(os.Stderr, "使用方法: mcpconfig server remove <サーバー名> --from <プロファイル名>")
		os.Exit(utils.ExitArgumentError)
	}

	if err := utils.ValidateName(profileName, "プロファイル"); err != nil {
		fmt.Fprintln(os.Stderr, "エラー:", err)
		os.Exit(utils.ExitArgumentError)
	}

	profileManager := profile.NewManager(cfg.ProfilesDir)
	if err := profileManager.RemoveServer(profileName, serverName); err != nil {
		fmt.Fprintln(os.Stderr, "エラー:", err)
		os.Exit(utils.ExitGeneralError)
	}
}

func handleShow(cfg *config.Config, args []string) {
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

	if fromPath == "" {
		fmt.Fprintln(os.Stderr, "エラー: 対象のMCP設定ファイルパスが指定されていません")
		fmt.Fprintln(os.Stderr, "使用方法: mcpconfig server show --from <パス> [--server <名前>]")
		os.Exit(utils.ExitArgumentError)
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