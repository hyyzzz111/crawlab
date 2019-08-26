package utils

import (
	"crawlab/constants"
	"crawlab/model"
	"github.com/gin-gonic/gin"
)

type UserContext struct {
	User *model.User
}
type Context struct {
	*gin.Context
}

func NewContext(context *gin.Context) *Context {
	return &Context{Context: context}
}

func (c *Context) CurrentUser() *UserContext {
	value, exists := c.Get(constants.ContextCurrentUserKey)
	user, ok := value.(*model.User)
	if !exists || !ok {
		return nil
	}
	return &UserContext{
		User: user,
	}
}
