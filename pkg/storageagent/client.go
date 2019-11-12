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

package storageagent

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/IBM/ibm-block-csi-operator/pkg/config"
	pb "github.com/IBM/ibm-block-csi-operator/pkg/storageagent/storageagent"
	"github.com/IBM/ibm-block-csi-operator/pkg/util"
	"github.com/go-logr/logr"
	"google.golang.org/grpc"
)

var validStart = `^[a-zA-Z_]`
var validLetter = `[^a-zA-Z0-9 \-._]`
var replace = "_"
var prefix = replace

var timeout = time.Second * 10

type storageClient struct {
	arrayAddress, username, password string
	address                          string
	logger                           logr.Logger
}

func NewStorageClient(arrayAddress, username, password string, logger logr.Logger) StorageClient {
	return &storageClient{
		arrayAddress: arrayAddress,
		username:     username,
		password:     password,
		address:      getEndpoint(),
		logger:       logger,
	}
}

// beautify formats the host name to a valid one.
// rules:
//     1. The name can contain letters, numbers, spaces, periods, dashes, and underscores.
//     2. The name must begin with a letter or an underscore.
//     3. The name must not begin or end with a space.
func beautify(hostName string) string {
	trimed := strings.TrimSpace(hostName)
	startReg := regexp.MustCompile(validStart)
	nameReg := regexp.MustCompile(validLetter)

	if !startReg.MatchString(trimed) {
		trimed = prefix + trimed
	}
	return nameReg.ReplaceAllString(trimed, "_")
}

func (c *storageClient) CreateHost(name string, iscsiPorts, fcPorts []string) error {
	hostName := beautify(name)

	resInterface, err := c.runGrpcCommand(
		"CreateHost",
		&pb.CreateHostRequest{Name: hostName, Iqns: iscsiPorts, Wwpns: fcPorts,
			Secrets: map[string]string{"management_address": c.arrayAddress, "username": c.username, "password": c.password}},
	)
	if err != nil {
		return err
	}
	res := resInterface.(*pb.CreateHostReply)
	c.logger.Info("Created host", "name", res.GetHost().GetName())
	return nil
}

func (c *storageClient) ListIscsiTargets() ([]*pb.IscsiTarget, error) {
	resInterface, err := c.runGrpcCommand(
		"ListIscsiTargets",
		&pb.ListIscsiTargetsRequest{
			Secrets: map[string]string{"management_address": c.arrayAddress, "username": c.username, "password": c.password}},
	)
	if err != nil {
		return nil, err
	}
	res := resInterface.(*pb.ListIscsiTargetsReply)
	return res.GetTargets(), nil
}

func (c *storageClient) runGrpcCommand(cmdName string, request interface{}, opts ...grpc.CallOption) (interface{}, error) {
	c.logger.Info("Starting command", "command", cmdName)

	conn, err := grpc.Dial(c.address, grpc.WithInsecure())
	if err != nil {
		c.logger.Error(err, "Failed to connect server", "address", c.address)
		return nil, err
	}
	defer conn.Close()
	client := pb.NewStorageAgentClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	otherArgs := []interface{}{ctx, request}
	for _, opt := range opts {
		otherArgs = append(otherArgs, opt)
	}
	returnValues, err := util.Invoke(client, cmdName, otherArgs...)
	if err != nil {
		c.logger.Error(err, "Failed to invoke command", "command", cmdName)
		return nil, err
	}

	res := returnValues[0].Interface()
	errInterface := returnValues[1].Interface()
	if errInterface != nil {
		c.logger.Error(err, "Failed to execute command", "command", cmdName)
		return nil, errInterface.(error)
	}

	c.logger.Info("Successfully executed command", "command", cmdName)
	return res, nil
}

func getEndpoint() string {
	address := os.Getenv(config.ENVEndpoint)
	if address == "" {
		panic(fmt.Sprintf("ENV %s is not set!", config.ENVEndpoint))
	}
	return address
}
