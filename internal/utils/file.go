package utils

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func LoadJSON(path string, v interface{}) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	
	decoder := json.NewDecoder(file)
	return decoder.Decode(v)
}

func SaveJSON(path string, v interface{}) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(v)
}

func ConfirmOverwrite(resourceType, name string) bool {
	if !IsInteractive() {
		return false
	}
	
	fmt.Printf("警告: %s '%s' は既に存在します\n", resourceType, name)
	fmt.Print("上書きしますか？ (y/N): ")
	
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return false
	}
	
	input = strings.TrimSpace(strings.ToLower(input))
	return input == "y" || input == "yes"
}

func IsInteractive() bool {
	stat, _ := os.Stdin.Stat()
	return (stat.Mode() & os.ModeCharDevice) != 0
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