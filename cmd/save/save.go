package save

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
		fmt.Fprintln(os.Stderr, "使用方法: mcpconfig save <プロファイル名> --from <パス>")
		os.Exit(utils.ExitArgumentError)
	}

	profileName := args[0]
	var fromPath string
	force := false

	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "--from", "-f":
			if i+1 >= len(args) {
				fmt.Fprintln(os.Stderr, "エラー: --from オプションに値が指定されていません")
				os.Exit(utils.ExitArgumentError)
			}
			fromPath = args[i+1]
			i++
		case "--force", "-F":
			force = true
		}
	}

	if fromPath == "" {
		// デフォルトパスを検索
		fromPath = config.FindMCPConfigPath()
		if fromPath == "" {
			fmt.Fprintln(os.Stderr, "エラー: MCP設定ファイルが見つかりません")
			fmt.Fprintln(os.Stderr, "使用方法: mcpconfig save <プロファイル名> --from <パス>")
			os.Exit(utils.ExitArgumentError)
		}
		fmt.Printf("MCP設定ファイルを自動検出しました: %s\n", fromPath)
	}

	utils.HandleArgumentError(utils.ValidateName(profileName, "プロファイル"))

	cfg, err := config.New()
	utils.HandleEnvironmentError(err)

	utils.HandleGeneralError(profile.Save(cfg, profileName, fromPath, force))
}