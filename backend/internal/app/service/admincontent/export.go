package admincontent

import (
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
	var joins []model.EnrollJoin
	queryBuilder := database.DB.Where("`enroll_join_enroll_id` = ?", enrollID)
	if startDay != "" {
		queryBuilder = queryBuilder.Where("`enroll_join_day` >= ?", startDay)
	}
	if endDay != "" {
		queryBuilder = queryBuilder.Where("`enroll_join_day` <= ?", endDay)
	}
	queryBuilder.Order("`enroll_join_add_time` DESC").Find(&joins)

	var enroll model.Enroll
	database.DB.Where("`id` = ?", enrollID).First(&enroll)

	userNames := map[string]string{}
	var users []model.User
	database.DB.Find(&users)
	for _, u := range users {
		userNames[u.MiniOpenID] = u.Name
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
