package service

import (
	"wecheckin-backend/backend/internal/database"
	"wecheckin-backend/backend/internal/model"
)

func GetNewsList() ([]model.News, error) {
	var newsList []model.News
	err := database.DB.Where("`NEWS_STATUS` = 1").Order("`NEWS_ORDER` ASC, `NEWS_ADD_TIME` DESC").Find(&newsList).Error
	if err != nil {
		return nil, err
	}
	return newsList, nil
}

func ViewNews(id string) (*model.News, error) {
	var news model.News
	err := database.DB.Where("`NEWS_STATUS` = 1 AND `id` = ?", id).First(&news).Error
	if err != nil {
		return nil, err
	}
	database.DB.Model(&news).UpdateColumn("NEWS_VIEW_CNT", news.ViewCnt+1)
	return &news, nil
}
