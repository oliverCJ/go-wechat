package util

import (
	"os"
	"runtime"
	"strconv"
	"strings"
	"syscall"

	"github.com/sirupsen/logrus"
)

type Log struct {
	ReportCaller bool
	Name         string
	Level        string
	Format       string
	init         bool
	LogPath      string
}

func (log *Log) SetDefaults() {
	if log.Name == "" {
		log.Name = "go-webchat"
	}

	if log.Level == "" {
		log.Level = "DEBUG"
	}

	if log.Format == "" {
		log.Format = "json"
	}

	log.ReportCaller = true
}

func (log *Log) Init() {
	if !log.init {
		log.Create()
		log.init = true
	}
}

func (log *Log) Create() {
	if log.Format == "json" {
		logrus.SetFormatter(&logrus.JSONFormatter{
			CallerPrettyfier: CallerPrettyfier,
		})
	} else {
		logrus.SetFormatter(&logrus.TextFormatter{
			ForceColors:      true,
			CallerPrettyfier: CallerPrettyfier,
		})
	}

	logrus.SetLevel(getLogLevel(log.Level))
	logrus.SetReportCaller(log.ReportCaller)
	logrus.AddHook(NewWeChatLLoggerHook(log.Name))

	if log.LogPath != "" {
		logrus.SetOutput(os.NewFile(uintptr(syscall.Stdout), log.LogPath))
	} else {
		logrus.SetOutput(os.Stdout)
	}

}

type WeChatLoggerHook struct {
	Name string
}

func NewWeChatLLoggerHook(name string) *WeChatLoggerHook {
	return &WeChatLoggerHook{
		Name: name,
	}
}

func (hook *WeChatLoggerHook) Fire(entry *logrus.Entry) error {
	entry.Data["name"] = hook.Name
	return nil
}

func (hook *WeChatLoggerHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func getLogLevel(l string) logrus.Level {
	level, err := logrus.ParseLevel(strings.ToLower(l))
	if err == nil {
		return level
	}
	return logrus.InfoLevel
}

func CallerPrettyfier(f *runtime.Frame) (function string, file string) {
	return f.Function + " line:" + strconv.FormatInt(int64(f.Line), 10), ""
}
