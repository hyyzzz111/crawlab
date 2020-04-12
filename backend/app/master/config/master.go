package config

type MasterNodeConfig struct {
	MemoryRegistry struct {
	}
	Cors struct {
	}
	Raft struct {
		InitialCluster []string `yaml:"initial_cluster"`
	}
}
