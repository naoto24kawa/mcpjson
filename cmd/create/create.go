package create

import (
	"fmt"
	"os"

	"github.com/naoto24kawa/mcpconfig/internal/config"
	"github.com/naoto24kawa/mcpconfig/cmd/profile"
	"github.com/naoto24kawa/mcpconfig/internal/utils"
)

func Execute(args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "エラー: プロファイル名が指定されていません")
		fmt.Fprintln(os.Stderr, "使用方法: mcpconfig create <プロファイル名>")
		os.Exit(utils.ExitArgumentError)
	}

	profileName := args[0]
	var templateName string

	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "--template", "-t":
			if i+1 >= len(args) {
				fmt.Fprintln(os.Stderr, "エラー: --template オプションに値が指定されていません")
				os.Exit(utils.ExitArgumentError)
			}
			templateName = args[i+1]
			i++
		}
	}

	if err := utils.ValidateName(profileName, "プロファイル"); err != nil {
		fmt.Fprintln(os.Stderr, "エラー:", err)
		os.Exit(utils.ExitArgumentError)
	}

	cfg, err := config.New()
	if err != nil {
		fmt.Fprintln(os.Stderr, "エラー:", err)
		os.Exit(utils.ExitEnvironment)
	}

	if err := profile.Create(cfg, profileName, templateName); err != nil {
		fmt.Fprintln(os.Stderr, "エラー:", err)
		os.Exit(utils.ExitGeneralError)
	}
}