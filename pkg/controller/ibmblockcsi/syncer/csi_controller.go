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
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/IBM/ibm-block-csi-operator/pkg/config"
	"github.com/IBM/ibm-block-csi-operator/pkg/internal/ibmblockcsi"
	"github.com/presslabs/controller-util/mergo/transformers"
	"github.com/presslabs/controller-util/syncer"
)

const (
	socketVolumeName                       = "socket-dir"
	controllerContainerName                = "ibm-block-csi-controller"
	controllerDriverRegistrarContainerName = "cluster-driver-registrar"
	provisionerContainerName               = "csi-provisioner"
	attacherContainerName                  = "csi-attacher"
	controllerLivenessProbeContainerName   = "liveness-probe"

	controllerContainerHealthPortName   = "healthz"
	controllerContainerHealthPortNumber = 9808
)

var controllerContainerHealthPort = intstr.FromInt(controllerContainerHealthPortNumber)

type csiControllerSyncer struct {
	driver *ibmblockcsi.IBMBlockCSI
	obj    runtime.Object
}

// NewCSIControllerSyncer returns a syncer for CSI controller
func NewCSIControllerSyncer(c client.Client, scheme *runtime.Scheme, driver *ibmblockcsi.IBMBlockCSI) syncer.Interface {
	obj := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:        config.GetNameForResource(config.CSIController, driver.Name),
			Namespace:   driver.Namespace,
			Annotations: driver.GetAnnotations(),
			Labels:      driver.GetLabels(),
		},
	}

	sync := &csiControllerSyncer{
		driver: driver,
		obj:    obj,
	}

	return syncer.NewObjectSyncer(config.CSIController.String(), driver.Unwrap(), obj, c, scheme, func() error {
		return sync.SyncFn()
	})
}

func (s *csiControllerSyncer) SyncFn() error {
	out := s.obj.(*appsv1.StatefulSet)

	out.Spec.Selector = metav1.SetAsLabelSelector(s.driver.GetCSIControllerSelectorLabels())
	out.Spec.ServiceName = config.GetNameForResource(config.CSIController, s.driver.Name)

	// ensure template
	out.Spec.Template.ObjectMeta.Labels = s.driver.GetCSIControllerPodLabels()
	out.Spec.Template.ObjectMeta.Annotations = s.driver.GetAnnotations()

	err := mergo.Merge(&out.Spec.Template.Spec, s.ensurePodSpec(), mergo.WithTransformers(transformers.PodSpec))
	if err != nil {
		return err
	}

	return nil
}

func (s *csiControllerSyncer) ensurePodSpec() corev1.PodSpec {
	fsGroup := config.ControllerUserID
	return corev1.PodSpec{
		Containers: s.ensureContainersSpec(),
		Volumes:    s.ensureVolumes(),
		SecurityContext: &corev1.PodSecurityContext{
			FSGroup:   &fsGroup,
			RunAsUser: &fsGroup,
		},
		Affinity: &corev1.Affinity{
			NodeAffinity: ensureNodeAffinity(),
		},
		ServiceAccountName: config.GetNameForResource(config.CSIControllerServiceAccount, s.driver.Name),
	}
}

func (s *csiControllerSyncer) ensureContainersSpec() []corev1.Container {
	// controller plugin container
	controllerPlugin := s.ensureContainer(controllerContainerName,
		s.driver.GetCSIControllerImage(),
		[]string{"--csi-endpoint=$(CSI_ENDPOINT)"},
	)
	controllerPlugin.Ports = ensurePorts(corev1.ContainerPort{
		Name:          controllerContainerHealthPortName,
		ContainerPort: controllerContainerHealthPortNumber,
	})

	//controllerPlugin.Resources = ensureResources(controllerContainerName)

	// cluster driver registrar sidecar
	registrar := s.ensureContainer(controllerDriverRegistrarContainerName,
		config.ClusterDriverRegistrarImage,
		[]string{"--csi-address=$(ADDRESS)", "--v=5"},
	)

	// csi provisioner sidecar
	provisioner := s.ensureContainer(provisionerContainerName,
		config.CSIProvisionerImage,
		[]string{"--csi-address=$(ADDRESS)", "--v=5"},
	)

	// csi attacher sidecar
	attacherImage := config.CSIAttacherImage
	attacher := s.ensureContainer(attacherContainerName,
		attacherImage,
		[]string{"--csi-address=$(ADDRESS)", "--v=5"},
	)

	// liveness probe sidecar
	livenessProbe := s.ensureContainer(controllerLivenessProbeContainerName,
		config.CSILivenessProbeImage,
		[]string{
			"--csi-address=/csi/csi.sock",
		},
	)

	return []corev1.Container{
		controllerPlugin,
		registrar,
		provisioner,
		attacher,
		livenessProbe,
	}
}

