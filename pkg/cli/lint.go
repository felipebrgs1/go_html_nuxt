package cli

import (
	"fmt"
	"os"

	"go_template/pkg/linter"
)

func RunLint() error {
	fmt.Println("Running framework linter...")

	l := linter.New(".")
	result, err := l.Run()
	if err != nil {
		return fmt.Errorf("failed to run linter: %w", err)
	}

	if len(result.Issues) == 0 {
		fmt.Println("No issues found.")
		return nil
	}

	fmt.Printf("%d issue(s) found:\n\n", len(result.Issues))

	for _, issue := range result.Issues {
		sevIcon := "[INFO]"
		switch issue.Severity {
		case linter.Error:
			sevIcon = "[ERROR]"
		case linter.Warning:
			sevIcon = "[WARNING]"
		}
		fmt.Printf("%s [%s] %s:%d  %s\n   -> %s\n\n", sevIcon, issue.Rule, issue.File, issue.Line, issue.Severity, issue.Message)
	}

	if result.HasErrors() {
		os.Exit(1)
	}
	return nil
}
