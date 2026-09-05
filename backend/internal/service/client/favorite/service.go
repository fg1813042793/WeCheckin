package favorite

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"

	"gorm.io/gorm"
	"wecheckin/backend/internal/model"
	"wecheckin/backend/internal/support/media"
	"wecheckin/backend/pkg/database"
)

type enrollObj struct {
	Cover []string `json:"cover"`
	Desc  string   `json:"desc"`
}

type ListItem struct {
	ID           string `json:"id"`
	Title        string `json:"title"`
	CreateTime   int64  `json:"_createTime"`
	JoinCount    *int   `json:"joinCount,omitempty"`
	CheckinCount *int   `json:"checkinCount,omitempty"`
	Img          string `json:"img,omitempty"`
	Desc         string `json:"desc,omitempty"`
}

func IsFav(userID, oid string) (bool, error) {
	return IsFavContext(context.Background(), userID, oid)
}

func IsFavContext(ctx context.Context, userID, oid string) (bool, error) {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	var cnt int64
	err := db.Model(&model.Favorite{}).Where("`fav_user_id` = ? AND `fav_oid` = ?", userID, oid).Count(&cnt).Error
	if err != nil {
		return false, err
	}
	return cnt > 0, nil
}

func UpdateFav(userID, title, typ, oid, path, addIP string) error {
	return UpdateFavContext(context.Background(), userID, title, typ, oid, path, addIP)
}

func UpdateFavContext(ctx context.Context, userID, title, typ, oid, path, addIP string) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	var fav model.Favorite
	result := db.Where("`fav_user_id` = ? AND `fav_oid` = ?", userID, oid).First(&fav)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		fav = model.Favorite{
			UserID:  userID,
			Title:   title,
			Type:    typ,
			OID:     oid,
			Path:    path,
			AddTime: database.Now(),
			AddIP:   addIP,
		}
		return db.Create(&fav).Error
	}
	if result.Error != nil {
		return result.Error
	}
	return db.Model(&fav).Updates(map[string]interface{}{
		"fav_title":     title,
		"fav_type":      typ,
		"fav_oid":       oid,
		"fav_path":      path,
		"fav_edit_time": database.Now(),
		"fav_edit_ip":   addIP,
	}).Error
}

func DelFav(userID, oid string) error {
	return DelFavContext(context.Background(), userID, oid)
}

func DelFavContext(ctx context.Context, userID, oid string) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	return db.Where("`fav_user_id` = ? AND `fav_oid` = ?", userID, oid).Delete(&model.Favorite{}).Error
}

func GetMyFavList(userID string) ([]ListItem, error) {
	return GetMyFavListContext(context.Background(), userID)
}

func GetMyFavListContext(ctx context.Context, userID string) ([]ListItem, error) {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	var favs []model.Favorite
	err := db.Where("`fav_user_id` = ?", userID).Order("`fav_add_time` DESC").Find(&favs).Error
	if err != nil {
		return nil, err
	}
	enrollByID, err := loadFavoriteEnrollsByID(db, favs)
	if err != nil {
		return nil, err
	}
	result := make([]ListItem, 0, len(favs))
	for _, f := range favs {
		item := ListItem{
			ID:         f.OID,
			Title:      f.Title,
			CreateTime: f.AddTime,
		}
		if enroll, ok := enrollByID[f.OID]; ok {
			enrichFavoriteItemWithEnroll(&item, enroll)
		}
		result = append(result, item)
	}
	return result, nil
}

func loadFavoriteEnrollsByID(db *gorm.DB, favs []model.Favorite) (map[string]model.Enroll, error) {
	oids := favoriteOIDs(favs)
	enrollByID := make(map[string]model.Enroll, len(oids))
	if len(oids) == 0 {
		return enrollByID, nil
	}

	var enrolls []model.Enroll
	if err := db.Select("id", "enroll_user_cnt", "enroll_join_cnt", "enroll_obj").
		Where("`id` IN ?", oids).
		Find(&enrolls).Error; err != nil {
		return nil, err
	}
	for _, enroll := range enrolls {
		enrollByID[strconv.FormatUint(uint64(enroll.ID), 10)] = enroll
	}
	return enrollByID, nil
}

func favoriteOIDs(favs []model.Favorite) []string {
	seen := make(map[string]struct{}, len(favs))
	oids := make([]string, 0, len(favs))
	for _, fav := range favs {
		if fav.OID == "" {
			continue
		}
		if _, ok := seen[fav.OID]; ok {
			continue
		}
		seen[fav.OID] = struct{}{}
		oids = append(oids, fav.OID)
	}
	return oids
}

func enrichFavoriteItemWithEnroll(item *ListItem, enroll model.Enroll) {
	joinCount := enroll.UserCnt
	checkinCount := enroll.JoinCnt
	item.JoinCount = &joinCount
	item.CheckinCount = &checkinCount
	if enroll.Obj == "" {
		return
	}

	var obj enrollObj
	if json.Unmarshal([]byte(enroll.Obj), &obj) != nil {
		return
	}
	if len(obj.Cover) > 0 {
		item.Img = media.FullURLWithStaticDomain(obj.Cover[0])
	}
	item.Desc = obj.Desc
}
