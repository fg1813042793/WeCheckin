package domain

import (
	"errors"
	"testing"
)

func TestStatusValid(t *testing.T) {
	validStatuses := []Status{
		StatusPending,
		StatusProcessing,
		StatusResolved,
		StatusClosed,
	}
	for _, status := range validStatuses {
		if !status.Valid() {
			t.Errorf("Status(%q).Valid() = false, want true", status)
		}
	}

	if Status("unknown").Valid() {
		t.Fatal("Status(unknown).Valid() = true, want false")
	}
}

func TestValidateTransition(t *testing.T) {
	statuses := []Status{
		StatusPending,
		StatusProcessing,
		StatusResolved,
		StatusClosed,
	}
	allowed := map[[2]Status]bool{
		{StatusPending, StatusProcessing}:  true,
		{StatusPending, StatusClosed}:      true,
		{StatusProcessing, StatusResolved}: true,
		{StatusResolved, StatusClosed}:     true,
		{StatusResolved, StatusProcessing}: true,
		{StatusClosed, StatusProcessing}:   true,
	}

	for _, from := range statuses {
		for _, to := range statuses {
			from, to := from, to
			t.Run(string(from)+"_to_"+string(to), func(t *testing.T) {
				err := ValidateTransition(from, to, "handled")
				if allowed[[2]Status{from, to}] {
					if err != nil {
						t.Fatalf("ValidateTransition(%q, %q) error = %v, want nil", from, to, err)
					}
					return
				}
				if !errors.Is(err, ErrTransitionNotAllowed) {
					t.Fatalf("ValidateTransition(%q, %q) error = %v, want ErrTransitionNotAllowed", from, to, err)
				}
			})
		}
	}
}

func TestValidateTransitionNoteRequirement(t *testing.T) {
	tests := []struct {
		name string
		from Status
		to   Status
		want error
	}{
		{name: "pending to processing", from: StatusPending, to: StatusProcessing},
		{name: "pending to closed", from: StatusPending, to: StatusClosed, want: ErrTransitionNoteRequired},
		{name: "processing to resolved", from: StatusProcessing, to: StatusResolved, want: ErrTransitionNoteRequired},
		{name: "resolved to closed", from: StatusResolved, to: StatusClosed, want: ErrTransitionNoteRequired},
		{name: "resolved to processing", from: StatusResolved, to: StatusProcessing, want: ErrTransitionNoteRequired},
		{name: "closed to processing", from: StatusClosed, to: StatusProcessing, want: ErrTransitionNoteRequired},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := ValidateTransition(test.from, test.to, " \t\n ")
			if !errors.Is(err, test.want) {
				t.Fatalf("ValidateTransition(%q, %q, whitespace) error = %v, want %v", test.from, test.to, err, test.want)
			}
		})
	}
}

func TestStatusAllowsSupplement(t *testing.T) {
	tests := []struct {
		status Status
		want   bool
	}{
		{status: StatusPending, want: true},
		{status: StatusProcessing, want: true},
		{status: StatusResolved, want: false},
		{status: StatusClosed, want: false},
		{status: Status("unknown"), want: false},
	}

	for _, test := range tests {
		if got := test.status.AllowsSupplement(); got != test.want {
			t.Errorf("Status(%q).AllowsSupplement() = %t, want %t", test.status, got, test.want)
		}
	}
}

func TestMessageTypeValid(t *testing.T) {
	validTypes := []MessageType{MessageTypeInitial, MessageTypeSupplement, MessageTypeStatus}
	for _, messageType := range validTypes {
		if !messageType.Valid() {
			t.Errorf("MessageType(%q).Valid() = false, want true", messageType)
		}
	}
	if MessageType("unknown").Valid() {
		t.Fatal("MessageType(unknown).Valid() = true, want false")
	}
}

func TestAuthorTypeValid(t *testing.T) {
	validTypes := []AuthorType{AuthorTypeUser, AuthorTypeAdmin}
	for _, authorType := range validTypes {
		if !authorType.Valid() {
			t.Errorf("AuthorType(%q).Valid() = false, want true", authorType)
		}
	}
	if AuthorType("unknown").Valid() {
		t.Fatal("AuthorType(unknown).Valid() = true, want false")
	}
}
