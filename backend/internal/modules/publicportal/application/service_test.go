package application

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	pm "wecheckin/backend/internal/model/publicportal"
)

type fakeStore struct {
	settings  pm.Settings
	items     []pm.Item
	sources   []Source
	actor     Actor
	err       error
	inTx      bool
	commits   int
	rollbacks int
	writeErr  error
}

func fixture() *fakeStore {
	return &fakeStore{settings: pm.Settings{ID: 1, Version: 1, Enabled: true, Title: "活动", PublicBaseURL: "https://forms.example.com"}, actor: Actor{ID: 7}, sources: []Source{{Kind: Survey, ID: 10, ManagementTitle: "内部名称", DisplayTitle: "公开名称", Status: 1, CanManage: true, SettingsValid: true}}}
}
func (f *fakeStore) Transaction(ctx context.Context, fn func(Store) error) error {
	next := *f
	next.items = append([]pm.Item{}, f.items...)
	next.inTx = true
	if err := fn(&next); err != nil {
		f.rollbacks++
		return err
	}
	next.inTx = false
	next.commits++
	*f = next
	return nil
}
func (f *fakeStore) Actor(context.Context, uint) (Actor, error)    { return f.actor, f.err }
func (f *fakeStore) Settings(context.Context) (pm.Settings, error) { return f.settings, f.err }
func (f *fakeStore) LockSettings(context.Context) (pm.Settings, error) {
	if !f.inTx {
		panic("missing transaction")
	}
	return f.settings, f.err
}
func (f *fakeStore) SaveSettings(_ context.Context, s pm.Settings, v uint64) error {
	if !f.inTx {
		panic("missing transaction")
	}
	if f.writeErr != nil {
		return f.writeErr
	}
	if v != f.settings.Version {
		return ErrVersionConflict
	}
	f.settings = s
	return nil
}
func (f *fakeStore) Items(context.Context) ([]pm.Item, error) {
	return append([]pm.Item{}, f.items...), f.err
}
func (f *fakeStore) InsertItem(_ context.Context, item *pm.Item) error {
	if !f.inTx {
		panic("missing transaction")
	}
	if f.writeErr != nil {
		return f.writeErr
	}
	item.ID = uint64(len(f.items) + 1)
	f.items = append(f.items, *item)
	return nil
}
func (f *fakeStore) UpdateItem(_ context.Context, item pm.Item, v uint64) error {
	if !f.inTx {
		panic("missing transaction")
	}
	if f.writeErr != nil {
		return f.writeErr
	}
	for i, old := range f.items {
		if old.ID == item.ID && old.Version == v {
			f.items[i] = item
			return nil
		}
	}
	return ErrVersionConflict
}
func (f *fakeStore) DeleteItem(_ context.Context, id, v uint64) error {
	if !f.inTx {
		panic("missing transaction")
	}
	for i, old := range f.items {
		if old.ID == id && old.Version == v {
			f.items = append(f.items[:i], f.items[i+1:]...)
			return nil
		}
	}
	return ErrVersionConflict
}
func (f *fakeStore) Sources(_ context.Context, k Kind, ids []uint, actor uint) ([]Source, error) {
	out := []Source{}
	for _, s := range f.sources {
		for _, id := range ids {
			if s.Kind == k && s.ID == id {
				if actor > 0 && !s.CanManage {
					s = Source{Kind: k, ID: id}
				}
				out = append(out, s)
			}
		}
	}
	return out, f.err
}
func (f *fakeStore) Options(context.Context, uint, OptionsQuery) (SourcePage, error) {
	return SourcePage{List: f.sources, Total: int64(len(f.sources))}, f.err
}

func TestPublicRevalidatesSourcesAndOnlyExposesWhitelist(t *testing.T) {
	f := fixture()
	f.items = []pm.Item{{ID: 1, Kind: "survey", SourceID: 10, Enabled: true, Version: 1}}
	s := NewService(f)
	got, err := s.List(context.Background())
	if err != nil || len(got.Items) != 1 {
		t.Fatalf("got=%+v err=%v", got, err)
	}
	if got.Items[0].URL != "/sf/10" || got.Items[0].Title != "公开名称" {
		t.Fatalf("item=%+v", got.Items[0])
	}
	b, _ := json.Marshal(got)
	if !strings.Contains(string(b), `"managementTitle":"内部名称"`) {
		t.Fatalf("public cards must include their source management name: %s", b)
	}
	for _, key := range []string{"settings", "schema", "version", "publicBaseUrl", "createdBy"} {
		if strings.Contains(string(b), key) {
			t.Fatalf("leaked %s: %s", key, b)
		}
	}
	for _, change := range []func(*Source){func(s *Source) { s.Visibility = 1 }, func(s *Source) { s.Status = 0 }, func(s *Source) { s.SettingsValid = false }} {
		original := f.sources[0]
		change(&f.sources[0])
		got, err = s.List(context.Background())
		if err != nil || len(got.Items) != 0 {
			t.Fatalf("source not hidden: %+v %v", got, err)
		}
		f.sources[0] = original
	}
	f.sources = nil
	got, err = s.List(context.Background())
	if err != nil || len(got.Items) != 0 {
		t.Fatal("deleted source visible")
	}
}

