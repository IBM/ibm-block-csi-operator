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
	"reflect"

	"google.golang.org/grpc"
	corev1 "k8s.io/api/core/v1"

	csiv1 "github.com/IBM/ibm-block-csi-operator/pkg/apis/csi/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Invoke calls an objet's method by name
func Invoke(any interface{}, name string, args ...interface{}) (values []reflect.Value, err error) {
	values = []reflect.Value{reflect.ValueOf(nil)}

	defer func() {
		if e := recover(); e != nil {
			err = e.(error)
		}
	}()

	method := reflect.ValueOf(any).MethodByName(name)
	methodType := method.Type()
	numIn := methodType.NumIn()

	if !methodType.IsVariadic() {
		if numIn != len(args) {
			return values, fmt.Errorf("Method %s must have %d params. Have %d", name, numIn, len(args))
		}
	} else {
		if numIn-1 > len(args) {
			return values, fmt.Errorf("Method %s must have minimum %d params. Have %d", name, numIn-1, len(args))
		}
	}

	in := make([]reflect.Value, len(args))
	for i := 0; i < len(args); i++ {
		var inType reflect.Type
		if methodType.IsVariadic() && i >= numIn-1 {
			inType = methodType.In(numIn - 1).Elem()
		} else {
			inType = methodType.In(i)
		}
		argValue := reflect.ValueOf(args[i])
		if !argValue.IsValid() {
			return values, fmt.Errorf("Method %s. Param[%d] must be %s. Have %s", name, i, inType, argValue.String())
		}
		argType := argValue.Type()
		if argType.ConvertibleTo(inType) {
			in[i] = argValue.Convert(inType)
		} else {
			return values, fmt.Errorf("Method %s. Param[%d] must be %s. Have %s", name, i, inType, argType)
		}
	}
	return method.Call(in), nil
}

// TestConnectivity tests the given addresses one by one and return the
// first successful one, if all failed, return empty
func TestConnectivity(addrs []string, port string) string {
	for _, addr := range addrs {
		var address string
		if port == "" {
			address = addr
		} else {
			address = fmt.Sprintf("%s:%s", addr, port)
		}
		conn, err := grpc.Dial(address, grpc.WithInsecure())
		if err == nil {
			conn.Close()
			return addr
		}
	}
	return ""
}

// GetNodeAddresses returns a node's addresses in a special order
func GetNodeAddresses(node *corev1.Node) []string {
	nodeAddresses := node.Status.Addresses
	addrs := []string{}

	// put internal ip first
	for _, addr := range nodeAddresses {
		if addr.Type == corev1.NodeInternalIP {
			addrs = append(addrs, addr.Address)
		}
	}

	// then external ip
	for _, addr := range nodeAddresses {
		if addr.Type == corev1.NodeExternalIP {
			addrs = append(addrs, addr.Address)
		}
	}

	// at last hostname
	for _, addr := range nodeAddresses {
		if addr.Type == corev1.NodeHostName {
			addrs = append(addrs, addr.Address)
		}
	}
	return addrs
}

func GetOperatorConfig(c client.Client) (*csiv1.Config, error) {
	configList := &csiv1.ConfigList{}
	err := c.List(context.TODO(), nil, configList)
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
