// Package logger 提供基于 zerolog 的结构化日志封装
package logger

import (
	"io"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

// Logger 是 zerolog.Logger 的别名
type Logger = zerolog.Logger

// New 创建一个新的 Logger,输出到 stdout 和日志文件
func New() *Logger {
	zerolog.TimeFieldFormat = time.RFC3339

	logDir := getLogDir()
	_ = os.MkdirAll(logDir, 0o755)
	file, err := os.OpenFile(
		logDir+"/pulse.log",
		os.O_CREATE|os.O_WRONLY|os.O_APPEND,
		0o644,
	)
	if err != nil {
		return newConsoleLogger(os.Stdout)
	}

	return newConsoleLogger(io.MultiWriter(os.Stdout, file))
}

// newConsoleLogger 创建带颜色的控制台日志
func newConsoleLogger(w io.Writer) *Logger {
	consoleWriter := zerolog.ConsoleWriter{
		Out: w,
		FormatTimestamp: func(i interface{}) string {
			return time.Now().Format(time.RFC3339)
		},
		FormatLevel: func(i interface{}) string {
			return "[" + strings.ToUpper(i.(string)) + "]"
		},
	}
	logger := zerolog.New(consoleWriter).With().Timestamp().Logger()
	return &logger
}

// getLogDir 获取日志目录
func getLogDir() string {
	if dir := os.Getenv("PULSE_LOG_DIR"); dir != "" {
		return dir
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "/tmp/pulse/logs"
	}
	return home + "/.pulse/logs"
}