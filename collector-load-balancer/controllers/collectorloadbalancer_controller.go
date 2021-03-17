/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"time"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	otelv1 "github.com/aws-observability/collector-load-balancer/api/v1"
	"github.com/aws-observability/collector-load-balancer/loadbalancer"
)

const (
	syncRetryInterval = time.Minute
)

// CollectorLoadBalancerReconciler reconciles a CollectorLoadBalancer object
type CollectorLoadBalancerReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme

	lb *loadbalancer.Server
}

// NOTE: this includes the rbc for both controller manager and prometheus service discovery
// TODO: need to add ingress, and nonResourceURLs /metrics

//+kubebuilder:rbac:groups=otel.k8s.aws,resources=collectorloadbalancers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=otel.k8s.aws,resources=collectorloadbalancers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=otel.k8s.aws,resources=collectorloadbalancers/finalizers,verbs=update
//+kubebuilder:rbac:groups="apps",resources=daemonsets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="apps",resources=statefulsets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=endpoints,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=nodes/metrics,verbs=get;list;watch

// Reconcile creates load balancer instance if not exists, and notify the instance there are some resource changing.
func (r *CollectorLoadBalancerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("collectorloadbalancer", req.NamespacedName)

	instanceName := loadbalancer.InstanceName{
		Namespace: req.Namespace,
		Name:      req.Name,
	}

	// Fetch
	var clb otelv1.CollectorLoadBalancer
	if err := r.Get(ctx, req.NamespacedName, &clb); err != nil {
		// Delete
		if apierrors.IsNotFound(err) {
			// Stop service discovery, sending targets to collectors etc.
			err = r.lb.RemoveInstance(ctx, instanceName)
			if err != nil {
				log.Error(err, "Remove existing instance failed")
				return ctrl.Result{Requeue: true, RequeueAfter: syncRetryInterval}, err
			}
			return ctrl.Result{}, nil
		}
		log.Error(err, "Fetch CollectorLoadBalancer failed")
		return ctrl.Result{Requeue: true, RequeueAfter: syncRetryInterval}, err
	}

	// Start/Fetch instance, which contains actual sync logic
	instance, err := r.lb.AddInstanceIfNotExists(clb)
	if err != nil {
		log.Error(err, "AddInstance failed")
		return ctrl.Result{Requeue: true, RequeueAfter: syncRetryInterval}, err
	}

	// Notify the instance it needs to reconcile because something (CRD, sts etc.) has changed.
	// The actual reconcile logic is async and may not execute immediately.
	// Returned error should be config or unexpected internal error.
	if err = instance.CRDReconcile(clb); err != nil {
		log.Error(err, "Sync on instance failed", "InstanceName", instanceName)
		return ctrl.Result{Requeue: true, RequeueAfter: syncRetryInterval}, err
	}

	return ctrl.Result{}, nil
}

var apiGVStr = otelv1.GroupVersion.String()

// SetupWithManager sets up the controller with the Manager.
func (r *CollectorLoadBalancerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// Create local index for sts owned by controller.
	err := mgr.GetFieldIndexer().IndexField(context.Background(), &appsv1.StatefulSet{}, loadbalancer.K8sControllerOwnerKey,
		func(rawObj client.Object) []string {
			sts := rawObj.(*appsv1.StatefulSet)
			owner := metav1.GetControllerOf(sts)
			if owner == nil {
				return nil
			}
			if owner.APIVersion != apiGVStr || owner.Kind != loadbalancer.K8sCRDKind {
				return nil
			}
			return []string{owner.Name}
		})
	if err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&otelv1.CollectorLoadBalancer{}).
		Owns(&appsv1.StatefulSet{}).
		Complete(r)
}

func (r *CollectorLoadBalancerReconciler) SetLB(lb *loadbalancer.Server) {
	r.lb = lb
}
