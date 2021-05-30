package gomigrate

import (
	"embed"
	"io/fs"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"sort"
)

var reValidFlywayFilename = regexp.MustCompile("(?i)^v(\\d+(_\\d+)*)__(.+)\\.sql$")

func GetMigrationsFromFlywayDir(sourcePath string) ([]Migration, error) {
	sortableMigrations := make(SortableMigrations, 0)
	err := filepath.WalkDir(sourcePath, func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}
		baseFilename := filepath.Base(path)
		matches := reValidFlywayFilename.FindStringSubmatch(baseFilename)
		if len(matches) == 0 {
			return nil
		}

		versionStr := matches[1]
		migrationVersion, err := ParseMigrationVersion(versionStr)
		if err != nil {
			return err
		}

		content, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}

		migration := &Migration{
			Name:    matches[0],
			Content: string(content),
		}
		sortableMigrations = append(sortableMigrations, &SortableMigration{
			M:       migration,
			Version: migrationVersion,
		})

		return nil
	})
	if err != nil {
		return nil, err
	}

	sort.Sort(sortableMigrations)
	migrations := make([]Migration, len(sortableMigrations))
	for i, sortableMigration := range sortableMigrations {
		migrations[i] = *sortableMigration.M
	}
	return migrations, nil
}

func GetMigrationsFromFlywayEmbedFS(embedFS embed.FS, subDirPath string) ([]Migration, error) {
	entries, err := embedFS.ReadDir(subDirPath)
	if err != nil {
		return nil, err
	}

	sortableMigrations := make(SortableMigrations, 0)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		matches := reValidFlywayFilename.FindStringSubmatch(entry.Name())
		if len(matches) == 0 {
			continue
		}

		versionStr := matches[1]
		migrationVersion, err := ParseMigrationVersion(versionStr)
		if err != nil {
			return nil, err
		}

		entryPath := subDirPath + "/" + entry.Name()
		if subDirPath == "." {
			entryPath = entry.Name()
		}
		content, err := embedFS.ReadFile(entryPath)
		if err != nil {
			return nil, err
		}
		migration := &Migration{
			Name:    matches[0],
			Content: string(content),
		}
		sortableMigrations = append(sortableMigrations, &SortableMigration{
			M:       migration,
			Version: migrationVersion,
		})
	}
	sort.Sort(sortableMigrations)
	migrations := make([]Migration, len(sortableMigrations))
	for i, sortableMigration := range sortableMigrations {
		migrations[i] = *sortableMigration.M
	}
	return migrations, nil
}