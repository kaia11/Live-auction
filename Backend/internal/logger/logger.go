package logger

import (
	"log"
	"os"
)

var std = log.New(os.Stdout, "", log.Ldate|log.Ltime)

func Info(format string, args ...any) {
	std.Printf("[INFO] "+format, args...)
}

func Error(format string, args ...any) {
	std.Printf("[ERROR] "+format, args...)
}

func Warn(format string, args ...any) {
	std.Printf("[WARN] "+format, args...)
}
