package home

import (
	"context"
	"encoding/json"
	"sort"

	"gorm.io/gorm"
	setupservice "wecheckin/backend/internal/service/admin/setup"
	"wecheckin/backend/internal/support/dept"
	"wecheckin/backend/internal/support/media"
	"wecheckin/backend/internal/support/publish"
	"wecheckin/backend/internal/model"
	"wecheckin/backend/pkg/database"
)

func enrollBaseQuery(db *gorm.DB, deptWhere string) *gorm.DB {
	return db.Model(&model.Enroll{}).Where("`enroll_status` = 1").Where(deptWhere)
}

func eventBaseQuery(db *gorm.DB, eventDeptWhere string) *gorm.DB {
	return db.Model(&model.Event{}).Where("`event_status` = 1").Where(eventDeptWhere)
}

type homePageConfig struct {
	VouchLimit int `json:"vouch_limit"`
	NewLimit   int `json:"new_limit"`
	HotLimit   int `json:"hot_limit"`
}

type eventObj struct {
	Cover []string `json:"cover"`
	Desc  string   `json:"desc"`
	Rules string   `json:"rules"`
}

type ListResponse struct {
	NewList   []ListItem `json:"newList"`
	HotList   []ListItem `json:"hotList"`
	VouchList []ListItem `json:"vouchList"`
}

type ListItem struct {
	ID         uint    `json:"id"`
	Img        string  `json:"img"`
	Title      string  `json:"title"`
	UserCnt    int     `json:"userCnt"`
	CateID     *string `json:"cateId,omitempty"`
	CateName   string  `json:"cateName"`
	Type       *int    `json:"type,omitempty"`
	CreateTime *int64  `json:"_createTime,omitempty"`
	Kind       string  `json:"kind"`
}

func loadHomePageConfig(ctx context.Context) homePageConfig {
	cfg := homePageConfig{VouchLimit: 10, NewLimit: 10, HotLimit: 10}
	setup, err := setupservice.GetSetupContext(ctx, "HOME_PAGE_CONFIG")
	if err != nil {
		return cfg
	}
	json.Unmarshal([]byte(setup.Value), &cfg)
	if cfg.VouchLimit <= 0 {
		cfg.VouchLimit = 10
	}
	if cfg.NewLimit <= 0 {
		cfg.NewLimit = 10
	}
	if cfg.HotLimit <= 0 {
		cfg.HotLimit = 10
	}
	return cfg
}

func populateEventFields(list []model.Event) []model.Event {
	for i := range list {
		var obj eventObj
		if list[i].Obj != "" {
			json.Unmarshal([]byte(list[i].Obj), &obj)
		}
		if len(obj.Cover) > 0 {
			list[i].Img = media.FullURLWithStaticDomain(obj.Cover[0])
		}
		list[i].Desc = obj.Desc
		list[i].Rules = obj.Rules
	}
	return list
}

func GetHomeList(userID string) (ListResponse, error) {
	return GetHomeListContext(context.Background(), userID)
}

