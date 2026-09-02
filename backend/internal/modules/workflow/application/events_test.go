package application

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

func TestLifecycleEventBusDispatchesByBusinessType(t *testing.T) {
	bus := NewLifecycleEventBus(nil)
	var leaveEvents []LifecycleEvent
	var purchaseEvents []LifecycleEvent

	if err := bus.Register("leave_request", LifecycleEventHandlerFunc(func(_ context.Context, event LifecycleEvent) error {
		leaveEvents = append(leaveEvents, event)
		return nil
	})); err != nil {
		t.Fatalf("register leave handler: %v", err)
	}
	if err := bus.Register("purchase_order", LifecycleEventHandlerFunc(func(_ context.Context, event LifecycleEvent) error {
		purchaseEvents = append(purchaseEvents, event)
		return nil
	})); err != nil {
		t.Fatalf("register purchase handler: %v", err)
	}

	want := LifecycleEvent{
		Type: LifecycleInstanceCompleted, InstanceID: "instance-1",
		BusinessType: "leave_request", BusinessKey: "leave-100", Status: "completed",
	}
	bus.Publish(context.Background(), want)

	if !reflect.DeepEqual(leaveEvents, []LifecycleEvent{want}) {
		t.Fatalf("leave events = %#v, want %#v", leaveEvents, []LifecycleEvent{want})
	}
	if len(purchaseEvents) != 0 {
		t.Fatalf("purchase handler received unrelated events: %#v", purchaseEvents)
	}
}

func TestLifecycleEventBusContinuesAfterHandlerFailure(t *testing.T) {
	handlerErr := errors.New("writeback failed")
	var handled int
	var reported []error
	bus := NewLifecycleEventBus(func(_ LifecycleEvent, err error) {
		reported = append(reported, err)
	})

	if err := bus.Register("leave_request", LifecycleEventHandlerFunc(func(context.Context, LifecycleEvent) error {
		return handlerErr
	})); err != nil {
		t.Fatalf("register failing handler: %v", err)
	}
	if err := bus.Register("leave_request", LifecycleEventHandlerFunc(func(context.Context, LifecycleEvent) error {
		handled++
		return nil
	})); err != nil {
		t.Fatalf("register succeeding handler: %v", err)
	}

	bus.Publish(context.Background(), LifecycleEvent{Type: LifecycleInstanceRejected, BusinessType: "leave_request"})

	if handled != 1 {
		t.Fatalf("successful handler calls = %d, want 1", handled)
	}
	if len(reported) != 1 || !errors.Is(reported[0], handlerErr) {
		t.Fatalf("reported errors = %#v, want %v", reported, handlerErr)
	}
}

func TestLifecycleEventBusRejectsInvalidRegistration(t *testing.T) {
	bus := NewLifecycleEventBus(nil)
	if err := bus.Register("", LifecycleEventHandlerFunc(func(context.Context, LifecycleEvent) error { return nil })); err == nil {
		t.Fatal("Register() error = nil for empty business type")
	}
	if err := bus.Register("leave_request", nil); err == nil {
		t.Fatal("Register() error = nil for nil handler")
	}
}

func TestBusinessStatusLifecycleHandlerMapsInstanceEvents(t *testing.T) {
	var updates []BusinessStatusUpdate
	handler := NewBusinessStatusLifecycleHandler(BusinessStatusUpdaterFunc(func(_ context.Context, update BusinessStatusUpdate) error {
		updates = append(updates, update)
		return nil
	}))
	bus := NewLifecycleEventBus(nil)
	if err := bus.Register("leave_request", handler); err != nil {
		t.Fatalf("register business status handler: %v", err)
	}

	bus.Publish(context.Background(), LifecycleEvent{
		Type: LifecycleTaskCompleted, InstanceID: "instance-1", TaskID: "task-1",
		BusinessType: "leave_request", BusinessKey: "leave-100", Status: "running",
	})
	bus.Publish(context.Background(), LifecycleEvent{
		Type: LifecycleInstanceCompleted, InstanceID: "instance-1", ActorID: "user-7",
		BusinessType: "leave_request", BusinessKey: "leave-100", Status: "completed",
	})

	want := []BusinessStatusUpdate{{
		BusinessKey: "leave-100", InstanceID: "instance-1", ActorID: "user-7",
		Status: "completed", EventType: LifecycleInstanceCompleted,
	}}
	if !reflect.DeepEqual(updates, want) {
		t.Fatalf("business status updates = %#v, want %#v", updates, want)
	}
}

func TestBusinessStatusLifecycleHandlerRejectsNilUpdater(t *testing.T) {
	if handler := NewBusinessStatusLifecycleHandler(nil); handler != nil {
		t.Fatalf("handler = %#v, want nil", handler)
	}
}
