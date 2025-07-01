package merge

import (
	"fmt"
	"os"

	"github.com/naoto24kawa/mcpjson/cmd/profile"
	"github.com/naoto24kawa/mcpjson/internal/config"
	"github.com/naoto24kawa/mcpjson/internal/utils"
)

func Execute(args []string) {
	if len(args) < 2 {
		fmt.Fprintln(os.Stderr, "エラー: 使用方法: mcpjson merge <合成先プロファイル名> <ソースプロファイル1> [ソースプロファイル2] ... [--force]")
		os.Exit(utils.ExitArgumentError)
	}

	destName := args[0]
	sourceNames := []string{}
	force := false

	// 引数を解析
	for i := 1; i < len(args); i++ {
		if args[i] == "--force" || args[i] == "-f" {
			force = true
		} else {
			sourceNames = append(sourceNames, args[i])
		}
	}

	if len(sourceNames) == 0 {
		fmt.Fprintln(os.Stderr, "エラー: 少なくとも1つのソースプロファイルを指定してください")
		os.Exit(utils.ExitArgumentError)
	}

	// 合成先プロファイル名の検証
	utils.HandleArgumentError(utils.ValidateName(destName, "プロファイル"))

	// 各ソースプロファイル名の検証
	for _, sourceName := range sourceNames {
		utils.HandleArgumentError(utils.ValidateName(sourceName, "プロファイル"))
	}

	cfg, err := config.New()
	utils.HandleEnvironmentError(err)

	utils.HandleGeneralError(profile.Merge(cfg, destName, sourceNames, force))
}
