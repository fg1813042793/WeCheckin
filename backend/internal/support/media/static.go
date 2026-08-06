package media

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"

	"gorm.io/gorm"

	"wecheckin/backend/internal/model"
	"wecheckin/backend/pkg/database"
)

const defaultStaticDomain = "http://localhost:8083"
const staticDomainSetupKey = "STATIC_DOMAIN"
const staticDomainCacheTTL = 5 * time.Minute

type staticDomainCacheEntry struct {
	value     string
	expiresAt time.Time
}

var staticDomainCache = struct {
	sync.RWMutex
	entry staticDomainCacheEntry
	ok    bool
}{}

func getStaticDomainCache(now time.Time) (string, bool) {
	staticDomainCache.RLock()
	entry := staticDomainCache.entry
	ok := staticDomainCache.ok
	staticDomainCache.RUnlock()
	if !ok || !now.Before(entry.expiresAt) {
		if ok {
			InvalidateStaticDomainCache()
		}
		return "", false
	}
	return entry.value, true
}

func setStaticDomainCache(value string, now time.Time) {
	staticDomainCache.Lock()
	staticDomainCache.entry = staticDomainCacheEntry{
		value:     value,
		expiresAt: now.Add(staticDomainCacheTTL),
	}
	staticDomainCache.ok = true
	staticDomainCache.Unlock()
}

func InvalidateStaticDomainCache() {
	staticDomainCache.Lock()
	staticDomainCache.entry = staticDomainCacheEntry{}
	staticDomainCache.ok = false
	staticDomainCache.Unlock()
}

// StaticDomain returns the configured static resource domain.
func StaticDomain() string {
	return StaticDomainContext(context.Background())
}

// StaticDomainContext returns the configured static resource domain using the caller context.
func StaticDomainContext(ctx context.Context) string {
	now := time.Now()
	if domain, ok := getStaticDomainCache(now); ok {
		return domain
	}
	db, cancel := database.WithContext(ctx)
	defer cancel()
	if db == nil {
		return defaultStaticDomain
	}
	var setup model.Setup
	if err := db.Where("`setup_key` = ?", staticDomainSetupKey).Take(&setup).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			setStaticDomainCache(defaultStaticDomain, now)
		}
		return defaultStaticDomain
	}
	domain := strings.TrimSpace(setup.Value)
	if domain == "" {
		setStaticDomainCache(defaultStaticDomain, now)
		return defaultStaticDomain
	}
	setStaticDomainCache(domain, now)
	return domain
}

// FullURLWithStaticDomain returns an absolute URL using the configured static domain.
func FullURLWithStaticDomain(path string) string {
	return FullURLWithStaticDomainContext(context.Background(), path)
}

// FullURLWithStaticDomainContext returns an absolute URL using the configured static domain.
func FullURLWithStaticDomainContext(ctx context.Context, path string) string {
	if path == "" || strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return FullURL(path, "")
	}
	return FullURL(path, StaticDomainContext(ctx))
}
