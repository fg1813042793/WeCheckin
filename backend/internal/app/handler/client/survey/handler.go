package survey

import (
	"github.com/cloudwego/hertz/pkg/app"

	surveyservice "wecheckin/backend/internal/app/service/survey"
)

type ClientSurveyHandler struct {
	survey    *surveyservice.SurveyService
	responses *surveyservice.ResponseService
}

func getUID(c *app.RequestContext) uint {
	uidVal, _ := c.Get("user_id")
	if uidVal == nil {
		return 0
	}
	switch v := uidVal.(type) {
	case uint:
		return v
	case int64:
		return uint(v)
	case float64:
		return uint(v)
	}
	return 0
}

func NewClientSurveyHandler() *ClientSurveyHandler { return &ClientSurveyHandler{} }

func (h *ClientSurveyHandler) lazyInit() {
	if h.survey == nil {
		h.survey = surveyservice.NewSurveyService()
	}
	if h.responses == nil {
		h.responses = surveyservice.NewResponseService()
	}
}
