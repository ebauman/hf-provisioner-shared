package controller

import (
	"context"
	"github.com/acorn-io/baaah"
	"github.com/acorn-io/baaah/pkg/restconfig"
	"github.com/acorn-io/baaah/pkg/router"
	"github.com/ebauman/crder"
	v1 "github.com/hobbyfarm/gargantua/pkg/apis/hobbyfarm.io/v1"
	labels2 "github.com/hobbyfarm/hf-provisioner-shared/labels"
	namespace "github.com/hobbyfarm/hf-provisioner-shared/namespace"
	"github.com/hobbyfarm/hf-provisioner-shared/provider"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/rest"
)

type Controller struct {
	Router     *router.Router
	Scheme     *runtime.Scheme
	restconfig *rest.Config
	Provider   provider.Provider
}

func NewController(provider provider.Provider) (*Controller, error) {
	scheme := runtime.NewScheme()

	cfg, err := restconfig.New(scheme)

	if err != nil {
		return nil, err
	}

	for _, ra := range provider.SchemeAdders() {
		utilruntime.Must(ra(scheme))
	}

	baseRouter, err := baaah.NewRouter(provider.Name(), &baaah.Options{
		Scheme:            scheme,
		DefaultRESTConfig: cfg,
		DefaultNamespace:  namespace.ResolveNamespace(),
	})

	providerRouter := registerProviderRouter(baseRouter, namespace.ResolveNamespace(), provider.Name())

	if err != nil {
		return nil, err
	}

	for _, ra := range provider.RouteAdders() {
		if err := ra(providerRouter); err != nil {
			return nil, err
		}
	}

	return &Controller{
		Router:     baseRouter,
		Scheme:     scheme,
		restconfig: cfg,
	}, nil
}

func (c *Controller) Start(ctx context.Context) error {
	crds := c.Provider.CRDs()

	if err := crder.InstallUpdateCRDs(c.restconfig, crds...); err != nil {
		return err
	}

	return c.Router.Start(ctx)
}

func registerProviderRouter(router *router.Router, ns string, providerName string) router.RouteBuilder {
	return router.Type(&v1.VirtualMachine{}).Namespace(ns).Selector(labels.SelectorFromSet(map[string]string{
		labels2.ProvisionerLabel: providerName,
	}))
}
