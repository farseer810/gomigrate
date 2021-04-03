package gomigrate

import (
	"errors"
	"fmt"
)

type MigrateStatus int

const (
	DefaultSchemaHistoryTableName = "gomigrate_schema_history"

	StatusUnknown MigrateStatus = iota
	StatusInstalled
	StatusReadyToInstall
	StatusMigrationMissing
	StatusMigrationModified
	StatusBrokenSchemaHistory
)

var (
	ErrInitializeFail          = errors.New("failed to initialize")
	ErrBrokenSchemaHistory     = errors.New("broken schema history detected")
	ErrDuplicatedMigrationName = errors.New("duplicated migration name detected")
	ErrInvalidMigrations       = errors.New("invalid migrations")
	ErrMigrationMissing        = fmt.Errorf("%w(missing)", ErrInvalidMigrations)
	ErrMigrationModified       = fmt.Errorf("%w(modified)", ErrInvalidMigrations)
)
