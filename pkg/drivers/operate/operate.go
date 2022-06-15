package operate

import (
	"FileTransfer/middlewares"
	_ "FileTransfer/pkg/drivers/alidrive"
	"FileTransfer/pkg/drivers/base"
	"FileTransfer/pkg/logging"
	"FileTransfer/pkg/setting"
	"fmt"
	log "github.com/sirupsen/logrus"
	"runtime/debug"
)

func Upload(file *middlewares.FileStream) (map[string]interface{}, error) {
	defer func() {
		_ = file.Close()
	}()
	driver, ok := base.GetDriver(setting.Drivers)
	if !ok {
		return base.Json{}, fmt.Errorf("no [%s] driver", setting.Drivers)
	}

	req, err := driver.Upload(file)
	if err != nil {
		logging.Logger.Errorf("upload error: %s", err.Error())
	}
	debug.FreeOSMemory()
	return req, err
}

func Delete(path string) error {
	driver, ok := base.GetDriver(setting.Drivers)
	if !ok {
		return fmt.Errorf("no [%s] driver", setting.Drivers)
	}
	err := driver.Delete(path)
	if err == nil {
		return nil
	}
	if err != nil {
		log.Errorf("delete error: %s", err.Error())
	}
	return err
}

func PersonalInfo() (map[string]interface{}, error) {
	driver, ok := base.GetDriver(setting.Drivers)
	if !ok {
		return base.Json{}, fmt.Errorf("no [%s] driver", setting.Drivers)
	}

	req, err := driver.PersonalInfo()
	if err != nil {
		logging.Logger.Errorf("upload error: %s", err.Error())
	}
	debug.FreeOSMemory()
	return map[string]interface{}{
		"total_size": req.TotalSize,
		"used_size":  req.UsedSize,
	}, err
}
