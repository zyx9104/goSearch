package logger

import (
	"bufio"
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
	Level = map[string]logrus.Level{
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
	debug := viper.GetBool("log.debug")
	Log = logrus.New()

	//配置控制台日志
	formatter := new(prefixed.TextFormatter)
	formatter.FullTimestamp = true
	formatter.TimestampFormat = "2006-01-02 15:04:05"

	ff, _ := os.OpenFile(os.DevNull, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	output := bufio.NewWriter(ff)
	if debug {
		Log.SetOutput(os.Stderr)
	} else {
		Log.SetOutput(output)
	}
	Log.SetFormatter(formatter)
	Log.SetLevel(Level[logLevel])

	//配置文件日志
	_, err := os.Stat(logPath)
	if os.IsNotExist(err) {
		err := os.Mkdir(logPath, os.ModePerm)
		if err != nil {
			return errors.New("failed to create log folder")
		}
	}

	cwd, _ := os.Getwd()
	baseLogPath := path.Join(cwd, "logs", logFilename)
	linkName := path.Join(cwd, "logs", logFilename+".log")
	Writer, err = rotateLogs.New(
		baseLogPath+"_%Y%m%d%H%M.log",
		rotateLogs.WithLinkName(linkName),         // 生成软链，指向最新日志文件
		rotateLogs.WithMaxAge(12*30*24*time.Hour), // 文件最大保存时间
		rotateLogs.WithRotationTime(time.Hour),    // 日志切割时间间隔
	)

	if err != nil {
		return err
	}

	fileFormatter := new(prefixed.TextFormatter)
	fileFormatter.FullTimestamp = true
	fileFormatter.TimestampFormat = "2006-01-02 15:04:05"
	fileFormatter.DisableColors = true
	fileFormatter.ForceFormatting = true

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

func Info(args ...interface{}) {
	Log.Info(args...)
}

func Infoln(args ...interface{}) {
	Log.Infoln(args...)
}

func Infof(format string, args ...interface{}) {
	Log.Infof(format, args...)
}

func Debug(args ...interface{}) {
	Log.Debug(args...)
}

func Debugln(args ...interface{}) {
	Log.Debugln(args...)
}

func Debugf(format string, args ...interface{}) {
	Log.Debugf(format, args...)
}

func Warn(args ...interface{}) {
	Log.Warn(args...)
}

func Warnln(args ...interface{}) {
	Log.Warnln(args...)
}

func Warnf(format string, args ...interface{}) {
	Log.Warnf(format, args...)
}

func Fatal(args ...interface{}) {
	Log.Fatal(args...)
}

func Fatalln(args ...interface{}) {
	Log.Fatalln(args...)
}

func Fatalf(format string, args ...interface{}) {
	Log.Fatalf(format, args...)
}

func Panic(args ...interface{}) {
	Log.Panic(args...)
}

func Panicln(args ...interface{}) {
	Log.Panicln(args...)
}

func Panicf(format string, args ...interface{}) {
	Log.Panicf(format, args...)
}
