package spider_service

import (
	"crawlab/model"
	"github.com/apex/log"
	"github.com/panjf2000/ants/v2"
	"sync"
)

const (
	Md5File = "md5.txt"
)

type SpiderManager struct {
	spiders   sync.Map
	fetchPool *ants.Pool
}

func NewSpiderManager() *SpiderManager {
	pool, _ := ants.NewPool(10)
	return &SpiderManager{spiders: sync.Map{}, fetchPool: pool}
}

func (sm *SpiderManager) InjectSpider(spider *model.Spider) *Spider {
	if sp, ok := sm.spiders.Load(spider.Id); ok {
		sp.(*Spider).LoadModel(spider)
		return sp.(*Spider)
	}
	lsp := &Spider{
		Spider:           spider,
		RWMutex:          sync.RWMutex{},
		shouldWaitUpdate: false,
		logger:           log.WithField("p", "SpiderManager").WithField("e", "Spider"),
	}
	sm.spiders.Store(spider.Id, lsp)
	return lsp
}
