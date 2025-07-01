package copy

import (
	"fmt"
	"os"

	"github.com/naoto24kawa/mcpjson/internal/config"
	"github.com/naoto24kawa/mcpjson/internal/server"
	"github.com/naoto24kawa/mcpjson/internal/utils"
)

func Execute(cfg *config.Config, args []string) {
	if len(args) < 2 {
		fmt.Fprintf(os.Stderr, "エラー: コピー元とコピー先のサーバー名を指定してください\n")
		printUsage()
		os.Exit(utils.ExitGeneralError)
	}

	srcName := args[0]
	destName := args[1]
	force := false

	// オプション解析
	for i := 2; i < len(args); i++ {
		switch args[i] {
		case "--force", "-f":
			force = true
		default:
			fmt.Fprintf(os.Stderr, "エラー: 不明なオプション '%s'\n", args[i])
			printUsage()
			os.Exit(utils.ExitGeneralError)
		}
	}

	if err := utils.ValidateName(srcName, "サーバーテンプレート"); err != nil {
		fmt.Fprintln(os.Stderr, "エラー:", err)
		os.Exit(utils.ExitArgumentError)
	}

	if err := utils.ValidateName(destName, "サーバーテンプレート"); err != nil {
		fmt.Fprintln(os.Stderr, "エラー:", err)
		os.Exit(utils.ExitArgumentError)
	}

	serverManager := server.NewManager(cfg.ServersDir)
	if err := serverManager.Copy(srcName, destName, force); err != nil {
		fmt.Fprintf(os.Stderr, "エラー: %v\n", err)
		os.Exit(utils.ExitGeneralError)
	}
}

func printUsage() {
	fmt.Println(`mcpjson server copy - サーバーテンプレートをコピー

使用方法:
  mcpjson server copy <コピー元サーバー名> <コピー先サーバー名> [--force]

引数:
  <コピー元サーバー名>    コピー元のサーバーテンプレート名
  <コピー先サーバー名>    コピー先のサーバーテンプレート名

オプション:
  --force, -f           既存サーバーがある場合に確認なしで上書き

説明:
  既存のサーバーテンプレートを別名でコピーします。元のサーバーテンプレートはそのまま残ります。`)
}
