package executor_service

import (
	"context"
	"crawlab/database"
	"crawlab/model"
	"crawlab/pkg/types"
	"crawlab/services/local_node"
	"crawlab/services/spider_service"
	"encoding/json"
	"github.com/apex/log"
	"github.com/cenkalti/backoff/v4"
	"github.com/gomodule/redigo/redis"
	"github.com/panjf2000/ants/v2"
	"github.com/spf13/viper"
	"time"
)

type Executor struct {
	sm     *spider_service.SpiderManager
	pool   *ants.Pool
	logger *log.Entry
}

func (e *Executor) SpiderManager() *spider_service.SpiderManager {
	return e.sm
}

func NewExecutor(pooSize int) *Executor {
	pool, _ := ants.NewPool(pooSize)
	return &Executor{
		sm:     spider_service.NewSpiderManager(),
		pool:   pool,
		logger: log.WithField("p", "Executor"),
	}
}

func (e *Executor) getTask() (task model.Task, sp model.Spider, err error) {

	// 获取当前节点
	node := local_node.CurrentNode()

	// 节点队列
	queueCur := "tasks:node:" + node.Id.Hex()

	// 节点队列任务
	var msg string
	if msg, err = database.RedisClient.LPop(queueCur); err != nil {
		// 节点队列没有任务，获取公共队列任务
		queuePub := "tasks:public"
		if msg, err = database.RedisClient.LPop(queuePub); err != nil {
		}
	}

	// 如果没有获取到任务，返回
	if msg == "" {
		return
	}
	// 反序列化
	tMsg := types.TaskMessage{}
	if err = json.Unmarshal([]byte(msg), &tMsg); err != nil {
		e.logger.Errorf("json string to struct error: %s", err.Error())
		return
	}

	// 获取任务
	task, err = model.GetTask(tMsg.Id)
	if err != nil {
		e.logger.Errorf("execute task, get task error: %s", err.Error())
		return
	}
	// 获取爬虫
	spiderMode, err := task.GetSpider()
	if err != nil {
		return
	}
	return task, spiderMode, nil
}
func (e *Executor) getTaskBlock(ctx context.Context) (model.Task, model.Spider, error) {
	bp := backoff.NewExponentialBackOff()
	bc := backoff.WithContext(bp, ctx)
	var task model.Task
	var spider model.Spider
	err := backoff.Retry(func() (err error) {
		task, spider, err = e.getTask()
		if err == redis.ErrNil {
			e.logger.Infof("wait task")
		}
		return err
	}, bc)
	return task, spider, err
}
func (e *Executor) createWorker(ctx context.Context) (*TaskWorker, error) {
	task, spider, err := e.getTaskBlock(ctx)
	if err != nil {
		return nil, err
	}
	e.sm.PutSpider(&spider)
	localSpider, _ := e.sm.GetSpider(task.SpiderId.Hex())
	return &TaskWorker{
		task:   task,
		spider: localSpider,
	}, err
}
func (e *Executor) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			e.pool.Release()
			break
		default:
			err := e.pool.Submit(func() {
				worker, err := e.createWorker(ctx)
				if err != nil {
					if err == redis.ErrNil {
						e.logger.Infof("pull task empty. waiting..")
						return
					}
					e.logger.WithError(err).Errorf("createWorker failed")
					return
				}
				worker.Execute()
			})
			if err != ants.ErrPoolOverload {
				time.Sleep(3 * time.Second)
			}
		}
	}

}

func (e *Executor) RunGitSync() {

}

var localExecutor *Executor

func Default() *Executor {
	return localExecutor
}
func InitExecutor() error {
	localExecutor := NewExecutor(2)
	if model.IsMaster() {
		go func() {
			localExecutor.RunGitSync()
		}()
	}
	// 如果不允许主节点运行任务，则跳过
	if model.IsMaster() && viper.GetString("setting.runOnMaster") == "N" {
		return nil
	}

	go func() {
		localExecutor.Run(context.TODO())
	}()
	return nil
}
