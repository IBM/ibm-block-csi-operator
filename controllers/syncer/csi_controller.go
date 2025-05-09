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
	"math"
	"strconv"
	os "runtime"

	"github.com/imdario/mergo"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	csiv1 "github.com/IBM/ibm-block-csi-operator/api/v1"
	"github.com/IBM/ibm-block-csi-operator/controllers/internal/crutils"
	"github.com/IBM/ibm-block-csi-operator/pkg/config"
	"github.com/IBM/ibm-block-csi-operator/pkg/util/boolptr"
	"github.com/presslabs/controller-util/pkg/mergo/transformers"
	"github.com/presslabs/controller-util/pkg/syncer"
)

const (
	socketVolumeName                     = "socket-dir"
	ControllerContainerName              = "ibm-block-csi-controller"
	provisionerContainerName             = "csi-provisioner"
	attacherContainerName                = "csi-attacher"
	snapshotterContainerName             = "csi-snapshotter"
	resizerContainerName                 = "csi-resizer"
	replicatorContainerName              = "csi-addons-replicator"
	volumeGroupContainerName             = "csi-volume-group"
	controllerLivenessProbeContainerName = "livenessprobe"

	commonMaxWorkersFlag  = "--worker-threads"
	resizerMaxWorkersFlag = "--workers"

	controllerContainerHealthPortName          = "healthz"
	controllerContainerDefaultHealthPortNumber = 9808
)

var TopologyEnabled = false

type csiControllerSyncer struct {
	driver *crutils.IBMBlockCSI
	obj    runtime.Object
}

// NewCSIControllerSyncer returns a syncer for CSI controller
func NewCSIControllerSyncer(c client.Client, scheme *runtime.Scheme, driver *crutils.IBMBlockCSI) syncer.Interface {
	obj := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:        config.GetNameForResource(config.CSIController, driver.Name),
			Namespace:   driver.Namespace,
			Annotations: driver.GetAnnotations("", ""),
			Labels:      driver.GetLabels(),
		},
		Spec: appsv1.StatefulSetSpec{
			Selector: metav1.SetAsLabelSelector(driver.GetCSIControllerSelectorLabels()),
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      driver.GetCSIControllerPodLabels(),
					Annotations: driver.GetAnnotations("", ""),
				},
				Spec: corev1.PodSpec{},
			},
		},
	}

	sync := &csiControllerSyncer{
		driver: driver,
		obj:    obj,
	}

	return syncer.NewObjectSyncer(config.CSIController.String(), driver.Unwrap(), obj, c, func() error {
		return sync.SyncFn()
	})
}

func (s *csiControllerSyncer) SyncFn() error {
	out := s.obj.(*appsv1.StatefulSet)

	out.Spec.Selector = metav1.SetAsLabelSelector(s.driver.GetCSIControllerSelectorLabels())
	out.Spec.ServiceName = config.GetNameForResource(config.CSIController, s.driver.Name)

	controllerLabels := s.driver.GetCSIControllerPodLabels()
	controllerAnnotations := s.driver.GetAnnotations("", "")

	// ensure template
	out.Spec.Template.ObjectMeta.Labels = controllerLabels

	out.ObjectMeta.Labels = controllerLabels
	ensureAnnotations(&out.Spec.Template.ObjectMeta, &out.ObjectMeta, controllerAnnotations)

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
		Affinity:           s.driver.Spec.Controller.Affinity,
		Tolerations:        s.driver.Spec.Controller.Tolerations,
		ServiceAccountName: config.GetNameForResource(config.CSIControllerServiceAccount, s.driver.Name),
	}
}

