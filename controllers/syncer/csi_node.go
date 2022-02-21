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
	"fmt"

	"github.com/imdario/mergo"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	csiv1 "github.com/IBM/ibm-block-csi-operator/api/v1"
	"github.com/IBM/ibm-block-csi-operator/controllers/internal/ibmblockcsi"
	"github.com/IBM/ibm-block-csi-operator/pkg/config"
	"github.com/IBM/ibm-block-csi-operator/pkg/util/boolptr"
	"github.com/presslabs/controller-util/mergo/transformers"
	"github.com/presslabs/controller-util/syncer"
)

const (
	registrationVolumeName              = "registration-dir"
	nodeContainerName                   = "ibm-block-csi-node"
	csiNodeDriverRegistrarContainerName = "csi-node-driver-registrar"
	nodeLivenessProbeContainerName      = "livenessprobe"

	nodeContainerHealthPortName          = "healthz"
	nodeContainerDefaultHealthPortNumber = 9808

	registrationVolumeMountPath = "/registration"
)

type csiNodeSyncer struct {
	driver *ibmblockcsi.IBMBlockCSI
	obj    runtime.Object
}

// NewCSINodeSyncer returns a syncer for CSI node
func NewCSINodeSyncer(c client.Client, scheme *runtime.Scheme, driver *ibmblockcsi.IBMBlockCSI,
	daemonSetRestartedKey string, daemonSetRestartedValue string) syncer.Interface {
	obj := &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:        config.GetNameForResource(config.CSINode, driver.Name),
			Namespace:   driver.Namespace,
			Annotations: driver.GetAnnotations(daemonSetRestartedKey, daemonSetRestartedValue),
			Labels:      driver.GetLabels(),
		},
	}

	sync := &csiNodeSyncer{
		driver: driver,
		obj:    obj,
	}

	return syncer.NewObjectSyncer(config.CSINode.String(), driver.Unwrap(), obj, c, func() error {
		return sync.SyncFn(daemonSetRestartedKey, daemonSetRestartedValue)
	})
}

func (s *csiNodeSyncer) SyncFn(daemonSetRestartedKey string, daemonSetRestartedValue string) error {
	out := s.obj.(*appsv1.DaemonSet)

	out.Spec.Selector = metav1.SetAsLabelSelector(s.driver.GetCSINodeSelectorLabels())

	// ensure template
	out.Spec.Template.ObjectMeta.Labels = s.driver.GetCSINodePodLabels()
	out.Spec.Template.ObjectMeta.Annotations = s.driver.GetAnnotations(daemonSetRestartedKey, daemonSetRestartedValue)

	err := mergo.Merge(&out.Spec.Template.Spec, s.ensurePodSpec(), mergo.WithTransformers(transformers.PodSpec))
	if err != nil {
		return err
	}

	return nil
}

func (s *csiNodeSyncer) ensurePodSpec() corev1.PodSpec {
	return corev1.PodSpec{
		Containers:         s.ensureContainersSpec(),
		Volumes:            s.ensureVolumes(),
		HostIPC:            true,
		HostNetwork:        true,
		ServiceAccountName: config.GetNameForResource(config.CSINodeServiceAccount, s.driver.Name),
		Affinity:           s.driver.Spec.Node.Affinity,
		Tolerations:        s.driver.Spec.Node.Tolerations,
	}
}

