package survey

// ResponseService 答卷服务
type ResponseService struct {
	Survey *SurveyService
}

func NewResponseService() *ResponseService {
	return &ResponseService{Survey: NewSurveyService()}
}
