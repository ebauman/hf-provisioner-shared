package instanceid

import (
	"context"
	v1 "k8s.io/api/core/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"strings"
	"testing"
)

func createClientWithConfigMap(cmName string, idKey string, idVal string) client.Client {
	return fake.NewFakeClient(&v1.ConfigMap{
		ObjectMeta: v12.ObjectMeta{
			Name:      cmName,
			Namespace: "default",
		},
		Data: map[string]string{
			idKey: idVal,
		},
	})
}

func Test_GetInstanceId_NonExistent(t *testing.T) {
	kclient := fake.NewFakeClient()

	_, err := GetInstanceId(context.Background(), kclient)

	if err == nil {
		t.Error("should have received error")
	}
}

func Test_GetInstanceId_Exists(t *testing.T) {
	var uuid = "839d1199-a61c-4dd6-bb88-bf120acb6041"
	kclient := createClientWithConfigMap(HobbyfarmInstanceIdName, InstanceIdKey, uuid)

	id, err := GetInstanceId(context.Background(), kclient)
	if err != nil {
		t.Error("should not receive error")
	}

	if id != uuid {
		t.Error("invalid id")
	}
}

func Test_GetInstanceId_WrongConfigmapName(t *testing.T) {
	var uuid = "839d1199-a61c-4dd6-bb88-bf120acb6041"
	kclient := createClientWithConfigMap("blahblah", InstanceIdKey, uuid)

	id, err := GetInstanceId(context.Background(), kclient)
	if err == nil {
		t.Error("should receive error")
	}

	if id != "" {
		t.Error("id should not be filled")
	}
}

func Test_GetInstanceId_WrongKeyName(t *testing.T) {
	var uuid = "839d1199-a61c-4dd6-bb88-bf120acb6041"
	kclient := createClientWithConfigMap(HobbyfarmInstanceIdName, "id", uuid)

	id, err := GetInstanceId(context.Background(), kclient)
	if err == nil {
		t.Error("should receive error")
	}

	if !strings.Contains(err.Error(), "key instance-id not found") {
		t.Error("wrong error received")
	}

	if id != "" {
		t.Error("id should not be filled")
	}
}

func Test_GetCreateInstanceId_InstanceIdExists(t *testing.T) {
	var uuid = "839d1199-a61c-4dd6-bb88-bf120acb6041"
	kclient := createClientWithConfigMap(HobbyfarmInstanceIdName, InstanceIdKey, uuid)

	id, err := GetOrCreateInstanceId(context.Background(), kclient)
	if err != nil {
		t.Error("should not receive error")
	}

	if id != uuid {
		t.Error("wrong id received")
	}
}

func Test_GetCreateInstanceId_NonPreExisting(t *testing.T) {
	kclient := fake.NewFakeClient()

	id, err := GetOrCreateInstanceId(context.Background(), kclient)
	if err != nil {
		t.Errorf("received error, should not have: %s", err.Error())
	}

	if id == "" {
		t.Error("should not receive empty id")
	}
}

func Test_GetCreateInstanceId_WrongExisting(t *testing.T) {
	var uuid = "839d1199-a61c-4dd6-bb88-bf120acb6041"
	kclient := createClientWithConfigMap(HobbyfarmInstanceIdName, "blah", uuid)

	id, err := GetOrCreateInstanceId(context.Background(), kclient)
	if err == nil {
		t.Errorf("should receive error, didn't")
	}

	if id != "" {
		t.Error("id should not be filled")
	}
}
