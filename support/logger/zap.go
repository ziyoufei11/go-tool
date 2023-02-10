package logger

import (
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/spf13/viper"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"gopkg.in/natefinch/lumberjack.v2"
)

var ZapLog *zap.Logger
var SugarLog *zap.SugaredLogger

func Init(configPrefix, logFileName string) {

	config := &LogConfig{LogFile: logFileName}
	err := viper.UnmarshalKey(configPrefix, &config)
	if err != nil {
		color.Red(err.Error())
		os.Exit(0)
	}

	if config.LogPath == "" {
		config.LogPath = "."
	}

	if config.LogFile == "" {
		config.LogFile = "error.log"
	}

	isAbsPath := filepath.IsAbs(config.LogPath)
	if !isAbsPath {
		path, err := os.Getwd()
		if err != nil {
			color.Red("获取运行目录失败 %s", err.Error())
			os.Exit(0)
		}

		config.LogPath = filepath.Join(path, config.LogPath)
		err = os.MkdirAll(config.LogPath, 0755)
		if err != nil {
			color.Red("创建日志目录失败 %s", err.Error())
			os.Exit(0)
		}
	}

	hook := lumberjack.Logger{
		Filename:   filepath.Join(config.LogPath, config.LogFile),
		MaxSize:    config.MaxSize,
		MaxBackups: config.MaxBackups,
		MaxAge:     config.MaxAge,
		Compress:   config.Compress,
	}

	stdLog := zapcore.NewCore(config.getEncoder(""), zapcore.AddSync(os.Stdout), zap.DebugLevel)
	fileLog := zapcore.NewCore(config.getEncoder("file"), zapcore.AddSync(&hook), zap.DebugLevel)

	//ZapLog
	ZapLog = zap.New(zapcore.NewTee(stdLog, fileLog), zap.AddCaller(), zap.Development())

	//sugar
	SugarLog = ZapLog.Sugar()
}
