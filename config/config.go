package config

import (
	"fmt"
	"github.com/acorn-io/baaah/pkg/router"
	v1 "github.com/hobbyfarm/gargantua/pkg/apis/hobbyfarm.io/v1"
	"github.com/hobbyfarm/hf-provisioner-shared/errors"
	namespace "github.com/hobbyfarm/hf-provisioner-shared/namespace"
	errors2 "k8s.io/apimachinery/pkg/api/errors"
	kclient "sigs.k8s.io/controller-runtime/pkg/client"
)

func ResolveConfigItemName(vmName string, req router.Request, item string) (string, error) {
	machine := &v1.VirtualMachine{}
	err := req.Client.Get(req.Ctx, kclient.ObjectKey{
		Namespace: namespace.ResolveNamespace(),
		Name:      vmName,
	}, machine)
	if err != nil {
		if errors2.IsNotFound(err) {
			return "", errors.NewNotFoundError("vm %s not found", vmName)
		}

		return "", fmt.Errorf("error while looking up machine with name %s: %s", vmName, err.Error())
	}

	return ResolveConfigItem(machine, req, item)
}

func ResolveConfigItem(obj *v1.VirtualMachine, req router.Request, item string) (string, error) {
	// go from most to least specific
	env := &v1.Environment{}
	err := req.Client.Get(req.Ctx, kclient.ObjectKey{
		Namespace: obj.Namespace,
		Name:      obj.Status.EnvironmentId,
	}, env)

	if err != nil {
		if errors2.IsNotFound(err) {
			return "", errors.NewNotFoundError("environment %s not found", obj.Status.EnvironmentId)
		}
		return "", fmt.Errorf("error while looking up environment for config key %s: %s", item, err.Error())
	}

	// first, check specifics for the template
	if val, ok := env.Spec.TemplateMapping[obj.Spec.VirtualMachineTemplateId][item]; ok {
		return val, nil
	}

	// if its not there, check the environment specs
	if val, ok := env.Spec.EnvironmentSpecifics[item]; ok {
		return val, nil
	}

	return "", fmt.Errorf("could not resolve config item with key %s", item)
}
