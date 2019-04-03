package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/google/go-cmp/cmp"
	apiClient "github.com/grid-x/ds-api/test/e2e/client"
	log "github.com/sirupsen/logrus"
)

const (
	DeviceAuthEndpoint             = "api/device/auth"
	DeviceDevicesEndpoint          = "api/device"
	DevicePodsEndpoint             = "api/device/pods"
	ManagementDevicesEndpoint      = "api/management/devices"
	ManagementPodsEndpoint         = "api/management/pods"
	ManagementDeploymentsEndpoint  = "api/management/deployments"
	ManagementApplicationsEndpoint = "api/management/applications"
	ManagementMaintenanceEndpoint  = "api/management/maintenance"
)

var (
	APIClient                     *apiClient.APIClient
	APIVersion                    = "application/vnd.gridx.ai.2018-12-04"
	logger                        = log.New()
	ExtractedDeviceUUIDs          []string
	ExtractedDeploymentUUIDs      []string
	ExtractedMaintenanceTaskUUIDs []string
	ExtractedPodUUIDs             []string
)

type Testcase struct {
	name             string
	endpoint         string
	version          string
	method           string
	body             []byte
	token            string
	expectedCode     int
	wanntErr         bool
	compareResponse  bool
	expectedResponse []byte
}

func TestMain(m *testing.M) {
	var (
		url = flag.String("url", "http://127.0.0.1:8080", "target url of DS-API")
	)
	flag.Parse()

	APIClient = apiClient.NewAPIClient(*url)
	os.Exit(m.Run())
}

func areEqualJSON(s1, s2 []byte) (bool, string, error) {
	var o1 interface{}
	var o2 interface{}

	var err error
	err = json.Unmarshal(s1, &o1)
	if err != nil {
		return false, "", fmt.Errorf("Error mashalling string 1 :: %s", string(s1))
	}
	err = json.Unmarshal(s2, &o2)
	if err != nil {
		return false, "", fmt.Errorf("Error mashalling string 2 :: %s", string(s2))
	}

	result := cmp.Diff(o1, o2)

	if result == "" {
		return true, "", nil
	}
	var acceptUUID = regexp.MustCompile(`(?m:^root(\[\".*\"\]\[.*\])?\[\"(metadata|spec)\"\]\[\"(id|accountID)\"\]:(\r\n|\r|\n)\t\-: "(.*)"(\r\n|\r|\n)\t\+: "@UUID"(\r\n|\r|\n)$)`)
	var uuid string
	uuidmatch := acceptUUID.FindStringSubmatch(result)

	if len(uuidmatch) == 8 && uuidmatch[5] != "" {
		uuid = uuidmatch[5]
	}
	res := acceptUUID.ReplaceAllString(result, "")

	if res == "" {
		return true, uuid, nil
	}
	return false, uuid, nil

}

func runTestcase(tc Testcase, t *testing.T) string {
	status, resp, err := APIClient.Request(tc.endpoint, tc.version, tc.method, tc.body, map[string]string{}, tc.token)
	if err != nil && !tc.wanntErr {
		t.Fatalf("unexpected error %v", err)
	}
	if err == nil && tc.wanntErr {
		t.Fatalf("wanted error but passed")
	}

	if status != tc.expectedCode {
		t.Fatalf("unexpected status code. wanted: %d got: %d with response %s", tc.expectedCode, status, string(resp))
	}

	if tc.expectedCode >= 200 && tc.expectedCode <= 204 && tc.compareResponse {
		eq, uuid, err := areEqualJSON(resp, tc.expectedResponse)
		if err != nil {
			t.Fatalf("unexpected error %v", err)
		}

		if !eq {
			t.Fatalf("unexpected response. wanted: %s got: %s", string(tc.expectedResponse), string(resp))
		}

		return uuid
	}
	return ""
}
