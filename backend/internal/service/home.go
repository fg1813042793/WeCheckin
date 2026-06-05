package service

import (
	"encoding/json"
	"sort"
	"strings"
	"wecheckin-backend/backend/internal/database"
	"wecheckin-backend/backend/internal/model"
	"gorm.io/gorm"
)

func GetSetup(key string) (*model.Setup, error) {
	var setup model.Setup
	err := database.DB.Where("`setup_key` = ?", key).First(&setup).Error
	if err != nil {
		return nil, err
	}
	return &setup, nil
}

func GetStaticDomain() string {
	var setup model.Setup
	if err := database.DB.Where("`setup_key` = ?", "STATIC_DOMAIN").First(&setup).Error; err != nil {
		return "http://localhost:8080"
	}
	domain := setup.Value
	if domain == "" {
		return "http://localhost:8080"
	}
	return domain
}

func GetFullURL(path string) string {
	if path == "" {
		return ""
	}
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return path
	}
	domain := GetStaticDomain()
	return domain + path
}

type enrollObj struct {
	Cover []string `json:"cover"`
	Desc  string   `json:"desc"`
}

func populateEnrollFields(list []model.Enroll) []model.Enroll {
	for i := range list {
		var obj enrollObj
		if list[i].Obj != "" {
			json.Unmarshal([]byte(list[i].Obj), &obj)
		}
		if len(obj.Cover) > 0 {
			list[i].Img = GetFullURL(obj.Cover[0])
		}
		list[i].Desc = obj.Desc

		if list[i].UserList != "" {
			var arr []map[string]interface{}
			if err := json.Unmarshal([]byte(list[i].UserList), &arr); err == nil {
				list[i].UserListArr = arr
			}
		}
	}
	return list
}

func enrollBaseQuery(deptWhere string) *gorm.DB {
	return database.DB.Model(&model.Enroll{}).Where("`enroll_status` = 1").Where(deptWhere)
}

func eventBaseQuery(eventDeptWhere string) *gorm.DB {
	return database.DB.Model(&model.Event{}).Where("`event_status` = 1").Where(eventDeptWhere)
}

type homePageConfig struct {
	VouchLimit int `json:"vouch_limit"`
	NewLimit   int `json:"new_limit"`
	HotLimit   int `json:"hot_limit"`
}

