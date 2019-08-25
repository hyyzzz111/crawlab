package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"net/http"
)

func GetRoutings(ctx *gin.Context) {
	if _, exists := ctx.Get("currentUser"); exists {

	}
	var installed bool
	if HasAdminAccount != 0 {
		installed = true
	}
	enable_register := viper.GetBool("server.settings.enable_register")

}
func GetSettings(context *gin.Context) {
	var installed bool
	if HasAdminAccount != 0 {
		installed = true
	}
	cache_time := viper.GetInt("server.settings.cache_life")
	if cache_time == 0 {
		cache_time = 600
	}
	context.JSON(http.StatusOK,
		Response{
			Status:  "ok",
			Message: "success",
			Data: gin.H{
				"installed":       installed,
				"cache_time":      cache_time,
				"enable_register": viper.GetBool("server.settings.enable_register"),
			},
		})
}
