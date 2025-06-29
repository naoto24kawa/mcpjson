package save

import (
	"fmt"
	"os"

	"github.com/naoto24kawa/mcpconfig/cmd/profile"
	"github.com/naoto24kawa/mcpconfig/internal/config"
	"github.com/naoto24kawa/mcpconfig/internal/utils"
)

// findMCPConfigFile searches for MCP configuration file in default locations
func findMCPConfigFile() (string, error) {
	localPath := config.GetDefaultMCPPath()
	if _, err := os.Stat(localPath); err == nil {
		return localPath, nil
	}

	foundPath := config.FindMCPConfigPath()
	if foundPath == "" {
		return "", fmt.Errorf("MCP設定ファイルが見つかりません\n使用方法: mcpconfig save [プロファイル名] --from <パス>")
	}

	fmt.Printf("MCP設定ファイルを自動検出しました: %s\n", foundPath)
	return foundPath, nil
}

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
		var err error
		fromPath, err = findMCPConfigFile()
		if err != nil {
			fmt.Fprintln(os.Stderr, "エラー:", err)
			os.Exit(utils.ExitArgumentError)
		}
	}

	utils.HandleArgumentError(utils.ValidateName(profileName, "プロファイル"))

	cfg, err := config.New()
	utils.HandleEnvironmentError(err)

	utils.HandleGeneralError(profile.Save(cfg, profileName, fromPath, force))
}
