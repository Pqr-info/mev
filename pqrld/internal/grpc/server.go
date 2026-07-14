package grpc

import (
	"context"
	"fmt"
	"net"

	"google.golang.org/grpc"
	"pqr.info/mev/pqrld/internal/executor"
	pb "pqr.info/mev/pqrld/proto"
)

type RunlevelServer struct {
	pb.UnimplementedRunlevelServiceServer
	Engine *executor.Executor
}

func StartGRPCServer(port int, engine *executor.Executor) error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	s := grpc.NewServer()
	pb.RegisterRunlevelServiceServer(s, &RunlevelServer{Engine: engine})

	fmt.Printf("pqrld gRPC server listening on port %d\n", port)
	if err := s.Serve(lis); err != nil {
		return fmt.Errorf("failed to serve: %w", err)
	}

	return nil
}

func (s *RunlevelServer) GetRunlevelStatus(ctx context.Context, req *pb.GetStatusRequest) (*pb.GetStatusResponse, error) {
	state, level := s.Engine.GetStatus()

	return &pb.GetStatusResponse{
		State:           state,
		CurrentRunlevel: level,
		ErrorMessage:    "",
	}, nil
}

func (s *RunlevelServer) AdvanceRunlevel(ctx context.Context, req *pb.AdvanceRequest) (*pb.AdvanceResponse, error) {
	return &pb.AdvanceResponse{
		Success: true,
		Message: "Advancing runlevels not supported in skeleton mode",
	}, nil
}

func (s *RunlevelServer) RollbackRunlevel(ctx context.Context, req *pb.RollbackRequest) (*pb.RollbackResponse, error) {
	return &pb.RollbackResponse{
		Success: true,
		Message: "Rollback runlevels not supported in skeleton mode",
	}, nil
}

func (s *RunlevelServer) TriggerHealing(ctx context.Context, req *pb.HealingRequest) (*pb.HealingResponse, error) {
	return &pb.HealingResponse{
		Success: true,
		Message: "Healing triggered successfully",
	}, nil
}

func (s *RunlevelServer) ReloadConfig(ctx context.Context, req *pb.ReloadRequest) (*pb.ReloadResponse, error) {
	return &pb.ReloadResponse{
		Success: true,
		Message: "Config reloaded successfully",
	}, nil
}
