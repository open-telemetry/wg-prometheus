package loadbalancer

import (
	"math"

	"go.uber.org/zap"

	otelv1 "github.com/aws-observability/collector-load-balancer/api/v1"
)

type Scaler struct {
	logger *zap.Logger
	cfg    otelv1.ScalerConfig
}

func NewScaler(cfg otelv1.ScalerConfig, logger *zap.Logger) *Scaler {
	return &Scaler{
		logger: logger.With(zap.String("Component", "clb/Scaler")),
		cfg:    cfg,
	}
}

func (s *Scaler) ExpectedReplicas(nTargets int) int {
	if !s.cfg.Enabled {
		panic("Scaler is not enabled")
	}

	cfg := s.cfg
	expected := int(math.Ceil(float64(nTargets) / float64(cfg.ExpectedTargetsPerCollector)))
	if expected < cfg.MinReplicas {
		expected = cfg.MinReplicas
	} else if expected > cfg.MaxReplicas {
		expected = cfg.MaxReplicas
	}
	s.logger.Info("Expected replica calculated", zap.Int("Expected", expected), zap.Int("Targets", nTargets),
		zap.Int("Min", cfg.MaxReplicas), zap.Int("Max", cfg.MaxReplicas))
	return expected
}
