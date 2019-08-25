package main

import (
	"crawlab/config"
	"crawlab/database"
	"crawlab/model"
	"crawlab/router"
	"crawlab/routes"
	"crawlab/services"
	"github.com/apex/log"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"runtime/debug"
	"sync/atomic"
)

func main() {
	app := gin.Default()

	// 初始化配置
	if err := config.InitConfig(""); err != nil {
		log.Error("init config error:" + err.Error())
		panic(err)
	}
	log.Info("初始化配置成功")

	// 初始化日志设置
	logLevel := viper.GetString("log.level")
	if logLevel != "" {
		log.SetLevelFromString(logLevel)
	}
	log.Info("初始化日志设置成功")

	// 初始化Mongodb数据库
	if err := database.InitMongo(); err != nil {
		log.Error("init mongodb error:" + err.Error())
		debug.PrintStack()
		panic(err)
	}
	log.Info("初始化Mongodb数据库成功")

	// 初始化Redis数据库
	if err := database.InitRedis(); err != nil {
		log.Error("init redis error:" + err.Error())
		debug.PrintStack()
		panic(err)
	}
	log.Info("初始化Redis数据库成功")

	if services.IsMaster() {
		// 初始化定时任务
		if err := services.InitScheduler(); err != nil {
			log.Error("init scheduler error:" + err.Error())
			debug.PrintStack()
			panic(err)
		}
		log.Info("初始化定时任务成功")
	}

	// 初始化任务执行器
	if err := services.InitTaskExecutor(); err != nil {
		log.Error("init task executor error:" + err.Error())
		debug.PrintStack()
		panic(err)
	}
	log.Info("初始化任务执行器成功")

	// 初始化节点服务
	if err := services.InitNodeService(); err != nil {
		log.Error("init node service error:" + err.Error())
		panic(err)
	}
	log.Info("初始化节点配置成功")

	// 初始化爬虫服务
	if err := services.InitSpiderService(); err != nil {
		log.Error("init spider service error:" + err.Error())
		debug.PrintStack()
		panic(err)
	}
	log.Info("初始化爬虫服务成功")

	//// 初始化用户服务
	//if err := services.InitUserService(); err != nil {
	//	log.Error("init user service error:" + err.Error())
	//	debug.PrintStack()
	//	panic(err)
	//}
	//log.Info("初始化用户服务成功")

	// 以下为主节点服务
	if services.IsMaster() {

		exists, err := model.HasAdminUser()
		if err != nil {
			panic(err)
		}
		var flag int32
		if exists {
			flag = 1
		} else {
			_ = services.InitUserService()
		}
		atomic.StoreInt32(&routes.HasAdminAccount, flag)
		router.RegisterMasterRoutes(app)
	} else {
		router.RegisterWorkerRoutes(app)
	}

	// 路由ping
	app.GET("/ping", routes.Ping)
	//fmt.Println(app.Routes())
	// 运行服务器
	host := viper.GetString("server.host")
	port := viper.GetString("server.port")
	if err := app.Run(host + ":" + port); err != nil {
		log.Error("run server error:" + err.Error())
		panic(err)
	}
}
