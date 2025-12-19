package services

import (
	"context"

	temporalclient "go.temporal.io/sdk/client"

	"github.com/Websoft9/apprun/core/internal/models"
)

// WorkflowService 工作流服务
type WorkflowService struct {
	temporalClient temporalclient.Client
}

// NewWorkflowService 创建工作流服务
func NewWorkflowService(temporalHost string) *WorkflowService {
	client, err := temporalclient.Dial(temporalclient.Options{
		HostPort: temporalHost,
	})
	if err != nil {
		panic(err) // TODO: 改进错误处理
	}

	return &WorkflowService{
		temporalClient: client,
	}
}

// StartWorkflowRequest 启动工作流请求
type StartWorkflowRequest struct {
	Name         string                 `json:"name" validate:"required"`
	Description  string                 `json:"description"`
	WorkflowType string                 `json:"workflow_type" validate:"required"`
	Input        map[string]interface{} `json:"input"`
}

// StartWorkflow 启动工作流
func (s *WorkflowService) StartWorkflow(ctx context.Context, req StartWorkflowRequest, tenantID, userID uint) (*models.Workflow, error) {
	// 启动 Temporal 工作流
	workflowOptions := temporalclient.StartWorkflowOptions{
		ID:        req.Name,
		TaskQueue: req.WorkflowType,
	}

	we, err := s.temporalClient.ExecuteWorkflow(
		ctx,
		workflowOptions,
		req.WorkflowType,
		req.Input,
	)

	if err != nil {
		return nil, err
	}

	// 创建工作流记录
	workflow := &models.Workflow{
		TenantID:    tenantID,
		Name:        req.Name,
		Description: req.Description,
		WorkflowID:  we.GetID(),
		RunID:       we.GetRunID(),
		Status:      "running",
		CreatedBy:   userID,
	}

	return workflow, nil
}

// Close 关闭客户端
func (s *WorkflowService) Close() {
	if s.temporalClient != nil {
		s.temporalClient.Close()
	}
}
