package main

import (
	"FileTransfer/middlewares"
	"FileTransfer/models"
	"FileTransfer/pkg/drivers/operate"
	"fmt"
)

func main() {
	err := middlewares.InitRedis()
	if err != nil {
		return
	}
	fmt.Println("无限循环开始：")
	middlewares.GetPools(func(fileId string) error {
		fmt.Printf("回调成功，删除File_id:%s\n", fileId)
		models.DeleteFilesByFileId(fileId)
		operate.Delete(fileId)
		return nil
	})
}
