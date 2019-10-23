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

const (
	NodeDriverRegistrarImage    = "quay.io/k8scsi/csi-node-driver-registrar:v1.2.0"
	CSIProvisionerImage         = "quay.io/k8scsi/csi-provisioner:v1.3.0"
	CSIAttacherImage            = "quay.io/k8scsi/csi-attacher:v1.2.1"
	CSILivenessProbeImage       = "quay.io/k8scsi/livenessprobe:v1.1.0"

	ControllerTag = "1.0.0"
	NodeTag       = "1.0.0"
	NodeAgentTag  = "1.0.0"

	DefaultNamespace = "kube-system"
	DefaultLogLevel  = "DEBUG"
	ControllerUserID = int64(9999)

	NodeAgentPort = "10086"
)
