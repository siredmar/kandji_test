package client

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	auth "github.com/grid-x/gxctl/pkg/auth"
)

const (
	baseURL = "https://api.ds.gridx.ai/"
)

var client = &http.Client{Timeout: 10 * time.Second}

func Request(endpoint string) []byte {
	token, err := auth.GenerateJWTToken()

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	req, err := http.NewRequest(http.MethodGet, baseURL+endpoint, nil)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	req.Header = map[string][]string{
		"Authorization": {fmt.Sprintf("Bearer %s", token)},
	}
	r, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer r.Body.Close()

	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	return bodyBytes
}
