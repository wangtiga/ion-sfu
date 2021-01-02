package logger

import (
	"time"

	"github.com/natefinch/lumberjack"
	"github.com/pion/ion-sfu/pkg/common"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	StampMilli = "2006-01-02 15:04:05.000"
)

type LogConfig struct {
	FileName    string
	Console     bool
	Level       string
	MaxFileSize int
	MaxDays     int
	Compress    bool
	Options     []zap.Option
}

type Logger struct {
	log         *zap.SugaredLogger
	atomicLevel *zap.AtomicLevel
	callerSkip  int
}

func New() *Logger {
	return &Logger{
		callerSkip: 1,
	}
}

func getLogLevel(level string) zapcore.Level {
	switch level {
	case "debug":
		return zap.DebugLevel
	case "info":
		return zap.InfoLevel
	case "warn":
		return zap.WarnLevel
	case "error":
		return zap.ErrorLevel
	case "panic":
		return zap.PanicLevel
	case "fatal":
		return zap.FatalLevel
	default:
		return zap.InfoLevel
	}
}

func zapToLoggerString(level zapcore.Level) string {
	switch level {
	case zap.DebugLevel:
		return "debug"
	case zap.InfoLevel:
		return "info"
	case zap.WarnLevel:
		return "warn"
	case zap.ErrorLevel:
		return "error"
	case zap.FatalLevel:
		return "fatal"
	default:
		return "info"
	}
}

func (l *Logger) Init(cfg *LogConfig) {
	hook := lumberjack.Logger{
		Filename:   cfg.FileName,
		MaxSize:    cfg.MaxFileSize,
		MaxBackups: 3,
		MaxAge:     cfg.MaxDays,
		LocalTime:  true,
		Compress:   cfg.Compress,
	}
	w := zapcore.AddSync(&hook)
	zapLevel := getLogLevel("info")
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.Format(StampMilli))
	}
	atomicLevel := zap.NewAtomicLevelAt(zapLevel)
	l.atomicLevel = &atomicLevel
	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderConfig),
		w,
		atomicLevel,
	)
	logger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(l.callerSkip))
	l.log = logger.Sugar()
}

func (l *Logger) SetCallerSkip(skip int) {
	l.callerSkip = skip
}

func (l *Logger) Stop() {
	_ = l.log.Sync()
}

func (l *Logger) Debug(args ...interface{}) {
	l.log.Debug(args...)
}

func (l *Logger) Debugf(format string, args ...interface{}) {
	l.log.Debugf(format, args...)
}

func (l *Logger) Info(args ...interface{}) {
	l.log.Info(args...)
}

func (l *Logger) Infof(format string, args ...interface{}) {
	l.log.Infof(format, args...)
}

func (l *Logger) Warn(args ...interface{}) {
	l.log.Warn(args...)
}

func (l *Logger) Warnf(format string, args ...interface{}) {
	l.log.Warnf(format, args...)
}

func (l *Logger) Error(args ...interface{}) {
	l.log.Error(args...)
}

func (l *Logger) Errorf(format string, args ...interface{}) {
	l.log.Errorf(format, args...)
}

func (l *Logger) Panic(args ...interface{}) {
	l.log.Panic(args...)
}

func (l *Logger) Panicf(format string, args ...interface{}) {
	l.log.Panicf(format, args...)
}

func (l *Logger) Fatal(args ...interface{}) {
	l.log.Panic(args...)
}

func (l *Logger) Fatalf(format string, args ...interface{}) {
	l.log.Panicf(format, args...)
}

func (l *Logger) GetLevel() string {
	if l.atomicLevel != nil {
		return l.atomicLevel.String()
	}
	return ""
}

func (l *Logger) SetLevel(level string) {
	if l.atomicLevel != nil {
		zapLevel := getLogLevel(level)
		l.atomicLevel.SetLevel(zapLevel)
	}
}

func (l *Logger) With(args ...interface{}) common.ILogger {
	nl := *l
	nl.log = l.log.With(args...)
	return &nl
}
