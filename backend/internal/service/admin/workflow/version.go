package workflowservice

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"wecheckin/backend/internal/model"
	"wecheckin/backend/internal/workflowcore"
	"wecheckin/backend/pkg/database"
)

const maxVersionPublishNoteLength = 500

type VersionSummary struct {
	ID                  uint   `json:"id"`
	DefinitionID        uint   `json:"definitionId"`
	Version             int    `json:"version"`
	Name                string `json:"name"`
	DeploymentID        string `json:"deploymentId"`
	PublishedBy         uint   `json:"publishedBy"`
	PublishedByName     string `json:"publishedByName"`
	PublishedAt         int64  `json:"publishedAt"`
	PublishNote         string `json:"publishNote"`
	RollbackFromVersion int    `json:"rollbackFromVersion"`
	ChangeBaseVersion   int    `json:"changeBaseVersion"`
	ChangeHeadline      string `json:"changeHeadline"`
	ChangeCount         int    `json:"changeCount"`
	InstanceCount       int64  `json:"instanceCount"`
	StartDraftCount     int64  `json:"startDraftCount"`
	IsCurrent           bool   `json:"isCurrent"`
	CanDelete           bool   `json:"canDelete"`
	DeleteBlockedReason string `json:"deleteBlockedReason"`
}

type RollbackRequest struct {
	Note string `json:"note"`
}

func GetVersionsContext(ctx context.Context, id uint) ([]VersionSummary, error) {
	if id == 0 {
		return nil, errors.New("流程定义 ID 无效")
	}
	db, cancel := database.WithContext(ctx)
	defer cancel()

	var definition model.WorkflowDefinition
	if err := db.First(&definition, id).Error; err != nil {
		return nil, definitionError(err)
	}
	var rows []model.WorkflowDefinitionVersion
	if err := db.Select(
		"id", "definition_id", "definition_version", "definition_source_json", "definition_metadata_json",
		"definition_change_base_version", "definition_change_summary_json", "definition_publish_note",
		"definition_rollback_from_version", "definition_deployment_id", "definition_published_by", "definition_published_at",
	).Where("definition_id = ?", id).Order("definition_version DESC").Find(&rows).Error; err != nil {
		return nil, err
	}
	instanceCounts, err := loadVersionReferenceCounts(db, &model.WorkflowProcessInstance{}, id)
	if err != nil {
		return nil, err
	}
	draftCounts, err := loadVersionReferenceCounts(db, &model.WorkflowStartDraft{}, id)
	if err != nil {
		return nil, err
	}
	publisherNames, err := loadVersionPublisherNames(db, rows)
	if err != nil {
		return nil, err
	}

	fallback := metadataFromDefinition(definition)
	snapshots := make(map[int]versionSnapshot, len(rows))
	for _, row := range rows {
		snapshot, _, err := versionSnapshotFromModel(row, fallback)
		if err != nil {
			return nil, err
		}
		snapshots[row.Version] = snapshot
	}
	versions := sortedVersionNumbers(rows)
	result := make([]VersionSummary, 0, len(rows))
	for _, row := range rows {
		summary, ok, err := decodeVersionChangeSummary(row.ChangeSummaryJSON)
		if err != nil {
			return nil, fmt.Errorf("解析流程定义 v%d 变更摘要失败: %w", row.Version, err)
		}
		if !ok {
			baseVersion := row.ChangeBaseVersion
			if baseVersion <= 0 {
				baseVersion = nearestLowerVersion(versions, row.Version)
			}
			summary = buildVersionChangeSummary(baseVersion, snapshots[baseVersion], snapshots[row.Version])
		}
		instances := instanceCounts[row.Version]
		drafts := draftCounts[row.Version]
		blockedReason := versionDeleteBlockReason(row.Version, definition.CurrentVersion, instances, drafts)
		result = append(result, VersionSummary{
			ID: row.ID, DefinitionID: row.DefinitionID, Version: row.Version,
			Name: snapshots[row.Version].Metadata.Name, DeploymentID: row.DeploymentID,
			PublishedBy: row.PublishedBy, PublishedByName: publisherNames[row.PublishedBy], PublishedAt: row.PublishedAt,
			PublishNote: row.PublishNote, RollbackFromVersion: row.RollbackFromVersion,
			ChangeBaseVersion: summary.BaseVersion, ChangeHeadline: summary.Headline, ChangeCount: summary.ChangeCount,
			InstanceCount: instances, StartDraftCount: drafts, IsCurrent: row.Version == definition.CurrentVersion,
			CanDelete: blockedReason == "", DeleteBlockedReason: blockedReason,
		})
	}
	return result, nil
}

