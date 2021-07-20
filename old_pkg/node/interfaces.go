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

package node

import pb "github.com/IBM/ibm-block-csi-operator/pkg/node/nodeagent"

type IscsiTarget struct {
	Portal, Port, Iqn string
}

type NodeClient interface {
	GetNodeInfo(name string) (*pb.Node, error)
	IscsiLogin(targets []string) error
	IscsiLogout(targets []string) error
}

type NodeInformer interface {
	GetNodeIscsiIQNs() ([]string, error)
	GetNodeFcWWPNs() ([]string, error)
}

type IscsiAdmin interface {
	DiscoverAndLoginPortals(portals []string) error
	DiscoverAndLogoutPortals(portals []string) error
	DiscoverAndLogin(portal string) error
	DiscoverAndLogout(portal string) error

	// Login performs an iscsi login for the specified target
	// portal is an address with port
	Login(tgtIQN, portal string) error

	// Logout performs an iscsi logout for the specified target
	// portal is an address with port
	Logout(tgtIQN, portal string) error

	// Discover performs an iscsi discoverydb for the specified target
	// portal is an address without port
	Discover(portal string) ([]*IscsiTarget, error)

	// DeleteDBEntry deletes the iscsi db entry fo the specified target
	DeleteDBEntry(tgtIQN string) error
}
