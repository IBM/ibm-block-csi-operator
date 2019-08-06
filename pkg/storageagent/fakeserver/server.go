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

package fakeserver

import (
	"context"
	"net"

	pb "github.com/IBM/ibm-block-csi-driver-operator/pkg/storageagent/storageagent"
	"google.golang.org/grpc"
)

// fakeServer is used to mock a nodeagent.NodeAgentServer.
type fakeServer struct {
	nodeName string
}

func (s *fakeServer) CreateHost(ctx context.Context, in *pb.CreateHostRequest) (*pb.CreateHostReply, error) {
	return &pb.CreateHostReply{Host: &pb.Host{Name: "fake-name"}}, nil
}

func (s *fakeServer) DeleteHost(ctx context.Context, in *pb.DeleteHostRequest) (*pb.DeleteHostReply, error) {
	return &pb.DeleteHostReply{}, nil
}

func (s *fakeServer) ListHosts(ctx context.Context, in *pb.ListHostsRequest) (*pb.ListHostsReply, error) {
	return &pb.ListHostsReply{}, nil
}

func (s *fakeServer) ListIscsiTargets(ctx context.Context, in *pb.ListIscsiTargetsRequest) (*pb.ListIscsiTargetsReply, error) {
	return &pb.ListIscsiTargetsReply{Targets: []*pb.IscsiTarget{{Address: "1.2.3.4"}}}, nil
}

var s *grpc.Server

func Serve(address string) error {
	conn, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}
	s = grpc.NewServer()
	pb.RegisterStorageAgentServer(s, &fakeServer{})
	if err := s.Serve(conn); err != nil {
		return err
	}
	return nil
}

func Stop() {
	s.Stop()
}
