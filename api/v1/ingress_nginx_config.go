package v1

type ParseIngressAnnotations interface {
	GetIngressAnnotations()
}

type Configuration struct {
	Servers []*Server `json:"servers"`
}

type Server struct {
	Name      string     `json:"name"`
	NameSpace string     `json:"name_space"`
	HostName  string     `json:"host_name"`
	Tls       SSLCert    `json:"tls"`
	Paths     []*Backend `json:"paths"`
}

type SSLCert struct {
	TlsKey    string `json:"tls-key"`
	TlsCrt    string `json:"tls-crt"`
	TlsNoPass bool   `json:"tls-no-pass"`
}

type Backend struct {
	Name           string                  `json:"name"`
	IngName        string                  `json:"ing_name"`
	NameSpace      string                  `json:"name_space"`
	Path           string                  `json:"path"`
	ServiceBackend *IngressServiceBackend  `json:"service_backend"`
	Port           int32                   `json:"port"`
	TargetPath     string                  `json:"target_path"`
	Annotations    ParseIngressAnnotations `json:"annotations"`
	RewritePath    string                  `json:"rewrite_path"`
	UpstreamName   string                  `json:"upstream_name"`
}
