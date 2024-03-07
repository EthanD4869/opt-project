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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// DistributeTrainSpec defines the desired state of DistributeTrain
type DistributeTrainSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of DistributeTrain. Edit distributetrain_types.go to remove/update
	Foo                string            `json:"foo,omitempty"`
	ServiceAccountName string            `json:"serviceAccountName,omitempty"`
	Env                []EnvSpec         `json:"env,omitempty"`
	Image              string            `json:"image,omitempty"`
	ImagePullPolicy    string            `json:"imagePullPolicy,omitempty"`
	MasterCmd          string            `json:"masterCmd,omitempty"`
	MasterResources    ResourceLimits    `json:"masterResources,omitempty"`
	NodeSelector       map[string]string `json:"nodeSelector,omitempty"`
	Size               int               `json:"size,omitempty"`
	SlaveCmd           string            `json:"slaveCmd,omitempty"`
	SlaveResources     ResourceLimits    `json:"slaveResources,omitempty"`
	Tolerations        []Toleration      `json:"tolerations,omitempty"`
	VolumeMounts       []VolumeMount     `json:"volumeMounts,omitempty"`
	Volumes            []Volume          `json:"volumes,omitempty"`
}

// EnvSpec defines the environment variable
type EnvSpec struct {
	Name  string `json:"name,omitempty"`
	Value string `json:"value,omitempty"`
}

// ResourceLimits defines the resource limits
type ResourceLimits struct {
	Limits ResourceValues `json:"limits,omitempty"`
}

// ResourceValues defines the CPU, memory, and GPU values
type ResourceValues struct {
	CPU       string `json:"cpu,omitempty"`
	Memory    string `json:"memory,omitempty"`
	NvidiaGPU string `json:"nvidia.com/gpu,omitempty"`
}

// Toleration defines the toleration values
type Toleration struct {
	// Add toleration fields as needed
}

// VolumeMount defines the volume mount configuration
type VolumeMount struct {
	MountPath string `json:"mountPath,omitempty"`
	Name      string `json:"name,omitempty"`
	ReadOnly  bool   `json:"readOnly,omitempty"`
}

// Volume defines the volume configuration
type Volume struct {
	HostPath HostPath `json:"hostPath,omitempty"`
	Name     string   `json:"name,omitempty"`
}

// HostPath defines the host path configuration
type HostPath struct {
	Path string `json:"path,omitempty"`
	Type string `json:"type,omitempty"`
}

// DistributeTrainStatus defines the observed state of DistributeTrain
type DistributeTrainStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// DistributeTrain is the Schema for the distributetrains API
type DistributeTrain struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DistributeTrainSpec   `json:"spec,omitempty"`
	Status DistributeTrainStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// DistributeTrainList contains a list of DistributeTrain
type DistributeTrainList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DistributeTrain `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DistributeTrain{}, &DistributeTrainList{})
}
