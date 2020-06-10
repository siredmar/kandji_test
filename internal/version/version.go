package version

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"text/tabwriter"
	"text/template"

	api "github.com/grid-x/gxctl/pkg/api"
)

const Unknown = "unknown"

var (
	GitCommit = Unknown // git commit SHA
	Version   = Unknown // gxctl Version
	BuildTime = Unknown // build datetime
)

type clientVersion struct {
	Version    string
	APIVersion string
	GoVersion  string
	GitCommit  string
	BuildTime  string
	OS         string
	Arch       string
}

// VersionDetails prints detailed version info
func VersionDetails() error {
	v := clientVersion{
		Version:    Version,
		APIVersion: api.APIVersion,
		GoVersion:  runtime.Version(),
		OS:         runtime.GOOS,
		Arch:       runtime.GOARCH,
		GitCommit:  GitCommit,
		BuildTime:  BuildTime,
	}

	versionTemplate, err := newVersionTemplate()
	if err != nil {
		return err
	}

	return prettyPrintVersion(os.Stdout, v, versionTemplate)
}

// UserAgent returns version info in HTTP User-Agent header format
func UserAgent() string {
	return fmt.Sprintf("gxctl/%s (%.7s; %s/%s; %s)", Version, GitCommit, runtime.GOOS, runtime.GOARCH, runtime.Version())
}

func newVersionTemplate() (*template.Template, error) {
	versionTemplate := `gridX:
	Version:	{{.Version}}
	APIVersion:	{{.APIVersion}}
	Built:	{{.BuildTime}}
	Git commit:	{{.GitCommit}}
Environment:
	OS/Arch:	{{.OS}}/{{.Arch}}
	Go version:	{{.GoVersion}}
`
	return template.New("version").Parse(versionTemplate)
}

func prettyPrintVersion(out io.Writer, info clientVersion, tmpl *template.Template) error {
	t := tabwriter.NewWriter(out, 2, 5, 3, ' ', 0)
	if err := tmpl.Execute(t, info); err != nil {
		return err
	}
	return t.Flush()
}
