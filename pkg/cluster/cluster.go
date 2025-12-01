package cluster

import "github.com/garrettladley/tailscale/pkg/server"

type Cluster struct {
	servers []*server.Server
}

func NewCluster(servers ...*server.Server) *Cluster {
	return &Cluster{
		servers: servers,
	}
}
