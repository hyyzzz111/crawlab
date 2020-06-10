package notification

import (
	"crawlab/constants"
	"crawlab/model"
	"crawlab/services/notification"
	"crawlab/utils"
	"fmt"
	"github.com/apex/log"
	"github.com/imroc/req"
	"net/http"
	"runtime/debug"
)

func GetTaskEmailMarkdownContent(t model.Task, s model.Spider) string {
	n, _ := model.GetNode(t.NodeId)
	errMsg := ""
	statusMsg := fmt.Sprintf(`<span style="color:green">%s</span>`, t.Status)
	if t.Status == constants.StatusError {
		errMsg = " with errors"
		statusMsg = fmt.Sprintf(`<span style="color:red">%s</span>`, t.Status)
	}
	return fmt.Sprintf(`
Your task has finished%s. Please find the task info below.

 | 
--: | :--
**Task ID:** | %s
**Task Status:** | %s
**Task Param:** | %s
**Spider ID:** | %s
**Spider Name:** | %s
**Node:** | %s
**Create Time:** | %s
**Start Time:** | %s
**Finish Time:** | %s
**Wait Duration:** | %.0f sec
**Runtime Duration:** | %.0f sec
**Total Duration:** | %.0f sec
**Number of Results:** | %d
**Error:** | <span style="color:red">%s</span>

Please login to Crawlab to view the details.
`,
		errMsg,
		t.Id,
		statusMsg,
		t.Param,
		s.Id.Hex(),
		s.Name,
		n.Name,
		utils.GetLocalTimeString(t.CreateTs),
		utils.GetLocalTimeString(t.StartTs),
		utils.GetLocalTimeString(t.FinishTs),
		t.WaitDuration,
		t.RuntimeDuration,
		t.TotalDuration,
		t.ResultCount,
		t.Error,
	)
}
func SendTaskEmail(u model.User, t model.Task, s model.Spider) {
	statusMsg := "has finished"
	if t.Status == constants.StatusError {
		statusMsg = "has an error"
	}
	title := fmt.Sprintf("[Crawlab] Task for \"%s\" %s", s.Name, statusMsg)
	if err := notification.SendMail(
		u.Email,
		u.Username,
		title,
		GetTaskEmailMarkdownContent(t, s),
	); err != nil {
		log.Errorf("mail error: " + err.Error())
		debug.PrintStack()
	}
}
func GetTaskMarkdownContent(t model.Task, s model.Spider) string {
	n, _ := model.GetNode(t.NodeId)
	errMsg := ""
	errLog := "-"
	statusMsg := fmt.Sprintf(`<font color="#00FF00">%s</font>`, t.Status)
	if t.Status == constants.StatusError {
		errMsg = `（有错误）`
		errLog = fmt.Sprintf(`<font color="#FF0000">%s</font>`, t.Error)
		statusMsg = fmt.Sprintf(`<font color="#FF0000">%s</font>`, t.Status)
	}
	return fmt.Sprintf(`
您的任务已完成%s，请查看任务信息如下。

> **任务ID:** %s  
> **任务状态:** %s  
> **任务参数:** %s  
> **爬虫ID:** %s  
> **爬虫名称:** %s  
> **节点:** %s  
> **创建时间:** %s  
> **开始时间:** %s  
> **完成时间:** %s  
> **等待时间:** %.0f秒   
> **运行时间:** %.0f秒  
> **总时间:** %.0f秒  
> **结果数:** %d  
> **错误:** %s  

请登录Crawlab查看详情。
`,
		errMsg,
		t.Id,
		statusMsg,
		t.Param,
		s.Id.Hex(),
		s.Name,
		n.Name,
		utils.GetLocalTimeString(t.CreateTs),
		utils.GetLocalTimeString(t.StartTs),
		utils.GetLocalTimeString(t.FinishTs),
		t.WaitDuration,
		t.RuntimeDuration,
		t.TotalDuration,
		t.ResultCount,
		errLog,
	)
}
func SendTaskDingTalk(u model.User, t model.Task, s model.Spider) {
	statusMsg := "已完成"
	if t.Status == constants.StatusError {
		statusMsg = "发生错误"
	}
	title := fmt.Sprintf("[Crawlab] \"%s\" 任务%s", s.Name, statusMsg)
	content := GetTaskMarkdownContent(t, s)
	if err := notification.SendMobileNotification(u.Setting.DingTalkRobotWebhook, title, content); err != nil {
		log.Errorf(err.Error())
		debug.PrintStack()
	}
}

func SendTaskWechat(u model.User, t model.Task, s model.Spider) {
	content := GetTaskMarkdownContent(t, s)
	if err := notification.SendMobileNotification(u.Setting.WechatRobotWebhook, "", content); err != nil {
		log.Errorf(err.Error())
		debug.PrintStack()
	}
}

func SendNotifications(u model.User, t model.Task, s model.Spider) {
	if u.Email != "" && utils.StringArrayContains(u.Setting.EnabledNotifications, constants.NotificationTypeMail) {
		go func() {
			SendTaskEmail(u, t, s)
		}()
	}

	if u.Setting.DingTalkRobotWebhook != "" && utils.StringArrayContains(u.Setting.EnabledNotifications, constants.NotificationTypeDingTalk) {
		go func() {
			SendTaskDingTalk(u, t, s)
		}()
	}

	if u.Setting.WechatRobotWebhook != "" && utils.StringArrayContains(u.Setting.EnabledNotifications, constants.NotificationTypeWechat) {
		go func() {
			SendTaskWechat(u, t, s)
		}()
	}
}

func SendWebHookRequest(u model.User, t model.Task, s model.Spider) {
	type RequestBody struct {
		Status   string       `json:"status"`
		Task     model.Task   `json:"task"`
		Spider   model.Spider `json:"spider"`
		UserName string       `json:"user_name"`
	}

	if s.IsWebHook && s.WebHookUrl != "" {
		// request header
		header := req.Header{
			"Content-Type": "application/json; charset=utf-8",
		}

		// request body
		reqBody := RequestBody{
			Status:   t.Status,
			UserName: u.Username,
			Task:     t,
			Spider:   s,
		}

		// make POST http request
		res, err := req.Post(s.WebHookUrl, header, req.BodyJSON(reqBody))
		if err != nil {
			log.Errorf("sent web hook request with error: " + err.Error())
			debug.PrintStack()
			return
		}
		if res.Response().StatusCode != http.StatusOK {
			log.Errorf(fmt.Sprintf("sent web hook request with error http code: %d, task_id: %s, status: %s", res.Response().StatusCode, t.Id, t.Status))
			debug.PrintStack()
			return
		}
		log.Infof(fmt.Sprintf("sent web hook request, task_id: %s, status: %s)", t.Id, t.Status))
	}
}
