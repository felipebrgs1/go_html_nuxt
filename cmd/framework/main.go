package main

import (
	"fmt"
	"os"

	"go_template/pkg/cli"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Uso: framework <comando>")
		fmt.Println("Comandos disponíveis:")
		fmt.Println("  dev    Inicia o servidor de desenvolvimento com hot-reload")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "dev":
		if err := cli.RunDev(); err != nil {
			fmt.Fprintf(os.Stderr, "Erro: %v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Fprintf(os.Stderr, "Comando desconhecido: %s\n", os.Args[1])
		os.Exit(1)
	}
}
