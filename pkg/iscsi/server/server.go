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

package server

import (
	"context"
	"net"

	"github.com/IBM/ibm-block-csi-driver-operator/pkg/iscsi"
	pb "github.com/IBM/ibm-block-csi-driver-operator/pkg/iscsi/iscsiagent"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

// server is used to implement iscsiagent.IscsiAgentServer.
type server struct{}

func (s *server) Login(ctx context.Context, in *pb.LoginRequest) (*pb.LoginReply, error) {
	err := iscsi.DiscoverAndLoginPortals(in.GetTargets())
	if err != nil {
		return nil, status.Convert(err).Err()
	}
	return &pb.LoginReply{}, nil
}

func (s *server) Logout(ctx context.Context, in *pb.LogoutRequest) (*pb.LogoutReply, error) {
	err := iscsi.DiscoverAndLogoutPortals(in.GetTargets())
	if err != nil {
		return nil, status.Convert(err).Err()
	}
	return &pb.LogoutReply{}, nil
}

func Serve(address string) error {
	conn, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}
	s := grpc.NewServer()
	pb.RegisterIscsiAgentServer(s, &server{})
	if err := s.Serve(conn); err != nil {
		return err
	}
	return nil
}
