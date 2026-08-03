package dingtalkh5

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"

	"gorm.io/gorm"

	"wecheckin/backend/internal/model"
	"wecheckin/backend/pkg/database"
)

func nextStatusAfterSelfSubmit(review DingTalkH5Review) string {
	if strings.TrimSpace(review.ManagerID) == "" {
		return ReviewStatusHRFinal
	}
	return ReviewStatusManagerReview
}

func notDeletedReviewQuery(db *gorm.DB) *gorm.DB {
	if db == nil {
		return db
	}
	return db.Where("`deleted_at` = 0")
}

func findVisibleReview(ctx context.Context, user *model.DingTalkH5PerfUser, reviewNo string) (*model.DingTalkH5PerfReview, error) {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	var review model.DingTalkH5PerfReview
	query := notDeletedReviewQuery(db).Where("review_no = ?", strings.TrimSpace(reviewNo))
	var err error
	query, err = applyReviewVisibilityScopeContext(ctx, db, query, user, reviewScopeSummary)
	if err != nil {
		return nil, err
	}
	if err := query.First(&review).Error; err != nil {
		return nil, fmt.Errorf("没有找到这张考评单")
	}
	return &review, nil
}

func mutateReview(ctx context.Context, user *model.DingTalkH5PerfUser, reviewNo string, action func(*gorm.DB, *model.DingTalkH5PerfReview) error) (*ReviewDTO, error) {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	var output model.DingTalkH5PerfReview
	err := db.Transaction(func(tx *gorm.DB) error {
		if err := notDeletedReviewQuery(tx).Where("review_no = ?", strings.TrimSpace(reviewNo)).First(&output).Error; err != nil {
			return fmt.Errorf("没有找到这张考评单")
		}
		visible, err := reviewInDataScopeContext(ctx, tx, user, output)
		if err != nil {
			return err
		}
		if !visible {
			return fmt.Errorf("没有找到这张考评单")
		}
		if err := action(tx, &output); err != nil {
			return err
		}
		return touchReviewUpdateAudit(ctx, tx, user, &output)
	})
	if err != nil {
		return nil, err
	}
	histories, err := historiesForReview(ctx, output.ID)
	if err != nil {
		return nil, err
	}
	dto := reviewDTO(output, histories)
	hydrateReviewDTOValues(&dto, loadReviewValueTemplatesContext(ctx))
	return &dto, nil
}

func addHistoryWithDB(db *gorm.DB, review *model.DingTalkH5PerfReview, user *model.DingTalkH5PerfUser, action string) error {
	byAccount := "system"
	byName := "system"
	if user != nil {
		byAccount = user.Account
		byName = user.Name
	}
	now := database.Now()
	history := model.DingTalkH5PerfHistory{
		ReviewID:  review.ID,
		ReviewNo:  review.ReviewNo,
		ByAccount: byAccount,
		ByName:    byName,
		Action:    action,
		AddTime:   now,
		EditTime:  now,
	}
	applyDingTalkH5CreateAudit(&history.DingTalkH5AuditFields, dingtalkH5AuditMetaForUserContext(dingtalkH5DBContext(db), db, user, now))
	return db.Create(&history).Error
}

func touchReviewUpdateAudit(ctx context.Context, db *gorm.DB, user *model.DingTalkH5PerfUser, review *model.DingTalkH5PerfReview) error {
	if db == nil || review == nil || review.ID == 0 {
		return nil
	}
	audit := dingtalkH5AuditMetaForUserContext(ctx, db, user, database.Now())
	applyDingTalkH5UpdateAudit(&review.DingTalkH5AuditFields, audit)
	review.EditTime = audit.CurrentMilli
	updates := dingtalkH5AuditUpdateValues(audit)
	if review.CreateBy == 0 {
		for key, value := range dingtalkH5CreateAuditValues(audit) {
			updates[key] = value
		}
	}
	return db.Model(&model.DingTalkH5PerfReview{}).
		Where("`id` = ? AND `deleted_at` = 0", review.ID).
		Updates(updates).Error
}

func historiesForReview(ctx context.Context, reviewID uint) ([]model.DingTalkH5PerfHistory, error) {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	var histories []model.DingTalkH5PerfHistory
	err := db.Where("review_id = ?", reviewID).Order("add_time ASC, id ASC").Find(&histories).Error
	return histories, err
}

