package gomigrate

import (
	"embed"
	"io/fs"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

var reValidFlywayFilename = regexp.MustCompile("(?i)^v(\\d+(_\\d+)?)__(.+)\\.sql$")

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
		version, err := strconv.ParseFloat(strings.Replace(versionStr, "_", ".", 1), 64)
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
			M:      migration,
			Weight: version,
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
		version, err := strconv.ParseFloat(strings.Replace(versionStr, "_", ".", 1), 64)
		if err != nil {
			return nil, err
		}

		content, err := embedFS.ReadFile(subDirPath + "/" + entry.Name())
		if err != nil {
			return nil, err
		}
		migration := &Migration{
			Name:    matches[0],
			Content: string(content),
		}
		sortableMigrations = append(sortableMigrations, &SortableMigration{
			M:      migration,
			Weight: version,
		})
	}
	sort.Sort(sortableMigrations)
	migrations := make([]Migration, len(sortableMigrations))
	for i, sortableMigration := range sortableMigrations {
		migrations[i] = *sortableMigration.M
	}
	return migrations, nil
}