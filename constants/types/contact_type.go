package types

type ContactType int

const (
	CONTACT_TYPE_UNKONWN ContactType = iota + 1 // 未知
	CONTACT_TYPE_TEMP                           // 临时聊天
	CONTACT_TYPE_MEMBER                         // 好友
	CONTACT_TYPE_GROUP                          // 群组
	CONTACT_TYPE_PUBLIC                         // 公众号
	CONTACT_TYPE_SPECIAL                        // 特殊用户，文件助手，官方账号等
)
