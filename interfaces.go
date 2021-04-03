package gomigrate

type MigrationExecutor interface {
	GetSchemaHistoryTableName() string
	SetSchemaHistoryTableName(tableName string) error
	SetMigrations(migrations []Migration)
	InitSchemaHistoryTable() error
	ShowMigrations() error
	InstallMigrations() error
}

type BaseExecutor struct {
	tableName string
}

func (b *BaseExecutor) GetSchemaHistoryTableName() string {
	if b.tableName == "" {
		return DefaultSchemaHistoryTableName
	}
	return b.tableName
}

func (b *BaseExecutor) SetSchemaHistoryTableName(tableName string) error {
	b.tableName = tableName
	return nil
}
