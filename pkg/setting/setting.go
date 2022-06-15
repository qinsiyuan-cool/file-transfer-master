package setting

import (
	"github.com/go-ini/ini"
	"github.com/robfig/cron/v3"
	"log"
	"time"
)

var (
	Cfg                 *ini.File
	Cron                *cron.Cron
	RunModel            string
	AddrIP              string
	DownDomain          string
	HttpPort            int
	LimitCountPerSecond int
	ReadTimeout         time.Duration
	WriteTimeout        time.Duration
	JwtSecret           string
	Drivers             string
	DriversPath         string
	DriveId             string
	RefreshToken        string
	RedisAddr           string
	RedisPort           string
	RedisPass           string
	RedisDb             int
)

func init() {
	var err error
	Cfg, err = ini.Load("config/app.ini")
	if err != nil {
		log.Fatalf("Fail to parse 'conf/app.ini':%v", err)
	}
	loadConfig()
}

func loadConfig() {
	//加载项目运行模式
	RunModel = Cfg.Section("").Key("RUN_MODE").MustString("debug")
	//加载项目运行HTTP配置
	server, err := Cfg.GetSection("server")
	if err != nil {
		log.Fatalf("Fail to get section 'server':%v", err)
	}
	AddrIP = server.Key("ADDR_IP").MustString("0.0.0.0")
	HttpPort = server.Key("HTTP_PORT").MustInt(8999)
	ReadTimeout = time.Duration(server.Key("READ_TIMEOUT").MustInt(60)) * time.Second
	WriteTimeout = time.Duration(server.Key("WRITE_TIMEOUT").MustInt(60)) * time.Second

	//加载项目APP配置
	app, err := Cfg.GetSection("app")
	if err != nil {
		log.Fatalf("Fail to get section 'app':%v", err)
	}
	DownDomain = app.Key("DOWN_DOMAIN").MustString("https://mapi.net.cn/d/")
	LimitCountPerSecond = app.Key("LIMIT_COUNT_PER_SECOND").MustInt(7)
	JwtSecret = app.Key("JWT_SECRET").MustString("!@)*#)!@U#@*!@!)")
	Drivers = app.Key("DRIVERS").MustString("AliDrive")
	DriversPath = app.Key("DRIVERS_PATH").MustString("")
	DriveId = app.Key("DRIVE_ID").MustString("1495890")
	RefreshToken = app.Key("REFRESH_TOKEN").MustString("5329aee40509488e8e518791c22859f1")
	//加载数据库配置
	//加载Redis缓存配置
	redisConfig, err := Cfg.GetSection("redis")
	if err != nil {
		log.Fatalf("Fail to get section 'app':%v", err)
	}
	RedisAddr = redisConfig.Key("REDIS_ADDR").MustString("127.0.0.1")
	RedisPort = redisConfig.Key("REDIS_PORT").MustString("6379")
	RedisDb = redisConfig.Key("REDIS_DB").MustInt(0)
	RedisPass = redisConfig.Key("REDIS_PASS").MustString("")

}
