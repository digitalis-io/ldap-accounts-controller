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
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	ldapv1 "ldap-accounts-controller/api/v1"
	ld "ldap-accounts-controller/ldap"
)

var (
	ldapUserOwnerKey = ".metadata.controller"
)

// LdapUserReconciler reconciles a LdapUser object
type LdapUserReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=ldap.digitalis.io,resources=ldapusers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=ldap.digitalis.io,resources=ldapusers/status,verbs=get;update;patch

func (r *LdapUserReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("ldapuser", req.NamespacedName)

	var ldapuser ldapv1.LdapUser
	if err := r.Get(ctx, req.NamespacedName, &ldapuser); err != nil {
		//log.Error(err, "unable to fetch ldap user")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	//! [finalizer]
	ldapuserFinalizerName := "ldap.digitalis.io/finalizer"
	if ldapuser.ObjectMeta.DeletionTimestamp.IsZero() {
		if !containsString(ldapuser.GetFinalizers(), ldapuserFinalizerName) {
			ldapuser.SetFinalizers(append(ldapuser.GetFinalizers(), ldapuserFinalizerName))
			if err := r.Update(context.Background(), &ldapuser); err != nil {
				return ctrl.Result{}, err
			}
		}
	} else {
		// The object is being deleted
		if containsString(ldapuser.GetFinalizers(), ldapuserFinalizerName) {
			// our finalizer is present, so lets handle any external dependency
			if err := ld.LdapDeleteUser(ldapuser.Spec); err != nil {
				log.Error(err, "Error deleting from LDAP")
				return ctrl.Result{}, err
			}

			// remove our finalizer from the list and update it.
			ldapuser.SetFinalizers(removeString(ldapuser.GetFinalizers(), ldapuserFinalizerName))
			if err := r.Update(context.Background(), &ldapuser); err != nil {
				return ctrl.Result{}, err
			}
		}

		// Stop reconciliation as the item is being deleted
		return ctrl.Result{}, nil
	}
	//! [finalizer]

	log.Info("Adding or updating LDAP user")
	err := ld.LdapAddUser(ldapuser.Spec)
	if err != nil {
		log.Error(err, "cannot add user to ldap")
	}
	ldapuser.Status.CreatedOn = time.Now().Format("2006-01-02 15:04:05")

	var ldapUsers ldapv1.LdapUserList
	if err := r.List(ctx, &ldapUsers, client.InNamespace(req.Namespace), client.MatchingFields{ldapUserOwnerKey: req.Name}); err != nil {
		log.Error(err, "unable to list ldap accounts")
		return ctrl.Result{}, err
	}

	for _, acc := range ldapUsers.Items {
		msg := fmt.Sprintf("Checking user %s", acc.Spec.Username)
		log.Info(msg)
		acc.Status.CreatedOn = time.Now().Format("2006-01-02 15:04:05")
	}

	return ctrl.Result{}, nil
}

// Helper functions to check and remove string from a slice of strings.
func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

func removeString(slice []string, s string) (result []string) {
	for _, item := range slice {
		if item == s {
			continue
		}
		result = append(result, item)
	}
	return
}

func (r *LdapUserReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if err := mgr.GetFieldIndexer().IndexField(&ldapv1.LdapUser{}, ldapUserOwnerKey, func(rawObj runtime.Object) []string {
		acc := rawObj.(*ldapv1.LdapUser)
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
		For(&ldapv1.LdapUser{}).
		WithEventFilter(pred).
		Complete(r)
}
