package action

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/grid-x/gxctl/pkg/api"
	"github.com/grid-x/gxctl/pkg/service"
)

const (
	GxctlManagedLabelKey = "gxctl.gridx.ai/managed"
)

type record struct {
	local    api.Deployment
	remote   api.Deployment
	fileName string
}

// Prune resources from remote state that don't exist in local state
func Prune(s *service.Service, dryRun bool, fileName string, yes bool, diffCmd string) error {
	var (
		err    error
		local  map[string][]byte
		remote api.Deployments
	)
	if local, err = api.GetFilesContentsToProcess(fileName); err != nil {
		return err
	}

	if remote, err = getDeployments(s.Client); err != nil {
		return err
	}

	state, err := buildPruneState(local, remote)
	if err != nil {
		return err
	}

	for id, rec := range state {
		var (
			ok  bool
			err error
		)
		if ok, err = shouldPrune(diffCmd, dryRun, yes, id, rec); err != nil {
			return err
		}
		if ok {
			if err := apply(id, &rec.local, false, s.Client); err != nil {
				return err
			}
		}
	}

	return nil
}

func buildPruneState(local map[string][]byte, remote api.Deployments) (map[string]*record, error) {
	state := make(map[string]*record)

	for fn, c := range local {
		res, resID, err := checkResourceFile(c, true)
		if err != nil {
			continue
		}
		if v, ok := res.(*api.Deployment); !ok {
			continue
		} else {
			state[resID] = &record{
				local:    *v,
				fileName: fn,
			}
		}
	}

	for _, deploy := range remote.Deployments {
		if isManaged(&deploy) {
			id := deploy.Meta().ID
			if _, ok := state[id]; !ok {
				state[id] = &record{}
			}
			if err := withoutManagedMeta(&deploy); err != nil {
				return nil, err
			}
			state[id].remote = deploy
		}
	}
	return state, nil
}

func shouldPrune(diffCmd string, dryRun bool, yes bool, id string, rec *record) (bool, error) {
	if rec == nil {
		return false, errors.New("nil record")
	}

	err, f1, f2 := writeDiffFiles(rec.local, rec.remote)
	defer os.Remove(f1)
	defer os.Remove(f2)
	if err != nil {
		return false, err
	}

	output, err := exec.Command(diffCmd, f1, f2).Output()
	if err != nil {
		switch err.(type) {
		case *exec.ExitError:
			// this is just an exit code error, no worries
		default: //couldnt run diff
			return false, err
		}
	}

	if len(output) == 0 {
		return false, nil
	}
	fmt.Printf("%s:\n", rec.fileName)
	fmt.Println(string(output))

	if !dryRun && (yes || confirmCli("Apply?")) {
		return true, nil
	}

	return false, nil
}

func confirmCli(msg string) bool {
	var s string

	fmt.Printf(msg + " (y/N): ")
	_, err := fmt.Scan(&s)
	if err != nil {
		return false
	}
	s = strings.TrimSpace(strings.ToLower(s))
	if s == "y" || s == "yes" {
		return true
	}
	return false
}

func isManaged(res api.Resource) bool {
	if res == nil {
		return false
	}

	meta := res.Meta()
	if meta == nil {
		return false
	}

	if meta.Labels == nil {
		return false
	}

	v, ok := meta.Labels[GxctlManagedLabelKey]
	return ok && v == "true"
}

func withManagedMeta(res api.Resource) error {
	if res == nil {
		return fmt.Errorf("res nil")
	}
	meta := res.Meta()
	if meta == nil {
		return fmt.Errorf("meta nil")
	}
	if meta.Labels == nil {
		meta.Labels = make(map[string]string)
	}
	meta.Labels[GxctlManagedLabelKey] = "true"

	return nil
}

func withoutManagedMeta(res api.Resource) error {
	if res == nil {
		return fmt.Errorf("res nil")
	}

	meta := res.Meta()
	if meta == nil {
		return nil
	}

	if meta.Labels == nil {
		return nil
	}

	delete(meta.Labels, GxctlManagedLabelKey)

	return nil
}

func withoutAnnotations(res api.Resource) {
	if res != nil {
		meta := res.Meta()
		if meta != nil && meta.Annotations != nil {
			meta.Annotations = map[string]string{}
		}
	}
}
