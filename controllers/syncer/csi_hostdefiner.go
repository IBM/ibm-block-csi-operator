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
	"github.com/IBM/ibm-block-csi-operator/controllers/internal/hostdefiner"
	"github.com/IBM/ibm-block-csi-operator/pkg/config"
	"github.com/IBM/ibm-block-csi-operator/pkg/util/boolptr"
	csiversion "github.com/IBM/ibm-block-csi-operator/version"
	"github.com/imdario/mergo"
	"github.com/presslabs/controller-util/mergo/transformers"
	"github.com/presslabs/controller-util/syncer"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	HostDefinerContainerName = "ibm-block-csi-hostdefiner"
)

type csiHostDefinerSyncer struct {
	driver *hostdefiner.HostDefiner
	obj    runtime.Object
}

var defaultAnnotations = labels.Set{
	"productID":      config.ProductName,
	"productName":    config.ProductName,
	"productVersion": csiversion.Version,
}

func NewCSIHostDefinerSyncer(c client.Client, scheme *runtime.Scheme, driver *hostdefiner.HostDefiner) syncer.Interface {
	obj := getDeploymentSkeleton(driver)

	sync := &csiHostDefinerSyncer{
		driver: driver,
		obj:    obj,
	}

	return syncer.NewObjectSyncer(config.CSIHostDefiner.String(), driver.Unwrap(), obj, c, func() error {
		return sync.SyncFn()
	})
}

func getDeploymentSkeleton(driver *hostdefiner.HostDefiner) *appsv1.Deployment {
	obj := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:        config.GetNameForResource(config.CSIHostDefiner, driver.Name),
			Namespace:   driver.Namespace,
			Annotations: driver.GetAnnotations("", ""),
			Labels:      driver.GetLabels(),
		},
		Spec: appsv1.DeploymentSpec{
			Selector: metav1.SetAsLabelSelector(driver.GetCSIHostDefinerSelectorLabels()),
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      driver.GetCSIHostDefinerPodLabels(),
					Annotations: driver.GetAnnotations("", ""),
				},
				Spec: corev1.PodSpec{},
			},
		},
	}
	return obj
}

func (s *csiHostDefinerSyncer) SyncFn() error {
	out := s.obj.(*appsv1.Deployment)
	out.Spec.Selector = metav1.SetAsLabelSelector(s.driver.GetCSIHostDefinerSelectorLabels())
	labels := s.driver.GetCSIHostDefinerPodLabels()
	out.Spec.Template.ObjectMeta.Labels = labels
	out.ObjectMeta.Labels = labels
	s.ensureAnnotations(out)

	err := mergo.Merge(&out.Spec.Template.Spec, s.ensurePodSpec(), mergo.WithTransformers(transformers.PodSpec))
	if err != nil {
		return err
	}

	return nil
}

func (s *csiHostDefinerSyncer) ensureAnnotations(deployment *appsv1.Deployment) {
	annotations := s.driver.GetAnnotations("", "")
	for k, _ := range defaultAnnotations {
		deployment.Spec.Template.ObjectMeta.Annotations[k] = annotations[k]
		deployment.ObjectMeta.Annotations[k] = annotations[k]
	}
}

func (s *csiHostDefinerSyncer) ensurePodSpec() corev1.PodSpec {
	return corev1.PodSpec{
		Containers:         s.ensureContainersSpec(),
		Affinity:           s.driver.Spec.HostDefiner.Affinity,
		Tolerations:        s.driver.Spec.HostDefiner.Tolerations,
		ServiceAccountName: config.GetNameForResource(config.CSIHostDefinerServiceAccount, s.driver.Name),
	}
}

func (s *csiHostDefinerSyncer) ensureContainersSpec() []corev1.Container {
	hostDefinerPlugin := s.ensureContainer(HostDefinerContainerName,
		s.driver.GetCSIHostDefinerImage(),
		[]string{},
	)

	hostDefinerPlugin.Resources = ensureResources("40m", "800m", "40Mi", "400Mi")

	hostDefinerPlugin.ImagePullPolicy = s.driver.Spec.HostDefiner.ImagePullPolicy

	return []corev1.Container{
		hostDefinerPlugin,
	}
}

func (s *csiHostDefinerSyncer) ensureContainer(name, image string, args []string) corev1.Container {
	sc := &corev1.SecurityContext{AllowPrivilegeEscalation: boolptr.False()}
	fillSecurityContextCapabilities(sc)
	return corev1.Container{
		Name:            name,
		Image:           image,
		Args:            args,
		Env:             s.getEnv(),
		SecurityContext: sc,
		Resources:       ensureDefaultResources(),
	}
}

func (s *csiHostDefinerSyncer) getEnv() []corev1.EnvVar {
	return []corev1.EnvVar{
		{
			Name:  "PREFIX",
			Value: s.getPrefix(),
		},
		{
			Name:  "CONNECTION",
			Value: s.getConnection(),
		},
	}
}

func (s *csiHostDefinerSyncer) getPrefix() string {
	return s.driver.Spec.HostDefiner.Prefix
}

func (s *csiHostDefinerSyncer) getConnection() string {
	return s.driver.Spec.HostDefiner.Connectivity
}
