package main

import (
	"fmt"
	"io"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
)

var (
	handlerDefaultFuncMap = template.FuncMap{
		"toFuncName": func(h string) string {
			// TODO: check for invalid chars
			reg := regexp.MustCompile("^[a-zA-Z]+")
			p := reg.ReplaceAllString(h, "")
			return strings.Title(p)
		},
		"untitle": func(s string) string {
			if len(s) < 1 {
				return s
			}
			return strings.ToLower(s[0:1]) + s[1:len(s)]
		},
		"title": func(s string) string {
			return strings.Title(s)
		},
		"takesInput": func(method string) bool {
			return method == "POST" || method == "PATCH"
		},
	}

	serviceDefaultFuncMap = template.FuncMap{
		"toPackageName": func(version string) string {
			reg := regexp.MustCompile("[^0-9]+")
			p := reg.ReplaceAllString(version, "")
			return "v" + p
		},
		"title": func(s string) string {
			return strings.Title(s)
		},
	}
)

type Handler struct {
	Group  string
	Name   string
	Method string
}

type handlerImpl struct {
	APIGroup   string
	APIVersion string
	Handlers   []Handler
}

func (h handlerImpl) Gen(w io.Writer) (bool, error) {
	data, err := Asset("templates/handler.go.tmpl")
	if err != nil {
		return true, err
	}
	// Create a template, add the function map, and parse the text.
	tmpl, err := template.New("handlerImpl").Funcs(handlerDefaultFuncMap).Parse(string(data))
	if err != nil {
		return true, err
	}

	// Run the template to verify the output.
	return true, tmpl.Execute(w, h)
}

type handlerTestImpl struct {
	APIGroup string
	Handlers []Handler
}

func (h handlerTestImpl) Gen(w io.Writer) (bool, error) {
	data, err := Asset("templates/handler_test.go.tmpl")
	if err != nil {
		return true, err
	}
	// Create a template, add the function map, and parse the text.
	tmpl, err := template.New("handlerTestImpl").Funcs(handlerDefaultFuncMap).Parse(string(data))
	if err != nil {
		return true, err
	}

	// Run the template to verify the output.
	return true, tmpl.Execute(w, h)
}

func handlerFromString(handlers string) ([]Handler, error) {
	result := []Handler{}
	groupAndHandlers := strings.Split(handlers, ",")
	for _, item := range groupAndHandlers {
		ls := strings.Split(item, ":")
		if len(ls) < 2 || len(ls) > 3 {
			return nil, fmt.Errorf("invalid parameter: %s", handlers)
		}
		group := strings.ToLower(ls[0])
		method := "GET"
		if len(ls) == 3 {
			method = strings.ToUpper(ls[2])
			switch method {
			case "GET", "POST", "PATCH", "DELETE":
			default:
				return nil, fmt.Errorf("method %s not supported", method)
			}
		}

		if strings.ToUpper(ls[1]) == "CRUD" {
			hs := []Handler{
				{
					Group:  group,
					Name:   "Create",
					Method: "POST",
				},
				{
					Group:  group,
					Name:   "List",
					Method: "GET",
				},
				{
					Group:  group,
					Name:   "Get",
					Method: "GET",
				},
				{
					Group:  group,
					Name:   "Update",
					Method: "PATCH",
				},
				{
					Group:  group,
					Name:   "Delete",
					Method: "DELETE",
				},
			}
			result = append(result, hs...)

		} else {
			result = append(result, Handler{
				Group:  ls[0],
				Name:   ls[1],
				Method: method,
			})
		}

	}

	return result, nil
}

type serviceImpl struct {
	Type       string
	APIVersion string
	APIGroups  []string
}

func (s serviceImpl) Gen(w io.Writer) (bool, error) {
	data, err := Asset("templates/service.go.tmpl")
	if err != nil {
		return true, err
	}

	// Create a template, add the function map, and parse the text.
	tmpl, err := template.New("serviceImpl").Funcs(serviceDefaultFuncMap).Parse(string(data))
	if err != nil {
		return true, err
	}

	// Run the template to verify the output.
	return true, tmpl.Execute(w, s)

}

type APIGenerator struct {
	basePath        string
	handlerFileName string
	serviceFileName string

	version  string
	handlers []Handler
}

func NewAPIGenerator(basePath, handlerFileName, serviceFileName, version, handlersStr string) (*APIGenerator, error) {
	handlers, err := handlerFromString(handlersStr)
	if err != nil {
		return nil, err
	}
	return &APIGenerator{
		basePath:        basePath,
		handlerFileName: handlerFileName,
		serviceFileName: serviceFileName,
		version:         version,
		handlers:        handlers,
	}, nil
}

func (a *APIGenerator) GetGenFiles() ([]string, map[string]GenFile) {
	handlersByGroup := make(map[string][]Handler)
	for _, h := range a.handlers {
		fmt.Printf("found handler: %+v \n", h)
		group := h.Group
		handlersByGroup[group] = append(handlersByGroup[group], h)
	}

	var hs []handlerImpl
	var hts []handlerTestImpl
	for group, handlers := range handlersByGroup {
		hs = append(hs, handlerImpl{
			APIGroup:   group,
			APIVersion: a.version,
			Handlers:   handlers,
		})
		hts = append(hts, handlerTestImpl{
			APIGroup: group,
			Handlers: handlers,
		})
	}

	dirs := []string{}
	gofiles := make(map[string]GenFile)
	si := serviceImpl{
		Type:       strings.Trim(a.basePath, "api/"),
		APIVersion: a.version,
		APIGroups:  []string{},
	}
	for _, h := range hs {
		dir := filepath.Join(a.basePath, a.version, h.APIGroup)
		dirs = append(dirs, dir)
		gofiles[filepath.Join(dir, a.handlerFileName)] = h
		si.APIGroups = append(si.APIGroups, h.APIGroup)
	}
	for _, h := range hts {
		dir := filepath.Join(a.basePath, a.version, h.APIGroup)
		gofiles[filepath.Join(dir, a.handlerFileName[0:len(a.handlerFileName)-3]+"_test.go")] = h
	}
	gofiles[filepath.Join(a.basePath, a.version, a.serviceFileName)] = si

	return dirs, gofiles
}