func (s *csiNodeSyncer) ensureContainersSpec() []corev1.Container {
	// node plugin container
	nodePlugin := s.ensureContainer(nodeContainerName,
		s.driver.GetCSINodeImage(),
		[]string{
			"--csi-endpoint=$(CSI_ENDPOINT)",
			"--hostname=$(KUBE_NODE_NAME)",
			"--config-file-path=./config.yaml",
			"--loglevel=$(CSI_LOGLEVEL)",
		},
	)

	nodePlugin.Resources = ensureResources("40m", "1000m", "40Mi", "400Mi")

	healthPort := s.driver.Spec.HealthPort
	if healthPort == 0 {
		healthPort = nodeContainerDefaultHealthPortNumber
	}
	nodePlugin.Ports = ensurePorts(corev1.ContainerPort{
		Name:          nodeContainerHealthPortName,
		ContainerPort: int32(healthPort),
	})

	nodePlugin.ImagePullPolicy = s.driver.Spec.Node.ImagePullPolicy

	nodeContainerHealthPort := intstr.FromInt(int(healthPort))
	nodePlugin.LivenessProbe = ensureProbe(10, 3, 10, corev1.Handler{
		HTTPGet: &corev1.HTTPGetAction{
			Path:   "/healthz",
			Port:   nodeContainerHealthPort,
			Scheme: corev1.URISchemeHTTP,
		},
	})

	nodePlugin.SecurityContext = &corev1.SecurityContext{
		Privileged:               boolptr.True(),
		AllowPrivilegeEscalation: boolptr.True(),
	}
	fillSecurityContextCapabilities(
		nodePlugin.SecurityContext,
		"CHOWN",
		"FSETID",
		"FOWNER",
		"SETGID",
		"SETUID",
		"DAC_OVERRIDE",
	)

	// node driver registrar sidecar
	registrar := s.ensureContainer(csiNodeDriverRegistrarContainerName,
		s.getCSINodeDriverRegistrarImage(),
		[]string{
			"--csi-address=$(ADDRESS)",
			"--kubelet-registration-path=$(DRIVER_REG_SOCK_PATH)",
			"--v=5",
		},
	)
	registrar.Lifecycle = &corev1.Lifecycle{
		PreStop: &corev1.Handler{
			Exec: &corev1.ExecAction{
				Command: []string{"/bin/sh", "-c", "rm -rf /registration/ibm-block-csi-driver-reg.sock /csi/csi.sock"},
			},
		},
	}
	registrar.SecurityContext = &corev1.SecurityContext{AllowPrivilegeEscalation: boolptr.False()}
	fillSecurityContextCapabilities(registrar.SecurityContext)
	registrar.ImagePullPolicy = s.getCSINodeDriverRegistrarPullPolicy()

	// liveness probe sidecar
	healthPortArg := fmt.Sprintf("--health-port=%v", healthPort)
	livenessProbe := s.ensureContainer(nodeLivenessProbeContainerName,
		s.getLivenessProbeImage(),
		[]string{
			"--csi-address=/csi/csi.sock",
			healthPortArg,
		},
	)
	livenessProbe.SecurityContext = &corev1.SecurityContext{AllowPrivilegeEscalation: boolptr.False()}
	fillSecurityContextCapabilities(livenessProbe.SecurityContext)
	livenessProbe.ImagePullPolicy = s.getCSINodeDriverRegistrarPullPolicy()

	return []corev1.Container{
		nodePlugin,
		registrar,
		livenessProbe,
	}
}

func (s *csiNodeSyncer) ensureContainer(name, image string, args []string) corev1.Container {
	return corev1.Container{
		Name:         name,
		Image:        image,
		Args:         args,
		Env:          s.getEnvFor(name),
		VolumeMounts: s.getVolumeMountsFor(name),
		Resources:    ensureDefaultResources(),
	}
}

func envVarFromField(name, fieldPath string) corev1.EnvVar {
	env := corev1.EnvVar{
		Name: name,
		ValueFrom: &corev1.EnvVarSource{
			FieldRef: &corev1.ObjectFieldSelector{
				APIVersion: config.APIVersion,
				FieldPath:  fieldPath,
			},
		},
	}
	return env
}

func (s *csiNodeSyncer) getEnvFor(name string) []corev1.EnvVar {

	switch name {
	case nodeContainerName:
		return []corev1.EnvVar{
			{
				Name:  "CSI_ENDPOINT",
				Value: config.CSINodeEndpoint,
			},
			{
				Name:  "CSI_LOGLEVEL",
				Value: "trace",
			},
			envVarFromField("KUBE_NODE_NAME", "spec.nodeName"),
		}

	case csiNodeDriverRegistrarContainerName:
		return []corev1.EnvVar{
			{
				Name:  "ADDRESS",
				Value: config.NodeSocketPath,
			},
			{
				Name:  "DRIVER_REG_SOCK_PATH",
				Value: config.NodeRegistrarSocketPath,
			},
		}
	}
	return nil
}

