package interaction

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func IsInteractive() bool {
	stat, _ := os.Stdin.Stat()
	return (stat.Mode() & os.ModeCharDevice) != 0
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

func Confirm(message string) bool {
	if !IsInteractive() {
		return false
	}

	fmt.Printf("%s (y/N): ", message)

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return false
	}

	input = strings.TrimSpace(strings.ToLower(input))
	return input == "y" || input == "yes"
}
