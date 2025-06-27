package common

import (
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/getsentry/sentry-go"
)

// parseGoStackTrace converts Go's debug.Stack() output to Sentry's stacktrace format
func parseGoStackTrace(stack []byte) *sentry.Stacktrace {
	lines := strings.Split(string(stack), "\n")
	var frames []sentry.Frame

	// Skip the first line which is just "goroutine N [running]:"
	for i := 1; i < len(lines)-1; i += 2 {
		// Each frame takes 2 lines:
		// Line 1: function name
		// Line 2: file:line +offset

		if i+1 >= len(lines) {
			break
		}

		functionLine := strings.TrimSpace(lines[i])
		fileLine := strings.TrimSpace(lines[i+1])

		if functionLine == "" || fileLine == "" {
			continue
		}

		frame := parseStackFrame(functionLine, fileLine)
		if frame != nil {
			frames = append(frames, *frame)
		}
	}

	// Reverse frames so the innermost frame is first (Sentry convention)
	for i, j := 0, len(frames)-1; i < j; i, j = i+1, j-1 {
		frames[i], frames[j] = frames[j], frames[i]
	}

	return &sentry.Stacktrace{
		Frames: frames,
	}
}

// parseStackFrame parses a single stack frame from Go's debug output
func parseStackFrame(functionLine, fileLine string) *sentry.Frame {
	// Parse function name
	// Format: "package.function(args...)"
	// or: "package.(*Type).method(args...)"
	funcName := functionLine
	if idx := strings.Index(funcName, "("); idx != -1 {
		funcName = funcName[:idx]
	}

	// Parse file and line
	// Format: "  /path/to/file.go:123 +0x456"
	parts := strings.Fields(fileLine)
	if len(parts) == 0 {
		return nil
	}

	fileAndLine := parts[0]
	lastColonIdx := strings.LastIndex(fileAndLine, ":")
	if lastColonIdx == -1 {
		return nil
	}

	filename := fileAndLine[:lastColonIdx]
	lineStr := fileAndLine[lastColonIdx+1:]

	lineNo, err := strconv.Atoi(lineStr)
	if err != nil {
		return nil
	}

	// Determine if this is application code or external library
	inApp := isApplicationCode(filename)

	// Clean up filename - use basename for readability
	baseName := filepath.Base(filename)

	return &sentry.Frame{
		Filename: baseName,
		AbsPath:  filename,
		Function: funcName,
		Lineno:   lineNo,
		InApp:    inApp,
		Module:   extractPackageName(funcName),
		Package:  extractPackagePath(filename),
		Colno:    0, // Go doesn't provide column numbers in stack traces
	}
}

// isApplicationCode determines if a file path belongs to application code
func isApplicationCode(filename string) bool {
	// Consider it application code if it's in our module
	if strings.Contains(filename, "decorebator.com") {
		return true
	}

	// Exclude Go runtime and standard library
	if strings.Contains(filename, "/usr/local/go/") ||
		strings.Contains(filename, "/go/pkg/") ||
		strings.Contains(filename, "runtime/") ||
		!strings.Contains(filename, "/") {
		return false
	}

	// Exclude common vendor/dependency paths
	if strings.Contains(filename, "/vendor/") ||
		strings.Contains(filename, "/node_modules/") ||
		strings.Contains(filename, "/.go/") {
		return false
	}

	return true
}

// extractPackageName extracts the package name from a function name
func extractPackageName(funcName string) string {
	// For "github.com/pkg/errors.New" -> "errors"
	// For "decorebator.com/internal/service.(*WordService).Create" -> "service"

	parts := strings.Split(funcName, ".")
	if len(parts) >= 2 {
		// Get the last part of the package path
		packagePart := parts[len(parts)-2]
		// Remove any receiver type info like "(*WordService)"
		if strings.Contains(packagePart, ")") {
			if idx := strings.LastIndex(packagePart, ")"); idx != -1 && idx+1 < len(packagePart) {
				return packagePart[idx+1:]
			}
		}
		return packagePart
	}
	return ""
}

// extractPackagePath extracts the package path from a file path
func extractPackagePath(filename string) string {
	// For "/go/src/decorebator.com/internal/service/word.go" -> "decorebator.com/internal/service"

	if strings.Contains(filename, "decorebator.com") {
		// Find the start of our module
		if idx := strings.Index(filename, "decorebator.com"); idx != -1 {
			// Extract from module start to the directory containing the file
			packagePath := filepath.Dir(filename[idx:])
			return packagePath
		}
	}

	return filepath.Dir(filename)
}

// getCurrentCaller returns information about the caller function
// skip=0 means the function calling getCurrentCaller
// skip=1 means the function calling the function that called getCurrentCaller, etc.
func getCurrentCaller(skip int) (file string, line int, function string, ok bool) {
	// Skip additional frames: getCurrentCaller itself + the requested skip
	pc, file, line, ok := runtime.Caller(skip + 1)
	if !ok {
		return
	}

	funcDetails := runtime.FuncForPC(pc)
	if funcDetails != nil {
		function = funcDetails.Name()
	}

	return
}

// createCallerFrame creates a Sentry frame for the current caller
func createCallerFrame(skip int) *sentry.Frame {
	file, line, function, ok := getCurrentCaller(skip + 1) // +1 to skip createCallerFrame itself
	if !ok {
		return nil
	}

	return &sentry.Frame{
		Filename: filepath.Base(file),
		AbsPath:  file,
		Function: function,
		Lineno:   line,
		InApp:    isApplicationCode(file),
		Module:   extractPackageName(function),
		Package:  extractPackagePath(file),
	}
}
