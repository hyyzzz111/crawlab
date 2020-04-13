package config

import "crawlab/embed/raftserver"

type MasterNodeConfig struct {
	MemoryRegistry struct {
	}
	Cors struct {
	}
	RaftServer *raftserver.Config
}
