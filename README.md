#Go SQL Migration Library
## How To Use
### basic usage
1, download
```shell
go get github.com/farseer810/gomigrate
```
2, get the migration executor you need(currently support MySQL)
```go
executor := NewMySQLMigrateExecutor("user:password@tcp(host:port)/your_db?charset=utf8")
```
3, install migrations
```go
migrations := []Migration{
  {
    Name:    "test_table1",
    Content: `create table if not exists test_table1(data text not null)`,
  },
  {
    Name:    "test_table2",
    Content: "create table if not exists test_table2(data text not null)",
  }
}
executor.SetMigrations(migrations)
executor.InstallMigrations()
```

### show migrations
This idea comes from Django Web Framework. It shows problems in your migrations and schema table.
```go
executor.SetMigrations(migrations)
executor.ShowMigrations()
```
and it prints out something like this:
```
+------+-------------+----------------+---------------------+--------------------+
| RANK | SCHEMA NAME | MIGRATION NAME | INSTALLED TIME      | STATUS             |
+------+-------------+----------------+---------------------+--------------------+
| 1    | test_table1 | test_table1    | 2021-04-03 07:52:58 | INSTALLED          |
| 2    | test_table2 | test_table2    | 2021-04-03 07:52:58 | INSTALLED          |
| -    |             | test_table3    | -                   | SCHEMA BROKEN      |
| 4    | test_table3 | test_table4    | 2021-04-03 07:52:58 | MIGRATION MODIFIED |
| 5    | test_table4 |                | 2021-04-03 07:52:58 | MIGRATION MISSING  |
| 6    | test_table5 |                | 2021-04-03 07:52:58 | MIGRATION MISSING  |
+------+-------------+----------------+---------------------+--------------------+
(1) to fix MIGRATION MISSING: provide the missing migrations
(2) to fix installed MIGRATION MODIFIED: recovery the installed but modified migrations. 
	Please DO NOT modify installed migrations
(3) to fix SCHEMA BROKEN: no cure (yet?:))
```

## Flyway Style Migrations
We provide a way to parse flyway style migrations from file system or embed.FS
```go
migrations, err := GetMigrationsFromFlywayDir("flyway_migrations_dir")
executor.SetMigrations(migrations)
```
or 
```go
//go:embed flyway_migrations_dir
var embedFS embed.FS

migrations, err := GetMigrationsFromFlywayEmbedFS(embedFS, flyway_migrations_dir)
executor.SetMigrations(migrations)
```

### what is flyway style migrations
, there's an example:
```
V1__test_table1_migration_name.sql
V2__test_table4_migration_name.sql
v1_1__test_table2_migration_name.sql
V1_2__test_table3_migration_name.sql
v3_1__test_table5_migration_name.sql
```
To summarize, a valid migration file name must fit the following requirements:
* file name must start with letter 'V', case doesn't matter
* what comes after the leading letter 'V' is **version number**, which is either integer, or a decimal with its 
  floating point replaced by underscore. **version number** determines the execution order of this migration
* an double underscore separates version number and **migration name**, **both of them must be unique**
* file name must end with ".sql", doesn't matter

So in the above example, the migrations will be executed in this order:
```
V1__test_table1_migration_name.sql
v1_1__test_table2_migration_name.sql
V1_2__test_table3_migration_name.sql
V2__test_table4_migration_name.sql
v3_1__test_table5_migration_name.sql
```

## Ground Rules
