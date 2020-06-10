package spider_service

import (
	"context"
	"crawlab/constants"
	"crawlab/database"
	"crawlab/model"
	"crawlab/utils"
	"errors"
	"github.com/apex/log"
	"github.com/globalsign/mgo"
	uuid "github.com/satori/go.uuid"
	"github.com/spf13/viper"
	"go.uber.org/atomic"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strings"
	"sync"
)

var (
	ErrNotZipFile = errors.New("not a valid zip file")
)

type Spider struct {
	*model.Spider
	logger *log.Entry
	//MD5
	Md5 string
	sync.RWMutex
	shouldWaitUpdate bool
	downloadLocker   atomic.Int32
}

func (s *Spider) LogPath() (dir string, err error) {
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
	if err = os.MkdirAll(tmpPath, 0755); err != nil {
		log.Errorf("mkdir other.tmppath error: %v", err.Error())
		return
	}
	// 创建临时文件
	tmpFilePath := filepath.Join(tmpPath, randomId.String()+".zip")
	dir = utils.OpenFile(tmpFilePath)
	defer utils.Close(dir)
	// 将该文件写入临时文件
	if _, err = io.Copy(dir, gridFs); err != nil {
		log.Errorf("copy file error: %s, file_id: %s", err.Error(), gridFs.Id())
		debug.PrintStack()
		return
	}
	return dir, func() {
		dir.Close()
		// 删除临时文件
		if err := os.Remove(tmpFilePath); err != nil {
			log.Errorf(err.Error())
			debug.PrintStack()
			return
		}
	}, err
}

