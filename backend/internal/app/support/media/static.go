package media

import (
	"context"

	"wecheckin-backend/backend/internal/model"
	"wecheckin-backend/backend/pkg/database"
)

const defaultStaticDomain = "http://localhost:8083"

// StaticDomain returns the configured static resource domain.
func StaticDomain() string {
	return StaticDomainContext(context.Background())
}

// StaticDomainContext returns the configured static resource domain using the caller context.
func StaticDomainContext(ctx context.Context) string {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	if db == nil {
		return defaultStaticDomain
	}
	var setup model.Setup
	if err := db.Where("`setup_key` = ?", "STATIC_DOMAIN").First(&setup).Error; err != nil {
		return defaultStaticDomain
	}
	if setup.Value == "" {
		return defaultStaticDomain
	}
	return setup.Value
}

// FullURLWithStaticDomain returns an absolute URL using the configured static domain.
func FullURLWithStaticDomain(path string) string {
	return FullURL(path, StaticDomain())
}
