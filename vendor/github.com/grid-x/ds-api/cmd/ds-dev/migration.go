package main

import (
	"fmt"
	"path/filepath"
	"time"
)

type MigrationGenerator struct {
	repositoryBasePath string
	name               string
	attributes         []Attribute
}

func NewMigrationGenerator(repPath, name, attributes string) (*MigrationGenerator, error) {
	attrs, err := attributesFromString(attributes)
	if err != nil {
		return nil, err
	}
	return &MigrationGenerator{
		repositoryBasePath: repPath,
		name:               name,
		attributes:         attrs,
	}, nil
}

func (g *MigrationGenerator) GetGenFiles() ([]string, map[string]GenFile) {
	migrationFileName := fmt.Sprintf("%s_create_table_%s.sql", time.Now().Format("2006-01-02_1504"), underscore(g.name))
	fp := filepath.Join(g.repositoryBasePath, "migrations", "sql")
	return []string{fp}, map[string]GenFile{
		filepath.Join(fp, migrationFileName): migrationSQLImpl{
			Name:       g.name,
			Attributes: g.attributes,
		},
	}
}
