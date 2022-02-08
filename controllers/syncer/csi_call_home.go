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

package syncer

import (
	"github.com/imdario/mergo"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/IBM/ibm-block-csi-operator/controllers/internal/ibmblockcsi"
	"github.com/IBM/ibm-block-csi-operator/pkg/config"
	"github.com/IBM/ibm-block-csi-operator/pkg/util/boolptr"
	"github.com/presslabs/controller-util/mergo/transformers"
	"github.com/presslabs/controller-util/syncer"
)

const (
	callHomeContainerName              = "ibm-block-call-home"
	callHomeLivenessProbeContainerName = "liveness-probe"
)

type callHomeSyncer struct {
	driver *ibmblockcsi.IBMBlockCSI
	obj    runtime.Object
}

// NewCallHomeSyncer returns a syncer for call home
func NewCallHomeSyncer(c client.Client, scheme *runtime.Scheme, driver *ibmblockcsi.IBMBlockCSI) syncer.Interface {
	obj := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:        config.GetNameForResource(config.CallHome, driver.Name),
			Namespace:   driver.Namespace,
			Annotations: driver.GetAnnotations("", ""),
			Labels:      driver.GetLabels(),
		},
	}

	sync := &callHomeSyncer{
		driver: driver,
		obj:    obj,
	}

	return syncer.NewObjectSyncer(config.CallHome.String(), driver.Unwrap(), obj, c, func() error {
		return sync.SyncFn()
	})
}

func (s *callHomeSyncer) SyncFn() error {
	out := s.obj.(*appsv1.StatefulSet)

	out.Spec.Selector = metav1.SetAsLabelSelector(s.driver.GetCallHomeSelectorLabels())
	out.Spec.ServiceName = config.GetNameForResource(config.CallHome, s.driver.Name)

	// ensure template
	out.Spec.Template.ObjectMeta.Labels = s.driver.GetCallHomePodLabels()
	out.Spec.Template.ObjectMeta.Annotations = s.driver.GetAnnotations("", "")

	err := mergo.Merge(&out.Spec.Template.Spec, s.ensurePodSpec(), mergo.WithTransformers(transformers.PodSpec))
	if err != nil {
		return err
	}

	return nil
}

func (s *callHomeSyncer) ensurePodSpec() corev1.PodSpec {
	fsGroup := config.ControllerUserID
	return corev1.PodSpec{
		Containers: s.ensureContainersSpec(),
		Volumes:    s.ensureVolumes(),
		SecurityContext: &corev1.PodSecurityContext{
			FSGroup:   &fsGroup,
			RunAsUser: &fsGroup,
		},
		ImagePullSecrets: s.getImagePullSecrets(),
		Affinity:         s.driver.Spec.CallHome.Affinity,
		Tolerations:      s.driver.Spec.CallHome.Tolerations,
	}
}

func (s *callHomeSyncer) getImagePullSecrets() []corev1.LocalObjectReference {
	var secrets []corev1.LocalObjectReference
	for _, s := range s.driver.Spec.CallHome.ImagePullSecrets {
		secrets = append(secrets, corev1.LocalObjectReference{Name: s})
	}
	return secrets
}

func (s *callHomeSyncer) ensureContainersSpec() []corev1.Container {
	callHomePlugin := s.ensureContainer(callHomeContainerName,
		s.driver.GetCallHomeImage(),
		[]string{"--csi-endpoint=$(CSI_ENDPOINT)"},
	)

	callHomePlugin.Resources = ensureResources("40m", "800m", "40Mi", "400Mi")

	healthPort := s.driver.Spec.HealthPort
	if healthPort == 0 {
		healthPort = controllerContainerDefaultHealthPortNumber
	}

	callHomePlugin.Ports = ensurePorts(corev1.ContainerPort{
		Name:          controllerContainerHealthPortName,
		ContainerPort: int32(healthPort),
	})
	callHomePlugin.ImagePullPolicy = s.driver.Spec.CallHome.ImagePullPolicy

	return []corev1.Container{
		callHomePlugin,
	}
}

func (s *callHomeSyncer) ensureContainer(name, image string, args []string) corev1.Container {
	sc := &corev1.SecurityContext{AllowPrivilegeEscalation: boolptr.False()}
	fillSecurityContextCapabilities(sc)
	return corev1.Container{
		Name:  name,
		Image: image,
		Args:  args,
		//EnvFrom:         s.getEnvSourcesFor(name),
		Env:             s.getEnvFor(name),
		VolumeMounts:    s.getVolumeMountsFor(name),
		SecurityContext: sc,
		Resources:       ensureDefaultResources(),
	}
}

func (s *callHomeSyncer) envVarFromSecret(sctName, name, key string, opt bool) corev1.EnvVar {
	env := corev1.EnvVar{
		Name: name,
		ValueFrom: &corev1.EnvVarSource{
			SecretKeyRef: &corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: sctName,
				},
				Key:      key,
				Optional: &opt,
			},
		},
	}
	return env
}

func (s *callHomeSyncer) getEnvFor(name string) []corev1.EnvVar {

	switch name {
	case callHomeContainerName:
		return []corev1.EnvVar{
			{
				Name:  "CSI_ENDPOINT",
				Value: config.CSIEndpoint,
			},
			{
				Name:  "CSI_LOGLEVEL",
				Value: config.DefaultLogLevel,
			},
		}

	case provisionerContainerName, attacherContainerName, snapshotterContainerName,
		resizerContainerName, replicatorContainerName:
		return []corev1.EnvVar{
			{
				Name:  "ADDRESS",
				Value: config.ControllerSocketPath,
			},
		}
	}
	return nil
}

func (s *callHomeSyncer) getVolumeMountsFor(name string) []corev1.VolumeMount {
	switch name {
	case controllerContainerName, provisionerContainerName, attacherContainerName, snapshotterContainerName,
		resizerContainerName, replicatorContainerName:
		return []corev1.VolumeMount{
			{
				Name:      socketVolumeName,
				MountPath: config.ControllerSocketVolumeMountPath,
			},
		}

	case controllerLivenessProbeContainerName:
		return []corev1.VolumeMount{
			{
				Name:      socketVolumeName,
				MountPath: config.ControllerLivenessProbeContainerSocketVolumeMountPath,
			},
		}
	}
	return nil
}

func (s *callHomeSyncer) ensureVolumes() []corev1.Volume {
	return []corev1.Volume{
		ensureVolume(socketVolumeName, corev1.VolumeSource{
			EmptyDir: &corev1.EmptyDirVolumeSource{},
		}),
	}
}
