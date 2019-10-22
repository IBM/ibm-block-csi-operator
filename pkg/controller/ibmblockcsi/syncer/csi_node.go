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
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/IBM/ibm-block-csi-operator/pkg/config"
	"github.com/IBM/ibm-block-csi-operator/pkg/internal/ibmblockcsi"
	"github.com/IBM/ibm-block-csi-operator/pkg/util/boolptr"
	"github.com/presslabs/controller-util/mergo/transformers"
	"github.com/presslabs/controller-util/syncer"
)

const (
	registrationVolumeName           = "registration-dir"
	nodeContainerName                = "ibm-block-csi-node"
	nodeDriverRegistrarContainerName = "node-driver-registrar"
	nodeLivenessProbeContainerName   = "liveness-probe"

	nodeContainerHealthPortName   = "healthz"
	nodeContainerHealthPortNumber = 9808

	registrationVolumeMountPath = "/registration"
)

var nodeContainerHealthPort = intstr.FromInt(nodeContainerHealthPortNumber)

type csiNodeSyncer struct {
	driver *ibmblockcsi.IBMBlockCSI
	obj    runtime.Object
}

// NewCSINodeSyncer returns a syncer for CSI node
func NewCSINodeSyncer(c client.Client, scheme *runtime.Scheme, driver *ibmblockcsi.IBMBlockCSI) syncer.Interface {
	obj := &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:        config.GetNameForResource(config.CSINode, driver.Name),
			Namespace:   driver.Namespace,
			Annotations: driver.GetAnnotations(),
			Labels:      driver.GetLabels(),
		},
	}

	sync := &csiNodeSyncer{
		driver: driver,
		obj:    obj,
	}

	return syncer.NewObjectSyncer(config.CSINode.String(), driver.Unwrap(), obj, c, scheme, func() error {
		return sync.SyncFn()
	})
}

func (s *csiNodeSyncer) SyncFn() error {
	out := s.obj.(*appsv1.DaemonSet)

	out.Spec.Selector = metav1.SetAsLabelSelector(s.driver.GetCSINodeSelectorLabels())

	// ensure template
	out.Spec.Template.ObjectMeta.Labels = s.driver.GetCSINodePodLabels()
	out.Spec.Template.ObjectMeta.Annotations = s.driver.GetAnnotations()

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
		HostNetwork:        true,
		DNSPolicy:          "ClusterFirstWithHostNet", // To have DNS options set along with hostNetwork, you have to specify DNS policy explicitly to 'ClusterFirstWithHostNet'
		ServiceAccountName: "default",
		Affinity: &corev1.Affinity{
			NodeAffinity: ensureNodeAffinity(),
		},
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
	nodePlugin.Ports = ensurePorts(corev1.ContainerPort{
		Name:          nodeContainerHealthPortName,
		ContainerPort: nodeContainerHealthPortNumber,
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
	registrar := s.ensureContainer(nodeDriverRegistrarContainerName,
		config.NodeDriverRegistrarImage,
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

	// liveness probe sidecar
	livenessProbe := s.ensureContainer(nodeLivenessProbeContainerName,
		config.CSILivenessProbeImage,
		[]string{
			"--csi-address=/csi/csi.sock",
		},
	)
	livenessProbe.SecurityContext = &corev1.SecurityContext{AllowPrivilegeEscalation: boolptr.False()}
	fillSecurityContextCapabilities(livenessProbe.SecurityContext)

	return []corev1.Container{
		nodePlugin,
		registrar,
		livenessProbe,
	}
}

func (s *csiNodeSyncer) ensureContainer(name, image string, args []string) corev1.Container {
	return corev1.Container{
		Name:            name,
		Image:           image,
		ImagePullPolicy: "IfNotPresent",
		Args:            args,
		Env:             s.getEnvFor(name),
		VolumeMounts:    s.getVolumeMountsFor(name),
		Resources:       ensureDefaultResources(),
		LivenessProbe: ensureProbe(10, 3, 10, corev1.Handler{
			HTTPGet: &corev1.HTTPGetAction{
				Path:   "/healthz",
				Port:   nodeContainerHealthPort,
				Scheme: corev1.URISchemeHTTP,
			},
		}),
	}
}

func envVarFromField(name, fieldPath string) corev1.EnvVar {
	env := corev1.EnvVar{
		Name: name,
		ValueFrom: &corev1.EnvVarSource{
			FieldRef: &corev1.ObjectFieldSelector{
				FieldPath: fieldPath,
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

	case nodeDriverRegistrarContainerName:
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
		}

	case nodeDriverRegistrarContainerName:
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
	}
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
