package application

import (
	"errors"

	"wecheckin/backend/internal/modules/userfeedback/domain"
)

var (
	ErrServiceUnavailable        = errors.New("feedback application service is not initialized")
	ErrInvalidArgument           = errors.New("invalid feedback argument")
	ErrFeedbackNotFound          = errors.New("feedback not found or inaccessible")
	ErrSupplementNotAllowed      = errors.New("feedback status does not allow supplements")
	ErrTransitionNotAllowed      = domain.ErrTransitionNotAllowed
	ErrTransitionNoteRequired    = domain.ErrTransitionNoteRequired
	ErrVersionConflict           = errors.New("feedback version conflict")
	ErrDuplicateRequest          = errors.New("duplicate feedback request")
	ErrTransactionOutcomeUnknown = errors.New("feedback transaction outcome is unknown")
	ErrDailyLimitExceeded        = errors.New("feedback daily limit exceeded")
	ErrAttachmentLimitExceeded   = errors.New("feedback attachment limit exceeded")
	ErrAttachmentTooLarge        = errors.New("feedback attachment is too large")
	ErrAttachmentTypeNotAllowed  = errors.New("feedback attachment type is not allowed")
	ErrStorageFailed             = errors.New("feedback attachment storage failed")
)