func TestPublicManagementNamesFollowSourceWithoutReplacingDisplayTitles(t *testing.T) {
	f := fixture()
	f.sources = []Source{
		{Kind: Survey, ID: 10, ManagementTitle: "DISC-研发岗位", DisplayTitle: "DISC测试", Status: 1, SettingsValid: true},
		{Kind: Exam, ID: 10, ManagementTitle: "DISC-岗位测验", DisplayTitle: "DISC测试", Status: 1, SettingsValid: true},
	}
	f.items = []pm.Item{{ID: 1, Kind: "survey", SourceID: 10, Enabled: true}, {ID: 2, Kind: "exam", SourceID: 10, Enabled: true}}
	s := NewService(f)
	got, err := s.List(context.Background())
	if err != nil || len(got.Items) != 2 {
		t.Fatalf("items=%+v err=%v", got, err)
	}
	raw, _ := json.Marshal(got.Items)
	for _, title := range []string{"DISC-研发岗位", "DISC-岗位测验"} {
		if !strings.Contains(string(raw), title) {
			t.Fatalf("missing source name %s", title)
		}
	}
	if got.Items[0].Title != "DISC测试" || got.Items[1].Title != "DISC测试" {
		t.Fatal("display titles changed")
	}
	f.sources[0].ManagementTitle = "DISC-产品岗位"
	got, err = s.Preview(context.Background(), 7)
	raw, _ = json.Marshal(got.Items)
	if err != nil || !strings.Contains(string(raw), "DISC-产品岗位") || strings.Contains(string(raw), "DISC-研发岗位") {
		t.Fatalf("stale management name: %s %v", raw, err)
	}
	f.sources[0].Visibility = 2
	got, err = s.List(context.Background())
	raw, _ = json.Marshal(got.Items)
	if err != nil || len(got.Items) != 1 || strings.Contains(string(raw), "DISC-产品岗位") {
		t.Fatalf("private source name exposed: %s %v", raw, err)
	}
}

func TestPortalCategoryFollowsCurrentSourceAcrossAllViews(t *testing.T) {
	f := fixture()
	f.sources[0].Kind = Exam
	f.items = []pm.Item{{ID: 1, Kind: "exam", SourceID: 10, Enabled: true}}
	s := NewService(f)
	ctx := context.Background()
	check := func(value interface{}, err error, category string) {
		t.Helper()
		raw, _ := json.Marshal(value)
		if err != nil || !strings.Contains(string(raw), `"category":"`+category+`"`) {
			t.Fatalf("current category missing: %s err=%v", raw, err)
		}
	}
	for _, category := range []string{"面试题", "培训测验", ""} {
		// JSON fixture also reproduces the pre-fix projection silently dropping category.
		if err := json.Unmarshal([]byte(`{"Category":"`+category+`"}`), &f.sources[0]); err != nil {
			t.Fatal(err)
		}
		public, err := s.List(ctx)
		check(public, err, category)
		if len(public.Items) != 1 || public.Items[0].Kind != Exam || public.Items[0].URL != "/ef/10" {
			t.Fatalf("category changed source identity: %+v", public)
		}
		preview, err := s.Preview(ctx, 7)
		check(preview, err, category)
		admin, err := s.Get(ctx, 7)
		check(admin, err, category)
		options, err := s.Options(ctx, 7, OptionsQuery{Kind: Exam})
		check(options, err, category)
	}
	_ = json.Unmarshal([]byte(`{"Category":"不可公开分类"}`), &f.sources[0])
	f.sources[0].CanManage = false
	admin, err := s.Get(ctx, 7)
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := json.Marshal(admin)
	if strings.Contains(string(raw), "不可公开分类") {
		t.Fatal("category leaked outside administrator scope")
	}
	f.sources[0].Visibility = 2
	public, err := s.List(ctx)
	if err != nil || len(public.Items) != 0 {
		t.Fatal("category caused private source to become public")
	}
}

