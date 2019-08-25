package routes

import (
	"crawlab/constants"
	"crawlab/model"
	"crawlab/services"
	"crawlab/utils"
	"github.com/gin-gonic/gin"
	"github.com/globalsign/mgo/bson"
	"github.com/pkg/errors"
	"net/http"
	"strings"
	"sync/atomic"
	"time"
)

var HasAdminAccount int32

type UserListRequestData struct {
	PageNum  int `form:"page_num"`
	PageSize int `form:"page_size"`
}

type UserRequestData struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func GetUser(c *gin.Context) {
	id := c.Param("id")

	user, err := model.GetUser(bson.ObjectIdHex(id))
	if err != nil {
		HandleError(http.StatusInternalServerError, c, err)
		return
	}

	c.JSON(http.StatusOK, Response{
		Status:  "ok",
		Message: "success",
		Data:    user,
	})
}

func GetUserList(c *gin.Context) {
	// 绑定数据
	data := UserListRequestData{}
	if err := c.ShouldBindQuery(&data); err != nil {
		HandleError(http.StatusBadRequest, c, err)
		return
	}
	if data.PageNum == 0 {
		data.PageNum = 1
	}
	if data.PageSize == 0 {
		data.PageNum = 10
	}

	// 获取用户列表
	users, err := model.GetUserList(nil, (data.PageNum-1)*data.PageSize, data.PageSize, "-create_ts")
	if err != nil {
		HandleError(http.StatusInternalServerError, c, err)
		return
	}

	// 获取总用户数
	total, err := model.GetUserListTotal(nil)
	if err != nil {
		HandleError(http.StatusInternalServerError, c, err)
		return
	}

	// 去除密码
	for i := range users {
		users[i].Password = ""
	}

	c.JSON(http.StatusOK, ListResponse{
		Status:  "ok",
		Message: "success",
		Data:    users,
		Total:   total,
	})
}

func PutAdminUser(context *gin.Context) {
	if old := atomic.SwapInt32(&HasAdminAccount, 1); old == 0 {
		exists, err := model.HasAdminUser()
		if err != nil {
			HandleError(http.StatusUnauthorized, context, err)
			atomic.StoreInt32(&HasAdminAccount, old)
			return
		}
		if exists {
			HandleError(http.StatusUnauthorized, context, errors.New("Admin User Has Been Registered."))
			return
		}

		// 绑定请求数据
		var reqData UserRequestData
		if err := context.ShouldBindJSON(&reqData); err != nil {
			HandleError(http.StatusBadRequest, context, err)
			return
		}

		// 添加用户
		adminUser := model.User{
			Username: strings.ToLower(reqData.Username),
			Password: utils.EncryptPassword(reqData.Password),
			Roles:    []string{constants.RoleAdmin},
		}
		if err := adminUser.Add(); err != nil {
			atomic.StoreInt32(&HasAdminAccount, old)
			HandleError(http.StatusUnauthorized, context, err)
			return
		}
		// 获取token
		tokenStr, err := services.GetToken(adminUser.Username)
		if err != nil {
			HandleError(http.StatusUnauthorized, context, errors.New("not authorized"))
			return
		}

		context.JSON(http.StatusOK, Response{
			Status:  "ok",
			Message: "success",
			Data:    tokenStr,
		})
	} else {
		HandleError(http.StatusUnauthorized, context, errors.New("Admin User Has Been Registered."))
		return
	}
}
func PutUser(c *gin.Context) {
	// 绑定请求数据
	var reqData UserRequestData
	if err := c.ShouldBindJSON(&reqData); err != nil {
		HandleError(http.StatusBadRequest, c, err)
		return
	}

	// 添加用户
	user := model.User{
		Username: strings.ToLower(reqData.Username),
		Password: utils.EncryptPassword(reqData.Password),
		Roles:    []string{constants.RoleNormal},
	}
	if err := user.Add(); err != nil {
		HandleError(http.StatusInternalServerError, c, err)
		return
	}

	c.JSON(http.StatusOK, Response{
		Status:  "ok",
		Message: "success",
	})
}

func PostUser(c *gin.Context) {
	id := c.Param("id")

	if !bson.IsObjectIdHex(id) {
		HandleErrorF(http.StatusBadRequest, c, "invalid id")
	}

	var item model.User
	if err := c.ShouldBindJSON(&item); err != nil {
		HandleError(http.StatusBadRequest, c, err)
		return
	}

	if err := model.UpdateUser(bson.ObjectIdHex(id), item); err != nil {
		HandleError(http.StatusInternalServerError, c, err)
		return
	}

	c.JSON(http.StatusOK, Response{
		Status:  "ok",
		Message: "success",
	})
}

func DeleteUser(c *gin.Context) {
	id := c.Param("id")

	if !bson.IsObjectIdHex(id) {
		HandleErrorF(http.StatusBadRequest, c, "invalid id")
		return
	}

	// 从数据库中删除该爬虫
	if err := model.RemoveUser(bson.ObjectIdHex(id)); err != nil {
		HandleError(http.StatusInternalServerError, c, err)
		return
	}

	c.JSON(http.StatusOK, Response{
		Status:  "ok",
		Message: "success",
	})
}

