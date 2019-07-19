package syncer

import (
	"github.com/imdario/mergo"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/IBM/ibm-block-csi-driver-operator/pkg/config"
	"github.com/IBM/ibm-block-csi-driver-operator/pkg/internal/ibmblockcsi"
	"github.com/IBM/ibm-block-csi-driver-operator/pkg/util/boolptr"
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
}

// NewCSINodeSyncer returns a syncer for CSI node
func NewCSINodeSyncer(c client.Client, scheme *runtime.Scheme, driver *ibmblockcsi.IBMBlockCSI) syncer.Interface {
	obj := &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:        config.GetNameForResource(config.CSINode, driver.Name),
			Namespace:   driver.Namespace,
			Annotations: driver.GetAnnotations(),
		},
	}

	sync := &csiNodeSyncer{
		driver: driver,
	}

	return syncer.NewObjectSyncer(config.CSINode.String(), driver.Unwrap(), obj, c, scheme, func(in runtime.Object) error {
		return sync.SyncFn(in)
	})
}

func (s *csiNodeSyncer) SyncFn(in runtime.Object) error {
	out := in.(*appsv1.DaemonSet)

	out.Spec.Selector = metav1.SetAsLabelSelector(s.driver.GetCSINodeComponentAnnotations())

	// ensure template
	out.Spec.Template.ObjectMeta.Labels = s.driver.GetCSINodeComponentAnnotations()

	err := mergo.Merge(&out.Spec.Template.Spec, s.ensurePodSpec(), mergo.WithTransformers(transformers.PodSpec))
	if err != nil {
		return err
	}

	return nil
}

func (s *csiNodeSyncer) ensurePodSpec() corev1.PodSpec {
	return corev1.PodSpec{
		Containers:  s.ensureContainersSpec(),
		Volumes:     s.ensureVolumes(),
		HostNetwork: true,
		DNSPolicy:   "ClusterFirstWithHostNet", // To have DNS options set along with hostNetwork, you have to specify DNS policy explicitly to 'ClusterFirstWithHostNet'
		//ServiceAccountName: config.GetNameForResource(config.CSINodeServiceAccount, s.driver.Name),
	}
}

func (s *csiNodeSyncer) ensureContainersSpec() []corev1.Container {
	// node plugin container
	nodePlugin := s.ensureContainer(nodeContainerName,
		s.driver.GetCSINodeImage(),
		[]string{
			"--csi-endpoint=$(CSI_ENDPOINT)",
			"--v=$(CSI_LOGLEVEL)",
		},
	)
	nodePlugin.Ports = ensurePorts(corev1.ContainerPort{
		Name:          nodeContainerHealthPortName,
		ContainerPort: nodeContainerHealthPortNumber,
	})

	nodePlugin.SecurityContext = &corev1.SecurityContext{
		Privileged: boolptr.True(),
	}

	//nodePlugin.Resources = ensureResources(nodeContainerName)

	nodePlugin.LivenessProbe = ensureProbe(10, 3, 10, corev1.Handler{
		HTTPGet: &corev1.HTTPGetAction{
			Path:   "/healthz",
			Port:   nodeContainerHealthPort,
			Scheme: corev1.URISchemeHTTP,
		},
	})

	// cluster driver registrar sidecar
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

	// liveness probe sidecar
	livenessProbe := s.ensureContainer(nodeLivenessProbeContainerName,
		config.CSILivenessProbeImage,
		[]string{
			"--csi-address=/csi/csi.sock",
			"--connection-timeout=3s",
		},
	)

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
				Value: "5",
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
	switch name {
	case nodeContainerName:
		return []corev1.VolumeMount{
			{
				Name:      socketVolumeName,
				MountPath: config.NodeSocketVolumeMountPath,
			},
			{
				Name:      "mountpoint-dir",
				MountPath: "/var/lib/kubelet/pods",
			},
			{
				Name:      "device-dir",
				MountPath: "/dev",
			},
			{
				Name:      "iscsi-dir",
				MountPath: "/etc/iscsi",
			},
			{
				Name:      "sys-dir",
				MountPath: "/sys",
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
		ensureVolume("socket-dir", ensureHostPathVolumeSource("/var/lib/kubelet/plugins/ibm-block-csi-driver", "DirectoryOrCreate")),
		ensureVolume("registration-dir", ensureHostPathVolumeSource("/var/lib/kubelet/plugins_registry", "Directory")),
		ensureVolume("device-dir", ensureHostPathVolumeSource("/dev", "Directory")),
		ensureVolume("iscsi-dir", ensureHostPathVolumeSource("/etc/iscsi", "Directory")),
		ensureVolume("sys-dir", ensureHostPathVolumeSource("/sys", "Directory")),
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
