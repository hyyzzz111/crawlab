package master

import (
	"crawlab/app/master/config"
	"crawlab/middlewares"
	mr "crawlab/pkg/core/registry/datasource/memory"
	"crawlab/routes"
	"crawlab/runtime"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"net/http"
	"strings"
)

func NewMasterNode(config *config.ApplicationConfig) *http.Server {
	grpcServer := grpc.NewServer()
	wsServer, _ := newWebsocketServer()
	if config.RegistryType == runtime.RegistryMaster {
		mr.BindGRPCServices(grpcServer)
	}
	ginServer := gin.Default()
	//TODO 定制CORS
	ginServer.Use(middlewares.CORSMiddleware())
	installGinAnonymousRoutes(ginServer)
	installGinAuthRoutes(ginServer)
	wsServer.installWebsocketService(ginServer, &config.Server.Websocket)
	return &http.Server{
		Addr: config.HostWithoutProtocol(),
		Handler: http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			if request.ProtoMajor != 2 {
				ginServer.ServeHTTP(writer, request)
				return
			}
			contentType := request.Header.Get("Content-Type")
			if strings.Contains(contentType, "application/grpc") {
				grpcServer.ServeHTTP(writer, request)
			}
		}),
	}
}

func installNodeAPI(g *gin.RouterGroup) {
	g.GET("/nodes", routes.GetNodeList)                            // 节点列表
	g.GET("/nodes/:id", routes.GetNode)                            // 节点详情
	g.POST("/nodes/:id", routes.PostNode)                          // 修改节点
	g.GET("/nodes/:id/tasks", routes.GetNodeTaskList)              // 节点任务列表
	g.GET("/nodes/:id/system", routes.GetSystemInfo)               // 节点任务列表
	g.DELETE("/nodes/:id", routes.DeleteNode)                      // 删除节点
	g.GET("/nodes/:id/langs", routes.GetLangList)                  // 节点语言环境列表
	g.GET("/nodes/:id/deps", routes.GetDepList)                    // 节点第三方依赖列表
	g.GET("/nodes/:id/deps/installed", routes.GetInstalledDepList) // 节点已安装第三方依赖列表
	g.POST("/nodes/:id/deps/install", routes.InstallDep)           // 节点安装依赖
	g.POST("/nodes/:id/deps/uninstall", routes.UninstallDep)       // 节点卸载依赖
	g.POST("/nodes/:id/langs/install", routes.InstallLang)         // 节点安装语言
}
func installSpiderAPI(g *gin.RouterGroup) {
	g.GET("/spiders", routes.GetSpiderList)                                            // 爬虫列表
	g.GET("/spiders/:id", routes.GetSpider)                                            // 爬虫详情
	g.PUT("/spiders", routes.PutSpider)                                                // 添加爬虫
	g.POST("/spiders", routes.UploadSpider)                                            // 上传爬虫
	g.POST("/spiders/:id", routes.PostSpider)                                          // 修改爬虫
	g.POST("/spiders/:id/publish", routes.PublishSpider)                               // 发布爬虫
	g.POST("/spiders/:id/upload", routes.UploadSpiderFromId)                           // 上传爬虫（ID）
	g.DELETE("/spiders", routes.DeleteSelectedSpider)                                  // 删除选择的爬虫
	g.DELETE("/spiders/:id", routes.DeleteSpider)                                      // 删除爬虫
	g.POST("/spiders/:id/copy", routes.CopySpider)                                     // 拷贝爬虫
	g.GET("/spiders/:id/tasks", routes.GetSpiderTasks)                                 // 爬虫任务列表
	g.GET("/spiders/:id/file/tree", routes.GetSpiderFileTree)                          // 爬虫文件目录树读取
	g.GET("/spiders/:id/file", routes.GetSpiderFile)                                   // 爬虫文件读取
	g.POST("/spiders/:id/file", routes.PostSpiderFile)                                 // 爬虫文件更改
	g.PUT("/spiders/:id/file", routes.PutSpiderFile)                                   // 爬虫文件创建
	g.PUT("/spiders/:id/dir", routes.PutSpiderDir)                                     // 爬虫目录创建
	g.DELETE("/spiders/:id/file", routes.DeleteSpiderFile)                             // 爬虫文件删除
	g.POST("/spiders/:id/file/rename", routes.RenameSpiderFile)                        // 爬虫文件重命名
	g.GET("/spiders/:id/dir", routes.GetSpiderDir)                                     // 爬虫目录
	g.GET("/spiders/:id/stats", routes.GetSpiderStats)                                 // 爬虫统计数据
	g.GET("/spiders/:id/schedules", routes.GetSpiderSchedules)                         // 爬虫定时任务
	g.GET("/spiders/:id/scrapy/spiders", routes.GetSpiderScrapySpiders)                // Scrapy 爬虫名称列表
	g.PUT("/spiders/:id/scrapy/spiders", routes.PutSpiderScrapySpiders)                // Scrapy 爬虫创建爬虫
	g.GET("/spiders/:id/scrapy/settings", routes.GetSpiderScrapySettings)              // Scrapy 爬虫设置
	g.POST("/spiders/:id/scrapy/settings", routes.PostSpiderScrapySettings)            // Scrapy 爬虫修改设置
	g.GET("/spiders/:id/scrapy/items", routes.GetSpiderScrapyItems)                    // Scrapy 爬虫 items
	g.POST("/spiders/:id/scrapy/items", routes.PostSpiderScrapyItems)                  // Scrapy 爬虫修改 items
	g.GET("/spiders/:id/scrapy/pipelines", routes.GetSpiderScrapyPipelines)            // Scrapy 爬虫 pipelines
	g.GET("/spiders/:id/scrapy/spider/filepath", routes.GetSpiderScrapySpiderFilepath) // Scrapy 爬虫 pipelines
	g.POST("/spiders/:id/git/sync", routes.PostSpiderSyncGit)                          // 爬虫 Git 同步
	g.POST("/spiders/:id/git/reset", routes.PostSpiderResetGit)                        // 爬虫 Git 重置
	g.POST("/spiders-cancel", routes.CancelSelectedSpider)                             // 停止所选爬虫任务
	g.POST("/spiders-run", routes.RunSelectedSpider)                                   // 运行所选爬虫
}
func installConfigSpiderAPI(g *gin.RouterGroup) {
	g.GET("/config_spiders/:id/config", routes.GetConfigSpiderConfig)           // 获取可配置爬虫配置
	g.POST("/config_spiders/:id/config", routes.PostConfigSpiderConfig)         // 更改可配置爬虫配置
	g.PUT("/config_spiders", routes.PutConfigSpider)                            // 添加可配置爬虫
	g.POST("/config_spiders/:id", routes.PostConfigSpider)                      // 修改可配置爬虫
	g.POST("/config_spiders/:id/upload", routes.UploadConfigSpider)             // 上传可配置爬虫
	g.POST("/config_spiders/:id/spiderfile", routes.PostConfigSpiderSpiderfile) // 上传可配置爬虫
	g.GET("/config_spiders_templates", routes.GetConfigSpiderTemplateList)      // 获取可配置爬虫模版列表
}