func ensureDefaultResources() corev1.ResourceRequirements {
	return ensureResources("50m", "100m", "50Mi", "100Mi")
}

func ensureResources(cpuRequests, cpuLimits, memoryRequests, memoryLimits string) corev1.ResourceRequirements {
	requests := corev1.ResourceList{
		corev1.ResourceRequestsCPU:    resource.MustParse(cpuRequests),
		corev1.ResourceRequestsMemory: resource.MustParse(memoryRequests),
	}
	limits := corev1.ResourceList{
		corev1.ResourceLimitsCPU:    resource.MustParse(cpuLimits),
		corev1.ResourceLimitsMemory: resource.MustParse(memoryLimits),
	}

	return corev1.ResourceRequirements{
		Limits:   limits,
		Requests: requests,
	}
}

func ensureNodeAffinity() *corev1.NodeAffinity {
	return &corev1.NodeAffinity{
		RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
			NodeSelectorTerms: []corev1.NodeSelectorTerm{
				{
					MatchExpressions: []corev1.NodeSelectorRequirement{
						{
							Key:      "beta.kubernetes.io/arch",
							Operator: corev1.NodeSelectorOpIn,
							Values:   []string{"amd64"},
						},
					},
				},
			},
		},
	}
}

func (s *csiControllerSyncer) ensureContainer(name, image string, args []string) corev1.Container {
	sc := &corev1.SecurityContext{}
	fillSecurityContextCapabilities(sc)
	return corev1.Container{
		Name:            name,
		Image:           image,
		ImagePullPolicy: "IfNotPresent",
		Args:            args,
		//EnvFrom:         s.getEnvSourcesFor(name),
		Env:             s.getEnvFor(name),
		VolumeMounts:    s.getVolumeMountsFor(name),
		SecurityContext: sc,
		Resources:       ensureDefaultResources(),
		LivenessProbe: ensureProbe(10, 3, 2, corev1.Handler{
			HTTPGet: &corev1.HTTPGetAction{
				Path:   "/healthz",
				Port:   controllerContainerHealthPort,
				Scheme: corev1.URISchemeHTTP,
			},
		}),
	}
}

func (s *csiControllerSyncer) envVarFromSecret(sctName, name, key string, opt bool) corev1.EnvVar {
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

func (s *csiControllerSyncer) getEnvFor(name string) []corev1.EnvVar {

	switch name {
	case controllerContainerName:
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

	case controllerDriverRegistrarContainerName, provisionerContainerName, attacherContainerName:
		return []corev1.EnvVar{
			{
				Name:  "ADDRESS",
				Value: config.ControllerSocketPath,
			},
		}
	}
	return nil
}

func (s *csiControllerSyncer) getVolumeMountsFor(name string) []corev1.VolumeMount {
	switch name {
	case controllerContainerName, controllerDriverRegistrarContainerName, provisionerContainerName, attacherContainerName:
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

func (s *csiControllerSyncer) ensureVolumes() []corev1.Volume {
	return []corev1.Volume{
		ensureVolume(socketVolumeName, corev1.VolumeSource{
			EmptyDir: &corev1.EmptyDirVolumeSource{},
		}),
	}
}

func ensurePorts(ports ...corev1.ContainerPort) []corev1.ContainerPort {
	return ports
}

func ensureProbe(delay, timeout, period int32, handler corev1.Handler) *corev1.Probe {
	return &corev1.Probe{
		InitialDelaySeconds: delay,
		TimeoutSeconds:      timeout,
		PeriodSeconds:       period,
		Handler:             handler,
		FailureThreshold:    5,
	}
}

func ensureVolume(name string, source corev1.VolumeSource) corev1.Volume {
	return corev1.Volume{
		Name:         name,
		VolumeSource: source,
	}
}
