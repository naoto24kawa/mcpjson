package server

import (
	"fmt"
	"os"

	"github.com/naoto24kawa/mcpconfig/cmd/server/add"
	"github.com/naoto24kawa/mcpconfig/cmd/server/delete"
	"github.com/naoto24kawa/mcpconfig/cmd/server/detail"
	"github.com/naoto24kawa/mcpconfig/cmd/server/list"
	"github.com/naoto24kawa/mcpconfig/cmd/server/remove"
	"github.com/naoto24kawa/mcpconfig/cmd/server/rename"
	"github.com/naoto24kawa/mcpconfig/cmd/server/save"
	"github.com/naoto24kawa/mcpconfig/cmd/server/show"
	"github.com/naoto24kawa/mcpconfig/internal/config"
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
		save.Execute(cfg, subArgs)
	case "list":
		list.Execute(cfg, subArgs)
	case "delete":
		delete.Execute(cfg, subArgs)
	case "rename":
		rename.Execute(cfg, subArgs)
	case "add":
		add.Execute(cfg, subArgs)
	case "remove":
		remove.Execute(cfg, subArgs)
	case "show":
		show.Execute(cfg, subArgs)
	case "detail":
		detail.Execute(cfg, subArgs)
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
  save <サーバー名> --server <サーバー名> --from <パス>    設定ファイルからサーバー保存
  save <サーバー名> --command <コマンド> [オプション]      手動でサーバー作成
  list [--detail]                                      サーバー一覧表示
  delete <サーバー名>                                   サーバー削除
  rename <現在のサーバー名> <新しいサーバー名>              サーバー名変更
  add <サーバー名> --to <プロファイル名>                  プロファイルにサーバー追加
  remove <サーバー名> --from <プロファイル名>             プロファイルからサーバー削除
  show --from <パス> [--server <サーバー名>]             設定ファイルからサーバー情報表示
  detail <サーバー名>                                   サーバーテンプレートの詳細を表示`)
}
