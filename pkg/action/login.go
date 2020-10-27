package action

import (
	"context"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"time"

	"github.com/toqueteos/webbrowser"

	"github.com/grid-x/gxctl/pkg/auth"
	"github.com/grid-x/gxctl/pkg/service"
)

const localAddress = "localhost:4445"

func Login(s *service.Service, openBrowser bool) error {
	l, err := net.Listen("tcp", localAddress)
	if err != nil {
		return err
	}
	l.Close()

	// https://auth0.com/docs/api-auth/tutorials/nonce
	nonce := getRandomString()

	tenant, err := s.Client.GetAuth0TenantFromAuthConfig()
	if err != nil {
		return err
	}
	clientID, err := s.Client.GetAuth0ClientIDFromAuthConfig()
	if err != nil {
		return err
	}

	loginLocation := fmt.Sprintf("https://%s/authorize?nonce=%s&scope=openid%%20email&response_type=id_token&client_id=%s&redirect_uri=http://%s/", tenant, nonce, clientID, localAddress)

	go func() {
		if openBrowser {
			time.Sleep(time.Second * 1)
			webbrowser.Open(loginLocation)
		}
	}()

	fmt.Printf("Setting up home route on %s\n", localAddress)
	fmt.Println("Press ctrl + c on Linux / Windows or cmd + c on OSX to end the process.")
	fmt.Printf("If your browser does not open automatically, navigate to:\n\n\t%s\n\n", loginLocation)

	r := http.NewServeMux()
	server := &http.Server{Addr: localAddress, Handler: r}

	var token string
	r.HandleFunc("/callback", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("error") != "" {
			http.Error(w, "error happened in callback: "+r.URL.Query().Get("error")+" "+r.URL.Query().Get("error_description")+" "+r.URL.Query().Get("error_debug"), http.StatusInternalServerError)
			return
		}
		token = r.URL.Query().Get("id_token")

		// TODO Make a goodlooking exitpage
		fmt.Fprint(w, auth.Finish)

		// Sanity
		time.Sleep(time.Second * 1)

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
		defer cancel()
		server.Shutdown(ctx)

		return
	}))

	r.HandleFunc("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, auth.Redirect)
		return
	}))

	server.ListenAndServe()

	if token != "" {
		if err := s.Client.SetTokenInAuthConfig(token); err != nil {
			return err
		}
		fmt.Println("Authenticated!")
		return nil
	}

	return fmt.Errorf("Unable to authenticate")
}

func getRandomString() string {
	const charset = "0123456789ABCDEFGHIJKLMNOPQRSTUVXYZabcdefghijklmnopqrstuvwxyz-._"

	var seededRand *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))

	b := make([]byte, 32)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}
