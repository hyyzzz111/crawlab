package executor_service

import (
	"context"
	"crawlab/constants"
	"crawlab/model"
	"crawlab/services"
	"crawlab/services/local_node"
	"crawlab/services/local_spider"
	"github.com/apex/log"
	"path/filepath"
	"runtime/debug"
	"time"
)

type TaskDelegate interface {
	OnStart()
}
type TaskWorker struct {
	task     model.Task
	spider   *spider_service.Spider
	taskUser model.User
}

func (t *TaskWorker) markTaskRunning() {
	// 获取当前节点
	node := local_node.CurrentNode()
	// 任务赋值
	t.task.NodeId = node.Id                                             // 任务节点信息
	t.task.StartTs = time.Now()                                         // 任务开始时间
	t.task.Status = constants.StatusRunning                             // 任务状态
	t.task.WaitDuration = t.task.StartTs.Sub(t.task.CreateTs).Seconds() // 等待时长
	// 储存任务
	_ = t.task.Save()
}
func (t *TaskWorker) markTaskFinish() {
	// 统计数据
	t.task.Status = constants.StatusFinished                               // 任务状态: 已完成
	t.task.FinishTs = time.Now()                                           // 结束时间
	t.task.RuntimeDuration = t.task.FinishTs.Sub(t.task.StartTs).Seconds() // 运行时长
	t.task.TotalDuration = t.task.FinishTs.Sub(t.task.CreateTs).Seconds()  // 总时长
	_ = t.task.Save()
}

func (t *TaskWorker) updateResultCount() {
	model.UpdateTaskResultCount(t.task.Id)
}
func (t *TaskWorker) watchEvent(ctx context.Context) {
	timer1 := time.NewTicker(time.Second * 5)
	timer2 := time.NewTicker(time.Second * 30)
	defer func() {
		timer1.Stop()
		t.updateResultCount()
	}()
	for {
		select {
		case <-ctx.Done():
			break
		case <-timer1.C:
			t.updateResultCount()
		case <-timer2.C:

		}
	}
}
func (t *TaskWorker) updateErrorLogs() {
	u, err := model.GetUser(t.task.UserId)
	if err != nil {
		return
	}
	if err := model.UpdateTaskErrorLogs(t.task.Id, u.Setting.ErrorRegexPattern); err != nil {
		return
	}
	if err := model.UpdateErrorLogCount(t.task.Id); err != nil {
		return
	}

}

func (t *TaskWorker) Execute() {
	// 获得触发任务用户
	user, err := model.GetUser(t.task.UserId)
	if err != nil {
		return
	}

	// 开始计时
	t.markTaskRunning()
	// 发送 Web Hook 请求 (任务开始)
	go services.SendWebHookRequest(user, t.task, *t.spider.Spider)

	ctx := context.Background()
	go t.watchEvent(ctx)
	defer func() {
		ctx.Done()
	}()
	// 创建日志目录
	t.task.LogPath, err = t.GetLogFilePath()
	if err != nil {
		return
	}
	if err = t.spider.Execute(t.task); err != nil {
		log.WithError(err).Errorf("TASK:%s:%w", t.spider.Cmd, err.Error())
		// 如果发生错误，则发送通知
		if user.Setting.NotificationTrigger == constants.NotificationTriggerOnTaskEnd || user.Setting.NotificationTrigger == constants.NotificationTriggerOnTaskError {
			services.SendNotifications(user, t.task, *t.spider.Spider)
		}
		// 发送 Web Hook 请求 (任务开始)
		go services.SendWebHookRequest(user, t.task, *t.spider.Spider)

		return
	}

	// 发送 Web Hook 请求 (任务结束)
	go services.SendWebHookRequest(user, t.task, *t.spider.Spider)

	// 如果是任务结束时发送通知，则发送通知
	if user.Setting.NotificationTrigger == constants.NotificationTriggerOnTaskEnd {
		services.SendNotifications(user, t.task, *t.spider.Spider)
	}
	t.markTaskFinish()
}
func (t *TaskWorker) GetLogFilePath() (dir string, err error) {
	// 日志目录
	fileDir, err := t.spider.GetLogDir()

	// 如果日志目录不存在，生成该目录
	if err != nil {
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