func TestPreviewBypassesOnlyMasterSwitchAndRequiresActor(t *testing.T) {
	f := fixture()
	f.settings.Enabled = false
	f.items = []pm.Item{{ID: 1, Kind: "survey", SourceID: 10, Enabled: true}, {ID: 2, Kind: "exam", SourceID: 11, Enabled: false}}
	s := NewService(f)
	got, err := s.List(context.Background())
	if err != nil || got.Enabled || len(got.Items) != 0 {
		t.Fatalf("disabled=%+v %v", got, err)
	}
	got, err = s.Preview(context.Background(), 7)
	if err != nil || !got.Enabled || len(got.Items) != 1 {
		t.Fatalf("preview=%+v %v", got, err)
	}
	f.sources[0].Visibility = 2
	got, err = s.Preview(context.Background(), 7)
	if err != nil || len(got.Items) != 0 {
		t.Fatal("preview exposed private source")
	}
	if _, err = s.Preview(context.Background(), 0); !errors.Is(err, ErrForbidden) {
		t.Fatalf("anonymous preview: %v", err)
	}
}

func TestPublicTimeStatesTemplatesOverridesAndOrdering(t *testing.T) {
	f := fixture()
	now := time.Now().UnixMilli()
	f.sources = append(f.sources, Source{Kind: Exam, ID: 11, ManagementTitle: "Exam", Status: 1, SettingsValid: true, FillTemplate: "https://evil.example", LoginRequired: true, PasswordRequired: true, StartTime: now + 60000})
	f.sources[0].EndTime = now - 60000
	f.sources[0].FillTemplate = "sf1"
	f.items = []pm.Item{{ID: 3, Kind: "exam", SourceID: 11, Enabled: true, Sort: 1}, {ID: 2, Kind: "survey", SourceID: 10, Enabled: true, Sort: 1, Title: "覆盖", Description: "说明"}}
	s := NewService(f)
	got, err := s.List(context.Background())
	if err != nil || len(got.Items) != 1 || got.Items[0].State != "upcoming" || got.Items[0].URL != "/ef/11" || !got.Items[0].LoginRequired || !got.Items[0].PasswordRequired {
		t.Fatalf("got=%+v err=%v", got, err)
	}
	f.settings.ShowEnded = true
	got, err = s.List(context.Background())
	if err != nil || len(got.Items) != 2 || got.Items[0].ID != 2 || got.Items[0].State != "ended" || got.Items[0].Title != "覆盖" || got.Items[0].Description != "说明" || !strings.Contains(got.Items[0].URL, "/sf1/") {
		t.Fatalf("got=%+v err=%v", got, err)
	}
}

func TestMutationScopeVersionsDuplicatesLimitAndRollback(t *testing.T) {
	f := fixture()
	s := NewService(f)
	ctx := context.Background()
	input := ItemInput{Kind: Survey, SourceID: 10, Enabled: true}
	item, err := s.CreateItem(ctx, 7, input)
	if err != nil || item.Version != 1 || f.commits != 1 {
		t.Fatalf("create=%+v %v", item, err)
	}
	if _, err = s.CreateItem(ctx, 7, input); !errors.Is(err, ErrDuplicateSource) {
		t.Fatalf("duplicate=%v", err)
	}
	if _, err = s.UpdateItem(ctx, 7, item.ID, ItemUpdate{ItemInput: input, Version: 99}); !errors.Is(err, ErrVersionConflict) {
		t.Fatalf("version=%v", err)
	}
	f.sources[0].CanManage = false
	admin, err := s.Get(ctx, 7)
	if err != nil || admin.Items[0].CanManage || admin.Items[0].ManagementTitle != "" || admin.Items[0].DisplayTitle != "" {
		t.Fatalf("unscoped=%+v %v", admin, err)
	}
	if _, err = s.DeleteItem(ctx, 7, item.ID, 1); !errors.Is(err, ErrForbidden) {
		t.Fatalf("out-of-scope creator delete=%v", err)
	}
	if _, err = s.UpdateItem(ctx, 7, item.ID, ItemUpdate{ItemInput: input, Version: 1}); !errors.Is(err, ErrForbidden) {
		t.Fatalf("out-of-scope update=%v", err)
	}
	f.sources = nil
	if _, err = s.DeleteItem(ctx, 7, item.ID, 1); err != nil {
		t.Fatalf("creator orphan cleanup=%v", err)
	}
	f = fixture()
	s = NewService(f)
	f.writeErr = errors.New("write failed")
	if _, err = s.CreateItem(ctx, 7, input); err == nil || len(f.items) != 0 || f.rollbacks != 1 {
		t.Fatalf("rollback=%+v %v", f, err)
	}
	f.writeErr = nil
	for i := 0; i < MaxItems; i++ {
		f.items = append(f.items, pm.Item{ID: uint64(i + 1), Kind: "exam", SourceID: uint(i + 1)})
	}
	if _, err = s.CreateItem(ctx, 7, input); !errors.Is(err, ErrItemLimit) {
		t.Fatalf("limit=%v", err)
	}
}

