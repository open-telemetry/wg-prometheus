package loadbalancer

import (
	"sort"

	"go.uber.org/zap"

	"github.com/aws-observability/collector-load-balancer/loadbalancer/promsd"
)

type ScheduledTarget struct {
	Targets map[promsd.TargetPath]promsd.Target
}

func newScheduledTarget() *ScheduledTarget {
	return &ScheduledTarget{Targets: make(map[promsd.TargetPath]promsd.Target)}
}

func (s *ScheduledTarget) DeepCopy() *ScheduledTarget {
	m := make(map[promsd.TargetPath]promsd.Target)
	for k, v := range s.Targets {
		m[k] = *v.DeepCopy()
	}
	return &ScheduledTarget{Targets: m}
}

// Scheduler schedules prometheus scrape Targets to collectors.
type Scheduler struct {
	logger *zap.Logger
}

func NewScheduler(logger *zap.Logger) *Scheduler {
	return &Scheduler{logger: logger.With(zap.String("Component", "clb/Scheduler"))}
}

// Schedule is a naive greedy scheduling, find the collector with least targets to make sure most collectors get similar amount of targets.
func (s *Scheduler) Schedule(currentStates map[CollectorId]*CollectorState, expectedReplicas int, discoveredTargets []promsd.Target) (map[CollectorId]*CollectorState, error) {
	targetsToSchedule := promsd.TargetsToMap(discoveredTargets)
	// FIXME: it seems a lot target are having common path ...
	// 2021-03-17T04:26:18.043Z DEBUG Scanned current state {"InstanceName": "default/collectorloadbalancer-sample", "Component": "clb/Scheduler", "DiscoveredTargets": 12, "StaleTargets": 0, "TargetsToSchedule": 4}
	s.logger.Debug("Convert targets to map", zap.Int("Map", len(targetsToSchedule)), zap.Int("List", len(discoveredTargets)))

	// If replicas does not change, we just do minor modification otherwise we reshard everything.
	reshard := len(currentStates) != expectedReplicas
	// Loop through existing collectors, discard collectors that will be terminated.
	newStates := make(map[CollectorId]*CollectorState)
	deleted := 0
	for i := 0; i < expectedReplicas; i++ {
		id := CollectorId{Ordinal: i}
		var newState *CollectorState
		if existingState, ok := currentStates[id]; ok {
			newState = existingState.DeepCopy()
		} else {
			newState = newCollectorState(id)
		}
		newStates[id] = newState

		// If reshard, simply ignore existing targets, it's naive because it will create churn, but it's a demo so ...
		if reshard {
			newState.ScheduledTarget = newScheduledTarget()
			continue
		}
		// Otherwise do the diff
		for p := range newState.ScheduledTarget.Targets {
			if _, ok := targetsToSchedule[p]; ok {
				// Update existing target to use latest labels
				newState.ScheduledTarget.Targets[p] = targetsToSchedule[p]
				// We no longer need to schedule these targets
				delete(targetsToSchedule, p)
				s.logger.Debug("Dropping scheduled target", zap.String("Target", string(p)), zap.Int("Collector", id.Ordinal))
			} else {
				// Remove target that is no longer discovered from scheduled
				delete(newState.ScheduledTarget.Targets, p)
				deleted++
			}
		}
	}
	s.logger.Debug("Scanned current state", zap.Int("DiscoveredTargets", len(discoveredTargets)),
		zap.Int("StaleTargets", deleted), zap.Int("TargetsToSchedule", len(targetsToSchedule)),
		zap.Int("CurrentCollectors", len(currentStates)), zap.Int("ExpectedReplicas", expectedReplicas),
		zap.Bool("Reshard", reshard))

	// No new target found
	if len(targetsToSchedule) == 0 {
		return newStates, nil
	}

	// Schedule remaining targets to collectors
	pq := newCollectorPQFromMap(newStates)
	sortedTargets := sortTargets(targetsToSchedule)
	for _, t := range sortedTargets {
		// The priority queue make sure we get the collector with least number of assigned targets
		cstate := pq.Pop()
		cstate.ScheduledTarget.Targets[t.Path] = t
		pq.Push(cstate)
	}
	s.logger.Debug("Scheduled new targets", zap.Int("ScheduledTargets", len(sortedTargets)))
	return newStates, nil
}

func sortTargets(targets map[promsd.TargetPath]promsd.Target) []promsd.Target {
	var l []promsd.Target
	for _, v := range targets {
		l = append(l, v)
	}
	// Simply sort by target's address, we should do more complex sorting (and grouping)
	sort.Slice(l, func(i, j int) bool {
		return l[i].Path < l[j].Path
	})
	return l
}
