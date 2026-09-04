package workflowservice

import (
	"strings"
	"testing"

	"wecheckin/backend/internal/workflowcore"
)

func TestBuildVersionChangeSummaryCapturesSemanticChanges(t *testing.T) {
	before := newDefaultDefinition("expense", "费用审批")
	before.Form = []workflowcore.FormField{{Key: "amount", Label: "金额", Type: "number"}}
	before.Nodes = append(before.Nodes[:1], workflowcore.Node{
		ID: "approve", Type: workflowcore.NodeTypeApproval, Name: "主管审批",
		Assignee: &workflowcore.Assignee{Type: "manager", Value: "1"},
	}, before.Nodes[1])
	before.Edges = []workflowcore.Edge{
		{ID: "start-approve", Source: "start", Target: "approve"},
		{ID: "approve-end", Source: "approve", Target: "end"},
	}

	after := before
	after.Name = "费用报销审批"
	after.Form = []workflowcore.FormField{
		{Key: "amount", Label: "报销金额", Type: "amount", Required: true},
		{Key: "reason", Label: "报销事由", Type: "textarea"},
	}
	after.Nodes = append([]workflowcore.Node(nil), before.Nodes...)
	after.Nodes[1].Name = "部门负责人审批"
	after.Nodes[1].Notification = &workflowcore.NotificationConfig{
		Enabled: true, Channels: []string{"in_app"}, Title: "待审批", Content: "请处理 {{workflowName}}",
	}

	summary := buildVersionChangeSummary(
		3,
		versionSnapshot{Metadata: versionMetadata{Name: "费用审批", Category: "财务"}, Definition: before},
		versionSnapshot{Metadata: versionMetadata{Name: "费用报销审批", Category: "财务"}, Definition: after},
	)
	if summary.BaseVersion != 3 || summary.ChangeCount < 5 {
		t.Fatalf("unexpected summary: %#v", summary)
	}
	joined := changeItemText(summary.Items)
	for _, want := range []string{"流程名称", "新增字段", "报销金额", "部门负责人审批", "通知"} {
		if !strings.Contains(joined, want) {
			t.Fatalf("change summary missing %q: %s", want, joined)
		}
	}
}

func TestBuildVersionChangeSummaryMarksFirstPublish(t *testing.T) {
	definition := newDefaultDefinition("leave", "请假审批")
	summary := buildVersionChangeSummary(0, versionSnapshot{}, versionSnapshot{
		Metadata:   versionMetadata{Name: "请假审批"},
		Definition: definition,
	})
	if summary.Headline != "首次发布" || summary.ChangeCount != 1 || len(summary.Items) != 1 {
		t.Fatalf("unexpected first publish summary: %#v", summary)
	}
}

func TestBuildVersionChangeSummaryCapturesApprovalResultNotificationChange(t *testing.T) {
	before := newDefaultDefinition("expense", "费用审批")
	before.Nodes = append(before.Nodes[:1], workflowcore.Node{
		ID: "approve", Type: workflowcore.NodeTypeApproval, Name: "主管审批",
		ApprovalMode: workflowcore.ApprovalModeSingle,
		Assignee:     &workflowcore.Assignee{Type: workflowcore.AssigneeTypeManager, Value: "direct_manager"},
	}, before.Nodes[1])
	before.Edges = []workflowcore.Edge{
		{ID: "start-approve", Source: "start", Target: "approve"},
		{ID: "approve-end", Source: "approve", Target: "end"},
	}
	after := before
	after.Nodes = append([]workflowcore.Node(nil), before.Nodes...)
	after.Nodes[1].ResultNotification = &workflowcore.NotificationConfig{
		Enabled: true, Channels: []string{workflowcore.NotificationChannelInApp},
		Title: "{{workflowName}}审批结果", Content: "{{nodeName}}{{result}}",
	}

	summary := buildVersionChangeSummary(
		1,
		versionSnapshot{Metadata: versionMetadata{Name: before.Name}, Definition: before},
		versionSnapshot{Metadata: versionMetadata{Name: after.Name}, Definition: after},
	)
	if summary.ChangeCount != 1 || summary.Items[0].Category != versionChangeCategoryNotification {
		t.Fatalf("result notification change summary = %#v", summary)
	}
}

func TestVersionDeleteBlockReasonProtectsCurrentAndReferencedVersions(t *testing.T) {
	cases := []struct {
		name          string
		version       int
		current       int
		instances     int64
		drafts        int64
		wantSubstring string
	}{
		{name: "current", version: 5, current: 5, wantSubstring: "当前"},
		{name: "instances", version: 4, current: 5, instances: 2, wantSubstring: "2 个流程实例"},
		{name: "drafts", version: 4, current: 5, drafts: 1, wantSubstring: "1 份发起草稿"},
		{name: "unused", version: 3, current: 5},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			reason := versionDeleteBlockReason(tc.version, tc.current, tc.instances, tc.drafts)
			if tc.wantSubstring == "" && reason != "" {
				t.Fatalf("unused version should be deletable, got %q", reason)
			}
			if tc.wantSubstring != "" && !strings.Contains(reason, tc.wantSubstring) {
				t.Fatalf("block reason = %q, want substring %q", reason, tc.wantSubstring)
			}
		})
	}
}

func TestBuildVersionChangeSummaryDoesNotDuplicateNestedFieldChanges(t *testing.T) {
	before := newDefaultDefinition("review", "评审")
	before.Form = []workflowcore.FormField{{
		Key: "group", Label: "评审信息", Type: "group",
		Fields: []workflowcore.FormField{{Key: "comment", Label: "原评价", Type: "textarea"}},
	}}
	after := before
	after.Form = []workflowcore.FormField{{
		Key: "group", Label: "评审信息", Type: "group",
		Fields: []workflowcore.FormField{{Key: "comment", Label: "评价", Type: "textarea"}},
	}}

	summary := buildVersionChangeSummary(
		1,
		versionSnapshot{Metadata: versionMetadata{Name: "评审"}, Definition: before},
		versionSnapshot{Metadata: versionMetadata{Name: "评审"}, Definition: after},
	)
	if summary.ChangeCount != 1 || len(summary.Items) != 1 || summary.Items[0].Detail != "原评价 -> 评价" {
		t.Fatalf("nested field change should be reported once, got %#v", summary)
	}
}

func TestNewDefinitionVersionModelRecordsRollbackHistory(t *testing.T) {
	definition := newDefaultDefinition("leave", "请假审批")
	snapshot := versionSnapshot{Metadata: versionMetadata{Name: "请假审批", Category: "人事"}, Definition: definition}
	summary := buildVersionChangeSummary(4, versionSnapshot{Metadata: versionMetadata{Name: "旧名称"}, Definition: definition}, snapshot)

	item, err := newDefinitionVersionModel(7, 5, 66, 123456, `{"name":"请假审批"}`, "<bpmn />", snapshot, summary, "回滚测试", 2)
	if err != nil {
		t.Fatalf("new version model: %v", err)
	}
	if item.DefinitionID != 7 || item.Version != 5 || item.ChangeBaseVersion != 4 || item.RollbackFromVersion != 2 {
		t.Fatalf("unexpected rollback version identity: %#v", item)
	}
	if item.MetadataJSON == "" || item.ChangeSummaryJSON == "" || len(item.ContentHash) != 64 || item.PublishNote != "回滚测试" {
		t.Fatalf("rollback version history is incomplete: %#v", item)
	}
}

func changeItemText(items []VersionChangeItem) string {
	var result strings.Builder
	for _, item := range items {
		result.WriteString(item.Title)
		result.WriteString(item.Detail)
	}
	return result.String()
}
