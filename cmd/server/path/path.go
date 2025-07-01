package path

import (
	"fmt"
	"os"

	"github.com/naoto24kawa/mcpjson/internal/config"
	"github.com/naoto24kawa/mcpjson/internal/server"
	"github.com/naoto24kawa/mcpjson/internal/utils"
)

func Execute(cfg *config.Config, args []string) {
	if len(args) != 1 {
		fmt.Fprintf(os.Stderr, "エラー: サーバーテンプレート名を指定してください\n")
		printUsage()
		os.Exit(utils.ExitGeneralError)
	}

	templateName := args[0]

	serverManager := server.NewManager(cfg.ServersDir)
	templatePath, err := serverManager.GetTemplatePath(templateName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "エラー: サーバーテンプレートパスの取得に失敗しました: %v\n", err)
		os.Exit(utils.ExitGeneralError)
	}

	fmt.Print(templatePath)
}

func printUsage() {
	fmt.Println(`mcpjson server path - サーバーテンプレートパス表示

使用方法:
  mcpjson server path <サーバーテンプレート名>

引数:
  <サーバーテンプレート名>    パスを表示するサーバーテンプレート名

説明:
  指定されたサーバーテンプレートファイルの絶対パスを表示します。`)
}
