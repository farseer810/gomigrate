package gomigrate

import (
	"database/sql"
	"fmt"
	"testing"
)

const mysqlTestSource = "root:123456@tcp(127.0.0.1:3306)/gomigrate_test?charset=utf8"

var mysqlTestMigrations = []Migration{
	{
		Name:    "test_table1",
		Content: `create table if not exists test_table1(id int unsigned not null auto_increment, data text not null, primary key(id))`,
	},
	{
		Name:    "test_table2",
		Content: "create table if not exists test_table2(id int unsigned not null auto_increment, data text not null, primary key(id))",
	},
	{
		Name:    "test_table3",
		Content: "create table if not exists test_table3(id int unsigned not null auto_increment, data text not null, primary key(id))",
	},
	{
		Name:    "test_table4",
		Content: "create table if not exists test_table4(id int unsigned not null auto_increment, data text not null, primary key(id))",
	},
	{
		Name:    "test_table5",
		Content: "create table if not exists test_table5(id int unsigned not null auto_increment, data text not null, primary key(id))",
	},
}

func insertMySQLTestdata(t *testing.T) (*mysqlMigrateExecutor, *sql.DB, func()) {
	// 创建Schema History表
	executor := NewMySQLMigrateExecutor(mysqlTestSource)
	err := executor.InitSchemaHistoryTable()
	if err != nil {
		t.Log("using mysql database source: " + mysqlTestSource)
		t.Error(err)
		t.FailNow()
	}

	// 连接数据库
	mysqlExecutor := executor.(*mysqlMigrateExecutor)
	db, err := mysqlExecutor.connectDB()
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	// 添加测试数据
	for i, testMigration := range mysqlTestMigrations {
		err = mysqlExecutor.addSchemaHistory(db, i+1, testMigration)
		if err != nil {
			t.Error(err)
			t.FailNow()
		}
	}

	// 结束释放资源
	clearFunc := func() {
		db.Exec(fmt.Sprintf("DROP TABLE `%s`", mysqlExecutor.GetSchemaHistoryTableName()))
		db.Exec("DROP TABLE `test_test1`")
		db.Exec("DROP TABLE `test_test2`")
		db.Exec("DROP TABLE `test_test3`")
		db.Exec("DROP TABLE `test_test4`")
		db.Exec("DROP TABLE `test_test5`")
		db.Close()
	}
	return mysqlExecutor, db, clearFunc
}

func TestGetSchemaHistories(t *testing.T) {
	executor, db, clear := insertMySQLTestdata(t)
	defer clear()

	// 获取数据并检验
	schemaHistories, err := executor.getSchemaHistories(db)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	if len(schemaHistories) != len(mysqlTestMigrations) {
		t.FailNow()
	}
	for i, testMigration := range mysqlTestMigrations {
		schemaHistory := schemaHistories[i]
		if schemaHistory.Rank != i+1 || schemaHistory.Name != testMigration.Name || schemaHistory.Content != testMigration.Content {
			t.FailNow()
		}
	}

	// 测试 Schema History不完整性检查 是否正常
	_, err = db.Exec(fmt.Sprintf("DELETE FROM `%s` where `rank` in (1, 3)", executor.GetSchemaHistoryTableName()))
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	schemaHistories, err = executor.getSchemaHistories(db)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	for _, schemaHistory := range schemaHistories {
		testMigration := mysqlTestMigrations[schemaHistory.Rank-1]
		if schemaHistory.Name != testMigration.Name || schemaHistory.Content != testMigration.Content {
			t.FailNow()
		}
	}
}

func TestMySQLShowMigrations(t *testing.T) {
	var err error
	executor := NewMySQLMigrateExecutor(mysqlTestSource)
	err = executor.InitSchemaHistoryTable()
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	mysqlExecutor := executor.(*mysqlMigrateExecutor)
	db, err := mysqlExecutor.connectDB()
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	defer db.Close()
	defer func() {
		db.Exec(fmt.Sprintf("DROP TABLE `%s`", executor.GetSchemaHistoryTableName()))
	}()

	executor.SetMigrations(mysqlTestMigrations)
	err = executor.ShowMigrations()
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
}

func TestMySQLInstallMigrations(t *testing.T) {
	executor := NewMySQLMigrateExecutor(mysqlTestSource)
	err := executor.InitSchemaHistoryTable()
	if err != nil {
		t.Log("using mysql database source: " + mysqlTestSource)
		t.Error(err)
		t.FailNow()
	}

	// 连接数据库
	mysqlExecutor := executor.(*mysqlMigrateExecutor)
	db, err := mysqlExecutor.connectDB()
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	defer db.Close()
	defer func() {
		db.Exec(fmt.Sprintf("DROP TABLE `%s`", executor.GetSchemaHistoryTableName()))
		db.Exec(fmt.Sprintf("DROP TABLE `test_table1`"))
		db.Exec(fmt.Sprintf("DROP TABLE `test_table2`"))
		db.Exec(fmt.Sprintf("DROP TABLE `test_table3`"))
		db.Exec(fmt.Sprintf("DROP TABLE `test_table4`"))
		db.Exec(fmt.Sprintf("DROP TABLE `test_table5`"))
	}()

	executor.SetMigrations(mysqlTestMigrations)

	err = executor.InstallMigrations()
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
}
