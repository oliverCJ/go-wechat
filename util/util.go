package util

import (
	"io/ioutil"
	"math/rand"
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

/**
 *  生成随机字符串
 *  index：取随机序列的前index个
 *  0-9:10
 *  0-9a-z:10+24
 *  0-9a-zA-Z:10+24+24
 *  length：需要生成随机字符串的长度
 */
func GetRandomString(index int, length int) string {
	str := "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	bytes := []byte(str)
	result := []byte{}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	for i := 0; i < length; i++ {
		result = append(result, bytes[r.Intn(index)])
	}
	return string(result)
}

func CacheData(data []byte, flag int, file string) (int, error) {
	if flag == 0 {
		flag = os.O_RDWR | os.O_CREATE
	}

	if fileHandle, err := os.OpenFile(file, flag, 0644); err != nil {
		logrus.Warningf("打开文件失败[file:%s, err:%s]", file, err.Error())
		return 0, err
	} else {
		defer fileHandle.Close()

		length, err := fileHandle.Write(data)
		if err != nil {
			logrus.Warningf("写入文件失败[file:%s, err:%s]", file, err.Error())
			return 0, err
		}
		return length, nil
	}
}

func LoadCacheData(file string) ([]byte, error) {
	if fileHandle, err := os.Open(file); err != nil {
		logrus.Warningf("打开文件失败[file:%s, err:%s]", file, err.Error())
		return nil, err
	} else {
		defer fileHandle.Close()

		readContent, err := ioutil.ReadAll(fileHandle)
		if err != nil {
			logrus.Warningf("读取文件失败[file:%s, err:%s]", file, err.Error())
			return nil, err
		}
		return readContent, nil
	}
}
