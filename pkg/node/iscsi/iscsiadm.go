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

package iscsi

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"github.com/IBM/ibm-block-csi-driver-operator/pkg/node"
	"github.com/pkg/errors"

	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var log = logf.Log.WithName("iscsi")
var execCommand = exec.Command

type iscsiAdmin struct{}

func NewIscsiAdmin() node.IscsiAdmin {
	return &iscsiAdmin{}
}

var iscsiCmd = func(args ...string) (string, error) {
	cmd := execCommand("iscsiadm", args...)
	var stdout bytes.Buffer
	var iscsiadmError error
	cmd.Stdout = &stdout
	cmd.Stderr = &stdout
	defer stdout.Reset()

	log.Info("Executing iscsiadm command", "command", "iscsiadm "+strings.Join(args, " "))

	// we're using Start and Wait because we want to grab exit codes
	err := cmd.Start()
	if err != nil {
		// This is usually a cmd not found so we'll set our own error here
		formattedOutput := strings.Replace(string(stdout.Bytes()), "\n", "", -1)
		iscsiadmError = fmt.Errorf("iscsiadm error: %s (%s)", formattedOutput, err.Error())
	} else {
		err = cmd.Wait()
		if err != nil {
			formattedOutput := strings.Replace(string(stdout.Bytes()), "\n", "", -1)
			iscsiadmError = fmt.Errorf("iscsiadm error: %s (%s)", formattedOutput, err.Error())
		}
	}

	return string(stdout.Bytes()), iscsiadmError
}

func (i *iscsiAdmin) DiscoverAndLoginPortals(portals []string) error {
	log.Info("Starting to login portals: " + strings.Join(portals, ", "))
	var err error
	var failedPortals = []string{}

	for _, portal := range portals {
		e := i.DiscoverAndLogin(portal)
		if e != nil {
			log.Error(e, "Failed to login portal "+portal)
			failedPortals = append(failedPortals, portal)
			if err == nil {
				err = e
			}
		}
	}
	log.Info("Finished to login portals")
	if err != nil {
		fp := strings.Join(failedPortals, ", ")
		return errors.WithMessage(err, "Failed to login portals "+fp)
	}
	return nil
}

func (i *iscsiAdmin) DiscoverAndLogoutPortals(portals []string) error {
	log.Info("Starting to logout portals: " + strings.Join(portals, ", "))
	var err error
	var failedPortals = []string{}

	for _, portal := range portals {
		e := i.DiscoverAndLogout(portal)
		if e != nil {
			log.Error(e, "Failed to logout portal "+portal)
			failedPortals = append(failedPortals, portal)
			if err == nil {
				err = e
			}
		}
	}
	log.Info("Finished to logout portals")
	if err != nil {
		fp := strings.Join(failedPortals, ", ")
		return errors.WithMessage(err, "Failed to logout portals "+fp)
	}
	return nil
}

func (i *iscsiAdmin) DiscoverAndLogin(portal string) error {
	targets, err := i.Discover(portal)
	if err != nil {
		return err
	}
	for _, target := range targets {
		targetErr := i.Login(target.Iqn, target.Portal+":"+target.Port)
		if targetErr != nil && err == nil {
			err = targetErr
		}
	}
	return err
}

func (i *iscsiAdmin) DiscoverAndLogout(portal string) error {
	targets, err := i.Discover(portal)
	if err != nil {
		return err
	}
	for _, target := range targets {
		targetErr := i.Logout(target.Iqn, target.Portal+":"+target.Port)
		if targetErr == nil {
			i.DeleteDBEntry(target.Iqn)
		}
		if targetErr != nil && err == nil {
			err = targetErr
		}
	}
	return err
}

// Login performs an iscsi login for the specified target
// portal is an address with port
func (i *iscsiAdmin) Login(tgtIQN, portal string) error {
	_, err := iscsiCmd([]string{"--mode", "node", "--targetname", tgtIQN, "--portal", portal, "--login"}...)
	return err
}

// Logout performs an iscsi logout for the specified target
// portal is an address with port
func (i *iscsiAdmin) Logout(tgtIQN, portal string) error {
	_, err := iscsiCmd([]string{"--mode", "node", "--targetname", tgtIQN, "--portal", portal, "--logout"}...)
	return err
}

// Discover performs an iscsi discoverydb for the specified target
// portal is an address without port
func (i *iscsiAdmin) Discover(portal string) ([]*node.IscsiTarget, error) {
	output, err := iscsiCmd([]string{"--mode", "discoverydb", "--type", "sendtargets", "--portal", portal, "--discover"}...)
	if err != nil {
		return nil, err
	}
	return extractIscsiTargets(output), nil
}

// DeleteDBEntry deletes the iscsi db entry fo the specified target
func (i *iscsiAdmin) DeleteDBEntry(tgtIQN string) error {
	_, err := iscsiCmd([]string{"--mode", "node", "--targetname", tgtIQN, "-o", "delete"}...)
	return err

}

// record format is 1.2.3.4:3260,1 iqn.xxxxxxxxx
func extractIscsiTargets(record string) []*node.IscsiTarget {
	targets := []*node.IscsiTarget{}
	records := strings.Split(record, "\n")
	for _, rec := range records {
		result := strings.Split(rec, " ")
		if len(result) != 2 {
			continue
		}
		target := &node.IscsiTarget{Iqn: result[1]}
		portalAndPort := strings.Split(strings.Split(result[0], ",")[0], ":")
		target.Portal = portalAndPort[0]
		target.Port = portalAndPort[1]
		targets = append(targets, target)
	}
	return targets
}
