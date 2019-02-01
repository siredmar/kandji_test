package main

import (
	"bufio"
	"fmt"
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"text/template"
	"time"
)

var (
	routesDefaultFuncMap = template.FuncMap{
		"packageName": func(input string) string {
			return strings.Replace(input, "-", "", -1)
		},
		"title": strings.Title,
		"untitle": func(s string) string {
			if len(s) < 1 {
				return s
			}
			return strings.ToLower(s[0:1]) + s[1:len(s)]
		},
	}
)

// RoutesGenerator generates routes based on the package structure.
type RoutesGenerator struct {
	// Type is the API type such as management or device.
	Type string
	// Output is the name of the file.
	Output string
	// VersionImports is a list of versioned services which also includes subpackages.
	VersionImports []Import
	// Services is a list of versioned services.
	Services []Service
	// Middlewares is a list of all middlewares needed by the handlers
	Middlewares []string
	// tmpl contains the template to generate the routes files.
	tmpl *template.Template
}

// Import represents the package imports.
type Import struct {
	// Version is the version of the service imported.
	Version string
	// Name is the local package name.
	Name string
	// Path is the import path.
	Path string
}

// Service represents a versioned service.
type Service struct {
	// Version is the version of the service.
	Version string
	// Name is the local name of the service instance.
	Name string
	// PkgName is local package name.
	PkgName string
	// Imports are the imports of this service.
	Imports []Import
	// Groups are the different handler groups such as registry, authentication,
	// etc.
	Groups []Group
}

// Group represents a set of endpoints.
type Group struct {
	// Version is the version of the service.
	Version string
	// Name is the name of the group.
	Name string
	// PkgName is local package name.
	PkgName string
	// Endpoints are the different endpoints available in that group.
	Endpoints []Endpoint
}

// Endpoint represents a HTTP REST endpoint.
type Endpoint struct {
	// PkgName is local package name.
	PkgName string
	// Action represent the endpoint action used by the access management.
	Action string
	// Resource represents the access management resource.
	Resource string
	// Path represents the URL path.
	Path string
	// Methods represents the allowed HTTP methods.
	Methods []string
	// ExpectBody reports if a request body needs to be read.
	ExpectBody bool
	// RequestStruct is the name of the request struct.
	RequestStruct string
	// HandlerName represents the called handler
	HandlerName string
	// Middlewares are the middlewares to use for this endpoint
	Middlewares []string
}

// NewRoutesGenerator generates the routes.
func NewRoutesGenerator(basePath, subRoute, since, output string) (*RoutesGenerator, error) {
	data, err := Asset("templates/routes.go.tmpl")
	if err != nil {
		return nil, err
	}
	g := &RoutesGenerator{
		Output: output,
		Type:   subRoute,
	}
	g.tmpl, err = template.New("routesImpl").Funcs(routesDefaultFuncMap).Parse(string(data))
	if err != nil {
		return nil, fmt.Errorf("NewRoutesGenerator: %+v", err)
	}

	// services contains the services by its version.
	services := make(map[string]Service)

	for pkg, files := range parsePackages(basePath) {
		service, err := collectService(g.Type, files)
		if err != nil {
			log.Printf("warn: pkg %q is not service: %v", pkg, err)
			continue
		}
		services[service.Version] = service
	}

	// Sort versions
	var keys []string
	for k := range services {
		keys = append(keys, k)
	}
	sort.Sort(sort.Reverse(sort.StringSlice(keys)))
	for _, k := range keys {
		if since != "" && k < since {
			// Skip versions older than since.
			continue
		}
		g.Services = append(g.Services, services[k])
		g.VersionImports = append(g.VersionImports, Import{
			Version: k,
			Name:    services[k].PkgName,
			Path:    "github.com/grid-x/ds-api/" + basePath + "/" + k,
		})
		g.VersionImports = append(g.VersionImports, services[k].Imports...)
	}

	middlewaresMap := make(map[string]bool)
	for _, s := range g.Services {
		for _, g := range s.Groups {
			for _, e := range g.Endpoints {
				for _, m := range e.Middlewares {
					middlewaresMap[strings.ToLower(m)] = true
				}
			}
		}
	}

	for k := range middlewaresMap {
		g.Middlewares = append(g.Middlewares, k)
	}

	return g, nil
}

