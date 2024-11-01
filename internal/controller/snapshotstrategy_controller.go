/*
Copyright 2024.

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

package controller

import (
	"context"
	"fmt"
	"os"
	"time"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	kindarocksv1alpha1 "github.com/db-operator/backup-operator/api/v1alpha1"
)

// SnapshotStrategyReconciler reconciles a SnapshotStrategy object
type SnapshotStrategyReconciler struct {
	client.Client
	Scheme        *runtime.Scheme
	Interval      time.Duration
	ScriptsFolder string
}

// +kubebuilder:rbac:groups=kinda.rocks,resources=snapshotstrategies,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=kinda.rocks,resources=snapshotstrategies/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=kinda.rocks,resources=snapshotstrategies/finalizers,verbs=update

// This controller is supposed to check whether scripts, that are listed
// in the SnapshotStrategy CR exist and are executable.
func (r *SnapshotStrategyReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	reconcilePeriod := r.Interval * time.Second
	reconcileResult := reconcile.Result{RequeueAfter: reconcilePeriod}
	// Fetch the Database custom resource
	snapsthotStrategyCr := &kindarocksv1alpha1.SnapshotStrategy{}
	err := r.Get(ctx, req.NamespacedName, snapsthotStrategyCr)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			// Requested object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcileResult, nil
		}
		// Error reading the object - requeue the request.
		return reconcileResult, err
	}

	// Update object status always when function exit abnormally or through a panic.
	defer func() {
		if err := r.Status().Update(ctx, snapsthotStrategyCr); err != nil {
			log.Error(err, "failed to update status")
		}
	}()

	// All the scripts (or other tools) must be stored in the
	// ScriptFolder.
	if err := r.VerifyScript(ctx, snapsthotStrategyCr.Spec.MysqlDumpScript); err != nil {
		log.Error(err, "mysql script is not valid")
		return reconcileResult, err
	}

	if err := r.VerifyScript(ctx, snapsthotStrategyCr.Spec.PostgresDumpScript); err != nil {
		log.Error(err, "postgres script is not valid")
		return reconcileResult, err
	}

	snapsthotStrategyCr.Status.ScriptsVerified = true
	return ctrl.Result{}, nil
}

func (r *SnapshotStrategyReconciler) VerifyScript(ctx context.Context, script string) error {
	log := log.FromContext(ctx)
	if len(script) > 0 {
		fullPath := fmt.Sprintf("%s/%s", r.ScriptsFolder, script)
		file, err := os.Stat(fullPath)
		if err != nil {
			return err
		}
		if file.Mode() != 493 {
			return fmt.Errorf("%s is not executable, current mode is %d", script, file.Mode())
		}
	}

	log.Info("script is not set, skipping")
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *SnapshotStrategyReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&kindarocksv1alpha1.SnapshotStrategy{}).
		Named("snapshotstrategy").
		Complete(r)
}
