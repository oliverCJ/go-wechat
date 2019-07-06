package errors

import "fmt"

type WeChatError struct {
	ErrorType string
	Msg string
	Desc string
}

func (w *WeChatError) WithMsg(msg string) *WeChatError {
	w.Msg = msg
	return w
}

func (w *WeChatError) WithDesc(desc string) *WeChatError {
	w.Desc = desc
	return w
}

func (w WeChatError) Error() string {
	return fmt.Sprintf("[%s]msg:%s,desc:%s", w.ErrorType, w.Msg, w.Desc)
}