func historiesByReviewIDs(ctx context.Context, reviewIDs []uint) (map[uint][]model.DingTalkH5PerfHistory, error) {
	result := make(map[uint][]model.DingTalkH5PerfHistory, len(reviewIDs))
	if len(reviewIDs) == 0 {
		return result, nil
	}
	db, cancel := database.WithContext(ctx)
	defer cancel()
	var histories []model.DingTalkH5PerfHistory
	if err := db.Where("review_id IN ?", reviewIDs).Order("review_id ASC, add_time ASC, id ASC").Find(&histories).Error; err != nil {
		return nil, err
	}
	for _, history := range histories {
		result[history.ReviewID] = append(result[history.ReviewID], history)
	}
	return result, nil
}

func collectReviewIDs(reviews []model.DingTalkH5PerfReview) []uint {
	items := make([]uint, 0, len(reviews))
	for _, review := range reviews {
		if review.ID > 0 {
			items = append(items, review.ID)
		}
	}
	return items
}

func usersByAccounts(ctx context.Context, accounts []string) (map[string]*model.DingTalkH5PerfUser, error) {
	result := map[string]*model.DingTalkH5PerfUser{}
	if len(accounts) == 0 {
		return result, nil
	}
	db, cancel := database.WithContext(ctx)
	defer cancel()
	var users []model.DingTalkH5PerfUser
	if err := db.Select("`user_mini_openid`, `user_name`").Where("`user_mini_openid` IN ?", accounts).Find(&users).Error; err != nil {
		return nil, err
	}
	for i := range users {
		account := NormalizeUserID(users[i].Account)
		if account == "" {
			continue
		}
		users[i].Account = account
		result[account] = &users[i]
	}
	return result, nil
}

func collectEmployeeAccounts(reviews []model.DingTalkH5PerfReview) []string {
	items := make([]string, 0, len(reviews))
	seen := map[string]struct{}{}
	for _, review := range reviews {
		appendUniqueAccount(&items, seen, review.EmployeeAccount)
	}
	return items
}

func collectReviewParticipantAccounts(reviews []model.DingTalkH5PerfReview) []string {
	items := make([]string, 0, len(reviews)*4)
	seen := map[string]struct{}{}
	for _, review := range reviews {
		appendUniqueAccount(&items, seen, review.EmployeeAccount)
		appendUniqueAccount(&items, seen, review.ManagerAccount)
		appendUniqueAccount(&items, seen, review.HRBPAccount)
		appendUniqueAccount(&items, seen, review.HRBPReviewerAccount)
	}
	return items
}

func appendUniqueAccount(items *[]string, seen map[string]struct{}, account string) {
	account = strings.TrimSpace(account)
	if account == "" {
		return
	}
	if _, ok := seen[account]; ok {
		return
	}
	seen[account] = struct{}{}
	*items = append(*items, account)
}

func matchesKeyword(review model.DingTalkH5PerfReview, employee *model.DingTalkH5PerfUser, keyword string) bool {
	keyword = strings.ToLower(strings.TrimSpace(keyword))
	if keyword == "" {
		return true
	}
	haystack := strings.ToLower(strings.Join([]string{
		review.ReviewNo,
		review.EmployeeAccount,
		review.ManagerAccount,
		review.HRBPAccount,
		review.Department,
		review.Period,
		review.NextPeriod,
		review.Status,
		review.ManagerGrade,
		review.HRBPGrade,
		review.FinalGrade,
	}, " "))
	if employee != nil {
		haystack += " " + strings.ToLower(employee.Name)
	}
	return strings.Contains(haystack, keyword)
}

func defaultObjectives(items []NextObjective, reviewNo string) []Objective {
	result := make([]Objective, 0, len(items))
	for index, item := range items {
		result = append(result, Objective{ID: fmt.Sprintf("%s-obj-%d", reviewNo, index+1), Target: item.Target, Weight: item.Weight, Completion: "", Result: ""})
	}
	return result
}

type reviewObjectiveSource struct {
	reviewNo string
	period   string
}

func currentObjectivesForNewReview(reviewNo string, templateDefaults []NextObjective, previous *model.DingTalkH5PerfReview) ([]Objective, reviewObjectiveSource) {
	if previous != nil {
		objectives := objectivesFromNextObjectivesForCurrentReview(decodeNextObjectives(previous.NextObjectivesJSON), reviewNo)
		if len(objectives) > 0 {
			return objectives, reviewObjectiveSource{
				reviewNo: strings.TrimSpace(previous.ReviewNo),
				period:   strings.TrimSpace(previous.Period),
			}
		}
	}
	return defaultObjectives(templateDefaults, reviewNo), reviewObjectiveSource{}
}

