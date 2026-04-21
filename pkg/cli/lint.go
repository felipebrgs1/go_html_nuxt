package cli

import (
	"fmt"
	"os"

	"go_template/pkg/linter"
)

func RunLint() error {
	fmt.Println("Executando linter do framework...")

	l := linter.New(".")
	result, err := l.Run()
	if err != nil {
		return fmt.Errorf("falha ao executar linter: %w", err)
	}

	if len(result.Issues) == 0 {
		fmt.Println("Nenhum problema encontrado.")
		return nil
	}

	fmt.Printf("%d problema(s) encontrado(s):\n\n", len(result.Issues))

	for _, issue := range result.Issues {
		sevIcon := "[INFO]"
		switch issue.Severity {
		case linter.Error:
			sevIcon = "[ERRO]"
		case linter.Warning:
			sevIcon = "[AVISO]"
		}
		fmt.Printf("%s [%s] %s:%d  %s\n   -> %s\n\n", sevIcon, issue.Rule, issue.File, issue.Line, issue.Severity, issue.Message)
	}

	if result.HasErrors() {
		os.Exit(1)
	}
	return nil
}
