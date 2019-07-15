package services

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"time"

	"github.com/mdp/qrterminal"
	"github.com/oliverCJ/go-wechat/constants/errors"
	"github.com/oliverCJ/go-wechat/global"
	"github.com/oliverCJ/go-wechat/util"
	"github.com/sirupsen/logrus"
	"github.com/tuotoo/qrcode"
)

// 登录相关

type LoginService struct {
	// 下游可用数据
	LoginData *BaseLoginData
	// 是否已经扫码
	tip string
	// 请求资源
	Request *util.Request
	// 登录二维码路径
	QrImagePath string
	// 等待扫码重试次数
	scanRetryTimes int
}

func NewLoginService(rootDir string) *LoginService {
	return &LoginService{
		Request:     util.NewRequest(),
		QrImagePath: rootDir + "/qrcode.jpg",
		LoginData: &BaseLoginData{
			BaseRequest: new(BaseRequest),
		},
		// 默认为未扫码
		tip: "1",
		// 默认重试次数3次
		scanRetryTimes: 3,
	}
}

func (login *LoginService) Login() error {
	err := login.getUUID()
	if err != nil {
		return err
	}
	err = login.getQRCode()
	if err != nil {
		return err
	}

	defer os.Remove(login.QrImagePath)

	err = login.showTermQrCode()
	if err != nil {
		return err
	}
	//err = login.openQrCode()
	//if err != nil {
	//	return err
	//}

	// 尝试登录
	for ; login.scanRetryTimes > 0; login.scanRetryTimes-- {
		retry, err := login.waitForScan()
		if err != nil {
			return err
		}
		if !retry {
			break
		}
		// 休息2秒
		time.Sleep(2 * time.Second)
	}

	return nil
}

// 获取UUID
func (login *LoginService) getUUID() error {
	params := url.Values{}
	params.Set("appid", global.Common.APPID)
	params.Set("fun", global.Common.FUN)
	params.Set("lang", global.Common.Lang)
	params.Set("_", strconv.FormatInt(time.Now().Unix(), 10))
	resp, err := login.Request.Request(http.MethodGet, global.Common.WXUrlBase.UUIDUrl, params, util.FORM_HEADER)
	if err != nil {
		return err
	}

	matches := regexp.MustCompile(`window.QRLogin.code = (\d+); window.QRLogin.uuid = "(\S+?)";`)
	matchResult := matches.FindStringSubmatch(string(resp))

	if len(matchResult) != 3 {
		logrus.Warningf("获取UUID失败,解析API数据失败,返回数据格式错误[resp：%s]", string(resp))
		return errors.LoginError.New().WithMsg("获取UUID失败").WithDesc(fmt.Sprintf("解析API数据失败，返回数据格式错误[resp：%s]", string(resp)))
	}

	returnCode, err := strconv.ParseInt(matchResult[1], 10, 64)
	if err != nil {
		logrus.Warningf("获取UUID失败，解析API数据失败，获取到错误的code数据[resp：%s，err:%s]", string(resp), err.Error())
		return errors.LoginError.New().WithMsg("获取UUID失败").WithDesc(fmt.Sprintf("解析API数据失败，获取到错误的code数据[resp：%s，err:%s]", string(resp), err.Error()))
	}

	if returnCode != 200 {
		logrus.Warningf("获取UUID失败,API返回错误的状态[resp：%s，code:%d]", string(resp), returnCode)
		return errors.LoginError.New().WithMsg("获取UUID失败").WithDesc(fmt.Sprintf("API返回错误的状态[resp：%s，code:%d]", string(resp), returnCode))
	}
	login.LoginData.UUID = matchResult[2]
	return nil
}

// 获取登录二维码
func (login *LoginService) getQRCode() error {
	if login.LoginData.UUID == "" {
		logrus.Warningf("获取登录二维码失败，没有找到正确的uuid")
		return errors.LoginError.New().WithMsg("获取登录二维码失败").WithDesc("没有找到正确的uuid")
	}

	params := url.Values{}
	params.Set("t", global.Common.WebWx)
	params.Set("_", strconv.FormatInt(time.Now().Unix(), 10))
	resp, err := login.Request.Request(http.MethodPost, fmt.Sprintf(global.Common.WXUrlBase.QRUrl, login.LoginData.UUID), params, util.FORM_HEADER)
	if err != nil {
		logrus.Warningf("获取登录二维码失败[err:%s]", err.Error())
		return errors.LoginError.New().WithMsg("获取登录二维码失败").WithDesc(err.Error())
	}

	dst, err := os.Create(login.QrImagePath)
	if err != nil {
		return err
	}
	defer dst.Close()
	_, err = dst.Write(resp)
	if err != nil {
		logrus.Warningf("创建二维码文件失败[err:%s]", err.Error())
		return errors.LoginError.New().WithMsg("创建二维码文件失败").WithDesc(err.Error())
	}
	return nil
}

