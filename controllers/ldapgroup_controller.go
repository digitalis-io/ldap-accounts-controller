/*
Copyright 2021 Digitalis.IO.

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
	"fmt"
	"time"

	"github.com/go-logr/logr"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	ldapv1 "ldap-accounts-controller/api/v1"
)

// LdapGroupReconciler reconciles a LdapGroup object
type LdapGroupReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

var (
	ldapGroupOwnerKey = ".metadata.controller"
)

// +kubebuilder:rbac:groups=ldap.digitalis.io,resources=ldapgroups,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=ldap.digitalis.io,resources=ldapgroups/status,verbs=get;update;patch

func (r *LdapGroupReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("ldapgroup", req.NamespacedName)

	// your logic here
	var ldapgroup ldapv1.LdapGroup
	if err := r.Get(ctx, req.NamespacedName, &ldapgroup); err != nil {
		// Check if it was deleted and ignore
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		log.Error(err, "unable to fetch ldap group")
		return ctrl.Result{}, err
	}
	var ldapGroups ldapv1.LdapGroupList
	if err := r.List(ctx, &ldapGroups, client.InNamespace(req.Namespace), client.MatchingFields{ldapGroupOwnerKey: req.Name}); err != nil {
		log.Error(err, "unable to list ldap accounts")
		return ctrl.Result{}, err
	}

	for _, acc := range ldapGroups.Items {
		msg := fmt.Sprintf("Checking group %s", acc.Spec.Name)
		log.Info(msg)
		acc.Status.CreatedOn = time.Now().Format("2006-01-02 15:04:05")
	}

	return ctrl.Result{}, nil
}

func (r *LdapGroupReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if err := mgr.GetFieldIndexer().IndexField(&ldapv1.LdapGroup{}, ldapGroupOwnerKey, func(rawObj runtime.Object) []string {
		acc := rawObj.(*ldapv1.LdapGroup)
		return []string{acc.Name}
	}); err != nil {
		return err
	}
	return ctrl.NewControllerManagedBy(mgr).
		For(&ldapv1.LdapGroup{}).
		Complete(r)
}
