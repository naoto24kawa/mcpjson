package save

import (
	"fmt"
	"os"

	"github.com/naoto24kawa/mcpconfig/internal/config"
	"github.com/naoto24kawa/mcpconfig/cmd/profile"
	"github.com/naoto24kawa/mcpconfig/internal/utils"
)

func Execute(args []string) {
	profileName, argsOffset := utils.ParseProfileName(args, config.DefaultProfileName)
	var fromPath string
	force := false

	for i := argsOffset; i < len(args); i++ {
		switch args[i] {
		case "--from", "-f":
			var err error
			fromPath, i, err = utils.ParseFlag(args, i, "--from")
			utils.HandleArgumentError(err)
		case "--force", "-F":
			force = true
		}
	}

	if fromPath == "" {
		fromPath = config.FindMCPConfigPath()
		if fromPath == "" {
			fmt.Fprintln(os.Stderr, "エラー: MCP設定ファイルが見つかりません")
			fmt.Fprintln(os.Stderr, "使用方法: mcpconfig save <プロファイル名> --from <パス>")
			os.Exit(utils.ExitArgumentError)
		}
		fmt.Printf("MCP設定ファイルを自動検出しました: %s\n", fromPath)
	}

	utils.HandleArgumentError(utils.ValidateName(profileName, "プロファイル"))

	cfg, err := config.New()
	utils.HandleEnvironmentError(err)

	utils.HandleGeneralError(profile.Save(cfg, profileName, fromPath, force))
}