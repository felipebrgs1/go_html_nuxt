package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"go_template/pkg/gonx/format"
)

// RunFmt formats all .gonx files under the project root, or a single .gonx
// file if projectRoot points to a file.
func RunFmt(projectRoot string) error {
	if projectRoot == "" {
		projectRoot = "."
	}

	abs, err := filepath.Abs(projectRoot)
	if err != nil {
		return fmt.Errorf("invalid path: %w", err)
	}

	info, err := os.Stat(abs)
	if err != nil {
		return fmt.Errorf("path not found: %w", err)
	}

	// Single-file mode
	if !info.IsDir() {
		if !strings.HasSuffix(abs, ".gonx") {
			return fmt.Errorf("not a .gonx file: %s", abs)
		}
		root := findProjectRoot(abs)
		opts, err := format.LoadConfig(filepath.Join(root, "gonx.toml"))
		if err != nil {
			return fmt.Errorf("failed to load gonx.toml: %w", err)
		}
		rel, _ := filepath.Rel(root, abs)
		changed, fmtErr := format.FormatGonxFile(abs, opts)
		if fmtErr != nil {
			fmt.Printf("  \u2717  %s  (%v)\n", rel, fmtErr)
			return nil
		}
		if changed {
			fmt.Printf("  \u2713  %s\n", rel)
		}
		fmt.Printf("\nDone: 1 file(s) formatted, 0 skipped.\n")
		return nil
	}

	// Directory mode (original behaviour)
	appDir := filepath.Join(abs, "app")
	if _, err := os.Stat(appDir); os.IsNotExist(err) {
		return fmt.Errorf("app/ directory not found in %s", abs)
	}

	opts, err := format.LoadConfig(filepath.Join(abs, "gonx.toml"))
	if err != nil {
		return fmt.Errorf("failed to load gonx.toml: %w", err)
	}

	formatted := 0
	skipped := 0

	err = filepath.Walk(appDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(info.Name(), ".gonx") {
			return nil
		}

		rel, _ := filepath.Rel(abs, path)
		changed, fmtErr := format.FormatGonxFile(path, opts)
		if fmtErr != nil {
			fmt.Printf("  \u2717  %s  (%v)\n", rel, fmtErr)
			skipped++
			return nil
		}
		if changed {
			fmt.Printf("  \u2713  %s\n", rel)
			formatted++
		}
		return nil
	})
	if err != nil {
		return err
	}

	fmt.Printf("\nDone: %d file(s) formatted, %d skipped.\n", formatted, skipped)
	return nil
}

// findProjectRoot walks upward from path looking for a directory that
// contains an "app/" sub-directory.  Falls back to the file's directory.
func findProjectRoot(filePath string) string {
	dir := filepath.Dir(filePath)
	for {
		if _, err := os.Stat(filepath.Join(dir, "app")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return filepath.Dir(filePath)
}
