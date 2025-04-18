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
	netv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// IngressSpec defines the desired state of Ingress
type IngressSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	// +optional
	IngressClassName string `json:"ingressClassName,omitempty" protobuf:"bytes,4,opt,name=ingressClassName"`
	// +optional
	DefaultBackend *IngressBackend `json:"defaultBackend,omitempty" protobuf:"bytes,1,opt,name=defaultBackend"`
	// When an ingress instance is created, the corresponding Secret resource will be automatically
	// +optional
	TLS []netv1.IngressTLS `json:"tls,omitempty" protobuf:"bytes,2,rep,name=tls"`
	// +optional
	Rules []IngressRule `json:"rules,omitempty" protobuf:"bytes,3,rep,name=rules"`
}

// IngressStatus defines the observed state of Ingress
type IngressStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Ingress is the Schema for the ingresses API
type Ingress struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   IngressSpec   `json:"spec,omitempty"`
	Status IngressStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// IngressList contains a list of Ingress
type IngressList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Ingress `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Ingress{}, &IngressList{})
}
