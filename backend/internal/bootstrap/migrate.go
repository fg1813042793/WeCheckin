package bootstrap

import (
	"fmt"
	"log"
	"os"
	"strings"

	"wecheckin-backend/backend/internal/model"
	"wecheckin-backend/backend/pkg/database"
)

type migrationStep struct {
	Name string
	Run  func() error
}

type migrationLogFunc func(format string, args ...interface{})

func runMigrationSteps(steps []migrationStep, logf migrationLogFunc) error {
	for _, step := range steps {
		if err := step.Run(); err != nil {
			return fmt.Errorf("migration step %s: %w", step.Name, err)
		}
		if logf != nil {
			logf("migration step completed: " + step.Name)
		}
	}
	return nil
}

func autoMigrateEnabled() bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv("WECHECKIN_AUTO_MIGRATE"))) {
	case "", "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		return true
	}
}

func autoMigrate() error {
	err := database.DB.AutoMigrate(
		&model.User{},
		&model.News{},
		&model.Enroll{},
		&model.EnrollJoin{},
		&model.EnrollUser{},
		&model.Favorite{},
		&model.Admin{},
		&model.Log{},
		&model.Setup{},
		&model.Role{},
		&model.SysDict{},
		&model.Department{},
		&model.UserDept{},
		&model.Menu{},
		&model.RoleMenu{},
		&model.AdminDept{},
		&model.RoleDept{},
		&model.Event{},
		&model.EventRole{},
		&model.EventParticipant{},
		&model.EventDynamic{},
		&model.EventScore{},
		&model.ExamQuestion{},
		&model.ExamPaper{},
		&model.Exam{},
		&model.ExamRecord{},
		&model.ExamResource{},
		&model.Survey{},
		&model.SurveyResponse{},
		&model.SurveyChannel{},
		&model.SurveyAILog{},
		&model.SurveyResource{},
		&model.SurveyQuestion{},
		&model.Notify{},
	)
	if err != nil {
		return err
	}
	return runMigrationSteps([]migrationStep{
		{
			Name: "adjust_event_scores_score_text",
			Run: func() error {
				return database.DB.Exec("ALTER TABLE `event_scores` MODIFY COLUMN `event_score_score` TEXT COMMENT '成绩'").Error
			},
		},
		{
			Name: "adjust_survey_schema_mediumtext",
			Run: func() error {
				return database.DB.Exec("ALTER TABLE `survey` MODIFY COLUMN `survey_schema` MEDIUMTEXT COMMENT 'formkit schema (JSON)'").Error
			},
		},
	}, log.Printf)
}
