package controller

import (
	"context"
	"github.com/acorn-io/baaah"
	"github.com/acorn-io/baaah/pkg/log"
	"github.com/acorn-io/baaah/pkg/restconfig"
	"github.com/acorn-io/baaah/pkg/router"
	"github.com/ebauman/crder"
	namespace "github.com/hobbyfarm/hf-provisioner-shared/namespace"
	"github.com/hobbyfarm/hf-provisioner-shared/provider"
	corev1 "k8s.io/api/core/v1"
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

	// give everything corev1 at least
	utilruntime.Must(corev1.AddToScheme(scheme))

	baseRouter, err := baaah.NewRouter(provider.Name(), &baaah.Options{
		Scheme:            scheme,
		DefaultRESTConfig: cfg,
		DefaultNamespace:  namespace.ResolveNamespace(),
	})

	if err != nil {
		return nil, err
	}

	for _, ra := range provider.RouteAdders() {
		if err := ra(baseRouter); err != nil {
			return nil, err
		}
	}

	if provider.Logger() != nil {
		log.SetLogger(provider.Logger())
	}

	return &Controller{
		Router:     baseRouter,
		Scheme:     scheme,
		restconfig: cfg,
		Provider:   provider,
	}, nil
}

func (c *Controller) Start(ctx context.Context) error {
	crds := c.Provider.CRDs()

	if err := crder.InstallUpdateCRDs(c.restconfig, crds...); err != nil {
		return err
	}

	return c.Router.Start(ctx)
}
