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

package client

import (
	"context"
	"time"

	pb "github.com/IBM/ibm-block-csi-driver-operator/pkg/node/nodeagent"
	"github.com/IBM/ibm-block-csi-driver-operator/pkg/util"
	"github.com/go-logr/logr"
	"google.golang.org/grpc"
)

var timeout = time.Second * 10

type nodeClient struct {
	address string
	logger  logr.Logger
}

func NewNodeClient(address string, logger logr.Logger) NodeClient {
	return &nodeClient{
		address: address,
		logger:  logger.WithValues("address", address),
	}
}

func (c *nodeClient) GetNodeInfo(name string) (*pb.Node, error) {
	resInterface, err := c.runGrpcCommand("GetNodeInfo", &pb.GetNodeInfoRequest{Name: name})
	if err != nil {
		return nil, err
	}
	res := resInterface.(*pb.GetNodeInfoReply)
	c.logger.Info("", "response", res)
	return res.GetNode(), nil
}

func (c *nodeClient) IscsiLogin(targets []string) error {
	resInterface, err := c.runGrpcCommand("IscsiLogin", &pb.IscsiLoginRequest{Targets: targets})
	if err != nil {
		return err
	}
	res := resInterface.(*pb.IscsiLoginReply)
	c.logger.Info("", "response", res)
	return nil
}

func (c *nodeClient) IscsiLogout(targets []string) error {
	resInterface, err := c.runGrpcCommand("IscsiLogout", &pb.IscsiLogoutRequest{Targets: targets})
	if err != nil {
		return err
	}
	res := resInterface.(*pb.IscsiLogoutReply)
	c.logger.Info("", "response", res)
	return nil
}

func (c *nodeClient) runGrpcCommand(cmdName string, request interface{}, opts ...grpc.CallOption) (interface{}, error) {
	c.logger.Info("Starting command", "command", cmdName)

	conn, err := grpc.Dial(c.address, grpc.WithInsecure())
	if err != nil {
		c.logger.Error(err, "Failed to connect server", "address", c.address)
		return nil, err
	}
	defer conn.Close()
	client := pb.NewNodeAgentClient(conn)

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
