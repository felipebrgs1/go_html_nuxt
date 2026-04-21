package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go_template/pkg/generator"
	"go_template/pkg/router"
	"go_template/pkg/server"
	"go_template/pkg/tailwind"
	"go_template/pkg/templ"
	"go_template/pkg/watcher"
)

const maxConsecutiveErrors = 5

func RunDev() error {
	fmt.Println("🚀 Iniciando servidor de desenvolvimento...")

	if err := compileAssets("."); err != nil {
		fmt.Printf("⚠️  Aviso ao compilar assets: %v\n", err)
	}
	if err := generateRoutes("."); err != nil {
		fmt.Printf("⚠️  Aviso ao gerar rotas: %v\n", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	devServer := server.NewDevServer(".")
	if err := devServer.Start(); err != nil {
		return fmt.Errorf("falha ao iniciar servidor: %w", err)
	}

	errorCount := 0
	shuttingDown := false

	restartFn := func() {
		if shuttingDown {
			return
		}

		if errorCount >= maxConsecutiveErrors {
			fmt.Printf("❌ Máximo de %d erros consecutivos atingido. Pausando reinícios.\n", maxConsecutiveErrors)
			fmt.Println("💡 Corrija o erro e salve o arquivo para tentar novamente.")
			time.Sleep(2 * time.Second)
			return
		}

		if err := compileAssets("."); err != nil {
			fmt.Printf("⚠️  Erro ao compilar assets: %v\n", err)
			errorCount++
			return
		}
		if err := generateRoutes("."); err != nil {
			fmt.Printf("⚠️  Erro ao gerar rotas: %v\n", err)
			errorCount++
			return
		}

		errorCount = 0
		devServer.Restart()
	}

	fw, err := watcher.NewFileWatcher(".", restartFn)
	if err != nil {
		return fmt.Errorf("falha ao iniciar watcher: %w", err)
	}

	if err := fw.Start(ctx); err != nil {
		return fmt.Errorf("falha ao observar arquivos: %w", err)
	}

	fmt.Println("👀 Hot-reload ativo. Pressione Ctrl+C para parar.")

	<-sigCh
	fmt.Println("\n🛑 Parando servidor...")

	shuttingDown = true
	fw.Disable()
	_ = fw.Close()
	cancel()

	// Força kill imediato no grupo de processos
	devServer.Kill()

	// Aguarda brevemente para garantir que a porta foi liberada
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

func compileAssets(root string) error {
	if templ.HasTempl(root) {
		if err := templ.Compile(root); err != nil {
			return fmt.Errorf("templ: %w", err)
		}
	}
	if tailwind.HasTailwind(root) {
		if err := tailwind.Compile(root); err != nil {
			return fmt.Errorf("tailwind: %w", err)
		}
	}
	return nil
}
