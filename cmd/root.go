package cmd

import (
	"fmt"
	"os"

	"github.com/naoto24kawa/mcpconfig/cmd/apply"
	"github.com/naoto24kawa/mcpconfig/cmd/create"
	"github.com/naoto24kawa/mcpconfig/cmd/delete"
	"github.com/naoto24kawa/mcpconfig/cmd/list"
	"github.com/naoto24kawa/mcpconfig/cmd/rename"
	"github.com/naoto24kawa/mcpconfig/cmd/reset"
	"github.com/naoto24kawa/mcpconfig/cmd/save"
	"github.com/naoto24kawa/mcpconfig/cmd/server"
	"github.com/naoto24kawa/mcpconfig/internal/config"
	"github.com/naoto24kawa/mcpconfig/internal/utils"
)

var Version = "dev"

func Execute() {
	if len(os.Args) == 1 {
		printUsage()
		os.Exit(0)
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	switch cmd {
	case "help", "-h", "--help":
		printUsage()
		os.Exit(0)
	case "version", "-v", "--version":
		fmt.Printf("mcpconfig version %s\n", Version)
		os.Exit(0)
	case "apply":
		apply.Execute(args)
	case "save":
		save.Execute(args)
	case "create":
		create.Execute(args)
	case "list":
		list.Execute(args)
	case "delete":
		delete.Execute(args)
	case "rename":
		rename.Execute(args)
	case "server":
		handleServer(args)
	case "reset":
		handleReset(args)
	default:
		fmt.Fprintf(os.Stderr, "エラー: 不明なコマンド '%s'\n", cmd)
		printUsage()
		os.Exit(utils.ExitGeneralError)
	}
}

func printUsage() {
	fmt.Printf(`mcpconfig - MCP設定ファイル管理ツール

使用方法:
  mcpconfig <コマンド> [オプション] [引数]

コマンド:
  apply [プロファイル名] --to <パス>         プロファイルを指定パスに適用 (デフォルト: %s)
  save [プロファイル名] --from <パス>        現在の設定をプロファイルとして保存 (デフォルト: %s)
  create [プロファイル名]                    新規プロファイルを作成 (デフォルト: %s)
  list [--detail]                           プロファイル一覧を表示
  delete [プロファイル名]                    プロファイルを削除 (デフォルト: %s)
  rename [現在の名前] <新しい名前>           プロファイル名を変更 (デフォルト: %s)
  server <サブコマンド>                      MCPサーバー管理
  reset <サブコマンド>                       開発用設定のリセット

注意: []で囲まれた引数は省略可能で、省略時はデフォルトプロファイル名 '%s' が使用されます

グローバルオプション:
  --help, -h      ヘルプを表示
  --version, -v   バージョンを表示

詳細は 'mcpconfig help <コマンド>' で確認してください`, 
		config.DefaultProfileName, 
		config.DefaultProfileName, 
		config.DefaultProfileName, 
		config.DefaultProfileName, 
		config.DefaultProfileName, 
		config.DefaultProfileName)
}


func handleServer(args []string) {
	if len(args) == 0 {
		server.PrintUsage()
		os.Exit(0)
	}

	cfg, err := config.New()
	if err != nil {
		fmt.Fprintln(os.Stderr, "エラー:", err)
		os.Exit(utils.ExitEnvironment)
	}

	server.Execute(cfg, args)
}

func handleReset(args []string) {
	if len(args) == 0 {
		reset.PrintUsage()
		os.Exit(0)
	}

	cfg, err := config.New()
	if err != nil {
		fmt.Fprintln(os.Stderr, "エラー:", err)
		os.Exit(utils.ExitEnvironment)
	}

	reset.Execute(cfg, args)
}