package api

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/grid-x/gxctl/pkg/errors"
)

func LookupID(prefix string, ids []string) (string, error) {
	if len(prefix) == 36 {
		//Not an prefix at all but the full ID
		return prefix, nil
	}

	var out string

	for _, i := range ids {
		if strings.HasPrefix(i, prefix) {
			if out != "" {
				s := fmt.Sprintf("Found more then one result for abbreviation %s", prefix)
				return out, errors.E(errors.Invalid, s)
			}
			out = i
		}
	}

	if out == "" {
		s := fmt.Sprintf("Found no result for abbreviation %s", prefix)
		return out, errors.E(errors.NotExists, s)
	}

	return out, nil
}

func GetFilesContentsToProcess(loc string) (map[string][]byte, error) {
	info, err := os.Stat(loc)
	if err != nil {
		return nil, err
	}

	ret := make(map[string][]byte)
	switch mode := info.Mode(); {
	case mode.IsDir():
		err := filepath.Walk(loc, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}

			b, err := readFile(path)
			if err != nil {
				return err
			}
			ret[path] = b
			return nil
		})

		return ret, err
	case mode.IsRegular():
		b, err := readFile(loc)
		if err != nil {
			return nil, err
		}

		ret[loc] = b
		return ret, nil
	}

	return nil, fmt.Errorf("Unknown error while processing filename")
}

func readFile(filepath string) ([]byte, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	return bytes, nil
}

func ComputeMetadataMap(current map[string]string, update map[string]string) map[string]string {
	result := make(map[string]string)

	// Prepare to remove entries, which are found in the current state but not in the update from the state.
	for k := range current {
		if _, ok := update[k]; !ok {
			result[k+"-"] = ""
		}
	}

	for k, v := range update {
		result[k] = v
	}

	return result
}
