package admincontent

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"wecheckin-backend/backend/internal/model"
	"wecheckin-backend/backend/pkg/database"
)

func GetEnrollJoinDataURL(enrollID string) (string, error) {
	return "", nil
}

func DeleteEnrollJoinDataExcel(enrollID string) error {
	filename := fmt.Sprintf("export_enroll_%s.csv", enrollID)
	os.Remove(filepath.Join("./uploads", filename))
	return nil
}

func ExportEnrollJoinDataExcel(enrollID, startDay, endDay string) (string, error) {
	return ExportEnrollJoinDataExcelContext(context.Background(), enrollID, startDay, endDay)
}

func ExportEnrollJoinDataExcelContext(ctx context.Context, enrollID, startDay, endDay string) (string, error) {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	var joins []model.EnrollJoin
	queryBuilder := db.Where("`enroll_join_enroll_id` = ?", enrollID)
	if startDay != "" {
		queryBuilder = queryBuilder.Where("`enroll_join_day` >= ?", startDay)
	}
	if endDay != "" {
		queryBuilder = queryBuilder.Where("`enroll_join_day` <= ?", endDay)
	}
	if err := queryBuilder.Order("`enroll_join_add_time` DESC").Find(&joins).Error; err != nil {
		return "", err
	}

	var enroll model.Enroll
	if err := db.Where("`id` = ?", enrollID).First(&enroll).Error; err != nil {
		return "", err
	}

	userIDs := make([]string, 0, len(joins))
	for _, join := range joins {
		userIDs = append(userIDs, join.UserID)
	}
	var users []model.User
	if len(userIDs) > 0 {
		if err := db.Select("user_mini_openid", "user_name").
			Where("`user_mini_openid` IN ?", uniqueNonEmptyStrings(userIDs)).
			Find(&users).Error; err != nil {
			return "", err
		}
	}
	userNames := make(map[string]string, len(users))
	for _, user := range users {
		userNames[user.MiniOpenID] = user.Name
	}

	filename := fmt.Sprintf("export_enroll_%s.csv", enrollID)
	filePath := filepath.Join("./uploads", filename)

	f, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	writer := csv.NewWriter(f)
	defer writer.Flush()

	f.WriteString("\xEF\xBB\xBF")

	writer.Write([]string{"打卡项目", enroll.Title})
	writer.Write([]string{"用户ID", "用户姓名", "打卡日期", "打卡内容", "打卡时间", "IP地址"})
	for _, j := range joins {
		joinTime := time.UnixMilli(j.AddTime).Format("2006-01-02 15:04:05")
		writer.Write([]string{
			j.UserID,
			userNames[j.UserID],
			j.Day,
			j.Forms,
			joinTime,
			j.AddIP,
		})
	}

	return filename, nil
}

func GetUserDataURL() (string, error) {
	return "", nil
}

func DeleteUserDataExcel() error {
	return nil
}

func ExportUserDataExcel() (string, error) {
	return "", nil
}
