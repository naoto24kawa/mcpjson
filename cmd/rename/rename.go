package rename

import (
	"fmt"
	"os"

	"github.com/naoto24kawa/mcpconfig/internal/config"
	"github.com/naoto24kawa/mcpconfig/cmd/profile"
	"github.com/naoto24kawa/mcpconfig/internal/utils"
)

func Execute(args []string) {
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