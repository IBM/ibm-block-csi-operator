package ibmblockcsi

import (
	"github.com/IBM/ibm-block-csi-driver-operator/pkg/config"
)

// TODO: improve this function
// SetDefaults set defaults if omitted in spec, returns true means CR should be updated on cluster.
// Replace it with kubernetes native default setter when it is available.
// https://kubernetes.io/docs/tasks/access-kubernetes-api/custom-resources/custom-resource-definitions/#defaulting
func (c *IBMBlockCSI) SetDefaults() bool {
	changed := false

	if c.Spec == nil {
		c.Spec.Controller.Repository = config.ControllerRepository
		c.Spec.Controller.Tag = config.ControllerTag
		c.Spec.Node.Repository = config.NodeRepository
		c.Spec.Node.Tag = config.NodeTag

		changed = true
	}

	return changed
}
