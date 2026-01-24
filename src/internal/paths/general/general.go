package general

import (
	"bufio"
	"bytes"
	"path/filepath"
	"runtime/debug"

	"github.com/LuckyGuessServices/planets/internal/luglog"
)

// Abs returns an absolute path to a file or directory.
//
// Panics, if an error occurs.
func Abs(somePath string) string {
	absolutePath, err := filepath.Abs(somePath)
	if err != nil {
		luglog.Panicf("Failed to detect absolute path for: '%s'", somePath)
	}

	return absolutePath
}

// PanicSource parses [debug.Stack] and detects the most possible (closest to "runtime/panic.go") panic path
// and its line number. 'pathPackage' contains "/path/to/package/function".
// 'pathFileWithLine' contains '/absolute/path/to/file:line-number'
func PanicSource() (pathPackage string, pathFileWithLine string) {
	searchSubstringBeforeRequiredLine := []byte("runtime/panic.go:")
	stackScanner := bufio.NewScanner(bytes.NewReader(debug.Stack()))
	defer func() {
		if err := stackScanner.Err(); err != nil {
			luglog.Print("Failed to scan debug.Stack(): ", err)
		}
	}()

	for stackScanner.Scan() {
		if bytes.Contains(stackScanner.Bytes(), searchSubstringBeforeRequiredLine) {
			break
		}
	}

	if !stackScanner.Scan() {
		return
	}
	line := stackScanner.Bytes()
	if len(line) == 0 {
		return
	}
	line, _, _ = bytes.Cut(line, []byte("("))
	pathPackage = string(bytes.TrimSpace(line))

	if !stackScanner.Scan() {
		return
	}
	line = stackScanner.Bytes()
	if len(line) == 0 {
		return
	}

	line, _, _ = bytes.Cut(line, []byte("+0x"))
	pathFileWithLine = string(bytes.TrimSpace(line))

	return
}
