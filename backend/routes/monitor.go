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
		return
	}
	c.JSON(http.StatusOK, Response{
		Status:  "ok",
		Message: "success",
		Data:    stats,
	})
}

func GetRedisStats(c *gin.Context) {
	stats, err := database.RedisClient.MemoryStats()
	if err != nil {
		HandleErrorF(http.StatusInternalServerError, c, err.Error())
		return
	}
	c.JSON(http.StatusOK, Response{
		Status:  "ok",
		Message: "success",
		Data:    stats,
	})
}
