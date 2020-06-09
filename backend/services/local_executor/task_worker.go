package local_executor

import (
	"context"
	"crawlab/constants"
	"crawlab/database"
	"crawlab/model"
	"crawlab/services"
	"crawlab/services/local_node"
	"crawlab/services/local_spider"
	"crawlab/utils"
	"encoding/json"
	"github.com/apex/log"
	"github.com/globalsign/mgo"
	"github.com/spf13/viper"
	"os"
	"path/filepath"
	"runtime/debug"
	"time"
)

type TaskWorker struct {
	task   model.Task
	spider local_spider.Spider
}

func (t *TaskWorker) Execute(spiderManager *local_spider.SpiderManager) {
	//如果需要跳过任务执行进行文件最长等待三秒钟
	t.spider.TryWaitUpdate(context.Background())
	// 开始计时
	tic := time.Now()

	// 获取当前节点
	node := local_node.CurrentNode()

	err := t.spider.Sync()
	if err != nil {
		return
	}
	// 创建日志目录
	t.task.LogPath, err = t.GetLogFilePath()
	if err != nil {
		return
	}
	// 工作目录
	cwd := filepath.Join(
		viper.GetString("spider.path"),
		t.spider.Name,
	)
	// 执行命令
	var cmd string
	if t.spider.Type == constants.Configurable {
		// 可配置爬虫命令
		cmd = "scrapy crawl config_spider"
	} else {
		// 自定义爬虫命令
		cmd = t.spider.Cmd
	}

	// 加入参数
	if t.task.Param != "" {
		cmd += " " + t.task.Param
	}
	// 获得触发任务用户
	user, err := model.GetUser(t.task.UserId)
	if err != nil {
		return
	}

	// 任务赋值
	t.task.NodeId = node.Id                                             // 任务节点信息
	t.task.StartTs = time.Now()                                         // 任务开始时间
	t.task.Status = constants.StatusRunning                             // 任务状态
	t.task.WaitDuration = t.task.StartTs.Sub(t.task.CreateTs).Seconds() // 等待时长
	// 储存任务
	_ = t.task.Save()

	//创建索引
	col := utils.GetSpiderCol(t.spider.Col, t.spider.Name)
	s, c := database.GetCol(col)
	defer s.Close()
	_ = c.EnsureIndex(mgo.Index{
		Key: []string{"task_id"},
	})

}
func (t *TaskWorker) GetLogFilePath() (dir string, err error) {
	// 日志目录
	fileDir := filepath.Join(viper.GetString("log.path"), t.spider.Id.Hex())

	// 如果日志目录不存在，生成该目录
	if err := os.MkdirAll(fileDir, 0777); err != nil {
		log.Errorf("execute task, make log dir error: %s", err.Error())
		debug.PrintStack()
		return "", err
	}
	// 时间戳
	ts := time.Now()
	tsStr := ts.Format("20060102150405")

	// stdout日志文件
	filePath := filepath.Join(fileDir, t.task.Id+"_"+tsStr+".log")
	return filePath, nil
}
