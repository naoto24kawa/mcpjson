package utils

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/tidwall/jsonc"
)

func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func LoadJSON(path string, v interface{}) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	// JSONC形式（コメント付きJSON）を標準JSONに変換
	jsonData := jsonc.ToJSON(data)

	return json.Unmarshal(jsonData, v)
}

func SaveJSON(path string, v interface{}) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

func LoadEnvFile(path string) (map[string]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("環境変数ファイルが見つかりません: '%s'", path)
	}
	defer file.Close()

	envMap := make(map[string]string)
	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("環境変数ファイルの形式が不正です: '%s' 行%d", path, lineNum)
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		value = strings.Trim(value, `"'`)

		envMap[key] = value
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("環境変数ファイルの読み込みに失敗しました: '%s'", path)
	}

	return envMap, nil
}
