/**
 * Copyright 2019 IBM Corp.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package crutils

import (
	"github.com/IBM/ibm-block-csi-operator/pkg/config"
	"github.com/IBM/ibm-block-csi-operator/pkg/util/boolptr"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	storagev1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	snapshotStorageApiGroup                  string = "snapshot.storage.k8s.io"
	securityOpenshiftApiGroup                string = "security.openshift.io"
	storageApiGroup                          string = "storage.k8s.io"
	rbacAuthorizationApiGroup                string = "rbac.authorization.k8s.io"
	replicationStorageOpenshiftApiGroup      string = "replication.storage.openshift.io"
	storageClassesResource                   string = "storageclasses"
	persistentVolumesResource                string = "persistentvolumes"
	persistentVolumeClaimsResource           string = "persistentvolumeclaims"
	persistentVolumeClaimsStatusResource     string = "persistentvolumeclaims/status"
	persistentVolumeClaimsFinalizersResource string = "persistentvolumeclaims/finalizers"
	volumeGroupClassesResource               string = "volumegroupclasses"
	volumeCroupContentsResource              string = "volumegroupcontents"
	volumeGroupsResources                    string = "volumegroups"
	volumeGroupsStatusResource               string = "volumegroups/status"
	volumeGroupsFinalizersResource           string = "volumegroups/finalizers"
	podsResource                             string = "pods"
	volumeAttachmentsResource                string = "volumeattachments"
	volumeAttachmentsStatusResource          string = "volumeattachments/status"
	volumeSnapshotClassesResource            string = "volumesnapshotclasses"
	volumeSnapshotsResource                  string = "volumesnapshots"
	volumeSnapshotContentsResource           string = "volumesnapshotcontents"
	volumeSnapshotContentsStatusResource     string = "volumesnapshotcontents/status"
	volumeReplicationClassesResource         string = "volumereplicationclasses"
	volumeReplicationsResource               string = "volumereplications"
	volumeReplicationsFinalizersResource     string = "volumereplications/finalizers"
	volumeReplicationsStatusResource         string = "volumereplications/status"
	eventsResource                           string = "events"
	nodesResource                            string = "nodes"
	csiNodesResource                         string = "csinodes"
	secretsResource                          string = "secrets"
	securityContextConstraintsResource       string = "securitycontextconstraints"
	verbGet                                  string = "get"
	verbList                                 string = "list"
	verbWatch                                string = "watch"
	verbCreate                               string = "create"
	verbUpdate                               string = "update"
	verbPatch                                string = "patch"
	verbDelete                               string = "delete"
)

func (c *IBMBlockCSI) GenerateCSIDriver() *storagev1.CSIDriver {
	return &storagev1.CSIDriver{
		ObjectMeta: metav1.ObjectMeta{
			Name: config.DriverName,
		},
		Spec: storagev1.CSIDriverSpec{
			AttachRequired: boolptr.True(),
			PodInfoOnMount: boolptr.False(),
		},
	}
}

func (c *IBMBlockCSI) GenerateControllerServiceAccount() *corev1.ServiceAccount {
	return getServiceAccount(c, config.CSIControllerServiceAccount)
}

func (c *IBMBlockCSI) GenerateNodeServiceAccount() *corev1.ServiceAccount {
	return getServiceAccount(c, config.CSINodeServiceAccount)
}

func getServiceAccount(c *IBMBlockCSI, serviceAccountResourceName config.ResourceName) *corev1.ServiceAccount {
	secrets := getImagePullSecrets(c.Spec.ImagePullSecrets)
	return &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      config.GetNameForResource(serviceAccountResourceName, c.Name),
			Namespace: c.Namespace,
			Labels:    c.GetLabels(),
		},
		ImagePullSecrets: secrets,
	}
}

func (c *IBMBlockCSI) GenerateExternalProvisionerClusterRole() *rbacv1.ClusterRole {
	return &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: config.GetNameForResource(config.ExternalProvisionerClusterRole, c.Name),
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{""},
				Resources: []string{secretsResource},
				Verbs:     []string{verbGet, verbList},
			},
			{
				APIGroups: []string{""},
				Resources: []string{persistentVolumesResource},
				Verbs:     []string{verbGet, verbList, verbWatch, verbCreate, verbDelete},
			},
			{
				APIGroups: []string{""},
				Resources: []string{persistentVolumeClaimsResource},
				Verbs:     []string{verbGet, verbList, verbWatch, verbUpdate},
			},
			{
				APIGroups: []string{storageApiGroup},
				Resources: []string{storageClassesResource},
				Verbs:     []string{verbGet, verbList, verbWatch},
			},
			{
				APIGroups: []string{""},
				Resources: []string{eventsResource},
				Verbs:     []string{verbList, verbWatch, verbCreate, verbUpdate, verbPatch},
			},
			{
				APIGroups: []string{snapshotStorageApiGroup},
				Resources: []string{volumeSnapshotsResource},
				Verbs:     []string{verbGet, verbList},
			},
			{
				APIGroups: []string{snapshotStorageApiGroup},
				Resources: []string{volumeSnapshotContentsResource},
				Verbs:     []string{verbGet, verbList},
			},
			{
				APIGroups: []string{storageApiGroup},
				Resources: []string{csiNodesResource},
				Verbs:     []string{verbGet, verbList, verbWatch},
			},
			{
				APIGroups: []string{""},
				Resources: []string{nodesResource},
				Verbs:     []string{verbGet, verbList, verbWatch},
			},
			{
				APIGroups: []string{storageApiGroup},
				Resources: []string{volumeAttachmentsResource},
				Verbs:     []string{verbGet, verbList, verbWatch},
			},
		},
	}
}

func (c *IBMBlockCSI) GenerateExternalProvisionerClusterRoleBinding() *rbacv1.ClusterRoleBinding {
	return &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: config.GetNameForResource(config.ExternalProvisionerClusterRoleBinding, c.Name),
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      config.GetNameForResource(config.CSIControllerServiceAccount, c.Name),
				Namespace: c.Namespace,
			},
		},
		RoleRef: rbacv1.RoleRef{
			Kind:     "ClusterRole",
			Name:     config.GetNameForResource(config.ExternalProvisionerClusterRole, c.Name),
			APIGroup: rbacAuthorizationApiGroup,
		},
	}
}

func (c *IBMBlockCSI) GenerateExternalAttacherClusterRole() *rbacv1.ClusterRole {
	return &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: config.GetNameForResource(config.ExternalAttacherClusterRole, c.Name),
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{""},
				Resources: []string{persistentVolumesResource},
				Verbs:     []string{verbGet, verbList, verbWatch, verbPatch},
			},
			{
				APIGroups: []string{storageApiGroup},
				Resources: []string{csiNodesResource},
				Verbs:     []string{verbGet, verbList, verbWatch},
			},
			{
				APIGroups: []string{storageApiGroup},
				Resources: []string{volumeAttachmentsResource},
				Verbs:     []string{verbGet, verbList, verbWatch, verbPatch},
			},
			{
				APIGroups: []string{storageApiGroup},
				Resources: []string{volumeAttachmentsStatusResource},
				Verbs:     []string{verbPatch},
			},
			{
				APIGroups: []string{""},
				Resources: []string{secretsResource},
				Verbs:     []string{verbGet, verbList},
			},
		},
	}
}

func (c *IBMBlockCSI) GenerateExternalAttacherClusterRoleBinding() *rbacv1.ClusterRoleBinding {
	return &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: config.GetNameForResource(config.ExternalAttacherClusterRoleBinding, c.Name),
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      config.GetNameForResource(config.CSIControllerServiceAccount, c.Name),
				Namespace: c.Namespace,
			},
		},
		RoleRef: rbacv1.RoleRef{
			Kind:     "ClusterRole",
			Name:     config.GetNameForResource(config.ExternalAttacherClusterRole, c.Name),
			APIGroup: rbacAuthorizationApiGroup,
		},
	}
}

func (c *IBMBlockCSI) GenerateExternalSnapshotterClusterRole() *rbacv1.ClusterRole {
	return &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: config.GetNameForResource(config.ExternalSnapshotterClusterRole, c.Name),
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{""},
				Resources: []string{"events"},
				Verbs:     []string{verbList, verbWatch, verbCreate, verbUpdate, verbPatch},
			},
			{
				APIGroups: []string{""},
				Resources: []string{secretsResource},
				Verbs:     []string{verbGet, verbList},
			},
			{
				APIGroups: []string{snapshotStorageApiGroup},
				Resources: []string{volumeSnapshotClassesResource},
				Verbs:     []string{verbGet, verbList, verbWatch},
			},
			{
				APIGroups: []string{snapshotStorageApiGroup},
				Resources: []string{volumeSnapshotContentsResource},
				Verbs:     []string{verbCreate, verbGet, verbList, verbWatch, verbUpdate, verbDelete, verbPatch},
			},
			{
				APIGroups: []string{snapshotStorageApiGroup},
				Resources: []string{volumeSnapshotContentsStatusResource},
				Verbs:     []string{verbUpdate, verbPatch},
			},
		},
	}
}

func (c *IBMBlockCSI) GenerateExternalSnapshotterClusterRoleBinding() *rbacv1.ClusterRoleBinding {
	return &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: config.GetNameForResource(config.ExternalSnapshotterClusterRoleBinding, c.Name),
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      config.GetNameForResource(config.CSIControllerServiceAccount, c.Name),
				Namespace: c.Namespace,
			},
		},
		RoleRef: rbacv1.RoleRef{
			Kind:     "ClusterRole",
			Name:     config.GetNameForResource(config.ExternalSnapshotterClusterRole, c.Name),
			APIGroup: rbacAuthorizationApiGroup,
		},
	}
}

func (c *IBMBlockCSI) GenerateExternalResizerClusterRole() *rbacv1.ClusterRole {
	return &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: config.GetNameForResource(config.ExternalResizerClusterRole, c.Name),
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{""},
				Resources: []string{persistentVolumesResource},
				Verbs:     []string{verbGet, verbList, verbWatch, verbPatch},
			},
			{
				APIGroups: []string{""},
				Resources: []string{persistentVolumeClaimsResource},
				Verbs:     []string{verbGet, verbList, verbWatch},
			},
			{
				APIGroups: []string{""},
				Resources: []string{podsResource},
				Verbs:     []string{verbGet, verbList, verbWatch},
			},
			{
				APIGroups: []string{""},
				Resources: []string{persistentVolumeClaimsStatusResource},
				Verbs:     []string{verbPatch},
			},
			{
				APIGroups: []string{""},
				Resources: []string{eventsResource},
				Verbs:     []string{verbList, verbWatch, verbCreate, verbUpdate, verbPatch},
			},
			{
				APIGroups: []string{""},
				Resources: []string{secretsResource},
				Verbs:     []string{verbGet, verbList, verbWatch},
			},
		},
	}
}

func (c *IBMBlockCSI) GenerateExternalResizerClusterRoleBinding() *rbacv1.ClusterRoleBinding {
	return &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: config.GetNameForResource(config.ExternalResizerClusterRoleBinding, c.Name),
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      config.GetNameForResource(config.CSIControllerServiceAccount, c.Name),
				Namespace: c.Namespace,
			},
		},
		RoleRef: rbacv1.RoleRef{
			Kind:     "ClusterRole",
			Name:     config.GetNameForResource(config.ExternalResizerClusterRole, c.Name),
			APIGroup: rbacAuthorizationApiGroup,
		},
	}
}

func (c *IBMBlockCSI) GenerateCSIAddonsReplicatorClusterRole() *rbacv1.ClusterRole {
	return &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: config.GetNameForResource(config.CSIAddonsReplicatorClusterRole, c.Name),
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{replicationStorageOpenshiftApiGroup},
				Resources: []string{volumeReplicationClassesResource},
				Verbs:     []string{verbGet, verbList, verbWatch},
			},
			{
				APIGroups: []string{replicationStorageOpenshiftApiGroup},
				Resources: []string{volumeReplicationsResource},
				Verbs:     []string{verbCreate, verbDelete, verbGet, verbList, verbPatch, verbUpdate, verbWatch},
			},
			{
				APIGroups: []string{replicationStorageOpenshiftApiGroup},
				Resources: []string{volumeReplicationsFinalizersResource},
				Verbs:     []string{verbUpdate},
			},
			{
				APIGroups: []string{replicationStorageOpenshiftApiGroup},
				Resources: []string{volumeReplicationsStatusResource},
				Verbs:     []string{verbGet, verbPatch, verbUpdate},
			},
			{
				APIGroups: []string{""},
				Resources: []string{secretsResource},
				Verbs:     []string{verbGet},
			},
		},
	}
}

func (c *IBMBlockCSI) GenerateCSIAddonsReplicatorClusterRoleBinding() *rbacv1.ClusterRoleBinding {
	return &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: config.GetNameForResource(config.CSIAddonsReplicatorClusterRoleBinding, c.Name),
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      config.GetNameForResource(config.CSIControllerServiceAccount, c.Name),
				Namespace: c.Namespace,
			},
		},
		RoleRef: rbacv1.RoleRef{
			Kind:     "ClusterRole",
			Name:     config.GetNameForResource(config.CSIAddonsReplicatorClusterRole, c.Name),
			APIGroup: rbacAuthorizationApiGroup,
		},
	}
}

func (c *IBMBlockCSI) GenerateVolumeGroupClusterRole() *rbacv1.ClusterRole {
	return &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: config.GetNameForResource(config.CSIVolumeGroupClusterRole, c.Name),
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{""},
				Resources: []string{volumeGroupsResources},
				Verbs:     []string{verbGet, verbList, verbWatch, verbCreate, verbUpdate, verbPatch, verbDelete},
			},
			{
				APIGroups: []string{""},
				Resources: []string{volumeGroupsStatusResource},
				Verbs:     []string{verbGet, verbUpdate, verbPatch},
			},
			{
				APIGroups: []string{""},
				Resources: []string{volumeGroupsFinalizersResource},
				Verbs:     []string{verbUpdate},
			},
			{
				APIGroups: []string{""},
				Resources: []string{volumeGroupsResources},
				Verbs:     []string{verbGet, verbList, verbWatch, verbCreate, verbUpdate, verbPatch, verbDelete},
			},
			{
				APIGroups: []string{""},
				Resources: []string{volumeGroupClassesResource},
				Verbs:     []string{verbGet, verbList, verbWatch},
			},
			{
				APIGroups: []string{""},
				Resources: []string{volumeCroupContentsResource},
				Verbs:     []string{verbGet, verbList, verbWatch},
			},
			{
				APIGroups: []string{""},
				Resources: []string{persistentVolumeClaimsStatusResource},
				Verbs:     []string{verbGet, verbUpdate, verbPatch},
			},
			{
				APIGroups: []string{""},
				Resources: []string{persistentVolumeClaimsFinalizersResource},
				Verbs:     []string{verbUpdate},
			},
			{
				APIGroups: []string{replicationStorageOpenshiftApiGroup},
				Resources: []string{volumeReplicationsFinalizersResource},
				Verbs:     []string{verbUpdate},
			},
			{
				APIGroups: []string{replicationStorageOpenshiftApiGroup},
				Resources: []string{volumeReplicationsStatusResource},
				Verbs:     []string{verbGet, verbPatch, verbUpdate},
			},
			{
				APIGroups: []string{""},
				Resources: []string{secretsResource},
				Verbs:     []string{verbGet},
			},
		},
	}
}

func (c *IBMBlockCSI) GenerateVolumeGroupClusterRoleBinding() *rbacv1.ClusterRoleBinding {
	return &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: config.GetNameForResource(config.CSIVolumeGroupClusterRoleBinding, c.Name),
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      config.GetNameForResource(config.CSIControllerServiceAccount, c.Name),
				Namespace: c.Namespace,
			},
		},
		RoleRef: rbacv1.RoleRef{
			Kind:     "ClusterRole",
			Name:     config.GetNameForResource(config.CSIVolumeGroupClusterRole, c.Name),
			APIGroup: rbacAuthorizationApiGroup,
		},
	}
}

func (c *IBMBlockCSI) GenerateSCCForControllerClusterRole() *rbacv1.ClusterRole {
	return &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: config.GetNameForResource(config.CSIControllerSCCClusterRole, c.Name),
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups:     []string{securityOpenshiftApiGroup},
				Resources:     []string{securityContextConstraintsResource},
				ResourceNames: []string{"anyuid"},
				Verbs:         []string{"use"},
			},
		},
	}
}

func (c *IBMBlockCSI) GenerateSCCForControllerClusterRoleBinding() *rbacv1.ClusterRoleBinding {
	return &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: config.GetNameForResource(config.CSIControllerSCCClusterRoleBinding, c.Name),
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      config.GetNameForResource(config.CSIControllerServiceAccount, c.Name),
				Namespace: c.Namespace,
			},
		},
		RoleRef: rbacv1.RoleRef{
			Kind:     "ClusterRole",
			Name:     config.GetNameForResource(config.CSIControllerSCCClusterRole, c.Name),
			APIGroup: rbacAuthorizationApiGroup,
		},
	}
}

func (c *IBMBlockCSI) GenerateSCCForNodeClusterRole() *rbacv1.ClusterRole {
	return &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: config.GetNameForResource(config.CSINodeSCCClusterRole, c.Name),
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups:     []string{securityOpenshiftApiGroup},
				Resources:     []string{securityContextConstraintsResource},
				ResourceNames: []string{"privileged"},
				Verbs:         []string{"use"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{nodesResource},
				Verbs:     []string{verbGet},
			},
		},
	}
}

func (c *IBMBlockCSI) GenerateSCCForNodeClusterRoleBinding() *rbacv1.ClusterRoleBinding {
	return &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: config.GetNameForResource(config.CSINodeSCCClusterRoleBinding, c.Name),
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      config.GetNameForResource(config.CSINodeServiceAccount, c.Name),
				Namespace: c.Namespace,
			},
		},
		RoleRef: rbacv1.RoleRef{
			Kind:     "ClusterRole",
			Name:     config.GetNameForResource(config.CSINodeSCCClusterRole, c.Name),
			APIGroup: rbacAuthorizationApiGroup,
		},
	}
}
