package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"testing"
	"time"

	apiClient "github.com/grid-x/ds-api/test/e2e/client"
)

func TestFullWorkflow(t *testing.T) {
	now := time.Now().Format(time.RFC3339)
	nowResponse := time.Now().UTC().Format(time.RFC3339)

	// Only the first device will report back some status vie device API
	// Consequently we do not need to do some token handling for the second device

	absPath, err := filepath.Abs("test-certs/device-private.pem")
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}
	privPEM, err := ioutil.ReadFile(absPath)
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}
	privKey, err := apiClient.GetPrivKeyFromPem(privPEM)
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}
	pubKeyPem, err := apiClient.GetPublicKeyPem(privKey)
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}

	/*
		- Device 1 (setup with proper keys to handle auth process)
		- Device 2
		- Application
		- Deployment for label gridx.de/channel=alpha
	*/
	tt := []Testcase{{
		name:     "Create device",
		endpoint: ManagementDevicesEndpoint,
		version:  APIVersion,
		method:   http.MethodPost,
		body: []byte(fmt.Sprintf(
			`{
				"spec":{
					"serialnumber":"abcdefghijklmno-123",
					"macAddress":"00-14-22-01-23-45",
					"publicKey":"%s",
					"maintenanceWindow":"Sun:02:00-Sun:03:00"
				}
			}`, apiClient.JSONEscape(pubKeyPem))),
		expectedCode:    201,
		wanntErr:        false,
		compareResponse: true,
		expectedResponse: []byte(fmt.Sprintf(
			`{
				"metadata":{
					"id":"@UUID"
				},
				"spec":{
					"serialnumber":"abcdefghijklmno-123",
					"macAddress":"00-14-22-01-23-45",
					"publicKey":"%s",
					"maintenanceWindow":"Sun:02:00-Sun:03:00"
				},
				"status":{}
			}`, apiClient.JSONEscape(pubKeyPem))),
	}, {
		name:     "Create device",
		endpoint: ManagementDevicesEndpoint,
		version:  APIVersion,
		method:   http.MethodPost,
		body: []byte(
			`{
				"spec":{
					"serialnumber":"abcdefghijklmno-987",
					"macAddress":"99-88-77-66-55-44",
					"publicKey":"ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC4H3QCpgn/3//9NMPWuOn69wDEgw8PnOqu4ucd7oJid3Yi0a8A63N1U+LvWwPrupKS5GE+D+Q2wDOraOzOHjWgAnZmI7JtqZCZgcDMTzUM6DmvUAwZZ0lCY1FqBHDtdWp/RWEDb6B4ZQIYxEvLqwZRGCCi5mE2aoMargbus6JKQrdtLJbOh4ybhm2F1aqlRmlhWmp4Hl2FvOHytMm5O8GxXMJ8TrjTzDvmMsqCgqAk2nqxe/oD1nTtcgHl/KSAhf/0mtfokMhCMFLLbBL/MtLfhBdXjob+BP6SWJS0E24a43OffnhgpvgY9vFwRdnZStM65khZXfQzzdx62mZdSs0f joel@deep-thought.local",
					"maintenanceWindow":"Sun:02:00-Sun:03:00"
				}
			}`),
		expectedCode:    201,
		wanntErr:        false,
		compareResponse: true,
		expectedResponse: []byte(
			`{
				"metadata":{
					"id":"@UUID"
				},
				"spec":{
					"serialnumber":"abcdefghijklmno-987",
					"macAddress":"99-88-77-66-55-44",
					"publicKey":"ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC4H3QCpgn/3//9NMPWuOn69wDEgw8PnOqu4ucd7oJid3Yi0a8A63N1U+LvWwPrupKS5GE+D+Q2wDOraOzOHjWgAnZmI7JtqZCZgcDMTzUM6DmvUAwZZ0lCY1FqBHDtdWp/RWEDb6B4ZQIYxEvLqwZRGCCi5mE2aoMargbus6JKQrdtLJbOh4ybhm2F1aqlRmlhWmp4Hl2FvOHytMm5O8GxXMJ8TrjTzDvmMsqCgqAk2nqxe/oD1nTtcgHl/KSAhf/0mtfokMhCMFLLbBL/MtLfhBdXjob+BP6SWJS0E24a43OffnhgpvgY9vFwRdnZStM65khZXfQzzdx62mZdSs0f joel@deep-thought.local",
					"maintenanceWindow":"Sun:02:00-Sun:03:00"
				},
				"status":{}
			}`),
	}, {
		name:     "Create application",
		endpoint: ManagementApplicationsEndpoint,
		version:  APIVersion,
		method:   http.MethodPost,
		body: []byte(
			`{
				"name":"testapp"
			}`),
		expectedCode:    201,
		wanntErr:        false,
		compareResponse: true,
		expectedResponse: []byte(
			`{
				"metadata":{
					"id":"testapp"
				},
				"name":"testapp"
			}`),
	}, {
		name:     "Create deployment",
		endpoint: ManagementDeploymentsEndpoint,
		version:  APIVersion,
		method:   http.MethodPost,
		body: []byte(
			`{
				"spec":{
					"app":"testapp",
					"selector":{
						"matchByLabels":{
							"gridx.de/channel":"alpha"
						}
					},
					"template":{
						"spec":{
							"containers":[
									{
										"name":"monitoring-agent",
										"image":"gridx/monitoring:foobar-u28391389"
									}
							]
						}
					}
				}
			}`),
		expectedCode:    201,
		wanntErr:        false,
		compareResponse: true,
		expectedResponse: []byte(
			`{
				"metadata":{
					"id":"@UUID"
				},
				"spec":{
					"app":"testapp",
					"selector":{
						"matchByLabels":{
							"gridx.de/channel":"alpha"
						}
					},
					"template":{
						"spec":{
							"containers":[
									{
										"name":"monitoring-agent",
										"image":"gridx/monitoring:foobar-u28391389"
									}
							]
						}
					}
				},
				"status":{}
			}`),
	}}
	ExtractedDeviceUUIDs = ExtractedDeviceUUIDs[:0]         // Sanity
	ExtractedDeploymentUUIDs = ExtractedDeploymentUUIDs[:0] // Sanity
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			uuid := runTestcase(tc, t)
			if uuid != "" {
				if tc.endpoint == ManagementDevicesEndpoint {
					ExtractedDeviceUUIDs = append(ExtractedDeviceUUIDs, uuid)
				}
				if tc.endpoint == ManagementDeploymentsEndpoint {
					ExtractedDeploymentUUIDs = append(ExtractedDeploymentUUIDs, uuid)
				}
			}
		})
	}

	token, err := APIClient.GetDeviceToken(privKey, string(pubKeyPem), DeviceAuthEndpoint, APIVersion)
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}

	/*
		- Device 1 - Report heartbeat
	*/
	tt = []Testcase{{
		name:     "Report Heartbeat",
		endpoint: fmt.Sprintf("%s/", DeviceDevicesEndpoint),
		version:  APIVersion,
		method:   http.MethodPatch,
		token:    token.String(),
		body: []byte(fmt.Sprintf(
			`{
				"status": {
					"lastHeartbeat":"%s"
				}
			}`, now)),
		expectedCode:    200,
		wanntErr:        false,
		compareResponse: true,
		expectedResponse: []byte(fmt.Sprintf(
			`{
				"metadata": {
					"id":"%s"
				},
				"spec":{
					"accountID":"@UUID",
					"serialnumber":"abcdefghijklmno-123",
					"macAddress":"00-14-22-01-23-45",
					"publicKey":"%s",
					"maintenanceWindow":"Sun:02:00-Sun:03:00"
				},
				"status":{
					"lastHeartbeat":"%s"
				}
			}`, ExtractedDeviceUUIDs[0], apiClient.JSONEscape(pubKeyPem), nowResponse)),
	}, {
		name:            "Get single device",
		endpoint:        fmt.Sprintf("%s/%s", ManagementDevicesEndpoint, ExtractedDeviceUUIDs[0]),
		version:         APIVersion,
		method:          http.MethodGet,
		body:            nil,
		expectedCode:    200,
		wanntErr:        false,
		compareResponse: true,
		expectedResponse: []byte(fmt.Sprintf(
			`{
				"metadata": {
					"id":"%s"
				},
				"spec":{
					"serialnumber":"abcdefghijklmno-123",
					"macAddress":"00-14-22-01-23-45",
					"publicKey":"%s",
					"maintenanceWindow":"Sun:02:00-Sun:03:00"
				},
				"status":{
					"lastHeartbeat":"%s"
				}
			}`, ExtractedDeviceUUIDs[0], apiClient.JSONEscape(pubKeyPem), nowResponse)),
	}}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			runTestcase(tc, t)
		})
	}

	/*
		- Device 1 - Patch label to gridx.de/channel=alpha
		- Device 2 - Patch label to gridx.de/channel=stable
	*/
	tt = []Testcase{{
		name:            "Get all pods",
		endpoint:        ManagementPodsEndpoint,
		version:         APIVersion,
		method:          http.MethodGet,
		body:            nil,
		expectedCode:    200,
		wanntErr:        false,
		compareResponse: true,
		expectedResponse: []byte(
			`{
				"pods":[]
			}`),
	}, {
		name:     "Patch single device",
		endpoint: fmt.Sprintf("%s/%s", ManagementDevicesEndpoint, ExtractedDeviceUUIDs[0]),
		version:  APIVersion,
		method:   http.MethodPatch,
		body: []byte(
			`{
				"metadata": {
					"labels":{"gridx.de/channel": "alpha"}
				}
			}`),
		expectedCode:    200,
		wanntErr:        false,
		compareResponse: true,
		expectedResponse: []byte(fmt.Sprintf(
			`{
				"metadata": {
					"id":"%s",
					"labels": {"gridx.de/channel": "alpha"}
				},
				"spec":{
					"serialnumber":"abcdefghijklmno-123",
					"macAddress":"00-14-22-01-23-45",
					"publicKey":"%s",
					"maintenanceWindow":"Sun:02:00-Sun:03:00"
				},
				"status":{
					"lastHeartbeat":"%s"
				}
			}`, ExtractedDeviceUUIDs[0], apiClient.JSONEscape(pubKeyPem), nowResponse)),
	}, {
		name:     "Patch single device",
		endpoint: fmt.Sprintf("%s/%s", ManagementDevicesEndpoint, ExtractedDeviceUUIDs[1]),
		version:  APIVersion,
		method:   http.MethodPatch,
		body: []byte(
			`{
				"metadata": {
					"labels":{"gridx.de/channel": "stable"}
				}
			}`),
		expectedCode:    200,
		wanntErr:        false,
		compareResponse: true,
		expectedResponse: []byte(fmt.Sprintf(
			`{
				"metadata": {
					"id":"%s",
					"labels": {"gridx.de/channel": "stable"}
				},
				"spec":{
					"serialnumber":"abcdefghijklmno-987",
					"macAddress":"99-88-77-66-55-44",
					"publicKey":"ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC4H3QCpgn/3//9NMPWuOn69wDEgw8PnOqu4ucd7oJid3Yi0a8A63N1U+LvWwPrupKS5GE+D+Q2wDOraOzOHjWgAnZmI7JtqZCZgcDMTzUM6DmvUAwZZ0lCY1FqBHDtdWp/RWEDb6B4ZQIYxEvLqwZRGCCi5mE2aoMargbus6JKQrdtLJbOh4ybhm2F1aqlRmlhWmp4Hl2FvOHytMm5O8GxXMJ8TrjTzDvmMsqCgqAk2nqxe/oD1nTtcgHl/KSAhf/0mtfokMhCMFLLbBL/MtLfhBdXjob+BP6SWJS0E24a43OffnhgpvgY9vFwRdnZStM65khZXfQzzdx62mZdSs0f joel@deep-thought.local",
					"maintenanceWindow":"Sun:02:00-Sun:03:00"
				},
				"status":{}
			}`, ExtractedDeviceUUIDs[1])),
	}}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			runTestcase(tc, t)
		})
	}

	// Wait for pods to be created
	time.Sleep(time.Second * 5)

	// Checking that there are now pods available for the first device matching alpha group
	tt = []Testcase{{
		name:            "Get all pods",
		endpoint:        ManagementPodsEndpoint,
		version:         APIVersion,
		method:          http.MethodGet,
		body:            nil,
		expectedCode:    200,
		wanntErr:        false,
		compareResponse: true,
		expectedResponse: []byte(fmt.Sprintf(
			`{
				"pods":[
					 {
							"metadata":{
								 "id":"@UUID",
								 "annotations":{
										"gridx.ai/app":"testapp"
								 }
							},
							"spec":{
								 "deviceID":"%s",
								 "config":{
										"containers":[
											 {
													"name":"monitoring-agent",
													"image":"gridx/monitoring:foobar-u28391389"
											 }
										]
								 }
							},
							"status":{
							}
					 }
				]
		 }`, ExtractedDeviceUUIDs[0])),
	}}
	ExtractedPodUUIDs = ExtractedPodUUIDs[:0] // Sanity
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			uuid := runTestcase(tc, t)
			if uuid != "" {
				ExtractedPodUUIDs = append(ExtractedPodUUIDs, uuid)
			}
		})
	}

	/*
		- Device 1 - Report pod start time
	*/
	tt = []Testcase{{
		name:     "Report Starttime",
		endpoint: fmt.Sprintf("%s/%s", DevicePodsEndpoint, ExtractedPodUUIDs[0]),
		version:  APIVersion,
		method:   http.MethodPatch,
		token:    token.String(),
		body: []byte(fmt.Sprintf(
			`{
				"status": {
					"startTime":"%s"
				}
			}`, now)),
		expectedCode:    200,
		wanntErr:        false,
		compareResponse: true,
		expectedResponse: []byte(fmt.Sprintf(
			`{
				"metadata":{
					"id":"%s",
					"annotations":{
							"gridx.ai/app":"testapp"
					}
				},
				"spec":{
					"deviceID":"%s",
					"config":{
							"containers":[
								{
										"name":"monitoring-agent",
										"image":"gridx/monitoring:foobar-u28391389"
								}
							]
					}
				},
				"status":{
					"startTime":"%s"
				}
			}`, ExtractedPodUUIDs[0], ExtractedDeviceUUIDs[0], nowResponse)),
	}, {
		name:            "Get single pod",
		endpoint:        fmt.Sprintf("%s/%s", ManagementPodsEndpoint, ExtractedPodUUIDs[0]),
		version:         APIVersion,
		method:          http.MethodGet,
		body:            nil,
		expectedCode:    200,
		wanntErr:        false,
		compareResponse: true,
		expectedResponse: []byte(fmt.Sprintf(
			`{
				"metadata":{
						"id":"%s",
						"annotations":{
							"gridx.ai/app":"testapp"
						}
				},
				"spec":{
						"deviceID":"%s",
						"config":{
							"containers":[
									{
										"name":"monitoring-agent",
										"image":"gridx/monitoring:foobar-u28391389"
									}
							]
						}
				},
				"status":{
					"startTime":"%s"
				}
			}`, ExtractedPodUUIDs[0], ExtractedDeviceUUIDs[0], nowResponse)),
	}}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			runTestcase(tc, t)
		})
	}

	// Cleanup
	tt = []Testcase{{
		name:             "Delete deployment",
		endpoint:         fmt.Sprintf("%s/%s", ManagementDeploymentsEndpoint, ExtractedDeploymentUUIDs[0]),
		version:          APIVersion,
		method:           http.MethodDelete,
		body:             nil,
		expectedCode:     200,
		wanntErr:         false,
		compareResponse:  true,
		expectedResponse: []byte(`{}`),
	}, {
		name:             "Delete device",
		endpoint:         fmt.Sprintf("%s/%s", ManagementDevicesEndpoint, ExtractedDeviceUUIDs[0]),
		version:          APIVersion,
		method:           http.MethodDelete,
		body:             nil,
		expectedCode:     200,
		wanntErr:         false,
		compareResponse:  true,
		expectedResponse: []byte(`{}`),
	}, {
		name:             "Delete device",
		endpoint:         fmt.Sprintf("%s/%s", ManagementDevicesEndpoint, ExtractedDeviceUUIDs[1]),
		version:          APIVersion,
		method:           http.MethodDelete,
		body:             nil,
		expectedCode:     200,
		wanntErr:         false,
		compareResponse:  true,
		expectedResponse: []byte(`{}`),
	}, {
		name:             "Delete application",
		endpoint:         fmt.Sprintf("%s/%s", ManagementApplicationsEndpoint, "testapp"),
		version:          APIVersion,
		method:           http.MethodDelete,
		body:             nil,
		expectedCode:     200,
		wanntErr:         false,
		compareResponse:  true,
		expectedResponse: []byte(`{}`),
	}}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			runTestcase(tc, t)
		})
	}
}
