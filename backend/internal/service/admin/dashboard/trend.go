package dashboard

import "time"

type dashboardTrendTimestamps struct {
	Users           []int64
	EnrollJoins     []int64
	SurveyResponses []int64
	ExamRecords     []int64
}

func normalizeTrendDays(days int) int {
	if days == 30 {
		return 30
	}
	return 7
}

func buildTrendPoints(now time.Time, days int, timestamps dashboardTrendTimestamps) []DashboardTrendPoint {
	if days <= 0 {
		days = 7
	}
	start := startOfDay(now).AddDate(0, 0, -(days - 1))
	points := make([]DashboardTrendPoint, days)
	indexes := make(map[string]int, days)
	for index := 0; index < days; index++ {
		date := start.AddDate(0, 0, index).Format("2006-01-02")
		points[index].Date = date
		indexes[date] = index
	}
	addTrendTimestamps(points, indexes, now.Location(), timestamps.Users, func(point *DashboardTrendPoint) {
		point.UserCount++
	})
	addTrendTimestamps(points, indexes, now.Location(), timestamps.EnrollJoins, func(point *DashboardTrendPoint) {
		point.EnrollJoinCount++
	})
	addTrendTimestamps(points, indexes, now.Location(), timestamps.SurveyResponses, func(point *DashboardTrendPoint) {
		point.SurveyResponseCount++
	})
	addTrendTimestamps(points, indexes, now.Location(), timestamps.ExamRecords, func(point *DashboardTrendPoint) {
		point.ExamRecordCount++
	})
	return points
}

func addTrendTimestamps(
	points []DashboardTrendPoint,
	indexes map[string]int,
	location *time.Location,
	timestamps []int64,
	add func(*DashboardTrendPoint),
) {
	for _, timestamp := range timestamps {
		if timestamp <= 0 {
			continue
		}
		date := time.UnixMilli(timestamp).In(location).Format("2006-01-02")
		index, exists := indexes[date]
		if !exists {
			continue
		}
		add(&points[index])
	}
}

func startOfDay(value time.Time) time.Time {
	return time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, value.Location())
}
