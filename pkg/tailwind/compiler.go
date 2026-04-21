package tailwind

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// Precisa de fmt para debug

// Compile roda `tailwindcss -i input.css -o public/styles.css` ou build único
func Compile(root string) error {
	input := filepath.Join(root, "assets", "global.css")
	output := filepath.Join(root, "public", "styles.css")
	config := filepath.Join(root, "tailwind.config.js")

	// Se não existe input.css, não faz nada
	if _, err := os.Stat(input); os.IsNotExist(err) {
		return nil
	}

	args := []string{"-i", input, "-o", output}
	if _, err := os.Stat(config); err == nil {
		args = append(args, "-c", config)
	}

	// Procura tailwindcss no PATH ou no diretório raiz
	bin := findTailwindBin(root)
	if bin == "" {
		return fmt.Errorf("tailwindcss não encontrado (instale com: curl -sLO https://github.com/tailwindlabs/tailwindcss/releases/latest/download/tailwindcss-linux-x64 && chmod +x tailwindcss-linux-x64 && mv tailwindcss-linux-x64 tailwindcss)")
	}

	cmd := exec.Command(bin, args...)
	cmd.Dir = root
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("tailwindcss falhou: %w", err)
	}
	return nil
}

func findTailwindBin(root string) string {
	// Tenta no PATH
	if path, err := exec.LookPath("tailwindcss"); err == nil {
		return path
	}
	// Tenta no diretório raiz do projeto (retorna path absoluto)
	local := filepath.Join(root, "tailwindcss")
	if info, err := os.Stat(local); err == nil && !info.IsDir() {
		abs, _ := filepath.Abs(local)
		if abs != "" {
			return abs
		}
		return local
	}
	return ""
}

// HasTailwind verifica se existe public/input.css
func HasTailwind(root string) bool {
	_, err := os.Stat(filepath.Join(root, "assets", "global.css"))
	return err == nil
}
