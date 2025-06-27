package utils

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	namePattern    = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	reservedWords  = []string{"help", "version", "list", "server", "apply", "save", "create", "delete", "rename", "add", "remove", "show"}
	envKeyPattern  = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_]*$`)
)

const MaxNameLength = 50

func ValidateName(name string, resourceType string) error {
	if name == "" {
		return fmt.Errorf("%s名が指定されていません", resourceType)
	}
	
	if len(name) > MaxNameLength {
		return fmt.Errorf("%s名は%d文字以内で指定してください", resourceType, MaxNameLength)
	}
	
	if !namePattern.MatchString(name) {
		return fmt.Errorf("%s名に使用できない文字が含まれています（使用可能: 英数字、ハイフン、アンダースコア）", resourceType)
	}
	
	for _, reserved := range reservedWords {
		if name == reserved {
			return fmt.Errorf("%s名に予約語 '%s' は使用できません", resourceType, reserved)
		}
	}
	
	return nil
}

func ParseEnvVars(envStr string) (map[string]string, error) {
	if envStr == "" {
		return nil, nil
	}
	
	envMap := make(map[string]string)
	pairs := strings.Split(envStr, ",")
	
	for _, pair := range pairs {
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("環境変数の形式が不正です: '%s'", pair)
		}
		
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		
		if !envKeyPattern.MatchString(key) {
			return nil, fmt.Errorf("環境変数名が不正です: '%s'", key)
		}
		
		envMap[key] = value
	}
	
	return envMap, nil
}

func ParseArgs(argsStr string) []string {
	if argsStr == "" {
		return nil
	}
	
	args := strings.Split(argsStr, ",")
	result := make([]string, 0, len(args))
	
	for _, arg := range args {
		trimmed := strings.TrimSpace(arg)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	
	return result
}