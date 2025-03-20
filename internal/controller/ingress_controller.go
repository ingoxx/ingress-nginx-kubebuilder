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

package controller

import (
	"context"
	"fmt"
	ingressv1 "github.com/ingoxx/ingress-nginx-kubebuilder/api/v1"
	"github.com/ingoxx/ingress-nginx-kubebuilder/internal/annotations"
	"github.com/ingoxx/ingress-nginx-kubebuilder/internal/config"
	"github.com/ingoxx/ingress-nginx-kubebuilder/internal/controller/store"
	"github.com/ingoxx/ingress-nginx-kubebuilder/internal/nginx"
	"github.com/ingoxx/ingress-nginx-kubebuilder/pkg/resources"
	"github.com/ingoxx/ingress-nginx-kubebuilder/pkg/utils/template_nginx"
	v1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
	"os"
	"path/filepath"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
	"time"
)

// IngressReconciler reconciles a Ingress object
type IngressReconciler struct {
	client.Client
	Scheme        *runtime.Scheme
	dynamicClient *dynamic.DynamicClient
	ctx           context.Context
	ingress       *ingressv1.Ingress
}

//+kubebuilder:rbac:groups=ingress.nginx.kubebuilder.io,resources=ingresses,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=ingress.nginx.kubebuilder.io,resources=ingresses/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=ingress.nginx.kubebuilder.io,resources=ingresses/finalizers,verbs=update
//+kubebuilder:rbac:groups=cert-manager.io,resources=certificates,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=cert-manager.io,resources=issuers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Ingress object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.17.2/pkg/reconcile
func (r *IngressReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	//logger := log.FromContext(ctx)

	// TODO(user): your logic here
	var ic = new(ingressv1.Ingress)

	if err := r.Get(ctx, req.NamespacedName, ic); err != nil {
		klog.Infof("ingress resource %s not found in namesapce %s, maybe has been deleted", req.NamespacedName.Name, req.NamespacedName.Namespace)
		r.clearConf(req.NamespacedName)
		return ctrl.Result{}, nil
	}

	//if !ic.ObjectMeta.DeletionTimestamp.IsZero() {
	//	klog.Infof("ingress resource %s has been deleted in namesapce %s", req.NamespacedName.Name, req.NamespacedName.Namespace)
	//	r.clearConf(req.NamespacedName)
	//	return ctrl.Result{}, nil
	//}

	r.ctx = ctx
	r.ingress = ic

	if info := r.checkController(); info != nil {
		return ctrl.Result{RequeueAfter: time.Second * time.Duration(30)}, nil
	}

	var key client.ObjectKey
	if ic.Spec.DefaultBackend != nil {
		key = types.NamespacedName{Name: ic.Spec.DefaultBackend.Service.Name, Namespace: ic.Namespace}
		if err := r.checkService(key); err != nil {
			return ctrl.Result{RequeueAfter: time.Second * time.Duration(30)}, nil
		}
	}

	var errList []error
	if len(ic.Spec.Rules) > 0 {
		for _, v := range ic.Spec.Rules {
			for _, h := range v.HTTP.Paths {
				key = types.NamespacedName{Name: h.Backend.Service.Name, Namespace: ic.Namespace}
				if err := r.checkService(key); err != nil {
					errList = append(errList, err)
				}
			}
		}
	}

	if len(errList) > 0 {
		return ctrl.Result{RequeueAfter: time.Second * time.Duration(30)}, nil
	}

	rs := r.GetReconcileInfo()
	rs.DynamicClientSet = r.dynamicClient
	rs.IngressInfos = store.NewIngressInfo(rs)

	if err := resources.ReconcileResource(rs); err != nil {
		return ctrl.Result{RequeueAfter: time.Second * time.Duration(15)}, nil
	}

	parsed, err := annotations.NewAnnotationExtractor(rs.IngressInfos).Extract(ic)
	if err != nil {
		klog.ErrorS(err, fmt.Sprintf("fail to parse annotations in ingress: %s, namespace: %s", req.Name, req.Namespace))
		return ctrl.Result{RequeueAfter: time.Second * time.Duration(15)}, nil
	}

	var ings = annotations.IngressAnnotations{
		ParsedAnnotations: parsed,
	}

	if err := NewNginxController(rs).GenerateConfigure(ings); err != nil {
		klog.ErrorS(err, fmt.Sprintf("error in ingress: %s, namespace: %s", req.Name, req.Namespace))
		return ctrl.Result{RequeueAfter: time.Second * time.Duration(15)}, nil
	}

	return ctrl.Result{}, nil
}