func loadHomePageConfig() homePageConfig {
	cfg := homePageConfig{VouchLimit: 10, NewLimit: 10, HotLimit: 10}
	var setup model.Setup
	if err := database.DB.Where("`setup_key` = ?", "HOME_PAGE_CONFIG").First(&setup).Error; err != nil {
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

func GetHomeList(userID string) (map[string]interface{}, error) {
	cfg := loadHomePageConfig()

	deptWhere := "(`enroll_publish_dept_ids` = '' OR `enroll_publish_dept_ids` IS NULL)"
	eventDeptWhere := "(`event_publish_dept_ids` = '' OR `event_publish_dept_ids` IS NULL)"
	if userID != "" {
		deptIDs := getUserDeptIDsByMiniOpenID(userID)
		if len(deptIDs) > 0 {
			overlap := buildDeptOverlap("enroll_publish_dept_ids", deptIDs)
			deptWhere = "(`enroll_publish_dept_ids` = '' OR `enroll_publish_dept_ids` IS NULL OR " + overlap + ")"
			eventOverlap := buildDeptOverlap("event_publish_dept_ids", deptIDs)
			eventDeptWhere = "(`event_publish_dept_ids` = '' OR `event_publish_dept_ids` IS NULL OR " + eventOverlap + ")"
		}
	}

	// --- Vouch (recommended) ---
	var enrollVouch []model.Enroll
	enrollBaseQuery(deptWhere).Where("`enroll_vouch` = 1").Order("`enroll_order` ASC, `enroll_add_time` DESC").Limit(cfg.VouchLimit).Find(&enrollVouch)
	enrollVouch = populateEnrollFields(enrollVouch)

	var eventVouch []model.Event
	eventBaseQuery(eventDeptWhere).Where("`event_vouch` = 1").Order("`event_order` ASC, `event_add_time` DESC").Limit(cfg.VouchLimit).Find(&eventVouch)
	eventVouch = populateEventFields(eventVouch)

	vouchList := make([]map[string]interface{}, 0, len(enrollVouch)+len(eventVouch))
	for _, e := range enrollVouch {
		vouchList = append(vouchList, map[string]interface{}{
			"id": e.ID, "img": e.Img, "title": e.Title, "userCnt": e.UserCnt,
			"cateId": e.CateID, "cateName": e.CateName, "kind": "enroll",
		})
	}
	for _, e := range eventVouch {
		kind := "activity"
		if e.Type == 2 {
			kind = "competition"
		}
		vouchList = append(vouchList, map[string]interface{}{
			"id": e.ID, "img": e.Img, "title": e.Title, "userCnt": e.UserCnt,
			"type": e.Type, "cateName": e.CateName, "kind": kind,
		})
	}

	// --- New (latest) ---
	var enrollNew []model.Enroll
	enrollBaseQuery(deptWhere).Order("`enroll_add_time` DESC").Limit(cfg.NewLimit).Find(&enrollNew)
	enrollNew = populateEnrollFields(enrollNew)

	var eventNew []model.Event
	eventBaseQuery(eventDeptWhere).Order("`event_add_time` DESC").Limit(cfg.NewLimit).Find(&eventNew)
	eventNew = populateEventFields(eventNew)

	type newItem struct {
		id      uint
		addTime int64
		data    map[string]interface{}
	}
	var newItems []newItem
	for _, e := range enrollNew {
		newItems = append(newItems, newItem{e.ID, e.AddTime, map[string]interface{}{
			"id": e.ID, "img": e.Img, "title": e.Title, "userCnt": e.UserCnt,
			"cateName": e.CateName, "_createTime": e.AddTime, "kind": "enroll",
		}})
	}
	for _, e := range eventNew {
		kind := "activity"
		if e.Type == 2 {
			kind = "competition"
		}
		newItems = append(newItems, newItem{e.ID, e.AddTime, map[string]interface{}{
			"id": e.ID, "img": e.Img, "title": e.Title, "userCnt": e.UserCnt,
			"cateName": e.CateName, "_createTime": e.AddTime, "kind": kind,
		}})
	}
	sort.Slice(newItems, func(i, j int) bool { return newItems[i].addTime > newItems[j].addTime })
	if len(newItems) > cfg.NewLimit {
		newItems = newItems[:cfg.NewLimit]
	}
	newList := make([]map[string]interface{}, len(newItems))
	for i, item := range newItems {
		newList[i] = item.data
	}

	// --- Hot (most participants) ---
	var enrollHot []model.Enroll
	enrollBaseQuery(deptWhere).Order("`enroll_join_cnt` DESC, `enroll_add_time` DESC").Limit(cfg.HotLimit).Find(&enrollHot)
	enrollHot = populateEnrollFields(enrollHot)

	var eventHot []model.Event
	eventBaseQuery(eventDeptWhere).Order("`event_user_cnt` DESC, `event_add_time` DESC").Limit(cfg.HotLimit).Find(&eventHot)
	eventHot = populateEventFields(eventHot)

	type hotItem struct {
		cnt     int
		addTime int64
		data    map[string]interface{}
	}
	var hotItems []hotItem
	for _, e := range enrollHot {
		hotItems = append(hotItems, hotItem{e.JoinCnt, e.AddTime, map[string]interface{}{
			"id": e.ID, "img": e.Img, "title": e.Title, "userCnt": e.UserCnt,
			"cateName": e.CateName, "_createTime": e.AddTime, "kind": "enroll",
		}})
	}
	for _, e := range eventHot {
		kind := "activity"
		if e.Type == 2 {
			kind = "competition"
		}
		hotItems = append(hotItems, hotItem{e.JoinCnt, e.AddTime, map[string]interface{}{
			"id": e.ID, "img": e.Img, "title": e.Title, "userCnt": e.UserCnt,
			"cateName": e.CateName, "_createTime": e.AddTime, "kind": kind,
		}})
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
	hotList := make([]map[string]interface{}, len(hotItems))
	for i, item := range hotItems {
		hotList[i] = item.data
	}

	result := map[string]interface{}{
		"newList":   newList,
		"hotList":   hotList,
		"vouchList": vouchList,
	}
	return result, nil
}
