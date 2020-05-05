package routes

import (
	"crawlab/constants"
	"crawlab/database"
	"crawlab/entity"
	"crawlab/services/rpc"
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

func GetNodeStats(c *gin.Context) {
	nodeId := c.Param("id")

	s := rpc.GetService(entity.RpcMessage{
		NodeId:  nodeId,
		Method:  constants.RpcGetNodeStats,
		Timeout: 60,
	})

	stats, err := s.ClientHandle()
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
