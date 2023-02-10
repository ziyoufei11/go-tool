package logger

import (
	"time"

	"go.uber.org/zap/zapcore"
)

type LogConfig struct {
	LogFile    string
	LogPath    string `yaml:"logPath"`    //日志文件路径
	MaxSize    int    `yaml:"maxSize"`    //日志文件大小MB
	MaxBackups int    `yaml:"maxBackups"` //日志文件最大备份数量
	MaxAge     int    `yaml:"maxAge"`     //日志最大保存多少天
	Compress   bool   `yaml:"compress"`   //是否启用gzp压缩
	UseCaller  bool   `yaml:"useCaller"`  //是否启用Zap Caller
}

//getEncoder 根据模式获取编码器
func (config *LogConfig) getEncoder(mode string) zapcore.Encoder {
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.FullCallerEncoder,
	}

	if config.UseCaller {
		encoderConfig.CallerKey = "caller"
		encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
	}

	encoderConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.Format("2006-01-02 15:04:05.000"))
	}

	if mode == "file" {
		return zapcore.NewConsoleEncoder(encoderConfig)
	}

	//控制台模式下显示颜色
	encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	return zapcore.NewConsoleEncoder(encoderConfig)
}
