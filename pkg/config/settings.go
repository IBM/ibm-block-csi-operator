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

package config

import "k8s.io/apimachinery/pkg/util/sets"

const (
	NodeDriverRegistrarImage = "quay.io/k8scsi/csi-node-driver-registrar:v1.2.0"
	CSIProvisionerImage      = "quay.io/k8scsi/csi-provisioner:v1.4.0"
	CSIAttacherImage         = "quay.io/k8scsi/csi-attacher:v1.2.1"
	CSILivenessProbeImage    = "quay.io/k8scsi/livenessprobe:v1.1.0"

	OpenShiftNodeDriverRegistrarImage = "registry.redhat.io/openshift4/ose-csi-driver-registrar:v4.3"
	OpenShiftCSIProvisionerImage      = "registry.redhat.io/openshift4/ose-csi-external-provisioner-rhel7:v4.3"
	OpenShiftCSIAttacherImage         = "registry.redhat.io/openshift4/ose-csi-external-attacher:v4.3"
	OpenShiftCSILivenessProbeImage    = "registry.redhat.io/openshift4/ose-csi-livenessprobe:v4.3"

	ControllerTag = "1.1.0"
	NodeTag       = "1.1.0"
	NodeAgentTag  = "1.0.0"

	DefaultNamespace = "kube-system"
	DefaultLogLevel  = "DEBUG"
	ControllerUserID = int64(9999)

	NodeAgentPort = "10086"
)

var ReplaceControllerVersions = sets.String{
	ControllerRepository + ":" + "1.0.0": sets.Empty{},
}

var ReplaceNodeVersions = sets.String{
	NodeRepository + ":" + "1.0.0": sets.Empty{},
}

var ReplaceCSIProvisionerVersions = sets.String{
	"quay.io/k8scsi/csi-provisioner:v1.3.0": sets.Empty{},
}

var ReplaceCSIAttacherVersions = sets.String{}

var ReplaceNodeDriverRegistrarVersions = sets.String{}

var ReplaceLivenessProbeVersions = sets.String{}

var ReplaceOpenShiftControllerVersions = sets.String{
	OpenShiftControllerRepository + ":" + "1.0.0": sets.Empty{},
}

var ReplaceOpenShiftNodeVersions = sets.String{
	OpenShiftNodeRepository + ":" + "1.0.0": sets.Empty{},
}

var ReplaceOpenShiftCSIProvisionerVersions = sets.String{
	"quay.io/k8scsi/csi-provisioner:v1.3.0": sets.Empty{},
}

var ReplaceOpenShiftCSIAttacherVersions = sets.String{
	"quay.io/k8scsi/csi-attacher:v1.2.1": sets.Empty{},
}

var ReplaceOpenShiftNodeDriverRegistrarVersions = sets.String{
	"quay.io/k8scsi/csi-node-driver-registrar:v1.2.0": sets.Empty{},
}

var ReplaceOpenShiftLivenessProbeVersions = sets.String{
	"quay.io/k8scsi/livenessprobe:v1.1.0": sets.Empty{},
}

func GetReplaceVersions(platform, image string) sets.String {
	switch platform {
	case OpenShift:
		switch image {
		case ControllerImage:
			return ReplaceOpenShiftControllerVersions
		case NodeImage:
			return ReplaceOpenShiftNodeVersions
		case CSIProvisioner:
			return ReplaceOpenShiftCSIProvisionerVersions
		case CSIAttacher:
			return ReplaceOpenShiftCSIAttacherVersions
		case CSINodeDriverRegistrar:
			return ReplaceOpenShiftNodeDriverRegistrarVersions
		case LivenessProbe:
			return ReplaceOpenShiftLivenessProbeVersions
		default:
			return sets.String{}
		}
	case Kubernetes:
		switch image {
		case ControllerImage:
			return ReplaceControllerVersions
		case NodeImage:
			return ReplaceNodeVersions
		case CSIProvisioner:
			return ReplaceCSIProvisionerVersions
		case CSIAttacher:
			return ReplaceCSIAttacherVersions
		case CSINodeDriverRegistrar:
			return ReplaceNodeDriverRegistrarVersions
		case LivenessProbe:
			return ReplaceLivenessProbeVersions
		default:
			return sets.String{}
		}
	default:
		return sets.String{}
	}
}
