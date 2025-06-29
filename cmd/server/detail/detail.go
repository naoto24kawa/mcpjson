package detail

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/naoto24kawa/mcpconfig/internal/config"
	"github.com/naoto24kawa/mcpconfig/internal/server"
	"github.com/naoto24kawa/mcpconfig/internal/utils"
)

func Execute(cfg *config.Config, args []string) {
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "エラー: サーバー名を指定してください\n")
		fmt.Println("使用方法: mcpconfig server detail <サーバー名>")
		os.Exit(utils.ExitGeneralError)
	}

	serverName := args[0]
	if err := showServerDetail(cfg, serverName); err != nil {
		fmt.Fprintln(os.Stderr, "エラー:", err)
		os.Exit(utils.ExitGeneralError)
	}
}

func showServerDetail(cfg *config.Config, serverName string) error {
	templatePath := filepath.Join(cfg.ServersDir, serverName+config.FileExtension)
	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		return fmt.Errorf("サーバーテンプレート '%s' が見つかりません", serverName)
	}

	var targetTemplate server.ServerTemplate
	if err := utils.LoadJSON(templatePath, &targetTemplate); err != nil {
		return fmt.Errorf("サーバーテンプレートの読み込みに失敗しました: %v", err)
	}

	jsonData, err := json.MarshalIndent(targetTemplate, "", "  ")
	if err != nil {
		return fmt.Errorf("JSONの生成に失敗しました: %v", err)
	}

	fmt.Println(string(jsonData))
	return nil
}