func installScheduleAPI(g *gin.RouterGroup) {
	g.GET("/schedules", routes.GetScheduleList)              // 定时任务列表
	g.GET("/schedules/:id", routes.GetSchedule)              // 定时任务详情
	g.PUT("/schedules", routes.PutSchedule)                  // 创建定时任务
	g.POST("/schedules/:id", routes.PostSchedule)            // 修改定时任务
	g.DELETE("/schedules/:id", routes.DeleteSchedule)        // 删除定时任务
	g.POST("/schedules/:id/disable", routes.DisableSchedule) // 禁用定时任务
	g.POST("/schedules/:id/enable", routes.EnableSchedule)   // 启用定时任务
}
func installUserAPI(g *gin.RouterGroup) {
	g.GET("/users", routes.GetUserList)       // 用户列表
	g.GET("/users/:id", routes.GetUser)       // 用户详情
	g.POST("/users/:id", routes.PostUser)     // 更改用户
	g.DELETE("/users/:id", routes.DeleteUser) // 删除用户
	g.PUT("/users-add", routes.PutUser)       // 添加用户
	g.GET("/me", routes.GetMe)                // 获取自己账户
	g.POST("/me", routes.PostMe)              // 修改自己账户
}
func installSystemAPI(g *gin.RouterGroup) {
	g.GET("/system/deps/:lang", routes.GetAllDepList)             // 节点所有第三方依赖列表
	g.GET("/system/deps/:lang/:dep_name/json", routes.GetDepJson) // 节点第三方依赖JSON
}
func installGlobalEnvAPI(g *gin.RouterGroup) {
	g.GET("/variables", routes.GetVariableList)      // 列表
	g.PUT("/variable", routes.PutVariable)           // 新增
	g.POST("/variable/:id", routes.PostVariable)     // 修改
	g.DELETE("/variable/:id", routes.DeleteVariable) // 删除
}
func installProjectAPI(g *gin.RouterGroup) {
	g.GET("/projects", routes.GetProjectList)       // 列表
	g.GET("/projects/tags", routes.GetProjectTags)  // 项目标签
	g.PUT("/projects", routes.PutProject)           // 修改
	g.POST("/projects/:id", routes.PostProject)     // 新增
	g.DELETE("/projects/:id", routes.DeleteProject) // 删除
}

func installChallengeAPI(g *gin.RouterGroup) {
	g.GET("/challenges", routes.GetChallengeList)          // 挑战列表
	g.POST("/challenges-check", routes.CheckChallengeList) // 检查挑战列表
}
func installOtherAPI(g *gin.RouterGroup) {
	g.PUT("/actions", routes.PutAction) // 新增操作

	// 统计数据
	g.GET("/stats/home", routes.GetHomeStats) // 首页统计数据
	// 文件
	g.GET("/file", routes.GetFile) // 获取文件
	// Git
	g.GET("/git/branches", routes.GetGitRemoteBranches) // 获取 Git 分支
	g.GET("/git/public-key", routes.GetGitSshPublicKey) // 获取 SSH 公钥
	g.GET("/git/commits", routes.GetGitCommits)         // 获取 Git Commits
	g.POST("/git/checkout", routes.PostGitCheckout)     // 获取 Git Commits
}

func installGinAuthRoutes(g *gin.Engine) {
	group := g.Group("/")
	installNodeAPI(group)
	installSpiderAPI(group)
	installConfigSpiderAPI(group)
	installScheduleAPI(group)
	installUserAPI(group)
	installProjectAPI(group)
	installGlobalEnvAPI(group)
	installChallengeAPI(group)
	installSystemAPI(group)
	installOtherAPI(group)
}

func installGinAnonymousRoutes(g *gin.Engine) {
	group := g.Group("/", middlewares.AuthorizationMiddleware())
	group.POST("/login", routes.Login)       // 用户登录
	group.PUT("/users", routes.PutUser)      // 添加用户
	group.GET("/setting", routes.GetSetting) // 获取配置信息
	// release版本
	group.GET("/version", routes.GetVersion)               // 获取发布的版本
	group.GET("/releases/latest", routes.GetLatestRelease) // 获取最近发布的版本
	// 文档
	group.GET("/docs", routes.GetDocs) // 获取文档数据
}
