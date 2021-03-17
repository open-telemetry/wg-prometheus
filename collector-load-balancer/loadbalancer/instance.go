package loadbalancer

import (
	"context"
	"time"

	corev1 "k8s.io/api/core/v1"

	"github.com/dyweb/gommon/errors"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	otelv1 "github.com/aws-observability/collector-load-balancer/api/v1"
	"github.com/aws-observability/collector-load-balancer/loadbalancer/configparser"
	"github.com/aws-observability/collector-load-balancer/loadbalancer/promsd"
)

// InstanceName is just name of the crd
type InstanceName struct {
	Namespace string
	Name      string
}

func (n *InstanceName) String() string {
	return n.Namespace + "/" + n.Name
}

func NewInstanceName(clb otelv1.CollectorLoadBalancer) InstanceName {
	return InstanceName{
		Namespace: clb.Namespace,
		Name:      clb.Name,
	}
}

// Instance maps to a single crd. It is NOT a collector instance, see collector.go ...
type Instance struct {
	logger *zap.Logger

	name InstanceName

	cancel      func()
	sdmgr       *promsd.DiscoveryManager
	stateStore  *InstanceStateStore
	stateSyncCh chan *InstanceState
	k8sClient   *InstanceK8sClient
	scheduler   *Scheduler
	dispatcher  *Dispatcher
	scaler      *Scaler
}

type InstanceOptions struct {
	// Same for all instances
	Logger    *zap.Logger
	CRDScheme *runtime.Scheme
	K8sClient client.Client

	// Changes based on CRD
	Name         InstanceName
	InitialState otelv1.CollectorLoadBalancer
}

func NewInstance(opts InstanceOptions) (*Instance, error) {
	logger := opts.Logger.With(
		zap.String("Component", "clb/Instance"),
		zap.String("InstanceName", opts.Name.String()),
	)
	childLogger := opts.Logger.With(
		zap.String("InstanceName", opts.Name.String()),
	)

	// Prometheus SD
	promCfg, err := configparser.ExtractPrometheusK8sSD([]byte(opts.InitialState.Spec.CollectorConfig))
	if err != nil {
		return nil, errors.Wrap(err, "invalid otel prometheus receiver config")
	}
	sdmgr, err := promsd.NewDiscoveryManager(childLogger, string(promCfg))
	if err != nil {
		return nil, errors.Wrap(err, "create service discovery manager failed")
	}
	// State Store
	stateStore := NewInstanceStateStore(opts.InitialState)
	// k8s client
	opts.Logger = childLogger
	k8sClient := NewInstanceK8sClient(opts)
	return &Instance{
		logger:      logger,
		name:        opts.Name,
		sdmgr:       sdmgr,
		stateStore:  stateStore,
		stateSyncCh: make(chan *InstanceState, 5), // buffered so sender won't block
		k8sClient:   k8sClient,
		scheduler:   NewScheduler(childLogger),
		dispatcher:  NewDispatcher(childLogger),
		scaler:      NewScaler(opts.InitialState.Spec.Scale, childLogger),
	}, nil
}

func (ins *Instance) Run(ctx context.Context) error {
	ins.logger.Info("Running instance")

	ctx, cancel := context.WithCancel(ctx)
	ins.cancel = cancel
	g, ctx := errgroup.WithContext(ctx)

	// Run prometheus service discovery manager in background.
	g.Go(func() error {
		return ins.sdmgr.Run(ctx)
	})
	// Watch for targets and trigger sync
	g.Go(func() error {
		targetsCh := ins.sdmgr.TargetsCh()
		for {
			select {
			case <-ctx.Done():
				return nil
			case targets := <-targetsCh:
				ins.logger.Debug("Got Targets", zap.Int("Count", len(targets)))
				ins.stateSyncCh <- ins.stateStore.UpdateTargets(targets) // notify the sync handler
			}
		}
	})

	// Actual sync loop, watches for sync signals from targets and CRDReconcile
	g.Go(func() error {
		gcTicker := time.NewTicker(time.Minute)
		for {
			select {
			case <-ctx.Done():
				return nil
			case <-gcTicker.C:
				ins.stateStore.GC()
			case state := <-ins.stateSyncCh:
				ins.logger.Info("Syncing State", zap.Int("StateId", state.Id), zap.String("StateReason", string(state.UpdateReason)))
				if err := ins.sync(state); err != nil {
					ins.logger.Error("Sync State failed", zap.Error(err), zap.Int("StateId", state.Id))
				} else {
					ins.logger.Info("Synced State done", zap.Int("StateId", state.Id))
				}
			}
		}
	})
	return g.Wait()
}