func (s *csiControllerSyncer) ensureContainersSpec() []corev1.Container {
	controllerPlugin := s.ensureContainer(ControllerContainerName,
		s.driver.GetCSIControllerImage(),
		[]string{"--csi-endpoint=$(CSI_ENDPOINT)"},
	)

	controllerPlugin.Resources = ensureResources("40m", "800m", "40Mi", "400Mi")

	healthPort := s.driver.Spec.HealthPort
	if healthPort == 0 {
		healthPort = controllerContainerDefaultHealthPortNumber
	}

	controllerPlugin.Ports = ensurePorts(corev1.ContainerPort{
		Name:          controllerContainerHealthPortName,
		ContainerPort: int32(healthPort),
	})
	controllerPlugin.ImagePullPolicy = s.driver.Spec.Controller.ImagePullPolicy

	controllerContainerHealthPort := intstr.FromInt(int(healthPort))
	controllerPlugin.LivenessProbe = ensureProbe(10, 100, 5, corev1.ProbeHandler{
		HTTPGet: &corev1.HTTPGetAction{
			Path:   "/healthz",
			Port:   controllerContainerHealthPort,
			Scheme: corev1.URISchemeHTTP,
		},
	})

	maxWorkersFlag := getCommonMaxWorkersFlag()

	provisionerArgs := []string{
		"--csi-address=$(ADDRESS)",
		"--v=5",
		"--timeout=120s",
		"--default-fstype=ext4",
		maxWorkersFlag,
	}
	if TopologyEnabled {
		provisionerArgs = append(provisionerArgs, "--feature-gates=Topology=true")
	}
	provisioner := s.ensureContainer(provisionerContainerName,
		s.getCSIProvisionerImage(),
		provisionerArgs,
	)
	provisioner.ImagePullPolicy = s.getCSIProvisionerPullPolicy()

	attacher := s.ensureContainer(attacherContainerName,
		s.getCSIAttacherImage(),
		[]string{"--csi-address=$(ADDRESS)", "--v=5", "--timeout=180s", maxWorkersFlag},
	)
	attacher.ImagePullPolicy = s.getCSIAttacherPullPolicy()

	snapshotter := s.ensureContainer(snapshotterContainerName,
		s.getCSISnapshotterImage(),
		[]string{
			"--csi-address=$(ADDRESS)",
			"--v=5",
			"--timeout=120s",
			maxWorkersFlag,
		},
	)
	snapshotter.ImagePullPolicy = s.getCSISnapshotterPullPolicy()

	resizer := s.ensureContainer(resizerContainerName,
		s.getCSIResizerImage(),
		[]string{
			"--csi-address=$(ADDRESS)",
			"--v=5",
			"--timeout=30s",
			"--handle-volume-inuse-error=false",
			getResizerMaxWorkersFlag(),
		},
	)
	resizer.ImagePullPolicy = s.getCSIResizerPullPolicy()

	leaderElectionNamespaceFlag := fmt.Sprintf("--leader-election-namespace=%s", s.driver.Namespace)
	driverNameFlag := fmt.Sprintf("--driver-name=%s", config.DriverName)
	replicator := s.ensureContainer(replicatorContainerName,
		s.getCSIAddonsReplicatorImage(),
		[]string{leaderElectionNamespaceFlag, driverNameFlag,
			"--csi-address=$(ADDRESS)", "--zap-log-level=5", "--rpc-timeout=30s"},
	)
	replicator.ImagePullPolicy = s.getCSIAddonsReplicatorPullPolicy()

	volumegroup := s.ensureContainer(volumeGroupContainerName,
		s.getCSIVolumeGroupImage(),
		[]string{
			driverNameFlag,
			"--csi-address=$(ADDRESS)",
			"--rpc-timeout=30s",
			"--multiple-vgs-to-pvc=false",
			"--disable-delete-pvcs=true",
		})
	volumegroup.ImagePullPolicy = s.getCSIVolumeGroupPullPolicy()

	healthPortArg := fmt.Sprintf("--health-port=%v", healthPort)
	livenessProbe := s.ensureContainer(controllerLivenessProbeContainerName,
		s.getLivenessProbeImage(),
		[]string{
			"--csi-address=/csi/csi.sock",
			healthPortArg,
		},
	)
	livenessProbe.ImagePullPolicy = s.getLivenessProbePullPolicy()

	return []corev1.Container{
		controllerPlugin,
		provisioner,
		attacher,
		snapshotter,
		resizer,
		replicator,
		volumegroup,
		livenessProbe,
	}
}

func ensureDefaultResources() corev1.ResourceRequirements {
	return ensureResources("20m", "200m", "20Mi", "200Mi")
}

