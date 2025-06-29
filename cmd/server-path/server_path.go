package serverpath

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/naoto24kawa/mcpconfig/internal/config"
	"github.com/naoto24kawa/mcpconfig/internal/server"
)

var ServerPathCmd = &cobra.Command{
	Use:   "server-path <template_name>",
	Short: "サーバーテンプレートファイルのパスを表示します",
	Long:  "指定されたサーバーテンプレートファイルの絶対パスを表示します。",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		templateName := args[0]

		cfg, err := config.New()
		if err != nil {
			return fmt.Errorf("設定の読み込みに失敗しました: %w", err)
		}

		serverManager := server.NewManager(cfg.ServersDir)
		templatePath, err := serverManager.GetTemplatePath(templateName)
		if err != nil {
			return fmt.Errorf("サーバーテンプレートパスの取得に失敗しました: %w", err)
		}

		fmt.Print(templatePath)
		return nil
	},
}
