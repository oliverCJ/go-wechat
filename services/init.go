package services

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/oliverCJ/go-wechat/constants/errors"
	"github.com/oliverCJ/go-wechat/constants/types"
	"github.com/oliverCJ/go-wechat/global"
	"github.com/oliverCJ/go-wechat/util"
	"github.com/sirupsen/logrus"
)

// 初始化相关

type InitService struct {
	// 基础登录数据
	LoginData *BaseLoginData
	// 初始化数据
	BaseUserData *BaseUserData
	// 请求资源
	Request *util.Request
}

func NewInitService(data *BaseLoginData) *InitService {
	return &InitService{
		LoginData: data,
		BaseUserData: &BaseUserData{
			GlobalMemberMap: make(map[string]TinyMemberInfo),
		},
		Request: util.NewRequest(),
	}
}

func (init *InitService) Init() error {
	err := init.getLoginPageInfo()
	if err != nil {
		return err
	}
	// 初始化登录数据
	err = init.loginInit()
	if err != nil {
		return err
	}
	err = init.statusNotify()
	if err != nil {
		return err
	}
	err = init.getContact()
	if err != nil {
		return err
	}

	return nil
}

// 获取登录公参
func (init *InitService) getLoginPageInfo() error {
	if init.LoginData.LoginRedirectUrl == "" {
		logrus.Warningf("获取登录公参失败，没有获取到正确的登录跳转地址")
		return errors.InitLoginError.New().WithMsg("获取登录公参失败").WithDesc("没有获取到正确的登录跳转地址")
	}

	resp, err := init.Request.Client.Get(init.LoginData.LoginRedirectUrl)
	if err != nil {
		logrus.Warningf("获取登录公参失败，获取登录公参失败[url:%s,err:%s]", init.LoginData.LoginRedirectUrl, err.Error())
		return errors.InitLoginError.New().WithMsg("获取登录公参失败").WithDesc(fmt.Sprintf("获取登录公参失败[url:%s,err:%s]", init.LoginData.LoginRedirectUrl, err.Error()))
	}

	// 解析公参
	if err = xml.NewDecoder(resp.Body.(io.Reader)).Decode(init.LoginData.BaseRequest); err != nil {
		logrus.Warningf("获取登录公参失败，解析登录公参失败[err:%s]", err.Error())
		return errors.InitLoginError.New().WithMsg("获取登录公参失败").WithDesc(fmt.Sprintf("解析登录公参失败[err:%s]", err.Error()))
	}

	// cookie
	init.LoginData.Cookie = resp.Cookies()

	// 增加DeviceID
	init.LoginData.BaseRequest.DeviceID = "e" + util.GetRandomString(10, 15)

	return nil
}

