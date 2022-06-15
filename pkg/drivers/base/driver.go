package base

import (
	"FileTransfer/middlewares"
	"github.com/go-resty/resty/v2"
	log "github.com/sirupsen/logrus"
	"net/http"
	"time"
)

type DriverConfig struct {
	Name          string
	OnlyProxy     bool // 必须使用代理（本机或者其他机器）
	OnlyLocal     bool // 必须本机返回的
	ApiProxy      bool // 使用API中转的
	NoNeedSetLink bool // 不需要设置链接的
	NoCors        bool // 不可以跨域
	LocalSort     bool // 本地排序
}

type Args struct {
	Path string
	IP   string
}

type Driver interface {
	// Config 配置
	Config() DriverConfig
	PersonalInfo() (*Personal, error)
	// Items 账号所需参数
	Items() []Item
	// Save 保存时处理
	Save() error
	// File 取文件
	File(path string) (*middlewares.File, error)
	// Files 取文件夹
	Files(path string) ([]middlewares.File, error)
	// Link 取链接
	Link(args Args) (*Link, error)
	// Path 取路径（文件或文件夹）
	Path(path string) (*middlewares.File, []middlewares.File, error)
	// Deprecated Proxy 代理处理
	//Proxy(r *http.Request, account *model.Account)
	// Preview 预览
	Preview(path string) (interface{}, error)
	// MakeDir 创建文件夹
	MakeDir(path string) error
	// Move 移动/改名
	Move(src string, dst string) error
	// Rename 改名
	Rename(src string, dst string) error
	// Copy 拷贝
	Copy(src string, dst string) error
	// Delete 删除
	Delete(path string) error
	// Upload 上传
	Upload(file *middlewares.FileStream) (map[string]interface{}, error)
	// TODO
	//Search(path string, keyword string, account *model.Account) ([]*model.File, error)
}

type Item struct {
	Name        string `json:"name"`
	Label       string `json:"label"`
	Type        string `json:"type"`
	Default     string `json:"default"`
	Values      string `json:"values"`
	Required    bool   `json:"required"`
	Description string `json:"description"`
}

var driversMap = map[string]Driver{}

func RegisterDriver(driver Driver) {
	log.Infof("register driver: [%s]", driver.Config().Name)
	driversMap[driver.Config().Name] = driver
}

func GetDriver(name string) (driver Driver, ok bool) {
	driver, ok = driversMap[name]
	return
}

var NoRedirectClient *resty.Client
var RestyClient = resty.New()
var HttpClient = &http.Client{}
var UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/87.0.4280.88 Safari/537.36"
var DefaultTimeout = time.Second * 20

func init() {
	NoRedirectClient = resty.New().SetRedirectPolicy(
		resty.RedirectPolicyFunc(func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}),
	)
	NoRedirectClient.SetHeader("user-agent", UserAgent)
	RestyClient.SetHeader("user-agent", UserAgent)
	RestyClient.SetRetryCount(3)
	RestyClient.SetTimeout(DefaultTimeout)
}
