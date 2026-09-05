package domain

import (
	"errors"
	"strings"
)

type Status string

const (
	StatusPending    Status = "pending"
	StatusProcessing Status = "processing"
	StatusResolved   Status = "resolved"
	StatusClosed     Status = "closed"
)

var (
	ErrTransitionNotAllowed   = errors.New("feedback status transition is not allowed")
	ErrTransitionNoteRequired = errors.New("feedback status transition note is required")
)

func (s Status) Valid() bool {
	switch s {
	case StatusPending, StatusProcessing, StatusResolved, StatusClosed:
		return true
	default:
		return false
	}
}

func (s Status) AllowsSupplement() bool {
	return s == StatusPending || s == StatusProcessing
}

func ValidateTransition(from, to Status, note string) error {
	requiresNote := false
	switch {
	case from == StatusPending && to == StatusProcessing:
	case from == StatusPending && to == StatusClosed:
		requiresNote = true
	case from == StatusProcessing && to == StatusResolved:
		requiresNote = true
	case from == StatusResolved && to == StatusClosed:
		requiresNote = true
	case from == StatusResolved && to == StatusProcessing:
		requiresNote = true
	case from == StatusClosed && to == StatusProcessing:
		requiresNote = true
	default:
		return ErrTransitionNotAllowed
	}

	if requiresNote && strings.TrimSpace(note) == "" {
		return ErrTransitionNoteRequired
	}
	return nil
}

type MessageType string

const (
	MessageTypeInitial    MessageType = "initial"
	MessageTypeSupplement MessageType = "supplement"
	MessageTypeStatus     MessageType = "status"
)

func (t MessageType) Valid() bool {
	switch t {
	case MessageTypeInitial, MessageTypeSupplement, MessageTypeStatus:
		return true
	default:
		return false
	}
}

type AuthorType string

const (
	AuthorTypeUser  AuthorType = "user"
	AuthorTypeAdmin AuthorType = "admin"
)

func (t AuthorType) Valid() bool {
	return t == AuthorTypeUser || t == AuthorTypeAdmin
}
