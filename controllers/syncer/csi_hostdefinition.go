/**
 * Copyright 2022 IBM Corp.
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

package syncer

import (
	"github.com/IBM/ibm-block-csi-operator/controllers/internal/hostdefinition"
	"github.com/IBM/ibm-block-csi-operator/pkg/config"
	"github.com/IBM/ibm-block-csi-operator/pkg/util/boolptr"
	"github.com/imdario/mergo"
	"github.com/presslabs/controller-util/mergo/transformers"
	"github.com/presslabs/controller-util/syncer"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	HostDefinitionContainerName = "ibm-block-csi-hostdefinition"
)

type csiHostDefinitionSyncer struct {
	driver *hostdefinition.HostDefinition
	obj    runtime.Object
}

func NewCSIHostDefinitionSyncer(c client.Client, scheme *runtime.Scheme, driver *hostdefinition.HostDefinition) syncer.Interface {
	obj := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:        config.GetNameForResource(config.CSIHostDefinition, driver.Name),
			Namespace:   driver.Namespace,
			Annotations: driver.GetAnnotations("", ""),
			Labels:      driver.GetLabels(),
		},
	}

	sync := &csiHostDefinitionSyncer{
		driver: driver,
		obj:    obj,
	}

	return syncer.NewObjectSyncer(config.CSIHostDefinition.String(), driver.Unwrap(), obj, c, func() error {
		return sync.SyncFn()
	})
}

func (s *csiHostDefinitionSyncer) SyncFn() error {
	out := s.obj.(*appsv1.Deployment)

	out.Spec.Selector = metav1.SetAsLabelSelector(s.driver.GetCSIHostDefinitionSelectorLabels())

	labels := s.driver.GetCSIHostDefinitionPodLabels()
	annotations := s.driver.GetAnnotations("", "")

	out.Spec.Template.ObjectMeta.Labels = labels
	out.Spec.Template.ObjectMeta.Annotations = annotations

	out.ObjectMeta.Labels = labels
	out.ObjectMeta.Annotations = annotations

	err := mergo.Merge(&out.Spec.Template.Spec, s.ensurePodSpec(), mergo.WithTransformers(transformers.PodSpec))
	if err != nil {
		return err
	}

	return nil
}

func (s *csiHostDefinitionSyncer) ensurePodSpec() corev1.PodSpec {
	return corev1.PodSpec{
		Containers:         s.ensureContainersSpec(),
		Affinity:           s.driver.Spec.HostDefinition.Affinity,
		Tolerations:        s.driver.Spec.HostDefinition.Tolerations,
		ServiceAccountName: config.GetNameForResource(config.CSIHostDefinitionServiceAccount, s.driver.Name),
	}
}

func (s *csiHostDefinitionSyncer) ensureContainersSpec() []corev1.Container {
	hostDefinitionPlugin := s.ensureContainer(HostDefinitionContainerName,
		s.driver.GetCSIHostDefinitionImage(),
		[]string{},
	)

	hostDefinitionPlugin.Resources = ensureResources("40m", "800m", "40Mi", "400Mi")

	hostDefinitionPlugin.ImagePullPolicy = s.driver.Spec.HostDefinition.ImagePullPolicy

	return []corev1.Container{
		hostDefinitionPlugin,
	}
}

func (s *csiHostDefinitionSyncer) ensureContainer(name, image string, args []string) corev1.Container {
	sc := &corev1.SecurityContext{AllowPrivilegeEscalation: boolptr.False()}
	fillSecurityContextCapabilities(sc)
	return corev1.Container{
		Name:            name,
		Image:           image,
		Args:            args,
		SecurityContext: sc,
		Resources:       ensureDefaultResources(),
	}
}
