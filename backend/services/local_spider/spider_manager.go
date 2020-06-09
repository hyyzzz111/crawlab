package local_spider

import (
	"context"
	"crawlab/database"
	"crawlab/model"
	"crawlab/utils"
	"github.com/apex/log"
	"github.com/globalsign/mgo"
	uuid "github.com/satori/go.uuid"
	"github.com/spf13/viper"
	"go.uber.org/atomic"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime/debug"
	"sync"
	"time"
)

const (
	Md5File = "md5.txt"
)

type Spider struct {
	*model.Spider
	//MD5
	Md5 string
	sync.RWMutex
	shouldWaitUpdate bool
	downloadLocker   atomic.Int32
}

func (s *Spider) LogPath() {
	// 日志目录
	fileDir := filepath.Join(viper.GetString("log.path"), s.Id.Hex())

	if err := os.MkdirAll(fileDir, 0777); err != nil {
		log.Errorf("execute task, make log dir error: %s", err.Error())
		debug.PrintStack()
		return "", err
	}
	return fileDir, nil
}
func (s *Spider) localPath() string {
	//目录不存在，则直接下载
	return filepath.Join(viper.GetString("spider.path"), s.Name)
}
func (s *Spider) fileMD5() string {
	return utils.ReadFileOneLine(filepath.Join(s.localPath(), Md5File))
}
func (s *Spider) downloadZipToTmp(gridFs *mgo.GridFile) (dir *os.File, cleanup func(), err error) {
	// 生成唯一ID
	randomId := uuid.NewV4()
	tmpPath := viper.GetString("other.tmppath")
	if err := os.MkdirAll(tmpPath, 0755); err != nil {
		log.Errorf("mkdir other.tmppath error: %v", err.Error())
		return
	}
	// 创建临时文件
	tmpFilePath := filepath.Join(tmpPath, randomId.String()+".zip")
	tmpFile := utils.OpenFile(tmpFilePath)
	defer utils.Close(tmpFile)
	// 将该文件写入临时文件
	if _, err := io.Copy(tmpFile, gridFs); err != nil {
		log.Errorf("copy file error: %s, file_id: %s", err.Error(), gridFs.Id())
		debug.PrintStack()
		return
	}
	return tmpFile, func() {
		tmpFile.Close()
		// 删除临时文件
		if err := os.Remove(tmpFilePath); err != nil {
			log.Errorf(err.Error())
			debug.PrintStack()
			return
		}
	}, err
}
func (s *Spider) run(fn func()) bool {
	s.tryWaitUpdate(context.Background())
	s.Sync()
	s.RLock()
	defer s.RUnlock()
	fn()
	return true
}
func (s *Spider) tryWaitUpdate(ctx context.Context) {
	if s.shouldWaitUpdate {
		timeout, _ := context.WithTimeout(ctx, time.Second*3)
		for {
			select {
			case <-timeout.Done():
				return
			default:
				if !s.shouldWaitUpdate {
					break
				}
			}
		}
	}
}
func (s *Spider) Sync() (err error) {
	if s.downloadLocker.Load() == 1 {
		return nil
	}

	session, gf := database.GetGridFs("files")
	defer session.Close()
	f, err := gf.OpenId(s.Id)
	defer utils.Close(f)

	if err != nil {
		return err
	}
	//如果同需要同步
	if s.fileMD5() != f.MD5() {
		if !s.downloadLocker.CAS(0, 1) {
			return
		}
		defer func() {
			s.shouldWaitUpdate = false
			s.Unlock()
			s.downloadLocker.Store(0)
		}()
		s.shouldWaitUpdate = true
		s.Lock()
		//下载爬虫代码到临时目录
		tmpFile, cleanup, err := s.downloadZipToTmp(f)
		if err != nil {
			return
		}
		defer cleanup()
		// 解压缩临时文件到目标文件夹
		dstPath := s.localPath()
		if err := utils.DeCompress(tmpFile, dstPath); err != nil {
			log.Errorf(err.Error())
			debug.PrintStack()
			return
		}
		//递归修改目标文件夹权限
		// 解决scrapy.setting中开启LOG_ENABLED 和 LOG_FILE时不能创建log文件的问题
		cmd := exec.Command("chmod", "-R", "777", dstPath)
		if err := cmd.Run(); err != nil {
			log.Errorf(err.Error())
			debug.PrintStack()
			return
		}
	}
	return nil
}

func (s *Spider) LoadModel(s2 *model.Spider) {
	s.shouldWaitUpdate = true
	s.Lock()
	defer func() {
		s.Unlock()
		s.shouldWaitUpdate = false
	}()
	s.Spider = s2
}

type SpiderManager struct {
	spiders sync.Map
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
