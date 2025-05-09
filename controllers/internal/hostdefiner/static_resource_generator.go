/**
 * Copyright 2025 IBM Corp.
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

package hostdefiner

import (
	"github.com/IBM/ibm-block-csi-operator/pkg/config"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *HostDefiner) GenerateHostDefinerClusterRoleBinding() *rbacv1.ClusterRoleBinding {
	return &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: config.GetNameForResource(config.HostDefinerClusterRoleBinding, c.Name),
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      config.GetNameForResource(config.HostDefinerServiceAccount, c.Name),
				Namespace: c.Namespace,
			},
		},
		RoleRef: rbacv1.RoleRef{
			Kind:     "ClusterRole",
			Name:     config.GetNameForResource(config.HostDefinerClusterRole, c.Name),
			APIGroup: config.RbacAuthorizationApiGroup,
		},
	}
}

func (c *HostDefiner) GenerateHostDefinerClusterRole() *rbacv1.ClusterRole {
	return &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: config.GetNameForResource(config.HostDefinerClusterRole, c.Name),
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{""},
				Resources: []string{config.SecretsResource},
				Verbs:     []string{config.VerbGet, config.VerbList, config.VerbWatch},
			},
			{
				APIGroups: []string{""},
				Resources: []string{config.PodsResource},
				Verbs:     []string{config.VerbList},
			},
			{
				APIGroups: []string{""},
				Resources: []string{config.EventsResource},
				Verbs:     []string{config.VerbCreate},
			},
			{
				APIGroups: []string{config.StorageApiGroup},
				Resources: []string{config.StorageClassesResource},
				Verbs:     []string{config.VerbGet, config.VerbList, config.VerbWatch},
			},
			{
				APIGroups: []string{config.StorageApiGroup},
				Resources: []string{config.CsiNodesResource},
				Verbs:     []string{config.VerbGet, config.VerbList, config.VerbWatch},
			},
			{
				APIGroups: []string{""},
				Resources: []string{config.NodesResource},
				Verbs:     []string{config.VerbGet, config.VerbPatch, config.VerbWatch, config.VerbList},
			},
			{
				APIGroups: []string{config.APIGroup},
				Resources: []string{config.HostDefinerResource},
				Verbs:     []string{config.VerbGet, config.VerbList, config.VerbWatch},
			},
			{
				APIGroups: []string{config.APIGroup},
				Resources: []string{config.HostDefinitionResource},
				Verbs:     []string{"*"},
			},
			{
				APIGroups: []string{config.APIGroup},
				Resources: []string{config.HostDefinitionStatusResource},
				Verbs:     []string{"*"},
			},
			{
				APIGroups: []string{config.AppsApiGroup},
				Resources: []string{config.DaemonSetResource},
				Verbs:     []string{config.VerbList},
			},
		},
	}
}

func (c *HostDefiner) GenerateServiceAccount() *corev1.ServiceAccount {
	secrets := []corev1.LocalObjectReference{}
	if len(c.Spec.ImagePullSecrets) > 0 {
		for _, s := range c.Spec.ImagePullSecrets {
			secrets = append(secrets, corev1.LocalObjectReference{Name: s})
		}
	}

	return &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      config.GetNameForResource(config.HostDefinerServiceAccount, c.Name),
			Namespace: c.Namespace,
			Labels:    c.GetLabels(),
		},
		ImagePullSecrets: secrets,
	}
}
