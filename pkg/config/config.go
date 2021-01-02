package config

import (
	"path/filepath"

	"github.com/pion/ion-sfu/pkg/common"
	"github.com/pion/ion-sfu/pkg/logger"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var _realPath string = "/tmp/sfu/log/"
var _logLevel zap.AtomicLevel = zap.NewAtomicLevelAt(zapcore.DebugLevel)
var _defaultLogger *logger.Logger

func init() {
	InitDefaultLogger(false)
}

func InitDefaultLogger(console bool) {
	_defaultLogger = NewFLogger("sfu.log", console)
}

func NewLogger(logTag string) common.ILogger {
	return _defaultLogger.With("LogTag", logTag)
}

func NewFLogger(fileName string, console bool) *logger.Logger {
	dbLogConf := &logger.LogConfig{
		FileName:    filepath.Join(_realPath, fileName), // ⽇志⽂件路径
		Level:       _logLevel.String(),
		MaxFileSize: 500,
		Console:     console,
	}
	logInstance := logger.New()
	logInstance.Init(dbLogConf)
	return logInstance
}
