package ibmblockcsi

import (
	csiv1 "github.com/IBM/ibm-block-csi-driver-operator/pkg/apis/csi/v1"
	"github.com/IBM/ibm-block-csi-driver-operator/pkg/config"
	csiversion "github.com/IBM/ibm-block-csi-driver-operator/version"
	"k8s.io/apimachinery/pkg/labels"
)

// IBMBlockCSI is the wrapper for csiv1.IBMBlockCSI type
type IBMBlockCSI struct {
	*csiv1.IBMBlockCSI
	ServerVersion string
}

// New returns a wrapper for csiv1.IBMBlockCSI
func New(c *csiv1.IBMBlockCSI, serverVersion string) *IBMBlockCSI {
	return &IBMBlockCSI{
		IBMBlockCSI:   c,
		ServerVersion: serverVersion,
	}
}

// Unwrap returns the csiv1.IBMBlockCSI object
func (c *IBMBlockCSI) Unwrap() *csiv1.IBMBlockCSI {
	return c.IBMBlockCSI
}

// GetAnnotations returns all the annotations to be set on all resources
func (c *IBMBlockCSI) GetAnnotations() labels.Set {
	labels := labels.Set{
		"app.kubernetes.io/name":       config.ProductName,
		"app.kubernetes.io/instance":   c.Name,
		"app.kubernetes.io/version":    csiversion.Version,
		"app.kubernetes.io/managed-by": config.Name,
	}

	if c.Annotations != nil {
		for k, v := range c.Annotations {
			labels[k] = v
		}
	}

	return labels
}

func (c *IBMBlockCSI) GetComponentAnnotations(component string) labels.Set {
	return labels.Set{
		"app.kubernetes.io/component": component,
	}
}

func (c *IBMBlockCSI) GetCSIControllerComponentAnnotations() labels.Set {
	return c.GetComponentAnnotations(config.CSIController.String())
}

func (c *IBMBlockCSI) GetCSINodeComponentAnnotations() labels.Set {
	return c.GetComponentAnnotations(config.CSINode.String())
}

func (c *IBMBlockCSI) GetCSIControllerAnnotations() labels.Set {
	labels := c.GetLabels()
	for k, v := range c.GetCSIControllerComponentAnnotations() {
		labels[k] = v
	}
	return labels
}

func (c *IBMBlockCSI) GetCSINodeAnnotations() labels.Set {
	labels := c.GetLabels()
	for k, v := range c.GetCSINodeComponentAnnotations() {
		labels[k] = v
	}
	return labels
}

func (c *IBMBlockCSI) GetCSIControllerImage() string {
	if c.Spec.Controller.Tag == "" {
		return c.Spec.Controller.Repository
	}
	return c.Spec.Controller.Repository + ":" + c.Spec.Controller.Tag
}

func (c *IBMBlockCSI) GetCSINodeImage() string {
	if c.Spec.Node.Tag == "" {
		return c.Spec.Node.Repository
	}
	return c.Spec.Node.Repository + ":" + c.Spec.Node.Tag
}
