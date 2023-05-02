package api

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"sigs.k8s.io/yaml"

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

func resolveIdentifierFromFile(bytes []byte) (string, error) {
	fullMeta, err := NewFullObjectMeta(bytes)
	if err != nil {
		return "", err
	}

	if fullMeta.Meta.Name != "" {
		return fullMeta.Meta.Name, nil
	}
	if fullMeta.Meta.Id != "" {
		return fullMeta.Meta.Id, nil
	}

	return "", fmt.Errorf("not found")
}

const (
	KNOWN_AFTER_APPLY = "(Known after apply)"
)

func CheckResourceFile(bytes []byte, readOnly bool) (Resource, string, error) {
	resID, err := resolveIdentifierFromFile(bytes)
	if err != nil && readOnly {
		resID = KNOWN_AFTER_APPLY
	}

	application, err := NewApplication(bytes, true)
	if err == nil {
		application.Name = resID
		return &application, resID, nil
	}

	deployment, err := NewDeployment(bytes, true)
	if err == nil {
		deployment.Metadata.ID = resID
		return &deployment, resID, nil
	}

	device, err := NewDevice(bytes, true)
	if err == nil {
		device.Metadata.ID = resID
		return &device, resID, nil
	}

	dcm, err := NewDeviceConfigMap(bytes, true)
	if err == nil {
		dcm.Metadata.ID = resID
		return &dcm, resID, nil
	}

	maintenanceTask, err := NewMaintenanceTask(bytes, true)
	if err == nil {
		maintenanceTask.Metadata.ID = resID
		return &maintenanceTask, resID, nil
	}

	// Nothing found
	return nil, "", errors.E(errors.Invalid, "Unsupported type")
}

type ResAssoc struct {
	ID       string
	Filename string
	Res      Resource
	Order    int
}

var (
	kindOrder = map[string]int{
		"Application":     0,
		"DeviceConfigMap": 1,
		"Deployment":      2,
	}
)

func sortByKind(assocs []ResAssoc) {
	for i := range assocs {
		kind := reflect.TypeOf(assocs[i].Res).Elem().Name()
		var order int
		var exists bool
		if order, exists = kindOrder[kind]; kind == "" || !exists {
			order = -1
		}
		assocs[i].Order = order
	}
	sort.SliceStable(assocs, func(i, j int) bool {
		s := assocs[i].Order
		t := assocs[j].Order

		// resources with no kind go last
		if s > -1 && t == -1 {
			return true
		}
		if t > -1 && s == -1 {
			return false
		}

		// fallback: sort by id
		if s == -1 && t == -1 || s == t {
			return assocs[i].ID < assocs[j].ID
		}

		// sort by kind
		return s < t
	})
}

func hasSupportedExtension(filename string) bool {
	for _, ext := range []string{"yaml", "yml", "json"} {
		if strings.HasSuffix(filename, ext) {
			return true
		}
	}
	return false
}

func GetResources(loc string, readOnly bool, checkForExtensionSupport bool) ([]ResAssoc, error) {
	contents, err := GetFilesContentsToProcess(loc)
	if err != nil {
		return nil, err
	}

	assocs := make([]ResAssoc, 0, len(contents))
	for n, c := range contents {
		var err error
		var res Resource
		var resID string

		if checkForExtensionSupport && !hasSupportedExtension(n) && loc != n {
			continue
		}

		res, resID, err = CheckResourceFile(c, readOnly)
		if err != nil {
			return nil, err
		}
		assoc := ResAssoc{
			ID:       resID,
			Filename: n,
			Res:      res,
			Order:    -1,
		}

		assocs = append(assocs, assoc)
	}

	sortByKind(assocs)
	return assocs, nil
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

			retVal, err := readFile(path)
			if err != nil {
				return err
			}

			for key, val := range retVal {
				b, err := yaml.YAMLToJSON(val)
				if err == nil {
					// TODO does it make sense to silently ignore errors here?
					val = b
				}

				// ignore null values - these are created e.g. if there is a `---` right at the document start
				if !bytes.Equal(b, []byte("null")) {
					ret[key] = val
				}
			}
			return nil
		})

		return ret, err
	case mode.IsRegular():
		retVal, err := readFile(loc)
		if err != nil {
			return nil, err
		}

		for key, val := range retVal {
			b, err := yaml.YAMLToJSON(val)
			if err == nil {
				// TODO does it make sense to silently ignore errors here?
				val = b
			}

			// ignore null values - these are created e.g. if there is a `---` right at the document start
			if !bytes.Equal(b, []byte("null")) {
				ret[key] = val
			}
		}
		return ret, nil
	}

	return nil, fmt.Errorf("Unknown error while processing filename")
}

const yamlSeparator = "\n---"

// splitYAMLDocument is a bufio.SplitFunc for splitting YAML streams into individual documents.
// The following function is taken from 'splitYAMLDocument' function in
// https://github.com/kubernetes/apimachinery/blob/master/pkg/util/yaml/decoder.go
func splitYAMLDocument(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}

	sep := len([]byte(yamlSeparator))

	if i := bytes.Index(data, []byte(yamlSeparator)); i >= 0 {
		// We have a potential document terminator
		i += sep
		after := data[i:]

		if len(after) == 0 {
			// we can't read any more characters
			if atEOF {
				return len(data), data[:len(data)-sep], nil
			}
			return 0, nil, nil
		}

		if i == sep {
			return sep, nil, nil
		}

		if j := bytes.IndexByte(after, '\n'); j >= 0 {
			return i + j + 1, data[0 : i-sep], nil
		}
		return 0, nil, nil
	}
	// If we're at EOF, we have a final, non-terminated line. Return it.
	if atEOF {
		return len(data), data, nil
	}
	// Request more data.
	return 0, nil, nil
}

func readFile(filepath string) (map[string][]byte, error) {
	f, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(f)
	buf := make([]byte, 4*1024)
	scanner.Buffer(buf, 5*1024*1024)
	scanner.Split(splitYAMLDocument)

	fileCounter := 0
	retVal := make(map[string][]byte)

	for scanner.Scan() {
		bytes := scanner.Bytes()
		// Do not add empty documents
		if len(bytes) > 1 {
			fileCounter += 1
			// Trim whitespace in both ends of each yaml docs.
			trimmedString := strings.TrimSpace(scanner.Text())
			retVal[strconv.Itoa(fileCounter)+"/"+filepath] = []byte(trimmedString)
		}
	}

	return retVal, nil
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