// 登录初始化
func (init *InitService) loginInit() error {
	params := url.Values{}
	params.Set("pass_ticket", init.LoginData.BaseRequest.PassTicket)
	params.Set("r", strconv.FormatInt(time.Now().Unix(), 10))

	BodyParams, err := json.Marshal(struct {
		BaseRequest *BaseRequest `json:"BaseRequest"`
	}{
		BaseRequest: init.LoginData.BaseRequest,
	})
	if err != nil {
		logrus.Warningf("登录初始化失败，格式化请求参数失败[param:%s,err:%s]", init.LoginData.BaseRequest, err.Error())
		return errors.InitLoginError.New().WithMsg("登录初始化失败").WithDesc(fmt.Sprintf("格式化请求参数失败[param:%s,err:%s]", init.LoginData.BaseRequest, err.Error()))
	}

	urlPath := fmt.Sprintf("%s?%s", global.Common.WXUrlBase.LoginInitUrl, params.Encode())
	resp, err := init.Request.Request(http.MethodPost, urlPath, BodyParams, util.JSON_HEADER)
	if err != nil {
		logrus.Warningf("登录初始化失败[err:%s]", err.Error())
		return errors.InitLoginError.New().WithMsg("登录初始化失败").WithDesc(err.Error())
	}

	type InitResponse struct {
		BaseResponse        *BaseResponse    `json:"BaseResponse"`
		User                User             `json:"User"`
		Count               int              `json:"Count"`
		ContactList         []User           `json:"ContactList"`
		SyncKey             SyncKey          `json:"SyncKey"`
		ChatSet             string           `json:"ChatSet"`
		SKey                string           `json:"SKey"`
		ClientVersion       int              `json:"ClientVersion"`
		SystemTime          int              `json:"SystemTime"`
		GrayScale           int              `json:"GrayScale"`
		InviteStartCount    int              `json:"InviteStartCount"`
		MPSubscribeMsgCount int              `json:"MPSubscribeMsgCount"`
		MPSubscribeMsgList  []MPSubscribeMsg `json:"MPSubscribeMsgList"`
		ClickReportInterval int              `json:"ClickReportInterval"`
	}

	respData := new(InitResponse)

	err = json.Unmarshal(resp, respData)
	if err != nil {
		logrus.Warningf("登录初始化失败,解析返回数据失败[err:%s]", err.Error())
		return errors.InitLoginError.New().WithMsg("登录初始化失败").WithDesc(fmt.Sprintf("解析返回数据失败[err:%s]", err.Error()))
	}
	if respData.BaseResponse.Ret != 0 {
		logrus.Warningf("登录初始化失败,接口请求失败[code:%d,err:%s]", respData.BaseResponse.Ret, respData.BaseResponse.ErrMsg)
		return errors.InitLoginError.New().WithMsg("登录初始化失败").WithDesc(fmt.Sprintf("接口请求失败[code:%d,err:%s]", respData.BaseResponse.Ret, respData.BaseResponse.ErrMsg))
	}

	logrus.Debugf("登录初始化成功[resp:%+v]", respData)

	init.BaseUserData.UserInfo = respData.User

	if len(respData.ContactList) > 0 {
		for _, item := range respData.ContactList {
			temp := TinyMemberInfo{
				UserName:    item.UserName,
				NickName:    item.NickName,
				DisplayName: item.DisplayName,
				HeadImgUrl:  item.HeadImgUrl,
				Sex:         item.Sex,
				Signature:   item.Signature,
				VerifyFlag:  item.VerifyFlag,
				Province:    item.Province,
				City:        item.City,
			}
			if _, ok := global.Common.SpecialUsers[item.UserName]; ok {
				temp.Type = types.CONTACT_TYPE_SPECIAL
			} else if item.UserName[:2] == "@@" { // 群组
				temp.Type = types.CONTACT_TYPE_GROUP
			} else if item.VerifyFlag&8 != 0 {
				temp.Type = types.CONTACT_TYPE_PUBLIC
			} else if item.UserName[:1] == "@" {
				temp.Type = types.CONTACT_TYPE_MEMBER
			}

			init.BaseUserData.GlobalMemberMap[item.UserName] = temp
			init.BaseUserData.ChatList = append(init.BaseUserData.ChatList, item)
		}
	}
	for _, v := range global.Common.SpecialUsers {
		temp := TinyMemberInfo{
			UserName:    v,
			NickName:    v,
			DisplayName: v,
		}
		init.BaseUserData.GlobalMemberMap[v] = temp
	}

	init.BaseUserData.MPSubscribeMsgList = respData.MPSubscribeMsgList
	init.BaseUserData.SyncCheckKey = respData.SyncKey
	if respData.SyncKey.Count > 0 {
		for i, item := range init.BaseUserData.SyncCheckKey.List {
			if i == 0 {
				init.BaseUserData.SyncCheckKeyStr = strconv.Itoa(item.Key) + "_" + strconv.Itoa(item.Val)
				continue
			}
			init.BaseUserData.SyncCheckKeyStr += "|" + strconv.Itoa(item.Key) + "_" + strconv.Itoa(item.Val)
		}
	}
	init.BaseUserData.SyncKey = init.BaseUserData.SyncCheckKey
	init.BaseUserData.SyncKeyStr = init.BaseUserData.SyncCheckKeyStr

	return nil
}

// 开启状态通知
func (init *InitService) statusNotify() error {
	params := url.Values{}
	params.Set("lang", global.Common.Lang)
	params.Set("pass_ticket", init.LoginData.BaseRequest.PassTicket)

	// 开启状态通知请求参数
	type StatusNotifyRequest struct {
		BaseRequest  *BaseRequest `json:"BaseRequest"`
		Code         int          `json:"Code"`
		FromUserName string       `json:"FromUserName"`
		ToUserName   string       `json:"ToUserName"`
		ClientMsgId  int32        `json:"ClientMsgId"`
	}
	// 开启状态通知返回参数
	type NotifyResp struct {
		BaseResponse *BaseResponse `json:"BaseResponse"`
		MsgID        string        `json:"MsgID"`
	}

	BodyParams, _ := json.Marshal(StatusNotifyRequest{
		BaseRequest:  init.LoginData.BaseRequest,
		Code:         3,
		FromUserName: init.BaseUserData.UserInfo.UserName,
		ToUserName:   init.BaseUserData.UserInfo.UserName,
		ClientMsgId:  int32(time.Now().Unix()),
	})

	urlPath := fmt.Sprintf("%s?%s", global.Common.WXUrlBase.LoginStatusNotifyUrl, params.Encode())
	resp, err := init.Request.Request(http.MethodPost, urlPath, BodyParams, util.JSON_HEADER)
	if err != nil {
		logrus.Warningf("开启状态通知失败[err:%s]", err.Error())
		return errors.InitLoginError.New().WithMsg("开启状态通知失败").WithDesc(err.Error())
	}

	respData := new(NotifyResp)
	_ = json.Unmarshal(resp, respData)

	if respData.BaseResponse.Ret != 0 {
		logrus.Warningf("开启状态通知失败,接口请求失败[code:%d,err:%s]", respData.BaseResponse.Ret, respData.BaseResponse.ErrMsg)
		return errors.InitLoginError.New().WithMsg("开启状态通知失败").WithDesc(fmt.Sprintf("接口请求失败[code:%d,err:%s]", respData.BaseResponse.Ret, respData.BaseResponse.ErrMsg))
	}

	return nil
}

