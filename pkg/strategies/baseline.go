package strategies

import (
	"context"
	"sync"
	"time"

	"github.com/garrettladley/tailscale/internal/xstrings"
	"github.com/garrettladley/tailscale/pkg/cluster"
	"github.com/garrettladley/tailscale/pkg/request"
	"github.com/garrettladley/tailscale/pkg/response"
)

func Baseline(ctx context.Context, cluster *cluster.Cluster, req request.Request) response.Response {
	var (
		responses = make([]response.Response, cluster.Size())
		wg        sync.WaitGroup
	)

	for i, server := range cluster.Servers {
		wg.Go(func() { responses[i] = server.Handle(ctx, req) })
	}

	return getBaselineFinalResponse(req.ID, responses...)
}

func getBaselineFinalResponse(requestID string, responses ...response.Response) response.Response {
	var (
		serverIDs         = make([]string, len(responses))
		maxLatency        time.Duration
		maxQueueWait      time.Duration
		maxProcessingTime time.Duration
		anySuccess        bool
		firstErr          error
	)

	for _, resp := range responses {
		serverIDs = append(serverIDs, resp.ServerID)

		maxLatency = max(maxLatency, resp.Latency)
		maxQueueWait = max(maxQueueWait, resp.QueueWait)
		maxProcessingTime = max(maxProcessingTime, resp.ProcessingTime)

		if !anySuccess && resp.Success {
			anySuccess = true
		}
		if !anySuccess && resp.Err != nil {
			firstErr = resp.Err
		}
	}

	var err error
	if !anySuccess && firstErr != nil {
		err = firstErr
	}

	serverID := xstrings.HashStrings("|", append([]string{"aggregated_servers"}, serverIDs...)...)
	return response.Response{
		RequestID:      requestID,
		Latency:        maxLatency,
		QueueWait:      maxQueueWait,
		ProcessingTime: maxProcessingTime,
		Success:        anySuccess,
		Err:            err,
		ServerID:       serverID,
	}
}
