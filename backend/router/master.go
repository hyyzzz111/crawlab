package router

import (
	"crawlab/middlewares"
	"crawlab/routes"
	"fmt"
	"github.com/gin-gonic/gin"
)

func RegisterMasterRoutes(app *gin.Engine) (err error) {
	app.Use(middlewares.CORSMiddleware())
	anonymousAccesses := app.Group("/")
	anonymousAccesses.Use(middlewares.TryGetAuthorizationMiddleware())
	{
		anonymousAccesses.GET("/tasks/:id/results/download", routes.DownloadTaskResultsCsv) // 下载任务结果

		anonymousAccesses.GET("/system/config", routes.GetSettings)
		anonymousAccesses.GET("/system/routing", routes.GetRoutings)
		anonymousAccesses.PUT("/system/register_admin", routes.PutAdminUser)

		//邀请注册
		anonymousAccesses.POST("/invitation/:id/validate", routes.ValidateInvitationURL)     //邀请列表
		anonymousAccesses.POST("/invitation/:id/register", routes.RegisterFromInvitationURL) //邀请列表
		anonymousAccesses.POST("/login", routes.Login)
	}
	authAccesses := app.Group("/")
	authAccesses.Use(middlewares.AuthorizationMiddleware())
	routeManger := DefaultManager()
	routeManger.RegisterRouterGroup("anonymous", anonymousAccesses)
	routeManger.RegisterRouterGroup("auth", authAccesses)

	{
		nodeRoutes := []Route{
			{Router: "auth", Method: GET, Path: "/nodes", I18n: "api.get.node_list", Alias: "get_node_list", Handler: routes.GetNodeList},                        // 节点列表
			{Router: "auth", Method: GET, Path: "/nodes/:id", I18n: "api.get.node_detail", Alias: "get_node_detail", Handler: routes.GetNode},                    // 节点详情
			{Router: "auth", Method: POST, Path: "/nodes/:id", I18n: "api.post.node_detail", Alias: "post_node_detail", Handler: routes.PostNode},                // 修改节点
			{Router: "auth", Method: GET, Path: "/nodes/:id/tasks", I18n: "api.get.node_id_tasks", Alias: "get_node_id_tasks", Handler: routes.GetNodeTaskList},  // 节点任务列表
			{Router: "auth", Method: GET, Path: "/nodes/:id/system", I18n: "api.get.node_id_system", Alias: "get_node_id_system", Handler: routes.GetSystemInfo}, // 节点系统信息
			{Router: "auth", Method: DELETE, Path: "/nodes/:id", I18n: "api.delete.nodes_id", Alias: "delete_nodes_id", Handler: routes.DeleteNode},              // 删除节点
		}
		routeManger.RegisterRoutes(nodeRoutes, "nodes", "nodes")
	}
	{
		spiderRoutes := []Route{
			// 爬虫
			{Router: "auth", Method: GET, Path: "/spiders", I18n: "api.get.spiders", Alias: "get_spiders", Handler: routes.GetSpiderList},                                  // 爬虫列表
			{Router: "auth", Method: GET, Path: "/spiders/:id", I18n: "api.get.spiders_id", Alias: "get_spiders_id", Handler: routes.GetSpider},                            // 爬虫详情
			{Router: "auth", Method: POST, Path: "/spiders", I18n: "api.post.spiders", Alias: "post_spiders", Handler: routes.GetSpider},                                   // 上传爬虫
			{Router: "auth", Method: POST, Path: "/spiders/:id", I18n: "api.post.spiders_id", Alias: "post_spiders_id", Handler: routes.GetSpider},                         // 修改爬虫
			{Router: "auth", Method: POST, Path: "/spiders/:id/publish", I18n: "api.post.spiders_id_publish", Alias: "post_spiders_id_publish", Handler: routes.GetSpider}, // 发布爬虫
			{Router: "auth", Method: DELETE, Path: "/spiders/:id", I18n: "api.delete.spiders_id", Alias: "delete_spiders_id", Handler: routes.DeleteSpider},                // 删除爬虫
			{Router: "auth", Method: GET, Path: "/spiders/:id/tasks", I18n: "api.get.spiders_id_tasks", Alias: "get_spiders_id_tasks", Handler: routes.GetSpiderTasks},     // 爬虫任务列表
			{Router: "auth", Method: GET, Path: "/spiders/:id/file", I18n: "api.get.spiders_id_file", Alias: "get_spiders_id_file", Handler: routes.GetSpiderFile},         // 爬虫文件读取
			{Router: "auth", Method: POST, Path: "/spiders/:id/file", I18n: "api.post.spider_id_file", Alias: "post_spiders_id_file", Handler: routes.PostSpiderFile},      // 爬虫目录写入
			{Router: "auth", Method: GET, Path: "/spiders/:id/dir", I18n: "api.get.spiders_id_file", Alias: "get_spiders_id_file", Handler: routes.GetSpiderDir},           // 爬虫目录
			{Router: "auth", Method: GET, Path: "/spiders/:id/stats", I18n: "api.get.spiders_id_stats", Alias: "get_spiders_id_stats", Handler: routes.GetSpiderStats},     // 爬虫统计数据
		}
		routeManger.RegisterRoutes(spiderRoutes, "spiders", "spiders")
	}
	{
		tasksRoutes := []Route{
			{Router: "auth", Method: GET, Path: "/tasks", I18n: "api.get.tasks", Alias: "get_tasks", Handler: routes.GetTaskList},                                    // 任务列表
			{Router: "auth", Method: GET, Path: "/tasks/:id", I18n: "api.get.tasks_id", Alias: "get_tasks_id", Handler: routes.GetTask},                              // 任务详情
			{Router: "auth", Method: PUT, Path: "/tasks", I18n: "api.put.tasks", Alias: "put_tasks", Handler: routes.PutTask},                                        // 派发任务
			{Router: "auth", Method: DELETE, Path: "/tasks/:id", I18n: "api.delete.tasks_id", Alias: "delete_tasks_id", Handler: routes.DeleteTask},                  // 删除任务
			{Router: "auth", Method: POST, Path: "/tasks/:id/cancel", I18n: "api.post.tasks_id_cancel", Alias: "post_tasks_id_cancel", Handler: routes.CancelTask},   // 取消任务
			{Router: "auth", Method: GET, Path: "/tasks/:id/log", I18n: "api.get.tasks_id_log", Alias: "get_tasks_id_log", Handler: routes.GetTaskLog},               // 任务日志
			{Router: "auth", Method: GET, Path: "/tasks/:id/results", I18n: "api.get.tasks_id_result", Alias: "get_tasks_id_result", Handler: routes.GetTaskResults}, // 任务结果
		}
		routeManger.RegisterRoutes(tasksRoutes, "tasks", "tasks")
	}
	{
		scheduleRoutes := []Route{
			//定时任务
			{Router: "auth", Method: GET, Path: "/schedules", I18n: "api.get.schedules", Alias: "get_schedules", Handler: routes.GetScheduleList},                   // 定时任务列表
			{Router: "auth", Method: GET, Path: "/schedules/:id", I18n: "api.get.schedules_id", Alias: "get_result_id", Handler: routes.GetSchedule},                // 定时任务详情
			{Router: "auth", Method: PUT, Path: "/schedules", I18n: "api.put.schedules", Alias: "put_schedules", Handler: routes.PutSchedule},                       // 创建定时任务
			{Router: "auth", Method: POST, Path: "/schedules/:id", I18n: "api.post.schedules_id", Alias: "post_schedules", Handler: routes.PostSchedule},            // 修改定时任务
			{Router: "auth", Method: DELETE, Path: "/schedules/:id", I18n: "api.delete.schedules_id", Alias: "delete_schedules_id", Handler: routes.DeleteSchedule}, // 删除定时任务
		}
		routeManger.RegisterRoutes(scheduleRoutes, "schedules", "schedules")
	}
	{
		stateRoutes := []Route{
			// 统计数据
			{Router: "auth", Method: GET, Path: "/stats/home", I18n: "api.get.stats_home", Alias: "get_stats_home", Handler: routes.GetHomeStats}, // 首页统计数据
		}
		routeManger.RegisterRoutes(stateRoutes, "stats", "stats")
	}

	{
		userRoutes := []Route{
			// 用户
			{Router: "auth", Method: GET, Path: "users", I18n: "api.get.users", Alias: "get_users", Handler: routes.GetUserList},                   // 用户列表
			{Router: "auth", Method: GET, Path: "/users/:id", I18n: "api.get.users_id", Alias: "get_users_id", Handler: routes.GetUser},            // 用户详情
			{Router: "auth", Method: PUT, Path: "/users", I18n: "api.put.users", Alias: "put_users", Handler: routes.PutUser},                      // 添加用户
			{Router: "auth", Method: POST, Path: "/users/:id", I18n: "api.post.users_id", Alias: "post_users_id", Handler: routes.PostUser},        // 更改用户
			{Router: "auth", Method: DELETE, Path: "/users/:id", I18n: "api.delete.users_id", Alias: "delete_user_id", Handler: routes.DeleteUser}, // 删除用户
			{Router: "auth", Method: GET, Path: "/me", I18n: "api.get.me", Alias: "get_me", Handler: routes.GetMe},
		}
		routeManger.RegisterRoutes(userRoutes, "users", "users")

	}
	{
		invitationRoutes := []Route{
			//邀请注册
			{Router: "auth", Method: GET, Path: "/invitations", I18n: "api.get.invitation_list", Alias: "get_invitation_list", Handler: routes.GetInvitationURLList},  //邀请列表
			{Router: "auth", Method: PUT, Path: "/invitation", I18n: "api.put.invitation", Alias: "put_invitation", Handler: routes.GenerateInvitationURL},            //生成邀请链接
			{Router: "auth", Method: POST, Path: "/invitation/:id", I18n: "api.post.invitation_id", Alias: "post_invitation_id", Handler: routes.UpdateInvitationURL}, //更新邀请链接配置
		}
		routeManger.RegisterRoutes(invitationRoutes, "invitations", "invitations")

	}
	err = routeManger.SetUp()
	fmt.Println(err)
	if err != nil {
		return err
	}
	return nil
	// 路由
	// 节点
	//authAccesses.GET("/nodes", routes.GetNodeList)               // 节点列表
	//authAccesses.GET("/nodes/:id", routes.GetNode)               // 节点详情
	//authAccesses.POST("/nodes/:id", routes.PostNode)             // 修改节点
	//authAccesses.GET("/nodes/:id/tasks", routes.GetNodeTaskList) // 节点任务列表
	//authAccesses.GET("/nodes/:id/system", routes.GetSystemInfo)  // 节点任务列表
	//authAccesses.DELETE("/nodes/:id", routes.DeleteNode)         // 删除节点
	// 爬虫
	//authAccesses.GET("/spiders", routes.GetSpiderList)              // 爬虫列表
	//authAccesses.GET("/spiders/:id", routes.GetSpider)              // 爬虫详情
	//authAccesses.POST("/spiders", routes.PutSpider)                 // 上传爬虫
	//authAccesses.POST("/spiders/:id", routes.PostSpider)            // 修改爬虫
	//authAccesses.POST("/spiders/:id/publish", routes.PublishSpider) // 发布爬虫
	//authAccesses.DELETE("/spiders/:id", routes.DeleteSpider)        // 删除爬虫
	//authAccesses.GET("/spiders/:id/tasks", routes.GetSpiderTasks)   // 爬虫任务列表
	//authAccesses.GET("/spiders/:id/file", routes.GetSpiderFile)     // 爬虫文件读取
	//authAccesses.POST("/spiders/:id/file", routes.PostSpiderFile)   // 爬虫目录写入
	//authAccesses.GET("/spiders/:id/dir", routes.GetSpiderDir)       // 爬虫目录
	//authAccesses.GET("/spiders/:id/stats", routes.GetSpiderStats)   // 爬虫统计数据
	// 任务
	//authAccesses.GET("/tasks", routes.GetTaskList)                // 任务列表
	//authAccesses.GET("/tasks/:id", routes.GetTask)                // 任务详情
	//authAccesses.PUT("/tasks", routes.PutTask)                    // 派发任务
	//authAccesses.DELETE("/tasks/:id", routes.DeleteTask)          // 删除任务
	//authAccesses.POST("/tasks/:id/cancel", routes.CancelTask)     // 取消任务
	//authAccesses.GET("/tasks/:id/log", routes.GetTaskLog)         // 任务日志
	//authAccesses.GET("/tasks/:id/results", routes.GetTaskResults) // 任务结果
	// 定时任务
	//authAccesses.GET("/schedules", routes.GetScheduleList)       // 定时任务列表
	//authAccesses.GET("/schedules/:id", routes.GetSchedule)       // 定时任务详情
	//authAccesses.PUT("/schedules", routes.PutSchedule)           // 创建定时任务
	//authAccesses.POST("/schedules/:id", routes.PostSchedule)     // 修改定时任务
	//authAccesses.DELETE("/schedules/:id", routes.DeleteSchedule) // 删除定时任务
	//// 统计数据
	//authAccesses.GET("/stats/home", routes.GetHomeStats) // 首页统计数据
	//// 用户
	//authAccesses.GET("/users", routes.GetUserList)       // 用户列表
	//authAccesses.GET("/users/:id", routes.GetUser)       // 用户详情
	//authAccesses.PUT("/users", routes.PutUser)           // 添加用户
	//authAccesses.POST("/users/:id", routes.PostUser)     // 更改用户
	//authAccesses.DELETE("/users/:id", routes.DeleteUser) // 删除用户
	//authAccesses.POST("/login", routes.Login)            // 用户登录
	//authAccesses.GET("/me", routes.GetMe)                // 获取自己账户
	////邀请注册
	//authAccesses.GET("/invitations", routes.GetInvitationURLList)    //邀请列表
	//authAccesses.PUT("/invitation", routes.GenerateInvitationURL)    //生成邀请链接
	//authAccesses.POST("/invitation/:id", routes.UpdateInvitationURL) //更新邀请链接配置

}
