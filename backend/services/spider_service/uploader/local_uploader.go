package uploader

import (
	"crawlab/model"
	"github.com/apex/log"
	uuid "github.com/satori/go.uuid"
	"runtime/debug"
)

type LocalFileUploader struct {
	spiderUploader
}

func NewLocalFileUploader(sp *model.Spider) *LocalFileUploader {
	return &LocalFileUploader{spiderUploader{Spider: sp}}
}
func (local *LocalFileUploader) fetchLocalFile(localPath string) (tmpZipPath string, err error) {
	randomId := uuid.NewV4()
	filename := local.Name + "." + randomId.String()
	return local.zipFile(filename, localPath)
}
func (local *LocalFileUploader) Upload(localPath string) (err error) {
	zipfile, err := local.fetchLocalFile(localPath)
	if err != nil {
		log.Errorf("upload spider error: " + err.Error())
		debug.PrintStack()
		return err
	}

	// 上传爬虫到GridFS
	if err := local.toGridFs(zipfile); err != nil {
		log.Errorf("upload spider error: " + err.Error())
		debug.PrintStack()
		return err
	}
	return nil
}
