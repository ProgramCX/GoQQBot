// Copyright 2026 ProgramCX. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

// Package logger provides a simple logger for GoQQBot.
package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

// LoggerColor logger colors
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorCyan   = "\033[36m"
	colorGray   = "\033[90m"
)

// Level defines log levels
type Level int

const (
	DebugLevel Level = iota
	LevelInfo
	WarnLevel
	ErrorLevel
)

var (
	levelColors = map[Level]string{
		DebugLevel: colorGray,
		LevelInfo:  colorCyan,
		WarnLevel:  colorYellow,
		ErrorLevel: colorRed,
	}

	levelPrefixes = map[Level]string{
		DebugLevel: "[Debug]",
		LevelInfo:  "[Info]",
		WarnLevel:  "[Warn]",
		ErrorLevel: "[Error]",
	}

	std *Logger
)

type Logger struct {
	level  Level
	output io.Writer
	mu     sync.Mutex
}

func init() {
	// output to both terminal and log file
	file, err := os.OpenFile("log_goqqbot.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		std = &Logger{level: DebugLevel, output: os.Stdout}
		return
	}

	multiWriter := io.MultiWriter(os.Stdout, file)

	std = &Logger{
		level:  DebugLevel,
		output: multiWriter,
	}
}

// SetLevel set log level
func (l *Logger) SetLevel(level Level) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

// getCallerInfo get caller info
// skip is the number of stack frames to ascend, with 0 identifying the caller of getCallerInfo.
// It returns caller file name, line number and function name.
func getCallerInfo(skip int) (file string, line int, funcName string) {
	pc, file, line, ok := runtime.Caller(skip + 1)
	if !ok {
		return "???", 0, "???"
	}

	file = filepath.Base(file)

	fn := runtime.FuncForPC(pc)
	if fn != nil {
		// discard package path
		funcName = filepath.Ext(fn.Name())
		if funcName != "" {
			funcName = funcName[1:] // discard leading '.'
		} else {
			funcName = fn.Name()
		}
	}

	return file, line, funcName
}

func (l *Logger) outputF(level Level, format string, v ...interface{}) {
	if level < l.level {
		return
	}

	//get logger caller info
	file, line, funcName := getCallerInfo(2)

	// format time
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	color := levelColors[level]
	prefix := levelPrefixes[level]
	reset := colorReset

	msg := fmt.Sprintf(format, v...)

	//[Info] 2026-02-25 15:41:05 main.go:57:main hello world
	logLine := fmt.Sprintf("%s%s%s %s %s:%d:%s %s%s\n",
		color, prefix, reset,
		timestamp,
		file, line, funcName,
		msg,
		reset,
	)

	_, err := l.output.Write([]byte(logLine))
	if err != nil {
		return
	}
}

// INFO log info message
func INFO(format string, v ...interface{}) {
	std.outputF(LevelInfo, format, v...)
}

// DEBUG log debug message
func DEBUG(format string, v ...interface{}) {
	std.outputF(DebugLevel, format, v...)
}

// WARN log warn message
func WARN(format string, v ...interface{}) {
	std.outputF(WarnLevel, format, v...)
}

// ERROR log error message
func ERROR(format string, v ...interface{}) {
	std.outputF(ErrorLevel, format, v...)
}
