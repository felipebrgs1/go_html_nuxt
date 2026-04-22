package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go_template/pkg/generator"
	"go_template/pkg/gonx"
	"go_template/pkg/router"
	"go_template/pkg/server"
	"go_template/pkg/tailwind"
	"go_template/pkg/watcher"
)

const maxConsecutiveErrors = 5

func RunDev(projectRoot string) error {
	if projectRoot == "" {
		projectRoot = "."
	}
	fmt.Printf("Starting development server in %s...\n", projectRoot)

	if err := compileAssets(projectRoot, true, false); err != nil {
		fmt.Printf("Warning during asset compilation: %v\n", err)
	}
	if err := generateRoutes(projectRoot); err != nil {
		fmt.Printf("Warning during route generation: %v\n", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	devServer := server.NewDevServer(projectRoot)
	if err := devServer.Start(); err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	errorCount := 0
	shuttingDown := false

	restartFn := func() {
		if shuttingDown {
			return
		}

		if errorCount >= maxConsecutiveErrors {
			fmt.Printf("Maximum of %d consecutive errors reached. Pausing restarts.\n", maxConsecutiveErrors)
			fmt.Println("Fix the error and save the file to try again.")
			time.Sleep(2 * time.Second)
			return
		}

		start := time.Now()
		if err := compileAssets(projectRoot, false, false); err != nil {
			fmt.Printf("Error during asset compilation: %v\n", err)
			errorCount++
			return
		}
		if err := generateRoutes(projectRoot); err != nil {
			fmt.Printf("Error during route generation: %v\n", err)
			errorCount++
			return
		}

		fmt.Printf("Build finished in %v\n", time.Since(start))
		errorCount = 0
		devServer.Restart()
	}

	fw, err := watcher.NewFileWatcher(projectRoot, restartFn)
	if err != nil {
		return fmt.Errorf("failed to start watcher: %w", err)
	}

	if err := fw.Start(ctx); err != nil {
		return fmt.Errorf("failed to watch files: %w", err)
	}

	fmt.Println("Hot-reload active. Press Ctrl+C to stop.")

	<-sigCh
	fmt.Println("\nStopping server...")

	shuttingDown = true
	fw.Disable()
	_ = fw.Close()
	cancel()

	// Force immediate kill of the process group
	devServer.Kill()

	// Wait briefly to ensure port is released
	time.Sleep(200 * time.Millisecond)

	return nil
}

func generateRoutes(root string) error {
	scanner := router.NewScanner(root)
	routes, err := scanner.Scan()
	if err != nil {
		return err
	}
	return generator.Generate(root, routes)
}

func compileAssets(root string, verbose bool, minify bool) error {
	if gonx.HasGonx(root) {
		if err := gonx.Compile(root, verbose, minify); err != nil {
			return fmt.Errorf("gonx: %w", err)
		}
	}
	if tailwind.HasTailwind(root) {
		if err := tailwind.Compile(root); err != nil {
			return fmt.Errorf("tailwind: %w", err)
		}
	}
	return nil
}

// RunBuild executa a compilação final (com minificação) para deploy
func RunBuild(projectRoot string) error {
	if projectRoot == "" {
		projectRoot = "."
	}
	fmt.Printf("Building for production in %s...\n", projectRoot)

	start := time.Now()
	// No build de deploy, sempre minificamos
	if err := compileAssets(projectRoot, true, true); err != nil {
		return err
	}
	if err := generateRoutes(projectRoot); err != nil {
		return err
	}

	fmt.Printf("Production build finished in %v\n", time.Since(start))
	return nil
}
