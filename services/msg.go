package services

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/oliverCJ/go-wechat/constants/errors"
	"github.com/oliverCJ/go-wechat/global"
	"github.com/oliverCJ/go-wechat/util"
	"github.com/sirupsen/logrus"
)

type MsgServices struct {
	LoginData *BaseLoginData
	UserData  *BaseUserData
	Request   *util.Request
	// 检查消息单独实例
	CheckRequest *util.Request

	InitService *InitService

	msgResp *SyncMsgResp
	// 是否自动回复
	autoReply bool
	// 消息读取通道
	MsgRead chan Message
	// 消息发送通道
	MsgSend chan SendMessage
	// 消息发送响应
	MsgSendResp chan SendMessageResp
}

func NewMsgService(initService *InitService, autoReply bool, msgRead chan Message, msgSend chan SendMessage, msgSendResp chan SendMessageResp) *MsgServices {
	// 检查消息需要设置cookie
	u, _ := url.Parse(global.HostWx)
	cookieRequest := util.NewRequest()
	cookieRequest.Client.Jar.SetCookies(u, initService.LoginData.Cookie)

	return &MsgServices{
		LoginData:    initService.LoginData,
		UserData:     initService.BaseUserData,
		Request:      util.NewRequest(),
		CheckRequest: cookieRequest,
		InitService:  initService,
		MsgRead:      msgRead,
		MsgSend:      msgSend,
		MsgSendResp:  msgSendResp,
		msgResp:      &SyncMsgResp{},
		autoReply:    autoReply,
	}
}

// 消息状态检查
func (msg *MsgServices) syncCheck() (selector int, continueCheck bool, err error) {
	params := url.Values{}
	curTime := strconv.FormatInt(time.Now().Unix(), 10)
	params.Set("r", curTime)
	params.Set("sid", msg.LoginData.BaseRequest.Wxsid)
	params.Set("uin", strconv.FormatInt(int64(msg.LoginData.BaseRequest.Wxuin), 10))
	params.Set("skey", msg.LoginData.BaseRequest.Skey)
	params.Set("deviceid", msg.LoginData.BaseRequest.DeviceID)
	params.Set("synckey", msg.UserData.SyncCheckKeyStr)
	params.Set("_", curTime)

	resp, err := msg.CheckRequest.Request(http.MethodGet, global.Common.WXUrlBase.SyncCheckUrl, params, util.JSON_HEADER)
	if err != nil {
		logrus.Warningf("消息检查失败[err:%s]", err.Error())
		return 0, false, errors.MsgError.New().WithMsg("消息检查失败").WithDesc(err.Error())
	}

	matches := regexp.MustCompile(`window.synccheck={retcode:"(\d+)",selector:"(\d+)"}`)
	matchResult := matches.FindStringSubmatch(string(resp))
	if len(matchResult) != 3 {
		logrus.Warningf("消息检查返回数据解析失败[resp:%s]", string(resp))
		return 0, false, errors.MsgError.New().WithMsg("消息检查返回数据解析失败").WithDesc(fmt.Sprintf("resp:%s", string(resp)))
	}

	logrus.Debugf("消息检查返回数据[%s]", string(resp))

	selector = 0
	continueCheck = false

	switch matchResult[1] {
	case "0": // 返回成功
		selector, _ = strconv.Atoi(matchResult[2])
		continueCheck = true
	case "-14": // TICKET错误
		logrus.Warningf("ticket错误")
		err = errors.MsgError.New().WithMsg("ticket错误")
	case "1": // 传入参数错误
		logrus.Warningf("传入参数错误")
		err = errors.MsgError.New().WithMsg("传入参数错误")
	case "1100": // 未登录
		logrus.Warningf("已退出登录")
		err = errors.MsgError.New().WithMsg("已退出登录")
	case "1101": // 在其他设备上登录
		logrus.Warningf("在其他设备上登录")
		err = errors.MsgError.New().WithMsg("在其他设备上登录")
	case "1102": //cookie值无效
		logrus.Warningf("cookie值无效")
		err = errors.MsgError.New().WithMsg("cookie值无效")
	case "1203": // 环境异常
		logrus.Warningf("不安全的登录环境")
		err = errors.MsgError.New().WithMsg("不安全的登录环境")
	case "1205": // 操作频繁
		logrus.Warningf("操作频繁，请稍后再试")
		err = errors.MsgError.New().WithMsg("操作频繁，请稍后再试")
	default:
		logrus.Warningf("其他错误")
		err = errors.MsgError.New().WithMsg("其他错误")
	}
	return
}

