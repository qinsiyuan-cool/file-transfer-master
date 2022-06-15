package alidrive

import (
	"FileTransfer/config"
	"FileTransfer/middlewares"
	"FileTransfer/pkg/drivers/base"
	"FileTransfer/pkg/logging"
	"FileTransfer/pkg/setting"
	"FileTransfer/pkg/utils"
	"bytes"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"math/big"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

type AliDrive struct{}

func (driver AliDrive) Config() base.DriverConfig {
	return base.DriverConfig{
		Name: "AliDrive",
	}
}

func (driver AliDrive) Items() []base.Item {
	return []base.Item{
		{
			Name:     "refresh_token",
			Label:    "refresh token",
			Type:     base.TypeString,
			Required: true,
		},
		{
			Name:     "root_folder",
			Label:    "root folder file_id",
			Type:     base.TypeString,
			Required: false,
		},
		{
			Name:     "order_by",
			Label:    "order_by",
			Type:     base.TypeSelect,
			Values:   "name,size,updated_at,created_at",
			Required: false,
		},
		{
			Name:     "order_direction",
			Label:    "order_direction",
			Type:     base.TypeSelect,
			Values:   "ASC,DESC",
			Required: false,
		},
		{
			Name:        "limit",
			Label:       "limit",
			Type:        base.TypeNumber,
			Required:    false,
			Description: ">0 and <=200",
		},
		{
			Name:  "bool_1",
			Label: "fast upload",
			Type:  base.TypeBool,
		},
	}
}

func (driver AliDrive) Save() error {

	return nil
}

func (driver AliDrive) File(path string) (*middlewares.File, error) {
	path = utils.ParsePath(path)
	if path == "/" {
		return &middlewares.File{
			Id:        setting.DriversPath,
			Name:      setting.Drivers,
			Size:      0,
			Type:      config.FOLDER,
			Driver:    setting.Drivers,
			UpdatedAt: string(time.Now().Unix()),
		}, nil
	}
	dir, name := filepath.Split(path)
	files, err := driver.Files(dir)
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		if file.Name == name {
			return &file, nil
		}
	}
	return nil, base.ErrPathNotFound
}

func (driver AliDrive) Files(path string) ([]middlewares.File, error) {
	path = utils.ParsePath(path)
	var rawFiles []AliFile
	file, err := driver.File(path)
	if err != nil {
		return nil, err
	}
	rawFiles, err = driver.GetFiles(file.Id)
	if err != nil {
		return nil, err
	}
	files := make([]middlewares.File, 0)
	for _, file := range rawFiles {
		files = append(files, *driver.FormatFile(&file))
	}
	return files, nil
}

func (driver AliDrive) Link(args base.Args) (*base.Link, error) {
	var err error
	var resp base.Json
	var e AliRespError
	_, err = aliClient.R().SetResult(&resp).
		SetError(&e).
		SetHeader("authorization", "Bearer\t"+middlewares.GetAccessToken()).
		SetBody(base.Json{
			"drive_id":   setting.DriveId,
			"file_id":    args.Path,
			"expire_sec": 14400,
		}).Post("https://api.aliyundrive.com/v2/file/get_download_url")
	if err != nil {
		return nil, err
	}
	if e.Code != "" {
		if e.Code == "AccessTokenInvalid" {
			err = driver.RefreshToken()
			if err != nil {
				return nil, err
			} else {
				return driver.Link(args)
			}
		}
		return nil, fmt.Errorf("%s", e.Message)
	}
	return &base.Link{
		Headers: []base.Header{
			{
				Name:  "Referer",
				Value: "https://www.aliyundrive.com/",
			},
		},
		Url: resp["url"].(string),
	}, nil
}

func (driver AliDrive) Path(path string) (*middlewares.File, []middlewares.File, error) {
	path = utils.ParsePath(path)
	logging.Logger.Debugf("ali path: %s", path)
	file, err := driver.File(path)
	if err != nil {
		return nil, nil, err
	}
	if !file.IsDir() {
		return file, nil, nil
	}
	files, err := driver.Files(path)
	if err != nil {
		return nil, nil, err
	}
	return nil, files, nil
}

