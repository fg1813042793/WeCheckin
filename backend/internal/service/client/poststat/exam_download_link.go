package poststat

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"sync"

	exportlinkapp "wecheckin/backend/internal/modules/examrecordexportlink/application"
)

const answerDownloadURLPlaceholder = "{answerDownloadUrl}"

var ErrExamAnswerDownloadLinkIssuerUnavailable = errors.New("exam answer download link issuer is not initialized")

type ExamAnswerDownloadLinkIssuer interface {
	Issue(context.Context, exportlinkapp.IssueInput) (string, error)
}

var examAnswerDownloadLinkRegistry struct {
	sync.RWMutex
	issuer ExamAnswerDownloadLinkIssuer
}

func ConfigureExamAnswerDownloadLinkIssuer(issuer ExamAnswerDownloadLinkIssuer) {
	examAnswerDownloadLinkRegistry.Lock()
	examAnswerDownloadLinkRegistry.issuer = issuer
	examAnswerDownloadLinkRegistry.Unlock()
}

func configuredExamAnswerDownloadLinkIssuer() ExamAnswerDownloadLinkIssuer {
	examAnswerDownloadLinkRegistry.RLock()
	defer examAnswerDownloadLinkRegistry.RUnlock()
	return examAnswerDownloadLinkRegistry.issuer
}

func issueExamAnswerDownloadURL(ctx context.Context, issuer ExamAnswerDownloadLinkIssuer, examID, recordID uint, rawSettings string, rules []Rule) (string, error) {
	needsLink := false
	for _, rule := range rules {
		if strings.Contains(rule.MessageTemplate, answerDownloadURLPlaceholder) {
			needsLink = true
			break
		}
	}
	if !needsLink {
		return "", nil
	}
	if issuer == nil {
		return "", ErrExamAnswerDownloadLinkIssuerUnavailable
	}
	settings := struct {
		AnswerDownloadBaseURL    string `json:"answerDownloadBaseUrl"`
		AnswerDownloadFormat     string `json:"answerDownloadFormat"`
		AnswerDownloadExpireDays int    `json:"answerDownloadExpireDays"`
	}{AnswerDownloadFormat: "pdf", AnswerDownloadExpireDays: 7}
	if json.Unmarshal([]byte(rawSettings), &settings) != nil {
		return "", exportlinkapp.ErrInvalidGrant
	}
	if strings.TrimSpace(settings.AnswerDownloadFormat) == "" {
		settings.AnswerDownloadFormat = "pdf"
	}
	if settings.AnswerDownloadExpireDays == 0 {
		settings.AnswerDownloadExpireDays = 7
	}
	baseURL := strings.TrimSpace(settings.AnswerDownloadBaseURL)
	if baseURL == "" {
		return "", exportlinkapp.ErrInvalidGrant
	}
	return issuer.Issue(ctx, exportlinkapp.IssueInput{
		ExamID: examID, RecordID: recordID,
		Format: settings.AnswerDownloadFormat, ExpireDays: settings.AnswerDownloadExpireDays,
		BaseURL: baseURL,
	})
}
