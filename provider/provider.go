package provider

import (
	"github.com/acorn-io/baaah/pkg/log"
	"github.com/acorn-io/baaah/pkg/router"
	"github.com/ebauman/crder"
	"k8s.io/apimachinery/pkg/runtime"
)

// RouteAdder is a function implemented by a provider to register
// Baaah routes for controller operation
type RouteAdder func(router *router.Router) error

// SchemeAdder is a function implemented by a provider to register
// types to a runtime scheme.
type SchemeAdder func(scheme *runtime.Scheme) error

// Provider is an instance of a hobbyfarm machine provider
type Provider interface {
	// Name is the name of the provider. Should return a short string uniqely identifying
	// the provider. This short string should identify the type of provider,
	// e.g. "aws" or "digitalocean", not a specific _instance_ of a provider
	Name() string

	// RouteAdders returns a slice of route adder functions. It is a slice
	// such that a provider could break up their route registrations into
	// multiple funcs
	RouteAdders() []RouteAdder

	// SchemeAdders returns a slice of scheme adder functions. It is a slice
	// such that a provider could register multiple apigroups / types with
	// the runtime scheme
	SchemeAdders() []SchemeAdder

	// CRDs returns a slice of CRD definitions to be installed by the controller
	CRDs() []crder.CRD

	// Logger returns the standard logger to be used for this controller
	Logger() log.Logger
}