func Login(c *gin.Context) {
	// 绑定请求数据
	var reqData UserRequestData
	if err := c.ShouldBindJSON(&reqData); err != nil {
		HandleError(http.StatusUnauthorized, c, errors.New("not authorized"))
		return
	}

	// 获取用户
	user, err := model.GetUserByUsername(strings.ToLower(reqData.Username))
	if err != nil {
		HandleError(http.StatusUnauthorized, c, errors.New("not authorized"))
		return
	}

	// 校验密码
	encPassword := utils.EncryptPassword(reqData.Password)
	if user.Password != encPassword {
		HandleError(http.StatusUnauthorized, c, errors.New("not authorized"))
		return
	}

	// 获取token
	tokenStr, err := services.GetToken(user.Username)
	if err != nil {
		HandleError(http.StatusUnauthorized, c, errors.New("not authorized"))
		return
	}

	c.JSON(http.StatusOK, Response{
		Status:  "ok",
		Message: "success",
		Data:    tokenStr,
	})
}

func GetMe(c *gin.Context) {
	user, exists := c.Get("currentUser")
	// 校验token
	if !exists {
		HandleError(http.StatusUnauthorized, c, errors.New("not authorized"))
		return
	}
	c.JSON(http.StatusOK, Response{
		Status:  "ok",
		Message: "success",
		Data: struct {
			*model.User
			Password string `json:"password,omitempty"`
		}{
			User: user.(*model.User),
		},
	})
}
func RegisterFromInvitationURL(c *gin.Context) {

	// 绑定请求数据
	var reqData struct {
		E               string `json:"e"`
		Username        string `json:"username"`
		Password        string `json:"password"`
		ConfirmPassword string `json:"confirm_password"`
	}
	if err := c.ShouldBindJSON(&reqData); err != nil {
		HandleError(http.StatusBadRequest, c, err)
		return
	}
	invitation, err := validateInvitationURL(c.Param("id"), reqData.E)
	if err != nil {
		HandleError(http.StatusBadRequest, c, err)
		return
	}
	//
	// 添加用户
	user := model.User{
		Id:       bson.NewObjectId(),
		Username: strings.ToLower(reqData.Username),
		Password: utils.EncryptPassword(reqData.Password),
		Roles:    invitation.Roles,
	}
	if err := user.Add(); err != nil {
		HandleError(http.StatusInternalServerError, c, err)
		return
	}
	if err := model.UpdateInvitation(invitation.Id, bson.M{"used": true, "account": user.Username}); err != nil {
		//TODO 暂时忽略更新失败的情况
		HandleError(http.StatusInternalServerError, c, err)
		return
	}
	c.JSON(http.StatusOK, Response{
		Status:  "ok",
		Message: "success",
	})

}
func validateInvitationURL(id string, e string) (*model.Invitation, error) {

	inv, err := model.GetInvitation(bson.ObjectIdHex(id))
	if err != nil {
		return nil, err
	}
	if !inv.Status || inv.ExpireTs.Before(time.Now()) || inv.Used || e != inv.EncryptRT {
		return nil, errors.New("bad request")
	}
	return inv, nil
}
func ValidateInvitationURL(c *gin.Context) {
	id := c.Param("id")
	var reqData struct {
		E string `json:"e"`
	}
	if err := c.ShouldBindJSON(&reqData); err != nil {
		HandleError(http.StatusBadRequest, c, err)
		return
	}
	_, err := validateInvitationURL(id, reqData.E)
	if err != nil {
		HandleError(http.StatusBadRequest, c, err)

		return
	}
	c.JSON(http.StatusOK, gin.H{
		"validate": true,
	})
}
func GetInvitationURLList(c *gin.Context) {
	r, err := model.GetInvitationList(bson.M{"status": true})
	if err != nil {
		HandleError(http.StatusBadRequest, c, errors.New("bad request"))
		return
	}
	var result []interface{}
	for _, invitation := range r {
		result = append(result, map[string]interface{}{
			"query": gin.H{
				"e": invitation.EncryptRT,
			},
			"id":        invitation.Id,
			"create_ts": invitation.CreateTs,
			"expire_ts": invitation.CreateTs,
			"used":      invitation.Used,
			"status":    invitation.Status,
			"roles":     invitation.Roles,
			"acls":      invitation.ACLs,
		})
	}
	c.JSON(http.StatusOK, result)
}

func GenerateInvitationURL(c *gin.Context) {
	var requestData struct {
		Roles []string `json:"roles"`
		Acls  []string `json:"acls"`
	}
	if err := c.ShouldBind(&requestData); err != nil {
		HandleError(http.StatusBadRequest, c, errors.New("bad request"))
		return
	}
	salt := utils.GetRandomString(20)
	id := utils.GetRandomString(20)
	requestData.Roles = []string{"admin"}
	invited := &model.Invitation{
		CreateTs:      time.Now(),
		UpdateTs:      time.Now(),
		Roles:         requestData.Roles,
		ACLs:          []string{},
		Salt:          salt,
		Token:         id,
		RegisterLimit: 1,
		Status:        true,
		ExpireTs:      time.Now().Add(time.Hour * 72),
		EncryptRT:     utils.MD5(utils.MD5(id+salt) + salt),
	}
	if err := invited.Add(); err != nil {
		HandleError(http.StatusBadRequest, c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{})
}

func UpdateInvitationURL(c *gin.Context) {
	var reqData struct {
		Status bool     `json:"status"`
		Roles  []string `json:"roles"`
		Acls   []string `json:"acls"`
	}
	if err := c.ShouldBindJSON(&reqData); err != nil {
		HandleError(http.StatusBadRequest, c, errors.New("bad request"))
		return
	}
	id := c.Param("id")

	err := model.UpdateInvitation(bson.ObjectIdHex(id), bson.M{
		"status": reqData.Status,
		"roles":  reqData.Roles,
		"acls":   reqData.Acls,
	})
	if err != nil {
		HandleError(http.StatusBadRequest, c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{})
}
