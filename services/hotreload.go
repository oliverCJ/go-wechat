package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/oliverCJ/go-wechat/constants/errors"
	"github.com/oliverCJ/go-wechat/util"
	"github.com/sirupsen/logrus"
)

type HotReloadService struct {
	// 登录数据
	loginData *BaseLoginData
	// 用户数据
	userData *BaseUserData
	// 是否重载登录信息
	HotReload bool
	// 项目路径
	RootDir string
}

type cacheStruct struct {
	LoginData *BaseLoginDataCache `json:"LoginData"`
	UserData  *BaseUserData       `json:"UserData"`
}

type BaseLoginDataCache struct {
	// uuid
	UUID string
	// 登录重定向地址
	LoginRedirectUrl string
	// 公参数据解析
	BaseRequest *BaseRequestCache
	Cookie      []*http.Cookie
}

// 公参用于存储
type BaseRequestCache struct {
	Ret        int    `json:"Ret"`
	Message    string `json:"Message"`
	Skey       string `json:"Skey"`
	Wxsid      string `json:"Wxsid"`
	Wxuin      int64  `json:"Wxuin"`
	PassTicket string `json:"PassTicket"`
	DeviceID   string `json:"DeviceID"`
}

func NewHotReloadService(hotReload bool, rootDir string, loginData *BaseLoginData, userData *BaseUserData) *HotReloadService {
	return &HotReloadService{
		HotReload: hotReload,
		RootDir:   rootDir,
		loginData: loginData,
		userData:  userData,
	}
}

func (h *HotReloadService) CacheLogin() error {
	// cache登录数据h
	if h.HotReload {
		cacheStruct := cacheStruct{
			LoginData: &BaseLoginDataCache{
				UUID:             h.loginData.UUID,
				LoginRedirectUrl: h.loginData.LoginRedirectUrl,
				Cookie:           h.loginData.Cookie,
				BaseRequest: &BaseRequestCache{
					Ret:        h.loginData.BaseRequest.Ret,
					Message:    h.loginData.BaseRequest.Message,
					Skey:       h.loginData.BaseRequest.Skey,
					Wxsid:      h.loginData.BaseRequest.Wxsid,
					Wxuin:      h.loginData.BaseRequest.Wxuin,
					PassTicket: h.loginData.BaseRequest.PassTicket,
					DeviceID:   h.loginData.BaseRequest.DeviceID,
				},
			},
			UserData: h.userData,
		}
		loginDataByte, err := json.Marshal(cacheStruct)
		if err != nil {
			logrus.Warningf("格式化用户信息失败[err:%s]", err.Error())
			// 不中断
			return nil
		}

		_, err = util.CacheData(loginDataByte, os.O_RDWR|os.O_CREATE|os.O_TRUNC, h.RootDir+"/auth.record")
		if err != nil {
			logrus.Warningf("存储用户信息失败[err:%s]", err.Error())
			return nil
		}
	}
	return nil
}

func LoadLogin(rootDir string, autoReply bool, msgRead chan Message, msgSend chan SendMessage, msgSendResp chan SendMessageResp) (*BaseLoginData, *BaseUserData, bool, error) {

	resp, err := util.LoadCacheData(rootDir + "/auth.record")
	if err != nil {
		// 不中断。直接重新登录
		return nil, nil, false, nil
	}

	oldCacheData := new(cacheStruct)

	err = json.Unmarshal(resp, oldCacheData)
	if err != nil {
		logrus.Warningf("解析已保存的登录数据失败[err:%s]", err.Error())
		// 不中断。重新登录
		return nil, nil, false, nil
	}

	loginData := new(BaseLoginData)
	loginData.UUID = oldCacheData.LoginData.UUID
	loginData.BaseRequest = &BaseRequest{
		Ret:        oldCacheData.LoginData.BaseRequest.Ret,
		Message:    oldCacheData.LoginData.BaseRequest.Message,
		Skey:       oldCacheData.LoginData.BaseRequest.Skey,
		Wxsid:      oldCacheData.LoginData.BaseRequest.Wxsid,
		Wxuin:      oldCacheData.LoginData.BaseRequest.Wxuin,
		PassTicket: oldCacheData.LoginData.BaseRequest.PassTicket,
		DeviceID:   oldCacheData.LoginData.BaseRequest.DeviceID,
	}
	loginData.Cookie = oldCacheData.LoginData.Cookie
	loginData.LoginRedirectUrl = oldCacheData.LoginData.LoginRedirectUrl

	// 尝试获取消息
	initService := NewInitService(loginData)
	initService.BaseUserData = oldCacheData.UserData
	msgService := NewMsgService(initService, autoReply, msgRead, msgSend, msgSendResp)
	err = msgService.SyncMsg()
	if err != nil {
		logrus.Warningf("热重启拉取消息发生错误[err:%s]", err.Error())
		return nil, nil, false, errors.HotReloadError.New().WithDesc("热重启拉取消息发生错误").WithDesc(fmt.Sprintf("err:%s", err.Error()))
	}
	err = msgService.ParseMsg()
	if err != nil {
		logrus.Warningf("热重启处理消息发生错误[err:%s]", err.Error())
		return nil, nil, true, errors.HotReloadError.New().WithDesc("热重启处理消息发生错误").WithDesc(fmt.Sprintf("err:%s", err.Error()))
	}

	return loginData, oldCacheData.UserData, true, nil
}