func TestSettingsValidationAndOptimisticVersion(t *testing.T) {
	f := fixture()
	s := NewService(f)
	ctx := context.Background()
	base := Settings{Enabled: true, Title: "页面", Version: 1}
	for _, logo := range []string{"//evil.example/x", "javascript:alert(1)", "https://user:pass@example.com/x", "/\\evil.example/x"} {
		input := base
		input.LogoURL = logo
		if _, err := s.SaveSettings(ctx, 7, input); !errors.Is(err, ErrInvalidLogo) {
			t.Errorf("logo %q accepted: %v", logo, err)
		}
	}
	base.LogoURL = "/uploads/logo.png"
	got, err := s.SaveSettings(ctx, 7, base)
	if err != nil || got.Version != 2 || f.settings.UpdatedAt <= 0 || f.settings.UpdatedBy != 7 {
		t.Fatalf("save=%+v %v", got, err)
	}
	if _, err = s.SaveSettings(ctx, 7, base); !errors.Is(err, ErrVersionConflict) {
		t.Fatalf("stale save=%v", err)
	}
	base.Version = 2
	base.Title = strings.Repeat("字", 101)
	if _, err = s.SaveSettings(ctx, 7, base); !errors.Is(err, ErrInvalidArgument) {
		t.Fatalf("title=%v", err)
	}
}

func TestPreviewAndPublicListWorkWithoutSavedDomain(t *testing.T) {
	f := fixture()
	f.settings.PublicBaseURL = ""
	f.settings.Enabled = false
	f.items = []pm.Item{{ID: 1, Kind: "survey", SourceID: 10, Enabled: true}}
	s := NewService(f)
	got, err := s.Preview(context.Background(), 7)
	if err != nil || !got.Enabled || len(got.Items) != 1 || got.Items[0].URL != "/sf/10" {
		t.Fatalf("preview=%+v error=%v", got, err)
	}
	got, err = s.List(context.Background())
	if err != nil || got.Enabled || len(got.Items) != 0 {
		t.Fatalf("closed page bypassed: %+v %v", got, err)
	}
	f.settings.Enabled = true
	got, err = s.List(context.Background())
	if err != nil || len(got.Items) != 1 || got.Items[0].URL != "/sf/10" {
		t.Fatalf("public=%+v error=%v", got, err)
	}
}

func TestLegacyDomainsNeverControlPublicLinksOrBlockSaving(t *testing.T) {
	for _, legacy := range []string{"", "https://old.example.com", "http://127.0.0.1:3000", "javascript:alert(1)", "https://user:pass@example.com/path"} {
		f := fixture()
		f.settings.PublicBaseURL = legacy
		f.items = []pm.Item{{ID: 1, Kind: "survey", SourceID: 10, Enabled: true}}
		s := NewService(f)
		got, err := s.List(context.Background())
		if err != nil || len(got.Items) != 1 || got.Items[0].URL != "/sf/10" {
			t.Fatalf("legacy domain affected output: %+v %v", got, err)
		}
		settings, err := s.SaveSettings(context.Background(), 7, Settings{Enabled: true, Title: "页面", Version: 1, PublicBaseURL: legacy})
		if err != nil || !settings.Enabled || settings.PublicBaseURL != "" {
			t.Fatalf("legacy domain blocked save: %+v %v", settings, err)
		}
		admin, err := s.Get(context.Background(), 7)
		if err != nil {
			t.Fatal(err)
		}
		raw, _ := json.Marshal(admin.Settings)
		if strings.Contains(string(raw), "publicBaseUrl") {
			t.Fatalf("removed field still returned: %s", raw)
		}
	}
}

