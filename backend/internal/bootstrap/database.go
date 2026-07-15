package bootstrap

import "log"

func InitBusiness(enableExam bool) {
	if autoMigrateEnabled() {
		if err := autoMigrate(); err != nil {
			log.Printf("Migration warning: %v (continuing)", err)
		}
	} else {
		log.Println("AutoMigrate disabled by WECHECKIN_AUTO_MIGRATE")
	}

	seedSetups()
	seedMenus(enableExam)
}
