package promsd

import (
	"github.com/dyweb/gommon/errors"

	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/config"
	"github.com/prometheus/prometheus/discovery/targetgroup"
	"github.com/prometheus/prometheus/pkg/labels"
	"github.com/prometheus/prometheus/pkg/relabel"
)

const (
	defaultScheme      = "http"
	defaultMetricsPath = "/metrics"
)

const (
	labelPodName          = "__meta_kubernetes_pod_nam"
	labelPodContainerName = "__meta_kubernetes_pod_container_name"
	labelPodNodeName      = "__meta_kubernetes_pod_node_name"
)

// TargetPath is the actual scrape path e.g. http://192.168.7.2:1234/metrics
type TargetPath string

type Target struct {
	Path   TargetPath        // Path is the actual scrape path e.g. http://192.168.7.2:1234/metrics
	Labels map[string]string // Labels contains__address__ and  __meta__ labels
	Source string            // Source is set by discovery provider
	Job    string            // Job is the job_name in scrape_configs

	// Decoded from label and source, can be used in grouping during scheduling

	Container string
	Pod       string
	Node      string
}

// ProcessTargetGroup flatten target groups. It applies to drop targets and flatten all the targets into a single slice.
func ProcessTargetGroup(discoveredTargets map[string][]*targetgroup.Group,
	scrapeConfigs map[string]*config.ScrapeConfig) ([]Target, int, error) {
	var targets []Target
	merr := errors.NewMultiErr()
	dropped := 0
	for jobName, groups := range discoveredTargets {
		scrapeConfig := scrapeConfigs[jobName]
		if scrapeConfig == nil {
			merr.Append(errors.Errorf("job not found from scrape configs %s", jobName))
			continue
		}

		for _, group := range groups {
			for _, target := range group.Targets {
				// target contains address and container specific labels like container name
				// group contains common labels like pod etc.
				merged := target.Clone().Merge(group.Labels)

				// Relabel but it's for checking if the target is going to get dropped.
				// This makes scheduling decision more accurate.
				relabeled := relabel.Process(labelSetToLabels(merged), scrapeConfig.RelabelConfigs...)
				if relabeled == nil {
					dropped++
					continue
				}

				// Get path, address is required
				address, ok := target[model.AddressLabel]
				if !ok {
					merr.Append(errors.Errorf("address not found from %v", target))
					continue
				}
				// TODO: use the relabled metrics path?
				scheme, ok := merged[model.SchemeLabel]
				if !ok {
					scheme = defaultScheme
				}
				metricsPath, ok := merged[model.MetricsPathLabel]
				if !ok {
					metricsPath = defaultMetricsPath
				}
				// FIXME: this path is not that unique ...
				path := scheme + "://" + address + metricsPath

				converted := Target{
					Path:   TargetPath(path),
					Labels: labelSetToMap(merged), // use labels without relabel
					Source: group.Source,
					Job:    jobName,
				}
				converted.DecodePodLabels()
				targets = append(targets, converted)
			}
		}
	}
	return targets, dropped, merr.ErrorOrNil()
}

func (t *Target) DecodePodLabels() {
	t.Container = t.Labels[labelPodContainerName]
	t.Pod = t.Labels[labelPodName]
	t.Node = t.Labels[labelPodNodeName]
}

func (t *Target) DeepCopy() *Target {
	t2 := *t
	l2 := make(map[string]string)
	for k, v := range t.Labels {
		l2[k] = v
	}
	t2.Labels = l2
	return &t2
}

func TargetsToMap(targets []Target) map[TargetPath]Target {
	m := make(map[TargetPath]Target)
	for _, t := range targets {
		m[t.Path] = t
	}
	return m
}

func MapToTargets(targets map[TargetPath]Target) []Target {
	var l []Target
	for _, v := range targets {
		l = append(l, v)
	}
	return l
}

func GetAddressFromLabels(labels map[string]string) (string, bool) {
	s, ok := labels[model.AddressLabel]
	return s, ok
}

func labelSetToLabels(labelSet model.LabelSet) labels.Labels {
	var ls labels.Labels
	for k, v := range labelSet {
		ls = append(ls, labels.Label{Name: string(k), Value: string(v)})
	}
	return ls
}

func labelSetToMap(labelSet model.LabelSet) map[string]string {
	m := make(map[string]string)
	for k, v := range labelSet {
		m[string(k)] = string(v)
	}
	return m
}
