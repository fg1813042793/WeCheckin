package service

import (
	"wecheckin-backend/backend/internal/database"
	"wecheckin-backend/backend/internal/model"
)

func GetSetup(key string) (*model.Setup, error) {
	var setup model.Setup
	err := database.DB.Where("`SETUP_KEY` = ?", key).First(&setup).Error
	if err != nil {
		return nil, err
	}
	return &setup, nil
}

func GetHomeList() (map[string]interface{}, error) {
	where := "`ENROLL_STATUS` = 1"

	var newList []model.Enroll
	database.DB.Where(where).Order("`ENROLL_ADD_TIME` DESC").Limit(10).Find(&newList)

	var vouchList []model.Enroll
	database.DB.Where(where).Order("`ENROLL_VOUCH` DESC, `ENROLL_ADD_TIME` DESC").Limit(10).Find(&vouchList)

	var hotList []model.Enroll
	database.DB.Where(where).Order("`ENROLL_USER_CNT` DESC, `ENROLL_ADD_TIME` DESC").Limit(10).Find(&hotList)

	result := map[string]interface{}{
		"banners":   []map[string]string{},
		"newList":   newList,
		"hotList":   hotList,
		"vouchList": vouchList,
	}
	return result, nil
}