func objectivesFromNextObjectivesForCurrentReview(items []NextObjective, reviewNo string) []Objective {
	items = sanitizeNextObjectives(items)
	result := make([]Objective, 0, len(items))
	for index, item := range items {
		result = append(result, Objective{
			ID:         fmt.Sprintf("%s-obj-%d", reviewNo, index+1),
			Target:     item.Target,
			Weight:     item.Weight,
			Completion: "",
			Result:     "",
		})
	}
	return result
}

func defaultNextObjectives(items []NextObjective, reviewNo string) []NextObjective {
	result := make([]NextObjective, 0, len(items))
	for index, item := range items {
		result = append(result, NextObjective{ID: fmt.Sprintf("%s-next-%d", reviewNo, index+1), Target: item.Target, Weight: item.Weight})
	}
	return result
}

func defaultValues(items []ValueTemplate) []ValueScore {
	result := make([]ValueScore, 0, len(items))
	for _, item := range items {
		result = append(result, valueScoreFromTemplate(item))
	}
	return result
}

func copySelfFields(review *model.DingTalkH5PerfReview, payload ReviewPayload) {
	review.ObjectivesJSON = encodeJSON(sanitizeObjectives(payload.Objectives))
	review.NextObjectivesJSON = encodeJSON(sanitizeNextObjectives(payload.NextObjectives))
	review.SelfSummary = strings.TrimSpace(payload.SelfSummary)
	values := mergeValues(decodeValues(review.ValuesJSON), sanitizeValues(payload.Values), "self")
	review.ValuesJSON = encodeJSON(values)
}

func validateSelfSubmitPayload(payload ReviewPayload) error {
	missing := []string{}
	if !selfObjectivesFilled(sanitizeObjectives(payload.Objectives)) {
		missing = append(missing, "本月目标")
	}
	if strings.TrimSpace(payload.SelfSummary) == "" {
		missing = append(missing, "思考总结")
	}
	if !allStageScoresFilled(sanitizeValues(payload.Values), "self") {
		missing = append(missing, "价值观自评")
	}
	if len(sanitizeNextObjectives(payload.NextObjectives)) == 0 {
		missing = append(missing, "下月目标")
	}
	if len(missing) > 0 {
		return fmt.Errorf("请完善：%s", strings.Join(missing, "、"))
	}
	return nil
}

func selfObjectivesFilled(items []Objective) bool {
	if len(items) == 0 {
		return false
	}
	for _, item := range items {
		if strings.TrimSpace(item.Target) == "" ||
			strings.TrimSpace(toString(item.Completion)) == "" ||
			strings.TrimSpace(item.Result) == "" {
			return false
		}
	}
	return true
}

func copyManagerFields(review *model.DingTalkH5PerfReview, payload ReviewPayload) {
	review.ManagerComment = strings.TrimSpace(payload.ManagerComment)
	review.ManagerGrade = strings.TrimSpace(payload.ManagerGrade)
	values := mergeValues(decodeValues(review.ValuesJSON), sanitizeValues(payload.Values), "manager")
	review.ValuesJSON = encodeJSON(values)
}

func copyHrbpFields(review *model.DingTalkH5PerfReview, payload ReviewPayload) {
	review.HRBPComment = strings.TrimSpace(payload.HRBPComment)
	review.HRBPGrade = strings.TrimSpace(payload.HRBPGrade)
	values := mergeValues(decodeValues(review.ValuesJSON), sanitizeValues(payload.Values), "hrbp")
	review.ValuesJSON = encodeJSON(values)
}

func sanitizeObjectives(items []Objective) []Objective {
	result := make([]Objective, 0, len(items))
	for index, item := range items {
		if index >= 12 {
			break
		}
		target := strings.TrimSpace(item.Target)
		if target == "" && index >= 3 {
			continue
		}
		id := strings.TrimSpace(item.ID)
		if id == "" {
			id = "obj-" + strconv.Itoa(index+1)
		}
		result = append(result, Objective{
			ID:         id,
			Target:     target,
			Weight:     clampNumber(item.Weight, 0, 100),
			Completion: normalizeScore(item.Completion, 0, 200),
			Result:     strings.TrimSpace(item.Result),
		})
	}
	return result
}

func sanitizeNextObjectives(items []NextObjective) []NextObjective {
	result := make([]NextObjective, 0, len(items))
	for index, item := range items {
		if index >= 12 {
			break
		}
		target := strings.TrimSpace(item.Target)
		if target == "" {
			continue
		}
		id := strings.TrimSpace(item.ID)
		if id == "" {
			id = "next-" + strconv.Itoa(index+1)
		}
		result = append(result, NextObjective{ID: id, Target: target, Weight: clampNumber(item.Weight, 0, 100)})
	}
	return result
}

