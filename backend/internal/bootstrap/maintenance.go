package bootstrap

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"unicode"

	"gorm.io/gorm"
)

type MaintenanceOptions struct {
	EnableExam       bool
	MigrationsDir    string
	RunSQLMigrations bool
}

type schemaMigration struct {
	ID         uint      `gorm:"column:id;primaryKey"`
	Version    string    `gorm:"column:migration_version"`
	Name       string    `gorm:"column:migration_name"`
	Checksum   string    `gorm:"column:migration_checksum"`
	Status     string    `gorm:"column:migration_status"`
	ExecutedAt time.Time `gorm:"column:migration_executed_at"`
	CreatedAt  time.Time `gorm:"column:created_at"`
	UpdatedAt  time.Time `gorm:"column:updated_at"`
}

func (schemaMigration) TableName() string { return "schema_migrations" }

func RunMaintenance(options MaintenanceOptions) error {
	if strings.TrimSpace(options.MigrationsDir) == "" {
		options.MigrationsDir = "migrations"
	}

	db, cancel := startupDB(context.Background())
	defer cancel()
	if db == nil {
		return fmt.Errorf("database is not initialized")
	}
	if err := ensureSchemaMigrationsTable(db); err != nil {
		return err
	}

	if err := runOnce(db, "bootstrap:auto_migrate", "GORM AutoMigrate 与兼容迁移", "", func() error {
		return autoMigrate(options.EnableExam)
	}); err != nil {
		return err
	}
	if err := runOnce(db, "bootstrap:seed_setups", "基础系统配置种子", "", seedSetups); err != nil {
		return err
	}
	if err := runOnce(db, "bootstrap:seed_permissions", "权限与菜单种子", "", func() error {
		return seedMenus(options.EnableExam)
	}); err != nil {
		return err
	}
	if options.RunSQLMigrations {
		if err := runVersionedSQLMigrations(options.MigrationsDir, nil); err != nil {
			return err
		}
	}
	return nil
}

func ensureSchemaMigrationsTable(db *gorm.DB) error {
	return db.Exec(`
CREATE TABLE IF NOT EXISTS schema_migrations (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  migration_version VARCHAR(160) NOT NULL,
  migration_name VARCHAR(255) NOT NULL DEFAULT '',
  migration_checksum VARCHAR(64) NOT NULL DEFAULT '',
  migration_status VARCHAR(20) NOT NULL DEFAULT 'success',
  migration_executed_at DATETIME(3) NOT NULL,
  created_at DATETIME(3) DEFAULT NULL,
  updated_at DATETIME(3) DEFAULT NULL,
  PRIMARY KEY (id),
  UNIQUE KEY idx_schema_migrations_version (migration_version)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='版本化迁移执行记录'
`).Error
}

