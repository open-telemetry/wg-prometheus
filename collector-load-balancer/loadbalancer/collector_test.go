package loadbalancer

import (
	"testing"

	"github.com/aws-observability/collector-load-balancer/loadbalancer/promsd"
)

func TestCollectorsPQ(t *testing.T) {
	genItems := func() []*CollectorState {
		return []*CollectorState{
			{
				Id:              CollectorId{Ordinal: 0},
				ScheduledTarget: &ScheduledTarget{Targets: map[promsd.TargetPath]promsd.Target{"a": {}}},
			},
			{
				Id:              CollectorId{Ordinal: 1},
				ScheduledTarget: &ScheduledTarget{Targets: map[promsd.TargetPath]promsd.Target{"b": {}, "c": {}, "d": {}}},
			},
			{
				Id:              CollectorId{Ordinal: 2},
				ScheduledTarget: &ScheduledTarget{Targets: map[promsd.TargetPath]promsd.Target{"e": {}, "f": {}}},
			},
			{
				Id:              CollectorId{Ordinal: 3},
				ScheduledTarget: &ScheduledTarget{Targets: map[promsd.TargetPath]promsd.Target{"g": {}, "h": {}, "i": {}, "j": {}}},
			},
		}
	}
	t.Run("sort", func(t *testing.T) {
		items := genItems()
		pq := newCollectorPQ(items)
		for pq.Len() > 0 {
			item := pq.Pop()
			t.Logf("score %d id %d", item.score(), item.Id.Ordinal)
		}
		//    collector_test.go:30: score 1 id 0
		//    collector_test.go:30: score 2 id 2
		//    collector_test.go:30: score 3 id 1
		//    collector_test.go:30: score 4 id 3
	})
	t.Run("update", func(t *testing.T) {
		items := genItems()
		pq := newCollectorPQ(items)
		item := pq.Pop()
		item.ScheduledTarget = &ScheduledTarget{Targets: map[promsd.TargetPath]promsd.Target{
			"a1": {}, "a2": {}, "a3": {}, "a4": {}, "a5": {}}}
		pq.Push(item)
		for pq.Len() > 0 {
			item := pq.Pop()
			t.Logf("score %d id %d", item.score(), item.Id.Ordinal)
		}
		//    collector_test.go:49: score 2 id 2
		//    collector_test.go:49: score 3 id 1
		//    collector_test.go:49: score 4 id 3
		//    collector_test.go:49: score 5 id 0
	})
}
