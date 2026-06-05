package service

import (
	"encoding/json"
	"wecheckin-backend/backend/internal/database"
	"wecheckin-backend/backend/internal/model"
)

func IsFav(userID, oid string) (bool, error) {
	var cnt int64
	err := database.DB.Model(&model.Favorite{}).Where("`fav_user_id` = ? AND `fav_oid` = ?", userID, oid).Count(&cnt).Error
	if err != nil {
		return false, err
	}
	return cnt > 0, nil
}

func UpdateFav(userID, title, typ, oid, path, addIP string) error {
	var fav model.Favorite
	result := database.DB.Where("`fav_user_id` = ? AND `fav_oid` = ?", userID, oid).First(&fav)
	if result.Error != nil {
		fav = model.Favorite{
			UserID:  userID,
			Title:   title,
			Type:    typ,
			OID:     oid,
			Path:    path,
			AddTime: database.Now(),
			AddIP:   addIP,
		}
		return database.DB.Create(&fav).Error
	}
	return database.DB.Model(&fav).Updates(map[string]interface{}{
		"fav_title":     title,
		"fav_type":      typ,
		"fav_oid":       oid,
		"fav_path":      path,
		"fav_edit_time": database.Now(),
		"fav_edit_ip":   addIP,
	}).Error
}

func DelFav(userID, oid string) error {
	return database.DB.Where("`fav_user_id` = ? AND `fav_oid` = ?", userID, oid).Delete(&model.Favorite{}).Error
}

func GetMyFavList(userID string) ([]map[string]interface{}, error) {
	var favs []model.Favorite
	err := database.DB.Where("`fav_user_id` = ?", userID).Order("`fav_add_time` DESC").Find(&favs).Error
	if err != nil {
		return nil, err
	}
	var result []map[string]interface{}
	for _, f := range favs {
		item := map[string]interface{}{
			"id":          f.OID,
			"title":       f.Title,
			"_createTime": f.AddTime,
		}
		if f.OID != "" {
			var enroll model.Enroll
			if database.DB.Where("id = ?", f.OID).First(&enroll).Error == nil {
				item["joinCount"] = enroll.UserCnt
				item["checkinCount"] = enroll.JoinCnt
				if enroll.Obj != "" {
					var obj enrollObj
					if json.Unmarshal([]byte(enroll.Obj), &obj) == nil {
						if len(obj.Cover) > 0 {
							item["img"] = GetFullURL(obj.Cover[0])
						}
						item["desc"] = obj.Desc
					}
				}
			}
		}
		result = append(result, item)
	}
	return result, nil
}
