package iscsi

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var log = logf.Log.WithName("iscsi")
var execCommand = exec.Command

type iscsiTarget struct {
	portal, port, iqn string
}

func iscsiCmd(args ...string) (string, error) {
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

	iscsiadmDebug(string(stdout.Bytes()), iscsiadmError)
	return string(stdout.Bytes()), iscsiadmError
}

func iscsiadmDebug(output string, cmdError error) {
	debugOutput := strings.Replace(output, "\n", "\\n", -1)
	log.Info("Output of iscsiadm command", "output", debugOutput)
	if cmdError != nil {
		log.Info("Error returned from iscsiadm command", "error", cmdError.Error())
	}
}

func DiscoverAndLogin(portal string) error {
	targets, err := Discover(portal)
	if err != nil {
		return err
	}
	for _, target := range targets {
		targetErr := Login(target.iqn, target.portal+":"+target.port)
		if targetErr != nil && err == nil {
			err = targetErr
		}
	}
	return err
}

func DiscoverAndLogout(portal string) error {
	targets, err := Discover(portal)
	if err != nil {
		return err
	}
	for _, target := range targets {
		targetErr := Logout(target.iqn, target.portal+":"+target.port)
		if targetErr == nil {
			DeleteDBEntry(target.iqn)
		}
		if targetErr != nil && err == nil {
			err = targetErr
		}
	}
	return err
}

// Login performs an iscsi login for the specified target
// portal is an address with port
func Login(tgtIQN, portal string) error {
	_, err := iscsiCmd([]string{"--mode", "node", "--targetname", tgtIQN, "--portal", portal, "--login"}...)
	return err
}

// Logout performs an iscsi logout for the specified target
// portal is an address with port
func Logout(tgtIQN, portal string) error {
	_, err := iscsiCmd([]string{"--mode", "node", "--targetname", tgtIQN, "--portal", portal, "--logout"}...)
	return err
}

// Discover performs an iscsi discoverydb for the specified target
// portal is an address without port
func Discover(portal string) ([]*iscsiTarget, error) {
	output, err := iscsiCmd([]string{"--mode", "discoverydb", "--type", "sendtargets", "--portal", portal, "--discover"}...)
	if err != nil {
		return nil, err
	}
	return extractIscsiTargets(output), nil
}

// DeleteDBEntry deletes the iscsi db entry fo the specified target
func DeleteDBEntry(tgtIQN string) error {
	_, err := iscsiCmd([]string{"--mode", "node", "--targetname", tgtIQN, "-o", "delete"}...)
	return err

}

// record format is 1.2.3.4:3260,1 iqn.xxxxxxxxx
func extractIscsiTargets(record string) []*iscsiTarget {
	targets := []*iscsiTarget{}
	records := strings.Split(record, "\n")
	for _, rec := range records {
		result := strings.Split(rec, " ")
		if len(result) != 2 {
			continue
		}
		target := &iscsiTarget{iqn: result[1]}
		portalAndPort := strings.Split(strings.Split(result[0], ",")[0], ":")
		target.portal = portalAndPort[0]
		target.port = portalAndPort[1]
		targets = append(targets, target)
	}
	return targets
}
