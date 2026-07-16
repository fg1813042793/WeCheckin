package bootstrap

import (
	"fmt"
	"log"
)

func InitBusiness(enableExam bool) error {
	if autoMigrateEnabled() {
		if err := autoMigrate(); err != nil {
			return fmt.Errorf("auto migrate: %w", err)
		}
	} else {
		log.Println("AutoMigrate disabled by WECHECKIN_AUTO_MIGRATE")
	}

	seedSetups()
	seedMenus(enableExam)
	return nil
}
