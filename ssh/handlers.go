package ssh

import (
	"fmt"
	"github.com/acorn-io/baaah/pkg/router"
	v1 "github.com/hobbyfarm/gargantua/pkg/apis/hobbyfarm.io/v1"
	config "github.com/hobbyfarm/hf-provisioner-shared/config"
	"github.com/hobbyfarm/hf-provisioner-shared/errors"
	labels2 "github.com/hobbyfarm/hf-provisioner-shared/labels"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	kclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// RequireSecret is a handler that requires a secret matching a virtual machine label to be present.
// It is meant to be used as a Middleware in Baaah
func RequireSecret(next router.Handler) router.Handler {
	return router.HandlerFunc(func(req router.Request, resp router.Response) error {
		vm := req.Object.(*v1.VirtualMachine)

		secretList := &corev1.SecretList{}
		err := req.List(secretList, &kclient.ListOptions{
			Namespace: vm.Namespace,
			LabelSelector: labels.SelectorFromSet(map[string]string{
				labels2.VirtualMachineLabel: vm.Name,
			}),
		})
		if err != nil {
			return err
		}

		if len(secretList.Items) == 0 {
			return nil
		}

		return next.Handle(req, resp)
	})
}

// SecretHandler creates an SSH secret with generated key for a virtualmachine.
// It will also look up a 'password' config value and add it if present.
// Secret is generic, keys are 'public_key', 'private_key', and 'password'
func SecretHandler(req router.Request, resp router.Response) error {
	vm := req.Object.(*v1.VirtualMachine)

	// try to get secret
	var secret *corev1.Secret
	var public, private string
	secret, err := GetSecret(req)
	if errors.IsNotFound(err) {
		secret = &corev1.Secret{}

		public, private, err = GenKeyPair()
		if err != nil {
			return err
		}

		secret.Data = map[string][]byte{}

		secret.Data["public_key"] = []byte(public)
		secret.Data["private_key"] = []byte(private)
	} else if err != nil {
		return err
	}

	secret.Name = fmt.Sprintf("%s-keys", vm.Name)

	if len(secret.Labels) == 0 {
		secret.Labels = map[string]string{}
	}

	secret.Labels[labels2.VirtualMachineLabel] = vm.Name
	secret.Namespace = vm.Namespace
	if password, err := config.ResolveConfigItem(vm, req, "password"); err == nil {
		secret.Data["password"] = []byte(password)
	}

	resp.Objects(secret)

	return nil
}

// GetSecret gets the first secret matching the virtual machine label
// If there are more than one secret found with this label, this method
// will only return the first from the List() request to kubernetes
func GetSecret(req router.Request) (*corev1.Secret, error) {
	// list secrets associated with the vm
	secretList := &corev1.SecretList{}
	err := req.List(secretList, &kclient.ListOptions{
		Namespace: req.Object.GetNamespace(),
		LabelSelector: labels.SelectorFromSet(map[string]string{
			labels2.VirtualMachineLabel: req.Object.GetName(),
		}),
	})
	if err != nil {
		return nil, err
	}

	if len(secretList.Items) > 0 {
		return &secretList.Items[0], nil
	}

	return nil, errors.NewNotFoundError("could not find secret for VirtualMachine %s", req.Object.GetName())
}