func (s *csiNodeSyncer) getVolumeMountsFor(name string) []corev1.VolumeMount {
	mountPropagationB := corev1.MountPropagationBidirectional

	switch name {
	case nodeContainerName:
		return []corev1.VolumeMount{
			{
				Name:      socketVolumeName,
				MountPath: config.NodeSocketVolumeMountPath,
			},
			{
				Name:             "mountpoint-dir",
				MountPath:        "/var/lib/kubelet/pods",
				MountPropagation: &mountPropagationB,
			},
			{
				Name:      "device-dir",
				MountPath: "/dev",
			},
			{
				Name:      "sys-dir",
				MountPath: "/sys",
			},
			{
				Name:             "host-dir",
				MountPath:        "/host",
				MountPropagation: &mountPropagationB,
			},
			{
				Name:      "iscsi",
				MountPath: "/etc/iscsi",
			},
		}

	case csiNodeDriverRegistrarContainerName:
		return []corev1.VolumeMount{
			{
				Name:      socketVolumeName,
				MountPath: config.NodeSocketVolumeMountPath,
			},
			{
				Name:      registrationVolumeName,
				MountPath: registrationVolumeMountPath,
			},
		}

	case nodeLivenessProbeContainerName:
		return []corev1.VolumeMount{
			{
				Name:      socketVolumeName,
				MountPath: config.NodeSocketVolumeMountPath,
			},
		}
	}
	return nil
}

func (s *csiNodeSyncer) ensureVolumes() []corev1.Volume {
	return []corev1.Volume{
		ensureVolume("mountpoint-dir", ensureHostPathVolumeSource("/var/lib/kubelet/pods", "Directory")),
		ensureVolume("socket-dir", ensureHostPathVolumeSource("/var/lib/kubelet/plugins/block.csi.ibm.com", "DirectoryOrCreate")),
		ensureVolume("registration-dir", ensureHostPathVolumeSource("/var/lib/kubelet/plugins_registry", "Directory")),
		ensureVolume("device-dir", ensureHostPathVolumeSource("/dev", "Directory")),
		ensureVolume("sys-dir", ensureHostPathVolumeSource("/sys", "Directory")),
		ensureVolume("host-dir", ensureHostPathVolumeSource("/", "Directory")),
		ensureVolume("iscsi", ensureHostPathVolumeSource("/etc/iscsi", "Directory")),
	}
}

func (s *csiNodeSyncer) getSidecarByName(name string) *csiv1.CSISidecar {
	return getSidecarByName(s.driver, name)
}

func (s *csiNodeSyncer) getCSINodeDriverRegistrarImage() string {
	sidecar := s.getSidecarByName(config.CSINodeDriverRegistrar)
	if sidecar != nil {
		return fmt.Sprintf("%s:%s", sidecar.Repository, sidecar.Tag)
	}
	return s.driver.GetDefaultSidecarImageByName(config.CSINodeDriverRegistrar)
}

func (s *csiNodeSyncer) getLivenessProbeImage() string {
	sidecar := s.getSidecarByName(config.LivenessProbe)
	if sidecar != nil {
		return fmt.Sprintf("%s:%s", sidecar.Repository, sidecar.Tag)
	}
	return s.driver.GetDefaultSidecarImageByName(config.LivenessProbe)
}

func (s *csiNodeSyncer) getCSINodeDriverRegistrarPullPolicy() corev1.PullPolicy {
	sidecar := s.getSidecarByName(config.CSINodeDriverRegistrar)
	if sidecar != nil && sidecar.ImagePullPolicy != "" {
		return sidecar.ImagePullPolicy
	}
	return corev1.PullIfNotPresent
}

func (s *csiNodeSyncer) getLivenessProbePullPolicy() corev1.PullPolicy {
	sidecar := s.getSidecarByName(config.LivenessProbe)
	if sidecar != nil && sidecar.ImagePullPolicy != "" {
		return sidecar.ImagePullPolicy
	}
	return corev1.PullIfNotPresent
}

func ensureHostPathVolumeSource(path, pathType string) corev1.VolumeSource {
	t := corev1.HostPathType(pathType)

	return corev1.VolumeSource{
		HostPath: &corev1.HostPathVolumeSource{
			Path: path,
			Type: &t,
		},
	}
}

func fillSecurityContextCapabilities(sc *corev1.SecurityContext, add ...string) {
	sc.Capabilities = &corev1.Capabilities{
		Drop: []corev1.Capability{"ALL"},
	}

	if len(add) > 0 {
		adds := []corev1.Capability{}
		for _, a := range add {
			adds = append(adds, corev1.Capability(a))
		}
		sc.Capabilities.Add = adds
	}
}
