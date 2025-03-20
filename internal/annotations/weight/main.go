package weight

import (
	"fmt"
	ingressv1 "github.com/ingoxx/ingress-nginx-kubebuilder/api/v1"
	"github.com/ingoxx/ingress-nginx-kubebuilder/internal/annotations/parser"
	"github.com/ingoxx/ingress-nginx-kubebuilder/internal/annotations/resolver"
	"github.com/ingoxx/ingress-nginx-kubebuilder/internal/errors"
	"k8s.io/klog/v2"
	"strconv"
)

const (
	useWeightAnnotation = "use-weight"
	useLbAnnotation     = "use-lb"
	lbPolicyAnnotation  = "lb-policy"
)

var weightAnnotation = parser.Annotation{
	Group: "allowCos",
	Annotations: parser.AnnotationFields{
		useWeightAnnotation: {
			Doc: "use weight, e.g: ` true or false`, optional",
		},
		useLbAnnotation: {
			Doc: "use Load balancing, e.g: ` true or false`, required",
		},
		lbPolicyAnnotation: {
			Doc: "lb policy, default: Round Robin, e.g: `least_conn,ip_hash,Random...`, required",
		},
	},
}

type Translate struct {
	Data string
}

type BackendWeight struct {
	UseWeight bool           `json:"use-weight"`
	UseLb     bool           `json:"use-lb"`
	LbPolicy  string         `json:"lb-policy"`
	Up        []UpstreamList `json:"up"`
}

type UpstreamList struct {
	SvcList  []string `json:"svc-list"`
	Upstream string   `json:"upstream"`
}

type weight struct {
	r resolver.Resolver
}

func NewParser(r resolver.Resolver) parser.IngressAnnotation {
	return &weight{
		r: r,
	}
}

func (r *weight) Parse(ing *ingressv1.Ingress) (interface{}, error) {
	var err error
	bw := &BackendWeight{}

	bw.UseLb, err = parser.GetBoolAnnotations(useLbAnnotation, ing, weightAnnotation.Annotations)
	if err != nil {
		if errors.IsValidationError(err) {
			klog.Warningf("%s is invalid, defaulting to false", useLbAnnotation)
		}
	}

	bw.UseWeight, err = parser.GetBoolAnnotations(useWeightAnnotation, ing, weightAnnotation.Annotations)
	if err != nil {
		if errors.IsValidationError(err) {
			klog.Warningf("%s is invalid, defaulting to false", useWeightAnnotation)
		}
	}

	bw.LbPolicy, err = parser.GetStringAnnotation(lbPolicyAnnotation, ing, weightAnnotation.Annotations)
	if err != nil {
		if errors.IsValidationError(err) {
			klog.Warningf("%s is invalid, defaulting to empty", lbPolicyAnnotation)
		}
	}

	if bw.UseLb {
		if err = r.validateLbBackend(ing, bw); err != nil {
			return nil, err
		}
	}

	return bw, nil
}

func (r *weight) validateLbBackend(ing *ingressv1.Ingress, config *BackendWeight) error {
	rules := ing.Spec.Rules
	for _, rs := range rules {
		var path string
		var wt string
		var upstreamName string
		var sl UpstreamList

		if len(rs.HTTP.Paths) < 1 {
			msg := fmt.Sprintf("at least two services are required to use the traffic allocation function, ingress name %s", ing.Name)
			return errors.NewNotSatisfiableError(msg)
		}

		for _, p := range rs.HTTP.Paths {
			if path == "" {
				path = p.Path
			}

			if path != p.Path {
				msg := fmt.Sprintf("when annotation %s is true, the path field of ingress must be the same, ingress name %s", useWeightAnnotation, ing.Name)
				return errors.NewNotSatisfiableError(msg)
			}

			svcPort := r.r.GetSvcPort(p.Backend)
			if svcPort == nil {
				return errors.NewMissResourcesError(p.Backend.Service.Name)
			}

			if upstreamName == "" {
				upstreamName = ing.Name + "-" + ing.Namespace + "-" + p.Backend.Service.Name
			} else {
				upstreamName += "-" + p.Backend.Service.Name
			}
			if config.UseWeight {
				if p.Backend.Service.Weight == nil {
					return errors.NewInvalidContent("weight", p.Backend.Service.Weight)
				}
				if r.inspectWeightVal(*p.Backend.Service.Weight).Data != "" {
					wt = r.inspectWeightVal(*p.Backend.Service.Weight).Data
				} else {
					wt = strconv.Itoa(int(*p.Backend.Service.Weight))
				}

				sl.SvcList = append(sl.SvcList, fmt.Sprintf("%s.%s.svc:%d weight=%s", p.Backend.Service.Name, ing.Namespace, *svcPort, wt))
			} else {
				sl.SvcList = append(sl.SvcList, fmt.Sprintf("%s.%s.svc:%d", p.Backend.Service.Name, ing.Namespace, *svcPort))
			}
		}
		sl.Upstream = upstreamName
		config.Up = append(config.Up, sl)
	}

	return nil
}

func (r *weight) Validate(ans map[string]string) error {
	return parser.CheckAnnotations(ans, weightAnnotation.Annotations)
}

func (r *weight) inspectWeightVal(weight int32) Translate {
	var t = Translate{}
	if weight == 0 {
		t.Data = "down"
	}

	return t
}
