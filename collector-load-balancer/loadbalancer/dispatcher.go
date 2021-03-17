package loadbalancer

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"go.uber.org/zap"

	"github.com/dyweb/gommon/errors"
	"google.golang.org/grpc"

	"github.com/aws-observability/collector-load-balancer/configmanager"
	"github.com/aws-observability/collector-load-balancer/loadbalancer/promsd"
	clbpb "github.com/aws-observability/collector-load-balancer/proto/generated/clb"
	cmpb "github.com/aws-observability/collector-load-balancer/proto/generated/clb/configmanager"
)

type Dispatcher struct {
	logger *zap.Logger
}

func NewDispatcher(logger *zap.Logger) *Dispatcher {
	return &Dispatcher{logger: logger}
}

func (d *Dispatcher) SendTargets(ctx context.Context, podIP string, cstate *CollectorState) error {
	// Create client
	// TODO: keep a pool of clients or use http w/o keep alive
	addr := fmt.Sprintf("%s:%d", podIP, configmanager.DefaultGRPCPort)
	d.logger.Debug("Dialing", zap.String("Addr", addr))
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	conn, err := grpc.DialContext(ctx, addr, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return errors.Wrap(err, "grpc client connection creation failed")
	}
	client := cmpb.NewConfigManagerServiceClient(conn)

	// Send
	res, err := d.sendTargets(ctx, client, cstate.ScheduledTarget.Targets)
	if err != nil {
		return errors.Wrapf(err, "send targets failed to %s", podIP)
	}
	// FIXME: res is nil when error is also nil, something wrong with grpc...?
	if res == nil {
		d.logger.Debug("GPRC res and error are both nil")
		return nil
	}
	d.logger.Debug("Sent target", zap.String("PodIP", podIP), zap.Int("UpdatedFiles", int(res.UpdatedFiles)))
	return nil
}

func (d *Dispatcher) sendTargets(ctx context.Context, client cmpb.ConfigManagerServiceClient, targets map[promsd.TargetPath]promsd.Target) (*cmpb.UpdatePrometheusFileSDRRes, error) {
	// Group targets by files, which is job name
	targetsByJob := make(map[string][]promsd.Target)
	for _, t := range targets {
		list := targetsByJob[t.Job]
		list = append(list, t)
		targetsByJob[t.Job] = list
	}

	// Convert to proto, someday you realize all you do is yaml -> proto -> json
	var files []*clbpb.PrometheusFileSD
	merr := errors.NewMultiErr()
	for job, targets := range targetsByJob {
		var converted []*clbpb.PrometheusTargetSet
		for _, t := range targets {
			address, ok := promsd.GetAddressFromLabels(t.Labels)
			// NOTE: this should NEVER happen as the sd result always have __address__
			if !ok {
				merr.Append(errors.Errorf("can't find address for %s %v", t.Path, t.Labels))
				continue
			}
			ct := clbpb.PrometheusTargetSet{
				Targets: []string{
					address,
				},
				Labels: t.Labels,
			}
			converted = append(converted, &ct)
		}
		f := clbpb.PrometheusFileSD{
			Location: filepath.Join(configmanager.DefaultTargetsBaseDir, job+".json"),
			Targets:  converted,
		}
		files = append(files, &f)
	}
	// Stop here because conversion error means our internal logic is generating corrupted data.
	if merr != nil {
		return nil, merr.ErrorOrNil()
	}

	req := &cmpb.UpdatePrometheusFileSDReq{Files: files}
	return client.UpdatePrometheusFileSD(ctx, req)
}
