package server

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"syscall"
	"time"
)

type DevServer struct {
	root        string
	cmd         *exec.Cmd
	mu          sync.Mutex
	port        string
	shuttingDown bool
	isRestart   bool
}

func NewDevServer(root string) *DevServer {
	return &DevServer{
		root: root,
		port: findAvailablePort(3000),
	}
}

func findAvailablePort(start int) string {
	for port := start; port < start+100; port++ {
		addr := fmt.Sprintf(":%d", port)
		ln, err := net.Listen("tcp", addr)
		if err == nil {
			_ = ln.Close()
			return fmt.Sprintf("%d", port)
		}
	}
	return "3000" // Fallback
}

func (s *DevServer) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.startProcess()
}

func (s *DevServer) Restart() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.shuttingDown {
		return
	}

	_ = s.stopProcessLocked()
	s.isRestart = true
	time.Sleep(300 * time.Millisecond)
	if err := s.startProcess(); err != nil {
		fmt.Printf("Error restarting server: %v\n", err)
	}
}

func (s *DevServer) Stop() error {
	s.mu.Lock()
	s.shuttingDown = true
	s.mu.Unlock()

	time.Sleep(100 * time.Millisecond)

	s.mu.Lock()
	defer s.mu.Unlock()

	return s.stopProcessLocked()
}

func (s *DevServer) Kill() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.shuttingDown = true
	if s.cmd == nil || s.cmd.Process == nil {
		return
	}

	pgid, err := syscall.Getpgid(s.cmd.Process.Pid)
	if err == nil && pgid > 0 {
		_ = syscall.Kill(-pgid, syscall.SIGKILL)
	}
	_ = s.cmd.Process.Kill()
}

func (s *DevServer) startProcess() error {
	if s.shuttingDown {
		return nil
	}

	entry := s.findEntryPoint()
	if entry == "" {
		fmt.Println("Warning: No entry point found. Waiting...")
		return nil
	}

	s.cmd = exec.Command("go", "run", entry)
	s.cmd.Dir = s.root
	s.cmd.Stdout = os.Stdout
	s.cmd.Stderr = os.Stderr
	s.cmd.Env = append(os.Environ(), fmt.Sprintf("PORT=%s", s.port), "GO_ENV=development")
	if s.isRestart {
		s.cmd.Env = append(s.cmd.Env, "GONX_RESTART=true")
	}
	s.cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	fmt.Printf("Executing: go run %s\n", entry)
	return s.cmd.Start()
}

func (s *DevServer) stopProcessLocked() error {
	if s.cmd == nil || s.cmd.Process == nil {
		return nil
	}

	pgid, err := syscall.Getpgid(s.cmd.Process.Pid)
	if err == nil && pgid > 0 {
		_ = syscall.Kill(-pgid, syscall.SIGTERM)
	} else {
		_ = s.cmd.Process.Signal(syscall.SIGTERM)
	}

	done := make(chan struct{})
	go func() {
		_ = s.cmd.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-time.After(800 * time.Millisecond):
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