func (s *Spider) Sync() (err error) {
	if s.downloadLocker.Load() == 1 {
		return nil
	}
	session, gf := database.GetGridFs("files")
	defer session.Close()
retryDownload:

	gridFsFile, err := gf.OpenId(s.Id)
	defer utils.Close(gridFsFile)

	if err != nil {
		return err
	}
	//如果同需要同步
	if s.fileMD5() != gridFsFile.MD5() {
		if !s.downloadLocker.CAS(0, 1) {
			return
		}
		s.shouldWaitUpdate = true
		defer func() {
			s.shouldWaitUpdate = false
			s.Unlock()
			s.downloadLocker.Store(0)
		}()

		randomId := uuid.NewV4()
		tmpPath := viper.GetString("other.tmppath")
		if err = os.MkdirAll(tmpPath, 0755); err != nil {
			log.Errorf("mkdir other.tmppath error: %v", err.Error())
			return
		}
		// 创建临时文件
		tmpFilePath := filepath.Join(tmpPath, randomId.String()+".zip")
		srcFile := utils.OpenFile(tmpFilePath)
		defer utils.Close(srcFile)
		// 将该文件写入临时文件
		if _, err = io.Copy(srcFile, gridFsFile); err != nil {
			log.Errorf("copy file error: %s, file_id: %s", err.Error(), gridFsFile.Id())
			debug.PrintStack()
			return
		}

		md5, err := utils.FileMd5(tmpFilePath)
		if err != nil {
			return err
		}
		if md5 != gridFsFile.MD5() {
			s.logger.Warnf("download zip file md5 not match gridfs md5. local:%s,grid:%s", md5, gridFsFile.MD5())
			_ = os.Remove(tmpFilePath)
			goto retryDownload
		}
		// 解压缩临时文件到目标文件夹
		dstPath := s.localPath()
		if err = os.MkdirAll(dstPath, 0755); err != nil {
			return err
		}
		if err := utils.DeCompress(srcFile, dstPath); err != nil {
			log.Errorf(err.Error())
			debug.PrintStack()
			return err
		}
		//递归修改目标文件夹权限
		// 解决scrapy.setting中开启LOG_ENABLED 和 LOG_FILE时不能创建log文件的问题
		cmd := exec.Command("chmod", "-R", "777", dstPath)
		if err := cmd.Run(); err != nil {
			log.Errorf(err.Error())
			debug.PrintStack()
			return err
		}
		s.Lock()
		//爬虫文件有变化，先删除本地文件
		if err := os.RemoveAll(s.localPath()); err != nil {
			log.Errorf("remove spider files error: %s, path: %s", err.Error(), s.localPath())
			debug.PrintStack()
		}
		s.InstallDeps()
		s.CreateMd5File(s.fileMD5())
		if model.IsMaster() {
			s.CheckIsScrapy()
		}
	}
	return nil
}
func (s *Spider) CreateMd5File(md5 string) {
	fileName := filepath.Join(s.localPath(), Md5File)
	file := utils.OpenFile(fileName)
	defer utils.Close(file)
	if file != nil {
		if _, err := file.WriteString(md5 + "\n"); err != nil {
			log.Errorf("file write string error: %s", err.Error())
			debug.PrintStack()
		}
	}
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
func (t *Spider) waitSynced() {
	for {
		if t.Spider.UploadingGridFsFile {
			runtime.Gosched()
			continue
		}
		break
	}
}
func (t *Spider) Execute(task model.Task) (err error) {
	//等待
	t.waitSynced()

	err = t.Sync()
	if err != nil {
		return nil
	}
	// 工作目录
	cwd := filepath.Join(
		viper.GetString("spider.path"),
		t.Name,
	)
	// 执行命令
	var cmd string
	if t.Type == constants.Configurable {
		// 可配置爬虫命令
		cmd = "scrapy crawl config_spider"
	} else {
		// 自定义爬虫命令
		cmd = t.Cmd
	}

	// 加入参数
	if task.Param != "" {
		cmd += " " + task.Param
	}
	// 获得触发任务用户
	user, err := model.GetUser(task.UserId)
	if err != nil || err != mgo.ErrNotFound {
		return err
	}

	//创建索引
	col := utils.GetSpiderCol(t.Col, t.Name)
	s, c := database.GetCol(col)
	defer s.Close()
	_ = c.EnsureIndex(mgo.Index{
		Key: []string{"task_id"},
	})

	t.RLock()
	defer t.RUnlock()
	// 执行Shell命令
	return ExecuteShellCmd(context.Background(), cmd, cwd, task, *t.Spider, user)
}
func (t *Spider) GetLogDir() (dir string, err error) {
	// 日志目录
	fileDir := filepath.Join(viper.GetString("log.path"), t.Id.Hex())

	// 如果日志目录不存在，生成该目录
	if err := os.MkdirAll(fileDir, 0777); err != nil {
		return "", err
	}

	return fileDir, nil
}
func (s *Spider) CheckIsScrapy() {
	if s.Spider.Type == constants.Configurable {
		return
	}
	if viper.GetString("setting.checkScrapy") != "Y" {
		return
	}
	s.Spider.IsScrapy = utils.Exists(path.Join(s.Spider.Src, "scrapy.cfg"))
	if s.Spider.IsScrapy {
		s.Spider.Cmd = "scrapy crawl"
	}
	if err := s.Spider.Save(); err != nil {
		log.Errorf(err.Error())
		debug.PrintStack()
		return
	}
}
func (s *Spider) InstallDeps() {
	langs := utils.GetLangList()
	for _, l := range langs {
		// no dep file name is found, skip
		if l.DepFileName == "" {
			continue
		}

		// no dep file found, skip
		if !utils.Exists(path.Join(s.Spider.Src, l.DepFileName)) {
			continue
		}

		// no dep install executable found, skip
		if !utils.Exists(l.DepExecutablePath) {
			continue
		}

		// command to install dependencies
		cmd := exec.Command(l.DepExecutablePath, strings.Split(l.InstallDepArgs, " ")...)
		// working directory
		cmd.Dir = s.Spider.Src

		// compatibility with node.js
		if l.ExecutableName == constants.Nodejs {
			deps, err := utils.GetPackageJsonDeps(path.Join(s.Spider.Src, l.DepFileName))
			if err != nil {
				continue
			}
			cmd = exec.Command(l.DepExecutablePath, strings.Split(l.InstallDepArgs+" "+strings.Join(deps, " "), " ")...)
		}

		// start executing command
		output, err := cmd.Output()
		if err != nil {
			log.Errorf("install dep error: " + err.Error())
			log.Errorf(string(output))
			debug.PrintStack()
		}
	}
}
