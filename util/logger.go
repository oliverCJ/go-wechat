package util

import (
	"os"
	"runtime"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
)

type Log struct {
	ReportCaller bool
	Name         string
	Level        string
	Format       string
	init         bool
	LogOutChan   chan string
	LogFile      *os.File
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
	logrus.AddHook(NewWeChatLoggerHook(log.Name, log.LogOutChan))

	if log.LogFile != nil {
		logrus.SetOutput(log.LogFile)
	} else {
		logrus.SetOutput(os.Stdout)
	}

}

type WeChatLoggerHook struct {
	Name    string
	OutChan chan string
}

func NewWeChatLoggerHook(name string, outChan chan string) *WeChatLoggerHook {
	return &WeChatLoggerHook{
		Name:    name,
		OutChan: outChan,
	}
}

func (hook *WeChatLoggerHook) Fire(entry *logrus.Entry) error {
	entry.Data["name"] = hook.Name
	content, err := entry.String()
	if err != nil {
		return err
	}
	if hook.OutChan != nil {
		hook.OutChan <- content
	}

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
