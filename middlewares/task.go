package middlewares

import (
	"FileTransfer/pkg/logging"
	"sync"
	"time"
)

// ProcessTimerCallback 定时要执行的回调函数
type ProcessTimerCallback func(id int, err error) error

// timeTask 定时任务类
type timeTask struct {
	id      int
	create  int
	timeOut int
	cb      ProcessTimerCallback
}

var mutex sync.Mutex

// 任务列表
var taskList []timeTask

// 初始化并启动定时任务

func init() {
	//setting.Cron = cron.New()
	//setting.Cron.AddFunc("@every 2s", ProcessTimerTaskHandler)
	//setting.Cron.Start()
	//logging.Logger.Infof("计划任务初始化成功")
}

// InserTimerTask 插入定时任务
func InserTimerTask(id int, timeout int, cb ProcessTimerCallback) {
	var task timeTask
	task.id = id
	task.create = int(time.Now().Unix())
	task.timeOut = timeout
	task.cb = cb
	logging.Logger.Infof("计划任务添加成功")

	mutex.Lock()
	defer mutex.Unlock()

	for i := 0; i < len(taskList); i++ {
		// 如果id存在，只刷新创建时间和超时时间
		if taskList[i].id == task.id {
			taskList[i].create = task.create
			taskList[i].timeOut = task.timeOut
			return
		}
	}

	// 如果id不存在。增加一个task
	taskList = append(taskList, task)
}

// RemoveTimeTask 删除定时任务
func RemoveTimeTask(id int) {
	mutex.Lock()
	defer mutex.Unlock()

	for i := 0; i < len(taskList); i++ {
		if taskList[i].id == id {
			taskList = append(taskList[:i], taskList[i+1:]...)
			return
		}
	}
}

// ProcessTimerTaskHandler 定时处理让任务
func ProcessTimerTaskHandler() {
	//logging.Logger.Infof("执行计划任务%d.", len(taskList))
	var task timeTask
	mutex.Lock()
	defer mutex.Unlock()
	if len(taskList) == 0 {
		return
	}
	for i := 0; i < len(taskList); {
		task = taskList[i]
		//logging.Logger.Infof("执行计划任务%d-%d-%d.", task.id, task.create, task.timeOut)
		if int(time.Now().Unix()) > task.timeOut {
			// 执行回调函数
			task.cb(task.id, nil)
			// 删除定时任务
			taskList = append(taskList[:i], taskList[i+1:]...)
		} else {
			i++
		}
	}
}
