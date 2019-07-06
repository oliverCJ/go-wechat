package util

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"

	"github.com/oliverCJ/go-wechat/global"

	"github.com/oliverCJ/go-wechat/constants/errors"
	"github.com/sirupsen/logrus"
)

const (
	FORM_HEADER string = "application/x-www-form-urlencoded"
	JSON_HEADER string = "application/json; charset=UTF-8"
)

type Request struct {
	Client *http.Client
}

func NewRequest() *Request {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil
	}

	transport := *(http.DefaultTransport.(*http.Transport))
	transport.ResponseHeaderTimeout = 1 * time.Minute
	transport.TLSClientConfig = &tls.Config{
		InsecureSkipVerify: true,
	}

	return &Request{
		Client: &http.Client{
			Transport: &transport,
			Jar:       jar,
			Timeout:   1 * time.Minute,
		},
	}
}

func (r *Request) Request(method string, requestUrl string, data interface{}, contentType string) (result []byte, err error) {
	var (
		resp = &http.Response{}
		req  = &http.Request{}
	)

	paramsString := ""

	switch data.(type) {
	case url.Values:
		paramsString = data.(url.Values).Encode()
	case string:
		paramsString = data.(string)
	case []byte:
		paramsString = string(data.([]byte))
	}

	logrus.Debugf("向微信API发起请求:[url:%s, method:%s, params:%s]", requestUrl, method, paramsString)

	switch method {
	case http.MethodPost:
		req, err = http.NewRequest(method, requestUrl, strings.NewReader(paramsString))
		if err != nil {
			logrus.Warningf("创建请求失败[err:%s]", err.Error())
			return nil, err
		}

	case http.MethodGet:
		if paramsString != "" {
			requestUrl = fmt.Sprintf("%s?%s", requestUrl, paramsString)
		}
		req, err = http.NewRequest(method, requestUrl, nil)
		if err != nil {
			logrus.Warningf("创建请求失败[err:%s]", err.Error())
			return nil, err
		}
	default:
		return nil, errors.RequestError.New().WithDesc("错误的method")
	}

	req.Header.Set("Content-Type", contentType)
	req.Header.Set("User-Agent", global.Common.UserAgent)

	resp, err = r.Client.Do(req)

	if err != nil {
		logrus.Errorf("请求微信服务器失败:[%s]", err.Error())
		return nil, errors.RequestError.New().WithDesc(fmt.Sprintf("请求微信服务器失败:[%s]", err.Error()))
	}

	defer resp.Body.Close()

	resultBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logrus.Errorf("获取微信API数据失败:[%s]", err.Error())
		return nil, errors.RequestError.New().WithDesc(fmt.Sprintf("获取微信API数据失败:[%s]", err.Error()))
	}

	result = resultBytes

	logrus.Debugf("微信API返回成功，数据长度:%d", len(resultBytes))

	return
}
