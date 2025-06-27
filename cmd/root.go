package cmd

import (
	"fmt"
	"os"

	"github.com/naoto24kawa/mcpconfig/cmd/profile"
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
		handleApply(args)
	case "save":
		handleSave(args)
	case "create":
		handleCreate(args)
	case "list":
		handleList(args)
	case "delete":
		handleDelete(args)
	case "rename":
		handleRename(args)
	case "server":
		handleServer(args)
	default:
		fmt.Fprintf(os.Stderr, "エラー: 不明なコマンド '%s'\n", cmd)
		printUsage()
		os.Exit(utils.ExitGeneralError)
	}
}

func printUsage() {
	fmt.Println(`mcpconfig - MCP設定ファイル管理ツール

使用方法:
  mcpconfig <コマンド> [オプション] [引数]

コマンド:
  apply <プロファイル名> --path <パス>     プロファイルを指定パスに適用
  save <プロファイル名> --from <パス>      現在の設定をプロファイルとして保存
  create <プロファイル名>                   新規プロファイルを作成
  list [--detail]                          プロファイル一覧を表示
  delete <プロファイル名>                   プロファイルを削除
  rename <現在の名前> <新しい名前>          プロファイル名を変更
  server <サブコマンド>                     MCPサーバー管理

グローバルオプション:
  --help, -h      ヘルプを表示
  --version, -v   バージョンを表示

詳細は 'mcpconfig help <コマンド>' で確認してください`)
}

func handleApply(args []string) {
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
		fmt.Fprintln(os.Stderr, "エラー: 適用先のパスが指定されていません")
		fmt.Fprintln(os.Stderr, "使用方法: mcpconfig apply <プロファイル名> --path <パス>")
		os.Exit(utils.ExitArgumentError)
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

	if err := profile.Apply(cfg, profileName, targetPath); err != nil {
		fmt.Fprintln(os.Stderr, "エラー:", err)
		os.Exit(utils.ExitGeneralError)
	}
}

func handleSave(args []string) {
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
		fmt.Fprintln(os.Stderr, "エラー: 保存元のパスが指定されていません")
		fmt.Fprintln(os.Stderr, "使用方法: mcpconfig save <プロファイル名> --from <パス>")
		os.Exit(utils.ExitArgumentError)
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

	if err := profile.Save(cfg, profileName, fromPath, force); err != nil {
		fmt.Fprintln(os.Stderr, "エラー:", err)
		os.Exit(utils.ExitGeneralError)
	}
}

func handleCreate(args []string) {
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

func handleList(args []string) {
	detail := false

	for _, arg := range args {
		switch arg {
		case "--detail", "-d":
			detail = true
		}
	}

	cfg, err := config.New()
	if err != nil {
		fmt.Fprintln(os.Stderr, "エラー:", err)
		os.Exit(utils.ExitEnvironment)
	}

	if err := profile.List(cfg, detail); err != nil {
		fmt.Fprintln(os.Stderr, "エラー:", err)
		os.Exit(utils.ExitGeneralError)
	}
}

func handleDelete(args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "エラー: プロファイル名が指定されていません")
		fmt.Fprintln(os.Stderr, "使用方法: mcpconfig delete <プロファイル名>")
		os.Exit(utils.ExitArgumentError)
	}

	profileName := args[0]
	force := false

	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "--force", "-f":
			force = true
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

	if err := profile.Delete(cfg, profileName, force); err != nil {
		fmt.Fprintln(os.Stderr, "エラー:", err)
		os.Exit(utils.ExitGeneralError)
	}
}

func handleRename(args []string) {
	if len(args) < 2 {
		fmt.Fprintln(os.Stderr, "エラー: プロファイル名が指定されていません")
		fmt.Fprintln(os.Stderr, "使用方法: mcpconfig rename <現在の名前> <新しい名前>")
		os.Exit(utils.ExitArgumentError)
	}

	oldName := args[0]
	newName := args[1]
	force := false

	for i := 2; i < len(args); i++ {
		switch args[i] {
		case "--force", "-f":
			force = true
		}
	}

	if err := utils.ValidateName(oldName, "プロファイル"); err != nil {
		fmt.Fprintln(os.Stderr, "エラー:", err)
		os.Exit(utils.ExitArgumentError)
	}

	if err := utils.ValidateName(newName, "プロファイル"); err != nil {
		fmt.Fprintln(os.Stderr, "エラー:", err)
		os.Exit(utils.ExitArgumentError)
	}

	cfg, err := config.New()
	if err != nil {
		fmt.Fprintln(os.Stderr, "エラー:", err)
		os.Exit(utils.ExitEnvironment)
	}

	if err := profile.Rename(cfg, oldName, newName, force); err != nil {
		fmt.Fprintln(os.Stderr, "エラー:", err)
		os.Exit(utils.ExitGeneralError)
	}
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