package exam

import "errors"

var (
	ErrExamNotPublished    = errors.New("exam not published")
	ErrExamNotStarted      = errors.New("exam not started")
	ErrExamEnded           = errors.New("exam ended")
	ErrExamMaxAttempts     = errors.New("exam max attempts reached")
	ErrExamPaperNotFound   = errors.New("exam paper not found")
	ErrExamRecordNotFound  = errors.New("exam record not found")
	ErrExamRecordSubmitted = errors.New("exam record already submitted")
)
