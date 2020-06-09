package local_executor

import (
	"context"
	"crawlab/database"
	"crawlab/model"
	"crawlab/services"
	"crawlab/services/local_node"
	"crawlab/services/local_spider"
	"encoding/json"
	"github.com/apex/log"
	"github.com/cenkalti/backoff/v4"
	"github.com/panjf2000/ants/v2"
	"time"
)

type Executor struct {
	sm   *local_spider.SpiderManager
	pool *ants.Pool
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
	tMsg := services.TaskMessage{}
	if err := json.Unmarshal([]byte(msg), &tMsg); err != nil {
		log.Errorf("json string to struct error: %s", err.Error())
		return
	}

	// 获取任务
	task, err = model.GetTask(tMsg.Id)
	if err != nil {
		log.Errorf("execute task, get task error: %s", err.Error())
		return
	}
	// 获取爬虫
	spiderMode, err := task.GetSpider()
	if err != nil {
		return
	}
	return task, spiderMode, nil
}
func (e *Executor) createWorker() (*TaskWorker, error) {
	task, spider, err := e.getTask()
	if err != nil {
		return nil, err
	}
	localSpider, ok := e.sm.GetSpider(task.SpiderId.Hex())
	if !ok {
		e.sm.PutSpider(&spider)
		localSpider, _ = e.sm.GetSpider(task.SpiderId.Hex())
	} else {
		localSpider.LoadModel(&spider)
	}
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
		default:
			bp := backoff.NewConstantBackOff(1 * time.Second)
			_ = backoff.Retry(func() error {
				err := e.pool.Submit(func() {
					worker, err := e.createWorker()
					if err != nil {
						log.WithError(err).Errorf("createWorker failed")
						return
					}
					worker.Execute(e.sm)
				})
				if err == ants.ErrPoolOverload {
					return err
				}
				return nil
			}, bp)
		}
	}

}
