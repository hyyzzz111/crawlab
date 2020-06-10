package spider_service

import (
	"bufio"
	"crawlab/constants"
	"crawlab/database"
	"crawlab/model"
	"crawlab/services/context"
	log2 "crawlab/services/spider_service/log"
	"crawlab/utils"
	"github.com/apex/log"
	"github.com/spf13/viper"
	"os"
	"os/exec"
	"path"
	"runtime"
	"runtime/debug"
	"strings"
	"sync"
	"syscall"
	"time"
)

// 执行shell命令
func ExecuteShellCmd(ctx context.Context, cmdStr string, cwd string, t model.Task, s model.Spider, u model.User) (err error) {
	log.Infof("cwd: %s", cwd)
	log.Infof("cmd: %s", cmdStr)

	// 生成执行命令
	var cmd *exec.Cmd
	if runtime.GOOS == constants.Windows {
		cmd = exec.Command("cmd", "/C", cmdStr)
	} else {
		cmd = exec.Command("sh", "-c", cmdStr)
	}
	// expire duration (in seconds)
	expireDuration := u.Setting.LogExpireDuration
	if expireDuration == 0 {
		// by default not expire
		expireDuration = constants.Infinite
	}
	session, coll := database.GetCol("logs")
	defer session.Close()
	logInput := log2.NewMongoInput(coll, t.Id, time.Duration(expireDuration))
	// 工作目录
	cmd.Dir = cwd
	cmd.Stdout = logInput
	cmd.Stderr = logInput

	// 环境变量配置
	envs := s.Envs
	if s.Type == constants.Configurable {
		// 数据库配置
		envs = append(envs, model.Env{Name: "CRAWLAB_MONGO_HOST", Value: viper.GetString("mongo.host")})
		envs = append(envs, model.Env{Name: "CRAWLAB_MONGO_PORT", Value: viper.GetString("mongo.port")})
		envs = append(envs, model.Env{Name: "CRAWLAB_MONGO_DB", Value: viper.GetString("mongo.db")})
		envs = append(envs, model.Env{Name: "CRAWLAB_MONGO_USERNAME", Value: viper.GetString("mongo.username")})
		envs = append(envs, model.Env{Name: "CRAWLAB_MONGO_PASSWORD", Value: viper.GetString("mongo.password")})
		envs = append(envs, model.Env{Name: "CRAWLAB_MONGO_AUTHSOURCE", Value: viper.GetString("mongo.authSource")})

		// 设置配置
		for envName, envValue := range s.Config.Settings {
			envs = append(envs, model.Env{Name: "CRAWLAB_SETTING_" + envName, Value: envValue})
		}
	}
	cmd = SetEnv(cmd, envs, t, s)

	//// 起一个goroutine来监控进程
	//ch := utils.TaskExecChanMap.ChanBlocked(t.Id)
	//
	//go FinishOrCancelTask(ch, cmd, s, t)

	// kill的时候，可以kill所有的子进程
	if runtime.GOOS != constants.Windows {
		cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	}

	// 启动进程
	if err := StartTaskProcess(cmd, t); err != nil {
		return err
	}

	// 同步等待进程完成
	if err := WaitTaskProcess(cmd, t, s); err != nil {
		return err
	}
	ch <- constants.TaskFinish
	return nil
}

// 设置环境变量
func SetEnv(cmd *exec.Cmd, envs []model.Env, task model.Task, spider model.Spider) *exec.Cmd {
	// 默认把Node.js的全局node_modules加入环境变量
	envPath := os.Getenv("PATH")
	homePath := os.Getenv("HOME")
	nodeVersion := "v10.19.0"
	nodePath := path.Join(homePath, ".nvm/versions/node", nodeVersion, "lib/node_modules")
	if !strings.Contains(envPath, nodePath) {
		_ = os.Setenv("PATH", nodePath+":"+envPath)
	}
	_ = os.Setenv("NODE_PATH", nodePath)

	// default results collection
	col := utils.GetSpiderCol(spider.Col, spider.Name)

	// 默认环境变量
	cmd.Env = append(os.Environ(), "CRAWLAB_TASK_ID="+task.Id)
	cmd.Env = append(cmd.Env, "CRAWLAB_COLLECTION="+col)
	cmd.Env = append(cmd.Env, "CRAWLAB_MONGO_HOST="+viper.GetString("mongo.host"))
	cmd.Env = append(cmd.Env, "CRAWLAB_MONGO_PORT="+viper.GetString("mongo.port"))
	if viper.GetString("mongo.db") != "" {
		cmd.Env = append(cmd.Env, "CRAWLAB_MONGO_DB="+viper.GetString("mongo.db"))
	}
	if viper.GetString("mongo.username") != "" {
		cmd.Env = append(cmd.Env, "CRAWLAB_MONGO_USERNAME="+viper.GetString("mongo.username"))
	}
	if viper.GetString("mongo.password") != "" {
		cmd.Env = append(cmd.Env, "CRAWLAB_MONGO_PASSWORD="+viper.GetString("mongo.password"))
	}
	if viper.GetString("mongo.authSource") != "" {
		cmd.Env = append(cmd.Env, "CRAWLAB_MONGO_AUTHSOURCE="+viper.GetString("mongo.authSource"))
	}
	cmd.Env = append(cmd.Env, "PYTHONUNBUFFERED=0")
	cmd.Env = append(cmd.Env, "PYTHONIOENCODING=utf-8")
	cmd.Env = append(cmd.Env, "TZ=Asia/Shanghai")
	cmd.Env = append(cmd.Env, "CRAWLAB_DEDUP_FIELD="+spider.DedupField)
	cmd.Env = append(cmd.Env, "CRAWLAB_DEDUP_METHOD="+spider.DedupMethod)
	if spider.IsDedup {
		cmd.Env = append(cmd.Env, "CRAWLAB_IS_DEDUP=1")
	} else {
		cmd.Env = append(cmd.Env, "CRAWLAB_IS_DEDUP=0")
	}

	//任务环境变量
	for _, env := range envs {
		cmd.Env = append(cmd.Env, env.Name+"="+env.Value)
	}

	// 全局环境变量
	variables := model.GetVariableList()
	for _, variable := range variables {
		cmd.Env = append(cmd.Env, variable.Key+"="+variable.Value)
	}
	return cmd
}
func StartTaskProcess(cmd *exec.Cmd, t model.Task) error {
	if err := cmd.Start(); err != nil {
		log.Errorf("start spider error:{}", err.Error())
		debug.PrintStack()

		t.Error = "start task error: " + err.Error()
		t.Status = constants.StatusError
		t.FinishTs = time.Now()
		_ = t.Save()
		return err
	}
	return nil
}

func WaitTaskProcess(cmd *exec.Cmd, t model.Task, s model.Spider) error {
	if err := cmd.Wait(); err != nil {
		log.Errorf("wait process finish error: %s", err.Error())
		debug.PrintStack()

		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode := exitError.ExitCode()
			log.Errorf("exit error, exit code: %d", exitCode)

			// 非kill 的错误类型
			if exitCode != -1 {
				// 非手动kill保存为错误状态
				t.Error = err.Error()
				t.FinishTs = time.Now()
				t.Status = constants.StatusError
				_ = t.Save()

				FinishUpTask(s, t)
			}
		}

		return err
	}

	return nil
}
