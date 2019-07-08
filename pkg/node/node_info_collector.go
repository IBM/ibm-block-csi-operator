package node

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

const iqnPath = "/etc/iscsi/initiatorname.iscsi"
const fcPath = "/sys/class/fc_host"
const portName = "port_name"
const portState = "port_state"
const portOnline = "Online"

func GetNodeIscsiIQN() (string, error) {
	if ok, err := exists(iqnPath); !ok || err != nil {
		return "", err
	}
	iqnLine, err := ioutil.ReadFile(iqnPath)
	if err != nil {
		return "", err
	}
	iqnLineStr := strings.TrimSpace(string(iqnLine))
	return strings.Split(iqnLineStr, "=")[1], nil
}

func GetNodeFcWWPNs() ([]string, error) {
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
