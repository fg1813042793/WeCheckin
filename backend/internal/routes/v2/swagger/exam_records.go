package swagger

import "wecheckin/backend/internal/model/assessment"

// ExamRecordListResponse contains stored exam records with client labels derived from User-Agent.
type ExamRecordListResponse struct {
	List  []assessment.ExamRecord `json:"list"`
	Total int64                   `json:"total"`
}
