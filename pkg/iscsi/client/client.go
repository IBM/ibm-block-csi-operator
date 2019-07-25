package client

import (
	"context"
	"time"

	pb "github.com/IBM/ibm-block-csi-driver-operator/pkg/iscsi/iscsiagent"
	"github.com/IBM/ibm-block-csi-driver-operator/pkg/util"
	"github.com/go-logr/logr"
	"google.golang.org/grpc"
)

var timeout = time.Second * 10

type iscsiClient struct {
	address string
	logger  logr.Logger
}

func NewIscsiClient(address string, logger logr.Logger) IscsiClient {
	return &iscsiClient{
		address: address,
		logger:  logger.WithValues("address", address),
	}
}

func (c *iscsiClient) Login(targets []string) error {
	resInterface, err := c.runGrpcCommand("Login", &pb.LoginRequest{Targets: targets})
	if err != nil {
		return err
	}
	res := resInterface.(*pb.LoginReply)
	c.logger.Info("", "response", res)
	return nil
}

func (c *iscsiClient) Logout(targets []string) error {
	resInterface, err := c.runGrpcCommand("Logout", &pb.LogoutRequest{Targets: targets})
	if err != nil {
		return err
	}
	res := resInterface.(*pb.LogoutReply)
	c.logger.Info("", "response", res)
	return nil
}

func (c *iscsiClient) runGrpcCommand(cmdName string, request interface{}, opts ...grpc.CallOption) (interface{}, error) {
	c.logger.Info("Starting command", "command", cmdName)

	conn, err := grpc.Dial(c.address, grpc.WithInsecure())
	if err != nil {
		c.logger.Error(err, "Failed to connect server", "address", c.address)
		return nil, err
	}
	defer conn.Close()
	client := pb.NewIscsiAgentClient(conn)

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
