package v1

type IngressRule struct {
	Host             string `json:"host,omitempty" protobuf:"bytes,1,opt,name=host"`
	IngressRuleValue `json:",inline,omitempty" protobuf:"bytes,2,opt,name=ingressRuleValue"`
}

type IngressRuleValue struct {
	HTTP *HTTPIngressRuleValue `json:"http,omitempty" protobuf:"bytes,1,opt,name=http"`
}

type HTTPIngressRuleValue struct {
	Paths []HTTPIngressPath `json:"paths" protobuf:"bytes,1,rep,name=paths"`
}

type PathType string

type HTTPIngressPath struct {
	Path     string         `json:"path,omitempty" protobuf:"bytes,1,opt,name=path"`
	PathType *PathType      `json:"pathType" protobuf:"bytes,3,opt,name=pathType"`
	Backend  IngressBackend `json:"backend" protobuf:"bytes,2,opt,name=backend"`
}

type IngressBackend struct {
	Service *IngressServiceBackend `json:"service,omitempty" protobuf:"bytes,4,opt,name=service"`
}

type IngressServiceBackend struct {
	Name   string             `json:"name" protobuf:"bytes,1,opt,name=name"`
	Port   ServiceBackendPort `json:"port,omitempty" protobuf:"bytes,2,opt,name=port"`
	Weight *int32             `json:"weight,omitempty" protobuf:"bytes,3,opt,name=weight"`
}

type ServiceBackendPort struct {
	Name   string `json:"name,omitempty" protobuf:"bytes,1,opt,name=name"`
	Number int32  `json:"number,omitempty" protobuf:"bytes,2,opt,name=number"`
}
