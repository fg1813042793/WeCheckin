package application

import "context"

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

type noopEventPublisher struct{}

func (noopEventPublisher) Publish(context.Context, LifecycleEvent) {}
