package spider_service

import (
	"crawlab/model"
	"github.com/apex/log"
	"github.com/gin-gonic/gin"
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

func (sm *SpiderManager) PutSpider(spider *model.Spider) *Spider {
	if sp, ok := sm.spiders.Load(spider.Id); ok {
		sp.(*Spider).LoadModel(spider)
		return sp.(*Spider)
	}
	lsp := &Spider{
		Spider:           spider,
		RWMutex:          sync.RWMutex{},
		shouldWaitUpdate: false,
		logger:           log.WithField("p", "SpiderManager").Fields("e", "Spider"),
	}
	sm.spiders.Store(spider.Id, lsp)
	return lsp
}
func (sm *SpiderManager) GetSpider(id string) (spider *Spider, exists bool) {
	s, ok := sm.spiders.Load(id)
	if !ok {
		return nil, false
	}
	return s.(*Spider), true
}
func (sm *SpiderManager) FetchLocalSpider(spider *model.Spider, localPath string) (err error) {
	curr := sm.PutSpider(spider)
	tmpfile, err := curr.fetchSpiderCodeFromLocal(localPath)
	if err != nil {
		return err
	}
	return sm.fetchPool.Submit(func() {
		curr.uploadSpiderFileToGridFs(tmpfile)
	})
}
func (sm *SpiderManager) FetchUploadZipSpider(spider *model.Spider, ctx *gin.Context) (err error) {

	curr := sm.PutSpider(spider)
	tmpfile, err := curr.fetchSpiderCodeFromUpload(ctx)
	if err != nil {
		return err
	}
	return sm.fetchPool.Submit(func() {
		curr.uploadSpiderFileToGridFs(tmpfile)
	})
}
func (sm *SpiderManager) FetchGitSpider(spider *model.Spider) (err error) {
	curr := sm.PutSpider(spider)
	tmpfile, err := curr.fetchSpiderCodeFromGit()
	if err != nil {
		return err
	}

	return sm.fetchPool.Submit(func() {
		curr.uploadSpiderFileToGridFs(tmpfile)
	})
}
