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

package util

import (
	"context"
	"fmt"

	csiv1 "github.com/IBM/ibm-block-csi-operator/pkg/apis/csi/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func GetOperatorConfig(c client.Client) (*csiv1.Config, error) {
	configList := &csiv1.ConfigList{}
	fmt.Println(configList)
	err := c.List(context.TODO(), configList)
	if err != nil {
		return nil, err
	}
	if len(configList.Items) == 0 {
		return nil, fmt.Errorf("No operator configuration is found.")
	}
	return &(configList.Items[0]), nil
}

func IsDefineHostEnabled(c client.Client) bool {
	conf, err := GetOperatorConfig(c)
	if err != nil {
		return false
	}
	return conf.Spec.DefineHost
}

func IsNodeAgentReady(c client.Client) bool {
	conf, err := GetOperatorConfig(c)
	if err != nil {
		return false
	}
	return conf.Status.NodeAgent.Phase == csiv1.NodeAgentPhaseRunning
}
