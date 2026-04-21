package server

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"syscall"
	"time"
)

type DevServer struct {
	root    string
	cmd     *exec.Cmd
	mu      sync.Mutex
	port    string
}

func NewDevServer(root string) *DevServer {
	return &DevServer{
		root: root,
		port: "3000",
	}
}

func (s *DevServer) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.startProcess()
}

func (s *DevServer) Restart() {
	s.mu.Lock()
	defer s.mu.Unlock()

	_ = s.stopProcess()
	// Aguarda liberação da porta (TIME_WAIT no Linux)
	time.Sleep(500 * time.Millisecond)
	if err := s.startProcess(); err != nil {
		fmt.Printf("Erro ao reiniciar servidor: %v\n", err)
	}
}

func (s *DevServer) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.stopProcess()
}

func (s *DevServer) startProcess() error {
	entry := s.findEntryPoint()
	if entry == "" {
		fmt.Println("⚠️  Nenhum ponto de entrada encontrado. Aguardando...")
		return nil
	}

	s.cmd = exec.Command("go", "run", entry)
	s.cmd.Dir = s.root
	s.cmd.Stdout = os.Stdout
	s.cmd.Stderr = os.Stderr
	s.cmd.Env = append(os.Environ(), fmt.Sprintf("PORT=%s", s.port))
	// Cria novo grupo de processos para poder matar todos os filhos (incluindo o binário compilado)
	s.cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	fmt.Printf("▶️  Executando: go run %s\n", entry)
	return s.cmd.Start()
}

func (s *DevServer) stopProcess() error {
	if s.cmd == nil || s.cmd.Process == nil {
		return nil
	}

	// Mata o grupo de processos inteiro (go run + binário compilado)
	pgid, err := syscall.Getpgid(s.cmd.Process.Pid)
	if err == nil {
		_ = syscall.Kill(-pgid, syscall.SIGTERM)
	}

	// Aguarda término com timeout
	done := make(chan struct{})
	go func() {
		_ = s.cmd.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-time.After(2 * time.Second):
		// Força kill no grupo
		if pgid > 0 {
			_ = syscall.Kill(-pgid, syscall.SIGKILL)
		}
		return s.cmd.Process.Kill()
	}
}

func (s *DevServer) findEntryPoint() string {
	candidates := []string{
		"app/main.go",
		"cmd/app/main.go",
		"cmd/server/main.go",
		"main.go",
	}

	for _, c := range candidates {
		path := filepath.Join(s.root, c)
		if _, err := os.Stat(path); err == nil {
			return c
		}
	}

	return ""
}
