package storageagent

import (
	"context"
	"os"
	"time"

	pb "github.com/IBM/ibm-block-csi-driver-operator/pkg/storageagent/storageagent"
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

func NewStoragClient(arrayAddress, username, password string, logger logr.Logger) StoragClient {
	return &storageClient{
		arrayAddress: arrayAddress,
		username:     username,
		password:     password,
		logger:       logger,
	}
}

func (c *storageClient) CreateHost(name string, iscsiPorts, fcPorts []string) error {
	c.logger.Info("Creating host", "name", name, "ports", append(iscsiPorts, fcPorts...))

	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		c.logger.Error(err, "Failed to create host", "name", name)
		return err
	}
	defer conn.Close()
	client := pb.NewStorageAgentClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	c.logger.Info("Starting to Create host", "name", name)
	res, err := client.CreateHost(
		ctx,
		&pb.CreateHostRequest{Name: name, Iqns: iscsiPorts, Wwpns: fcPorts,
			Secrets: map[string]string{"management_address": c.arrayAddress, "username": c.username, "password": c.password}})
	if err != nil {
		c.logger.Error(err, "Failed to create host", "name", name)
		return err
	}
	c.logger.Info("Created host", "name", res.GetHost().GetName())
	return nil
}

func setEndpoint() {
	address = os.Getenv("ENDPOINT")
	if address == "" {
		panic("env is not set")
	}
}
