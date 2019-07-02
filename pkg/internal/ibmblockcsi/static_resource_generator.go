package ibmblockcsi

import (
	"github.com/IBM/ibm-block-csi-driver-operator/pkg/config"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *IBMBlockCSI) GenerateControllerServiceAccount() *corev1.ServiceAccount {
	return &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      config.GetNameForResource(config.CSIControllerServiceAccount, c.Name),
			Namespace: c.Namespace,
		},
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
				Resources: []string{"secrets"},
				Verbs:     []string{"get", "list"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"persistentvolumes"},
				Verbs:     []string{"get", "list", "watch", "create", "delete"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"persistentvolumeclaims"},
				Verbs:     []string{"get", "list", "watch", "update"},
			},
			{
				APIGroups: []string{"storage.k8s.io"},
				Resources: []string{"storageclasses"},
				Verbs:     []string{"get", "list", "watch"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"events"},
				Verbs:     []string{"list", "watch", "create", "update", "patch"},
			},
			{
				APIGroups: []string{"snapshot.storage.k8s.io"},
				Resources: []string{"volumesnapshots"},
				Verbs:     []string{"get", "list"},
			},
			{
				APIGroups: []string{"snapshot.storage.k8s.io"},
				Resources: []string{"volumesnapshotcontents"},
				Verbs:     []string{"get", "list"},
			},
			{
				APIGroups: []string{c.GetCSIAPIGroup()},
				Resources: []string{"csinodes"},
				Verbs:     []string{"get", "list", "watch"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"nodes"},
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
			APIGroup: "rbac.authorization.k8s.io",
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
				Verbs:     []string{"get", "list", "watch", "update"},
			},
			{
				APIGroups: []string{c.GetCSIAPIGroup()},
				Resources: []string{"csinodes"},
				Verbs:     []string{"get", "list", "watch"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"nodes"},
				Verbs:     []string{"get", "list", "watch"},
			},
			{
				APIGroups: []string{"storage.k8s.io"},
				Resources: []string{"volumeattachments"},
				Verbs:     []string{"get", "list", "watch", "update"},
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
			APIGroup: "rbac.authorization.k8s.io",
		},
	}
}

func (c *IBMBlockCSI) GenerateClusterDriverRegistrarClusterRole() *rbacv1.ClusterRole {
	rules := []rbacv1.PolicyRule{
		{
			APIGroups: []string{c.GetCSIAPIGroup()},
			Resources: []string{"csidrivers"},
			Verbs:     []string{"create", "delete"},
		},
	}
	if c.ServerVersion == "1.13" {
		rules = append(rules, rbacv1.PolicyRule{
			APIGroups: []string{"apiextensions.k8s.io"},
			Resources: []string{"customresourcedefinitions"},
			Verbs:     []string{"create", "list", "watch", "delete"},
		})
	}

	return &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: config.GetNameForResource(config.ClusterDriverRegistrarClusterRole, c.Name),
		},
		Rules: rules,
	}
}

func (c *IBMBlockCSI) GenerateClusterDriverRegistrarClusterRoleBinding() *rbacv1.ClusterRoleBinding {
	return &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: config.GetNameForResource(config.ClusterDriverRegistrarClusterRoleBinding, c.Name),
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
			Name:     config.GetNameForResource(config.ClusterDriverRegistrarClusterRole, c.Name),
			APIGroup: "rbac.authorization.k8s.io",
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
				Resources: []string{"persistentvolumes"},
				Verbs:     []string{"get", "list", "watch"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"persistentvolumeclaims"},
				Verbs:     []string{"get", "list", "watch"},
			},
			{
				APIGroups: []string{"storage.k8s.io"},
				Resources: []string{"storageclasses"},
				Verbs:     []string{"get", "list", "watch"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"events"},
				Verbs:     []string{"list", "watch", "create", "update", "patch"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"secrets"},
				Verbs:     []string{"get", "list"},
			},
			{
				APIGroups: []string{"snapshot.storage.k8s.io"},
				Resources: []string{"volumesnapshotclasses"},
				Verbs:     []string{"get", "list", "watch"},
			},
			{
				APIGroups: []string{"snapshot.storage.k8s.io"},
				Resources: []string{"volumesnapshots"},
				Verbs:     []string{"get", "list", "watch", "update"},
			},
			{
				APIGroups: []string{"snapshot.storage.k8s.io"},
				Resources: []string{"volumesnapshotcontents"},
				Verbs:     []string{"create", "get", "list", "watch", "update", "delete"},
			},
			{
				APIGroups: []string{"apiextensions.k8s.io"},
				Resources: []string{"customresourcedefinitions"},
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
			APIGroup: "rbac.authorization.k8s.io",
		},
	}
}

func (c *IBMBlockCSI) GetCSIAPIGroup() string {
	if c.ServerVersion == "1.13" {
		return "csi.storage.k8s.io"
	}
	return "storage.k8s.io"
}
