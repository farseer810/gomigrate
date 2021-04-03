package gomigrate

import "testing"

func TestMigrationTableName(t *testing.T) {
	executor := &BaseExecutor{}
	if executor.GetSchemaHistoryTableName() != DefaultSchemaHistoryTableName {
		t.FailNow()
	}
	err := executor.SetSchemaHistoryTableName("testtable")
	if err != nil {
		t.FailNow()
	}
	if executor.GetSchemaHistoryTableName() != "testtable" {
		t.FailNow()
	}
}
