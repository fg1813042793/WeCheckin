package model

import (
	accountmodel "wecheckin/backend/internal/model/account"
	adminmodel "wecheckin/backend/internal/model/admin"
	assessmentmodel "wecheckin/backend/internal/model/assessment"
	contentmodel "wecheckin/backend/internal/model/content"
	dingtalkh5model "wecheckin/backend/internal/model/dingtalkh5"
	interactionmodel "wecheckin/backend/internal/model/interaction"
	organizationmodel "wecheckin/backend/internal/model/organization"
	permissionmodel "wecheckin/backend/internal/model/permission"
	scheduledtaskmodel "wecheckin/backend/internal/model/scheduledtask"
	systemmodel "wecheckin/backend/internal/model/system"
	workflowmodel "wecheckin/backend/internal/model/workflow"
)

type Admin = adminmodel.Admin
type Log = adminmodel.Log

type User = accountmodel.User
type UserDept = accountmodel.UserDept
type UserRole = accountmodel.UserRole
type UserReportingRelation = accountmodel.UserReportingRelation
type UserFormField = accountmodel.UserFormField

type Department = organizationmodel.Department
type Position = organizationmodel.Position
type Role = organizationmodel.Role

type Permission = permissionmodel.Permission
type PermissionGrant = permissionmodel.PermissionGrant

type Setup = systemmodel.Setup
type SysDict = systemmodel.SysDict

type News = contentmodel.News
type Favorite = contentmodel.Favorite
type Notify = contentmodel.Notify

type Enroll = interactionmodel.Enroll
type EnrollJoin = interactionmodel.EnrollJoin
type EnrollUser = interactionmodel.EnrollUser
type Event = interactionmodel.Event
type EventRole = interactionmodel.EventRole
type EventParticipant = interactionmodel.EventParticipant
type EventDynamic = interactionmodel.EventDynamic
type EventScore = interactionmodel.EventScore

type ExamQuestion = assessmentmodel.ExamQuestion
type ExamPaper = assessmentmodel.ExamPaper
type Exam = assessmentmodel.Exam
type ExamRecord = assessmentmodel.ExamRecord
type ExamResource = assessmentmodel.ExamResource
type Survey = assessmentmodel.Survey
type SurveyResponse = assessmentmodel.SurveyResponse
type SurveyChannel = assessmentmodel.SurveyChannel
type SurveyAILog = assessmentmodel.SurveyAILog
type SurveyResource = assessmentmodel.SurveyResource
type SurveyQuestion = assessmentmodel.SurveyQuestion

type DingTalkH5AuditFields = dingtalkh5model.DingTalkH5AuditFields
type DingTalkH5PerfUser = dingtalkh5model.DingTalkH5PerfUser
type DingTalkH5CorpConfig = dingtalkh5model.DingTalkH5CorpConfig
type DingTalkH5UserBinding = dingtalkh5model.DingTalkH5UserBinding
type DingTalkH5PerfReview = dingtalkh5model.DingTalkH5PerfReview
type DingTalkH5PerfHistory = dingtalkh5model.DingTalkH5PerfHistory
type DingTalkH5PerfTemplate = dingtalkh5model.DingTalkH5PerfTemplate

type WorkflowDefinition = workflowmodel.Definition
type WorkflowDefinitionVersion = workflowmodel.DefinitionVersion
type WorkflowProcessInstance = workflowmodel.ProcessInstance
type WorkflowStartDraft = workflowmodel.StartDraft
type WorkflowProcessToken = workflowmodel.ProcessToken
type WorkflowProcessTask = workflowmodel.ProcessTask
type WorkflowProcessVariable = workflowmodel.ProcessVariable
type WorkflowProcessHistory = workflowmodel.ProcessHistory
type WorkflowInstanceParticipant = workflowmodel.InstanceParticipant
type WorkflowNotificationOutbox = workflowmodel.NotificationOutbox
type WorkflowOrgApproverIdentity = workflowmodel.OrgApproverIdentity
type WorkflowOrgApproverAssignment = workflowmodel.OrgApproverAssignment

type ScheduledTask = scheduledtaskmodel.Task
type ScheduledTaskRun = scheduledtaskmodel.Run
type ScheduledTaskRunLog = scheduledtaskmodel.RunLog

const (
	DefinitionStatusDisabled  = workflowmodel.DefinitionStatusDisabled
	DefinitionStatusDraft     = workflowmodel.DefinitionStatusDraft
	DefinitionStatusPublished = workflowmodel.DefinitionStatusPublished

	OrgApproverIdentityStatusDisabled = workflowmodel.OrgApproverIdentityStatusDisabled
	OrgApproverIdentityStatusEnabled  = workflowmodel.OrgApproverIdentityStatusEnabled
	OrgApproverAssignmentStatusOff    = workflowmodel.OrgApproverAssignmentStatusOff
	OrgApproverAssignmentStatusOn     = workflowmodel.OrgApproverAssignmentStatusOn
	OrgApproverSubjectTypeDepartment  = workflowmodel.OrgApproverSubjectTypeDepartment
	OrgApproverSubjectTypeUser        = workflowmodel.OrgApproverSubjectTypeUser

	ReportingRelationTypeDirect = accountmodel.ReportingRelationTypeDirect
	ReportingRelationTypeDotted = accountmodel.ReportingRelationTypeDotted
	ReportingRelationStatusOff  = accountmodel.ReportingRelationStatusOff
	ReportingRelationStatusOn   = accountmodel.ReportingRelationStatusOn
)

var ParseJSON = interactionmodel.ParseJSON
