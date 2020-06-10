package uploader

import (
	"crawlab/database"
	"crawlab/model"
	"crawlab/services/spider_service"
	"crawlab/utils"
	"github.com/apex/log"
	"github.com/cenkalti/backoff/v4"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/spf13/viper"
	"os"
	"path/filepath"
	"runtime/debug"
	"time"
)

type spiderUploader struct {
	*model.Spider
}

func (s *spiderUploader) toGridFs(tmpZipPath string) (err error) {

	//根据zip生成md5
	md5Str, err := utils.FileMd5(tmpZipPath)
	if err != nil {
		return err
	}
	// 获取 GridFS 实例
	session, gf := database.GetGridFs("files")
	defer session.Close()
	filename := s.Name + ".zip"

	// 判断文件是否已经存在
	var gfFile *model.GridFs
	bp := backoff.NewConstantBackOff(time.Second * 3)
	backoff.Retry(func() error {
		err := gf.Find(bson.M{"filename": filename}).One(&gfFile)
		if err == nil || err == mgo.ErrNotFound {
			s.Spider.UploadingGridFsFile = true
			s.Spider.LastUploadedGridFsFileTime = time.Now()
			return s.Spider.Save()
		}
		return err
	}, bp)
	defer func() {
		backoff.Retry(func() error {
			s.Spider.UploadingGridFsFile = false
			return s.Spider.Save()
		}, bp)
	}()
	if gfFile != nil {
		if md5Str == gfFile.Md5 {
			return nil
		}
		if err := gf.RemoveId(gfFile.Id); err != nil && err != mgo.ErrNotFound {
			log.Errorf(err.Error())
			debug.PrintStack()
			return err
		}
	}
	// 上传到GridFs
	fid, err := spider_service.UploadToGridFs(filename, tmpZipPath)
	if err != nil {
		log.Errorf("upload to grid fs error: %s", err.Error())
		return err
	}

	// 保存爬虫 FileId
	s.FileId = fid

	return s.Save()
}
func (s *spiderUploader) zipFile(filename string, srcPath string) (zipFilePath string, err error) {

	// 打包为 zip 文件
	files, err := utils.GetFilesFromDir(srcPath)
	if err != nil {
		return "", err
	}

	tmpPath := viper.GetString("other.tmppath")
	zipFilePath = filepath.Join(tmpPath, filename+".zip")
	if err = os.MkdirAll(tmpPath, 0777); err != nil {
		log.Errorf("mkdir other.tmppath error: %v", err.Error())
		return "", nil
	}
	if err := utils.Compress(files, zipFilePath); err != nil {
		return "", err
	}
	return zipFilePath, nil
}
