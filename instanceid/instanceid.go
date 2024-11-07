package instanceid

import (
	"context"
	"fmt"
	"github.com/hobbyfarm/hf-provisioner-shared/namespace"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/uuid"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	HobbyfarmInstanceIdName = "hobbyfarm-instance-id"
	InstanceIdKey           = "instance-id"
)

// GetOrCreateInstanceId attempts first to get an existing hobbyfarm instance id.
// If one does not exist it creates it, stores it, and returns the value.
func GetOrCreateInstanceId(ctx context.Context, kclient client.Client) (string, error) {
	id, err := GetInstanceId(ctx, kclient)
	if err != nil {
		if !errors.IsNotFound(err) {
			return "", err
		}

		return createInstanceId(ctx, kclient)
	}

	return id, nil
}

// GetInstanceId attempts to retrieve a HobbyFarm installation's instance id.
// This is defined in a ConfigMap called "instance-id", with a key of "instance-id"
// On success, returns the string instance id, or "", err
func GetInstanceId(ctx context.Context, kclient client.Client) (string, error) {
	// Attempt to get instance id from namespace
	cfg := &v1.ConfigMap{}

	err := kclient.Get(ctx, client.ObjectKey{
		Name:      HobbyfarmInstanceIdName,
		Namespace: namespace.ResolveNamespace(),
	}, cfg)
	if err != nil {
		return "", err
	}

	if val, ok := cfg.Data[InstanceIdKey]; !ok {
		return "", fmt.Errorf("key instance-id not found in instance configmap")
	} else {
		return val, nil
	}
}

// createInstanceId creates an instance id (uuid) and stores it in a configmap)
func createInstanceId(ctx context.Context, kclient client.Client) (string, error) {
	id := uuid.NewUUID()

	cm := &v1.ConfigMap{
		ObjectMeta: v12.ObjectMeta{
			Name:      HobbyfarmInstanceIdName,
			Namespace: namespace.ResolveNamespace(),
		},
		Data: map[string]string{
			InstanceIdKey: string(id),
		},
	}

	err := kclient.Create(ctx, cm)
	if err != nil {
		return "", err
	}

	return string(id), nil
}
