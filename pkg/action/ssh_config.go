package action

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/grid-x/gxctl/pkg/errors"
	"github.com/grid-x/gxctl/pkg/service"
)

var configWant = map[string]map[string]string{
	"*.gridbox-tunnel": {
		"ProxyCommand":          "gxctl ssh tunnel --profile='*' $(echo %h | cut -d'.' -f1)",
		"ServerAliveInterval":   "30",
		"StrictHostKeyChecking": "no",
		"HashKnownHosts":        "no",
		"User":                  "root",
		"ControlMaster":         "auto",
		"ControlPath":           "~/.ssh/master-%r@%h:%p",
		"ControlPersist":        "no",
		"LogLevel":              "ERROR",
	},
	"*.gridbox": {
		"User":                   "root",
		"HostName":               "127.0.0.1",
		"Port":                   "22222",
		"UserKnownHostsFile":     "/dev/null",
		"StrictHostKeyChecking":  "no",
		"LogLevel":               "ERROR",
		"PubkeyAcceptedKeyTypes": "+ssh-rsa",
		"ProxyCommand":           "ssh -o 'ForwardAgent yes' $(echo %n | cut -d'.' -f1).gridbox-tunnel 'ssh-add -q /keys/id_wssh_rsa && nc %h %p'",
	},
}

func SSHConfigPrint(s *service.Service) error {
	var b strings.Builder

	for host, entry := range configWant {
		fmt.Fprintf(&b, "Host %v\n", host)
		for k, v := range entry {
			fmt.Fprintf(&b, "  %v %v\n", k, v)
		}
		fmt.Fprintf(&b, "\n")
	}
	fmt.Print(b.String())
	return nil
}

func SSHConfigCheck(s *service.Service) error {
	var diffs []sshConfigDiff
	for host := range configWant {
		configHave, err := sshGetConfig(host)
		if err != nil {
			return err
		}
		for k, v := range configWant[host] {
			diff, err := sshDiff(host, k, configHave, v)
			if err != nil {
				return err
			}
			if diff != nil {
				diffs = append(diffs, *diff)
			}
		}
	}

	var res []string
	for _, d := range diffs {
		s := fmt.Sprintf("[%v] %v\n  have: %v\n  want: %v", d.host, d.key, d.have, d.want)
		res = append(res, s)
	}

	if len(diffs) > 0 {
		return errors.E(
			errors.Validation,
			res,
		)
	}

	return nil
}

func sshGetConfig(host string) (map[string]string, error) {
	res := make(map[string]string)

	cmd := exec.Command("ssh", "-G", host)
	b, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	lines := strings.Split(string(b), "\n")
	for _, l := range lines {
		f := strings.SplitN(l, " ", 2)
		if len(f) == 1 {
			continue
		}
		res[strings.TrimSpace(f[0])] = f[1]
	}

	return res, nil
}

type sshConfigDiff struct {
	host string
	key  string
	have string
	want string
}

func sshDiff(host, key string, configHave map[string]string, want string) (*sshConfigDiff, error) {
	have, haveKey := configHave[strings.ToLower(key)]
	d := &sshConfigDiff{
		host: host,
		key:  key,
		have: have,
		want: want,
	}

	if d.have == "false" && (d.want == "no" || d.want == "off") {
		return nil, nil
	}

	if strings.HasPrefix(d.want, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}
		d.have = strings.Replace(d.have, home, "~", 1)
	}

	switch strings.ToLower(d.key) {
	case "controlpath":
		d.have = strings.Replace(d.have, "root", "%r", 1)
		d.have = strings.Replace(d.have, "*.gridbox-tunnel", "%h", 1)
		d.have = strings.Replace(d.have, "22", "%p", 1)
	case "pubkeyacceptedkeytypes":
		// The keyword was renamed in OpenSSH 8.5
		// https://www.openssh.com/txt/release-8.5
		if acceptedAlgorithms, haveRenamedKey := configHave["pubkeyacceptedalgorithms"]; !haveKey && haveRenamedKey {
			d.have = acceptedAlgorithms
		}

		if strings.Contains(d.have, strings.TrimLeft(d.want, "+")) {
			return nil, nil
		}
	default:
	}

	if d.have == d.want {
		return nil, nil
	}

	return d, nil
}
