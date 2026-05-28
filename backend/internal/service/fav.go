package service

import (
	"wecheckin-backend/backend/internal/database"
	"wecheckin-backend/backend/internal/model"
)

func IsFav(userID, oid string) (bool, error) {
	var cnt int64
	err := database.DB.Model(&model.Favorite{}).Where("`FAV_USER_ID` = ? AND `FAV_OID` = ?", userID, oid).Count(&cnt).Error
	if err != nil {
		return false, err
	}
	return cnt > 0, nil
}

func UpdateFav(userID, title, typ, oid, path, addIP string) error {
	var fav model.Favorite
	result := database.DB.Where("`FAV_USER_ID` = ? AND `FAV_OID` = ?", userID, oid).First(&fav)
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
		"FAV_TITLE":     title,
		"FAV_TYPE":      typ,
		"FAV_OID":       oid,
		"FAV_PATH":      path,
		"FAV_EDIT_TIME": database.Now(),
		"FAV_EDIT_IP":   addIP,
	}).Error
}

func DelFav(userID, oid string) error {
	return database.DB.Where("`FAV_USER_ID` = ? AND `FAV_OID` = ?", userID, oid).Delete(&model.Favorite{}).Error
}

func GetMyFavList(userID string) ([]model.Favorite, error) {
	var list []model.Favorite
	err := database.DB.Where("`FAV_USER_ID` = ?", userID).Order("`FAV_ADD_TIME` DESC").Find(&list).Error
	if err != nil {
		return nil, err
	}
	return list, nil
}