func (msg *MsgServices) SyncMsg() error {
	params := url.Values{}
	params.Set("sid", msg.LoginData.BaseRequest.Wxsid)
	params.Set("skey", msg.LoginData.BaseRequest.Skey)
	params.Set("pass_ticket", msg.LoginData.BaseRequest.PassTicket)

	BodyParams, _ := json.Marshal(struct {
		BaseRequest *BaseRequest
		SyncKey     SyncKey
		rr          int64
	}{
		BaseRequest: msg.LoginData.BaseRequest,
		SyncKey:     msg.UserData.SyncKey,
		rr:          ^time.Now().Unix(),
	})

	urlPath := fmt.Sprintf("%s?%s", global.Common.WXUrlBase.WebWXSyncUrl, params.Encode())
	resp, err := msg.Request.Request(http.MethodPost, urlPath, BodyParams, util.JSON_HEADER)
	if err != nil {
		logrus.Warningf("消息拉取失败[err:%s]", err.Error())
		return errors.MsgError.New().WithMsg("消息拉取失败").WithDesc(err.Error())
	}

	logrus.Debugf("获取到消息[%s]", string(resp))

	respData := &SyncMsgResp{}
	err = json.Unmarshal(resp, respData)
	if err != nil {
		logrus.Warningf("消息解析失败[msg:%+s, err:%s]", string(resp), err.Error())
		return errors.MsgError.New().WithMsg("消息解析失败").WithDesc(fmt.Sprintf("[msg:%+s, err:%s]", string(resp), err.Error()))
	}

	if respData.BaseResponse.Ret != 0 {
		logrus.Warningf("获取消息失败,接口请求失败[code:%d,err:%s]", respData.BaseResponse.Ret, respData.BaseResponse.ErrMsg)
		return errors.MsgError.New().WithMsg("获取消息失败").WithDesc(fmt.Sprintf("接口请求失败[code:%d,err:%s]", respData.BaseResponse.Ret, respData.BaseResponse.ErrMsg))
	}

	// 用于同步消息的key
	msg.UserData.SyncKey = respData.SyncKey
	msg.UserData.SyncKeyStr = ""
	for i, item := range respData.SyncKey.List {
		if i == 0 {
			msg.UserData.SyncKeyStr = strconv.Itoa(item.Key) + "_" + strconv.Itoa(item.Val)
			continue
		}
		msg.UserData.SyncKeyStr += "|" + strconv.Itoa(item.Key) + "_" + strconv.Itoa(item.Val)
	}

	// 用于检查的key
	msg.UserData.SyncCheckKey = respData.SyncCheckKey
	msg.UserData.SyncCheckKeyStr = ""
	for i, item := range respData.SyncCheckKey.List {
		if i == 0 {
			msg.UserData.SyncCheckKeyStr = strconv.Itoa(item.Key) + "_" + strconv.Itoa(item.Val)
			continue
		}
		msg.UserData.SyncCheckKeyStr += "|" + strconv.Itoa(item.Key) + "_" + strconv.Itoa(item.Val)
	}

	msg.msgResp = respData

	return nil
}

func (msg *MsgServices) ParseMsg() error {
	if len(msg.msgResp.AddMsgList) > 0 {
		for _, v := range msg.msgResp.AddMsgList {
			logrus.Debugf("收到消息:%+v", v)
			message := Message{}
			msgType := v.(map[string]interface{})["MsgType"].(float64)
			message.MsgType = int(msgType)
			message.FromUserName = v.(map[string]interface{})["FromUserName"].(string)
			message.ToUserName = v.(map[string]interface{})["ToUserName"].(string)
			message.RealUserName = message.FromUserName
			if nickName, ok := msg.UserData.GlobalMemberMap[message.FromUserName]; ok {
				message.FromUserNickName = nickName.NickName
			}
			if nickName, ok := msg.UserData.GlobalMemberMap[message.ToUserName]; ok {
				message.ToUserNickName = nickName.NickName
			}
			message.RealUserNickName = message.FromUserNickName

			message.Content = v.(map[string]interface{})["Content"].(string)
			message.FormatContent = strings.Replace(message.Content, "&lt;", "<", -1)
			message.FormatContent = strings.Replace(message.Content, "&gt;", ">", -1)
			message.FormatContent = strings.Replace(message.Content, " ", " ", 1)

			if message.ToUserName == "filehelper" {
				message.FromUserNickName = msg.UserData.UserInfo.NickName
			}

			switch message.MsgType {
			case 1: // 文本消息
				// 群组消息
				if message.FromUserName[:2] == "@@" {
					groupMemberMatches := regexp.MustCompile(`@(\S+):`)
					matchResult := groupMemberMatches.FindStringSubmatch(message.Content)
					if len(matchResult) == 2 {
						message.RealUserName = "@" + matchResult[1]
						user, _ := msg.InitService.SearchMemberInfo(message.RealUserName, message.FromUserName)
						if user != nil {
							message.RealUserNickName = user.NickName
						}
					}

					contentSlice := strings.Split(message.FormatContent, ":<br/>")
					message.FormatContent = contentSlice[1]
					// 取群名
					if group, ok := msg.UserData.GlobalMemberMap[message.FromUserName]; ok {
						message.FromGroupNickName = group.DisplayName
					}
				} else {
					if msg.autoReply {
						//TODO
					}
				}
				msg.MsgRead <- message
			case 3, 47: // 图片
				message.FormatContent = "[收到图片表情,请在手机上查看]"
				msg.MsgRead <- message
			case 34: // 语音
				message.FormatContent = "[收到语音消息,请在手机上查看]"
				msg.MsgRead <- message
			case 37: // 好友请求
				message.FormatContent = "[收到好友请求,请在手机上查看]"
				msg.MsgRead <- message
			case 42: // 分享名片
			case 43: // 小视频
				message.FormatContent = "[收到视频消息,请在手机上查看]"
				msg.MsgRead <- message
			case 48: // 定位消息
				message.FormatContent = "[收到定位消息,请在手机上查看]"
				msg.MsgRead <- message
			case 49: // 多媒体消息
			case 50:
			case 51: // 状态通知，访问了某一个聊天页面
			case 52:
			case 53:
			case 62: // 短视频
			case 9999: //系统通知
			case 10000: // 系统消息
			case 10002: // 撤回消息
			default: // 未知消息
				msg.MsgRead <- Message{
					FormatContent: fmt.Sprintf("未知消息:%s", v),
				}
			}
		}
	}
	return nil
}

