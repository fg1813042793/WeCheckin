package application

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"

	workflowdomain "wecheckin/backend/internal/modules/workflow/domain"
)

type LifecycleEventType string

const (
	LifecycleInstanceStarted     LifecycleEventType = "instance_started"
	LifecycleTaskCompleted       LifecycleEventType = "task_completed"
	LifecycleInstanceCompleted   LifecycleEventType = "instance_completed"
	LifecycleInstanceRejected    LifecycleEventType = "instance_rejected"
	LifecycleInstanceWithdrawn   LifecycleEventType = "instance_withdrawn"
	LifecycleInstanceCancelled   LifecycleEventType = "instance_cancelled"
	LifecycleInstanceFormRevised LifecycleEventType = "workflow.instance.form_revised"
)

type LifecycleEvent struct {
	Type              LifecycleEventType     `json:"type"`
	InstanceID        string                 `json:"instanceId"`
	TaskID            string                 `json:"taskId,omitempty"`
	ActorID           string                 `json:"actorId,omitempty"`
	BusinessType      string                 `json:"businessType,omitempty"`
	BusinessKey       string                 `json:"businessKey,omitempty"`
	Status            string                 `json:"status"`
	FormRevision      int64                  `json:"formRevision,omitempty"`
	RevisionRequestID string                 `json:"revisionRequestId,omitempty"`
	FormPatch         map[string]interface{} `json:"formPatch,omitempty"`
}

func NewFormRevisedBusinessEvent(
	instance workflowdomain.ProcessInstance,
	revisionRequestID string,
	formRevision int64,
	formPatch map[string]interface{},
) LifecycleEvent {
	return LifecycleEvent{
		Type: LifecycleInstanceFormRevised, InstanceID: instance.ID,
		BusinessType: instance.BusinessType, BusinessKey: instance.BusinessKey,
		Status: string(instance.Status), FormRevision: formRevision,
		RevisionRequestID: strings.TrimSpace(revisionRequestID),
		FormPatch:         cloneEventFormPatch(formPatch),
	}
}

func cloneEventFormPatch(patch map[string]interface{}) map[string]interface{} {
	if patch == nil {
		return nil
	}
	result := make(map[string]interface{}, len(patch))
	for key, value := range patch {
		result[key] = value
	}
	return result
}

type EventPublisher interface {
	Publish(context.Context, LifecycleEvent)
}

type LifecycleEventDispatcher interface {
	Dispatch(context.Context, LifecycleEvent) error
}

type WorkflowBusinessEvent struct {
	ID        string
	DedupeKey string
	Event     LifecycleEvent
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
	_ = bus.Dispatch(ctx, event)
}

func (bus *LifecycleEventBus) Dispatch(ctx context.Context, event LifecycleEvent) error {
	if bus == nil {
		return nil
	}
	businessType := strings.TrimSpace(event.BusinessType)
	if businessType == "" {
		return nil
	}
	bus.mu.RLock()
	handlers := append([]LifecycleEventHandler(nil), bus.handlers[businessType]...)
	bus.mu.RUnlock()
	var firstErr error
	for _, handler := range handlers {
		if err := invokeLifecycleEventHandler(ctx, handler, event); err != nil {
			if firstErr == nil {
				firstErr = err
			}
			if bus.onError != nil {
				bus.onError(event, err)
			}
		}
	}
	return firstErr
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

func DefaultLifecycleEventBus() *LifecycleEventBus {
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
