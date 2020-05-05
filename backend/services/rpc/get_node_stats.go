package rpc

import (
	"crawlab/entity"
	"encoding/json"
	"github.com/apex/log"
	"github.com/shirou/gopsutil/mem"
	"runtime/debug"
)

type GetNodeStatsService struct {
	msg entity.RpcMessage
}

func (s *GetNodeStatsService) ServerHandle() (entity.RpcMessage, error) {
	stats, _ := GetLocalNodeStats()

	// 序列化
	resultStr, _ := json.Marshal(stats)
	s.msg.Result = string(resultStr)
	return s.msg, nil
}

func (s *GetNodeStatsService) ClientHandle() (o interface{}, err error) {
	// 发起 RPC 请求，获取服务端数据
	s.msg, err = ClientFunc(s.msg)()
	if err != nil {
		return o, err
	}

	var output entity.NodeStats
	if err := json.Unmarshal([]byte(s.msg.Result), &output); err != nil {
		return o, err
	}
	o = output

	return
}

func GetLocalNodeStats() (stats entity.NodeStats, err error) {
	m, err := mem.VirtualMemory()
	if err != nil {
		log.Errorf("get memory stats error: " + err.Error())
		debug.PrintStack()
		return stats, err
	}
	stats.TotalMemory = m.Total
	stats.MemoryUsage = m.Used
	return stats, nil
}
