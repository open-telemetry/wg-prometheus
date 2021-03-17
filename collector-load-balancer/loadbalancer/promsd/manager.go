package promsd

import (
	"context"

	"golang.org/x/sync/errgroup"

	"github.com/dyweb/gommon/errors"
	"github.com/prometheus/prometheus/config"
	"github.com/prometheus/prometheus/discovery"
	"github.com/prometheus/prometheus/discovery/kubernetes"
	"go.uber.org/zap"

	_ "github.com/prometheus/prometheus/discovery/kubernetes" // register kubernetes_sd_configs

	"github.com/aws-observability/collector-load-balancer/loadbalancer/promsd/copiedlogger"
)

// manager.go starts a new prometheus service discovery

var _ kubernetes.SDConfig // make jump to definition easier ...

// DiscoveryManager wraps discovery.Manager.
type DiscoveryManager struct {
	mgr           *discovery.Manager
	logger        *zap.Logger
	initialConfig string
	targetsCh     chan []Target
	scrapeConfigs map[string]*config.ScrapeConfig // saved for relabel and drop targets early
}

// NewDiscoveryManager creates a prometheus discovery manger
// by extracting discovery config from a full prometheus configuration.
func NewDiscoveryManager(logger *zap.Logger, promCfg string) (*DiscoveryManager, error) {
	// Only validate the config
	_, _, err := newPrometheusScrapeConfig(promCfg)
	if err != nil {
		return nil, err
	}
	return &DiscoveryManager{
		logger:        logger.With(zap.String("Component", "clb/promsd/DiscoveryManager")),
		initialConfig: promCfg,
		targetsCh:     make(chan []Target),
	}, nil
}

func newPrometheusScrapeConfig(promCfg string) (map[string]discovery.Configs, map[string]*config.ScrapeConfig, error) {
	fullCfg, err := config.Load(promCfg)
	if err != nil {
		return nil, nil, errors.Wrap(err, "decode prometheus config failed")
	}
	scrapeConfigs := make(map[string]*config.ScrapeConfig)
	dCfg := make(map[string]discovery.Configs)
	for _, sCfg := range fullCfg.ScrapeConfigs {
		dCfg[sCfg.JobName] = sCfg.ServiceDiscoveryConfigs
		scrapeConfigs[sCfg.JobName] = sCfg
	}
	return dCfg, scrapeConfigs, nil
}

// Run starts discovery and blocks until error or context cancellation.
func (m *DiscoveryManager) Run(ctx context.Context) error {
	g, ctx := errgroup.WithContext(ctx)

	m.mgr = discovery.NewManager(ctx, copiedlogger.NewZapToGokitLogAdapter(m.logger))
	if err := m.loadConfig(m.initialConfig); err != nil {
		return err
	}

	g.Go(func() error {
		// Run starts consuming targets from discovery implementations
		m.logger.Info("Prometheus DiscoveryManager starting")
		if err := m.mgr.Run(); err != nil {
			m.logger.Error("DiscoverManger failed", zap.Error(err))
			return errors.Wrap(err, "discovery manager failed")
		}
		m.logger.Info("Prometheus Discovery Manager stopped")
		return nil
	})
	g.Go(func() error {
		for {
			select {
			case <-ctx.Done():
				return nil
			case targets := <-m.mgr.SyncCh():
				converted, dropped, err := ProcessTargetGroup(targets, m.scrapeConfigs)
				if err != nil {
					return errors.Wrap(err, "convert discovered target failed")
				}
				m.logger.Info("Processed Targets", zap.Int("Converted", len(converted)), zap.Int("Dropped", dropped))
				m.targetsCh <- converted
			}
		}
	})

	if err := g.Wait(); err != nil {
		m.logger.Error("DiscoveryManager stopped on error", zap.Error(err))
		return err
	}

	// Stopped due to context cancellation
	m.logger.Info("DiscoveryManager stopped")
	return nil
}

// TargetsCh sends out all the discovered targets, i.e. it's not incremental.
// For example, first message contains 5 targets, then a new pod got discovered,
// the second message would contain 6 targets instead of just the new pod.
func (m *DiscoveryManager) TargetsCh() <-chan []Target {
	return m.targetsCh
}

func (m *DiscoveryManager) ReloadConfig(promCfg string) error {
	if m.mgr == nil {
		return errors.New("discovery mgr is nil")
	}
	return m.loadConfig(promCfg)
}

func (m *DiscoveryManager) loadConfig(promCfg string) error {
	dCfgs, sCfgs, err := newPrometheusScrapeConfig(promCfg)
	if err != nil {
		return err
	}
	m.logger.Info("Decoded scrape config", zap.Int("JobCount", len(dCfgs)))
	m.scrapeConfigs = sCfgs

	// ApplyConfig actually starts discovery implementations in background ...
	if err := m.mgr.ApplyConfig(dCfgs); err != nil {
		return errors.Wrap(err, "apply discovery config failed")
	}
	return nil
}
