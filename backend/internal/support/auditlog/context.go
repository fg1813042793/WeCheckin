package auditlog

import (
	"context"
	"sync"
)

type Change struct {
	Field  string      `json:"field"`
	Before interface{} `json:"before"`
	After  interface{} `json:"after"`
}

type Details struct {
	EventID        string
	ActorID        uint
	ActorName      string
	ActorType      string
	ReasonCode     string
	Reason         string
	TargetType     string
	TargetID       string
	TargetIDs      []string
	TargetName     string
	AffectedCount  int
	RelatedType    string
	RelatedID      string
	Outcome        string
	SnapshotStatus string
	SnapshotReason string
	Changes        []Change
}

type Recorder struct {
	mu      sync.Mutex
	details Details
}

type contextKey struct{}

func WithRecorder(ctx context.Context) (context.Context, *Recorder) {
	r := &Recorder{}
	return context.WithValue(ctx, contextKey{}, r), r
}

// Update is a no-op outside an audited request, including existing background jobs.
func Update(ctx context.Context, update func(*Details)) {
	if ctx == nil {
		return
	}
	r, ok := ctx.Value(contextKey{}).(*Recorder)
	if !ok {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	update(&r.details)
}

func (r *Recorder) Snapshot() Details {
	r.mu.Lock()
	defer r.mu.Unlock()
	d := r.details
	d.Changes = append([]Change(nil), d.Changes...)
	d.TargetIDs = append([]string(nil), d.TargetIDs...)
	return d
}

func SetActor(ctx context.Context, id uint, name, actorType string) {
	Update(ctx, func(d *Details) { d.ActorID, d.ActorName, d.ActorType = id, name, actorType })
}

func SetReason(ctx context.Context, code, reason string) {
	Update(ctx, func(d *Details) { d.ReasonCode, d.Reason = code, reason })
}

func SetChanges(ctx context.Context, changes []Change) {
	Update(ctx, func(d *Details) { d.Changes = changes })
}

func CurrentEventID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	r, ok := ctx.Value(contextKey{}).(*Recorder)
	if !ok {
		return ""
	}
	return r.Snapshot().EventID
}
