package gomigrate

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
)

type mysqlMigrateExecutor struct {
	BaseExecutor
	connSource string
	migrations []Migration
}

func NewMySQLMigrateExecutor(connSource string) MigrationExecutor {
	db, err := sql.Open("mysql", connSource)
	if err != nil {
		panic(err)
	}
	db.SetConnMaxLifetime(time.Minute * 5)
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	return &mysqlMigrateExecutor{
		connSource: connSource,
	}
}

func (m *mysqlMigrateExecutor) connectDB() (*sql.DB, error) {
	db, err := sql.Open("mysql", m.connSource)
	if err != nil {
		return nil, err
	}
	db.SetConnMaxLifetime(time.Minute * 5)
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	return db, nil
}

func (m *mysqlMigrateExecutor) SetMigrations(migrations []Migration) {
	m.migrations = make([]Migration, len(migrations))
	for i := range migrations {
		m.migrations[i] = migrations[i]
	}
}

func (m *mysqlMigrateExecutor) InitSchemaHistoryTable() error {
	db, err := m.connectDB()
	if err != nil {
		return err
	}
	defer db.Close()

	// 建表
	createTableSQL := strings.Join([]string{
		fmt.Sprintf("CREATE TABLE IF NOT EXISTS `%s`(", m.GetSchemaHistoryTableName()),
		"`rank` INT(11) NOT NULL COMMENT 'rank',",
		"`name` VARCHAR(156) NOT NULL COMMENT 'schema name',",
		"`content` TEXT COMMENT 'schema content',",
		"`installed_time` DATETIME NOT NULL COMMENT 'installed time',",
		"PRIMARY KEY(`rank`),",
		"UNIQUE KEY `uniq_idx_name`(`name`)",
		") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT 'DO NOT touch this unless you know what you are doing';",
	}, "\n")
	_, err = db.Exec(createTableSQL)
	return err
}

func (m *mysqlMigrateExecutor) addSchemaHistory(db *sql.DB, rank int, migration Migration) error {
	_, err := db.Exec(fmt.Sprintf("INSERT INTO `%s`(`rank`, `name`, `content`, `installed_time`) VALUES(?, ?, ?, ?)", m.GetSchemaHistoryTableName()),
		rank,
		migration.Name,
		migration.Content,
		time.Now(),
	)
	return err
}

func (m *mysqlMigrateExecutor) getSchemaHistories(db *sql.DB) ([]SchemaHistory, error) {
	rows, err := db.Query(fmt.Sprintf("SELECT `rank`, `name`, `content`, `installed_time` FROM `%s` ORDER BY `rank` ASC", m.GetSchemaHistoryTableName()))
	if err != nil {
		return nil, err
	}

	schemaHistories := make([]SchemaHistory, 0)

	var installedTimeStr string
	for rows.Next() {
		schemaHistory := SchemaHistory{}
		rows.Scan(&schemaHistory.Rank, &schemaHistory.Name, &schemaHistory.Content, &installedTimeStr)
		if err != nil {
			return nil, err
		}
		installedTime, err := time.ParseInLocation("2006-01-02 15:04:05", installedTimeStr, time.Local)
		if err != nil {
			return nil, err
		}
		schemaHistory.InstalledTime = installedTime
		schemaHistories = append(schemaHistories, schemaHistory)
	}
	defer rows.Close()

	return schemaHistories, err
}

func (m *mysqlMigrateExecutor) CheckMigrations() (err error) {
	// 检查存不存在重复的Migration名称
	if m.migrations != nil {
		migrationNameSet := make(map[string]int)
		for _, migration := range m.migrations {
			if migrationNameSet[migration.Name] == 0 {
				migrationNameSet[migration.Name] = 1

			} else if migrationNameSet[migration.Name] == 1 {
				if err == nil {
					err = fmt.Errorf("%w: %s", ErrDuplicatedMigrationName, migration.Name)
				} else {
					err = fmt.Errorf("%w, %s", err, migration.Name)
				}
				migrationNameSet[migration.Name] = 2
			}
		}
		if err != nil {
			return err
		}
	}

	err = m.InitSchemaHistoryTable()
	if err != nil {
		return err
	}

	db, err := m.connectDB()
	if err != nil {
		return err
	}
	defer db.Close()

	schemaHistories, err := m.getSchemaHistories(db)
	if len(schemaHistories) == 0 {
		return nil
	}

	for i, schemaHistory := range schemaHistories {
		if schemaHistory.Rank != i+1 {
			return ErrBrokenSchemaHistory
		}
	}

	if len(schemaHistories) > len(m.migrations) {
		return ErrMigrationMissing
	}

	for i, schemaHistory := range schemaHistories {
		if schemaHistory.Name != m.migrations[i].Name || schemaHistory.Content != m.migrations[i].Content {
			return ErrMigrationModified
		}
	}
	return nil
}