// 获取联系人
func (init *InitService) getContact() error {
	params := url.Values{}
	params.Set("pass_ticket", init.LoginData.BaseRequest.PassTicket)
	params.Set("skey", init.LoginData.BaseRequest.Skey)
	params.Set("r", strconv.FormatInt(time.Now().Unix(), 10))

	BodyParams, _ := json.Marshal(struct {
		BaseRequest *BaseRequest `json:"BaseRequest"`
	}{
		BaseRequest: init.LoginData.BaseRequest,
	})

	urlPath := fmt.Sprintf("%s?%s", global.Common.WXUrlBase.LoginContactUrl, params.Encode())
	resp, err := init.Request.Request(http.MethodPost, urlPath, BodyParams, util.JSON_HEADER)
	if err != nil {
		logrus.Warningf("获取联系人失败[err:%s]", err.Error())
		return errors.InitLoginError.New().WithMsg("获取联系人失败").WithDesc(err.Error())
	}

	type MemberResp struct {
		BaseResponse *BaseResponse `json:"BaseResponse"`
		MemberCount  int           `json:"MemberCount"`
		MemberList   []Member      `json:"MemberList"`
		Seq          int           `json:"Seq"`
	}

	respData := new(MemberResp)

	_ = json.Unmarshal(resp, respData)
	if respData.BaseResponse.Ret != 0 {
		logrus.Warningf("获取联系人失败,接口请求失败[code:%d,err:%s]", respData.BaseResponse.Ret, respData.BaseResponse.ErrMsg)
		return errors.InitLoginError.New().WithMsg("获取联系人失败").WithDesc(fmt.Sprintf("接口请求失败[code:%d,err:%s]", respData.BaseResponse.Ret, respData.BaseResponse.ErrMsg))
	}

	// 处理联系人
	if respData.MemberCount > 0 {
		for _, item := range respData.MemberList {
			temp := TinyMemberInfo{
				UserName:    item.UserName,
				NickName:    item.NickName,
				DisplayName: item.DisplayName,
				HeadImgUrl:  item.HeadImgUrl,
				Sex:         item.Sex,
				Signature:   item.Signature,
				VerifyFlag:  item.VerifyFlag,
				Province:    item.Province,
				City:        item.City,
			}

			if item.ContactFlag == 2 { // 群组
				temp.Type = types.CONTACT_TYPE_GROUP
				init.BaseUserData.ContactList.Group = append(init.BaseUserData.ContactList.Group, item)
			} else if item.ContactFlag == 3 { // 公众号
				temp.Type = types.CONTACT_TYPE_PUBLIC
				init.BaseUserData.ContactList.PublicUser = append(init.BaseUserData.ContactList.PublicUser, item)
			} else if item.ContactFlag == 1 { // 联系人
				temp.Type = types.CONTACT_TYPE_MEMBER
				init.BaseUserData.ContactList.MemberList = append(init.BaseUserData.ContactList.MemberList, item)
			} else if _, ok := global.Common.SpecialUsers[item.UserName]; ok {
				temp.Type = types.CONTACT_TYPE_SPECIAL
				init.BaseUserData.ContactList.Special = append(init.BaseUserData.ContactList.Special, item)
			} else {
				temp.Type = types.CONTACT_TYPE_UNKONWN
				init.BaseUserData.ContactList.Unknown = append(init.BaseUserData.ContactList.Unknown, item)
			}

			init.BaseUserData.GlobalMemberMap[item.UserName] = temp
		}
	}

	// 映射chatset
	if len(init.BaseUserData.ChatSet) > 0 {
		init.BaseUserData.ChatList = []User{}
		for _, v := range init.BaseUserData.ChatSet {
			if value, ok := init.BaseUserData.GlobalMemberMap[v]; ok {
				init.BaseUserData.ChatList = append(init.BaseUserData.ChatList, User{
					UserName:    value.UserName,
					NickName:    value.NickName,
					DisplayName: value.DisplayName,
					HeadImgUrl:  value.HeadImgUrl,
					Sex:         value.Sex,
					Signature:   value.Signature,
					VerifyFlag:  value.VerifyFlag,
					Province:    value.Province,
					City:        value.City,
				})
			}
		}
	}
	return nil
}

// TODO 批量获取联系人详情
func (init *InitService) BatchGetContactInfo() {

}
