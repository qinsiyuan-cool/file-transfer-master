package alidrive

import (
	"FileTransfer/middlewares"
	"FileTransfer/pkg/drivers/base"
	"FileTransfer/pkg/setting"
	"errors"
	"fmt"
	"github.com/go-resty/resty/v2"
	jsoniter "github.com/json-iterator/go"
	log "github.com/sirupsen/logrus"
)

var aliClient = resty.New()

func (driver AliDrive) FormatFile(file *AliFile) *middlewares.File {
	f := &middlewares.File{
		Id:        file.FileId,
		Name:      file.Name,
		Size:      file.Size,
		UpdatedAt: file.UpdatedAt,
		Thumbnail: file.Thumbnail,
		Driver:    driver.Config().Name,
		Url:       file.Url,
	}
	f.Type = file.GetType()
	return f
}

func (driver AliDrive) GetFiles(fileId string) ([]AliFile, error) {
	marker := "first"
	res := make([]AliFile, 0)
	for marker != "" {
		if marker == "first" {
			marker = ""
		}
		var resp AliFiles
		var e AliRespError
		_, err := aliClient.R().
			SetResult(&resp).
			SetError(&e).
			SetHeader("authorization", "Bearer\t"+middlewares.GetAccessToken()).
			SetBody(base.Json{
				"drive_id":                setting.DriveId,
				"fields":                  "*",
				"image_thumbnail_process": "image/resize,w_400/format,jpeg",
				"image_url_process":       "image/resize,w_1920/format,jpeg",
				"limit":                   200,
				"marker":                  marker,
				"order_by":                "updated_at",
				"order_direction":         "DESC",
				"parent_file_id":          fileId,
				"video_thumbnail_process": "video/snapshot,t_0,f_jpg,ar_auto,w_300",
				"url_expire_sec":          14400,
			}).Post("https://api.aliyundrive.com/v2/file/list")
		if err != nil {
			return nil, err
		}
		if e.Code != "" {
			if e.Code == "AccessTokenInvalid" {
				err = driver.RefreshToken()
				if err != nil {
					return nil, err
				} else {
					return driver.GetFiles(fileId)
				}
			}
			return nil, fmt.Errorf("%s", e.Message)
		}
		marker = resp.NextMarker
		res = append(res, resp.Items...)
	}
	return res, nil
}

func (driver AliDrive) GetFile(path string) (*AliFile, error) {
	//dir, name := filepath.Split(path)
	//dir = utils.ParsePath(dir)
	//_, err := driver.Files(dir)
	//if err != nil {
	//	return nil, err
	//}
	//parentFiles_, _ := base.GetCache(dir, account)
	//parentFiles, _ := parentFiles_.([]AliFile)
	//for _, file := range parentFiles {
	//	if file.Name == name {
	//		if file.Type == "file" {
	//			return &file, err
	//		} else {
	//			return nil, fmt.Errorf("not file")
	//		}
	//	}
	//}
	return nil, base.ErrPathNotFound
}

func (driver AliDrive) RefreshToken() error {
	url := "https://auth.aliyundrive.com/v2/account/token"
	var resp base.TokenResp
	var e AliRespError
	_, err := aliClient.R().
		//ForceContentType("application/json").
		SetBody(base.Json{"refresh_token": middlewares.GetRefreshToken(), "grant_type": "refresh_token"}).
		SetResult(&resp).
		SetError(&e).
		Post(url)
	if err != nil {
		return err
	}
	log.Debugf("%+v,%+v", resp, e)
	if e.Code != "" {
		return fmt.Errorf("failed to refresh token: %s", e.Message)
	} else {
	}
	middlewares.SetRefreshToken(resp.RefreshToken)
	middlewares.SetAccessToken(resp.AccessToken)
	return nil
}

func (driver AliDrive) rename(fileId, name string) error {
	var resp base.Json
	var e AliRespError
	_, err := aliClient.R().SetResult(&resp).SetError(&e).
		SetHeader("authorization", "Bearer\t"+middlewares.GetAccessToken()).
		SetBody(base.Json{
			"check_name_mode": "refuse",
			"drive_id":        setting.DriveId,
			"file_id":         fileId,
			"name":            name,
		}).Post("https://api.aliyundrive.com/v3/file/update")
	if err != nil {
		return err
	}
	if e.Code != "" {
		if e.Code == "AccessTokenInvalid" {
			err = driver.RefreshToken()
			if err != nil {
				return err
			} else {
				return driver.rename(fileId, name)
			}
		}
		return fmt.Errorf("%s", e.Message)
	}
	if resp["name"] == name {
		return nil
	}
	return fmt.Errorf("%+v", resp)
}

func (driver AliDrive) batch(srcId, dstId string, url string) error {
	var e AliRespError
	res, err := aliClient.R().SetError(&e).
		SetHeader("authorization", "Bearer\t"+middlewares.GetAccessToken()).
		SetBody(base.Json{
			"requests": []base.Json{
				{
					"headers": base.Json{
						"Content-Type": "application/json",
					},
					"method": "POST",
					"id":     srcId,
					"body": base.Json{
						"drive_id":          setting.DriveId,
						"file_id":           srcId,
						"to_drive_id":       setting.DriveId,
						"to_parent_file_id": dstId,
					},
					"url": url,
				},
			},
			"resource": "file",
		}).Post("https://api.aliyundrive.com/v3/batch")
	if err != nil {
		return err
	}
	if e.Code != "" {
		if e.Code == "AccessTokenInvalid" {
			err = driver.RefreshToken()
			if err != nil {
				return err
			} else {
				//_ = model.SaveAccount(account)
				return driver.batch(srcId, dstId, url)
			}
		}
		return fmt.Errorf("%s", e.Message)
	}
	status := jsoniter.Get(res.Body(), "responses", 0, "status").ToInt()
	if status < 400 && status >= 100 {
		return nil
	}
	return errors.New(res.String())
}

func init() {
	base.RegisterDriver(&AliDrive{})
	aliClient.
		SetTimeout(base.DefaultTimeout).
		SetRetryCount(3).
		SetHeader("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/87.0.4280.88 Safari/537.36").
		SetHeader("content-type", "application/json").
		SetHeader("origin", "https://www.aliyundrive.com")
}
