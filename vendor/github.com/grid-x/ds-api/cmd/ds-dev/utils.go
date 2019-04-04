package main

import (
	"bytes"
	"fmt"
	"go/format"
	"io"
	"io/ioutil"
	"regexp"
	"strings"
)

type GenFile interface {
	// should return true and nil if it is a go file
	Gen(io.Writer) (bool, error)
}

func writeToGoFile(file string, content []byte) error {
	// format go code
	formatted, err := format.Source(content)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(file, formatted, 0644)
}

func writeGenFiles(fs map[string]GenFile) error {
	for filename, gofile := range fs {
		var buf bytes.Buffer
		isGenFile, err := gofile.Gen(&buf)
		if err != nil {
			return err
		}

		if isGenFile {
			if err := writeToGoFile(filename, buf.Bytes()); err != nil {
				return fmt.Errorf("error writing file %s: %+v", filename, err)
			}
		} else {
			return ioutil.WriteFile(filename, buf.Bytes(), 0644)
		}
	}
	return nil
}

var camel = regexp.MustCompile("(^[^A-Z0-9]*|[A-Z0-9]*)([A-Z0-9][^A-Z]+|$)")

func underscore(s string) string {
	s = strings.Title(s)
	var a []string
	for _, sub := range camel.FindAllStringSubmatch(s, -1) {
		if sub[1] != "" {
			a = append(a, sub[1])
		}
		if sub[2] != "" {
			a = append(a, sub[2])
		}
	}
	return strings.ToLower(strings.Join(a, "_"))
}

func counter() func() int {
	i := -1
	return func() int {
		i++
		return i
	}
}

func contains(s1 string, s2 string) bool {
	return strings.Contains(s1, s2)
}
