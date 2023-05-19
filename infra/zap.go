package infra

import (
	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"voice-dispatch/config"
)

var Zaplog *zap.Logger

func ZapInit() {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoder := zapcore.NewJSONEncoder(encoderConfig)
	core := zapcore.NewCore(encoder, getLogWriterByJack(), zapcore.DebugLevel)
	Zaplog = zap.New(core)
}

func getLogWriterByJack() zapcore.WriteSyncer {
	filename := config.AppConfig.FileName
	maxSize := config.AppConfig.MaxSize
	maxBackups := config.AppConfig.MaxBackups
	maxAge := config.AppConfig.MaxAge

	if len(filename) == 0 {
		filename = "./logs/log.log"
	}

	if maxSize == 0 {
		maxSize = 10
	}

	if maxBackups == 0 {
		maxBackups = 5
	}

	if maxAge == 0 {
		maxAge = 30
	}

	logger := &lumberjack.Logger{
		Filename:   filename,
		MaxSize:    maxSize,    // 单位是mb
		MaxBackups: maxBackups, // 备份数量 切割文件之前 会把文件做备份
		MaxAge:     maxAge,     // 备份天数
		Compress:   false,      //默认不压缩
	}

	return zapcore.AddSync(logger)
}

func Info(msg string, fields ...zapcore.Field) {
	Zaplog.Info(msg, fields...)
}

func Error(msg string, fields ...zapcore.Field) {
	Zaplog.Error(msg, fields...)
}
