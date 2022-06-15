package controller

import (
	mw "FileTransfer/middlewares"
	"FileTransfer/models"
	"FileTransfer/pkg/drivers/operate"
	"FileTransfer/pkg/ec"
	"FileTransfer/pkg/setting"
	"FileTransfer/pkg/utils"
	"fmt"
	"github.com/astaxie/beego/validation"
	"github.com/gin-gonic/gin"
	"github.com/unknwon/com"
	"net/http"
)

type RetData struct {
	Path string `json:"path"`
	St   string `json:"st"`
}

func UploadFiles(c *gin.Context) {

	form, err := c.MultipartForm()
	if err != nil {
		utils.RespFunc(c, http.StatusBadRequest, ec.ERROR, err)
	}

	files := form.File["files"]

	var dataMap []map[string]interface{}
	for i, file := range files {
		open, err := file.Open()
		fileSize := uint64(file.Size)
		//1g=1073741824,1M=1048576
		if fileSize > 1073741824 {
			utils.RespFunc(c, http.StatusBadRequest, ec.ERROR, "请上传1GB以内的文件")
			return
		}
		fileStream := mw.FileStream{
			File:       open,
			Size:       fileSize,
			ParentPath: "/", //setting.DriversPath
			Name:       file.Filename,
			MIMEType:   file.Header.Get("Content-Type"),
		}
		dmp, err := operate.Upload(&fileStream)
		if err != nil {
			if i != 0 {
				//删除缓存
				//_ = base.DeleteCache(path_, account)
			}
			utils.RespFunc(c, http.StatusBadRequest, ec.ERROR, fmt.Sprintf("%s", err))
			return
		}
		dmp["size"] = utils.FormatFileSize(fileSize)
		dataMap = append(dataMap, dmp)
	}
	password := utils.RandPass(6)
	var expireTime int
	//存入数据库
	for _, dm := range dataMap {
		data := make(map[string]interface{})
		data["file_name"] = dm["file_name"]
		data["file_id"] = dm["file_id"]
		data["upload_id"] = dm["upload_id"]
		data["password"] = password
		data["uid"] = 0
		data["views"] = 0
		data["downloads"] = 0
		data["file_size"] = dm["size"]
		autoId, expire := models.AddFiles(data)

		if autoId > 0 {
			expireTime = expire
			mw.AddPoolByZSet(mw.QueueData{
				ID:     float64(autoId),
				FileId: dm["file_id"].(string),
				Expire: expire,
			})
		}
	}
	//组装返回数据Data
	var retData = make(map[string]interface{}, 0)

	retData["code"] = password
	retData["expire"] = expireTime
	retData["files"] = dataMap
	retData["url"] = setting.DownDomain + password

	utils.RespFunc(c, http.StatusOK, ec.SUCCESS, retData)
}

func ShowDetailFiles(c *gin.Context) {
	fileCode := com.StrTo(c.PostForm("fileCode")).String()
	valid := validation.Validation{}

	valid.Required(fileCode, "fileCode").Message("取件码不能为空")

	if valid.HasErrors() {
		utils.RespFunc(c, http.StatusBadRequest, 400, valid.Errors)
		return
	}
	isAuth := utils.PostDataCheckAuth(c, []string{"fileCode"})
	if !isAuth {
		utils.RespFunc(c, http.StatusBadRequest, 400, "参数校验失败")
		return
	}

	data := make(map[string]interface{}, 0)
	if models.ExistFileByPass(fileCode) {
		fileDetail := models.GetFile(fileCode)
		models.UpdateViews(fileDetail.ID)
		data["id"] = fileDetail.ID
		data["file_name"] = fileDetail.FileName
		data["password"] = fileDetail.Password
		data["ctime"] = fileDetail.Ctime
		data["views"] = fileDetail.Views
		data["downloads"] = fileDetail.Downloads
		data["expire"] = fileDetail.Expire
		data["file_id"] = fileDetail.FileId
		data["file_size"] = fileDetail.FileSize
		data["url"] = setting.DownDomain + fileDetail.Password

		utils.RespFunc(c, http.StatusOK, ec.SUCCESS, data)
		return
	} else {
		utils.RespFunc(c, http.StatusBadRequest, ec.ERROR_NOT_EXIST_ARTICLE, "")
		return
	}

}
