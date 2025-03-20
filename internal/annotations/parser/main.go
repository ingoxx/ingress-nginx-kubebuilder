package parser

import (
	"fmt"
	ingressv1 "github.com/ingoxx/ingress-nginx-kubebuilder/api/v1"
	kerr "github.com/ingoxx/ingress-nginx-kubebuilder/internal/errors"
	"strconv"
	"strings"
)

const (
	DefaultAnnotationsPrefix = "ingress.nginx.kubebuilder.io"
)

var (
	AnnotationsPrefix = DefaultAnnotationsPrefix
)

type AnnotationFields map[string]AnnotationConfig

type Annotation struct {
	Group       string
	Annotations AnnotationFields
}

type AnnotationConfig struct {
	Doc       string
	Validator AnnotationValidator
}

type IngressAnnotation interface {
	Parse(*ingressv1.Ingress) (interface{}, error)
	Validate(map[string]string) error
}

func TrimAnnotationPrefix(annotation string) string {
	return strings.TrimPrefix(annotation, AnnotationsPrefix+"/")
}

func GetAnnotationWithPrefix(suffix string) string {
	return fmt.Sprintf("%v/%v", AnnotationsPrefix, suffix)
}

type ingAnnotations map[string]string

func (a ingAnnotations) parseString(name string) (string, error) {
	val, ok := a[name]
	if ok {
		if val == "" {
			return "", kerr.NewInvalidContent(name, val)
		}

		return val, nil
	}

	return "", kerr.ErrMissingAnnotations
}

func (a ingAnnotations) parseStringSlice(name string) ([]string, error) {
	var data = make([]string, 0)
	val, ok := a[name]
	if ok {
		if val == "" {
			return data, kerr.NewInvalidContent(name, val)
		}

		return data, nil
	}

	return data, kerr.ErrMissingAnnotations
}

func (a ingAnnotations) parseBool(name string) (bool, error) {
	val, ok := a[name]
	if ok {
		b, err := strconv.ParseBool(val)
		if err != nil {
			return false, kerr.NewInvalidContent(name, val)
		}
		return b, nil
	}
	return false, kerr.ErrMissingAnnotations
}

func GetStringAnnotation(name string, ing *ingressv1.Ingress, field AnnotationFields) (string, error) {
	key, err := CheckAnnotationsKey(name, ing, field)
	if err != nil {
		return "", err
	}
	return ingAnnotations(ing.GetAnnotations()).parseString(key)
}

func GetBoolAnnotations(name string, ing *ingressv1.Ingress, field AnnotationFields) (bool, error) {
	key, err := CheckAnnotationsKey(name, ing, field)
	if err != nil {
		return false, err
	}
	return ingAnnotations(ing.GetAnnotations()).parseBool(key)
}