//func (driver AliDrive) Proxy(r *http.Request, account *model.Account) {
//	r.Header.Del("Origin")
//	r.Header.Set("Referer", "https://www.aliyundrive.com/")
//}

func (driver AliDrive) Preview(path string) (interface{}, error) {
	file, err := driver.GetFile(path)
	if err != nil {
		return nil, err
	}
	// office
	var resp base.Json
	var e AliRespError
	var url string
	req := base.Json{
		"drive_id": setting.DriveId,
		"file_id":  file.FileId,
	}
	switch file.Category {
	case "doc":
		{
			url = "https://api.aliyundrive.com/v2/file/get_office_preview_url"
			req["access_token"] = middlewares.GetAccessToken()
		}
	case "video":
		{
			url = "https://api.aliyundrive.com/v2/file/get_video_preview_play_info"
			req["category"] = "live_transcoding"
		}
	default:
		return nil, base.ErrNotSupport
	}
	_, err = aliClient.R().SetResult(&resp).SetError(&e).
		SetHeader("authorization", "Bearer\t"+middlewares.GetAccessToken()).
		SetBody(req).Post(url)
	if err != nil {
		return nil, err
	}
	if e.Code != "" {
		return nil, fmt.Errorf("%s", e.Message)
	}
	return resp, nil
}

func (driver AliDrive) MakeDir(path string) error {
	dir, name := filepath.Split(path)
	parentFile, err := driver.File(dir)
	if err != nil {
		return err
	}
	if !parentFile.IsDir() {
		return base.ErrNotFolder
	}
	var resp base.Json
	var e AliRespError
	_, err = aliClient.R().SetResult(&resp).SetError(&e).
		SetHeader("authorization", "Bearer\t"+middlewares.GetAccessToken()).
		SetBody(base.Json{
			"check_name_mode": "refuse",
			"drive_id":        setting.DriveId,
			"name":            name,
			"parent_file_id":  parentFile.Id,
			"type":            "folder",
		}).Post("https://api.aliyundrive.com/adrive/v2/file/createWithFolders")
	if e.Code != "" {
		if e.Code == "AccessTokenInvalid" {
			err = driver.RefreshToken()
			if err != nil {
				return err
			} else {
				return driver.MakeDir(path)
			}
		}
		return fmt.Errorf("%s", e.Message)
	}
	if resp["file_name"] == name {
		return nil
	}
	return fmt.Errorf("%+v", resp)
}

func (driver AliDrive) Move(src string, dst string) error {
	dstDir, _ := filepath.Split(dst)
	srcFile, err := driver.File(src)
	if err != nil {
		return err
	}
	dstDirFile, err := driver.File(dstDir)
	if err != nil {
		return err
	}
	err = driver.batch(srcFile.Id, dstDirFile.Id, "/file/move")
	return err
}

func (driver AliDrive) Rename(src string, dst string) error {
	_, dstName := filepath.Split(dst)
	srcFile, err := driver.File(src)
	if err != nil {
		return err
	}
	err = driver.rename(srcFile.Id, dstName)
	return err
}

func (driver AliDrive) Copy(src string, dst string) error {
	dstDir, _ := filepath.Split(dst)
	srcFile, err := driver.File(src)
	if err != nil {
		return err
	}
	dstDirFile, err := driver.File(dstDir)
	if err != nil {
		return err
	}
	err = driver.batch(srcFile.Id, dstDirFile.Id, "/file/copy")
	return err
}

func (driver AliDrive) Delete(path string) error {
	var e AliRespError
	res, err := aliClient.R().SetError(&e).
		SetHeader("authorization", "Bearer\t"+middlewares.GetAccessToken()).
		SetBody(base.Json{
			"drive_id": setting.DriveId,
			"file_id":  path,
		}).Post("https://api.aliyundrive.com/v2/recyclebin/trash")
	if err != nil {
		return err
	}
	if e.Code != "" {
		if e.Code == "AccessTokenInvalid" {
			err = driver.RefreshToken()
			if err != nil {
				return err
			} else {
				return driver.Delete(path)
			}
		}
		return fmt.Errorf("%s", e.Message)
	}
	if res.StatusCode() < 400 {
		return nil
	}
	return errors.New(res.String())
}

