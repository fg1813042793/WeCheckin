package application

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"
)

type LifecycleEventType string

const (
	LifecycleInstanceStarted   LifecycleEventType = "instance_started"
	LifecycleTaskCompleted     LifecycleEventType = "task_completed"
	LifecycleInstanceCompleted LifecycleEventType = "instance_completed"
	LifecycleInstanceRejected  LifecycleEventType = "instance_rejected"
	LifecycleInstanceWithdrawn LifecycleEventType = "instance_withdrawn"
	LifecycleInstanceCancelled LifecycleEventType = "instance_cancelled"
)

type LifecycleEvent struct {
	Type         LifecycleEventType `json:"type"`
	InstanceID   string             `json:"instanceId"`
	TaskID       string             `json:"taskId,omitempty"`
	ActorID      string             `json:"actorId,omitempty"`
	BusinessType string             `json:"businessType,omitempty"`
	BusinessKey  string             `json:"businessKey,omitempty"`
	Status       string             `json:"status"`
}

type EventPublisher interface {
	Publish(context.Context, LifecycleEvent)
}

type LifecycleEventHandler interface {
	HandleLifecycleEvent(context.Context, LifecycleEvent) error
}

type LifecycleEventHandlerFunc func(context.Context, LifecycleEvent) error

func (handler LifecycleEventHandlerFunc) HandleLifecycleEvent(ctx context.Context, event LifecycleEvent) error {
	return handler(ctx, event)
}

type LifecycleEventErrorHandler func(LifecycleEvent, error)

type BusinessStatusUpdate struct {
	BusinessKey string             `json:"businessKey"`
	InstanceID  string             `json:"instanceId"`
	ActorID     string             `json:"actorId,omitempty"`
	Status      string             `json:"status"`
	EventType   LifecycleEventType `json:"eventType"`
}

type BusinessStatusUpdater interface {
	UpdateWorkflowStatus(context.Context, BusinessStatusUpdate) error
}

type BusinessStatusUpdaterFunc func(context.Context, BusinessStatusUpdate) error

func (updater BusinessStatusUpdaterFunc) UpdateWorkflowStatus(ctx context.Context, update BusinessStatusUpdate) error {
	return updater(ctx, update)
}

type businessStatusLifecycleHandler struct {
	updater BusinessStatusUpdater
}

func NewBusinessStatusLifecycleHandler(updater BusinessStatusUpdater) LifecycleEventHandler {
	if updater == nil {
		return nil
	}
	return &businessStatusLifecycleHandler{updater: updater}
}

func (handler *businessStatusLifecycleHandler) HandleLifecycleEvent(ctx context.Context, event LifecycleEvent) error {
	if handler == nil || handler.updater == nil || !isBusinessStatusEvent(event.Type) {
		return nil
	}
	return handler.updater.UpdateWorkflowStatus(ctx, BusinessStatusUpdate{
		BusinessKey: event.BusinessKey,
		InstanceID:  event.InstanceID,
		ActorID:     event.ActorID,
		Status:      event.Status,
		EventType:   event.Type,
	})
}

func isBusinessStatusEvent(eventType LifecycleEventType) bool {
	switch eventType {
	case LifecycleInstanceStarted, LifecycleInstanceCompleted, LifecycleInstanceRejected,
		LifecycleInstanceWithdrawn, LifecycleInstanceCancelled:
		return true
	default:
		return false
	}
}

type LifecycleEventBus struct {
	mu       sync.RWMutex
	handlers map[string][]LifecycleEventHandler
	onError  LifecycleEventErrorHandler
}

func NewLifecycleEventBus(onError LifecycleEventErrorHandler) *LifecycleEventBus {
	return &LifecycleEventBus{
		handlers: make(map[string][]LifecycleEventHandler),
		onError:  onError,
	}
}

func (bus *LifecycleEventBus) Register(businessType string, handler LifecycleEventHandler) error {
	if bus == nil {
		return errors.New("流程生命周期事件总线未初始化")
	}
	businessType = strings.TrimSpace(businessType)
	if businessType == "" {
		return errors.New("流程业务类型不能为空")
	}
	if handler == nil {
		return errors.New("流程业务回写处理器不能为空")
	}
	bus.mu.Lock()
	defer bus.mu.Unlock()
	bus.handlers[businessType] = append(bus.handlers[businessType], handler)
	return nil
}

func (bus *LifecycleEventBus) Publish(ctx context.Context, event LifecycleEvent) {
	if bus == nil {
		return
	}
	businessType := strings.TrimSpace(event.BusinessType)
	if businessType == "" {
		return
	}
	bus.mu.RLock()
	handlers := append([]LifecycleEventHandler(nil), bus.handlers[businessType]...)
	bus.mu.RUnlock()
	for _, handler := range handlers {
		if err := invokeLifecycleEventHandler(ctx, handler, event); err != nil && bus.onError != nil {
			bus.onError(event, err)
		}
	}
}

func invokeLifecycleEventHandler(ctx context.Context, handler LifecycleEventHandler, event LifecycleEvent) (err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			err = fmt.Errorf("流程业务回写处理器异常: %v", recovered)
		}
	}()
	return handler.HandleLifecycleEvent(ctx, event)
}

var defaultLifecycleEventBus = NewLifecycleEventBus(func(event LifecycleEvent, err error) {
	log.Printf("[WorkflowLifecycle] businessType=%s businessKey=%s instanceId=%s event=%s writeback error: %v",
		event.BusinessType, event.BusinessKey, event.InstanceID, event.Type, err)
})

func DefaultLifecycleEventPublisher() EventPublisher {
	return defaultLifecycleEventBus
}

func RegisterLifecycleEventHandler(businessType string, handler LifecycleEventHandler) error {
	return defaultLifecycleEventBus.Register(businessType, handler)
}

func RegisterBusinessStatusUpdater(businessType string, updater BusinessStatusUpdater) error {
	handler := NewBusinessStatusLifecycleHandler(updater)
	if handler == nil {
		return errors.New("流程业务状态回写器不能为空")
	}
	return RegisterLifecycleEventHandler(businessType, handler)
}

type noopEventPublisher struct{}

func (noopEventPublisher) Publish(context.Context, LifecycleEvent) {}
