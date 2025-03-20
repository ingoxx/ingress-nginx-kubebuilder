package controller

import (
	ingressv1 "github.com/ingoxx/ingress-nginx-kubebuilder/api/v1"
	"github.com/ingoxx/ingress-nginx-kubebuilder/internal/nginx"
	"github.com/ingoxx/ingress-nginx-kubebuilder/pkg/utils/template_nginx"
)

type ConfHandler struct {
}

func NewConfHandler() ConfHandler {
	return ConfHandler{}
}

func (c ConfHandler) UpdateDefaultConf(parser *template_nginx.RenderTemplate) error {
	var servers = new(ingressv1.Server)
	var cfg = struct {
		Server *ingressv1.Server
	}{
		Server: servers,
	}

	if err := parser.Render(cfg); err != nil {
		return err
	}

	if err := nginx.Reload(parser.GenerateName); err != nil {
		return err
	}

	return nil
}
