package resolver

import (
	ingressv1 "github.com/ingoxx/ingress-nginx-kubebuilder/api/v1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Resolver interface {
	GetSecret(client.ObjectKey) (*corev1.Secret, error)
	GetDefaultService() (*corev1.Service, error)
	GetService(string) (*corev1.Service, error)
	GetHostName() []string
	GetSvcPort(interface{}) *int32
	GetTlsData(client.ObjectKey) (map[string][]byte, error)
	GetUpstreamName([]ingressv1.HTTPIngressPath, interface{}) string
}
