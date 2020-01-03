package cmd

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/spf13/cobra"
	"github.com/toqueteos/webbrowser"

	"github.com/grid-x/gxctl/pkg/client"
	gxtemplate "github.com/grid-x/gxctl/pkg/template"
)

type Login struct {
	Command *cobra.Command
}

func NewLogin(parent *cobra.Command, client *client.APIClient) *Login {
	var loginCmd = &cobra.Command{
		Use:                   "login [OPTIONS]",
		DisableFlagsInUseLine: true,
		Short:                 "lelele login token",
		Long:                  `TODO`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// https://auth0.com/docs/api-auth/tutorials/nonce
			nonce := getRandomString()

			tenant, err := client.GetAuth0TenantFromAuthConfig()
			if err != nil {
				return err
			}
			loginLocation := fmt.Sprintf("https://%s.eu.auth0.com/authorize?nonce=%s&scope=openid%%20email&response_type=id_token&client_id=KCKGmjmTAR1R1c2Wa86JDG5M2dOZswLt&redirect_uri=http://localhost:4445/", tenant, nonce)
			webbrowser.Open(loginLocation)

			fmt.Println("Setting up home route on :4445")
			fmt.Println("Press ctrl + c on Linux / Windows or cmd + c on OSX to end the process.")
			fmt.Printf("If your browser does not open automatically, navigate to:\n\n\t%s\n\n", loginLocation)

			r := mux.NewRouter()
			server := &http.Server{Addr: fmt.Sprintf(":%d", 4445), Handler: r}
			var shutdown = func() {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
				defer cancel()
				_ = server.Shutdown(ctx)
			}

			var token string
			r.Path("/callback").Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Query().Get("error") != "" {
					http.Error(w, "error happened in callback: "+r.URL.Query().Get("error")+" "+r.URL.Query().Get("error_description")+" "+r.URL.Query().Get("error_debug"), http.StatusInternalServerError)
					return
				}
				token = r.URL.Query().Get("id_token")

				// TODO Make a goodlooking exitpage and close the tab via JS after a few secs.
				w.Write([]byte("Authenticated! You can now close the Tab."))
				go shutdown()
				return
			}))

			// Serve HTML page to populate id_token from client -> server by redirecting to /callback and attaching the token as query parameter
			r.PathPrefix("/").Handler(http.FileServer(http.Dir("./pkg/auth")))

			server.ListenAndServe()

			if token != "" {
				if err := client.SetTokenInAuthConfig(token); err != nil {
					return err
				}
				fmt.Println("Authenticated!")
				return nil
			}

			return fmt.Errorf("Unable to authenticate")
		},
	}

	loginCmd.SetHelpTemplate(gxtemplate.HelpTemplate())
	loginCmd.SetUsageTemplate(gxtemplate.UsageTemplate())
	parent.AddCommand(loginCmd)

	return &Login{
		Command: loginCmd,
	}
}

func getRandomString() string {
	const charset = "0123456789ABCDEFGHIJKLMNOPQRSTUVXYZabcdefghijklmnopqrstuvwxyz-._"

	var seededRand *rand.Rand = rand.New(
		rand.NewSource(time.Now().UnixNano()))

	b := make([]byte, 32)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}
