package workflowcore

import (
	"errors"
	"testing"
)

func TestValidateFormDataAcceptsStructuredAndLegacyAttachments(t *testing.T) {
	maximum := 2.0
	fields := []FormField{{
		Key:   "attachments",
		Label: "附件",
		Type:  FormFieldTypeAttachment,
		Rules: []FormValidationRule{{
			ID:   "attachment_count",
			Type: FormRuleSelectionCount,
			Max:  &maximum,
		}},
	}}

	structured := map[string]interface{}{
		"attachments": []interface{}{
			map[string]interface{}{
				"id":       "uploads/2026/09/03/expense.pdf",
				"name":     "报销凭证.pdf",
				"url":      "/uploads/2026/09/03/expense.pdf",
				"mimeType": "application/pdf",
				"size":     float64(1024),
			},
		},
	}
	if err := ValidateFormData(fields, structured, false); err != nil {
		t.Fatalf("structured attachment should be valid: %v", err)
	}

	legacy := map[string]interface{}{"attachments": []interface{}{"/uploads/legacy.pdf"}}
	if err := ValidateFormData(fields, legacy, false); err != nil {
		t.Fatalf("legacy string attachment should remain valid: %v", err)
	}
}

func TestValidateFormDataRejectsInvalidStructuredAttachment(t *testing.T) {
	fields := []FormField{{Key: "attachments", Label: "附件", Type: FormFieldTypeAttachment}}
	data := map[string]interface{}{
		"attachments": []interface{}{
			map[string]interface{}{
				"id":   "missing-url",
				"name": "无地址.pdf",
				"size": float64(10),
			},
		},
	}

	err := ValidateFormData(fields, data, false)
	if !errors.Is(err, ErrFormDataInvalid) {
		t.Fatalf("invalid structured attachment error = %v, want ErrFormDataInvalid", err)
	}
}
