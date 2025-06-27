package delete

import (
	"fmt"
	"os"

	"github.com/naoto24kawa/mcpconfig/internal/config"
	"github.com/naoto24kawa/mcpconfig/cmd/profile"
	"github.com/naoto24kawa/mcpconfig/internal/utils"
)

func Execute(args []string) {
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