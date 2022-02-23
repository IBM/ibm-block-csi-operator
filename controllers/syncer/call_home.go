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
	batchv1 "k8s.io/api/batch/v1"
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
	callHomeContainerName = "ibm-block-csi-call-home"
	secretVolumeName      = "call-home-secret-dir"
	CronSchedule          = "0 0 * * *" // runs once a day at midnight
	jobsHistoryLimit      = int32(1)
)

type callHomeSyncer struct {
	driver *ibmblockcsi.IBMBlockCSI
	obj    runtime.Object
}

// NewCallHomeSyncer returns a syncer for call home
func NewCallHomeSyncer(c client.Client, scheme *runtime.Scheme, driver *ibmblockcsi.IBMBlockCSI) syncer.Interface {
	obj := &batchv1.CronJob{
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
	out := s.obj.(*batchv1.CronJob)

	out.Spec.Schedule = CronSchedule

	// ensure template
	out.Spec.JobTemplate.ObjectMeta.Labels = s.driver.GetCallHomePodLabels()
	out.Spec.JobTemplate.ObjectMeta.Annotations = s.driver.GetAnnotations("", "")
	out.Spec.JobTemplate.Spec.Template.ObjectMeta.Labels = s.driver.GetCallHomePodLabels()
	successfulJobsHistoryLimit := jobsHistoryLimit
	out.Spec.SuccessfulJobsHistoryLimit = &successfulJobsHistoryLimit

	err := mergo.Merge(&out.Spec.JobTemplate.Spec.Template.Spec, s.ensurePodSpec(), mergo.WithTransformers(transformers.PodSpec))
	if err != nil {
		return err
	}

	return nil
}

func (s *callHomeSyncer) ensurePodSpec() corev1.PodSpec {
	return corev1.PodSpec{
		Containers:         s.ensureContainersSpec(),
		Volumes:            s.ensureVolumes(),
		ServiceAccountName: config.GetNameForResource(config.CallHomeServiceAccount, s.driver.Name),
		Affinity:           s.driver.Spec.CallHome.Affinity,
		Tolerations:        s.driver.Spec.CallHome.Tolerations,
		RestartPolicy:      "OnFailure",
	}
}

func (s *callHomeSyncer) ensureContainersSpec() []corev1.Container {
	callHomeContainer := s.ensureContainer(callHomeContainerName,
		s.driver.GetCallHomeImage(),
		[]string{""},
	)

	callHomeContainer.Resources = ensureResources("40m", "800m", "40Mi", "400Mi")

	callHomeContainer.ImagePullPolicy = s.driver.Spec.CallHome.ImagePullPolicy

	return []corev1.Container{
		callHomeContainer,
	}
}

func (s *callHomeSyncer) ensureContainer(name, image string, args []string) corev1.Container {
	sc := &corev1.SecurityContext{AllowPrivilegeEscalation: boolptr.False()}
	fillSecurityContextCapabilities(sc)
	return corev1.Container{
		Name:            name,
		Image:           image,
		Args:            args,
		Env:             s.getEnv(),
		VolumeMounts:    s.getVolumeMounts(),
		SecurityContext: sc,
		Resources:       ensureDefaultResources(),
	}
}

func (s *callHomeSyncer) getEnv() []corev1.EnvVar {
	return []corev1.EnvVar{
		{
			Name:  "CSI_LOGLEVEL",
			Value: config.DefaultLogLevel,
		},
	}

}

func (s *callHomeSyncer) getVolumeMounts() []corev1.VolumeMount {
	return []corev1.VolumeMount{
		{
			Name:      secretVolumeName,
			MountPath: config.CallHomeSecretVolumeMountPath,
		},
	}

}

func (s *callHomeSyncer) ensureVolumes() []corev1.Volume {
	return []corev1.Volume{
		ensureVolume(secretVolumeName, corev1.VolumeSource{
			Secret: &corev1.SecretVolumeSource{SecretName: s.driver.Spec.CallHome.SecretName},
		}),
	}
}
