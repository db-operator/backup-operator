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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DbBackupSpec defines the desired state of DbBackup.
type DbBackupSpec struct {
	SnapshotStrategy *StrategyRef `json:"snapshotStrategy"`
	StoreStrategy    *StrategyRef `json:"storeStrategy"`
	Target           *Target      `json:"target"`
}

type StrategyRef struct {
	Name    string `json:"name"`
	EnvFrom *From  `json:"envFrom"`
}

type From struct {
	Kind string `json:"kind"`
	Name string `json:"name"`
}

// Target should be a database definition
type Target struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

// DbBackupStatus defines the observed state of DbBackup.
type DbBackupStatus struct {
	Size     string `json:"size"`
	Uploaded bool   `json:"uploaded"`
	Queued   bool   `json:"queued"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// DbBackup is the Schema for the dbbackups API.
type DbBackup struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DbBackupSpec   `json:"spec,omitempty"`
	Status DbBackupStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// DbBackupList contains a list of DbBackup.
type DbBackupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DbBackup `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DbBackup{}, &DbBackupList{})
}
