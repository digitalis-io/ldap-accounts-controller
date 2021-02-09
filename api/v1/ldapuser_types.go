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

// LdapUserSpec defines the desired state of LdapUser
type LdapUserSpec struct {
	Username string `json:"username"`
	UID      string `json:"uid"`
	GID      string `json:"gid"`
	Password string `json:"password"`
	Homedir  string `json:"homedir,omitempty"`
	Shell    string `json:"shell,omitempty"`
}

// LdapUserStatus defines the observed state of LdapUser
type LdapUserStatus struct {
	CreatedOn string `json:"createdOn,omitempty"`
	UpdatedOn string `json:"updatedOn,omitempty"`
}

// +kubebuilder:object:root=true

// LdapUser is the Schema for the ldapusers API
type LdapUser struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   LdapUserSpec   `json:"spec,omitempty"`
	Status LdapUserStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// LdapUserList contains a list of LdapUser
type LdapUserList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []LdapUser `json:"items"`
}

func init() {
	SchemeBuilder.Register(&LdapUser{}, &LdapUserList{})
}
