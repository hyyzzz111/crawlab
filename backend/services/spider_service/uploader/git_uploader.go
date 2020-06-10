package uploader

import (
	"github.com/apex/log"
	"github.com/cenkalti/backoff/v4"
	uuid "github.com/satori/go.uuid"
	"github.com/spf13/viper"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/http"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/ssh"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"runtime/debug"
	"strings"
	"time"
)

type GitUploader struct {
	spiderUploader
}

func (git *GitUploader) Update() (err error) {
	zipfile, err := git.fetchGit()
	if err != nil {
		log.Errorf("upload spider error: " + err.Error())
		debug.PrintStack()
		return err
	}

	// 上传爬虫到GridFS
	if err := git.toGridFs(zipfile); err != nil {
		log.Errorf("upload spider error: " + err.Error())
		debug.PrintStack()
		return err
	}
	return nil
}
func (s *GitUploader) fetchGit() (tmpFilePath string, err error) {
	randomId := uuid.NewV4()
	tmpPath := viper.GetString("other.tmppath")
	tmpFilePath = filepath.Join(tmpPath, randomId.String())

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
	defer func() {
		if err != nil {

			s.GitSyncError = err.Error()
			_ = backoff.Retry(func() error {
				return s.Save()

			}, backoff.NewConstantBackOff(1*time.Second))
		}
	}()
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
				return "", err
			}
		}
	}

	rep, err := git.PlainClone(tmpFilePath, false, cloneOption)
	if err != nil {
		return "", err
	}
	ref, err := rep.Reference(currentBranch, true)
	if err != nil {
		return "", err
	}
	if s.Spider.GitHash == ref.Hash().String() {
		return "", nil
	}

	filename := s.Name + "." + randomId.String()
	return s.zipFile(filename, tmpFilePath)
}
