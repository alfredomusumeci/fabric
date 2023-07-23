package server

import (
	"context"
	"net"

	"google.golang.org/grpc"

	pb "github.com/hyperledger/fabric-protos-go/blocc"
)

type CLIServer struct {
	pb.UnimplementedCliServiceServer
}

func (s *CLIServer) ExecuteFunction(ctx context.Context, req *pb.ExecuteFunctionRequest) (*pb.ExecuteFunctionResponse, error) {
	return &pb.ExecuteFunctionResponse{
		Result: "OK",
		Error:  "",
	}, nil
}

func ConnectToCLI() {
	logger.Debug("Attempting to connect to CLI")
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		logger.Debugf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterCliServiceServer(s, &CLIServer{})
	if err := s.Serve(lis); err != nil {
		logger.Debugf("failed to serve: %v", err)
	}
}
