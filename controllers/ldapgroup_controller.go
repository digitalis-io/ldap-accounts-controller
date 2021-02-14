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
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	ldapv1 "ldap-accounts-controller/api/v1"
	ld "ldap-accounts-controller/ldap"
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

	//! [finalizer]
	ldapgroupFinalizerName := "ldap.digitalis.io/finalizer"
	if ldapgroup.ObjectMeta.DeletionTimestamp.IsZero() {
		if !containsString(ldapgroup.GetFinalizers(), ldapgroupFinalizerName) {
			ldapgroup.SetFinalizers(append(ldapgroup.GetFinalizers(), ldapgroupFinalizerName))
			if err := r.Update(context.Background(), &ldapgroup); err != nil {
				return ctrl.Result{}, err
			}
		}
	} else {
		// The object is being deleted
		if containsString(ldapgroup.GetFinalizers(), ldapgroupFinalizerName) {
			// our finalizer is present, so lets handle any external dependency
			if err := ld.DeleteGroup(ldapgroup.Spec); err != nil {
				log.Error(err, "Error deleting from LDAP")
				return ctrl.Result{}, err
			}

			// remove our finalizer from the list and update it.
			ldapgroup.SetFinalizers(removeString(ldapgroup.GetFinalizers(), ldapgroupFinalizerName))
			if err := r.Update(context.Background(), &ldapgroup); err != nil {
				return ctrl.Result{}, err
			}
		}

		// Stop reconciliation as the item is being deleted
		return ctrl.Result{}, nil
	}
	//! [finalizer]

	log.Info("Adding or updating LDAP group")
	err := ld.AddGroup(ldapgroup.Spec)
	if err != nil {
		log.Error(err, "cannot add group to ldap")
	}
	ldapgroup.Status.CreatedOn = time.Now().Format("2006-01-02 15:04:05")

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
	//! [pred]
	pred := predicate.Funcs{
		CreateFunc:  func(event.CreateEvent) bool { return true },
		DeleteFunc:  func(event.DeleteEvent) bool { return false },
		GenericFunc: func(event.GenericEvent) bool { return true },
		UpdateFunc: func(e event.UpdateEvent) bool {
			oldGeneration := e.MetaOld.GetGeneration()
			newGeneration := e.MetaNew.GetGeneration()
			// Generation is only updated on spec changes (also on deletion),
			// not metadata or status
			// Filter out events where the generation hasn't changed to
			// avoid being triggered by status updates
			return oldGeneration != newGeneration
		},
	}
	//! [pred]
	return ctrl.NewControllerManagedBy(mgr).
		For(&ldapv1.LdapGroup{}).
		WithEventFilter(pred).
		Complete(r)
}
