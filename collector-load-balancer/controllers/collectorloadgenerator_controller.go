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

	apierrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	otelv1 "github.com/aws-observability/collector-load-balancer/api/v1"
)

// CollectorLoadGeneratorReconciler reconciles a CollectorLoadGenerator object
type CollectorLoadGeneratorReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=otel.k8s.aws,resources=collectorloadgenerators,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=otel.k8s.aws,resources=collectorloadgenerators/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=otel.k8s.aws,resources=collectorloadgenerators/finalizers,verbs=update

// Reconcile
// TODO: clg can have most logic inside the reconcile to make things simple
func (r *CollectorLoadGeneratorReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("collectorloadgenerator", req.NamespacedName)

	var clg otelv1.CollectorLoadGenerator
	if err := r.Get(ctx, req.NamespacedName, &clg); err != nil {
		if apierrors.IsNotFound(err) {
			// TODO: should wrap an experiment as done
			return ctrl.Result{}, nil
		}
		log.Error(err, "Unable to fetch CollectorBalancer")
		return ctrl.Result{Requeue: true, RequeueAfter: syncRetryInterval}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *CollectorLoadGeneratorReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&otelv1.CollectorLoadGenerator{}).
		Complete(r)
}
