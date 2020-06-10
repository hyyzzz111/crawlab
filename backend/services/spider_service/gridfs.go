package spider_service

import (
	"crawlab/database"
	"crawlab/utils"
	"fmt"
	"github.com/apex/log"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"os"
	"runtime/debug"
)

// 上传zip文件到GridFS
func UploadToGridFs(fileName string, filePath string) (fid bson.ObjectId, err error) {
	fid = ""

	// 获取MongoDB GridFS连接
	s, gf := database.GetGridFs("files")
	defer s.Close()

	// 创建一个新GridFS文件
	f, err := gf.Create(fileName)
	if err != nil {
		log.Errorf("create file error: " + err.Error())
		debug.PrintStack()
		return
	}

	// 分片读取爬虫zip文件
	err = ReadFileByStep(filePath, WriteToGridFS, f)
	if err != nil {
		log.Errorf("read file by step error: " + err.Error())
		debug.PrintStack()
		return "", err
	}

	// 删除zip文件
	if err = os.Remove(filePath); err != nil {
		log.Errorf("remove file error: " + err.Error())
		debug.PrintStack()
		return
	}

	// 关闭文件，提交写入
	if err = f.Close(); err != nil {
		log.Errorf("close file error: " + err.Error())
		debug.PrintStack()
		return "", err
	}

	// 文件ID
	fid = f.Id().(bson.ObjectId)

	return fid, nil
}

// 写入grid fs
func WriteToGridFS(content []byte, f *mgo.GridFile) {
	if _, err := f.Write(content); err != nil {
		debug.PrintStack()
		return
	}
}

//分片读取大文件
func ReadFileByStep(filePath string, handle func([]byte, *mgo.GridFile), fileCreate *mgo.GridFile) error {
	f, err := os.OpenFile(filePath, os.O_RDONLY, 0777)
	if err != nil {
		log.Infof("can't opened this file")
		return err
	}
	defer utils.Close(f)
	s := make([]byte, 4096)
	for {
		switch nr, err := f.Read(s[:]); true {
		case nr < 0:
			_, _ = fmt.Fprintf(os.Stderr, "cat: error reading: %s\n", err.Error())
			debug.PrintStack()
		case nr == 0: // EOF
			return nil
		case nr > 0:
			handle(s[0:nr], fileCreate)
		}
	}
}
