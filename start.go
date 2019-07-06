package go_wechat

import (
	"github.com/oliverCJ/go-wechat/services"
	"github.com/sirupsen/logrus"
)

var gw = New()

func GetWeChat() *weChat {
	return gw
}

// 设置是否热重启
func SetHoReload(set bool) {
	gw.SetHoReload(set)
}

//TODO 设置是否保存历史记录
func SetCacheHistory(set bool) {
	gw.SetCacheHistory(set)
}

// 设置日志级别
func SetLog(logLevel, logPath string) {
	gw.SetLog(logLevel, logPath)
}

// 设置日志存储根目录
func SetRootPath(dir string) {
	gw.rootPath = dir
}

// 获取读取消息通道操作符
func GetReadChan() <-chan services.Message {
	return gw.readChan
}

// 获取发送通道操作符
func GetSendChan() chan<- services.SendMessage {
	return gw.sendChan
}

// 获取发送消息响应通道操作符
func GetSendRespChan() <-chan services.SendMessageResp {
	return gw.sendChanResp
}

// 获取关闭通道操作符
func GetCloseChan() <-chan bool {
	return gw.closeChan
}

// 获取联系人列表
func GetContact() services.ContactList {
	return gw.userData.ContactList
}

// 获取聊天列表
func GetChatList() []services.User {
	return gw.userData.ChatList
}

// 获取订阅消息
func GetMPSubscribeMsgList() []services.MPSubscribeMsg {
	return gw.userData.MPSubscribeMsgList
}

// 获取全局用户mao
func GetGlobalMemberMap() map[string]services.TinyMemberInfo {
	return gw.userData.GlobalMemberMap
}

// 获取登录用户信息
func GetUserInfo() services.User {
	return gw.userData.UserInfo
}

// 登录
func Login() (*services.LoginService, error) {
	loginService := services.NewLoginService(gw.rootPath)
	err := loginService.Login()
	if err != nil {
		return nil, err
	}

	return loginService, nil
}

// 初始化信息，联系人等
func ContactInit(loginService *services.LoginService) (*services.InitService, error) {
	InitService := services.NewInitService(loginService.LoginData)
	err := InitService.Init()
	if err != nil {
		return nil, err
	}
	return InitService, nil
}

func MsgInit(initService *services.InitService) (*services.MsgServices, error) {
	MsgService := services.NewMsgService(initService, gw.autoReplay, gw.readChan, gw.sendChan, gw.sendChanResp)

	// 子协程检测并获取消息
	go MsgService.SyncDaemon(gw.closeChan)
	// 子协程检测并发送消息
	go MsgService.SendMsgDaemon(gw.closeChan)
	return MsgService, nil
}

func Start() error {
	gw.Init()
	if gw.hotReload {
		// 加载并恢复场景
		loginData, userData, ok, err := services.LoadLogin(gw.rootPath, gw.autoReplay, gw.readChan, gw.sendChan, gw.sendChanResp)
		if err == nil && ok {
			contactService := services.NewInitService(loginData)
			contactService.BaseUserData = userData
			msgService, err := MsgInit(contactService)
			if err != nil {
				return err
			}
			gw.userData = msgService.UserData
			gw.loginData = msgService.LoginData
			return nil
		}
	}
	loginService, err := Login()
	if err != nil {
		return err
	}
	contactService, err := ContactInit(loginService)
	if err != nil {
		return err
	}
	msgService, err := MsgInit(contactService)
	if err != nil {
		return err
	}
	gw.userData = msgService.UserData
	gw.loginData = msgService.LoginData
	return nil
}

func Stop() {
	// 存储数据
	if gw.loginData != nil && gw.userData != nil && gw.hotReload {
		hotReload := services.NewHotReloadService(gw.hotReload, gw.rootPath, gw.loginData, gw.userData)
		err := hotReload.CacheLogin()
		if err != nil {
			logrus.Warningf("保存登录信息失败[err:%s]", err.Error())
		}
	}

	logrus.Infof("收到结束请求, bye")
}
