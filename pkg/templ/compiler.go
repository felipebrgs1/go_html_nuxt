package templ

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Compile roda `templ generate` no diretório raiz do projeto
func Compile(root string) error {
	bin := findTemplBin()
	if bin == "" {
		return fmt.Errorf("templ não encontrado (instale com: go install github.com/a-h/templ/cmd/templ@latest)")
	}

	cmd := exec.Command(bin, "generate")
	cmd.Dir = root
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("templ generate falhou: %w", err)
	}
	return nil
}

func findTemplBin() string {
	if path, err := exec.LookPath("templ"); err == nil {
		return path
	}

	// Tenta no GOPATH/bin
	home, _ := os.UserHomeDir()
	if home != "" {
		gopathBin := filepath.Join(home, "go", "bin", "templ")
		if info, err := os.Stat(gopathBin); err == nil && !info.IsDir() {
			return gopathBin
		}
	}

	// Tenta via go env GOPATH
	out, err := exec.Command("go", "env", "GOPATH").Output()
	if err == nil {
		gp := strings.TrimSpace(string(out))
		if gp != "" {
			gopathBin := filepath.Join(gp, "bin", "templ")
			if info, err := os.Stat(gopathBin); err == nil && !info.IsDir() {
				return gopathBin
			}
		}
	}

	return ""
}


// HasTempl verifica se o projeto usa templ (tem arquivos .templ)
func HasTempl(root string) bool {
	return exists(root, "app/layouts") || exists(root, "app/pages")
}

func exists(root, dir string) bool {
	_, err := os.Stat(root + "/" + dir)
	return err == nil
}
