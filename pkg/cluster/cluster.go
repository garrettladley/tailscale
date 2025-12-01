package cluster

import "github.com/garrettladley/tailscale/pkg/server"

type Cluster struct {
	Servers []*server.Server
}

func NewCluster(servers ...*server.Server) *Cluster {
	return &Cluster{
		Servers: servers,
	}
}

func (c *Cluster) Size() int {
	return len(c.Servers)
}