func (m *mysqlMigrateExecutor) ShowMigrations() error {
	err := m.InitSchemaHistoryTable()
	if err != nil {
		return err
	}
	db, err := m.connectDB()
	if err != nil {
		return err
	}
	defer db.Close()

	schemaHistories, err := m.getSchemaHistories(db)
	if err != nil && !errors.Is(err, ErrBrokenSchemaHistory) {
		return err
	}

	t := table.NewWriter()
	t.AppendHeader(table.Row{"Rank", "Schema Name", "Migration Name", "Installed Time", "Status"})

	colorError := text.FgRed
	colorSuccess := text.FgGreen

	schemaHistoryMap := make(map[int]*SchemaHistory)
	maxRank := 0
	for i, schemaHistory := range schemaHistories {
		if maxRank < schemaHistory.Rank {
			maxRank = schemaHistory.Rank
		}
		schemaHistoryMap[schemaHistory.Rank] = &schemaHistories[i]
	}
	// 计算需要遍历的最大值
	maxLen := maxRank
	if maxLen < len(m.migrations) {
		maxLen = len(m.migrations)
	}

	migrateInfos := make([]struct {
		SchemaHistory *SchemaHistory
		Migration     *Migration
		Status        MigrateStatus
	}, maxLen)
	for i := 0; i < maxLen; i++ {
		migrateInfos[i].SchemaHistory = schemaHistoryMap[i+1]
		if i < len(m.migrations) {
			migrateInfos[i].Migration = &m.migrations[i]
		}
		if i+1 <= maxRank {
			if i < len(m.migrations) {
				if schemaHistoryMap[i+1] == nil {
					migrateInfos[i].Status = StatusBrokenSchemaHistory
				} else {
					if schemaHistoryMap[i+1].Name != m.migrations[i].Name || schemaHistoryMap[i+1].Content != m.migrations[i].Content {
						migrateInfos[i].Status = StatusMigrationModified
					} else {
						migrateInfos[i].Status = StatusInstalled
					}
				}
			} else {
				migrateInfos[i].Status = StatusMigrationMissing
			}

		} else {
			migrateInfos[i].Status = StatusReadyToInstall
		}
	}

	hasMissingMigration := false
	hasModifiedMigration := false
	hasBrokenSchema := false
	for _, migrateInfo := range migrateInfos {
		rankText := "-"
		schemaNameText := ""
		migrationNameText := ""
		installedTimeText := "-"

		statusText := colorError.Sprint("UNKNOWN")
		if migrateInfo.Status == StatusInstalled {
			statusText = colorSuccess.Sprint("INSTALLED")
		} else if migrateInfo.Status == StatusReadyToInstall {
			statusText = colorSuccess.Sprint("READY TO INSTALL")
		} else if migrateInfo.Status == StatusMigrationMissing {
			hasMissingMigration = true
			statusText = colorError.Sprint("MIGRATION MISSING")
		} else if migrateInfo.Status == StatusMigrationModified {
			hasModifiedMigration = true
			statusText = colorError.Sprint("MIGRATION MODIFIED")
		} else if migrateInfo.Status == StatusBrokenSchemaHistory {
			hasBrokenSchema = true
			statusText = colorError.Sprint("SCHEMA BROKEN")
		}
		if migrateInfo.SchemaHistory != nil {
			rankText = fmt.Sprint(migrateInfo.SchemaHistory.Rank)
			schemaNameText = migrateInfo.SchemaHistory.Name
			installedTimeText = migrateInfo.SchemaHistory.InstalledTime.Format("2006-01-02 15:04:05")
		}
		if migrateInfo.Migration != nil {
			migrationNameText = migrateInfo.Migration.Name
		}
		t.AppendRow([]interface{}{rankText, schemaNameText, migrationNameText, installedTimeText, statusText})
	}

	helpTips := "all is well"
	if hasMissingMigration || hasModifiedMigration || hasBrokenSchema {
		tips := make([]string, 0)
		if hasMissingMigration {
			tips = append(tips, "to fix MIGRATION MISSING: provide the missing migrations")
		}
		if hasModifiedMigration {
			tips = append(tips, "to fix installed MIGRATION MODIFIED: recovery the installed but modified migrations. \n\tPlease DO NOT modify installed migrations")
		}
		if hasBrokenSchema {
			tips = append(tips, "to fix SCHEMA BROKEN: no cure (yet?:))")
		}

		for i := range tips {
			tips[i] = fmt.Sprintf("(%d) %s", i+1, tips[i])
		}
		helpTips = strings.Join(tips, "\n")
	}

	fmt.Println(t.Render())
	fmt.Println(helpTips)
	return nil
}

func (m *mysqlMigrateExecutor) InstallMigrations() error {
	err := m.CheckMigrations()
	if err != nil {
		return err
	}

	db, err := m.connectDB()
	if err != nil {
		return err
	}
	defer db.Close()
	schemaHistories, err := m.getSchemaHistories(db)
	if err != nil {
		return err
	}
	uninstallMigrations := make([]Migration, 0)
	if m.migrations != nil {
		uninstallMigrations = m.migrations[len(schemaHistories):]
	}

	baseRank := len(schemaHistories)
	for i, uninstallMigration := range uninstallMigrations {
		_, err = db.Exec(uninstallMigration.Content)
		if err != nil {
			return err
		}
		err = m.addSchemaHistory(db, baseRank+i+1, uninstallMigration)
		if err != nil {
			return err
		}
	}

	return nil
}
