package gomigrate

import (
	"crypto/sha1"
	"encoding/hex"
	"strconv"
	"strings"
	"time"
)

type SchemaHistory struct {
	Migration
	Rank          int
	InstalledTime time.Time
}

type Migration struct {
	Name    string
	Content string
}

type MigrationVersion []int

func ParseMigrationVersion(version string) (MigrationVersion, error) {
	versions := strings.Split(strings.ReplaceAll(version, "_", "."), ".")
	migrationVersion := make(MigrationVersion, 0)
	for _, version := range versions {
		versionInt, err := strconv.ParseInt(version, 10, 64)
		if err != nil {
			return nil, err
		}
		migrationVersion = append(migrationVersion, int(versionInt))
	}
	return migrationVersion, nil
}

func (m MigrationVersion) String() string {
	versions := make([]string, 0, len(m))
	for _, version := range m {
		versions = append(versions, strconv.FormatInt(int64(version), 10))
	}
	return strings.Join(versions, ".")
}

func (m *Migration) GetContentHash() string {
	sum := sha1.Sum([]byte(m.Content))
	return hex.EncodeToString(sum[:])
}

type SortableMigration struct {
	M       *Migration
	Version []int
}

type SortableMigrations []*SortableMigration

func (s SortableMigrations) Len() int {
	return len(s)
}

func (s SortableMigrations) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s SortableMigrations) Less(i, j int) bool {
	minVersionLen := len(s[i].Version)
	if minVersionLen > len(s[j].Version) {
		minVersionLen = len(s[j].Version)
	}
	for t := 0; t < minVersionLen; t++ {
		if s[i].Version[t] < s[j].Version[t] {
			return true
		} else if s[i].Version[t] > s[j].Version[t] {
			return false
		}
	}
	if len(s[i].Version) < len(s[j].Version) {
		return true
	}
	return false
}
