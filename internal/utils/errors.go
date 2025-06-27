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