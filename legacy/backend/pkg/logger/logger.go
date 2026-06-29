package logger

import (
	"os"
	"sync"

	"github.com/sirupsen/logrus"
)

var (
	once     sync.Once
	instance *logrus.Logger
)

// L 返回全局 logger 实例
func L() *logrus.Logger {
	once.Do(func() {
		instance = logrus.New()
		instance.SetOutput(os.Stdout)
		instance.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02 15:04:05.000",
			ForceColors:     true,
		})
		instance.SetLevel(logrus.DebugLevel)
	})
	return instance
}

// SetLevel 根据字符串配置日志级别
func SetLevel(level string) {
	l := L()
	switch level {
	case "trace":
		l.SetLevel(logrus.TraceLevel)
	case "debug":
		l.SetLevel(logrus.DebugLevel)
	case "info":
		l.SetLevel(logrus.InfoLevel)
	case "warn", "warning":
		l.SetLevel(logrus.WarnLevel)
	case "error":
		l.SetLevel(logrus.ErrorLevel)
	default:
		l.SetLevel(logrus.InfoLevel)
	}
	l.Infof("Log level set to %s", level)
}