// 展示二维码
func (login LoginService) openQrCode() error {
	if login.QrImagePath == "" {
		logrus.Warningf("展示二维码失败，没有找到二维码文件")
		return errors.LoginError.New().WithMsg("展示二维码失败").WithDesc("没有找到二维码文件")
	}
	return exec.Command("open", login.QrImagePath).Start()
}

// 命令行展示二维码
func (login LoginService) showTermQrCode() error {
	fi, err := os.Open(login.QrImagePath)
	if err != nil {
		logrus.Warningf("展示二维码失败，打开二维码文件失败[err:%s]", err.Error())
		return err
	}
	defer fi.Close()
	qrMatrix, err := qrcode.Decode(fi)
	if err != nil {
		logrus.Warningf("展示二维码失败，解析二维码文件失败[err:%s]", err.Error())
		return err
	}
	qrterminal.Generate(qrMatrix.Content, qrterminal.H, os.Stdout)
	return nil
}

// 等待扫描
func (login *LoginService) waitForScan() (retry bool, err error) {
	if login.scanRetryTimes == 0 {
		logrus.Warningf("扫码登录失败,重试次数已达上限，请重新启动程序")
		return false, errors.LoginError.New().WithMsg("扫码登录失败").WithDesc("重试次数已达上限，请重新启动程序")
	}
	params := url.Values{}
	params.Set("tip", login.tip)
	params.Set("uuid", login.LoginData.UUID)
	params.Set("_", strconv.FormatInt(time.Now().Unix(), 10))
	resp, err := login.Request.Request(http.MethodGet, global.Common.WXUrlBase.LoginUrl, params, util.FORM_HEADER)
	if err != nil {
		logrus.Warningf("获取登录信息失败[err:%s]", err.Error())
		return false, errors.LoginError.New().WithMsg("获取登录信息失败").WithDesc(err.Error())
	}

	matches := regexp.MustCompile(`window.code=(\d+);`)
	matchResult := matches.FindStringSubmatch(string(resp))
	if len(matchResult) != 2 {
		logrus.Warningf("解析登录信息code失败[resp:%s]", string(resp))
		return false, errors.LoginError.New().WithMsg("解析登录信息code失败")
	}
	returnCode, err := strconv.ParseInt(matchResult[1], 10, 64)
	if err != nil {
		logrus.Warningf("解析登录信息code失败，获取到错误的code数据[resp：%s，err:%s]", string(resp), err.Error())
		return false, errors.LoginError.New().WithMsg("解析登录信息code失败").WithDesc(fmt.Sprintf("获取到错误的code数据[resp：%s，err:%s]", string(resp), err.Error()))
	}
	switch returnCode {
	case 201: // 已扫码，但是未点击登录
		logrus.Debugf("扫码但是没有点击登录，请重试")
		login.tip = "0"
		return true, nil
	case 200: // 扫码登录成功
		login.tip = "0"
		reRedirectMatches := regexp.MustCompile(`window.redirect_uri="(\S+?)"`)
		reRedirectMatchResult := reRedirectMatches.FindStringSubmatch(string(resp))
		if len(reRedirectMatchResult) != 2 {
			logrus.Warningf("解析登录重定向地址失败[resp:%s]", string(resp))
			return false, errors.LoginError.New().WithMsg("解析登录重定向地址失败")
		}
		login.LoginData.LoginRedirectUrl = reRedirectMatchResult[1] + "&fun=new&version=v2"
		logrus.Debugf("获取登录重定向地址成功:%s", login.LoginData.LoginRedirectUrl)
		return false, nil
	case 400:
		login.tip = "1"
		logrus.Warningf("二维码失效，请重启程序")
		return false, errors.LoginError.New().WithMsg("二维码失效，请重启程序")
	case 408: // 未扫码
		login.tip = "1"
		return true, nil
	case 0: // 扫码超时
		login.tip = "1"
		logrus.Warningf("扫码登录失败，扫码超时，需要重新启动程序")
		return false, errors.LoginError.New().WithMsg("扫码登录失败").WithDesc("扫码超时，请重新启动程序")
	default: // 其他错误
		login.tip = "1"
		logrus.Warningf("扫码登录失败，发生未知错误")
		return false, errors.LoginError.New().WithMsg("扫码登录失败").WithDesc("发生未知错误，请稍后再试")
	}
}
