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
	systemmodel "wecheckin/backend/internal/model/system"
)

type Admin = adminmodel.Admin
type Log = adminmodel.Log

type User = accountmodel.User
type UserDept = accountmodel.UserDept
type UserRole = accountmodel.UserRole
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

var ParseJSON = interactionmodel.ParseJSON
