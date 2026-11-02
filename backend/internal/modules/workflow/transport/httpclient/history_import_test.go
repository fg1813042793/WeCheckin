package httpclient

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/route/param"
	"wecheckin/backend/internal/model"
	workflowapp "wecheckin/backend/internal/modules/workflow/application"
	"wecheckin/backend/internal/support/dingtalkh5session"
	"wecheckin/backend/internal/workflowcore"
)

type historyImportServiceStub struct {
	runtimeServiceStub
	preset *workflowcore.HistoryImportConfig
}

func (stub *historyImportServiceStub) ListPublishedDefinitionsForStarter(_ context.Context, actorID string) ([]workflowapp.PublishedDefinition, error) {
	stub.actorID = actorID
	return []workflowapp.PublishedDefinition{{ID: 7, HistoryImport: stub.preset}}, nil
}

func (stub *historyImportServiceStub) GetPublishedDefinitionForStarter(_ context.Context, definitionID uint, actorID string) (*workflowapp.PublishedDefinition, error) {
	stub.actorID = actorID
	return &workflowapp.PublishedDefinition{ID: definitionID, HistoryImport: stub.preset}, nil
}

func TestHistoryImportPublishedTransportForH5AndLegacyClient(t *testing.T) {
	for _, h5 := range []bool{false, true} {
		for _, list := range []bool{false, true} {
			for _, enabled := range []bool{false, true} {
				preset := &workflowcore.HistoryImportConfig{
					Enabled: enabled, SourceMode: "latest_history", SourceStatus: "completed", MappingMode: "mapping", Overwrite: true,
					Mappings: []workflowcore.HistoryImportMapping{{SourceKey: "old", TargetKey: "new", Columns: []workflowcore.HistoryImportColumnMapping{{SourceKey: "source", TargetKey: "target"}}}},
				}
				stub := &historyImportServiceStub{preset: preset}
				handler := NewRuntimeHandler(stub)
				c, actor := newUserContext(42), "42"
				if h5 {
					c, actor = app.NewContext(1), "66"
					dingtalkh5session.SetAuth(c, &model.DingTalkH5PerfUser{ID: 66}, "token")
				}
				c.Params = append(c.Params, param.Param{Key: "id", Value: "7"})
				if list {
					handler.ListDefinitions(context.Background(), c)
				} else {
					handler.GetDefinition(context.Background(), c)
				}
				if stub.actorID != actor {
					t.Fatalf("authenticated actor=%s, want %s", stub.actorID, actor)
				}
				var body struct {
					Code int             `json:"code"`
					Data json.RawMessage `json:"data"`
				}
				if err := json.Unmarshal(c.Response.Body(), &body); err != nil {
					t.Fatal(err)
				}
				var result workflowapp.PublishedDefinition
				if list {
					var definitions []workflowapp.PublishedDefinition
					if err := json.Unmarshal(body.Data, &definitions); err != nil {
						t.Fatal(err)
					}
					if len(definitions) != 1 {
						t.Fatalf("unexpected definition list: %s", c.Response.Body())
					}
					result = definitions[0]
				} else if err := json.Unmarshal(body.Data, &result); err != nil {
					t.Fatal(err)
				}
				if result.ID != 7 || !reflect.DeepEqual(result.HistoryImport, preset) {
					t.Fatalf("h5=%v list=%v enabled=%v response=%s", h5, list, enabled, c.Response.Body())
				}
			}
		}
	}
}
