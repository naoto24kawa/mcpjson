package apply

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
		fmt.Fprintln(os.Stderr, "使用方法: mcpconfig apply <プロファイル名> --path <パス>")
		os.Exit(utils.ExitArgumentError)
	}

	profileName := args[0]
	var targetPath string

	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "--path", "-p":
			if i+1 >= len(args) {
				fmt.Fprintln(os.Stderr, "エラー: --path オプションに値が指定されていません")
				os.Exit(utils.ExitArgumentError)
			}
			targetPath = args[i+1]
			i++
		}
	}

	if targetPath == "" {
		// デフォルトパスを使用
		targetPath = config.GetDefaultMCPConfigPath()
		fmt.Printf("適用先が指定されていないため、デフォルトパスを使用します: %s\n", targetPath)
	}

	utils.HandleArgumentError(utils.ValidateName(profileName, "プロファイル"))

	cfg, err := config.New()
	utils.HandleEnvironmentError(err)

	utils.HandleGeneralError(profile.Apply(cfg, profileName, targetPath))
}