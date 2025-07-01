package detail

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/naoto24kawa/mcpjson/internal/config"
	"github.com/naoto24kawa/mcpjson/internal/profile"
	"github.com/naoto24kawa/mcpjson/internal/utils"
)

func Execute(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("使用方法: mcpconfig detail <プロファイル名>")
	}

	return showProfileDetail(args[0])
}

func showProfileDetail(profileName string) error {
	cfg, err := config.New()
	if err != nil {
		return fmt.Errorf("設定の初期化に失敗しました: %v", err)
	}

	profilePath := filepath.Join(cfg.ProfilesDir, profileName+config.FileExtension)
	if _, err := os.Stat(profilePath); os.IsNotExist(err) {
		return fmt.Errorf("プロファイル '%s' が見つかりません", profileName)
	}

	var targetProfile profile.Profile
	if err := utils.LoadJSON(profilePath, &targetProfile); err != nil {
		return fmt.Errorf("プロファイルの読み込みに失敗しました: %v", err)
	}

	jsonData, err := json.MarshalIndent(targetProfile, "", "  ")
	if err != nil {
		return fmt.Errorf("JSONの生成に失敗しました: %v", err)
	}

	fmt.Println(string(jsonData))
	return nil
}
