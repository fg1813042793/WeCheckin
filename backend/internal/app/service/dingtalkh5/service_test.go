package dingtalkh5

import (
	"encoding/json"
	"testing"

	"wecheckin-backend/backend/internal/model"
)

func TestNextStatusAfterSelfSubmit(t *testing.T) {
	withManager := DingTalkH5Review{EmployeeID: "lip", ManagerID: "cube", HRBPID: "lucky", Status: ReviewStatusDraft}
	if got := nextStatusAfterSelfSubmit(withManager); got != ReviewStatusManagerReview {
		t.Fatalf("next status with manager = %q, want %q", got, ReviewStatusManagerReview)
	}

	withoutManager := DingTalkH5Review{EmployeeID: "nick", ManagerID: "", HRBPID: "hrbp", Status: ReviewStatusDraft}
	if got := nextStatusAfterSelfSubmit(withoutManager); got != ReviewStatusHRFinal {
		t.Fatalf("next status without manager = %q, want %q", got, ReviewStatusHRFinal)
	}
}

func TestNormalizeUserIDKeepsStableAccountCharacters(t *testing.T) {
	if got := NormalizeUserID(" Rock.Admin_01 "); got != "rock.admin_01" {
		t.Fatalf("normalized id = %q", got)
	}
}

func TestUserPayloadCarriesPositionToDTO(t *testing.T) {
	user, err := sanitizeUserPayload(UserPayload{
		ID:               "lip",
		Name:             "Lip",
		Role:             "employee",
		Position:         "  Java 工程师  ",
		DepartmentLevel1: "M/H业务",
		DepartmentLevel2: "研发部",
		DepartmentLevel3: "Java开发一组",
		ManagerID:        "cube",
		HRBPID:           "lucky",
	}, nil)
	if err != nil {
		t.Fatalf("sanitize user payload: %v", err)
	}
	if got := user.Position; got != "Java 工程师" {
		t.Fatalf("model position = %q, want %q", got, "Java 工程师")
	}

	dto := userDTO(user)
	if got := dto.Position; got != "Java 工程师" {
		t.Fatalf("dto position = %q, want %q", got, "Java 工程师")
	}
}

func TestPerfUserMetadataRoundTripsThroughUserObj(t *testing.T) {
	user := model.DingTalkH5PerfUser{
		Account:                "lip",
		Name:                   "Lip",
		Role:                   "manager",
		Position:               "研发经理",
		Department:             "M/H业务 / 研发部",
		DepartmentLevel1:       "M/H业务",
		DepartmentLevel2:       "研发部",
		DepartmentLevel3:       "Java开发一组",
		ManagerAccount:         "david",
		HRBPAccount:            "nick",
		ResponsibleDepartments: encodeJSON([]string{"研发部", "产品部"}),
		Obj:                    `{"theme":"blue"}`,
	}

	raw := encodePerfUserObj(user.Obj, user)
	var obj map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &obj); err != nil {
		t.Fatalf("decode user obj: %v", err)
	}
	if obj["theme"] != "blue" {
		t.Fatalf("existing user_obj keys should be preserved, got %#v", obj["theme"])
	}

	stored := model.DingTalkH5PerfUser{Account: "lip", Name: "Lip", Obj: raw, Status: 1}
	hydratePerfUser(&stored)
	if stored.Role != "manager" || stored.Position != "研发经理" || stored.ManagerAccount != "david" || stored.HRBPAccount != "nick" {
		t.Fatalf("metadata was not restored: %#v", stored)
	}
	if got := decodeStringList(stored.ResponsibleDepartments); len(got) != 2 || got[0] != "研发部" || got[1] != "产品部" {
		t.Fatalf("responsible departments = %#v", got)
	}
}
