package workflow
package workflow

import (
	"gofr.dev/pkg/gofr"

	"github.com/Websoft9/apprun/core/internal/services"
)

// StartWorkflow 启动工作流处理器
func StartWorkflow(workflowService *services.WorkflowService) func(*gofr.Context) (interface{}, error) {
	return func(ctx *gofr.Context) (interface{}, error) {
		var req services.StartWorkflowRequest

		if err := ctx.Bind(&req); err != nil {
			return nil, &gofr.HTTP{
				StatusCode: 400,
				Message:    "invalid request body",
			}
		}

		tenantID := ctx.Get("tenant_id").(uint)
		userID := ctx.Get("user_id").(uint)

		workflow, err := workflowService.StartWorkflow(ctx, req, tenantID, userID)
		if err != nil {
			return nil, err
		}

		// 保存工作流记录到数据库
		if err := ctx.GORM().Create(workflow).Error; err != nil {
			return nil, err
		}

		return workflow, nil
	}
}

// GetWorkflow 获取工作流详情
func GetWorkflow(workflowService *services.WorkflowService) func(*gofr.Context) (interface{}, error) {
	return func(ctx *gofr.Context) (interface{}, error) {
		// TODO: 实现获取工作流详情
		return map[string]string{
			"message": "get workflow endpoint - to be implemented",
		}, nil
	}
}

// SignalWorkflow 发送信号到工作流
func SignalWorkflow(workflowService *services.WorkflowService) func(*gofr.Context) (interface{}, error) {
	return func(ctx *gofr.Context) (interface{}, error) {
		// TODO: 实现发送信号
		return map[string]string{
			"message": "signal workflow endpoint - to be implemented",
		}, nil
	}
}

// CancelWorkflow 取消工作流
func CancelWorkflow(workflowService *services.WorkflowService) func(*gofr.Context) (interface{}, error) {
	return func(ctx *gofr.Context) (interface{}, error) {
		// TODO: 实现取消工作流
		return map[string]string{
			"message": "cancel workflow endpoint - to be implemented",
		}, nil
	}
}
