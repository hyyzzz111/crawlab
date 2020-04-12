package config

import (
	"crawlab/runtime"
	"net"
	"strconv"
)

type WorkerNodeConfig struct {
}
type RegistryCenterConfig struct {
	Type runtime.RegistryType
}

var DefaultConfig = new(ApplicationConfig)

type ApplicationConfig struct {
	Mode         runtime.EnvMode      `yaml:"mode" type:"enum"`
	NodeType     runtime.NodeType     `yaml:"node_type" type:"enum" default:""`
	RegistryType runtime.RegistryType `yaml:"registry_type" type:"enum" enums:"master,etcd" default:"master"`
	Server       struct {
		Port      int    `yaml:"port" default:"8080"`
		Host      string `yaml:"host" default:"0.0.0.0"`
		SSL       SSL
		Websocket Websocket `yaml:"websocket"`
	}
	Master MasterNodeConfig `yaml:"master"`
	Worker WorkerNodeConfig `yaml:"worker"`
}
type Websocket struct {
	Enable bool   `yaml:"enable"` //TODO
	Path   string `yaml:"path" default:"/ws"`
}

func (a *ApplicationConfig) HostWithoutProtocol() string {
	host := a.Server.Host
	port := a.Server.Port
	return net.JoinHostPort(host, strconv.Itoa(port))
}
