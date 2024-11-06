package instanceid

import (
	"context"
	"fmt"
	"github.com/hobbyfarm/hf-provisioner-shared/namespace"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// GetInstanceId attempts to retrieve a HobbyFarm installation's instance id.
// This is defined in a ConfigMap called "instance-id", with a key of "instance-id"
// On success, returns the string instance id, or "", err
func GetInstanceId(ctx context.Context, kclient client.Client) (string, error) {
	// Attempt to get instance id from namespace
	cfg := &v1.ConfigMap{}

	err := kclient.Get(ctx, client.ObjectKey{
		Name:      "instance-id",
		Namespace: namespace.ResolveNamespace(),
	}, cfg)
	if err != nil {
		return "", err
	}

	if val, ok := cfg.Data["instance-id"]; !ok {
		return "", fmt.Errorf("key instance-id not found in instance configmap")
	} else {
		return val, nil
	}
}
