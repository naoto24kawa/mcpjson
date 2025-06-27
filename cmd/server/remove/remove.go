package remove

import (
	"fmt"
	"os"

	"github.com/naoto24kawa/mcpconfig/internal/config"
	"github.com/naoto24kawa/mcpconfig/internal/profile"
	"github.com/naoto24kawa/mcpconfig/internal/utils"
)

func Execute(cfg *config.Config, args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "エラー: サーバー名が指定されていません")
		os.Exit(utils.ExitArgumentError)
	}

	serverName := args[0]
	var profileName string

	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "--from", "-f":
			if i+1 >= len(args) {
				fmt.Fprintln(os.Stderr, "エラー: --from オプションに値が指定されていません")
				os.Exit(utils.ExitArgumentError)
			}
			profileName = args[i+1]
			i++
		}
	}

	if profileName == "" {
		fmt.Fprintln(os.Stderr, "エラー: 削除元のプロファイル名が指定されていません")
		fmt.Fprintln(os.Stderr, "使用方法: mcpconfig server remove <サーバー名> --from <プロファイル名>")
		os.Exit(utils.ExitArgumentError)
	}

	if err := utils.ValidateName(profileName, "プロファイル"); err != nil {
		fmt.Fprintln(os.Stderr, "エラー:", err)
		os.Exit(utils.ExitArgumentError)
	}

	profileManager := profile.NewManager(cfg.ProfilesDir)
	if err := profileManager.RemoveServer(profileName, serverName); err != nil {
		fmt.Fprintln(os.Stderr, "エラー:", err)
		os.Exit(utils.ExitGeneralError)
	}
}