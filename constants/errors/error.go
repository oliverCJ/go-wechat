package errors

type TypeError int

const (
	LoginError TypeError = iota + 1
	InitLoginError
	RequestError
	MsgError
	HotReloadError
)

func (l TypeError) New() *WeChatError {
	return &WeChatError{
		ErrorType: l.ErrorType(),
		Msg:       l.Msg(),
	}
}

func (l TypeError) Error() string {
	return l.New().Error()
}

func (l TypeError) ErrorType() string {
	switch l {
	case LoginError:
		return "LoginError"
	case InitLoginError:
		return "InitLoginError"
	case RequestError:
		return "RequestError"
	case MsgError:
		return "MsgError"
	case HotReloadError:
		return "HotReloadError"
	}
	return "UNKNOWN"
}

func (l TypeError) Msg() string {
	switch l {
	case LoginError:
		return "登录错误"
	case InitLoginError:
		return "初始化登录信息错误"
	case RequestError:
		return "请求发生错误"
	case MsgError:
		return "获取信息错误"
	case HotReloadError:
		return "热重启失败"
	}
	return "-"
}
