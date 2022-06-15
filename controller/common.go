package controller

import (
	"FileTransfer/pkg/drivers/operate"
	"FileTransfer/pkg/utils"
	"github.com/gin-gonic/gin"
	"net/http"
)

func CopyRight(c *gin.Context) {
	personal, _ := operate.PersonalInfo()
	data := map[string]interface{}{
		"web_site": "点点笔记",
		"web_url":  "https://mapi.net.cn/",
		"driver":   "阿里云盘",
		"miit":     "粤ICP备2020114467号",
		"uses":     personal["used_size"],
		"total":    personal["total_size"],
	}
	utils.RespFunc(c, http.StatusOK, 200, data)

}
