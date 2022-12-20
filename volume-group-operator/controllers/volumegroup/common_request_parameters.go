package volumegroup

import "github.com/IBM/volume-group-operator/pkg/client"

type CommonRequestParameters struct {
	Name          string
	VolumeGroupID string
	VolumeIds     []string
	Parameters    map[string]string
	Secrets       map[string]string
	VolumeGroup   client.VolumeGroup
}
