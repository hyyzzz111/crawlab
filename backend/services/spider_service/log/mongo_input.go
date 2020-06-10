package log

import (
	"crawlab/model"
	"crawlab/utils"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"sync"
	"time"
)

type MongoInput struct {
	collection     *mgo.Collection
	seq            int64
	taskId         string
	expireDuration time.Duration
	cacheLogItems  []*model.LogItem
	sync.Mutex
}

func NewMongoInput(collection *mgo.Collection, taskId string, expireDuration time.Duration) *MongoInput {
	return &MongoInput{collection: collection, taskId: taskId, expireDuration: expireDuration}
}

func (g *MongoInput) Flush() {
	g.Lock()
	defer g.Unlock()
	bulk := g.collection.Bulk()
	for _, item := range g.cacheLogItems {
		bulk.Insert(item)
	}
	bulk.Run()
}
func (g *MongoInput) Write(p []byte) (n int, err error) {
	g.seq++
	l := &model.LogItem{
		Id:       bson.NewObjectId(),
		Seq:      g.seq,
		Message:  utils.BytesToString(p),
		TaskId:   g.taskId,
		Ts:       time.Now(),
		ExpireTs: time.Now().Add(g.expireDuration),
	}
	g.cacheLogItems = append(g.cacheLogItems, l)
	if len(g.cacheLogItems) > 20 {
		g.Flush()
	}
	return 0, nil
}
