package logger

import (
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type logger struct {
	*zap.SugaredLogger
}

type Logger interface {
	Debug(args ...interface{})
	Info(args ...interface{})
	Warning(args ...interface{})
	Error(args ...interface{})
	Fatal(args ...interface{})
	Panic(args ...interface{})
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warningf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Fatalf(format string, args ...interface{})
	Panicf(format string, args ...interface{})
	Debugln(args ...interface{})
	Infoln(args ...interface{})
	Warningln(args ...interface{})
	Errorln(args ...interface{})
	Fatalln(args ...interface{})
	Panicln(args ...interface{})
	GetSugar() *zap.SugaredLogger
	Sync()
}

func NewLogger() (Logger, error) {
	config := zap.NewProductionConfig()
	config.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(time.RFC3339) // or time.RFC3339Nano or "2006-01-02 15:04:05"
	zapLogger, err := config.Build()
	if err != nil {
		return nil, err
	}

	sugar := zapLogger.Sugar()

	return &logger{
		sugar,
	}, nil
}

func (l *logger) GetSugar() *zap.SugaredLogger {
	return l.SugaredLogger
}

// func (l *logger) Debug(args ...interface{}) {
// }

// func (l *logger) Info(args ...interface{}) {
// 	l.Info(args)
// }

func (l *logger) Warning(args ...interface{}) {
	l.SugaredLogger.Warn(args...)
}

// func (l *logger) Error(args ...interface{}) {
// 	l.Error(args...)
// }

// func (l *logger) Fatal(args ...interface{}) {
// 	l.Fatal(args...)
// }

// func (l *logger) Panic(args ...interface{}) {
// 	l.Panic(args...)
// }

// func (l *logger) Debugf(format string, args ...interface{}) {
// 	l.Debugf(format, args...)
// }

// func (l *logger) Infof(format string, args ...interface{}) {
// 	l.Infof(format, args...)
// }

func (l *logger) Warningf(format string, args ...interface{}) {
	l.SugaredLogger.Warnf(format, args...)
}

// func (l *logger) Errorf(format string, args ...interface{}) {
// 	l.Errorf(format, args...)
// }

// func (l *logger) Fatalf(format string, args ...interface{}) {
// 	l.Fatalf(format, args...)
// }

// func (l *logger) Panicf(format string, args ...interface{}) {
// 	l.Panicf(format, args...)
// }

// func (l *logger) Debugln(args ...interface{}) {
// 	l.Debugln(args...)
// }

// func (l *logger) Infoln(args ...interface{}) {
// 	l.Info(args...)
// }

func (l *logger) Warningln(args ...interface{}) {
	l.SugaredLogger.Warnln(args...)
}

// func (l *logger) Errorln(args ...interface{}) {
// 	l.Error(args...)
// }

// func (l *logger) Fatalln(args ...interface{}) {
// 	l.Fatal(args...)
// }

// func (l *logger) Panicln(args ...interface{}) {
// 	l.Panic(args...)
// }

func (l *logger) Sync() {
	l.SugaredLogger.Sync()
}
