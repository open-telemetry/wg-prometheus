package loadbalancer

import (
	"testing"

	otelv1 "github.com/aws-observability/collector-load-balancer/api/v1"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestScaler_ExpectedReplicas(t *testing.T) {
	s := NewScaler(otelv1.ScalerConfig{
		Enabled:                     true,
		ExpectedTargetsPerCollector: 3,
		MinReplicas:                 1,
		MaxReplicas:                 10,
	}, zap.NewExample())

	assert.Equal(t, 1, s.ExpectedReplicas(0))
	assert.Equal(t, 1, s.ExpectedReplicas(3))
	assert.Equal(t, 2, s.ExpectedReplicas(5))
	assert.Equal(t, 2, s.ExpectedReplicas(6))
	assert.Equal(t, 10, s.ExpectedReplicas(3000))
}
