package loadbalancer

import (
	"testing"

	otelv1 "github.com/aws-observability/collector-load-balancer/api/v1"
	"github.com/stretchr/testify/assert"
)

func TestInstanceStateStore_UpdateTargets(t *testing.T) {
	store := NewInstanceStateStore(otelv1.CollectorLoadBalancer{})
	state := store.ActiveState()
	assert.Equal(t, 0, state.Id)
	store.UpdateTargets(nil)
	assert.Equal(t, 0, state.Id)
	assert.Equal(t, 1, store.LastState().Id)
}
