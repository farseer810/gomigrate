package gomigrate

import (
	"crypto/sha1"
	"encoding/hex"
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

func (m *Migration) GetContentHash() string {
	sum := sha1.Sum([]byte(m.Content))
	return hex.EncodeToString(sum[:])
}

type SortableMigration struct {
	M      *Migration
	Weight float64
}

type SortableMigrations []*SortableMigration

func (s SortableMigrations) Len() int {
	return len(s)
}

func (s SortableMigrations) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s SortableMigrations) Less(i, j int) bool {
	return s[i].Weight < s[j].Weight
}
