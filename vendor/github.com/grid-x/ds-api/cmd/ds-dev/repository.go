package main

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"text/template"
	"time"
)

var (
	untitle = func(s string) string {
		if len(s) < 1 {
			return s
		}
		return strings.ToLower(s[0:1]) + s[1:len(s)]
	}

	repositoryDefaultFuncMap = template.FuncMap{
		"untitle":    untitle,
		"title":      strings.Title,
		"tolower":    strings.ToLower,
		"dbColumn":   underscore,
		"jsonField":  untitle,
		"counter":    counter,
		"underscore": underscore,
		"dbType": func(t string) string {
			switch t {
			case "string":
				return "TEXT"
			case "int", "int32", "uint32":
				return "INT"
			case "int64", "uint64":
				return "BIGINT"
			case "time.Time":
				return "TIMESTAMP WITH TIME ZONE DEFAULT now()"
			case "*time.Time":
				return "TIMESTAMP WITH TIME ZONE"
			default:
				return "TODO -- unsupported"
			}
			return ""
		},
	}
)

type Attribute struct {
	Name, Type string
}

func attributesFromString(s string) ([]Attribute, error) {
	result := []Attribute{
		{
			Name: "UUID",
			Type: "string",
		},
		{
			Name: "CreatedAt",
			Type: "time.Time",
		},
		{
			Name: "UpdatedAt",
			Type: "time.Time",
		},
		{
			Name: "DeletedAt",
			Type: "*time.Time",
		},
	}

	for _, a := range strings.Split(s, ",") {
		if len(a) == 0 {
			continue
		}
		nameAndType := strings.Split(a, ":")
		if len(nameAndType) != 2 {
			return nil, fmt.Errorf("invalid argument: %s", s)
		}

		result = append(result, Attribute{
			Name: nameAndType[0],
			Type: nameAndType[1],
		})
	}

	return result, nil
}

type repositoryGoImpl struct {
	Plural, Singular string
	Attributes       []Attribute
}

func (r repositoryGoImpl) Gen(w io.Writer) (bool, error) {
	data, err := Asset("templates/repository.go.tmpl")
	if err != nil {
		return true, err
	}
	t, err := template.New("repositoryGoImpl").Funcs(repositoryDefaultFuncMap).Parse(string(data))
	if err != nil {
		return true, err
	}

	return true, t.Execute(w, r)
}

type repositoryTestGoImpl struct {
	Singular, Plural string
}

func (r repositoryTestGoImpl) Gen(w io.Writer) (bool, error) {
	data, err := Asset("templates/repository_test.go.tmpl")
	if err != nil {
		return true, err
	}
	t, err := template.New("repositoryTestGoImpl").Funcs(repositoryDefaultFuncMap).Parse(string(data))
	if err != nil {
		return true, err
	}

	return true, t.Execute(w, r)
}

type modelGoImpl struct {
	Name       string
	Attributes []Attribute
}

func (m modelGoImpl) Gen(w io.Writer) (bool, error) {
	data, err := Asset("templates/model.go.tmpl")
	if err != nil {
		return true, err
	}
	t, err := template.New("modelGoImpl").Funcs(repositoryDefaultFuncMap).Parse(string(data))
	if err != nil {
		return true, err
	}

	return true, t.Execute(w, m)
}

type migrationSQLImpl struct {
	Name       string
	Attributes []Attribute
}

func (m migrationSQLImpl) Gen(w io.Writer) (bool, error) {
	data, err := Asset("templates/migration.sql.tmpl")
	if err != nil {
		return false, err
	}
	t, err := template.New("migrationGoImpl").Funcs(repositoryDefaultFuncMap).Parse(string(data))
	if err != nil {
		return false, err
	}

	return false, t.Execute(w, m)
}

type RepositoryGenerator struct {
	repositoryBasePath string
	modelBasePath      string
	plural, singular   string
	filename           string
	attributes         []Attribute

	includeMigration, includeModel, includeTests bool
}

func NewRepositoryGenerator(repositoryBasePath, modelBasePath, plural, singular, filename, attrs string, includeMigration, includeModel, includeTests bool) (*RepositoryGenerator, error) {

	attributes, err := attributesFromString(attrs)
	if err != nil {
		return nil, err
	}
	return &RepositoryGenerator{
		repositoryBasePath: repositoryBasePath,
		modelBasePath:      modelBasePath,
		plural:             plural,
		singular:           singular,
		filename:           filename,
		attributes:         attributes,
		includeMigration:   includeMigration,
		includeModel:       includeModel,
		includeTests:       includeTests,
	}, nil
}

func (r *RepositoryGenerator) GetGenFiles() ([]string, map[string]GenFile) {
	migrationFileName := fmt.Sprintf("%s_create_table_%s.sql", time.Now().Format("2006-01-02_1504"), underscore(r.plural))
	dirs := []string{r.modelBasePath, filepath.Join(r.repositoryBasePath, "migrations", "sql")}
	genfiles := map[string]GenFile{
		filepath.Join(r.repositoryBasePath, r.filename): repositoryGoImpl{
			Plural:     r.plural,
			Singular:   r.singular,
			Attributes: r.attributes,
		},
	}

	if r.includeTests {
		genfiles[filepath.Join(r.repositoryBasePath, r.filename[0:len(r.filename)-3]+"_test.go")] = repositoryTestGoImpl{
			Plural:   r.plural,
			Singular: r.singular,
		}
	}

	if r.includeModel {
		genfiles[filepath.Join(r.modelBasePath, r.filename)] = modelGoImpl{
			Name:       r.singular,
			Attributes: r.attributes,
		}
	}
	if r.includeMigration {
		genfiles[filepath.Join(r.repositoryBasePath, "migrations", "sql", migrationFileName)] = migrationSQLImpl{
			Name:       r.plural,
			Attributes: r.attributes,
		}
	}

	return dirs, genfiles
}
