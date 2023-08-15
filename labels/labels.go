package labels

import "k8s.io/apimachinery/pkg/labels"

const ProvisionerLabel = "hobbyfarm.io/provisioner"
const VirtualMachineLabel = "provisioning.hobbyfarm.io/virtual-machine"

func VMLabelSelector(vmName string) labels.Selector {
	return labels.SelectorFromSet(map[string]string{
		VirtualMachineLabel: vmName,
	})
}