func (msg *MsgServices) sendMsg(message SendMessage) error {
	params := url.Values{}
	params.Set("pass_ticket", msg.LoginData.BaseRequest.PassTicket)

	rand.Seed(time.Now().Unix())
	randMsgId := strconv.FormatInt(time.Now().Unix()<<4, 10) + strconv.Itoa(rand.Int())[0:4]

	reqBodyParam, _ := json.Marshal(struct {
		BaseRequest *BaseRequest
		Msg         struct {
			Type         int
			Content      string
			FromUserName string
			ToUserName   string
			LocalID      string
			ClientMsgId  string
			MediaId      string
		}
		Scene int
	}{
		BaseRequest: msg.LoginData.BaseRequest,
		Msg: struct {
			Type         int
			Content      string
			FromUserName string
			ToUserName   string
			LocalID      string
			ClientMsgId  string
			MediaId      string
		}{
			Type:         1,
			Content:      message.Content,
			FromUserName: msg.UserData.UserInfo.UserName,
			ToUserName:   message.ToUserName,
			LocalID:      message.LocalID,
			ClientMsgId:  randMsgId,
			MediaId:      "",
		},
	})

	urlPath := fmt.Sprintf("%s?%s", global.Common.WXUrlBase.WebWXSendMsgUrl, params.Encode())
	resp, err := msg.Request.Request(http.MethodPost, urlPath, reqBodyParam, util.JSON_HEADER)
	if err != nil {
		logrus.Warningf("消息发送失败[msg:%+v, err:%s]", message, err.Error())
		return errors.MsgError.New().WithMsg("消息发送失败").WithDesc(fmt.Sprintf("[msg:%+v, err:%s]", message, err.Error()))
	}

	respData := SendMessageResp{}
	_ = json.Unmarshal(resp, respData)
	msg.MsgSendResp <- respData

	return nil
}

func (msg *MsgServices) SendMsgDaemon(close chan<- bool) {
	for {
		select {
		case m := <-msg.MsgSend:
			err := msg.sendMsg(m)
			if err != nil {
				logrus.Warningf("消息发送失败[err:%s]", err.Error())
			}
		}
	}
}

func (msg *MsgServices) SyncDaemon(close chan<- bool) {
	for {
		checkTime := time.Now()
		selector, contineCheck, err := msg.syncCheck()
		if err != nil {
			logrus.Warningf("检查消息发生错误[err:%s]", err.Error())
			close <- true
			return
		}
		if !contineCheck {
			close <- true
			return
		}
		switch selector {
		case 2, 3: // 新消息
			err := msg.SyncMsg()
			if err != nil {
				logrus.Warningf("拉取消息发生错误[err:%s]", err.Error())
				close <- true
				return
			}
			err = msg.ParseMsg()
			if err != nil {
				logrus.Warningf("处理消息发生错误[err:%s]", err.Error())
			}
		case 4: // 通讯录更新
			logrus.Infof("通讯录发生变更")
			// 更新通讯录
			_ = msg.InitService.GetContact()
			// TODO
		case 6: //
		case 7: // 进入或离开聊天界面
		case 0: // 无事件
		}
		logrus.Debugf("diff time:%v", time.Now().Sub(checkTime).Seconds())
		if time.Now().Sub(checkTime).Seconds() <= 20 {
			time.Sleep(time.Second * time.Duration(time.Now().Sub(checkTime).Seconds()))
		}
	}
}
