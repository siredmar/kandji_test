package cmd

import (
	"io"
	"runtime"
	"text/tabwriter"
	"text/template"

	"github.com/spf13/cobra"

	api "github.com/grid-x/gxctl/pkg/api"
	gxtemplate "github.com/grid-x/gxctl/pkg/template"
)

type Version struct {
	Command *cobra.Command
}

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

func NewVersion(parent *cobra.Command) *Version {
	var versionCmd = &cobra.Command{
		Use:                   "version [OPTIONS]",
		DisableFlagsInUseLine: true,
		Short:                 "Version infos",
		Long:                  `TODO`,
		RunE: func(cmd *cobra.Command, args []string) error {
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

			return prettyPrintVersion(cmd.OutOrStdout(), v, versionTemplate)
		},
	}

	versionCmd.SetHelpTemplate(gxtemplate.HelpTemplate())
	versionCmd.SetUsageTemplate(gxtemplate.UsageTemplate())
	parent.AddCommand(versionCmd)

	return &Version{
		Command: versionCmd,
	}
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
