package local_spider

import (
	"context"
	"crawlab/constants"
	"crawlab/database"
	"crawlab/model"
	"crawlab/services"
	"crawlab/services/spider_handler"
	"crawlab/utils"
	"github.com/apex/log"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	uuid "github.com/satori/go.uuid"
	"github.com/spf13/viper"
	"go.uber.org/atomic"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/http"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/ssh"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"runtime/debug"
	"strings"
	"sync"
	"time"
)

type Spider struct {
	*model.Spider
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
func (s *Spider) Run(fn func()) bool {
	s.RLock()
	defer s.RUnlock()
	fn()
	return true
}
func (s *Spider) TryWaitUpdate(ctx context.Context, duration time.Duration) {
	if s.shouldWaitUpdate {
		timeout, _ := context.WithTimeout(ctx, duration)
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

func (s *Spider) publishFromGit() (err error) {
	randomId := uuid.NewV4()
	tmpPath := viper.GetString("other.tmppath")
	tmpFilePath := filepath.Join(tmpPath, randomId.String())

	if err = os.MkdirAll(tmpFilePath, 0755|os.ModeDir); err != nil {
		log.Errorf("mkdir other.tmppath error: %v", err.Error())
		return
	}
	currentBranch := plumbing.NewBranchReferenceName(s.GitBranch)
	var cloneOption = &git.CloneOptions{
		URL:               s.GitUrl,
		RemoteName:        "origin",
		SingleBranch:      true,
		ReferenceName:     currentBranch,
		Depth:             1,
		RecurseSubmodules: 1,
	}

	if s.GitUsername != "" && s.GitPassword != "" {

		cloneOption.Auth = &http.BasicAuth{
			Username: s.GitUsername,
		}
	} else {
		if !strings.HasPrefix(s.GitUrl, "http") {
			// 为 SSH
			regex := regexp.MustCompile("^(?:ssh://?)?([0-9a-zA-Z_]+)@")
			res := regex.FindStringSubmatch(s.GitUrl)
			username := s.GitUsername
			if username == "" {
				if len(res) > 1 {
					username = res[1]
				} else {
					username = "git"
				}
			}
			cloneOption.Auth, err = ssh.NewPublicKeysFromFile(username, path.Join(os.Getenv("HOME"), ".ssh", "id_rsa"), "")
			if err != nil {
				log.Error(err.Error())
				debug.PrintStack()
				services.SaveSpiderGitSyncError(*s.Spider, err.Error())
				return err
			}
		}
	}

	rep, err := git.PlainClone(tmpFilePath, false, cloneOption)
	if err != nil {
		services.SaveSpiderGitSyncError(*s.Spider, err.Error())
		return err
	}
	ref, err := rep.Reference(currentBranch, true)
	if err != nil {
		return err
	}
	if s.Spider.GitHash == ref.Hash().String() {
		return nil
	}
	// 打包为 zip 文件
	files, err := utils.GetFilesFromDir(tmpFilePath)
	if err != nil {
		return err
	}
	tmpZipPath := filepath.Join(viper.GetString("other.tmppath"), s.Name+"."+randomId.String()+".zip")
	spiderZipFileName := s.Name + ".zip"
	if err := utils.Compress(files, tmpZipPath); err != nil {
		return err
	}
	//根据zip生成md5
	// 获取 GridFS 实例
	session, gf := database.GetGridFs("files")
	defer session.Close()

	// 判断文件是否已经存在
	var gfFile model.GridFs
	if err := gf.Find(bson.M{"filename": spiderZipFileName}).One(&gfFile); err == nil {
		//if gfFile.
		// 已经存在文件，则删除
		log.Errorf(gfFile.Id.Hex() + " already exists. removing...")
		if err := gf.RemoveId(gfFile.Id); err != nil {
			log.Errorf(err.Error())
			debug.PrintStack()
			return err
		}
	}

	// 上传到GridFs
	fid, err := services.UploadToGridFs(spiderZipFileName, tmpFilePath)
	if err != nil {
		log.Errorf("upload to grid fs error: %s", err.Error())
		return err
	}

	// 保存爬虫 FileId
	spider.FileId = fid
	if err := spider.Save(); err != nil {
		return err
	}

	// 获取爬虫同步实例
	spiderSync := spider_handler.SpiderSync{
		Spider: spider,
	}

	// 获取gfFile
	gfFile2 := model.GetGridFs(spider.FileId)

	// 生成MD5
	spiderSync.CreateMd5File(gfFile2.Md5)

	// 检查是否为 Scrapy 爬虫
	spiderSync.CheckIsScrapy()

	return nil

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
		if err == mgo.ErrNotFound {
			if s.IsGit {

			}

		}
		return err
	}
	//如果同需要同步
	if s.fileMD5() != f.MD5() {
		if !s.downloadLocker.CAS(0, 1) {
			return
		}
		s.shouldWaitUpdate = true
		defer func() {
			s.shouldWaitUpdate = false
			s.Unlock()
			s.downloadLocker.Store(0)
		}()

		//下载爬虫代码到临时目录
		tmpFile, cleanup, err := s.downloadZipToTmp(f)
		if err != nil {
			return err
		}
		defer cleanup()
		// 解压缩临时文件到目标文件夹
		dstPath := s.localPath()
		if err = os.MkdirAll(dstPath, 0755); err != nil {
			return err
		}
		if err := utils.DeCompress(tmpFile, dstPath); err != nil {
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

func (t *Spider) Execute(task model.Task) error {
	//如果需要跳过任务执行进行文件最长等待三秒钟
	//t.spider.TryWaitUpdate(context.Background(),time.Second*3)
	// 开始计时
	var err error
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
	if err != nil {
		return nil
	}

	//创建索引
	col := utils.GetSpiderCol(t.Col, t.Name)
	s, c := database.GetCol(col)
	defer s.Close()
	_ = c.EnsureIndex(mgo.Index{
		Key: []string{"task_id"},
	})
	err = t.Sync()
	if err != nil {
		return nil
	}
	t.RLock()
	defer t.RUnlock()
	// 执行Shell命令
	return services.ExecuteShellCmd(cmd, cwd, task, *t.Spider, user)
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