func sanitizeValues(items []ValueScore) []ValueScore {
	result := make([]ValueScore, 0, len(items))
	for _, item := range items {
		id := strings.TrimSpace(item.ID)
		if id == "" {
			continue
		}
		result = append(result, ValueScore{
			ID:         id,
			Name:       strings.TrimSpace(item.Name),
			Definition: strings.TrimSpace(item.Definition),
			Rubric:     sanitizeValueRubric(item.Rubric),
			Self:       normalizeScore(item.Self, 0, 50),
			Manager:    normalizeScore(item.Manager, 0, 50),
			HRBP:       normalizeScore(item.HRBP, 0, 50),
			HR:         normalizeScore(item.HR, 0, 50),
		})
	}
	return result
}

func mergeValues(existing []ValueScore, incoming []ValueScore, field string) []ValueScore {
	if len(existing) == 0 {
		existing = incoming
	}
	index := map[string]ValueScore{}
	for _, item := range incoming {
		index[item.ID] = item
	}
	for idx, item := range existing {
		next, ok := index[item.ID]
		if !ok {
			continue
		}
		switch field {
		case "self":
			item.Self = next.Self
		case "manager":
			item.Manager = next.Manager
		case "hrbp":
			item.HRBP = next.HRBP
		case "hr":
			item.HR = next.HR
		}
		item = mergeValueScoreMeta(item, next)
		existing[idx] = item
	}
	return existing
}

func valueScoreFromTemplate(item ValueTemplate) ValueScore {
	return ValueScore{
		ID:         strings.TrimSpace(item.ID),
		Name:       strings.TrimSpace(item.Name),
		Definition: strings.TrimSpace(item.Definition),
		Rubric:     cloneValueRubric(item.Rubric),
		Self:       "",
		Manager:    "",
		HRBP:       "",
		HR:         "",
	}
}

func mergeValueScoreMeta(existing ValueScore, incoming ValueScore) ValueScore {
	if strings.TrimSpace(existing.Name) == "" {
		existing.Name = strings.TrimSpace(incoming.Name)
	}
	if strings.TrimSpace(existing.Definition) == "" {
		existing.Definition = strings.TrimSpace(incoming.Definition)
	}
	if len(existing.Rubric) == 0 {
		existing.Rubric = cloneValueRubric(incoming.Rubric)
	}
	return existing
}

func enrichValueScoresWithTemplate(values []ValueScore, templates []ValueTemplate) []ValueScore {
	if len(values) == 0 {
		return values
	}
	templateByID := make(map[string]ValueTemplate, len(templates))
	for _, item := range templates {
		id := strings.TrimSpace(item.ID)
		if id != "" {
			templateByID[id] = item
		}
	}
	result := make([]ValueScore, 0, len(values))
	for _, item := range values {
		if tpl, ok := templateByID[strings.TrimSpace(item.ID)]; ok {
			if strings.TrimSpace(tpl.Name) != "" {
				item.Name = strings.TrimSpace(tpl.Name)
			}
			if strings.TrimSpace(tpl.Definition) != "" {
				item.Definition = strings.TrimSpace(tpl.Definition)
			}
			if len(tpl.Rubric) > 0 {
				item.Rubric = cloneValueRubric(tpl.Rubric)
			}
		}
		result = append(result, item)
	}
	return result
}

func hydrateReviewDTOValues(dto *ReviewDTO, templates []ValueTemplate) {
	if dto == nil {
		return
	}
	dto.Values = enrichValueScoresWithTemplate(dto.Values, templates)
}

func loadReviewValueTemplatesContext(ctx context.Context) []ValueTemplate {
	tpl, err := LoadTemplateContext(ctx)
	if err != nil {
		return nil
	}
	return tpl.Values
}

func cloneValueRubric(items []ValueRubric) []ValueRubric {
	if len(items) == 0 {
		return nil
	}
	result := make([]ValueRubric, len(items))
	copy(result, items)
	return result
}

func normalizeScore(value interface{}, min, max float64) interface{} {
	if value == nil {
		return ""
	}
	switch typed := value.(type) {
	case string:
		typed = strings.TrimSpace(typed)
		if typed == "" {
			return ""
		}
		number, err := strconv.ParseFloat(typed, 64)
		if err != nil {
			return ""
		}
		return clampNumber(number, min, max)
	case float64:
		if math.IsNaN(typed) || math.IsInf(typed, 0) {
			return ""
		}
		return clampNumber(typed, min, max)
	case float32:
		return clampNumber(float64(typed), min, max)
	case int:
		return clampNumber(float64(typed), min, max)
	case int64:
		return clampNumber(float64(typed), min, max)
	default:
		number, err := strconv.ParseFloat(toString(value), 64)
		if err != nil {
			return ""
		}
		return clampNumber(number, min, max)
	}
}