func (r *IngressReconciler) GetReconcileInfo() *store.IngressReconciler {
	si := &store.IngressReconciler{
		Client:  r.Client,
		Scheme:  r.Scheme,
		Ingress: r.ingress,
		Context: r.ctx,
	}

	return si
}

func (r *IngressReconciler) checkService(key client.ObjectKey) error {
	svc := new(v1.Service)
	rs := r.GetReconcileInfo()

	if err := r.Get(rs.Context, key, svc); err != nil {
		if errors.IsNotFound(err) {
			klog.ErrorS(err, fmt.Sprintf("no service with name %v found in namespace %v: %v", key.Name, key.Namespace, err))
			return fmt.Errorf("no svc resource available")

		}
		klog.ErrorS(err, fmt.Sprintf("unexpected error searching service with name %v in namespace %v: %v", key.Name, key.Namespace, err))
		return fmt.Errorf("unable to fetch svc resource")
	}

	return nil
}

func (r *IngressReconciler) createDynamicClientSet() *dynamic.DynamicClient {
	configSet, err := clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
	if err != nil {
		inClusterConfig, err := rest.InClusterConfig()
		if err != nil {
			klog.Fatalf("fail to create clusterConfig: %v", err)
		}
		configSet = inClusterConfig
	}

	dynamicClient, err := dynamic.NewForConfig(configSet)
	if err != nil {
		klog.Fatalf("fail to create dynamicClient: %v", err)
	}

	return dynamicClient
}

func (r *IngressReconciler) checkController() error {
	ic := new(netv1.IngressClass)
	getAnnotations := r.ingress.GetAnnotations()
	if r.ingress.Spec.IngressClassName == "" && getAnnotations[nginxAnnotationKey] == "" {
		klog.Infoln("the current controller can be used by adding ingressClass or annotating specified values")
		return fmt.Errorf("select available ingress nginx controller")
	}

	if r.ingress.Annotations[nginxAnnotationKey] == nginxAnnotationVal {
		return nil
	}

	key := types.NamespacedName{Name: r.ingress.Spec.IngressClassName, Namespace: r.ingress.Namespace}
	if err := r.Get(r.ctx, key, ic); err != nil {
		return err
	}

	if ic.Spec.Controller != controller {
		klog.Infoln("neither ingressClass nor nginxAnnotationVal value matches the current controller")
		return fmt.Errorf("pls select available ingress nginx controller")
	}

	return nil
}

func (r *IngressReconciler) clearConf(key client.ObjectKey) {
	conf := filepath.Join(config.ConfDir, key.Name+"-"+key.Namespace+".conf")
	if _, err := os.Stat(conf); err != nil {
		defaultConf := strings.Split(config.MainConf, ".")
		pr := &template_nginx.RenderTemplate{
			GenerateName:       defaultConf[0],
			RenderTemplateName: config.DefaultTmpl,
			MainTemplateName:   config.NginxTmpl,
		}

		if err := NewConfHandler().UpdateDefaultConf(pr); err != nil {
			return
		}

		return
	}

	nginx.CleanConf(conf)
}

// SetupWithManager sets up the controller with the Manager.
func (r *IngressReconciler) SetupWithManager(mgr ctrl.Manager) error {
	go nginx.Start()
	r.dynamicClient = r.createDynamicClientSet()
	return ctrl.NewControllerManagedBy(mgr).
		For(&ingressv1.Ingress{}).
		Complete(r)
}
