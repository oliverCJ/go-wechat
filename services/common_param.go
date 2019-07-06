package services

import (
	"encoding/xml"
	"net/http"

	"github.com/oliverCJ/go-wechat/constants/types"
)

// 基础登录数据
type BaseLoginData struct {
	// uuid
	UUID string
	// 登录重定向地址
	LoginRedirectUrl string
	// 公参数据解析
	BaseRequest *BaseRequest
	Cookie []*http.Cookie
}

type BaseUserData struct {
	// 登录用户信息
	UserInfo User
	// 联系人列表
	ContactList ContactList
	// 当前聊天列表（可能包含非联系人，比如临时聊天群）
	ChatList []User
	// chatset
	ChatSet    []string

	SyncKey    SyncKey
	SyncKeyStr string
	SyncCheckKey SyncKey
	SyncCheckKeyStr string
	// 订阅信息
	MPSubscribeMsgList []MPSubscribeMsg
	// 所有联系人汇总
	GlobalMemberMap map[string]TinyMemberInfo
}

// 精简联系人信息（主要是为了构建全局MAP，便于查找）
type TinyMemberInfo struct {
	UserName    string
	NickName    string
	DisplayName string
	HeadImgUrl  string
	Sex         int
	Signature   string
	VerifyFlag  int
	Province    string
	City        string
	Type        types.ContactType
}

// 公参
type BaseRequest struct {
	XMLName    xml.Name `xml:"error" json:"-"`
	Ret        int      `xml:"ret" json:"-"`
	Message    string   `xml:"message" json:"-"`
	Skey       string   `xml:"skey" json:"Skey"`
	Wxsid      string   `xml:"wxsid" json:"Sid"`
	Wxuin      int64    `xml:"wxuin" json:"Uin"`
	PassTicket string   `xml:"pass_ticket" json:"-"`
	DeviceID   string   `xml:"-" json:"DeviceID"`
}

// 返回数据基础定义
type BaseResponse struct {
	Ret    int
	ErrMsg string
}

// 当前登录用户定义
type User struct {
	UserName          string `json:"UserName"`
	Uin               int64  `json:"Uin"`
	NickName          string `json:"NickName"`
	HeadImgUrl        string `json:"HeadImgUrl"`
	RemarkName        string `json:"RemarkName"`
	PYInitial         string `json:"PYInitial"`
	PYQuanPin         string `json:"PYQuanPin"`
	RemarkPYInitial   string `json:"RemarkPYInitial"`
	RemarkPYQuanPin   string `json:"RemarkPYQuanPin" `
	HideInputBarFlag  int    `json:"HideInputBarFlag"`
	StarFriend        int    `json:"StarFriend"`
	Sex               int    `json:"Sex"`
	Signature         string `json:"Signature"`
	AppAccountFlag    int    `json:"AppAccountFlag"`
	VerifyFlag        int    `json:"VerifyFlag"`
	ContactFlag       int    `json:"ContactFlag"`
	WebWxPluginSwitch int    `json:"WebWxPluginSwitch"`
	HeadImgFlag       int    `json:"HeadImgFlag"`
	SnsFlag           int    `json:"SnsFlag"`
	Province          string `json:"Province"`
	City              string `json:"City"`
	Alias             string `json:"Alias"`
	DisplayName       string `json:"DisplayName"`
}

