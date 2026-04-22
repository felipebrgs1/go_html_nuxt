package cli

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// scriptBlockRe matches a <script>...</script> block (capturing its content).
var scriptBlockRe = regexp.MustCompile(`(?s)(<script[^>]*>)(.*?)(</script>)`)

// styleBlockRe matches a <style>...</style> block (capturing its content).
var styleBlockRe = regexp.MustCompile(`(?s)(<style[^>]*>)(.*?)(</style>)`)

// RunFmt formats all .gonx files under the project root:
//   - <script> blocks are formatted via gofmt
//   - <style>  blocks are formatted via prettier (if available)
//   - <template> blocks are formatted via prettier (if available)
func RunFmt(projectRoot string) error {
	if projectRoot == "" {
		projectRoot = "."
	}

	abs, err := filepath.Abs(projectRoot)
	if err != nil {
		return fmt.Errorf("invalid path: %w", err)
	}

	appDir := filepath.Join(abs, "app")
	if _, err := os.Stat(appDir); os.IsNotExist(err) {
		return fmt.Errorf("app/ directory not found in %s", abs)
	}

	hasPrettier := prettierAvailable()
	formatted := 0
	skipped := 0

	err = filepath.Walk(appDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(info.Name(), ".gonx") {
			return nil
		}

		rel, _ := filepath.Rel(abs, path)
		changed, fmtErr := formatGonxFile(path, hasPrettier)
		if fmtErr != nil {
			fmt.Printf("  ✗  %s  (%v)\n", rel, fmtErr)
			skipped++
			return nil
		}
		if changed {
			fmt.Printf("  ✔  %s\n", rel)
			formatted++
		}
		return nil
	})
	if err != nil {
		return err
	}

	fmt.Printf("\nDone: %d file(s) formatted, %d skipped.\n", formatted, skipped)
	if !hasPrettier {
		fmt.Println("  Tip: install prettier to also format <template> and <style> blocks.")
	}
	return nil
}

// formatGonxFile reads path, formats each block, and writes back if changed.
func formatGonxFile(path string, hasPrettier bool) (changed bool, err error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return false, err
	}
	original := string(raw)
	result := original

	// --- Format <script> via gofmt ---
	result, err = formatBlock(result, scriptBlockRe, func(content string) (string, error) {
		return runGofmt(content)
	})
	if err != nil {
		return false, fmt.Errorf("gofmt: %w", err)
	}

	// --- Format <style> via prettier (optional) ---
	if hasPrettier {
		result, err = formatBlock(result, styleBlockRe, func(content string) (string, error) {
			return runPrettierOn(content, "css")
		})
		if err != nil {
			// Non-fatal: prettier may not support this input
			err = nil
		}
	}

	if result == original {
		return false, nil
	}

	return true, os.WriteFile(path, []byte(result), 0o644)
}

// formatBlock finds the first match of re in src, applies fn to the inner
// content, and returns the updated string.
// Ensures opening and closing tags stay on their own lines.
func formatBlock(src string, re *regexp.Regexp, fn func(string) (string, error)) (string, error) {
	match := re.FindStringSubmatchIndex(src)
	if match == nil {
		return src, nil
	}
	// Groups: [full, open-tag, content, close-tag]
	_, openEnd := match[2], match[3]
	contentStart, contentEnd := match[4], match[5]
	closeStart := match[6]
	inner := src[contentStart:contentEnd]

	formatted, err := fn(inner)
	if err != nil {
		return src, err
	}

	if strings.TrimSpace(formatted) == "" {
		return src[:openEnd] + "\n" + src[closeStart:], nil
	}

	if !strings.HasPrefix(formatted, "\n") {
		formatted = "\n" + formatted
	}
	if !strings.HasSuffix(formatted, "\n") {
		formatted = formatted + "\n"
	}

	return src[:openEnd] + formatted + src[closeStart:], nil
}

// runGofmt pipes content through `gofmt`.
func runGofmt(script string) (string, error) {
	// gofmt needs a valid Go file.
	// If it doesn't have a package, we inject one.
	trimmed := strings.TrimSpace(script)
	if trimmed == "" {
		return "", nil
	}

	needsPackage := !strings.Contains(trimmed, "package ")
	input := script
	if needsPackage {
		input = "package _gonx_fmt\n" + script
	}

	cmd := exec.Command("gofmt")
	cmd.Stdin = strings.NewReader(input)
	var out, stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return script, fmt.Errorf("%s", strings.TrimSpace(stderr.String()))
	}

	result := out.String()
	if needsPackage {
		// More robust removal: find the first line that starts with 'package ' and remove it
		lines := strings.Split(result, "\n")
		var filtered []string
		found := false
		for _, line := range lines {
			if !found && strings.HasPrefix(strings.TrimSpace(line), "package _gonx_fmt") {
				found = true
				continue
			}
			filtered = append(filtered, line)
		}
		result = strings.Join(filtered, "\n")
	}
	return strings.TrimSpace(result), nil
}

// runPrettierOn formats content using prettier with the given parser.
func runPrettierOn(content, parser string) (string, error) {
	cmd := exec.Command("prettier", "--parser", parser, "--stdin-filepath", "input."+parser)
	cmd.Stdin = strings.NewReader(content)
	var out, stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return content, fmt.Errorf("%s", strings.TrimSpace(stderr.String()))
	}
	return out.String(), nil
}

// prettierAvailable returns true if prettier is on PATH.
func prettierAvailable() bool {
	_, err := exec.LookPath("prettier")
	return err == nil
}
