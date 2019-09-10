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
	"strconv"

	"github.com/imdario/mergo"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/IBM/ibm-block-csi-operator/pkg/config"
	operatorconfig "github.com/IBM/ibm-block-csi-operator/pkg/internal/config"
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

type nodeAgentSyncer struct {
	operatorConfig *operatorconfig.Config
}

// NewNodeAgentSyncer returns a syncer for node agent
func NewNodeAgentSyncer(c client.Client, scheme *runtime.Scheme, operatorConfig *operatorconfig.Config) syncer.Interface {
	obj := &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      config.GetNameForResource(config.NodeAgent, operatorConfig.Name),
			Namespace: operatorConfig.Namespace,
			// Annotations: operatorConfig.GetAnnotations(),
		},
	}

	sync := &nodeAgentSyncer{
		operatorConfig: operatorConfig,
	}

	return syncer.NewObjectSyncer(config.NodeAgent.String(), operatorConfig.Unwrap(), obj, c, scheme, func(in runtime.Object) error {
		return sync.SyncFn(in)
	})
}

func (s *nodeAgentSyncer) SyncFn(in runtime.Object) error {
	out := in.(*appsv1.DaemonSet)

	out.Spec.Selector = metav1.SetAsLabelSelector(s.operatorConfig.GetNodeAgentPodLabels())
	out.Spec.Template.ObjectMeta.Labels = s.operatorConfig.GetNodeAgentPodLabels()

	err := mergo.Merge(&out.Spec.Template.Spec, s.ensurePodSpec(), mergo.WithTransformers(transformers.PodSpec))
	if err != nil {
		return err
	}

	return nil
}

func (s *nodeAgentSyncer) ensurePodSpec() corev1.PodSpec {
	return corev1.PodSpec{
		Containers:  s.ensureContainersSpec(),
		Volumes:     s.ensureVolumes(),
		HostNetwork: true,
		DNSPolicy:   "ClusterFirstWithHostNet", // To have DNS options set along with hostNetwork, you have to specify DNS policy explicitly to 'ClusterFirstWithHostNet'
	}
}

func (s *nodeAgentSyncer) ensureContainersSpec() []corev1.Container {
	port, _ := strconv.Atoi(s.operatorConfig.Spec.NodeAgent.Port)

	nodeAgent := s.ensureContainer("node-agent",
		s.operatorConfig.GetNodeAgentImage(),
	)
	nodeAgent.Ports = ensurePorts(corev1.ContainerPort{
		Name:          "grpc",
		ContainerPort: int32(port),
	})

	nodeAgent.SecurityContext = &corev1.SecurityContext{
		Privileged: boolptr.True(),
		Capabilities: &corev1.Capabilities{
			Add: []corev1.Capability{"SYS_ADMIN"},
		},
		AllowPrivilegeEscalation: boolptr.True(),
	}

	return []corev1.Container{
		nodeAgent,
	}
}

func (s *nodeAgentSyncer) ensureContainer(name, image string) corev1.Container {
	return corev1.Container{
		Name:            name,
		Image:           image,
		ImagePullPolicy: "IfNotPresent",
		Env:             s.getEnvFor(name),
		VolumeMounts:    s.getVolumeMountsFor(name),
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

func (s *nodeAgentSyncer) getEnvFor(name string) []corev1.EnvVar {
	return []corev1.EnvVar{
		{
			Name:  "ADDRESS",
			Value: fmt.Sprintf(":%s", s.operatorConfig.Spec.NodeAgent.Port),
		},
		envVarFromField("NODE_NAME", "spec.nodeName"),
	}
}

func (s *nodeAgentSyncer) getVolumeMountsFor(name string) []corev1.VolumeMount {
	return []corev1.VolumeMount{
		{
			Name:      "lib-modules",
			MountPath: "/lib/modules",
		},
		{
			Name:      "sysfs",
			MountPath: "/sys",
		},
		{
			Name:      "dev",
			MountPath: "/dev",
		},
		{
			Name:      "iscsi",
			MountPath: "/etc/iscsi",
		},
	}
}

func (s *nodeAgentSyncer) ensureVolumes() []corev1.Volume {
	return []corev1.Volume{
		ensureVolume("lib-modules", ensureHostPathVolumeSource("/lib/modules", "Directory")),
		ensureVolume("sysfs", ensureHostPathVolumeSource("/sys", "Directory")),
		ensureVolume("dev", ensureHostPathVolumeSource("/dev", "Directory")),
		ensureVolume("iscsi", ensureHostPathVolumeSource("/etc/iscsi", "DirectoryOrCreate")),
	}
}

func ensureVolume(name string, source corev1.VolumeSource) corev1.Volume {
	return corev1.Volume{
		Name:         name,
		VolumeSource: source,
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

func ensurePorts(ports ...corev1.ContainerPort) []corev1.ContainerPort {
	return ports
}
