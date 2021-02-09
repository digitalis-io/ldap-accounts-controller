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

	ldapv1 "ldap-accounts-controller/api/v1"
	ld "ldap-accounts-controller/ldap"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
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
		// Check if it was deleted and ignore
		if apierrors.IsNotFound(err) {
			// Delete fom LDAP
			// FIXME: does not work
			err := ld.LdapDeleteUser(ldapuser.Spec)
			if err != nil {
				log.Error(err, "Could not delete user from ldap")
			}
			return ctrl.Result{}, nil
		}
		log.Error(err, "unable to fetch ldap user")
		return ctrl.Result{}, err
	}

	var existsInLdap bool
	ldapEntry, err := ld.LdapGet("username", ldapuser.Spec.Username)
	if err != nil && ldapEntry.Username != "" {
		existsInLdap = false
	} else {
		existsInLdap = true
	}

	if !existsInLdap {
		err := ld.LdapAddUser(ldapuser.Spec)
		if err != nil {
			log.Error(err, "cannot add user to ldap")
		}
	}

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

func (r *LdapUserReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if err := mgr.GetFieldIndexer().IndexField(&ldapv1.LdapUser{}, ldapUserOwnerKey, func(rawObj runtime.Object) []string {
		acc := rawObj.(*ldapv1.LdapUser)
		return []string{acc.Name}
	}); err != nil {
		return err
	}
	return ctrl.NewControllerManagedBy(mgr).
		For(&ldapv1.LdapUser{}).
		Complete(r)
}