func TestCanDisableOriginalUnpublishedSourceButCannotEnableOrReplace(t *testing.T) {
	for _, private := range []bool{false, true} {
		f := fixture()
		f.items = []pm.Item{{ID: 1, Kind: "survey", SourceID: 10, Enabled: true, Version: 1, CreatedBy: 7}}
		if private {
			f.sources[0].Visibility = 1
		} else {
			f.sources[0].Status = 0
		}
		s := NewService(f)
		input := ItemUpdate{ItemInput: ItemInput{Kind: Survey, SourceID: 10, Enabled: false, Title: "下架说明", Sort: 19}, Version: 1}
		got, err := s.UpdateItem(context.Background(), 7, 1, input)
		if err != nil || got.Enabled || got.Version != 2 || got.Title != "下架说明" {
			t.Fatalf("disable=%+v %v", got, err)
		}
		input.Version = 2
		input.Enabled = true
		if _, err = s.UpdateItem(context.Background(), 7, 1, input); !errors.Is(err, ErrSourceUnavailable) {
			t.Fatalf("re-enable=%v", err)
		}
		input.Enabled = false
		f.sources[0].ID = 11
		input.SourceID = 11
		if _, err = s.UpdateItem(context.Background(), 7, 1, input); !errors.Is(err, ErrSourceUnavailable) {
			t.Fatalf("orphan edit=%v", err)
		}
	}
}

func TestOrphanCleanupRequiresCreatorOrReservedSuperAdmin(t *testing.T) {
	f := fixture()
	f.sources = nil
	f.items = []pm.Item{{ID: 1, Kind: "survey", SourceID: 10, CreatedBy: 8, Version: 1}}
	s := NewService(f)
	ctx := context.Background()
	got, err := s.Get(ctx, 7)
	if err != nil || got.Items[0].CanManage {
		t.Fatalf("unowned orphan=%+v %v", got, err)
	}
	if _, err = s.DeleteItem(ctx, 7, 1, 1); !errors.Is(err, ErrForbidden) {
		t.Fatalf("unowned cleanup=%v", err)
	}
	f.actor.ReservedSuperAdmin = true
	got, err = s.Get(ctx, 7)
	if err != nil || !got.Items[0].CanManage {
		t.Fatalf("super orphan=%+v %v", got, err)
	}
	if _, err = s.DeleteItem(ctx, 7, 1, 1); err != nil || len(f.items) != 0 {
		t.Fatalf("super cleanup=%v", err)
	}
}
func TestReplacementMustBeScopedAndEligibleEvenWhenDisabled(t *testing.T) {
	f := fixture()
	f.items = []pm.Item{{ID: 1, Kind: "survey", SourceID: 10, CreatedBy: 7, Version: 1}}
	f.sources = append(f.sources, Source{ID: 20, Kind: Exam, Status: 0, CanManage: true, SettingsValid: true})
	s := NewService(f)
	input := ItemUpdate{ItemInput: ItemInput{Kind: Exam, SourceID: 20, Enabled: false}, Version: 1}
	if _, err := s.UpdateItem(context.Background(), 7, 1, input); !errors.Is(err, ErrSourceUnavailable) {
		t.Fatalf("unpublished replacement=%v", err)
	}
	f.sources[1].Status = 1
	f.sources[1].CanManage = false
	if _, err := s.UpdateItem(context.Background(), 7, 1, input); !errors.Is(err, ErrForbidden) {
		t.Fatalf("unscoped replacement=%v", err)
	}
	f.sources[1].CanManage = true
	got, err := s.UpdateItem(context.Background(), 7, 1, input)
	if err != nil || got.SourceID != 20 || got.Version != 2 || got.Kind != Exam {
		t.Fatalf("replacement=%+v %v", got, err)
	}
}
func TestOptionsBoundsAndTypedProjection(t *testing.T) {
	s := NewService(fixture())
	ctx := context.Background()
	for _, q := range []OptionsQuery{{Kind: "bad"}, {Kind: Survey, Page: -1}, {Kind: Survey, PageSize: 101}, {Kind: Survey, Keyword: strings.Repeat("字", 256)}} {
		if _, err := s.Options(ctx, 7, q); !errors.Is(err, ErrInvalidArgument) {
			t.Fatalf("query=%+v err=%v", q, err)
		}
	}
	got, err := s.Options(ctx, 7, OptionsQuery{Kind: Survey})
	if err != nil || got.Page != 1 || got.PageSize != 20 || len(got.List) != 1 || got.List[0].DisplayTitle != "公开名称" {
		t.Fatalf("options=%+v %v", got, err)
	}
	if _, err = s.Options(ctx, 0, OptionsQuery{Kind: Survey}); !errors.Is(err, ErrForbidden) {
		t.Fatalf("anonymous options=%v", err)
	}
}