// collectService search the pkgs for services.
func collectService(apiType string, files []*ast.File) (Service, error) {
	s := Service{}
	var err error
	for _, file := range files {
		s.Version, err = collectVersion(file)
		if err != nil {
			return s, err
		}
		s.PkgName = file.Name.Name + strings.Title(apiType)
		cleanVersion := strings.Replace(s.Version, "-", "", -1)
		s.Name = fmt.Sprintf("service%s", cleanVersion)
		imports, err := collectImports(file)
		if err != nil {
			return s, err
		}
		for key, path := range imports {
			s.Imports = append(s.Imports, Import{
				Version: s.Version,
				Name:    key,
				Path:    path,
			})
		}

		fields, err := collectFields(file)
		if err != nil {
			return s, err
		}
		for key, value := range fields {
			pkgName, _ := splitType(value)
			s.Groups = append(s.Groups, Group{
				Name:    key,
				PkgName: pkgName,
			})
		}
	}

	// Sort groups
	s.Groups, err = parseGroups(s)
	if err != nil {
		return s, err
	}
	sort.Sort(byName(s.Groups))

	// Check if import is necessary.
	importPkgs := []string{}
	for i := range s.Groups {
		for j := range s.Groups[i].Endpoints {
			if s.Groups[i].Endpoints[j].ExpectBody {
				// Add pkg to import
				importPkgs = append(importPkgs, s.Groups[i].PkgName)
				// Rename endpoints pkg name.
				s.Groups[i].Endpoints[j].PkgName =
					fmt.Sprintf("%s%s", s.PkgName, strings.Title(s.Groups[i].PkgName))

			}
		}
	}
	imports := []Import{}
	for i := range s.Imports {
		for _, pkg := range importPkgs {
			if s.Imports[i].Name == pkg {
				s.Imports[i].Name = fmt.Sprintf("%s%s", s.PkgName, strings.Title(s.Imports[i].Name))
				imports = append(imports, s.Imports[i])
			}
		}
	}
	s.Imports = make([]Import, len(imports))
	copy(s.Imports, imports)
	return s, nil
}

type byName []Group

func (s byName) Len() int {
	return len(s)
}
func (s byName) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s byName) Less(i, j int) bool {
	return s[i].Name < s[j].Name
}

//parseGroups parses groups by going through the service's imports.
func parseGroups(s Service) ([]Group, error) {
	for i := range s.Groups {
		var found *Import
		for j := range s.Imports {
			if s.Imports[j].Name == s.Groups[i].PkgName {
				found = &s.Imports[j]
			}
		}
		if found == nil {

			continue
		}
		pkgs := parsePackages(strings.TrimPrefix(found.Path, "github.com/grid-x/ds-api/"))
		for _, pkg := range pkgs {
			if err := parseGroup(pkg, &s.Groups[i]); err != nil {
				return s.Groups, nil
			}
		}
	}
	return s.Groups, nil
}

// splitType splits a field type into pkgName and pkgType.
func splitType(t string) (string, string) {
	p := strings.SplitN(t, ".", 2)
	if len(p) < 2 {
		return "", p[0]
	}
	return p[0], p[1]
}

// parseGroup parses the group based on the ast files.
func parseGroup(files []*ast.File, g *Group) error {
	for _, file := range files {
		endpoints, err := collectEndpoints(file)
		if err != nil {
			return err
		}
		g.Endpoints = append(g.Endpoints, endpoints...)
	}
	return nil
}

// collectVersion retrieves the version of the package name of an ast file.
func collectVersion(f *ast.File) (string, error) {
	name := f.Name.Name
	if !strings.HasPrefix(name, "v") {
		return "", fmt.Errorf("pkg name does not start with v")
	}
	if len(name) != 9 {
		return "", fmt.Errorf("pkg name does not have 9 characters")
	}
	version := fmt.Sprintf("%s-%s-%s", name[1:5], name[5:7], name[7:])
	_, err := time.Parse("2006-01-02", version)
	if err != nil {
		return "", fmt.Errorf("pkg name is not a valid version: %v", err)
	}
	return version, nil
}

func collectEndpoints(f *ast.File) ([]Endpoint, error) {
	var endpoints []Endpoint
	ast.Inspect(f, func(n ast.Node) bool {
		decl, ok := n.(*ast.FuncDecl)
		if !ok {
			return true
		}
		if decl.Recv == nil {
			return true
		}
		expr, ok := decl.Recv.List[0].Type.(*ast.StarExpr)
		if !ok {
			return true
		}
		ident, ok := expr.X.(*ast.Ident)
		if !ok {
			return true
		}
		if ident.Name != "Service" {
			return false
		}
		endpoint, err := collectEndpoint(decl)
		if err != nil {
			return true
		}
		endpoint.PkgName = f.Name.Name
		endpoints = append(endpoints, endpoint)
		return true
	})
	return endpoints, nil
}

// collectImports collects the import statements.
func collectImports(f *ast.File) (map[string]string, error) {
	imports := map[string]string{}
	for _, i := range f.Imports {
		v, err := strconv.Unquote(i.Path.Value)
		if err != nil {
			return nil, err
		}
		_, identifier := path.Split(v)
		if i.Name != nil {
			identifier = i.Name.Name
		}
		imports[identifier] = v
	}
	return imports, nil
}

