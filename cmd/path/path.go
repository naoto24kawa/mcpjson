package path

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/naoto24kawa/mcpconfig/cmd/profile"
	"github.com/naoto24kawa/mcpconfig/internal/config"
)

var PathCmd = &cobra.Command{
	Use:   "path [profile_name]",
	Short: "プロファイルファイルのパスを表示します",
	Long:  "指定されたプロファイルファイルの絶対パスを表示します。プロファイル名を省略した場合はデフォルトプロファイルのパスを表示します。",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		profileName := config.DefaultProfileName
		if len(args) > 0 {
			profileName = args[0]
		}

		cfg, err := config.New()
		if err != nil {
			return fmt.Errorf("設定の読み込みに失敗しました: %w", err)
		}

		profilePath, err := profile.GetProfilePath(cfg, profileName)
		if err != nil {
			return fmt.Errorf("プロファイルパスの取得に失敗しました: %w", err)
		}

		fmt.Print(profilePath)
		return nil
	},
}