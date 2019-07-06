package global

type WXUrlBase struct {
	//  查询uuid的url
	UUIDUrl string
	// 获取QR的url
	QRUrl string
	// 轮询登录的rul
	LoginUrl string
	// 推送登录URL
	PushLoginUrl string
	// 登出的url
	LogoutUrl string

	// 获取登录公参url
	LoginParamUrl string
	// 获取登录初始化信息path
	LoginInitUrl string
	// 获取联系人列表path
	LoginContactUrl string
	// 批量查询联系人详情path
	LoginContactBatchUrl string
	// 开启状态通知path
	LoginStatusNotifyUrl string
	// 获取微信头像
	WebWXGetHeadImgUrl string

	// 消息检查
	SyncCheckUrl string
	// 获取消息
	WebWXSyncUrl string
	// 发送消息
	WebWXSendMsgUrl string
	// 上传媒体文件
	WebWXUploadMediaUrl string
	// 发送图片消息
	WebWXSendMsgImgUrl string
	// 发送视频消息
	WebWXSendVideoMsgUrl string
}

var (
	HostLogin = "https://login.weixin.qq.com"
	HostWx    = "https://wx2.qq.com"
	HostPush  = "https://webpush.wx2.qq.com"
	HostFile  = "https://file.wx2.qq.com"
)

// 通用全局配置，这里的配置一般情况无需修改
var Common = struct {
	// 链接配置
	WXUrlBase WXUrlBase
	// 定值appid
	APPID string
	// 定值fun
	FUN string
	// 字符集
	Lang string

	WebWx string

	SpecialUsers map[string]string

	UserAgent string
}{

	WXUrlBase: WXUrlBase{
		UUIDUrl:      HostLogin + "/jslogin",                            // jslogin?appid=<appid>&fun=new&lang=zh_CN&_=<_>
		QRUrl:        HostLogin + "/qrcode/%s",                          // qrcode/<uuid>
		LoginUrl:     HostLogin + "/cgi-bin/mmwebwx-bin/login",          // cgi-bin/mmwebwx-bin/login?loginicon=true&uuid=<uuid>&tip=<tip>&r=<r>&_=<_>
		LogoutUrl:    HostWx + "/cgi-bin/mmwebwx-bin/webwxlogout",       // cgi-bin/mmwebwx-bin/webwxlogout?redirect=1&type=<type>&skey=<skey>
		PushLoginUrl: HostWx + "/cgi-bin/mmwebwx-bin/webwxpushloginurl", ///cgi-bin/mmwebwx-bin/webwxpushloginurl?uin=<uin>

		LoginInitUrl:         HostWx + "/cgi-bin/mmwebwx-bin/webwxinit",            // /cgi-bin/mmwebwx-bin/webwxinit?r=<r>&lang=zh_CN&pass_ticket=<pass_ticket>
		LoginStatusNotifyUrl: HostWx + "/cgi-bin/mmwebwx-bin/webwxstatusnotify",    ///cgi-bin/mmwebwx-bin/webwxstatusnotify?lang=zh_CN&pass_ticket=<pass_ticket>
		LoginContactUrl:      HostWx + "/cgi-bin/mmwebwx-bin/webwxgetcontact",      // /cgi-bin/mmwebwx-bin/webwxgetcontact?lang=zh_CN&pass_ticket=<pass_ticket>&r=<r>&seq=0&skey=<skey>
		LoginContactBatchUrl: HostWx + "/cgi-bin/mmwebwx-bin/webwxbatchgetcontact", // /cgi-bin/mmwebwx-bin/webwxbatchgetcontact?type=ex&r=<r>&lang=zh_CN&pass_ticket=<pass_ticket>
		WebWXGetHeadImgUrl:   HostWx + "/cgi-bin/mmwebwx-bin/webwxgetheadimg",      // /cgi-bin/mmwebwx-bin/webwxgetheadimg?seq=<seq>&username=<username>

		SyncCheckUrl:         HostPush + "/cgi-bin/mmwebwx-bin/synccheck",        // /cgi-bin/mmwebwx-bin/synccheck?r=<r>&skey=<skey>&sid=<sid>&uin=<uin>&deviceid=<deviceid>&synckey=<synckey>&_=<_>
		WebWXSyncUrl:         HostWx + "/cgi-bin/mmwebwx-bin/webwxsync",          // /cgi-bin/mmwebwx-bin/webwxsync?sid=<sid>&skey=<skey>&pass_ticket=<pass_ticket>
		WebWXSendMsgUrl:      HostWx + "/cgi-bin/mmwebwx-bin/webwxsendmsg",       ///cgi-bin/mmwebwx-bin/webwxsendmsg?pass_ticket=<pass_ticket>
		WebWXUploadMediaUrl:  HostFile + "/cgi-bin/mmwebwx-bin/webwxuploadmedia", // /cgi-bin/mmwebwx-bin/webwxuploadmedia?f=json
		WebWXSendMsgImgUrl:   HostWx + "/cgi-bin/mmwebwx-bin/webwxsendmsgimg",    // /cgi-bin/mmwebwx-bin/webwxsendmsgimg?fun=async&f=json&lang=zh_CN&pass_ticket=<pass_ticket>
		WebWXSendVideoMsgUrl: HostWx + "/cgi-bin/mmwebwx-bin/webwxsendvideomsg",  // /cgi-bin/mmwebwx-bin/webwxsendvideomsg?fun=async&f=json

	},

	APPID: "wx782c26e4c19acffb",
	FUN:   "fun",
	Lang:  "zh_CN",
	WebWx: "webwx",
	SpecialUsers: map[string]string{
		"newsapp":               "newsapp",
		"fmessage":              "fmessage",
		"filehelper":            "filehelper",
		"weibo":                 "weibo",
		"qqmail":                "qqmail",
		"tmessage":              "tmessage",
		"qmessage":              "qmessage",
		"qqsync":                "qqsync",
		"floatbottle":           "floatbottle",
		"lbsapp":                "lbsapp",
		"shakeapp":              "shakeapp",
		"medianote":             "medianote",
		"qqfriend":              "qqfriend",
		"readerapp":             "readerapp",
		"blogapp":               "blogapp",
		"facebookapp":           "facebookapp",
		"masssendapp":           "masssendapp",
		"meishiapp":             "meishiapp",
		"feedsapp":              "feedsapp",
		"voip":                  "voip",
		"blogappweixin":         "blogappweixin",
		"weixin":                "weixin",
		"brandsessionholder":    "brandsessionholder",
		"weixinreminder":        "weixinreminder",
		"wxid_novlwrv3lqwv11":   "wxid_novlwrv3lqwv11",
		"gh_22b87fa7cb3c":       "gh_22b87fa7cb3c",
		"officialaccounts":      "officialaccounts",
		"notification_messages": "notification_messages",
		"wxitil":                "wxitil",
		"userexperience_alarm":  "userexperience_alarm",
	},

	UserAgent: "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_11_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/54.0.2840.71 Safari/537.36",
}
