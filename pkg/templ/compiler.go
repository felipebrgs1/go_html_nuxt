package templ

import (
	"fmt"
	"os"
	"os/exec"
)

// Compile roda `templ generate` no diretório raiz do projeto
func Compile(root string) error {
	cmd := exec.Command("templ", "generate")
	cmd.Dir = root
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("templ generate falhou: %w", err)
	}
	return nil
}

// HasTempl verifica se o projeto usa templ (tem arquivos .templ)
func HasTempl(root string) bool {
	return exists(root, "app/layouts") || exists(root, "app/pages")
}

func exists(root, dir string) bool {
	_, err := os.Stat(root + "/" + dir)
	return err == nil
}
