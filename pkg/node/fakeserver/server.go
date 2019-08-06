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

	pb "github.com/IBM/ibm-block-csi-driver-operator/pkg/node/nodeagent"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// fakeServer is used to mock a nodeagent.NodeAgentServer.
type fakeServer struct {
	nodeName string
}

func (s *fakeServer) GetNodeInfo(ctx context.Context, in *pb.GetNodeInfoRequest) (*pb.GetNodeInfoReply, error) {
	if in.GetName() != s.nodeName {
		return nil, status.Error(codes.InvalidArgument, "node name mismatch")
	}

	res := LoadResponse("GetNodeInfo", in)
	if res == nil {
		return nil, status.Error(codes.InvalidArgument, "no responses available, you should run StoreResponse first")
	}
	return res.(*pb.GetNodeInfoReply), nil
}

func (s *fakeServer) IscsiLogin(ctx context.Context, in *pb.IscsiLoginRequest) (*pb.IscsiLoginReply, error) {
	return &pb.IscsiLoginReply{}, nil
}

func (s *fakeServer) IscsiLogout(ctx context.Context, in *pb.IscsiLogoutRequest) (*pb.IscsiLogoutReply, error) {
	return &pb.IscsiLogoutReply{}, nil
}

var s *grpc.Server

func Serve(address, nodeName string) error {
	conn, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}
	s = grpc.NewServer()
	pb.RegisterNodeAgentServer(s, &fakeServer{nodeName: nodeName})
	if err := s.Serve(conn); err != nil {
		return err
	}
	return nil
}

func Stop() {
	s.Stop()
}
