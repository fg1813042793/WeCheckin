package enroll

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"wecheckin/backend/internal/model"
	"wecheckin/backend/internal/support/media"
	"wecheckin/backend/pkg/database"
)

func ViewEnroll(id, userID string) (*model.Enroll, error) {
	return ViewEnrollContext(context.Background(), id, userID)
}

func ViewEnrollContext(ctx context.Context, id, userID string) (*model.Enroll, error) {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	var enroll model.Enroll
	err := db.Where("`id` = ?", id).First(&enroll).Error
	if err != nil {
		return nil, err
	}
	db.Model(&enroll).UpdateColumn("enroll_view_cnt", enroll.ViewCnt+1)

	// Check if current user has joined
	if userID != "" {
		var euCnt int64
		db.Model(&model.EnrollUser{}).Where("`enroll_user_enroll_id` = ? AND `enroll_user_mini_openid` = ?", id, userID).Count(&euCnt)
		if euCnt > 0 {
			enroll.IsJoin = true
		}
		var jCnt int64
		db.Model(&model.EnrollJoin{}).Where("`enroll_join_enroll_id` = ? AND `enroll_join_user_id` = ? AND `enroll_join_day` = ?", id, userID, time.Now().Format("2006-01-02")).Count(&jCnt)
		if jCnt > 0 {
			enroll.MyEnrollJoinID = "1"
		}
	}

	// Parse OBJ for img/desc/content
	var objMap map[string]interface{}
	if enroll.Obj != "" {
		json.Unmarshal([]byte(enroll.Obj), &objMap)
	}
	if objMap != nil {
		if covers, ok := objMap["cover"].([]interface{}); ok && len(covers) > 0 {
			enroll.Img = media.FullURLWithStaticDomain(fmt.Sprintf("%v", covers[0]))
		}
		if desc, ok := objMap["desc"].(string); ok {
			enroll.Desc = desc
		}
		if c, ok := objMap["content"].([]interface{}); ok {
			for _, item := range c {
				if m, ok := item.(map[string]interface{}); ok {
					entry := map[string]string{}
					if t, ok := m["type"].(string); ok {
						entry["type"] = t
					}
					if v, ok := m["val"].(string); ok {
						entry["val"] = v
					}
					enroll.Content = append(enroll.Content, entry)
				}
			}
		}
	}

	// Format start/end
	if enroll.Start > 0 {
		enroll.StartStr = time.UnixMilli(enroll.Start).Format("2006-01-02")
	} else {
		enroll.StartStr = "-"
	}
	if enroll.End > 0 {
		enroll.EndStr = time.UnixMilli(enroll.End).Format("2006-01-02")
	} else {
		enroll.EndStr = "-"
	}

	// Status
	now := time.Now().UnixMilli()
	if enroll.Status == 0 {
		enroll.StatusDesc = "已停用"
	} else if enroll.End > 0 && now > enroll.End {
		enroll.StatusDesc = "已结束"
	} else if enroll.Start > 0 && now < enroll.Start {
		enroll.StatusDesc = "未开始"
	} else {
		enroll.StatusDesc = "进行中"
	}

	// DayList from join records
	var days []string
	db.Model(&model.EnrollJoin{}).
		Where("`enroll_join_enroll_id` = ?", id).
		Select("DISTINCT `enroll_join_day`").
		Order("`enroll_join_day` ASC").
		Pluck("enroll_join_day", &days)
	for _, d := range days {
		t, err := time.Parse("2006-01-02", d)
		if err != nil {
			continue
		}
		enroll.DayList = append(enroll.DayList, map[string]string{
			"day":   d,
			"month": fmt.Sprintf("%d月", t.Month()),
			"date":  fmt.Sprintf("%d", t.Day()),
		})
	}

	// RankList from enroll_users
	var enrollUsers []model.EnrollUser
	if err := db.Where("`enroll_user_enroll_id` = ?", id).
		Order("`enroll_user_join_cnt` DESC, `enroll_user_day_cnt` DESC").
		Find(&enrollUsers).Error; err != nil {
		return nil, err
	}

	userMap := map[string]model.User{}
	userOpenIDs := rankUserOpenIDs(enrollUsers)
	if len(userOpenIDs) > 0 {
		var users []model.User
		if err := db.Select("user_mini_openid", "user_name", "user_pic").
			Where("`user_mini_openid` IN ?", userOpenIDs).
			Find(&users).Error; err != nil {
			return nil, err
		}
		for _, u := range users {
			userMap[u.MiniOpenID] = u
		}
	}

	for _, eu := range enrollUsers {
		u, ok := userMap[eu.MiniOpenID]
		name := eu.MiniOpenID
		avatar := ""
		if ok {
			name = u.Name
			avatar = media.FullURLWithStaticDomain(u.Pic)
		}
		enroll.RankList = append(enroll.RankList, map[string]interface{}{
			"userName":   name,
			"userAvatar": avatar,
			"name":       name,
			"avatar":     avatar,
			"joinCount":  eu.JoinCnt,
			"lastDay":    eu.LastDay,
		})
	}

	return &enroll, nil
}

func rankUserOpenIDs(enrollUsers []model.EnrollUser) []string {
	seen := make(map[string]struct{}, len(enrollUsers))
	openIDs := make([]string, 0, len(enrollUsers))
	for _, enrollUser := range enrollUsers {
		if enrollUser.MiniOpenID == "" {
			continue
		}
		if _, ok := seen[enrollUser.MiniOpenID]; ok {
			continue
		}
		seen[enrollUser.MiniOpenID] = struct{}{}
		openIDs = append(openIDs, enrollUser.MiniOpenID)
	}
	return openIDs
}
