package gomigrate

import (
	"embed"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

const (
	flywayTestdataDirPath = "testdata_flyway"
	flywayEmbedTestdataDirPath = "testdata/embed_flyway/test1/test2"
	validTestdataNum = 5
)

var flywayTestdatas = []flywayTestdata{
	{
		Version: "V1",
		Name:    "test_table1",
		Content: `create table if not exists test_table1(id int unsigned not null auto_increment, data text not null, primary key(id))`,
	},
	{
		Version: "V1_1",
		Name:    "test_table2",
		Content: "create table if not exists test_table2(id int unsigned not null auto_increment, data text not null, primary key(id))",
	},
	{
		Version: "v1_2",
		Name:    "test__table3",
		Content: "create table if not exists test_table3(id int unsigned not null auto_increment, data text not null, primary key(id))",
	},
	{
		Version: "V2",
		Name:    "test___table4",
		Content: "create table if not exists test_table4(id int unsigned not null auto_increment, data text not null, primary key(id))",
	},
	{
		Version: "v3",
		Name:    "test_table__5",
		Content: "create table if not exists test_table5(id int unsigned not null auto_increment, data text not null, primary key(id))",
	},
	{
		Version: "4_1",
		Name:    "test_table__5",
		Content: "create table if not exists test_table5(id int unsigned not null auto_increment, data text not null, primary key(id))",
	},
	{
		Version: "v4.2",
		Name:    "test_table__5",
		Content: "create table if not exists test_table5(id int unsigned not null auto_increment, data text not null, primary key(id))",
	},
}

//go:embed testdata
var testdataFS embed.FS

type flywayTestdata struct {
	Version string
	Name    string
	Content string
}

func TestGetMigrationsFromFlywayDir(t *testing.T) {
	err := os.MkdirAll(flywayTestdataDirPath, 0777)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	defer os.RemoveAll(flywayTestdataDirPath)

	for _, testdata := range flywayTestdatas {
		filename := testdata.Version + "__" + testdata.Name + ".sql"
		err = ioutil.WriteFile(filepath.Join(flywayTestdataDirPath, filename), []byte(testdata.Content), 0777)
		if err != nil {
			t.Error(err)
			t.FailNow()
		}
	}

	migrations, err := GetMigrationsFromFlywayDir(flywayTestdataDirPath)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if len(migrations) != validTestdataNum {
		t.FailNow()
	}
	for i := 0; i < validTestdataNum; i++ {
		testdata := flywayTestdatas[i]
		if migrations[i].Name != testdata.Version+"__"+testdata.Name+".sql" || migrations[i].Content != testdata.Content {
			t.FailNow()
		}
	}
}

func TestGetMigrationsFromFlywayEmbedFS(t *testing.T) {
	migrations, err := GetMigrationsFromFlywayEmbedFS(testdataFS, flywayEmbedTestdataDirPath)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if len(migrations) != validTestdataNum {
		t.FailNow()
	}
	for i := 0; i < validTestdataNum; i++ {
		testdata := flywayTestdatas[i]
		if migrations[i].Name != testdata.Version+"__"+testdata.Name+".sql" || migrations[i].Content != testdata.Content {
			t.FailNow()
		}
	}
}