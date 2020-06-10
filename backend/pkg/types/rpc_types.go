package types

import (
	"crawlab/utils"
	"encoding/json"
)

// 任务消息
type TaskMessage struct {
	Id  string
	Cmd string
}

// 序列化任务消息
func (m *TaskMessage) ToString() (string, error) {
	data, err := json.Marshal(&m)
	if err != nil {
		return "", err
	}
	return utils.BytesToString(data), err
}
