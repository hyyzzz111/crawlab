package rpc

import (
	"crawlab/entity"
	"encoding/json"
	"github.com/apex/log"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
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
	output.MemoryUsagePercent = float64(output.MemoryUsage) / float64(output.MemoryTotal) * 100
	o = output

	return
}

func GetLocalNodeStats() (stats entity.NodeStats, err error) {
	// get memory stats
	m, err := mem.VirtualMemory()
	if err != nil {
		log.Errorf("get memory stats error: " + err.Error())
		debug.PrintStack()
		return stats, err
	}
	stats.MemoryTotal = m.Total
	stats.MemoryUsage = m.Used

	// get cpu stats
	cpuPercentages, err := cpu.Percent(0, false)
	if err != nil {
		log.Errorf("get cpu stats error: " + err.Error())
		debug.PrintStack()
		return stats, err
	}
	stats.CpuUsagePercent = cpuPercentages[0]

	// get disk usage
	du, err := disk.Usage("/")
	if err != nil {
		log.Errorf("get disk stats error: " + err.Error())
		debug.PrintStack()
		return stats, err
	}
	stats.DiskTotal = du.Total
	stats.DiskUsage = du.Used
	stats.DiskUsagePercent = du.UsedPercent
	return stats, nil
}
