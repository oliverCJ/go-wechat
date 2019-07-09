//ignore +build
package main

import (
	"fmt"
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

	logOut := make(chan string)
	logFile, _ := os.OpenFile(os.DevNull, os.O_RDWR, 0644)
	defer logFile.Close()
	go_wechat.SetLog("debug", logOut, logFile)

	closeChan := go_wechat.GetCloseChan()
	stopCh := make(chan os.Signal, 1)
	signal.Notify(stopCh, os.Interrupt, syscall.SIGTERM)

	go func() {
		for {
			select {
			case log := <- logOut:
				fmt.Println("hook:" + log)
			case <-closeChan:
				logrus.Infof("意外中断")
				return
			case <-stopCh:
				logrus.Infof("停止接收日志")
				return
			}
		}
	}()

	if err := go_wechat.Start(); err == nil {

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
				logrus.Infof("%s 发送消息:%s", msg.RealUserNickName, msg.FormatContent)
			}
		}
	}
}
