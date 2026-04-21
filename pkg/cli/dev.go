package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"go_template/pkg/generator"
	"go_template/pkg/router"
	"go_template/pkg/server"
	"go_template/pkg/watcher"
)

func RunDev() error {
	fmt.Println("🚀 Iniciando servidor de desenvolvimento...")

	// Gera rotas iniciais
	if err := generateRoutes("."); err != nil {
		fmt.Printf("⚠️  Aviso ao gerar rotas: %v\n", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Canal para graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	// Inicializa o servidor
	devServer := server.NewDevServer(".")
	if err := devServer.Start(); err != nil {
		return fmt.Errorf("falha ao iniciar servidor: %w", err)
	}

	// Callback de restart: regenera rotas e reinicia servidor
	restartFn := func() {
		if err := generateRoutes("."); err != nil {
			fmt.Printf("⚠️  Aviso ao gerar rotas: %v\n", err)
		}
		devServer.Restart()
	}

	// Inicializa o watcher
	fw, err := watcher.NewFileWatcher(".", restartFn)
	if err != nil {
		return fmt.Errorf("falha ao iniciar watcher: %w", err)
	}
	defer fw.Close()

	if err := fw.Start(ctx); err != nil {
		return fmt.Errorf("falha ao observar arquivos: %w", err)
	}

	fmt.Println("👀 Hot-reload ativo. Pressione Ctrl+C para parar.")

	// Aguarda sinal de interrupção
	<-sigCh
	fmt.Println("\n🛑 Parando servidor...")

	if err := devServer.Stop(); err != nil {
		return fmt.Errorf("falha ao parar servidor: %w", err)
	}

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