func ensureResources(cpuRequests, cpuLimits, memoryRequests, memoryLimits string) corev1.ResourceRequirements {
	requests := corev1.ResourceList{
		corev1.ResourceCPU:    resource.MustParse(cpuRequests),
		corev1.ResourceMemory: resource.MustParse(memoryRequests),
	}
	limits := corev1.ResourceList{
		corev1.ResourceCPU:    resource.MustParse(cpuLimits),
		corev1.ResourceMemory: resource.MustParse(memoryLimits),
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
							Key:      "kubernetes.io/arch",
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
	case ControllerContainerName:
		return []corev1.EnvVar{
			{
				Name:  "CSI_ENDPOINT",
				Value: config.CSIEndpoint,
			},
			{
				Name:  "CSI_LOGLEVEL",
				Value: config.DefaultLogLevel,
			},
			{
				Name:  "ENABLE_CALL_HOME",
				Value: s.driver.Spec.EnableCallHome,
			},
			{
				Name:  "ODF_VERSION_FOR_CALL_HOME",
				Value: s.driver.Spec.ODFVersionForCallHome,
			},
			{
				// TODO consider a different type of port. now uint16
				Name:  "SVC_SSH_PORT",
				Value: strconv.FormatUint(uint64(s.driver.Spec.SvcSshPort), 10),
			},
		}

	case provisionerContainerName, attacherContainerName, snapshotterContainerName,
		resizerContainerName, replicatorContainerName, volumeGroupContainerName:
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
	case ControllerContainerName, provisionerContainerName, attacherContainerName, snapshotterContainerName,
		resizerContainerName, replicatorContainerName, volumeGroupContainerName:
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

func (s *csiControllerSyncer) getSidecarByName(name string) *csiv1.CSISidecar {
	return getSidecarByName(s.driver, name)
}

func (s *csiControllerSyncer) getSidecarImageByName(name string) string {
	sidecar := s.getSidecarByName(name)
	if sidecar != nil {
		return fmt.Sprintf("%s:%s", sidecar.Repository, sidecar.Tag)
	}
	return s.driver.GetDefaultSidecarImageByName(name)
}

func (s *csiControllerSyncer) getCSIAttacherImage() string {
	return s.getSidecarImageByName(config.CSIAttacher)
}

func (s *csiControllerSyncer) getCSIProvisionerImage() string {
	return s.getSidecarImageByName(config.CSIProvisioner)
}

func (s *csiControllerSyncer) getLivenessProbeImage() string {
	return s.getSidecarImageByName(config.LivenessProbe)
}

func (s *csiControllerSyncer) getCSISnapshotterImage() string {
	return s.getSidecarImageByName(config.CSISnapshotter)
}

func (s *csiControllerSyncer) getCSIResizerImage() string {
	return s.getSidecarImageByName(config.CSIResizer)
}

func (s *csiControllerSyncer) getCSIAddonsReplicatorImage() string {
	return s.getSidecarImageByName(config.CSIAddonsReplicator)
}

func (s *csiControllerSyncer) getCSIVolumeGroupImage() string {
	return s.getSidecarImageByName(config.CSIVolumeGroup)
}

func (s *csiControllerSyncer) getSidecarPullPolicy(sidecarName string) corev1.PullPolicy {
	sidecar := s.getSidecarByName(sidecarName)
	if sidecar != nil && sidecar.ImagePullPolicy != "" {
		return sidecar.ImagePullPolicy
	}
	return corev1.PullIfNotPresent
}

func (s *csiControllerSyncer) getCSIAttacherPullPolicy() corev1.PullPolicy {
	return s.getSidecarPullPolicy(config.CSIAttacher)
}

func (s *csiControllerSyncer) getCSIProvisionerPullPolicy() corev1.PullPolicy {
	return s.getSidecarPullPolicy(config.CSIProvisioner)
}

func (s *csiControllerSyncer) getLivenessProbePullPolicy() corev1.PullPolicy {
	return s.getSidecarPullPolicy(config.LivenessProbe)
}

func (s *csiControllerSyncer) getCSISnapshotterPullPolicy() corev1.PullPolicy {
	return s.getSidecarPullPolicy(config.CSISnapshotter)
}

func (s *csiControllerSyncer) getCSIResizerPullPolicy() corev1.PullPolicy {
	return s.getSidecarPullPolicy(config.CSIResizer)
}

func (s *csiControllerSyncer) getCSIAddonsReplicatorPullPolicy() corev1.PullPolicy {
	return s.getSidecarPullPolicy(config.CSIAddonsReplicator)
}

func (s *csiControllerSyncer) getCSIVolumeGroupPullPolicy() corev1.PullPolicy {
	return s.getSidecarPullPolicy(config.CSIVolumeGroup)
}

func ensurePorts(ports ...corev1.ContainerPort) []corev1.ContainerPort {
	return ports
}

func ensureProbe(delay, timeout, period int32, handler corev1.ProbeHandler) *corev1.Probe {
	return &corev1.Probe{
		InitialDelaySeconds: delay,
		TimeoutSeconds:      timeout,
		PeriodSeconds:       period,
		ProbeHandler:        handler,
		SuccessThreshold:    1,
		FailureThreshold:    30,
	}
}

func ensureVolume(name string, source corev1.VolumeSource) corev1.Volume {
	return corev1.Volume{
		Name:         name,
		VolumeSource: source,
	}
}

func getSidecarByName(driver *crutils.IBMBlockCSI, name string) *csiv1.CSISidecar {
	for _, sidecar := range driver.Spec.Sidecars {
		if sidecar.Name == name {
			return &sidecar
		}
	}
	return nil
}

func getMaxWorkersCount() int {
	cpuCount := os.NumCPU()
	maxWorkers := math.Min(float64(cpuCount), 32) / 2
	return int(math.Max(maxWorkers, 2))
}

func getMaxWorkersFlag(flag string) string {
	maxWorkersCount := getMaxWorkersCount()
	return fmt.Sprintf("%s=%d", flag, maxWorkersCount)
}

func getCommonMaxWorkersFlag() string {
	return getMaxWorkersFlag(commonMaxWorkersFlag)
}

func getResizerMaxWorkersFlag() string {
	return getMaxWorkersFlag(resizerMaxWorkersFlag)
}
