package reset

import (
	"fmt"
	"os"

	"github.com/naoto24kawa/mcpconfig/internal/config"
	"github.com/naoto24kawa/mcpconfig/internal/interaction"
	"github.com/naoto24kawa/mcpconfig/internal/profile"
	"github.com/naoto24kawa/mcpconfig/internal/server"
	"github.com/naoto24kawa/mcpconfig/internal/utils"
)

func Execute(cfg *config.Config, args []string) {
	if len(args) == 0 {
		PrintUsage()
		os.Exit(0)
	}

	cmd := args[0]
	cmdArgs := args[1:]

	var force bool
	for i, arg := range cmdArgs {
		if arg == "--force" || arg == "-f" {
			force = true
			cmdArgs = append(cmdArgs[:i], cmdArgs[i+1:]...)
			break
		}
	}

	switch cmd {
	case "all":
		if err := resetAll(force); err != nil {
			fmt.Fprintln(os.Stderr, "エラー:", err)
			os.Exit(utils.ExitGeneralError)
		}
	case "profiles":
		if err := resetProfiles(force); err != nil {
			fmt.Fprintln(os.Stderr, "エラー:", err)
			os.Exit(utils.ExitGeneralError)
		}
	case "servers":
		if err := resetServers(force); err != nil {
			fmt.Fprintln(os.Stderr, "エラー:", err)
			os.Exit(utils.ExitGeneralError)
		}
	default:
		fmt.Fprintf(os.Stderr, "エラー: 不明なリセットコマンド '%s'\n", cmd)
		PrintUsage()
		os.Exit(utils.ExitGeneralError)
	}
}

func PrintUsage() {
	fmt.Println(`mcpconfig reset - 開発用設定のリセット

使用方法:
  mcpconfig reset <サブコマンド> [オプション]

サブコマンド:
  all       すべての設定をリセット (プロファイル + サーバーテンプレート)
  profiles  すべてのプロファイルを削除
  servers   すべてのサーバーテンプレートを削除

オプション:
  --force, -f  確認なしで実行

例:
  mcpconfig reset all                すべての設定をリセット
  mcpconfig reset profiles --force   確認なしですべてのプロファイルを削除
  mcpconfig reset servers            すべてのサーバーテンプレートを削除`)
}


func resetAll(force bool) error {
	cfg, err := config.New()
	if err != nil {
		return fmt.Errorf("設定の初期化に失敗しました: %w", err)
	}
	
	if !force {
		fmt.Println("以下の設定がすべて削除されます:")
		fmt.Println("  - すべてのプロファイル")
		fmt.Println("  - すべてのサーバーテンプレート")
		fmt.Println()
		
		if !interaction.Confirm("本当にすべての設定をリセットしますか？") {
			fmt.Println("リセットをキャンセルしました")
			return nil
		}
	}

	if err := resetProfilesWithConfig(cfg, true); err != nil {
		fmt.Printf("プロファイルのリセットに失敗しました: %v\n", err)
	}

	if err := resetServersWithConfig(cfg, true); err != nil {
		fmt.Printf("サーバーテンプレートのリセットに失敗しました: %v\n", err)
	}

	fmt.Println("すべての設定をリセットしました")
	return nil
}

func resetProfiles(force bool) error {
	cfg, err := config.New()
	if err != nil {
		return fmt.Errorf("設定の初期化に失敗しました: %w", err)
	}
	
	return resetProfilesWithConfig(cfg, force)
}

func resetServers(force bool) error {
	cfg, err := config.New()
	if err != nil {
		return fmt.Errorf("設定の初期化に失敗しました: %w", err)
	}
	
	return resetServersWithConfig(cfg, force)
}

func resetProfilesWithConfig(cfg *config.Config, force bool) error {
	profileManager := profile.NewManager(cfg.ProfilesDir)
	return profileManager.Reset(force)
}

func resetServersWithConfig(cfg *config.Config, force bool) error {
	serverManager := server.NewManager(cfg.ServersDir)
	return serverManager.Reset(force)
}