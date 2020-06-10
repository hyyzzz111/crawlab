package uploader

import (
	"crawlab/services/spider_service"
	"crawlab/utils"
	"github.com/apex/log"
	"github.com/gin-gonic/gin"
	"github.com/h2non/filetype"
	"github.com/h2non/filetype/matchers"
	uuid "github.com/satori/go.uuid"
	"github.com/spf13/viper"
	"os"
	"path/filepath"
	"runtime/debug"
)

type GinUploader struct {
	spiderUploader
}

func (gin *GinUploader) Upload(ctx *gin.Context) (err error) {
	zipfile, err := gin.fetchFileUpload(ctx)
	if err != nil {
		log.Errorf("upload spider error: " + err.Error())
		debug.PrintStack()
		return err
	}

	// 上传爬虫到GridFS
	if err := gin.toGridFs(zipfile); err != nil {
		log.Errorf("upload spider error: " + err.Error())
		debug.PrintStack()
		return err
	}
	return nil
}
func (gin *GinUploader) fetchFileUpload(ctx *gin.Context) (tmpPath string, err error) {
	uploadFile, err := ctx.FormFile("file")
	if err != nil {

		return "", err
	}
	file, err := uploadFile.Open()
	if err != nil {
		return "", err
	}
	defer utils.Close(file)
	head := make([]byte, 261)
	_, err = file.Read(head)
	if err != nil {
		return "", err
	}
	if !filetype.IsType(head, matchers.TypeZip) {
		return "", spider_service.ErrNotZipFile
	}
	tmpPath = viper.GetString("other.tmppath")
	if err = os.MkdirAll(tmpPath, 0755); err != nil {
		log.Errorf("mkdir other.tmppath error: %v", err.Error())
		return
	}
	randomId := uuid.NewV4()
	tmpFilePath := filepath.Join(tmpPath, randomId.String()+".zip")
	if err := os.MkdirAll(tmpFilePath, os.ModePerm); err != nil {
		return "", err
	}

	if err := ctx.SaveUploadedFile(uploadFile, tmpFilePath); err != nil {
		return "", err
	}
	return tmpFilePath, nil
}