func (ins *Instance) Stop(_ context.Context) error {
	ins.logger.Info("Stopping instance")
	ins.cancel()
	return nil
}

// CRDReconcile should be called the crd controller's Reconcile func.
// It indicates CRD itself has changed or resource it is watching has changed.
// However the source of change is not exposed in controller runtime.
func (ins *Instance) CRDReconcile(clb otelv1.CollectorLoadBalancer) error {
	// TODO(sync): apply prometheus config if needed
	// TODO(sync): if entire collector config changed, we should update sts as well
	ins.logger.Debug("CRDReconcile")
	ins.stateSyncCh <- ins.stateStore.UpdateCRD(clb)
	return nil
}

func (ins *Instance) sync(triggerState *InstanceState) error {
	var activeState, lastState *InstanceState
	ins.stateStore.Transaction(func() {
		activeState = ins.stateStore.ActiveStateNoLock()
		lastState = ins.stateStore.LastStateNoLock()
	})
	logger := ins.logger.With(zap.Int("ActiveStateId", activeState.Id),
		zap.Int("TriggerStateId", triggerState.Id), zap.Int("LastStateId", lastState.Id))
	if activeState.Id > triggerState.Id {
		logger.Debug("Ignore stale trigger state")
		return nil
	}

	// Always fetch latest info
	ctx := context.TODO()
	clb, err := ins.k8sClient.GetLatestCRD(ctx)
	if err != nil {
		return err
	}
	stsList, err := ins.k8sClient.GetStatefulSets(ctx)
	if err != nil {
		return err
	}
	if len(stsList.Items) == 0 {
		// sync will get triggered again because crd watches sts
		return ins.k8sClient.CreateStatefulSet(ctx, clb)
	}
	discoveredTargets := lastState.DiscoveredTargets

	// Update sts replicas (if needed)
	sts := stsList.Items[0]
	expectedReplicas := clb.Spec.Replicas
	// Scale by ourself
	if clb.Spec.Scale.Enabled {
		expectedReplicas = int32(ins.scaler.ExpectedReplicas(len(discoveredTargets)))
	}
	if err := ins.k8sClient.UpdateStatefulSetReplicas(ctx, sts, expectedReplicas); err != nil {
		return err
	}
	// TODO: update scale related status, but it's not that useful (for correctness) unless we apply HPA on CRD itself
	// Applying HPA on sts should be better because we can have more than one sts in the future.
	// TODO: need to redeploy the collectors if collector config itself has changed i.e. update sts pod spec

	// Schedule Targets based on expected replicas, state from last sync and Targets from latest state
	if len(lastState.DiscoveredTargets) == 0 {
		logger.Debug("No discovered target")
		// We should generates a signal/metric so the scaler (e.g. hpa or our own scaler impl) knowsthey can scale down to minimal.
		// But if a a collector has more than just prometheus receiver, then it should have a non 0 minimal value.
		return nil
	}
	existingSchedule := activeState.CollectorStates
	newSchedule, err := ins.scheduler.Schedule(existingSchedule, int(expectedReplicas), discoveredTargets)
	if err != nil {
		logger.Error("Schedule failed", zap.Error(err))
		return err
	}

	// Save our state
	var newActiveState *InstanceState
	ins.stateStore.Transaction(func() {
		newActiveState = ins.stateStore.UpdateCollectorStatesNoLock(newSchedule)
		// NOTE: we don't update DiscoveredTarget, CRD etc. because lastState we have when the scheduling start
		// may no longer be the last state now.
		newActiveState.CRDUsedInSchedule = clb

		ins.stateStore.SetActiveStateNoLock(newActiveState)
	})

	// Send out targets
	podList, err := ins.k8sClient.GetPodsInStatefulSet(ctx, sts)
	if err != nil {
		return err
	}
	for _, pod := range podList.Items {
		ordinal := getOrdinal(pod.Name)
		// Ignore pod that is going to be gone
		if ordinal >= int(expectedReplicas) {
			continue
		}
		// Ignore pod that is not ready
		if pod.Status.Phase != corev1.PodRunning || pod.Status.PodIP == "" {
			continue
		}

		cid := CollectorId{Ordinal: ordinal}
		cstate := newActiveState.CollectorStates[cid]
		podIP := pod.Status.PodIP
		logger.Debug("Sending targets", zap.Int("Ordinal", cid.Ordinal), zap.String("PodIP", podIP),
			zap.Int("NumTargets", len(cstate.ScheduledTarget.Targets)))
		if err := ins.dispatcher.SendTargets(ctx, podIP, cstate); err != nil {
			logger.Debug("Send target failed", zap.Error(err))
		}
	}

	return nil
}
