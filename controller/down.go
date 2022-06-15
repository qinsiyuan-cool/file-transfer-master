package controller

import (
	"FileTransfer/models"
	"FileTransfer/pkg/drivers/base"
	"FileTransfer/pkg/ec"
	"FileTransfer/pkg/setting"
	"FileTransfer/pkg/utils"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/unknwon/com"
	"net/http"
)

func Down(c *gin.Context) {
	fileCode := com.StrTo(c.Param("fileCode")).String()
	driver, ok := base.GetDriver(setting.Drivers)
	if !ok {
		utils.RespFunc(c, http.StatusBadRequest, ec.ERROR, fmt.Errorf("no [%s] driver", setting.Drivers))
		return
	}
	if models.ExistFileByPass(fileCode) {
		fileDetail := models.GetFile(fileCode)
		models.UpdateDownload(fileDetail.ID)
		link, err := driver.Link(base.Args{Path: fileDetail.FileId, IP: c.ClientIP()})
		if err != nil {
			utils.RespFunc(c, http.StatusBadRequest, ec.ERROR, err.Error())
		}
		c.Redirect(302, link.Url)
		return
	} else {
		utils.RespFunc(c, http.StatusBadRequest, ec.ERROR_NOT_EXIST_ARTICLE, "")
		return
	}

}
