package util

import (
	"fmt"
	"time"
)


type Logger struct {
	showTimestamp bool
}


func NewLogger(showTimestamp bool) *Logger {
	return &Logger{
		showTimestamp: showTimestamp,
	}
}

func (l *Logger) timestamp() string {
	if l.showTimestamp {
		return time.Now().Format("15:04:05") + " "
	}
	return ""
}

func (l *Logger) log(symbol, message string) {
	fmt.Printf("%s%s %s\n", l.timestamp(), symbol, message)
}


func (l *Logger) Success(message string) {
	l.log("[+]", message)
}


func (l *Logger) Info(message string) {
	l.log("[*]", message)
}


func (l *Logger) Warn(message string) {
	l.log("[!]", message)
}


func (l *Logger) Error(message string) {
	l.log("[-]", message)
}


func (l *Logger) Question(message string) {
	l.log("[?]", message)
}


func (l *Logger) Debug(message string) {
	l.log("[D]", message)
}
