package proto

import (
	"context"

	"google.golang.org/grpc"
)

type RunlevelServiceServer interface {
	GetRunlevelStatus(context.Context, *GetStatusRequest) (*GetStatusResponse, error)
	AdvanceRunlevel(context.Context, *AdvanceRequest) (*AdvanceResponse, error)
	RollbackRunlevel(context.Context, *RollbackRequest) (*RollbackResponse, error)
	TriggerHealing(context.Context, *HealingRequest) (*HealingResponse, error)
	ReloadConfig(context.Context, *ReloadRequest) (*ReloadResponse, error)
	mustEmbedUnimplementedRunlevelServiceServer()
}

type UnimplementedRunlevelServiceServer struct{}

func (UnimplementedRunlevelServiceServer) GetRunlevelStatus(context.Context, *GetStatusRequest) (*GetStatusResponse, error) {
	return nil, nil
}
func (UnimplementedRunlevelServiceServer) AdvanceRunlevel(context.Context, *AdvanceRequest) (*AdvanceResponse, error) {
	return nil, nil
}
func (UnimplementedRunlevelServiceServer) RollbackRunlevel(context.Context, *RollbackRequest) (*RollbackResponse, error) {
	return nil, nil
}
func (UnimplementedRunlevelServiceServer) TriggerHealing(context.Context, *HealingRequest) (*HealingResponse, error) {
	return nil, nil
}
func (UnimplementedRunlevelServiceServer) ReloadConfig(context.Context, *ReloadRequest) (*ReloadResponse, error) {
	return nil, nil
}
func (UnimplementedRunlevelServiceServer) mustEmbedUnimplementedRunlevelServiceServer() {}

func RegisterRunlevelServiceServer(s *grpc.Server, srv RunlevelServiceServer) {
	// gRPC register stub
}

type GetStatusRequest struct{}

type GetStatusResponse struct {
	State           string `json:"state"`
	CurrentRunlevel string `json:"current_runlevel"`
	ErrorMessage    string `json:"error_message"`
}

type AdvanceRequest struct {
	Runlevel string `json:"runlevel"`
}

type AdvanceResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type RollbackRequest struct {
	Runlevel string `json:"runlevel"`
}

type RollbackResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type HealingRequest struct {
	Runlevel string `json:"runlevel"`
}

type HealingResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type ReloadRequest struct{}

type ReloadResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}
