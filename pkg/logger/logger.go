// Package logger 提供基于 zerolog 的结构化日志封装
package logger

import (
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
)

// Logger 是 zerolog.Logger 的薄封装,提供 Pulse 统一的日志接口
type Logger = zerolog.Logger

// New 创建一个新的 Logger,输出到 stdout 和日志文件
func New() *Logger {
	// 设置时间格式
	zerolog.TimeFieldFormat = time.RFC3339

	// 创建多输出(stdout + 文件)
	logDir := getLogDir()
	_ = os.MkdirAll(logDir, 0o755)
	file, err := os.OpenFile(
		logDir+"/pulse.log",
		os.O_CREATE|os.O_WRONLY|os.O_APPEND,
		0o644,
	)
	if err != nil {
		// fallback 到 stdout
		return newConsoleLogger(os.Stdout)
	}

	return newConsoleLogger(io.MultiWriter(os.Stdout, file))
}

// newConsoleLogger 创建带颜色的控制台日志
func newConsoleLogger(w io.Writer) *Logger {
	consoleWriter := zerolog.ConsoleWriter{
		Out:        w,
		TimeFormat: time.RFC3339,
	}
	return zerolog.New(consoleWriter).With().Timestamp().Logger()
}

// getLogDir 获取日志目录
func getLogDir() string {
	if dir := os.Getenv("PULSE_LOG_DIR"); dir != "" {
		return dir
	}
	// 默认:~/.pulse/logs
	home, err := os.UserHomeDir()
	if err != nil {
		return "/tmp/pulse/logs"
	}
	return home + "/.pulse/logs"
}