func GetVersionChangesContext(ctx context.Context, id uint, version, compareTo int) (*VersionChangeSummary, error) {
	if id == 0 || version <= 0 || compareTo < 0 || compareTo == version {
		return nil, errors.New("流程版本参数无效")
	}
	db, cancel := database.WithContext(ctx)
	defer cancel()

	var definition model.WorkflowDefinition
	if err := db.First(&definition, id).Error; err != nil {
		return nil, definitionError(err)
	}
	var target model.WorkflowDefinitionVersion
	if err := db.First(&target, "definition_id = ? AND definition_version = ?", id, version).Error; err != nil {
		return nil, versionError(err)
	}
	if compareTo == 0 {
		if summary, ok, err := decodeVersionChangeSummary(target.ChangeSummaryJSON); err != nil {
			return nil, err
		} else if ok {
			return &summary, nil
		}
		compareTo = target.ChangeBaseVersion
	}
	if compareTo <= 0 {
		if err := db.Model(&model.WorkflowDefinitionVersion{}).
			Where("definition_id = ? AND definition_version < ?", id, version).
			Select("COALESCE(MAX(definition_version), 0)").Scan(&compareTo).Error; err != nil {
			return nil, err
		}
	}

	fallback := metadataFromDefinition(definition)
	targetSnapshot, _, err := versionSnapshotFromModel(target, fallback)
	if err != nil {
		return nil, err
	}
	var baseSnapshot versionSnapshot
	if compareTo > 0 {
		var base model.WorkflowDefinitionVersion
		if err := db.First(&base, "definition_id = ? AND definition_version = ?", id, compareTo).Error; err != nil {
			return nil, versionError(err)
		}
		baseSnapshot, _, err = versionSnapshotFromModel(base, fallback)
		if err != nil {
			return nil, err
		}
	}
	summary := buildVersionChangeSummary(compareTo, baseSnapshot, targetSnapshot)
	return &summary, nil
}

func DeleteVersionContext(ctx context.Context, id uint, version int) error {
	if id == 0 || version <= 0 {
		return errors.New("流程版本参数无效")
	}
	db, cancel := database.WithContext(ctx)
	defer cancel()
	return db.Transaction(func(tx *gorm.DB) error {
		var definition model.WorkflowDefinition
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&definition, id).Error; err != nil {
			return definitionError(err)
		}
		var item model.WorkflowDefinitionVersion
		if err := tx.First(&item, "definition_id = ? AND definition_version = ?", id, version).Error; err != nil {
			return versionError(err)
		}
		instances, err := countVersionReferences(tx, &model.WorkflowProcessInstance{}, id, version)
		if err != nil {
			return err
		}
		drafts, err := countVersionReferences(tx, &model.WorkflowStartDraft{}, id, version)
		if err != nil {
			return err
		}
		if reason := versionDeleteBlockReason(version, definition.CurrentVersion, instances, drafts); reason != "" {
			return errors.New(reason)
		}
		if err := preserveDependentVersionSummary(tx, definition, item); err != nil {
			return err
		}
		return tx.Unscoped().Delete(&item).Error
	})
}

