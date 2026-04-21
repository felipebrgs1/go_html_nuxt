package watcher

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
)

type FileWatcher struct {
	watcher  *fsnotify.Watcher
	root     string
	onChange func()
	debounce *time.Timer
}

func NewFileWatcher(root string, onChange func()) (*FileWatcher, error) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	return &FileWatcher{
		watcher:  w,
		root:     root,
		onChange: onChange,
	}, nil
}

func (fw *FileWatcher) Start(ctx context.Context) error {
	// Adiciona diretórios recursivamente
	if err := fw.addDirs(fw.root); err != nil {
		return err
	}

	go fw.loop(ctx)
	return nil
}

func (fw *FileWatcher) addDirs(root string) error {
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			return nil
		}

		// Ignora diretórios comuns que não devem ser observados
		base := filepath.Base(path)
		if base == ".git" || base == "node_modules" || base == "vendor" || base == "tmp" || base == ".framework" {
			return filepath.SkipDir
		}

		return fw.watcher.Add(path)
	})
}

func (fw *FileWatcher) loop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case event, ok := <-fw.watcher.Events:
			if !ok {
				return
			}
			if fw.shouldTrigger(event) {
				fw.trigger()
			}
		case err, ok := <-fw.watcher.Errors:
			if !ok {
				return
			}
			fmt.Printf("Erro no watcher: %v\n", err)
		}
	}
}

func (fw *FileWatcher) shouldTrigger(event fsnotify.Event) bool {
	// Ignora arquivos que não são .go ou .templ
	ext := filepath.Ext(event.Name)
	if ext != ".go" && ext != ".templ" {
		return false
	}

	// Reage a criação, escrita e renomeio
	return event.Op&fsnotify.Write == fsnotify.Write ||
		event.Op&fsnotify.Create == fsnotify.Create ||
		event.Op&fsnotify.Rename == fsnotify.Rename
}

func (fw *FileWatcher) trigger() {
	// Debounce para evitar múltiplos reinícios rápidos
	if fw.debounce != nil {
		fw.debounce.Stop()
	}
	fw.debounce = time.AfterFunc(300*time.Millisecond, func() {
		fmt.Println("🔄 Mudança detectada. Reiniciando servidor...")
		fw.onChange()
	})
}

func (fw *FileWatcher) Close() error {
	return fw.watcher.Close()
}
