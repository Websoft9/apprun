package services

import (
	"context"
	"encoding/json"

	temporalclient "go.temporal.io/sdk/client"
	"gofr.dev/pkg/gofr"
)

// EventService 事件服务（NATS + Temporal 集成）
type EventService struct {
	temporalClient temporalclient.Client
}

// NewEventService 创建事件服务
func NewEventService(temporalHost string) *EventService {
	// 初始化 Temporal 客户端
	client, err := temporalclient.Dial(temporalclient.Options{
		HostPort: temporalHost,
	})
	if err != nil {
		panic(err) // TODO: 改进错误处理
	}

	return &EventService{
		temporalClient: client,
	}
}

// SubscribeToEvents 订阅事件并触发相应的工作流
func (s *EventService) SubscribeToEvents(app *gofr.App) {
	// 订阅用户注册事件
	app.Subscribe("user.registered", func(ctx *gofr.Context) error {
		var event map[string]interface{}
		if err := json.Unmarshal(ctx.Request.Body, &event); err != nil {
			ctx.Logger.Error("Failed to unmarshal event", err)
			return err
		}

		userData := event["data"].(map[string]interface{})

		// 触发用户入职工作流
		workflowOptions := temporalclient.StartWorkflowOptions{
			ID:        "onboarding-" + userData["email"].(string),
			TaskQueue: "onboarding",
		}

		we, err := s.temporalClient.ExecuteWorkflow(
			context.Background(),
			workflowOptions,
			"OnboardingWorkflow",
			userData,
		)

		if err != nil {
			ctx.Logger.Error("Failed to start onboarding workflow", err)
			return err
		}

		ctx.Logger.Info("Onboarding workflow started", "workflow_id", we.GetID())
		return nil
	})

	// 订阅用户登录事件
	app.Subscribe("user.logged_in", func(ctx *gofr.Context) error {
		var event map[string]interface{}
		if err := json.Unmarshal(ctx.Request.Body, &event); err != nil {
			ctx.Logger.Error("Failed to unmarshal event", err)
			return err
		}

		// 记录登录统计
		ctx.Logger.Info("User logged in", "event", event)
		return nil
	})
}

// Close 关闭 Temporal 客户端
func (s *EventService) Close() {
	if s.temporalClient != nil {
		s.temporalClient.Close()
	}
}
