/**
 * Copyright 2022 IBM Corp.
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

package hostdefiner

import (
	"fmt"

	csiv1 "github.com/IBM/ibm-block-csi-operator/api/v1"
	"github.com/IBM/ibm-block-csi-operator/controllers/internal/common"
	"github.com/IBM/ibm-block-csi-operator/pkg/config"
	csiversion "github.com/IBM/ibm-block-csi-operator/version"
	"k8s.io/apimachinery/pkg/labels"
)

type HostDefiner struct {
	*csiv1.HostDefiner
}

func New(hd *csiv1.HostDefiner) *HostDefiner {
	return &HostDefiner{
		HostDefiner: hd,
	}
}

func (hd *HostDefiner) Unwrap() *csiv1.HostDefiner {
	return hd.HostDefiner
}

func (hd *HostDefiner) GetHostDefinerPodLabels() labels.Set {
	return labels.Merge(hd.GetLabels(), hd.GetHostDefinerSelectorLabels())
}

func (hd *HostDefiner) GetLabels() labels.Set {
	labels := labels.Set{
		"app.kubernetes.io/name":       config.ProductName,
		"app.kubernetes.io/instance":   hd.Name,
		"app.kubernetes.io/version":    csiversion.Version,
		"app.kubernetes.io/managed-by": config.Name,
		"csi":                          "ibm",
		"product":                      config.ProductName,
		"release":                      fmt.Sprintf("v%s", csiversion.Version),
	}

	if hd.Labels != nil {
		for k, v := range hd.Labels {
			if !labels.Has(k) {
				labels[k] = v
			}
		}
	}

	return labels
}

func (hd *HostDefiner) GetHostDefinerSelectorLabels() labels.Set {
	return common.GetSelectorLabels(config.HostDefiner.String())
}

func (hd *HostDefiner) GetAnnotations() labels.Set {
	labels := labels.Set{
		"productID":      config.ProductName,
		"productName":    config.ProductName,
		"productVersion": csiversion.Version,
	}

	if hd.Annotations != nil {
		for k, v := range hd.Annotations {
			if !labels.Has(k) {
				labels[k] = v
			}
		}
	}

	return labels
}

func (hd *HostDefiner) GetHostDefinerImage() string {
	if hd.Spec.HostDefiner.Tag == "" {
		return hd.Spec.HostDefiner.Repository
	}
	return hd.Spec.HostDefiner.Repository + ":" + hd.Spec.HostDefiner.Tag
}
