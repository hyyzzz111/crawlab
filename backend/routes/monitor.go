package routes

import (
	"crawlab/database"
	"github.com/gin-gonic/gin"
	"github.com/globalsign/mgo/bson"
	"net/http"
)

func GetMongoStats(c *gin.Context) {
	s, db := database.GetDb()
	defer s.Close()
	stats := bson.M{}
	if err := db.Run("dbstats", &stats); err != nil {
		HandleErrorF(http.StatusInternalServerError, c, err.Error())
	}
	c.JSON(http.StatusOK, Response{
		Status:  "ok",
		Message: "success",
		Data:    stats,
	})
}
