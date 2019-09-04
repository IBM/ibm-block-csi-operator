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

package informer

import (
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/IBM/ibm-block-csi-operator/pkg/node"
)

var iscsiPath = "/etc/iscsi"
var iscsiFile = "initiatorname.iscsi"
var iscsiFullPath = path.Join(iscsiPath, iscsiFile)
var fcPath = "/sys/class/fc_host"

const portName = "port_name"
const portState = "port_state"
const portOnline = "Online"

type informer struct{}

func NewInformer() node.NodeInformer {
	return &informer{}
}

func (i *informer) GetNodeIscsiIQNs() ([]string, error) {
	if ok, err := exists(iscsiFullPath); !ok || err != nil {
		return nil, err
	}
	iqnLine, err := ioutil.ReadFile(iscsiFullPath)
	if err != nil {
		return nil, err
	}
	iqnLineStr := strings.TrimSpace(string(iqnLine))
	return []string{strings.Split(iqnLineStr, "=")[1]}, nil
}

func (i *informer) GetNodeFcWWPNs() ([]string, error) {
	if ok, err := exists(fcPath); !ok || err != nil {
		return nil, err
	}

	hostDirs, err := ioutil.ReadDir(fcPath)
	if err != nil {
		return nil, err
	}

	wwpns := []string{}
	for _, hostDir := range hostDirs {
		if !hostDir.IsDir() {
			continue
		}
		hostName := hostDir.Name()
		hostPath := filepath.Join(fcPath, hostName)
		hostPortName := filepath.Join(hostPath, portName)
		hostPortState := filepath.Join(hostPath, portState)
		if ok, err := exists(hostPortName); !ok || err != nil {
			continue
		}
		if ok, err := exists(hostPortState); !ok || err != nil {
			continue
		}
		state, err := ioutil.ReadFile(hostPortState)
		if err != nil {
			continue
		}
		if strings.TrimSpace(string(state)) == portOnline {
			name, err := ioutil.ReadFile(hostPortName)
			if err != nil {
				continue
			}
			wwpns = append(wwpns, strings.TrimSpace(string(name)))
		}
	}
	return wwpns, nil
}

func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}
