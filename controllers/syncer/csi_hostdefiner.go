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
	"strconv"

	"github.com/IBM/ibm-block-csi-operator/controllers/internal/hostdefiner"
	"github.com/IBM/ibm-block-csi-operator/pkg/config"
	"github.com/IBM/ibm-block-csi-operator/pkg/util/boolptr"
	"github.com/imdario/mergo"
	"github.com/presslabs/controller-util/pkg/mergo/transformers"
	"github.com/presslabs/controller-util/pkg/syncer"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	HostDefinerContainerName = "ibm-block-csi-host-definer"
)

type hostDefinerSyncer struct {
	driver *hostdefiner.HostDefiner
	obj    runtime.Object
}

func NewHostDefinerSyncer(c client.Client, scheme *runtime.Scheme, driver *hostdefiner.HostDefiner) syncer.Interface {
	obj := getDeploymentSkeleton(driver)

	sync := &hostDefinerSyncer{
		driver: driver,
		obj:    obj,
	}

	return syncer.NewObjectSyncer(config.HostDefiner.String(), driver.Unwrap(), obj, c, func() error {
		return sync.SyncFn()
	})
}

func getDeploymentSkeleton(driver *hostdefiner.HostDefiner) *appsv1.Deployment {
	obj := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:        config.GetNameForResource(config.HostDefiner, driver.Name),
			Namespace:   driver.Namespace,
			Annotations: driver.GetAnnotations(),
			Labels:      driver.GetLabels(),
		},
		Spec: appsv1.DeploymentSpec{
			Selector: metav1.SetAsLabelSelector(driver.GetHostDefinerSelectorLabels()),
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      driver.GetHostDefinerPodLabels(),
					Annotations: driver.GetAnnotations(),
				},
				Spec: corev1.PodSpec{},
			},
		},
	}
	return obj
}

func (s *hostDefinerSyncer) SyncFn() error {
	out := s.obj.(*appsv1.Deployment)
	out.Spec.Selector = metav1.SetAsLabelSelector(s.driver.GetHostDefinerSelectorLabels())
	labels := s.driver.GetHostDefinerPodLabels()
	out.Spec.Template.ObjectMeta.Labels = labels
	out.ObjectMeta.Labels = labels
	ensureAnnotations(&out.Spec.Template.ObjectMeta, &out.ObjectMeta, s.driver.GetAnnotations())

	err := mergo.Merge(&out.Spec.Template.Spec, s.ensurePodSpec(), mergo.WithTransformers(transformers.PodSpec))
	if err != nil {
		return err
	}

	return nil
}

func (s *hostDefinerSyncer) ensurePodSpec() corev1.PodSpec {
	return corev1.PodSpec{
		Containers:         s.ensureContainersSpec(),
		Affinity:           s.driver.Spec.HostDefiner.Affinity,
		Tolerations:        s.driver.Spec.HostDefiner.Tolerations,
		ServiceAccountName: config.GetNameForResource(config.HostDefinerServiceAccount, s.driver.Name),
	}
}

func (s *hostDefinerSyncer) ensureContainersSpec() []corev1.Container {
	hostDefinerPlugin := s.ensureContainer(HostDefinerContainerName,
		s.driver.GetHostDefinerImage(),
		[]string{},
	)

	hostDefinerPlugin.Resources = ensureResources("40m", "800m", "40Mi", "400Mi")

	hostDefinerPlugin.ImagePullPolicy = s.driver.Spec.HostDefiner.ImagePullPolicy

	return []corev1.Container{
		hostDefinerPlugin,
	}
}

func (s *hostDefinerSyncer) ensureContainer(name, image string, args []string) corev1.Container {
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

func (s *hostDefinerSyncer) getEnv() []corev1.EnvVar {
	return []corev1.EnvVar{
		{
			Name:  "PREFIX",
			Value: s.driver.Spec.HostDefiner.Prefix,
		},
		{
			Name:  "CONNECTIVITY_TYPE",
			Value: s.driver.Spec.HostDefiner.ConnectivityType,
		},
		{
			Name:  "ALLOW_DELETE",
			Value: strconv.FormatBool(s.driver.Spec.HostDefiner.AllowDelete),
		},
		{
			Name:  "DYNAMIC_NODE_LABELING",
			Value: strconv.FormatBool(s.driver.Spec.HostDefiner.DynamicNodeLabeling),
		},
                {
                        Name:  "PORT_SET",
                        Value: s.driver.Spec.HostDefiner.PortSet,
                },
	}
}
