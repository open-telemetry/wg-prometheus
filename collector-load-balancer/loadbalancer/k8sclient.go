package loadbalancer

import (
	"context"
	"regexp"
	"strconv"

	"github.com/dyweb/gommon/errors"
	"go.uber.org/zap"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	otelv1 "github.com/aws-observability/collector-load-balancer/api/v1"
	"github.com/aws-observability/collector-load-balancer/configmanager"
	"github.com/aws-observability/collector-load-balancer/loadbalancer/configparser"
)

// InstanceK8sClient is a wrapper on underlying client for a specific crd i.e. instance.
type InstanceK8sClient struct {
	logger       *zap.Logger
	initialState otelv1.CollectorLoadBalancer
	scheme       *runtime.Scheme
	client       client.Client
}

func NewInstanceK8sClient(opts InstanceOptions) *InstanceK8sClient {
	return &InstanceK8sClient{
		logger:       opts.Logger.With(zap.String("Component", "clb/InstanceK8sClient")),
		initialState: opts.InitialState,
		scheme:       opts.CRDScheme,
		client:       opts.K8sClient,
	}
}

func (c *InstanceK8sClient) GetLatestCRD(ctx context.Context) (otelv1.CollectorLoadBalancer, error) {
	var clb otelv1.CollectorLoadBalancer
	err := c.client.Get(ctx, client.ObjectKeyFromObject(&c.initialState), &clb)
	return clb, err
}

func (c *InstanceK8sClient) GetStatefulSets(ctx context.Context) (appsv1.StatefulSetList, error) {
	var stsList appsv1.StatefulSetList
	if err := c.client.List(ctx, &stsList, client.InNamespace(c.initialState.Namespace),
		client.MatchingFields{K8sControllerOwnerKey: c.initialState.Name}); err != nil {
		return appsv1.StatefulSetList{}, err
	}
	if len(stsList.Items) > 1 {
		return stsList, errors.Errorf("too many statefulsets, expect one got %d", len(stsList.Items))
	}
	return stsList, nil
}

func (c *InstanceK8sClient) CreateStatefulSet(ctx context.Context, clb otelv1.CollectorLoadBalancer) error {
	c.logger.Info("Creating StatefulSet")
	otelCfg, err := configparser.ReplacePrometheusK8sSD([]byte(clb.Spec.CollectorConfig), configmanager.DefaultTargetsBaseDir)
	if err != nil {
		return errors.Wrap(err, "replace prometheus k8s sd failed")
	}
	newSts := newStatefulSetManifest(clb, string(otelCfg))
	if err := ctrl.SetControllerReference(&clb, &newSts, c.scheme); err != nil {
		return errors.Wrap(err, "set controller reference failed")
	}
	if err = c.client.Create(ctx, &newSts); err != nil {
		return errors.Wrap(err, "create StatefulSet failed")
	}
	c.logger.Info("Created StatefulSet")
	return nil
}

func (c *InstanceK8sClient) UpdateStatefulSetReplicas(ctx context.Context, sts appsv1.StatefulSet, replicas int32) error {
	if *sts.Spec.Replicas == replicas {
		return nil
	}
	c.logger.Info("Changing StatefulSet replicas", zap.Int32("Old", *sts.Spec.Replicas), zap.Int32("New", replicas))
	expected := sts.DeepCopy()
	expected.Spec.Replicas = &replicas
	patch := client.MergeFrom(&sts)
	if err := c.client.Patch(context.TODO(), expected, patch); err != nil {
		c.logger.Error("Patch StatefulSet failed", zap.Error(err))
		return errors.Wrap(err, "patch StatefulSet failed")
	}
	c.logger.Info("Changed StatefulSet replicas", zap.Int32("Replicas", replicas))
	return nil
}

func (c *InstanceK8sClient) GetPodsInStatefulSet(ctx context.Context, sts appsv1.StatefulSet) (corev1.PodList, error) {
	var podsList corev1.PodList
	labelSelector, err := metav1.LabelSelectorAsSelector(sts.Spec.Selector)
	if err != nil {
		return podsList, err
	}
	err = c.client.List(ctx, &podsList, client.InNamespace(sts.Namespace), client.MatchingLabelsSelector{Selector: labelSelector})
	return podsList, err
}

var statefulPodRegex = regexp.MustCompile("(.*)-([0-9]+)$")

// getOrdinal returns position of a pod within a StatefulSet
func getOrdinal(podName string) int {
	ordinal := -1
	subMatches := statefulPodRegex.FindStringSubmatch(podName)
	if len(subMatches) < 3 {
		return ordinal
	}
	if i, err := strconv.ParseInt(subMatches[2], 10, 32); err == nil {
		ordinal = int(i)
	}
	return ordinal
}

func newStatefulSetManifest(clb otelv1.CollectorLoadBalancer, collectorConfig string) appsv1.StatefulSet {
	// NOTE: we rely on owner reference and do NOT use the following labels when fetching sts.
	labels := map[string]string{
		"app.kubernetes.io/managed-by": "collector-load-balancer",
		"app.kubernetes.io/instance":   clb.Namespace + "__" + clb.Name, // NOTE: label value can't be /
	}
	// TODO: should add prometheus annotation, e.g. scrape metrics of collector itself
	annotations := make(map[string]string)
	podLabels := copyMap(labels)
	podAnnotations := copyMap(annotations)
	replicas := clb.Spec.Replicas
	return appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:        clb.Name + "-collector",
			Namespace:   clb.Namespace,
			Labels:      labels,
			Annotations: annotations,
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      podLabels,
					Annotations: podAnnotations,
				},
				Spec: corev1.PodSpec{
					// TODO: service account name, but it is not very useful unless we enable auth proxy on the manager
					Containers: []corev1.Container{
						{
							Name:            "aoc2",
							Image:           clb.Spec.CollectorImage,
							ImagePullPolicy: corev1.PullAlways,
							// NOTE: use env to pass collector config, ugly but easy to debug
							Env: []corev1.EnvVar{
								{
									Name:  AOTENVKey,
									Value: collectorConfig,
								},
								{
									Name:  "AWS_REGION", // For EMF exporter
									Value: "us-west-2",
								},
							},
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: configmanager.DefaultGRPCPort,
								},
								{
									ContainerPort: configmanager.DefaultHTTPPort,
								},
							},
						},
					},
				},
			},
		},
	}
}

func copyMap(src map[string]string) map[string]string {
	cp := make(map[string]string)
	for k, v := range src {
		cp[k] = v
	}
	return cp
}