// Collect endpoint uses the comments of the endpoint such as:
// @name: TODO
// @description: TODO
// @action: system:UploadLogs
// @resource: system:TODO
// @endpoint: POST /system/TODO
// @middlewares: auth,foo,bar
func collectEndpoint(n *ast.FuncDecl) (Endpoint, error) {
	endpoint := Endpoint{
		HandlerName: n.Name.String(),
	}
	scan := bufio.NewScanner(strings.NewReader(n.Doc.Text()))
	for scan.Scan() {
		if !strings.HasPrefix(scan.Text(), "@") {
			continue
		}
		parts := strings.SplitN(scan.Text(), ":", 2)
		key, value := parts[0], parts[1]
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		switch key {
		case "@resource":
			endpoint.Resource = value
		case "@action":
			endpoint.Action = value
		case "@endpoint":
			parts := strings.Split(value, " ")
			endpoint.Path = parts[1]
			endpoint.Methods = []string{parts[0]}
			if parts[0] == "GET" {
				endpoint.Methods = append(endpoint.Methods, "OPTIONS")
			}
		case "@middlewares":
			// we support multiple middlewares separated by ','
			middlewares := strings.Split(value, ",")
			endpoint.Middlewares = append(endpoint.Middlewares, middlewares...)
		case "@name", "@description":
			// will be handled in the future
		default:
			if strings.HasPrefix(key, "@") {
				log.Printf("WARNING: found unknown directive: %s", key)
			}
		}
	}

	for _, arg := range n.Type.Params.List {
		switch n := arg.Type.(type) {
		case *ast.Ident:
			endpoint.ExpectBody = true
			endpoint.RequestStruct = n.Name
		default:
			log.Printf("collectEndpoint: %s: found additional parameter %s; ignored", endpoint.HandlerName, arg.Type)
		}
	}

	return endpoint, nil
}

// collectFields returns a map with fieldName as key and the value is
// the full qualified type.
func collectFields(n ast.Node) (map[string]string, error) {
	fields := map[string]string{}
	ast.Inspect(n, func(n ast.Node) bool {
		t, ok := n.(*ast.TypeSpec)
		if !ok {
			return true
		}

		if !t.Name.IsExported() {
			return false
		}
		if t.Name.Name != "Service" {
			return false
		}

		x, ok := t.Type.(*ast.StructType)
		if !ok {
			return true
		}
		for _, field := range x.Fields.List {
			if len(field.Names) == 0 {
				continue
			}
			fieldType := collectType(field.Type)
			if fieldType != "" {
				if !field.Names[0].IsExported() {
					return false
				}
				fields[field.Names[0].Name] = fieldType
			}

		}
		return true

	})
	return fields, nil
}

// collectType returns full qualified type.
func collectType(n ast.Node) string {
	var fieldType string
	ast.Inspect(n, func(n ast.Node) bool {
		p, ok := n.(*ast.StarExpr)
		if !ok {
			return true
		}
		sel, ok := p.X.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		ident, ok := sel.X.(*ast.Ident)
		if !ok {
			return true
		}
		if !sel.Sel.IsExported() {
			return false
		}
		fieldType = fmt.Sprintf("%s.%s", ident.String(), sel.Sel.Name)
		return true
	})
	return fieldType
}

// GetGenFiles returns a list of files to be generated.
func (g *RoutesGenerator) GetGenFiles() map[string]GenFile {
	files := make(map[string]GenFile)
	filename := g.Output
	if filename == "" {
		filename = fmt.Sprintf("routes_%s_gen.go", g.Type)
	}
	files[filename] = g
	return files
}

// Gen writes the generated files to the writer.
func (g *RoutesGenerator) Gen(w io.Writer) (bool, error) {
	return true, g.tmpl.Execute(w, g)
}

// parsePackages goes through the directory and returns the ast files.
func parsePackages(dir string) map[string][]*ast.File {
	pkgs := make(map[string][]*ast.File)
	found := false
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			return nil
		}
		pkg, err := build.Default.ImportDir(path, 0)
		if _, ok := err.(*build.NoGoError); ok {
			return nil
		} else if err != nil {
			return err
		}
		found = true
		files := parsePackage(path, pkg.GoFiles)
		pkgs[pkg.Dir] = append(pkgs[pkg.Dir], files...)
		return filepath.SkipDir
	})
	if err != nil {
		log.Fatalf("parsePackages: %+v", err)
	}
	if !found {
		log.Fatalf("parsePackages: no go files")
	}
	return pkgs
}

// parsePackage parses files provided.
func parsePackage(dir string, files []string) []*ast.File {
	var astFiles []*ast.File
	fs := token.NewFileSet()
	for _, file := range files {
		if !strings.HasSuffix(file, ".go") {
			continue
		}
		name := filepath.Join(dir, file)
		parsed, err := parser.ParseFile(fs, name, nil, parser.ParseComments)
		if err != nil {
			log.Printf("parsePackage(%s): %+v", name, err)
			continue
		}
		astFiles = append(astFiles, parsed)
	}
	return astFiles
}