func GetHomeListContext(ctx context.Context, userID string) (ListResponse, error) {
	db, cancel := database.WithContext(ctx)
	defer cancel()

	cfg := loadHomePageConfig(ctx)

	deptWhere := "(`enroll_publish_dept_ids` = '' OR `enroll_publish_dept_ids` IS NULL)"
	eventDeptWhere := "(`event_publish_dept_ids` = '' OR `event_publish_dept_ids` IS NULL)"
	if userID != "" {
		deptIDs := dept.UserDeptIDsByMiniOpenIDContext(ctx, userID)
		if len(deptIDs) > 0 {
			overlap := publish.DeptOverlap("enroll_publish_dept_ids", deptIDs)
			deptWhere = "(`enroll_publish_dept_ids` = '' OR `enroll_publish_dept_ids` IS NULL OR " + overlap + ")"
			eventOverlap := publish.DeptOverlap("event_publish_dept_ids", deptIDs)
			eventDeptWhere = "(`event_publish_dept_ids` = '' OR `event_publish_dept_ids` IS NULL OR " + eventOverlap + ")"
		}
	}

	var enrollVouch []model.Enroll
	if err := enrollBaseQuery(db, deptWhere).Where("`enroll_vouch` = 1").Order("`enroll_order` ASC, `add_time` DESC").Limit(cfg.VouchLimit).Find(&enrollVouch).Error; err != nil {
		return ListResponse{}, err
	}
	enrollVouch = populateEnrollFields(enrollVouch)

	var eventVouch []model.Event
	if err := eventBaseQuery(db, eventDeptWhere).Where("`event_vouch` = 1").Order("`event_order` ASC, `add_time` DESC").Limit(cfg.VouchLimit).Find(&eventVouch).Error; err != nil {
		return ListResponse{}, err
	}
	eventVouch = populateEventFields(eventVouch)

	vouchList := make([]ListItem, 0, len(enrollVouch)+len(eventVouch))
	for _, e := range enrollVouch {
		vouchList = append(vouchList, enrollListItem(e, true, false))
	}
	for _, e := range eventVouch {
		vouchList = append(vouchList, eventListItem(e, true, false))
	}

	var enrollNew []model.Enroll
	if err := enrollBaseQuery(db, deptWhere).Order("`add_time` DESC").Limit(cfg.NewLimit).Find(&enrollNew).Error; err != nil {
		return ListResponse{}, err
	}
	enrollNew = populateEnrollFields(enrollNew)

	var eventNew []model.Event
	if err := eventBaseQuery(db, eventDeptWhere).Order("`add_time` DESC").Limit(cfg.NewLimit).Find(&eventNew).Error; err != nil {
		return ListResponse{}, err
	}
	eventNew = populateEventFields(eventNew)

	type newItem struct {
		id      uint
		addTime int64
		data    ListItem
	}
	var newItems []newItem
	for _, e := range enrollNew {
		newItems = append(newItems, newItem{e.ID, e.AddTime, enrollListItem(e, false, true)})
	}
	for _, e := range eventNew {
		newItems = append(newItems, newItem{e.ID, e.AddTime, eventListItem(e, false, true)})
	}
	sort.Slice(newItems, func(i, j int) bool { return newItems[i].addTime > newItems[j].addTime })
	if len(newItems) > cfg.NewLimit {
		newItems = newItems[:cfg.NewLimit]
	}
	newList := make([]ListItem, len(newItems))
	for i, item := range newItems {
		newList[i] = item.data
	}

	var enrollHot []model.Enroll
	if err := enrollBaseQuery(db, deptWhere).Order("`enroll_join_cnt` DESC, `add_time` DESC").Limit(cfg.HotLimit).Find(&enrollHot).Error; err != nil {
		return ListResponse{}, err
	}
	enrollHot = populateEnrollFields(enrollHot)

	var eventHot []model.Event
	if err := eventBaseQuery(db, eventDeptWhere).Order("`event_user_cnt` DESC, `add_time` DESC").Limit(cfg.HotLimit).Find(&eventHot).Error; err != nil {
		return ListResponse{}, err
	}
	eventHot = populateEventFields(eventHot)

	type hotItem struct {
		cnt     int
		addTime int64
		data    ListItem
	}
	var hotItems []hotItem
	for _, e := range enrollHot {
		hotItems = append(hotItems, hotItem{e.JoinCnt, e.AddTime, enrollListItem(e, false, true)})
	}
	for _, e := range eventHot {
		hotItems = append(hotItems, hotItem{e.JoinCnt, e.AddTime, eventListItem(e, false, true)})
	}
	sort.Slice(hotItems, func(i, j int) bool {
		if hotItems[i].cnt != hotItems[j].cnt {
			return hotItems[i].cnt > hotItems[j].cnt
		}
		return hotItems[i].addTime > hotItems[j].addTime
	})
	if len(hotItems) > cfg.HotLimit {
		hotItems = hotItems[:cfg.HotLimit]
	}
	hotList := make([]ListItem, len(hotItems))
	for i, item := range hotItems {
		hotList[i] = item.data
	}

	return ListResponse{
		NewList:   newList,
		HotList:   hotList,
		VouchList: vouchList,
	}, nil
}

func enrollListItem(e model.Enroll, includeCateID, includeCreateTime bool) ListItem {
	item := ListItem{
		ID:       e.ID,
		Img:      e.Img,
		Title:    e.Title,
		UserCnt:  e.UserCnt,
		CateName: e.CateName,
		Kind:     "enroll",
	}
	if includeCateID {
		item.CateID = &e.CateID
	}
	if includeCreateTime {
		item.CreateTime = &e.AddTime
	}
	return item
}

func eventListItem(e model.Event, includeType, includeCreateTime bool) ListItem {
	item := ListItem{
		ID:       e.ID,
		Img:      e.Img,
		Title:    e.Title,
		UserCnt:  e.UserCnt,
		CateName: e.CateName,
		Kind:     eventKind(e),
	}
	if includeType {
		item.Type = &e.Type
	}
	if includeCreateTime {
		item.CreateTime = &e.AddTime
	}
	return item
}

func eventKind(e model.Event) string {
	if e.Type == 2 {
		return "competition"
	}
	return "activity"
}
