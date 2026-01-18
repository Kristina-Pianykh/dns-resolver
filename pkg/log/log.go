package log

import (
	"log"
	"os"
)

func debugEnabled() bool {
	return os.Getenv("LOG_LEVEL") == "debug"
}

func Debug(msg string, args ...any) {
	if debugEnabled() {
		log.Printf("[DEBUG] "+msg, args...)
	}
}

func Info(msg string, args ...any) {
	log.Printf("[INFO] "+msg, args...)
}

func Error(msg string, args ...any) {
	log.Printf("[ERROR] "+msg, args...)
}

func Warn(msg string, args ...any) {
	log.Printf("[Warn] "+msg, args...)
}
