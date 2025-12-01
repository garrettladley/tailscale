package cluster

import "github.com/garrettladley/tailscale/pkg/server"

type Cluster struct {
	servers []*server.Server
}
