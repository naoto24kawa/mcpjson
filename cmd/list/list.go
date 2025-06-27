package list

import (
	"fmt"
	"os"

	"github.com/naoto24kawa/mcpconfig/internal/config"
	"github.com/naoto24kawa/mcpconfig/cmd/profile"
	"github.com/naoto24kawa/mcpconfig/internal/utils"
)

func Execute(args []string) {
	detail := false

	for _, arg := range args {
		switch arg {
		case "--detail", "-d":
			detail = true
		}
	}

	cfg, err := config.New()
	if err != nil {
		fmt.Fprintln(os.Stderr, "エラー:", err)
		os.Exit(utils.ExitEnvironment)
	}

	if err := profile.List(cfg, detail); err != nil {
		fmt.Fprintln(os.Stderr, "エラー:", err)
		os.Exit(utils.ExitGeneralError)
	}
}