package dict

import (
	"errors"
	"strings"
	"testing"

	"wecheckin/backend/internal/model"
)

func TestValidateTypeCodeRequiresStableLowercaseKey(t *testing.T) {
	if got, err := validateTypeCode(" content_type "); err != nil || got != "content_type" {
		t.Fatalf("validateTypeCode() = %q, %v", got, err)
	}
	for _, value := range []string{"", "ContentType", "content type", "-content"} {
		if _, err := validateTypeCode(value); !errors.Is(err, ErrInvalidTypeCode) {
			t.Fatalf("validateTypeCode(%q) error = %v", value, err)
		}
	}
}

func TestLegacyTypePlaceholderIsRecognizedWithoutDroppingOtherEmptyValues(t *testing.T) {
	placeholder := model.SysDict{TypeCode: "content_type", TypeName: "内容分类", Label: "内容分类", Value: "", Sort: 0, Remark: ""}
	if !isLegacyTypePlaceholder(placeholder) {
		t.Fatal("legacy type placeholder must be hidden from dictionary items")
	}
	placeholder.Remark = "historical empty value"
	if isLegacyTypePlaceholder(placeholder) {
		t.Fatal("non-placeholder historical empty values must be preserved")
	}
}

func TestLegacyTypePlaceholderSQLCanBeReusedDuringTypeRename(t *testing.T) {
	if legacyPlaceholderPredicateSQL == "" || legacyPlaceholderSQL != "NOT ("+legacyPlaceholderPredicateSQL+")" {
		t.Fatalf("legacy placeholder SQL predicates are inconsistent: %q / %q", legacyPlaceholderPredicateSQL, legacyPlaceholderSQL)
	}
}

func TestDictionaryTypeJoinUsesExplicitLegacyCollation(t *testing.T) {
	for _, column := range []string{"d.dict_type_code", "t.dict_type_code"} {
		if !strings.Contains(dictTypeJoinSQL, column+" COLLATE utf8mb4_general_ci") {
			t.Fatalf("dictionary type join must normalize %s collation: %q", column, dictTypeJoinSQL)
		}
	}
}

func TestNormalizeStatusAcceptsOnlyEnabledAndDisabled(t *testing.T) {
	for _, status := range []int{0, 1} {
		if got, err := normalizeStatus(status); err != nil || got != status {
			t.Fatalf("normalizeStatus(%d) = %d, %v", status, got, err)
		}
	}
	if _, err := normalizeStatus(2); !errors.Is(err, ErrInvalidStatus) {
		t.Fatalf("normalizeStatus(2) error = %v", err)
	}
}
