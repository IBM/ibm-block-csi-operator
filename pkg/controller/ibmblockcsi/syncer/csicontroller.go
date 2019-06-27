package syncer

import (
	"github.com/imdario/mergo"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/kubernetes/pkg/apis/core"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/IBM/ibm-block-csi-driver-operator/pkg/config"
	"github.com/IBM/ibm-block-csi-driver-operator/pkg/internal/ibmblockcsi"
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

	controllerContainerHealthPortName = "healthz"
	controllerContainerHealthPort     = 9808
)

type csiControllerSyncer struct {
	driver *ibmblockcsi.IBMBlockCSI
}

// NewCSIControllerSyncer returns a syncer for CSI controller
func NewCSIControllerSyncer(c client.Client, scheme *runtime.Scheme, driver *ibmblockcsi.IBMBlockCSI) syncer.Interface {
	obj := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name: config.GetNameForResource(config.CSIController),
			//Namespace: driver.Namespace,
			Namespace:   config.DefaultNamespace,
			Annotations: driver.GetAnnotations(),
		},
	}

	sync := &csiControllerSyncer{
		driver: driver,
	}

	return syncer.NewObjectSyncer(config.CSIController, driver.Unwrap(), obj, c, scheme, func(in runtime.Object) error {
		return sync.SyncFn(in)
	})
}

func (s *csiControllerSyncer) SyncFn(in runtime.Object) error {
	out := in.(*appsv1.StatefulSet)

	out.Spec.Selector = metav1.SetAsLabelSelector(s.driver.GetCSIControllerComponentAnnotations())
	out.Spec.ServiceName = config.GetNameForResource(config.CSIController)

	// ensure template
	out.Spec.Template.ObjectMeta.Labels = s.driver.GetCSIControllerComponentAnnotations()

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
		ServiceAccountName: config.GetNameForResource(config.CSIControllerServiceAccount),
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
		ContainerPort: controllerContainerHealthPort,
	})

	//controllerPlugin.Resources = ensureResources(controllerContainerName)

	controllerPlugin.LivenessProbe = ensureProbe(10, 3, 2, corev1.Handler{
		HTTPGet: &corev1.HTTPGetAction{
			Path:   "/healthz",
			Port:   controllerContainerHealthPort,
			Scheme: corev1.URISchemeHTTP,
		},
	})

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
	attacher := s.ensureContainer(attacherContainerName,
		config.CSIAttacherImage,
		[]string{"--csi-address=$(ADDRESS)", "--v=5"},
	)

	// liveness probe sidecar
	livenessProbe := s.ensureContainer(controllerLivenessProbeContainerName,
		config.CSILivenessProbeImage,
		[]string{
			"--csi-address=/csi/csi.sock",
			"--connection-timeout=3s",
		},
	)

	return []core.Container{
		controllerPlugin,
		registrar,
		provisioner,
		attacher,
		livenessProbe,
	}
}

func (s *csiControllerSyncer) ensureContainer(name, image string, args []string) corev1.Container {
	return corev1.Container{
		Name:            name,
		Image:           image,
		ImagePullPolicy: "IfNotPresent",
		Args:            args,
		//EnvFrom:         s.getEnvSourcesFor(name),
		Env:          s.getEnvFor(name),
		VolumeMounts: s.getVolumeMountsFor(name),
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
				MountPath: ControllerSocketVolumeMountPath,
			},
		}

	case controllerLivenessProbeContainerName:
		return []corev1.VolumeMount{
			{
				Name:      socketVolumeName,
				MountPath: ControllerLivenessProbeContainerSocketVolumeMountPath,
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

func ensureResources(name string) corev1.ResourceRequirements {
	limits := corev1.ResourceList{
		corev1.ResourceCPU: resource.MustParse("50m"),
	}
	requests := corev1.ResourceList{
		corev1.ResourceCPU: resource.MustParse("10m"),
	}

	switch name {
	case containerExporterName:
		limits = corev1.ResourceList{
			corev1.ResourceCPU: resource.MustParse("100m"),
		}
	}

	return corev1.ResourceRequirements{
		Limits:   limits,
		Requests: requests,
	}
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
