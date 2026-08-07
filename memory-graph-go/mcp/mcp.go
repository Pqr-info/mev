package mcp

import (
	"context"
)

type MCPClient struct {
	// Add fields for Valkey connection etc.
}

func NewMCPClient() *MCPClient {
	return &MCPClient{}
}

func (m *MCPClient) GetGlobalState(ctx context.Context, key string) (map[string]interface{}, error) {
	return nil, nil
}

func (m *MCPClient) SetGlobalState(ctx context.Context, key string, value interface{}) error {
	return nil
}

func (m *MCPClient) PublishEvent(ctx context.Context, missionID string, event map[string]interface{}) error {
	return nil
}
