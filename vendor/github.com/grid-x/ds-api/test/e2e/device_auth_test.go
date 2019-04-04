package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"testing"

	apiClient "github.com/grid-x/ds-api/test/e2e/client"
)

func TestAuthEndpoint(t *testing.T) {
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

	//First create a device
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
		expectedResponse: []byte(
			`{
				"metadata":{
					"id":"@UUID"
				},
				"spec":{
					"serialnumber":"abcdefghijklmno-123",
					"macAddress":"00-14-22-01-23-45",
					"publicKey":"-----BEGIN PUBLIC KEY-----\nMIIBojANBgkqhkiG9w0BAQEFAAOCAY8AMIIBigKCAYEAzAP3nC7TFtrWplw2pgP+\np5CJu1Qyw8Tm8/4ktSenaugQYWsTNIOq9fbS133GNIELHiXp0ZwJa216cpebMQ3I\nq+pTtfUA3T1TD9x8NdlWELlYotR/b96ImEvMp2Bvb39yZZ5ZjcBSKWq4y1SH8aF+\nDpb6hiORT24bHXjxh0LQaL7t72RYT/InvIxw2cd7MyzXA49V3SmvNIr3bd7OfZw5\nFly/EkZecAhfN1CshRq/pRux8CasBtb4SpuWj0/wKzBvqEKA1jo1Tu+72+AoxdC8\nITypxUDNyIzAsjrZjSDXWDeU9AnEGWah9jiF4Rw1Z0wtniZh94fkJRxutqTvV7D7\necK6AAxA2ZuJm+fRZITho9gBVs7c3MBbn7DrfYl1+kS8sn/8mxcdD1BwB5/K/zLv\nhmtyFWc/JBlbWoOh+4BpWvvsCCqe+01CL+sxICaFFs71j9JX2FHU9rQa4kkN71WN\ngSPzDDU16xx3azJmgt5ocBO+V41dKTynRbBnGHEvths1AgMBAAE=\n-----END PUBLIC KEY-----\n",
					"maintenanceWindow":"Sun:02:00-Sun:03:00"
				},
				"status":{}
			}`),
	}}
	ExtractedDeviceUUIDs = ExtractedDeviceUUIDs[:0] // Sanity
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			uuid := runTestcase(tc, t)
			if uuid != "" {
				ExtractedDeviceUUIDs = append(ExtractedDeviceUUIDs, uuid)
			}
		})
	}

	_, err = APIClient.GetDeviceToken(privKey, string(pubKeyPem), DeviceAuthEndpoint, APIVersion)
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}

	// Cleanup
	tt = []Testcase{{
		name:             "Delete device",
		endpoint:         fmt.Sprintf("%s/%s", ManagementDevicesEndpoint, ExtractedDeviceUUIDs[0]),
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
