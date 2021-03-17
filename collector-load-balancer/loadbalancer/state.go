package loadbalancer

import (
	"sync"

	otelv1 "github.com/aws-observability/collector-load-balancer/api/v1"
	"github.com/aws-observability/collector-load-balancer/loadbalancer/promsd"
)

// InstanceState is a snapshot of a single instance.
// It is immutable after it is saved into InstanceStateStore.
type InstanceState struct {
	Id                int // Id starts from 0 and increases by 1, it is only unique within the lifetime of the InstanceStateStore.
	UpdateReason      StateUpdateReason
	CRD               otelv1.CollectorLoadBalancer
	CRDUsedInSchedule otelv1.CollectorLoadBalancer
	DiscoveredTargets []promsd.Target
	CollectorStates   map[CollectorId]*CollectorState
}

func (s *InstanceState) ShallowCopy() *InstanceState {
	s2 := *s
	return &s2
}

type StateUpdateReason string

const (
	StateReasonInstanceCreated       StateUpdateReason = "InstanceCreated"
	StateReasonNewTargetsFound       StateUpdateReason = "NewTargetsFound"
	StateReasonCRDReconcile          StateUpdateReason = "CRDReconcile"
	StateReasonCollectorStateUpdated StateUpdateReason = "CollectorStateUpdated"
	StateReasonRetryRPC              StateUpdateReason = "RetryRPC"
	StateReasonScaleUp               StateUpdateReason = "ScaleUp"
	StateReasonScaleDown             StateUpdateReason = "ScaleDown"
)

const (
	minInstanceStates = 5
	maxInstanceStates = 25
)

// InstanceStateStore saves a series of state in time order.
// It is thread safe.
type InstanceStateStore struct {
	mu       sync.Mutex
	nextId   int
	activeId int
	states   map[int]*InstanceState
}

func NewInstanceStateStore(clb otelv1.CollectorLoadBalancer) *InstanceStateStore {
	return &InstanceStateStore{
		nextId:   1,
		activeId: 0,
		states: map[int]*InstanceState{
			0: {
				Id:           0,
				CRD:          clb,
				UpdateReason: StateReasonInstanceCreated,
			},
		},
	}
}

// GC make sure old states got released so the actual gc can recycle the map and slice referred by those states.
func (s *InstanceStateStore) GC() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.states) <= maxInstanceStates {
		return
	}
	// Make sure we don't delete the active one.
	// nextId = 26
	// activeId = 25, bound = min(25, 26 - 5) = 21
	// activeId = 10, bound = min(10, 26 - 5) = 10
	bound := minInt(s.activeId, s.nextId-minInstanceStates)
	for id := range s.states {
		if id < bound {
			delete(s.states, id)
		}
	}
}

func (s *InstanceStateStore) ActiveState() *InstanceState {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.ActiveStateNoLock()
}

func (s *InstanceStateStore) ActiveStateNoLock() *InstanceState {
	return s.states[s.activeId]
}

func (s *InstanceStateStore) SetActiveStateNoLock(a *InstanceState) {
	s.activeId = a.Id
}

func (s *InstanceStateStore) LastState() *InstanceState {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.LastStateNoLock()
}

func (s *InstanceStateStore) LastStateNoLock() *InstanceState {
	return s.states[s.nextId-1]
}

func (s *InstanceStateStore) UpdateTargets(targets []promsd.Target) *InstanceState {
	s.mu.Lock()
	defer s.mu.Unlock()

	state := s.newtState()
	state.DiscoveredTargets = targets
	state.UpdateReason = StateReasonNewTargetsFound
	return state
}

func (s *InstanceStateStore) UpdateCRD(clb otelv1.CollectorLoadBalancer) *InstanceState {
	s.mu.Lock()
	defer s.mu.Unlock()

	state := s.newtState()
	state.CRD = clb
	state.UpdateReason = StateReasonCRDReconcile
	return state
}

func (s *InstanceStateStore) UpdateCollectorStatesNoLock(cstates map[CollectorId]*CollectorState) *InstanceState {
	state := s.newtState()
	state.CollectorStates = cstates
	state.UpdateReason = StateReasonCollectorStateUpdated
	return state
}

func (s *InstanceStateStore) Transaction(f func()) {
	s.mu.Lock()
	defer s.mu.Unlock()

	f()
}

func (s *InstanceStateStore) newtState() *InstanceState {
	// Only do shallow copy because state should be read only
	// If Targets, crd need to be updated, they will be updated as a whole.
	state := s.states[s.nextId-1].ShallowCopy()
	id := s.nextId
	s.nextId++
	state.Id = id
	s.states[id] = state
	return state
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
