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

package ibmblockcsi

import (
	"github.com/IBM/ibm-block-csi-operator/pkg/config"
	"github.com/IBM/ibm-block-csi-operator/pkg/util/boolptr"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	storagev1beta1 "k8s.io/api/storage/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	snapshotStorageApiGroup              string = "snapshot.storage.k8s.io"
	securityOpenshiftApiGroup            string = "security.openshift.io"
	apiExtensionsApiGroup                string = "apiextensions.k8s.io"
	storageApiGroup                      string = "storage.k8s.io"
	rbacAuthorizationApiGroup            string = "rbac.authorization.k8s.io"
	storageClassesResource               string = "storageclasses"
	persistentVolumesResource            string = "persistentvolumes"
	persistentVolumeClaimsResource       string = "persistentvolumeclaims"
	volumeAttachmentsResource            string = "volumeattachments"
	volumeSnapshotClassesResource        string = "volumesnapshotclasses"
	volumeSnapshotsResource              string = "volumesnapshots"
	volumeSnapshotsStatusResource        string = "volumesnapshots/status"
	volumeSnapshotContentsResource       string = "volumesnapshotcontents"
	volumeSnapshotContentsStatusResource string = "volumesnapshotcontents/status"
	eventsResource                       string = "events"
	nodesResource                        string = "nodes"
	csiNodesResource                     string = "csinodes"
	secretsResource                      string = "secrets"
	securityContextConstraintsResource   string = "securitycontextconstraints"
	customResourceDefinitionsResource    string = "customresourcedefinitions"
)

func (c *IBMBlockCSI) GenerateCSIDriver() *storagev1beta1.CSIDriver {
	return &storagev1beta1.CSIDriver{
		ObjectMeta: metav1.ObjectMeta{
			Name: config.DriverName,
		},
		Spec: storagev1beta1.CSIDriverSpec{
			AttachRequired: boolptr.True(),
			PodInfoOnMount: boolptr.False(),
		},
	}
}

func (c *IBMBlockCSI) GenerateControllerServiceAccount() *corev1.ServiceAccount {
	secrets := []corev1.LocalObjectReference{}
	if len(c.Spec.ImagePullSecrets) > 0 {
		for _, s := range c.Spec.ImagePullSecrets {
			secrets = append(secrets, corev1.LocalObjectReference{Name: s})
		}
	}

	return &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      config.GetNameForResource(config.CSIControllerServiceAccount, c.Name),
			Namespace: c.Namespace,
			Labels:    c.GetLabels(),
		},
		ImagePullSecrets: secrets,
	}
}

func (c *IBMBlockCSI) GenerateNodeServiceAccount() *corev1.ServiceAccount {
	secrets := []corev1.LocalObjectReference{}
	if len(c.Spec.ImagePullSecrets) > 0 {
		for _, s := range c.Spec.ImagePullSecrets {
			secrets = append(secrets, corev1.LocalObjectReference{Name: s})
		}
	}

	return &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      config.GetNameForResource(config.CSINodeServiceAccount, c.Name),
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
				Verbs:     []string{"get", "list"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{persistentVolumesResource},
				Verbs:     []string{"get", "list", "watch", "create", "delete"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{persistentVolumeClaimsResource},
				Verbs:     []string{"get", "list", "watch", "update"},
			},
			{
				APIGroups: []string{storageApiGroup},
				Resources: []string{storageClassesResource},
				Verbs:     []string{"get", "list", "watch"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{eventsResource},
				Verbs:     []string{"list", "watch", "create", "update", "patch"},
			},
			{
				APIGroups: []string{storageApiGroup},
				Resources: []string{csiNodesResource},
				Verbs:     []string{"get", "list", "watch"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{nodesResource},
				Verbs:     []string{"get", "list", "watch"},
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
				Resources: []string{"persistentvolumes"},
				Verbs:     []string{"get", "list", "watch", "update", "patch"},
			},
			{
				APIGroups: []string{storageApiGroup},
				Resources: []string{csiNodesResource},
				Verbs:     []string{"get", "list", "watch"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{nodesResource},
				Verbs:     []string{"get", "list", "watch"},
			},
			{
				APIGroups: []string{storageApiGroup},
				Resources: []string{volumeAttachmentsResource},
				Verbs:     []string{"get", "list", "watch", "update", "patch"},
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
				Resources: []string{persistentVolumesResource},
				Verbs:     []string{"get", "list", "watch"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{persistentVolumeClaimsResource},
				Verbs:     []string{"get", "list", "watch"},
			},
			{
				APIGroups: []string{storageApiGroup},
				Resources: []string{storageClassesResource},
				Verbs:     []string{"get", "list", "watch"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"events"},
				Verbs:     []string{"list", "watch", "create", "update", "patch"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{secretsResource},
				Verbs:     []string{"get", "list"},
			},
			{
				APIGroups: []string{snapshotStorageApiGroup},
				Resources: []string{volumeSnapshotClassesResource},
				Verbs:     []string{"get", "list", "watch"},
			},
			{
				APIGroups: []string{snapshotStorageApiGroup},
				Resources: []string{volumeSnapshotsResource},
				Verbs:     []string{"get", "list", "watch", "update"},
			},
			{
				APIGroups: []string{snapshotStorageApiGroup},
				Resources: []string{volumeSnapshotsStatusResource},
				Verbs:     []string{"update"},
			},
			{
				APIGroups: []string{snapshotStorageApiGroup},
				Resources: []string{volumeSnapshotContentsResource},
				Verbs:     []string{"create", "get", "list", "watch", "update", "delete"},
			},
			{
				APIGroups: []string{snapshotStorageApiGroup},
				Resources: []string{volumeSnapshotContentsStatusResource},
				Verbs:     []string{"get", "update"},
			},
			{
				APIGroups: []string{apiExtensionsApiGroup},
				Resources: []string{customResourceDefinitionsResource},
				Verbs:     []string{"create", "list", "watch", "delete"},
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
