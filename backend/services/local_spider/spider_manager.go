package local_spider

import (
	"crawlab/model"
	"sync"
)

const (
	Md5File = "md5.txt"
)

type SpiderManager struct {
	spiders sync.Map
}

func NewSpiderManager() *SpiderManager {
	return &SpiderManager{spiders: sync.Map{}}
}

func (sm *SpiderManager) PutSpider(spider *model.Spider) {
	if _, ok := sm.spiders.Load(spider.Id); ok {
		return
	}
	sm.spiders.Store(spider.Id, &Spider{
		Spider:           spider,
		RWMutex:          sync.RWMutex{},
		shouldWaitUpdate: false,
	})
}
func (sm *SpiderManager) GetSpider(id string) (spider *Spider, exists bool) {
	s, ok := sm.spiders.Load(id)
	if !ok {
		return nil, false
	}
	return s.(*Spider), true
}

func (sm *SpiderManager) PublishSpider(spider *model.Spider) {

}
