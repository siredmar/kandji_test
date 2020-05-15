package action

import (
	"io"
	"os"
	"runtime"
	"text/tabwriter"
	"text/template"

	api "github.com/grid-x/gxctl/pkg/api"
	"github.com/grid-x/gxctl/pkg/service"
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

var (
	gitCommit string
	buildTime string
	version   string
)

func Version(s *service.Service) error {
	v := clientVersion{
		Version:    version,
		APIVersion: api.APIVersion,
		GoVersion:  runtime.Version(),
		OS:         runtime.GOOS,
		Arch:       runtime.GOARCH,
		GitCommit:  gitCommit,
		BuildTime:  buildTime,
	}

	versionTemplate, err := newVersionTemplate()
	if err != nil {
		return err
	}

	return prettyPrintVersion(os.Stdout, v, versionTemplate)
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
