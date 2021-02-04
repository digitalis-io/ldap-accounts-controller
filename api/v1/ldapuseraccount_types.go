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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// LdapUserAccountSpec defines the desired state of LdapUserAccount
type LdapUserAccountSpec struct {
	// Important: Run "make" to regenerate code after modifying this file

	Username string `json:"username"`
	Uid      string `json:"uid"`
	Password string `json:"password"`
}

// LdapUserAccountStatus defines the observed state of LdapUserAccount
type LdapUserAccountStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true

// LdapUserAccount is the Schema for the ldapuseraccounts API
type LdapUserAccount struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   LdapUserAccountSpec   `json:"spec,omitempty"`
	Status LdapUserAccountStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// LdapUserAccountList contains a list of LdapUserAccount
type LdapUserAccountList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []LdapUserAccount `json:"items"`
}

func init() {
	SchemeBuilder.Register(&LdapUserAccount{}, &LdapUserAccountList{})
}