func clampNumber(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return math.Round(value*10) / 10
}

func allStageScoresFilled(values []ValueScore, field string) bool {
	if len(values) == 0 {
		return false
	}
	for _, item := range values {
		var value interface{}
		switch field {
		case "self":
			value = item.Self
		case "manager":
			value = item.Manager
		case "hrbp":
			value = item.HRBP
		default:
			value = item.HR
		}
		if value == nil || strings.TrimSpace(toString(value)) == "" {
			return false
		}
	}
	return true
}

func managerReviewStarted(review model.DingTalkH5PerfReview) bool {
	if strings.TrimSpace(review.ManagerGrade) != "" || strings.TrimSpace(review.ManagerComment) != "" {
		return true
	}
	for _, item := range decodeValues(review.ValuesJSON) {
		if strings.TrimSpace(toString(item.Manager)) != "" {
			return true
		}
	}
	return false
}

func shouldSkipHrbpStage(ctx context.Context, review model.DingTalkH5PerfReview) bool {
	return review.ManagerAccount != "" && review.ManagerAccount == review.HRBPAccount
}

func returnReview(ctx context.Context, user *model.DingTalkH5PerfUser, reviewNo, fromStatus, toStatus, action string) (*ReviewDTO, error) {
	return mutateReview(ctx, user, reviewNo, func(db *gorm.DB, review *model.DingTalkH5PerfReview) error {
		if review.Status != fromStatus {
			return fmt.Errorf("当前阶段不能退回")
		}
		if fromStatus == ReviewStatusManagerReview && review.ManagerAccount != user.Account {
			return fmt.Errorf("当前阶段不能退回员工修改")
		}
		review.Status = toStatus
		review.EditTime = database.Now()
		if err := db.Save(review).Error; err != nil {
			return err
		}
		return addHistoryWithDB(db, review, user, action)
	})
}

func returnReason(payload ReviewPayload) string {
	reason := normalizeReviewReason(payload.ReturnReason)
	if reason == "" {
		return "未填写原因"
	}
	return reason
}

func normalizeReviewReason(reason string) string {
	reason = strings.TrimSpace(reason)
	runes := []rune(reason)
	if len(runes) > 200 {
		return string(runes[:200])
	}
	return reason
}

func departmentFromUser(user model.DingTalkH5PerfUser) string {
	if text := departmentText(user.DepartmentLevel1, user.DepartmentLevel2, user.DepartmentLevel3); text != "" {
		return text
	}
	return user.Department
}

func validMonth(value string) bool {
	if len(value) != 7 || value[4] != '-' {
		return false
	}
	year, errYear := strconv.Atoi(value[:4])
	month, errMonth := strconv.Atoi(value[5:])
	return errYear == nil && errMonth == nil && year >= 2000 && month >= 1 && month <= 12
}

func fallback(value, fallbackValue string) string {
	value = strings.TrimSpace(value)
	if value != "" {
		return value
	}
	return strings.TrimSpace(fallbackValue)
}

func effectiveGrade(review ReviewDTO) string {
	if review.FinalGrade != "" {
		return review.FinalGrade
	}
	return review.HRBPGrade
}

func objectiveTotal(items []Objective) float64 {
	var total float64
	for _, item := range items {
		completion := numericValue(item.Completion)
		total += item.Weight * completion / 100
	}
	return math.Round(total*10) / 10
}

func valueTotal(items []ValueScore, field string) interface{} {
	var total float64
	var count int
	for _, item := range items {
		var value interface{}
		switch field {
		case "self":
			value = item.Self
		case "manager":
			value = item.Manager
		case "hrbp":
			value = item.HRBP
		default:
			value = item.HR
		}
		if strings.TrimSpace(toString(value)) == "" {
			continue
		}
		total += numericValue(value)
		count++
	}
	if count == 0 {
		return ""
	}
	return math.Round(total*10) / 10
}

func numericValue(value interface{}) float64 {
	switch typed := value.(type) {
	case float64:
		return typed
	case int:
		return float64(typed)
	case string:
		number, _ := strconv.ParseFloat(strings.TrimSpace(typed), 64)
		return number
	default:
		number, _ := strconv.ParseFloat(toString(value), 64)
		return number
	}
}

func sortReviews(items []ReviewDTO) {
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].Period == items[j].Period {
			return items[i].ID > items[j].ID
		}
		return items[i].Period > items[j].Period
	})
}