func runOnce(db *gorm.DB, version, name, checksum string, run func() error) error {
	version = strings.TrimSpace(version)
	if version == "" {
		return fmt.Errorf("migration version is required")
	}
	var existing schemaMigration
	err := db.Where("migration_version = ?", version).First(&existing).Error
	if err == nil {
		if checksum != "" && existing.Checksum != "" && existing.Checksum != checksum {
			return fmt.Errorf("migration %s checksum changed", version)
		}
		return nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	if run != nil {
		if err := run(); err != nil {
			return fmt.Errorf("%s: %w", version, err)
		}
	}
	now := time.Now()
	return db.Create(&schemaMigration{
		Version:    version,
		Name:       name,
		Checksum:   checksum,
		Status:     "success",
		ExecutedAt: now,
		CreatedAt:  now,
		UpdatedAt:  now,
	}).Error
}

func runVersionedSQLMigrations(migrationsDir string, logf migrationLogFunc) error {
	db, cancel := startupDB(context.Background())
	defer cancel()
	if db == nil {
		return fmt.Errorf("database is not initialized")
	}
	if err := ensureSchemaMigrationsTable(db); err != nil {
		return err
	}
	files, err := filepath.Glob(filepath.Join(migrationsDir, "*.sql"))
	if err != nil {
		return err
	}
	sort.Strings(files)
	for _, file := range files {
		file := file
		content, err := os.ReadFile(file)
		if err != nil {
			return err
		}
		version := strings.TrimSuffix(filepath.Base(file), filepath.Ext(file))
		checksumBytes := sha256.Sum256(content)
		checksum := hex.EncodeToString(checksumBytes[:])
		if err := runOnce(db, version, filepath.Base(file), checksum, func() error {
			return executeSQLMigration(db, version, string(content))
		}); err != nil {
			return err
		}
		if logf != nil {
			logf("versioned migration checked: " + filepath.Base(file))
		}
	}
	return nil
}

func executeSQLMigration(db *gorm.DB, name, content string) error {
	statements := splitSQLStatements(content)
	return db.Transaction(func(tx *gorm.DB) error {
		for index, statement := range statements {
			if strings.TrimSpace(statement) == "" {
				continue
			}
			if err := tx.Exec(statement).Error; err != nil {
				return fmt.Errorf("%s statement %d: %w", name, index+1, err)
			}
		}
		return nil
	})
}

func splitSQLStatements(input string) []string {
	statements := make([]string, 0)
	var current strings.Builder
	inSingle := false
	inDouble := false
	inBacktick := false
	inLineComment := false
	inBlockComment := false

	for i := 0; i < len(input); i++ {
		ch := input[i]
		var next byte
		if i+1 < len(input) {
			next = input[i+1]
		}

		if inLineComment {
			current.WriteByte(ch)
			if ch == '\n' {
				inLineComment = false
			}
			continue
		}
		if inBlockComment {
			current.WriteByte(ch)
			if ch == '*' && next == '/' {
				current.WriteByte(next)
				i++
				inBlockComment = false
			}
			continue
		}

		if inSingle {
			current.WriteByte(ch)
			if ch == '\\' && next != 0 {
				current.WriteByte(next)
				i++
				continue
			}
			if ch == '\'' {
				if next == '\'' {
					current.WriteByte(next)
					i++
					continue
				}
				inSingle = false
			}
			continue
		}
		if inDouble {
			current.WriteByte(ch)
			if ch == '\\' && next != 0 {
				current.WriteByte(next)
				i++
				continue
			}
			if ch == '"' {
				if next == '"' {
					current.WriteByte(next)
					i++
					continue
				}
				inDouble = false
			}
			continue
		}
		if inBacktick {
			current.WriteByte(ch)
			if ch == '`' {
				inBacktick = false
			}
			continue
		}

		if ch == '-' && next == '-' && isSQLCommentBoundary(input, i+2) {
			current.WriteByte(ch)
			current.WriteByte(next)
			i++
			inLineComment = true
			continue
		}
		if ch == '#' {
			current.WriteByte(ch)
			inLineComment = true
			continue
		}
		if ch == '/' && next == '*' {
			current.WriteByte(ch)
			current.WriteByte(next)
			i++
			inBlockComment = true
			continue
		}
		if ch == '\'' {
			current.WriteByte(ch)
			inSingle = true
			continue
		}
		if ch == '"' {
			current.WriteByte(ch)
			inDouble = true
			continue
		}
		if ch == '`' {
			current.WriteByte(ch)
			inBacktick = true
			continue
		}
		if ch == ';' {
			statement := strings.TrimSpace(current.String())
			if statement != "" {
				statements = append(statements, statement)
			}
			current.Reset()
			continue
		}
		current.WriteByte(ch)
	}
	if statement := strings.TrimSpace(current.String()); statement != "" {
		statements = append(statements, statement)
	}
	return statements
}

func isSQLCommentBoundary(input string, pos int) bool {
	if pos >= len(input) {
		return true
	}
	return unicode.IsSpace(rune(input[pos]))
}
