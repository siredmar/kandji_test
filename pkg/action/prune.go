package action

import (
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

// Prune resources from remote state that don't exist in local state
func Prune(s *service.Service, dryRun bool, fileName string, yes bool, diffCmd string) error {

	contents, err := api.GetFilesContentsToProcess(fileName)
	if err != nil {
		return err
	}

	type record struct {
		local    api.Deployment
		remote   api.Deployment
		fileName string
	}
	state := make(map[string]*record)

	for fn, c := range contents {
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

	deploys, err := getDeployments(s.Client)
	if err != nil {
		return err
	}

	for _, deploy := range deploys.Deployments {
		if isManaged(&deploy) {
			id := deploy.Meta().ID
			if _, ok := state[id]; !ok {
				state[id] = &record{}
			}
			err = withoutManagedMeta(&deploy)
			if err != nil {
				return err
			}
			state[id].remote = deploy
		}
	}

	for id, rec := range state {
		err, f1, f2 := writeDiffFiles(rec.local, rec.remote)
		defer os.Remove(f1)
		defer os.Remove(f2)
		if err != nil {
			return err
		}

		output, err := exec.Command(diffCmd, f1, f2).Output()
		if err != nil {
			switch err.(type) {
			case *exec.ExitError:
				// this is just an exit code error, no worries
			default: //couldnt run diff
				return err
			}
		}

		if len(output) == 0 {
			continue
		}
		fmt.Printf("%s:\n", rec.fileName)
		fmt.Println(string(output))

		if !dryRun && (yes || confirmCli("Apply?")) {
			err := apply(id, &rec.local, false, s.Client)
			if err != nil {
				return err
			}
		}
	}

	return nil
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
