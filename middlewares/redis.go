package middlewares

import (
	"FileTransfer/pkg/logging"
	"FileTransfer/pkg/setting"
	"encoding/json"
	"github.com/go-redis/redis"
	"time"
)

var RedisConn *redis.Client
var Pool = "q_list"

// ProcessCallback 要执行的回调函数
type ProcessCallback func(fileId string) error

type QueueData struct {
	ID     float64 `json:"id"`
	FileId string  `json:"file_id"`
	Expire int     `json:"expire"`
}

func InitRedis() (err error) {
	logging.Logger.Infof("%s-%s-%s", setting.RedisAddr, setting.RedisPort, setting.RedisPass)
	RedisConn = redis.NewClient(&redis.Options{
		Network:  "tcp",
		Addr:     setting.RedisAddr + ":" + string(setting.RedisPort),
		Password: setting.RedisPass,
		DB:       setting.RedisDb,
	})
	_, err = RedisConn.Ping().Result()
	return nil
}
func AddPoolByLPush(data QueueData) error {
	logging.Logger.Infof("%s", data)
	marshal, err := json.Marshal(data)
	if err != nil {
		return err
	}
	logging.Logger.Infof("%s", marshal)
	RedisConn.LPush(Pool, marshal)
	return nil
}
func AddPoolByZSet(data QueueData) error {
	logging.Logger.Infof("%s", data)
	marshal, err := json.Marshal(data)
	if err != nil {
		return err
	}
	err = RedisConn.ZAdd(Pool, redis.Z{Score: data.ID, Member: marshal}).Err()
	if err != nil {
		panic(err)
	}
	return nil
}

func GetPools(cb ProcessCallback) error {
	var queueData QueueData
	for {
		poolJson, _ := RedisConn.ZRange(Pool, 0, -1).Result()
		if len(poolJson) > 0 {
			first := poolJson[0]
			err := json.Unmarshal([]byte(first), &queueData)
			if err != nil {
				continue
			}
			create := int(time.Now().Unix())
			if create > queueData.Expire {
				RedisConn.ZRemRangeByRank(Pool, 0, 0)
				cb(queueData.FileId)
			}
		} else {
			continue
		}
		time.Sleep(60 * time.Second)
	}
	return nil
}

func GetData() {

}
