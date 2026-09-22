// Package common provides shared logging, error handling, validation and
// git-repo helpers used across all plaesy subcommands. It ports the shared
// API surface of scripts/bash/common.sh and scripts/powershell/common.ps1.
package common

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const (
	colorRed    = "\033[0;31m"
	colorGreen  = "\033[0;32m"
	colorYellow = "\033[1;33m"
	colorBlue   = "\033[0;34m"
	colorGray   = "\033[0;37m"
	colorReset  = "\033[0m"
)

// DebugMode mirrors PLAESY_DEBUG. LogFile mirrors PLAESY_LOG_FILE.
var (
	DebugMode = os.Getenv("PLAESY_DEBUG") == "true"
	LogFile   = os.Getenv("PLAESY_LOG_FILE")
)

func scriptName() string {
	return filepath.Base(os.Args[0])
}

func logMessage(level, message, color string) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	name := scriptName()
	line := fmt.Sprintf("[%s] [%s] [%s] %s", timestamp, level, name, message)

	if level == "WARNING" || level == "ERROR" {
		fmt.Fprintf(os.Stderr, "%s[%s] [%s] [%s]%s %s\n", color, timestamp, level, name, colorReset, message)
	} else {
		fmt.Printf("%s[%s] [%s] [%s]%s %s\n", color, timestamp, level, name, colorReset, message)
	}

	if LogFile != "" {
		_ = os.MkdirAll(filepath.Dir(LogFile), 0o755)
		f, err := os.OpenFile(LogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
		if err == nil {
			defer f.Close()
			fmt.Fprintln(f, line)
		}
	}
}

func LogDebug(format string, args ...any) {
	if DebugMode {
		logMessage("DEBUG", fmt.Sprintf(format, args...), colorGray)
	}
}

func LogInfo(format string, args ...any) {
	logMessage("INFO", fmt.Sprintf(format, args...), colorBlue)
}

func LogSuccess(format string, args ...any) {
	logMessage("SUCCESS", fmt.Sprintf(format, args...), colorGreen)
}

func LogWarning(format string, args ...any) {
	logMessage("WARNING", fmt.Sprintf(format, args...), colorYellow)
}

func LogError(format string, args ...any) {
	logMessage("ERROR", fmt.Sprintf(format, args...), colorRed)
}
