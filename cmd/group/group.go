package group

import (
	"fmt"
	"os"

	"github.com/naoto24kawa/mcpjson/internal/config"
	"github.com/naoto24kawa/mcpjson/internal/utils"
)

func Execute(cfg *config.Config, args []string) {
	if len(args) == 0 {
		PrintUsage()
		os.Exit(0)
	}

	subCmd := args[0]

	switch subCmd {
	case "list":
		fmt.Println("グループ機能は現在開発中です")
	default:
		fmt.Fprintf(os.Stderr, "エラー: 不明なサブコマンド 'group %s'\n", subCmd)
		PrintUsage()
		os.Exit(utils.ExitGeneralError)
	}
}

func PrintUsage() {
	fmt.Println(`mcpjson group - グループ管理

使用方法:
  mcpjson group <サブコマンド> [オプション]

サブコマンド:
  list [--detail]                                    グループ一覧表示

注意: グループ機能は現在開発中です`)
}
