package go_wechat

import (
	"github.com/oliverCJ/go-wechat/services"
	"github.com/oliverCJ/go-wechat/util"
)

type weChat struct {
	// 设置为true将复用登录信息
	hotReload bool
	// 设置为true将保留历史消息
	cacheHistory bool
	// 日志
	log *util.Log
	// 项目目录
	rootPath string
	// 消息发送通道
	sendChan chan services.SendMessage
	// 消息发送响应
	sendChanResp chan services.SendMessageResp
	// 消息接收通道
	readChan chan services.Message
	// 微信监听意外中断会通过此通道发送信号
	closeChan chan bool
	// 是否自动回复
	autoReplay bool
	// 用户数据
	userData *services.BaseUserData
	// 登录数据
	loginData *services.BaseLoginData
}

func New() *weChat {
	return &weChat{}
}

func (w *weChat) Init() {
	// 初始化日志
	if w.log == nil {
		w.log = new(util.Log)
	}
	w.log.SetDefaults()
	w.log.Init()

	w.readChan = make(chan services.Message, 10)
	w.sendChan = make(chan services.SendMessage, 10)
	w.sendChanResp = make(chan services.SendMessageResp, 10)
	w.closeChan = make(chan bool, 0)
}

func (w *weChat) SetLog(logLevel, logPath string) {
	w.log = new(util.Log)
	w.log.LogPath = logPath
	w.log.Level = logLevel
}

func (w *weChat) SetHoReload(set bool) {
	w.hotReload = set
}

func (w *weChat) SetCacheHistory(set bool) {
	w.cacheHistory = set
}

func (w *weChat) SetRootPath(dir string) {
	w.rootPath = dir
}
