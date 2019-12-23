package api

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
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
				return out, errors.New(s)
			}
			out = i
		}
	}

	if out == "" {
		s := fmt.Sprintf("Found no result for abbreviation %s", prefix)
		return out, errors.New(s)
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
		// do directory stuff
		files, err := ioutil.ReadDir(loc)
		if err != nil {
			return nil, err
		}

		for _, f := range files {
			path := loc + "/" + f.Name()
			b, err := readFile(path)
			if err != nil {
				return nil, err
			}

			ret[path] = b
		}

		return ret, nil
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
