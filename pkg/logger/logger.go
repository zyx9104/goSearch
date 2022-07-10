package logger

import (
	"errors"
	rotateLogs "github.com/lestrrat/go-file-rotatelogs"
	"github.com/rifflock/lfshook"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"

	"io"
	"os"
	"path"
	"time"
)

var (
	Log   *logrus.Logger
	level = map[string]logrus.Level{
		"":      logrus.DebugLevel,
		"panic": logrus.PanicLevel,
		"fatal": logrus.FatalLevel,
		"error": logrus.ErrorLevel,
		"warn":  logrus.WarnLevel,
		"info":  logrus.InfoLevel,
		"debug": logrus.DebugLevel,
		"trace": logrus.TraceLevel,
	}
	Writer io.Writer
)

func Init() error {
	logPath := viper.GetString("log.path")
	logLevel := viper.GetString("log.level")
	logFilename := viper.GetString("log.filename")
	Log = logrus.New()

	//配置控制台日志
	formatter := new(prefixed.TextFormatter)
	formatter.FullTimestamp = true
	formatter.TimestampFormat = "2006-01-02 15:04:05"

	Log.SetFormatter(formatter)
	Log.SetOutput(os.Stderr)
	Log.SetLevel(level[logLevel])

	//配置文件日志
	_, err := os.Stat(logPath)
	if os.IsNotExist(err) {
		err := os.Mkdir(logPath, os.ModePerm)
		if err != nil {
			return errors.New("failed to create log folder")
		}
	}

	baseLogPath := path.Join(logPath, logFilename)
	Writer, err = rotateLogs.New(
		baseLogPath+"_%Y%m%d%H%M.log",
		rotateLogs.WithLinkName(baseLogPath),      // 生成软链，指向最新日志文件
		rotateLogs.WithMaxAge(7*24*time.Hour),     // 文件最大保存时间
		rotateLogs.WithRotationTime(24*time.Hour), // 日志切割时间间隔
	)

	if err != nil {
		return err
	}

	fileFormatter := new(prefixed.TextFormatter)
	fileFormatter.FullTimestamp = true
	fileFormatter.TimestampFormat = "2006-01-02 15:04:05"
	fileFormatter.DisableColors = true

	lfHook := lfshook.NewHook(lfshook.WriterMap{
		logrus.DebugLevel: Writer, // 为不同级别设置不同的输出目的
		logrus.InfoLevel:  Writer,
		logrus.WarnLevel:  Writer,
		logrus.ErrorLevel: Writer,
		logrus.FatalLevel: Writer,
		logrus.PanicLevel: Writer,
	}, fileFormatter)
	Log.AddHook(lfHook)

	return nil
}

func Infoln(args ...interface{}) {
	Log.Infoln(args...)
}

func Infof(format string, args ...interface{}) {
	Log.Infof(format, args...)
}
