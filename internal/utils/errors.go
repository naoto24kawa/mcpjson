package utils

import (
	"fmt"
	"os"
)

const (
	ExitSuccess        = 0
	ExitGeneralError   = 1
	ExitResourceError  = 2
	ExitFileError      = 3
	ExitFormatError    = 4
	ExitEnvironment    = 5
	ExitServerError    = 6
	ExitArgumentError  = 7
	ExitReferenceError = 8
)

// HandleError handles errors with standardized error reporting and exit codes
func HandleError(err error, exitCode int) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "エラー:", err)
		os.Exit(exitCode)
	}
}

// HandleArgumentError handles argument validation errors
func HandleArgumentError(err error) {
	HandleError(err, ExitArgumentError)
}

// HandleGeneralError handles general errors
func HandleGeneralError(err error) {
	HandleError(err, ExitGeneralError)
}

// HandleEnvironmentError handles environment setup errors
func HandleEnvironmentError(err error) {
	HandleError(err, ExitEnvironment)
}

// ParseProfileName extracts profile name from args, using default if not provided
func ParseProfileName(args []string, defaultName string) (profileName string, argsOffset int) {
	if len(args) < 1 {
		fmt.Printf("プロファイル名が指定されていないため、デフォルト '%s' を使用します\n", defaultName)
		return defaultName, 0
	}
	// Check if the first argument is an option (starts with -)
	if len(args[0]) > 0 && args[0][0] == '-' {
		fmt.Printf("プロファイル名が指定されていないため、デフォルト '%s' を使用します\n", defaultName)
		return defaultName, 0
	}
	return args[0], 1
}

// ParseFlag parses a flag value from args at the given index
func ParseFlag(args []string, index int, flag string) (string, int, error) {
	if index+1 >= len(args) {
		return "", index, fmt.Errorf("%s オプションに値が指定されていません", flag)
	}
	return args[index+1], index + 1, nil
}

// ParseRenameArgs parses rename command arguments (special case)
func ParseRenameArgs(args []string, defaultName string) (oldName, newName string, argsOffset int, err error) {
	if len(args) < 1 {
		return "", "", 0, fmt.Errorf("新しいプロファイル名が指定されていません\n使用方法: mcpconfig rename [現在のプロファイル名] <新しいプロファイル名>")
	}
	
	if len(args) < 2 {
		oldName = defaultName
		newName = args[0]
		fmt.Printf("元のプロファイル名が指定されていないため、デフォルト '%s' を使用します\n", oldName)
		return oldName, newName, 1, nil
	}
	
	return args[0], args[1], 2, nil
}