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

	"github.com/keikoproj/flippy/controllers/Reconciler"
	"github.com/keikoproj/flippy/pkg/k8s-utils/k8s"
	"github.com/keikoproj/flippy/pkg/k8s-utils/utils"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	k8sLog "sigs.k8s.io/controller-runtime/pkg/log"

	crdv1 "github.com/keikoproj/flippy/api/v1"
)

// FlippyReconciler reconciles a FlippyConfig object
type FlippyReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=webapp.my.domain,resources=guestbooks,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=webapp.my.domain,resources=guestbooks/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=webapp.my.domain,resources=guestbooks/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// Reconcile function compares the state specified by
// the FlippyConfig object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *FlippyReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = k8sLog.FromContext(ctx)

	result := crdv1.FlippyConfig{
		TypeMeta:   v1.TypeMeta{},
		ObjectMeta: v1.ObjectMeta{},
		Spec:       crdv1.FlippyConfigSpec{},
		Status:     crdv1.FlippyConfigStatus{},
	}

	err := r.Client.Get(ctx, req.NamespacedName, &result)
	if err != nil {
		log.Error("Failed to get CRD for ", req, " Error - ", err)
	}

	_, clientset, err := utils.Setup()

	if err != nil {
		return ctrl.Result{}, err
	}

	_, argoclientset, err := utils.SetupArgoRollout()

	if err != nil {
		log.Error("Failed to Setup Argo Rollout Client", err)
		argoclientset = nil
	}

	clientSetWrapper := k8s.ClientSet{K8sClientSet: clientset, ArgoRolloutClientSet: argoclientset}
	Reconciler.Process.Handle(result, clientSetWrapper, k8s.K8s)

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *FlippyReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&crdv1.FlippyConfig{}).
		Complete(r)
}