func preserveDependentVersionSummary(tx *gorm.DB, definition model.WorkflowDefinition, deleting model.WorkflowDefinitionVersion) error {
	var next model.WorkflowDefinitionVersion
	err := tx.Where("definition_id = ? AND definition_version > ?", definition.ID, deleting.Version).
		Order("definition_version ASC").First(&next).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	if strings.TrimSpace(next.ChangeSummaryJSON) != "" || (next.ChangeBaseVersion > 0 && next.ChangeBaseVersion != deleting.Version) {
		return nil
	}
	fallback := metadataFromDefinition(definition)
	baseSnapshot, _, err := versionSnapshotFromModel(deleting, fallback)
	if err != nil {
		return err
	}
	nextSnapshot, _, err := versionSnapshotFromModel(next, fallback)
	if err != nil {
		return err
	}
	summaryJSON, err := encodeVersionChangeSummary(buildVersionChangeSummary(deleting.Version, baseSnapshot, nextSnapshot))
	if err != nil {
		return err
	}
	return tx.Model(&model.WorkflowDefinitionVersion{}).
		Where("id = ?", next.ID).
		Updates(map[string]interface{}{
			"definition_change_base_version": deleting.Version,
			"definition_change_summary_json": summaryJSON,
		}).Error
}

func RollbackVersionContext(ctx context.Context, adminID, id uint, version int, request RollbackRequest) (*PublishResponse, error) {
	if id == 0 || version <= 0 {
		return nil, errors.New("流程版本参数无效")
	}
	note, err := normalizeVersionPublishNote(request.Note)
	if err != nil {
		return nil, err
	}
	db, cancel := database.WithContext(ctx)
	defer cancel()
	var result PublishResponse
	err = db.Transaction(func(tx *gorm.DB) error {
		var definition model.WorkflowDefinition
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&definition, id).Error; err != nil {
			return definitionError(err)
		}
		if definition.CurrentVersion < 1 {
			return errors.New("流程尚未发布，不能回滚")
		}
		if version == definition.CurrentVersion {
			return errors.New("当前版本无需回滚")
		}
		var target, current model.WorkflowDefinitionVersion
		if err := tx.First(&target, "definition_id = ? AND definition_version = ?", id, version).Error; err != nil {
			return versionError(err)
		}
		if err := tx.First(&current, "definition_id = ? AND definition_version = ?", id, definition.CurrentVersion).Error; err != nil {
			return versionError(err)
		}

		fallback := metadataFromDefinition(definition)
		targetSnapshot, _, err := versionSnapshotFromModel(target, fallback)
		if err != nil {
			return err
		}
		currentSnapshot, _, err := versionSnapshotFromModel(current, fallback)
		if err != nil {
			return err
		}
		if validationErrors := workflowcore.ValidateDefinition(targetSnapshot.Definition); len(validationErrors) > 0 {
			return workflowcore.ValidationErrors(validationErrors)
		}
		if jsonEqual(currentSnapshot, targetSnapshot) {
			return errors.New("目标版本与当前版本内容相同，无需回滚")
		}
		bpmn, err := workflowcore.CompileBPMN(targetSnapshot.Definition)
		if err != nil {
			return err
		}
		if note == "" {
			note = fmt.Sprintf("回滚至 v%d", version)
		}
		newVersion := definition.CurrentVersion + 1
		summary := buildVersionChangeSummary(definition.CurrentVersion, currentSnapshot, targetSnapshot)
		versionItem, err := newDefinitionVersionModel(
			definition.ID, newVersion, adminID, database.Now(), target.SourceJSON, string(bpmn),
			targetSnapshot, summary, note, version,
		)
		if err != nil {
			return err
		}
		if err := tx.Create(&versionItem).Error; err != nil {
			return err
		}
		metadata := targetSnapshot.Metadata
		if err := tx.Model(&model.WorkflowDefinition{}).Where("id = ?", definition.ID).Updates(map[string]interface{}{
			"definition_name":            metadata.Name,
			"definition_description":     metadata.Description,
			"definition_category":        metadata.Category,
			"definition_logo_url":        metadata.LogoURL,
			"definition_current_version": newVersion,
			"definition_status":          model.DefinitionStatusPublished,
			"definition_draft_json":      target.SourceJSON,
			"definition_edit_user_id":    adminID,
			"definition_edit_time":       versionItem.PublishedAt,
			"updated_at":                 gorm.Expr("CURRENT_TIMESTAMP"),
		}).Error; err != nil {
			return err
		}
		result = PublishResponse{DefinitionID: definition.ID, Version: newVersion, BPMNXML: string(bpmn)}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func newDefinitionVersionModel(definitionID uint, version int, adminID uint, publishedAt int64, sourceJSON, bpmnXML string, snapshot versionSnapshot, summary VersionChangeSummary, note string, rollbackFrom int) (model.WorkflowDefinitionVersion, error) {
	metadataJSON, err := encodeVersionMetadata(snapshot.Metadata)
	if err != nil {
		return model.WorkflowDefinitionVersion{}, err
	}
	summaryJSON, err := encodeVersionChangeSummary(summary)
	if err != nil {
		return model.WorkflowDefinitionVersion{}, err
	}
	contentHash, err := versionContentHash(snapshot)
	if err != nil {
		return model.WorkflowDefinitionVersion{}, err
	}
	validationJSON, err := json.Marshal([]workflowcore.ValidationError{})
	if err != nil {
		return model.WorkflowDefinitionVersion{}, err
	}
	return model.WorkflowDefinitionVersion{
		DefinitionID: definitionID, Version: version, SourceJSON: sourceJSON,
		MetadataJSON: metadataJSON, ChangeBaseVersion: summary.BaseVersion, ChangeSummaryJSON: summaryJSON,
		PublishNote: note, ContentHash: contentHash, RollbackFromVersion: rollbackFrom,
		BPMNXML: bpmnXML, ValidationJSON: string(validationJSON), PublishedBy: adminID, PublishedAt: publishedAt,
	}, nil
}

func normalizeVersionPublishNote(note string) (string, error) {
	note = strings.TrimSpace(note)
	if utf8.RuneCountInString(note) > maxVersionPublishNoteLength {
		return "", fmt.Errorf("发布说明不能超过 %d 个字符", maxVersionPublishNoteLength)
	}
	return note, nil
}

func loadVersionReferenceCounts(db *gorm.DB, modelValue interface{}, definitionID uint) (map[int]int64, error) {
	var rows []struct {
		Version int   `gorm:"column:definition_version"`
		Total   int64 `gorm:"column:total"`
	}
	if err := db.Model(modelValue).Select("definition_version, COUNT(*) AS total").
		Where("definition_id = ?", definitionID).Group("definition_version").Scan(&rows).Error; err != nil {
		return nil, err
	}
	result := make(map[int]int64, len(rows))
	for _, row := range rows {
		result[row.Version] = row.Total
	}
	return result, nil
}

func countVersionReferences(db *gorm.DB, modelValue interface{}, definitionID uint, version int) (int64, error) {
	var count int64
	err := db.Model(modelValue).Where("definition_id = ? AND definition_version = ?", definitionID, version).Count(&count).Error
	return count, err
}

func loadVersionPublisherNames(db *gorm.DB, versions []model.WorkflowDefinitionVersion) (map[uint]string, error) {
	ids := make([]uint, 0)
	seen := make(map[uint]struct{})
	for _, version := range versions {
		if version.PublishedBy == 0 {
			continue
		}
		if _, exists := seen[version.PublishedBy]; exists {
			continue
		}
		seen[version.PublishedBy] = struct{}{}
		ids = append(ids, version.PublishedBy)
	}
	result := make(map[uint]string, len(ids))
	if len(ids) == 0 {
		return result, nil
	}
	var admins []model.Admin
	if err := db.Select("id", "user_name").Where("id IN ?", ids).Find(&admins).Error; err != nil {
		return nil, err
	}
	for _, admin := range admins {
		result[admin.ID] = strings.TrimSpace(admin.Name)
	}
	return result, nil
}

func versionDeleteBlockReason(version, currentVersion int, instanceCount, draftCount int64) string {
	if version == currentVersion {
		return "当前发布版本不能删除"
	}
	parts := make([]string, 0, 2)
	if instanceCount > 0 {
		parts = append(parts, fmt.Sprintf("%d 个流程实例", instanceCount))
	}
	if draftCount > 0 {
		parts = append(parts, fmt.Sprintf("%d 份发起草稿", draftCount))
	}
	if len(parts) > 0 {
		return "该版本已被" + strings.Join(parts, "、") + "引用，不能删除"
	}
	return ""
}

func nearestLowerVersion(versions []int, target int) int {
	result := 0
	for _, version := range versions {
		if version >= target {
			break
		}
		result = version
	}
	return result
}

func versionError(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return errors.New("流程版本不存在")
	}
	return err
}
