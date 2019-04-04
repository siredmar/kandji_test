package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"testing"
	"time"

	"github.com/grid-x/ds-api/pkg/ssh"
	apiClient "github.com/grid-x/ds-api/test/e2e/client"
)

func TestSSHEndpoint(t *testing.T) {
	// Only the first device will report back some status vie device API

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
	}}
	ExtractedDeviceUUIDs = ExtractedDeviceUUIDs[:0] // Sanity
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			uuid := runTestcase(tc, t)
			if uuid != "" {
				if tc.endpoint == ManagementDevicesEndpoint {
					ExtractedDeviceUUIDs = append(ExtractedDeviceUUIDs, uuid)
				}
			}
		})
	}

	token, err := APIClient.GetDeviceToken(privKey, string(pubKeyPem), DeviceAuthEndpoint, APIVersion)
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}

	// CLIENT
	managementEndpoint := fmt.Sprintf("%s/%s/ssh", ManagementDevicesEndpoint, ExtractedDeviceUUIDs[0])
	clientConn, err := APIClient.GetWebsocketConnection(managementEndpoint, "", APIVersion)
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}

	// Make sure connection is getting closed
	defer clientConn.Close()

	// Create client Channel
	clientC := make(chan []byte)

	go func() {
		defer close(clientC)
		for {
			_, message, err := clientConn.ReadMessage()
			if err != nil {
				fmt.Println("Connection closed")
			}
			clientC <- message
		}
	}()

	time.Sleep(time.Second * 10)

	// In the meantime the pod should have been come up and the device should connect...

	// DEVICE
	deviceEndpoint := fmt.Sprintf("%s/ssh", DeviceDevicesEndpoint)
	deviceConn, err := APIClient.GetWebsocketConnection(deviceEndpoint, token.String(), APIVersion)

	// Create device Channel
	deviceC := make(chan []byte)

	go func() {
		defer close(deviceC)
		for {
			_, message, err := deviceConn.ReadMessage()
			if err != nil {
				fmt.Println("Connection closed")
			}
			deviceC <- message
		}
	}()

	time.Sleep(time.Second * 3)

	// All set now... Let's check the commuincation

	// First of all, Device should get a CreateProcessMessage after client has registered
	var createProcess ssh.CreateProcessMessage
	err = json.Unmarshal(<-deviceC, &createProcess)
	if err != nil {
		t.Fatalf("%v", err)
	}
	// Store sessionID for later communication
	sessionID := createProcess.ID

	// Devicse answers with ProcessCreatedMessage
	processCreatedDevice := ssh.NewProcessCreatedMessage(sessionID)
	deviceConn.WriteJSON(processCreatedDevice)

	// Client should receive a ProcessCreatedMessage after device has created the process
	var processCreated ssh.ProcessCreatedMessage
	err = json.Unmarshal(<-clientC, &processCreated)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if processCreated.ID != sessionID {
		t.Fatalf("bad session id: " + processCreated.ID)
	}

	// Client sends a command message to the Device
	executeClient := ssh.NewExecuteCommandMessage(sessionID, []byte("whoami"))
	clientConn.WriteJSON(executeClient)

	// Device should receive a ExecuteCommandMessage
	var executeDevice ssh.ExecuteCommandMessage
	err = json.Unmarshal(<-deviceC, &executeDevice)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if executeDevice.ID != sessionID {
		t.Fatalf("bad session id: " + executeDevice.ID)
	}
	if string(executeDevice.Command) != "whoami" {
		t.Fatalf("bad command: " + string(executeDevice.Command))
	}

	// Device answers with ProcessOutputMessage
	processOutputDevice := ssh.NewProcessOutputMessage(sessionID, []byte("The great P."))
	deviceConn.WriteJSON(processOutputDevice)

	// Client should receive a ProcessOutputMessage
	var processOutput ssh.ProcessOutputMessage
	err = json.Unmarshal(<-clientC, &processOutput)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if processOutput.ID != sessionID {
		t.Fatalf("bad session id: " + processOutput.ID)
	}
	if string(processOutput.Data) != "The great P." {
		t.Fatalf("bad data: " + string(processOutput.Data))
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
