package loadbalancer

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/aws-observability/collector-load-balancer/loadbalancer/promsd"
)

func TestScheduler_Schedule(t *testing.T) {
	sched := NewScheduler(zap.NewExample())
	t.Run("first time", func(t *testing.T) {
		states, err := sched.Schedule(nil, 2, p2Targets("p1", "p2"))
		require.NoError(t, err)
		for _, s := range states {
			t.Logf("%d %v", s.Id.Ordinal, s.ScheduledTarget)
		}
	})

	t.Run("same replicas", func(t *testing.T) {
		current := map[CollectorId]*CollectorState{
			CollectorId{Ordinal: 0}: {
				Id: CollectorId{Ordinal: 0},
				ScheduledTarget: &ScheduledTarget{
					Targets: map[promsd.TargetPath]promsd.Target{
						"c0p1": {Path: "c0p1"},
						"c0p3": {Path: "c0p3"},
					},
				},
			},
			CollectorId{Ordinal: 1}: {
				Id: CollectorId{Ordinal: 1},
				ScheduledTarget: &ScheduledTarget{
					Targets: map[promsd.TargetPath]promsd.Target{
						"c1p2": {Path: "c1p2"},
						"c2p4": {Path: "c2p4"},
						"c2p5": {Path: "c2p5"},
					},
				},
			},
		}

		t.Run("same targets do nothing", func(t *testing.T) {
			allTargets := p2Targets("c0p1", "c0p3", "c1p2", "c2p4", "c2p5")
			newStates, err := sched.Schedule(current, 2, allTargets)
			require.NoError(t, err)
			assert.Equal(t, current, newStates)
		})

		t.Run("remove stale targets", func(t *testing.T) {
			allTargets := p2Targets("c0p1", "c1p2", "c2p4")
			newStates, err := sched.Schedule(current, 2, allTargets)
			require.NoError(t, err)
			assert.Equal(t, map[CollectorId]*CollectorState{
				CollectorId{Ordinal: 0}: {
					Id: CollectorId{Ordinal: 0},
					ScheduledTarget: &ScheduledTarget{
						Targets: map[promsd.TargetPath]promsd.Target{
							"c0p1": {Path: "c0p1"},
						},
					},
				},
				CollectorId{Ordinal: 1}: {
					Id: CollectorId{Ordinal: 1},
					ScheduledTarget: &ScheduledTarget{
						Targets: map[promsd.TargetPath]promsd.Target{
							"c1p2": {Path: "c1p2"},
							"c2p4": {Path: "c2p4"},
						},
					},
				},
			}, newStates)
		})

		t.Run("new target goes to collector with least targets", func(t *testing.T) {
			allTargets := p2Targets("c0p1", "c0p3", "c1p2", "c2p4", "c2p5", "c1p6")
			newStates, err := sched.Schedule(current, 2, allTargets)
			require.NoError(t, err)
			assert.Equal(t, map[CollectorId]*CollectorState{
				CollectorId{Ordinal: 0}: {
					Id: CollectorId{Ordinal: 0},
					ScheduledTarget: &ScheduledTarget{
						Targets: map[promsd.TargetPath]promsd.Target{
							"c0p1": {Path: "c0p1"},
							"c0p3": {Path: "c0p3"},
							"c1p6": {Path: "c1p6"},
						},
					},
				},
				CollectorId{Ordinal: 1}: {
					Id: CollectorId{Ordinal: 1},
					ScheduledTarget: &ScheduledTarget{
						Targets: map[promsd.TargetPath]promsd.Target{
							"c1p2": {Path: "c1p2"},
							"c2p4": {Path: "c2p4"},
							"c2p5": {Path: "c2p5"},
						},
					},
				},
			}, newStates)
		})
	})

	t.Run("reshard", func(t *testing.T) {
		current := map[CollectorId]*CollectorState{
			CollectorId{Ordinal: 0}: {
				Id: CollectorId{Ordinal: 0},
				ScheduledTarget: &ScheduledTarget{
					Targets: map[promsd.TargetPath]promsd.Target{
						"c0p1": {Path: "c0p1"},
						"c0p3": {Path: "c0p3"},
					},
				},
			},
			CollectorId{Ordinal: 1}: {
				Id: CollectorId{Ordinal: 1},
				ScheduledTarget: &ScheduledTarget{
					Targets: map[promsd.TargetPath]promsd.Target{
						"c1p2": {Path: "c1p2"},
						"c2p4": {Path: "c2p4"},
						"c2p5": {Path: "c2p5"},
					},
				},
			},
		}

		t.Run("scale up", func(t *testing.T) {
			allTargets := p2Targets("c0p1", "c0p3", "c1p2", "c2p4", "c2p5")
			newStates, err := sched.Schedule(current, 3, allTargets)
			require.NoError(t, err)
			assert.Equal(t, 3, len(newStates))
			// reshard is not e... deterministic ...
			for _, s := range newStates {
				t.Logf("id %d targets %d", s.Id.Ordinal, len(s.ScheduledTarget.Targets))
			}
		})

		t.Run("scale down", func(t *testing.T) {
			allTargets := p2Targets("c0p1", "c0p3", "c1p2", "c2p4", "c2p5")
			newStates, err := sched.Schedule(current, 1, allTargets)
			require.NoError(t, err)
			assert.Equal(t, 1, len(newStates))
			// reshard is not e... deterministic ... (actually here it is because there is just one shard ...
			for _, s := range newStates {
				t.Logf("id %d targets %d", s.Id.Ordinal, len(s.ScheduledTarget.Targets))
			}
		})
	})
}

func p2Targets(ps ...string) []promsd.Target {
	var targets []promsd.Target
	for _, p := range ps {
		targets = append(targets, promsd.Target{Path: promsd.TargetPath(p)})
	}
	return targets
}
