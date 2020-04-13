package raftserver

type ClusterState int
//go:generate enumer -type ClusterState -linecomment -trimprefix ClusterState
const(
	ClusterStateNew ClusterState = iota
	ClusterStateExists
)
type Config struct {
	InitialCluster []string `cli:"initial-cluster"`
	InitialClusterToken string `yaml:"initial-cluster-token" cli:"initial-cluster-token" required:""`
	ClusterState ClusterState `yaml:"initial-cluster-state" type:"enum"`
}

