//ignore +build
package main

import (
	"os"
	"os/signal"
	"syscall"

	go_wechat "github.com/oliverCJ/go-wechat"
	"github.com/sirupsen/logrus"
)

func main() {
	startDaemon()
}

func startDaemon() {
	dir, _ := os.Getwd()
	go_wechat.GetWeChat()
	go_wechat.SetHoReload(true)
	go_wechat.SetRootPath(dir)
	stopCh := make(chan os.Signal, 1)
	signal.Notify(stopCh, os.Interrupt, syscall.SIGTERM)
	if err := go_wechat.Start(); err == nil {
		closeChan := go_wechat.GetCloseChan()
		for {
			select {
			case <-stopCh:
				go_wechat.Stop()
				return
			case <-closeChan:
				logrus.Infof("意外中断")
				return
			case msg := <-go_wechat.GetReadChan():
				logrus.Infof("消息详情:%+v", msg)
				logrus.Infof("%s 发送消息:%s", msg.FromUserNickName, msg.FormatContent)
			}
		}
	}
}
