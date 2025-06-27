package add

import (
	"fmt"
	"os"

	"github.com/naoto24kawa/mcpconfig/internal/config"
	"github.com/naoto24kawa/mcpconfig/internal/profile"
	"github.com/naoto24kawa/mcpconfig/internal/utils"
)

func Execute(cfg *config.Config, args []string) {
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