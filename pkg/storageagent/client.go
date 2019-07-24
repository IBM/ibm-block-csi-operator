package storageagent

import (
	"context"
	"os"
	"time"

	pb "github.com/IBM/ibm-block-csi-driver-operator/pkg/storageagent/storageagent"
	"github.com/IBM/ibm-block-csi-driver-operator/pkg/util"
	"github.com/go-logr/logr"
	"google.golang.org/grpc"
)

var address string

func init() {
	setEndpoint()
}

var timeout = time.Second * 10

type storageClient struct {
	arrayAddress, username, password string
	logger                           logr.Logger
}

func NewStorageClient(arrayAddress, username, password string, logger logr.Logger) StorageClient {
	return &storageClient{
		arrayAddress: arrayAddress,
		username:     username,
		password:     password,
		logger:       logger,
	}
}

func (c *storageClient) CreateHost(name string, iscsiPorts, fcPorts []string) error {
	resInterface, err := c.runGrpcCommand(
		"CreateHost",
		&pb.CreateHostRequest{Name: name, Iqns: iscsiPorts, Wwpns: fcPorts,
			Secrets: map[string]string{"management_address": c.arrayAddress, "username": c.username, "password": c.password}},
	)
	if err != nil {
		return err
	}
	res := resInterface.(*pb.CreateHostReply)
	c.logger.Info("Created host", "name", res.GetHost().GetName())
	return nil
}

func (c *storageClient) runGrpcCommand(cmdName string, request interface{}, opts ...grpc.CallOption) (interface{}, error) {
	c.logger.Info("Starting command", "command", cmdName)

	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		c.logger.Error(err, "Failed to connect server", "address", address)
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

func setEndpoint() {
	address = os.Getenv("ENDPOINT")
	if address == "" {
		panic("env is not set")
	}
}