// 联系人信息
type Member struct {
	Uin              int64  `json:"Uin"`
	UserName         string `json:"UserName"`
	NickName         string `json:"NickName"`
	HeadImgUrl       string `json:"HeadImgUrl"`
	ContactFlag      int    `json:"ContactFlag"`
	MemberCount      int    `json:"MemberCount"`
	MemberList       []User `json:"MemberList"`
	RemarkName       string `json:"RemarkName"`
	HideInputBarFlag int    `json:"HideInputBarFlag"`
	Sex              int    `json:"Sex"`
	Signature        string `json:"Signature"`
	VerifyFlag       int    `json:"VerifyFlag"`
	OwnerUin         int    `json:"OwnerUin"`
	PYInitial        string `json:"PYInitial"`
	PYQuanPin        string `json:"PYQuanPin"`
	RemarkPYInitial  string `json:"RemarkPYInitial"`
	RemarkPYQuanPin  string `json:"RemarkPYQuanPin"`
	StarFriend       int    `json:"StarFriend"`
	AppAccountFlag   int    `json:"AppAccountFlag"`
	Statues          int    `json:"Statues"`
	AttrStatus       int    `json:"AttrStatus"`
	Province         string `json:"Province"`
	City             string `json:"City"`
	Alias            string `json:"Alias"`
	SnsFlag          int    `json:"SnsFlag"`
	UniFriend        int    `json:"UniFriend"`
	DisplayName      string `json:"DisplayName"`
	ChatRoomId       int    `json:"ChatRoomId"`
	KeyWord          string `json:"KeyWord"`
	EncryChatRoomId  string `json:"EncryChatRoomId"`
	IsOwner          int    `json:"IsOwner"`
}

type SyncKey struct {
	Count int      `json:"Count"`
	List  []KeyVal `json:"List"`
}

type KeyVal struct {
	Key int `json:"Key"`
	Val int `json:"Val"`
}

// 订阅信息
type MPSubscribeMsg struct {
	// 订阅号
	UserName string `json:"UserName"`
	// 文章数量
	MPArticleCount int `json:"MPArticleCount"`
	// 文章列表
	MPArticleList []MPArticle `json:"MPArticleList"`
	// 发布时间
	Time int32 `json:"Time"`
	// 订阅号名称
	NickName string `json:"NickName"`
}

// 订阅文章信息
type MPArticle struct {
	// 标题
	Title string `json:"Title"`
	// 简述
	Digest string `json:"Digest"`
	// 封面
	Cover string `json:"Cover"`
	// 链接地址
	Url string `json:"Url"`
}

type ContactList struct {
	// 联系人列表
	MemberList []Member
	// 分组
	Group []Member
	// 公众号
	PublicUser []Member
	// 特殊
	Special []Member
	// 未知类型
	Unknown []Member
}

// 获取消息原始结构
type SyncMsgResp struct {
	BaseResponse BaseResponse
	// 新消息
	AddMsgList []interface{} `json:"AddMsgList"`
	// 联系人修改
	ModContactList []interface{} `json:"ModContactList"`
	// 联系人删除
	DelContactList []interface{} `json:"DelContactList"`
	ContinueFlag   int           `json:"ContinueFlag"`
	SyncKey        SyncKey       `json:"SyncKey"`
	SyncCheckKey   SyncKey       `json:"SyncCheckKey"`
}


type Message struct {
	MsgId int64
	// 消息发送者
	FromUserName string
	// 消息接收者
	ToUserName string
	// 消息类型
	MsgType       int
	PlayLength    int
	RecommendInfo []string
	// 消息内容
	Content              string
	// 格式化后的消息
	FormatContent		 string
	StatusNotifyUserName string
	StatusNotifyCode     int
	Status               int
	VoiceLength          int
	ForwardFlag          int
	AppMsgType           int
	AppInfo              AppInfo
	Url                  string
	ImgStatus            int
	ImgWidth             int
	ImgHeight            int
	MediaId              string
	FileName             string
	FileSize             string
	// 消息发送者名称
	FromUserNickName     string
	ToUserNickName       string
	// 消息发送的群名
	FromGroupNickName    string
	// 消息创建时间
	CreateTime int32
}

type AppInfo struct {
	Type  int
	AppID string
}

type SendMessage struct {
	ToUserName string
	// 消息内容
	Content string
	// 本地id
	LocalID string
}

type SendMessageResp struct {
	BaseRequest BaseRequest
	MsgID string
	LocalID string
}