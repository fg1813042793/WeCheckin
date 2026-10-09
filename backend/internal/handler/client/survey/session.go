package survey

import "time"

const surveySessionMaxAge = 12 * time.Hour

func shouldResetSurveySession(now, startAt int64) bool {
	if now <= 0 || startAt <= 0 || startAt > now {
		return true
	}
	return time.Duration(now-startAt)*time.Millisecond > surveySessionMaxAge
}
