package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

// ANSI Color Codes
const (
	Reset  = "\033[0m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Blue   = "\033[34m"
	Cyan   = "\033[36m"
	Gray   = "\033[90m"
)

func Debug(format string, v ...any) {
	if os.Getenv("DEBUG") != "1" {
		return
	}
	output("DEBUG", Cyan, fmt.Sprintf(format, v...))
}

func Info(format string, v ...any) {
	output("INFO", Blue, fmt.Sprintf(format, v...))
}

func Warn(format string, v ...any) {
	output("WARN", Yellow, fmt.Sprintf(format, v...))
}

func Error(format string, v ...any) {
	output("ERROR", Red, fmt.Sprintf(format, v...))
}

func Fatal(format string, v ...any) {
	output("FATAL", Red, fmt.Sprintf(format, v...))
	os.Exit(1)
}

func output(level, color, message string) {
	_, file, line, ok := runtime.Caller(2)
	if !ok {
		file = "unknown"
		line = 0
	} else {
		file = filepath.Base(file)
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05")

	fmt.Printf("%s[%s]%s %s[%s]%s %s[%s:%d]%s %s\n",
		Gray, timestamp, Reset,
		color, level, Reset,
		Gray, file, line, Reset,
		message,
	)
}
