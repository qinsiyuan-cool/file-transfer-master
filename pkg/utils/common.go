package utils

import (
	"FileTransfer/config"
	"FileTransfer/pkg/ec"
	"crypto/md5"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"github.com/gin-gonic/gin"
	"math/rand"
	"path"
	"strings"
	"time"
)

type Resp struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func RespFunc(c *gin.Context, httpCode int, code int, data interface{}) {
	c.JSON(httpCode, Resp{
		Code:    code,
		Message: ec.GetMsg(code),
		Data: func(data interface{}) interface{} {
			if data == nil {
				return ""
			}
			return data
		}(data),
	})
}
func GetSHA1Encode(data string) string {
	h := sha1.New()
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

func GetMD5Encode(data string) string {
	h := md5.New()
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

func Get16MD5Encode(data string) string {
	return GetMD5Encode(data)[8:24]
}

func SignWithPassword(name, password string) string {
	return Get16MD5Encode(fmt.Sprintf("alist-%s-%s", password, name))
}

func SignWithToken(name, token string) string {
	return Get16MD5Encode(fmt.Sprintf("alist-%s-%s", token, name))
}

// GetFileType get file type
func GetFileType(ext string) int {
	if ext == "" {
		return config.UNKNOWN
	}
	ext = strings.ToLower(strings.TrimPrefix(ext, "."))
	if IsContain(config.OfficeTypes, ext) {
		return config.OFFICE
	}
	if IsContain(config.AudioTypes, ext) {
		return config.AUDIO
	}
	if IsContain(config.VideoTypes, ext) {
		return config.VIDEO
	}
	if IsContain(config.TextTypes, ext) {
		return config.TEXT
	}
	if IsContain(config.ImageTypes, ext) {
		return config.IMAGE
	}
	return config.UNKNOWN
}
func IsContain(items []string, item string) bool {
	for _, eachItem := range items {
		if eachItem == item {
			return true
		}
	}
	return false
}

func Base(path string) string {
	idx := strings.LastIndex(path, "/")
	if idx == -1 {
		return path
	}
	return path[idx+1:]
}

func Dir(path string) string {
	idx := strings.LastIndex(path, "/")
	if idx == 0 {
		return "/"
	}
	if idx == -1 {
		return path
	}
	return path[:idx]
}
func Ext(name string) string {
	return strings.TrimPrefix(path.Ext(name), ".")
}

func RandPass(lenNum int) string {
	var chars = []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z", "A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M", "N", "O", "P", "Q", "R", "S", "T", "U", "V", "W", "X", "Y", "Z", "1", "2", "3", "4", "5", "6", "7", "8", "9", "0"}
	str := strings.Builder{}
	length := len(chars)
	rand.Seed(time.Now().UnixNano()) //重新播种，否则值不会变
	for i := 0; i < lenNum; i++ {
		str.WriteString(chars[rand.Intn(length)])

	}
	return str.String()
}

// FormatFileSize 字节的单位转换 保留两位小数
func FormatFileSize(fileSize uint64) (size string) {
	if fileSize < 1024 {
		//return strconv.FormatInt(fileSize, 10) + "B"
		return fmt.Sprintf("%.2fB", float64(fileSize)/float64(1))
	} else if fileSize < (1024 * 1024) {
		return fmt.Sprintf("%.2fKB", float64(fileSize)/float64(1024))
	} else if fileSize < (1024 * 1024 * 1024) {
		return fmt.Sprintf("%.2fMB", float64(fileSize)/float64(1024*1024))
	} else if fileSize < (1024 * 1024 * 1024 * 1024) {
		return fmt.Sprintf("%.2fGB", float64(fileSize)/float64(1024*1024*1024))
	} else if fileSize < (1024 * 1024 * 1024 * 1024 * 1024) {
		return fmt.Sprintf("%.2fTB", float64(fileSize)/float64(1024*1024*1024*1024))
	} else { //if fileSize < (1024 * 1024 * 1024 * 1024 * 1024 * 1024)
		return fmt.Sprintf("%.2fEB", float64(fileSize)/float64(1024*1024*1024*1024*1024))
	}
}
