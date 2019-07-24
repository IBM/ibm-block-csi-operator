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
