package watcher

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"

	"github.com/fsnotify/fsnotify"
)

type FileWatcher struct {
	watcher     *fsnotify.Watcher
	root        string
	onChange    func()
	debounce    *time.Timer
	disabled    atomic.Bool
	cooldownEnd atomic.Value // time.Time
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
	if err := fw.addDirs(fw.root); err != nil {
		return err
	}

	go fw.loop(ctx)
	return nil
}

func (fw *FileWatcher) Disable() {
	fw.disabled.Store(true)
}

func (fw *FileWatcher) SetCooldown(d time.Duration) {
	fw.cooldownEnd.Store(time.Now().Add(d))
}

func (fw *FileWatcher) addDirs(root string) error {
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			return nil
		}

		base := filepath.Base(path)
		// Ignora diretórios que causam loops ou não são código-fonte
		if base == ".git" || base == "node_modules" || base == "vendor" ||
			base == "tmp" || base == ".framework" || base == "public" ||
			base == ".gonx" {
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
	if fw.disabled.Load() {
		return false
	}

	if end, ok := fw.cooldownEnd.Load().(time.Time); ok && time.Now().Before(end) {
		return false
	}

	ext := filepath.Ext(event.Name)
	if ext != ".go" && ext != ".templ" && ext != ".gonx" {
		return false
	}

	base := filepath.Base(event.Name)
	if strings.HasSuffix(base, "_templ.go") || strings.HasSuffix(base, "_gonx.go") {
		return false
	}

	return event.Op&fsnotify.Write == fsnotify.Write ||
		event.Op&fsnotify.Create == fsnotify.Create ||
		event.Op&fsnotify.Rename == fsnotify.Rename
}

func (fw *FileWatcher) trigger() {
	if fw.debounce != nil {
		fw.debounce.Stop()
	}
	fw.debounce = time.AfterFunc(400*time.Millisecond, func() {
		if fw.disabled.Load() {
			return
		}
		fmt.Println("🔄 Mudança detectada. Reiniciando servidor...")
		fw.onChange()
	})
}

func (fw *FileWatcher) Close() error {
	fw.disabled.Store(true)
	if fw.debounce != nil {
		fw.debounce.Stop()
	}
	return fw.watcher.Close()
}
