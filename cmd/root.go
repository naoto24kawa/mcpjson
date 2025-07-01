package cmd

import (
	"fmt"
	"os"

	"github.com/naoto24kawa/mcpjson/cmd/apply"
	"github.com/naoto24kawa/mcpjson/cmd/copy"
	"github.com/naoto24kawa/mcpjson/cmd/create"
	"github.com/naoto24kawa/mcpjson/cmd/delete"
	"github.com/naoto24kawa/mcpjson/cmd/detail"
	"github.com/naoto24kawa/mcpjson/cmd/list"
	"github.com/naoto24kawa/mcpjson/cmd/merge"
	"github.com/naoto24kawa/mcpjson/cmd/path"
	"github.com/naoto24kawa/mcpjson/cmd/rename"
	"github.com/naoto24kawa/mcpjson/cmd/reset"
	"github.com/naoto24kawa/mcpjson/cmd/save"
	"github.com/naoto24kawa/mcpjson/cmd/server"
	serverpath "github.com/naoto24kawa/mcpjson/cmd/server-path"
	"github.com/naoto24kawa/mcpjson/internal/config"
	"github.com/naoto24kawa/mcpjson/internal/utils"
)

var Version = "dev"

type CommandRouter struct{}

func Execute() {
	router := &CommandRouter{}

	if len(os.Args) == 1 {
		printUsage()
		os.Exit(0)
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	router.Route(cmd, args)
}

func (r *CommandRouter) Route(cmd string, args []string) {
	switch cmd {
	case "help", "-h", "--help":
		r.handleHelp()
	case "version", "-v", "--version":
		r.handleVersion()
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
	case "copy":
		copy.Execute(args)
	case "merge":
		merge.Execute(args)
	case "detail":
		r.handleDetail(args)
	case "server":
		r.handleServer(args)
	case "reset":
		r.handleReset(args)
	case "path":
		r.handlePath(args)
	case "server-path":
		r.handleServerPath(args)
	default:
		r.handleUnknownCommand(cmd)
	}
}

func (r *CommandRouter) handleHelp() {
	printUsage()
	os.Exit(0)
}

func (r *CommandRouter) handleVersion() {
	fmt.Printf("mcpconfig version %s\n", Version)
	os.Exit(0)
}

func (r *CommandRouter) handleDetail(args []string) {
	if err := detail.Execute(args); err != nil {
		fmt.Fprintln(os.Stderr, "エラー:", err)
		os.Exit(utils.ExitGeneralError)
	}
}

func (r *CommandRouter) handlePath(args []string) {
	path.PathCmd.SetArgs(args)
	if err := path.PathCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "エラー:", err)
		os.Exit(utils.ExitGeneralError)
	}
}

func (r *CommandRouter) handleServerPath(args []string) {
	serverpath.ServerPathCmd.SetArgs(args)
	if err := serverpath.ServerPathCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "エラー:", err)
		os.Exit(utils.ExitGeneralError)
	}
}

func (r *CommandRouter) handleUnknownCommand(cmd string) {
	fmt.Fprintf(os.Stderr, "エラー: 不明なコマンド '%s'\n", cmd)
	printUsage()
	os.Exit(utils.ExitGeneralError)
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
  copy [コピー元] <コピー先>                 プロファイルをコピー (デフォルト: %s)
  merge <合成先> <ソース1> [ソース2]...      複数のプロファイルを合成
  path [プロファイル名]                      プロファイルファイルのパスを表示 (デフォルト: %s)
  detail <プロファイル名>                    プロファイルの詳細を表示
  server <サブコマンド>                      MCPサーバー管理
  server-path <テンプレート名>               サーバーテンプレートファイルのパスを表示
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
		config.DefaultProfileName,
		config.DefaultProfileName,
		config.DefaultProfileName)
}

func (r *CommandRouter) handleServer(args []string) {
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

func (r *CommandRouter) handleReset(args []string) {
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