type UploadResp struct {
	FileName     string `json:"file_name"`
	FileId       string `json:"file_id"`
	UploadId     string `json:"upload_id"`
	PartInfoList []struct {
		UploadUrl string `json:"upload_url"`
	} `json:"part_info_list"`

	RapidUpload bool `json:"rapid_upload"`
}

func (driver AliDrive) Upload(file *middlewares.FileStream) (map[string]interface{}, error) {
	var err error
	if file == nil {
		return base.Json{}, nil
	}

	parentFile := &middlewares.File{
		Id:        setting.DriversPath,
		Name:      setting.Drivers,
		Size:      0,
		Type:      config.FOLDER,
		Driver:    setting.Drivers,
		UpdatedAt: string(time.Now().Unix()),
	}

	const DEFAULT int64 = 10485760
	var count = int(math.Ceil(float64(file.GetSize()) / float64(DEFAULT)))

	partInfoList := make([]base.Json, 0, count)
	for i := 1; i <= count; i++ {
		partInfoList = append(partInfoList, base.Json{"part_number": i})
	}

	reqBody := base.Json{
		"check_name_mode": "auto_rename",
		"drive_id":        setting.DriveId,
		"name":            file.GetFileName(),
		"parent_file_id":  parentFile.Id,
		"part_info_list":  partInfoList,
		"size":            file.GetSize(),
		"type":            "file",
	}

	bool := true
	if bool {
		buf := make([]byte, 1024)
		n, _ := file.Read(buf[:])
		reqBody["pre_hash"] = utils.GetSHA1Encode(string(buf[:n]))
		file.File = io.NopCloser(io.MultiReader(bytes.NewReader(buf[:n]), file.File))
	} else {
		reqBody["content_hash_name"] = "none"
		reqBody["proof_version"] = "v1"
	}

	var resp UploadResp
	var e AliRespError
	client := aliClient.R().SetResult(&resp).SetError(&e).SetHeader("authorization", "Bearer\t"+middlewares.GetAccessToken()).SetBody(reqBody)

	_, err = client.Post("https://api.aliyundrive.com/adrive/v2/file/createWithFolders")
	if err != nil {
		return base.Json{}, err
	}
	if e.Code != "" && e.Code != "PreHashMatched" {
		if e.Code == "AccessTokenInvalid" {
			err = driver.RefreshToken()
			if err != nil {
				return base.Json{}, err
			} else {
				return driver.Upload(file)
			}
		}
		return base.Json{}, fmt.Errorf("%s", e.Message)
	}

	if e.Code == "PreHashMatched" && bool {
		tempFile, err := ioutil.TempFile("data/temp", "file-*")
		if err != nil {
			return base.Json{}, err
		}

		defer tempFile.Close()
		defer os.Remove(tempFile.Name())

		delete(reqBody, "pre_hash")
		h := sha1.New()
		if _, err = io.Copy(io.MultiWriter(tempFile, h), file.File); err != nil {
			return base.Json{}, err
		}
		reqBody["content_hash"] = hex.EncodeToString(h.Sum(nil))
		reqBody["content_hash_name"] = "sha1"
		reqBody["proof_version"] = "v1"

		/*
			js 隐性转换太坑不知道有没有bug
			var n = e.access_token，
			r = new BigNumber('0x'.concat(md5(n).slice(0, 16)))，
			i = new BigNumber(t.file.size)，
			o = i ? r.mod(i) : new gt.BigNumber(0);
			(t.file.slice(o.toNumber(), Math.min(o.plus(8).toNumber(), t.file.size)))
		*/
		buf := make([]byte, 8)
		r, _ := new(big.Int).SetString(utils.GetMD5Encode(middlewares.GetAccessToken())[:16], 16)
		i := new(big.Int).SetUint64(file.Size)
		o := r.Mod(r, i)
		n, _ := io.NewSectionReader(tempFile, o.Int64(), 8).Read(buf[:8])
		reqBody["proof_code"] = base64.StdEncoding.EncodeToString(buf[:n])

		_, err = client.Post("https://api.aliyundrive.com/adrive/v2/file/createWithFolders")
		if err != nil {
			return base.Json{}, err
		}
		if e.Code != "" && e.Code != "PreHashMatched" {
			return base.Json{}, fmt.Errorf("%s", e.Message)
		}

		if resp.RapidUpload {
			return map[string]interface{}{
				"file_id":   resp.FileId,
				"file_name": resp.FileName,
				"upload_id": resp.UploadId,
			}, nil
		}

		if _, err = tempFile.Seek(0, io.SeekStart); err != nil {
			return base.Json{}, err
		}
		file.File = tempFile
	}

	for _, partInfo := range resp.PartInfoList {
		req, err := http.NewRequest("PUT", partInfo.UploadUrl, io.LimitReader(file.File, DEFAULT))
		if err != nil {
			return base.Json{}, err
		}
		res, err := base.HttpClient.Do(req)
		if err != nil {
			return base.Json{}, err
		}
		logging.Logger.Debugf("%+v", res)
		//res, err := base.BaseClient.R().
		//	SetHeader("Content-Type","").
		//	SetBody(byteData).Put(resp.PartInfoList[i].UploadUrl)
		//if err != nil {
		//	return err
		//}
		//log.Debugf("put to %s : %d,%s", resp.PartInfoList[i].UploadUrl, res.StatusCode(),res.String())
	}
	var resp2 base.Json
	_, err = aliClient.R().SetResult(&resp2).SetError(&e).
		SetHeader("authorization", "Bearer\t"+middlewares.GetAccessToken()).
		SetBody(base.Json{
			"drive_id":  setting.DriveId,
			"file_id":   resp.FileId,
			"upload_id": resp.UploadId,
		}).Post("https://api.aliyundrive.com/v2/file/complete")
	if err != nil {
		return base.Json{}, err
	}
	if e.Code != "" && e.Code != "PreHashMatched" {
		//if e.Code == "AccessTokenInvalid" {
		//	err = driver.RefreshToken(account)
		//	if err != nil {
		//		return err
		//	} else {
		//		_ = model.SaveAccount(account)
		//		return driver.Upload(file, account)
		//	}
		//}
		return base.Json{}, fmt.Errorf("%s", e.Message)
	}
	if resp2["file_id"] == resp.FileId {
		return map[string]interface{}{
			"file_id":   resp2["file_id"],
			"file_name": resp2["name"],
			"upload_id": resp2["upload_id"],
		}, nil
	}
	return base.Json{}, fmt.Errorf("%+v", resp2)
}

func (driver AliDrive) PersonalInfo() (*base.Personal, error) {
	var err error
	var resp base.Json
	var e AliRespError
	_, err = aliClient.R().SetResult(&resp).
		SetError(&e).
		SetHeader("authorization", "Bearer\t"+middlewares.GetAccessToken()).
		SetBody(base.Json{
			"drive_id": setting.DriveId,
		}).Post("https://api.aliyundrive.com/v2/databox/get_personal_info")
	if err != nil {
		return nil, err
	}
	if e.Code != "" {
		if e.Code == "AccessTokenInvalid" {
			err = driver.RefreshToken()
			if err != nil {
				return nil, err
			} else {
				return driver.PersonalInfo()
			}
		}
		return nil, fmt.Errorf("%s", e.Message)
	}

	return &base.Personal{
		TotalSize: resp["personal_space_info"].(map[string]interface{})["total_size"],
		UsedSize:  resp["personal_space_info"].(map[string]interface{})["used_size"],
	}, nil
}

var _ base.Driver = (*AliDrive)(nil)
