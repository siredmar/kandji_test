package main

import (
	"fmt"
	"net/http"
	"testing"
)

func TestDevicesEndpoint(t *testing.T) {
	tt := []Testcase{{
		name:            "Get all devices",
		endpoint:        ManagementDevicesEndpoint,
		version:         APIVersion,
		method:          http.MethodGet,
		body:            nil,
		expectedCode:    200,
		wanntErr:        false,
		compareResponse: true,
		expectedResponse: []byte(
			`{
				"devices":[]
			}`),
	}, {
		name:     "Create device",
		endpoint: ManagementDevicesEndpoint,
		version:  APIVersion,
		method:   http.MethodPost,
		body: []byte(
			`{
				"spec":{
					"serialnumber":"abcdefghijklmno-12",
					"macAddress":"00-14-22-01-23-45",
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
					"serialnumber":"abcdefghijklmno-12",
					"macAddress":"00-14-22-01-23-45",
					"publicKey":"ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC4H3QCpgn/3//9NMPWuOn69wDEgw8PnOqu4ucd7oJid3Yi0a8A63N1U+LvWwPrupKS5GE+D+Q2wDOraOzOHjWgAnZmI7JtqZCZgcDMTzUM6DmvUAwZZ0lCY1FqBHDtdWp/RWEDb6B4ZQIYxEvLqwZRGCCi5mE2aoMargbus6JKQrdtLJbOh4ybhm2F1aqlRmlhWmp4Hl2FvOHytMm5O8GxXMJ8TrjTzDvmMsqCgqAk2nqxe/oD1nTtcgHl/KSAhf/0mtfokMhCMFLLbBL/MtLfhBdXjob+BP6SWJS0E24a43OffnhgpvgY9vFwRdnZStM65khZXfQzzdx62mZdSs0f joel@deep-thought.local",
					"maintenanceWindow":"Sun:02:00-Sun:03:00"
				},
				"status":{}
			}`),
	}, {
		name:     "Create device",
		endpoint: ManagementDevicesEndpoint,
		version:  APIVersion,
		method:   http.MethodPost,
		body: []byte(
			`{
				"spec":{
					"Serialnumber":"abcdefghijklmno-34"
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
					"serialnumber":"abcdefghijklmno-34",
					"maintenanceWindow":"Sun:04:00-Sun:06:00"
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

	tt = []Testcase{{
		name:             "Get non existing device",
		endpoint:         fmt.Sprintf("%s/%s", ManagementDevicesEndpoint, "Nothingthere"),
		version:          APIVersion,
		method:           http.MethodGet,
		body:             nil,
		expectedCode:     404,
		wanntErr:         true,
		compareResponse:  false,
		expectedResponse: nil,
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
				"metadata":{
					"id":"%s"
				},
				"spec":{
					"serialnumber":"abcdefghijklmno-12",
					"macAddress":"00-14-22-01-23-45",
					"publicKey":"ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC4H3QCpgn/3//9NMPWuOn69wDEgw8PnOqu4ucd7oJid3Yi0a8A63N1U+LvWwPrupKS5GE+D+Q2wDOraOzOHjWgAnZmI7JtqZCZgcDMTzUM6DmvUAwZZ0lCY1FqBHDtdWp/RWEDb6B4ZQIYxEvLqwZRGCCi5mE2aoMargbus6JKQrdtLJbOh4ybhm2F1aqlRmlhWmp4Hl2FvOHytMm5O8GxXMJ8TrjTzDvmMsqCgqAk2nqxe/oD1nTtcgHl/KSAhf/0mtfokMhCMFLLbBL/MtLfhBdXjob+BP6SWJS0E24a43OffnhgpvgY9vFwRdnZStM65khZXfQzzdx62mZdSs0f joel@deep-thought.local",
					"maintenanceWindow":"Sun:02:00-Sun:03:00"
				},
				"status":{}
			}`, ExtractedDeviceUUIDs[0])),
	}, {
		// Changing macAddress to "99-88-77-66-55-55"
		// Changing maintenanceWindow to "Sun:22:00-Sun:23:00"
		name:     "Patch single device",
		endpoint: fmt.Sprintf("%s/%s", ManagementDevicesEndpoint, ExtractedDeviceUUIDs[0]),
		version:  APIVersion,
		method:   http.MethodPatch,
		body: []byte(
			`{
				"spec":{
					"macAddress":"99-88-77-66-55-55",
					"maintenanceWindow":"Sun:22:00-Sun:23:00"
				}
			}`),
		expectedCode:    200,
		wanntErr:        false,
		compareResponse: true,
		expectedResponse: []byte(fmt.Sprintf(
			`{
				"metadata":{
					"id":"%s"
				},
				"spec":{
					"serialnumber":"abcdefghijklmno-12",
					"macAddress":"99-88-77-66-55-55",
					"publicKey":"ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC4H3QCpgn/3//9NMPWuOn69wDEgw8PnOqu4ucd7oJid3Yi0a8A63N1U+LvWwPrupKS5GE+D+Q2wDOraOzOHjWgAnZmI7JtqZCZgcDMTzUM6DmvUAwZZ0lCY1FqBHDtdWp/RWEDb6B4ZQIYxEvLqwZRGCCi5mE2aoMargbus6JKQrdtLJbOh4ybhm2F1aqlRmlhWmp4Hl2FvOHytMm5O8GxXMJ8TrjTzDvmMsqCgqAk2nqxe/oD1nTtcgHl/KSAhf/0mtfokMhCMFLLbBL/MtLfhBdXjob+BP6SWJS0E24a43OffnhgpvgY9vFwRdnZStM65khZXfQzzdx62mZdSs0f joel@deep-thought.local",
					"maintenanceWindow":"Sun:22:00-Sun:23:00"
				},
				"status":{}
			}`, ExtractedDeviceUUIDs[0])),
	}, {
		// Adding label to "gridx.de/channel": "stable"
		name:     "Patch single device",
		endpoint: fmt.Sprintf("%s/%s", ManagementDevicesEndpoint, ExtractedDeviceUUIDs[0]),
		version:  APIVersion,
		method:   http.MethodPatch,
		body: []byte(
			`{
				"metadata":{
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
					"serialnumber":"abcdefghijklmno-12",
					"macAddress":"99-88-77-66-55-55",
					"publicKey":"ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC4H3QCpgn/3//9NMPWuOn69wDEgw8PnOqu4ucd7oJid3Yi0a8A63N1U+LvWwPrupKS5GE+D+Q2wDOraOzOHjWgAnZmI7JtqZCZgcDMTzUM6DmvUAwZZ0lCY1FqBHDtdWp/RWEDb6B4ZQIYxEvLqwZRGCCi5mE2aoMargbus6JKQrdtLJbOh4ybhm2F1aqlRmlhWmp4Hl2FvOHytMm5O8GxXMJ8TrjTzDvmMsqCgqAk2nqxe/oD1nTtcgHl/KSAhf/0mtfokMhCMFLLbBL/MtLfhBdXjob+BP6SWJS0E24a43OffnhgpvgY9vFwRdnZStM65khZXfQzzdx62mZdSs0f joel@deep-thought.local",
					"maintenanceWindow":"Sun:22:00-Sun:23:00"
				},
				"status":{}
			}`, ExtractedDeviceUUIDs[0])),
	}, {
		name:            "Get all devices",
		endpoint:        ManagementDevicesEndpoint,
		version:         APIVersion,
		method:          http.MethodGet,
		body:            nil,
		expectedCode:    200,
		wanntErr:        false,
		compareResponse: true,
		expectedResponse: []byte(fmt.Sprintf(
			`{
				"devices":[{
					"metadata":{
						"id":"%s",
						"labels": {"gridx.de/channel": "stable"}
					},
					"spec":{
						"serialnumber":"abcdefghijklmno-12",
						"macAddress":"99-88-77-66-55-55",
						"publicKey":"ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC4H3QCpgn/3//9NMPWuOn69wDEgw8PnOqu4ucd7oJid3Yi0a8A63N1U+LvWwPrupKS5GE+D+Q2wDOraOzOHjWgAnZmI7JtqZCZgcDMTzUM6DmvUAwZZ0lCY1FqBHDtdWp/RWEDb6B4ZQIYxEvLqwZRGCCi5mE2aoMargbus6JKQrdtLJbOh4ybhm2F1aqlRmlhWmp4Hl2FvOHytMm5O8GxXMJ8TrjTzDvmMsqCgqAk2nqxe/oD1nTtcgHl/KSAhf/0mtfokMhCMFLLbBL/MtLfhBdXjob+BP6SWJS0E24a43OffnhgpvgY9vFwRdnZStM65khZXfQzzdx62mZdSs0f joel@deep-thought.local",
						"maintenanceWindow":"Sun:22:00-Sun:23:00"
					},
					"status":{}
				},{

					"metadata":{
						"id":"%s"
					},
					"spec":{
						"serialnumber":"abcdefghijklmno-34",
						"maintenanceWindow":"Sun:04:00-Sun:06:00"
					},
					"status":{}
				}]
			}`, ExtractedDeviceUUIDs[0], ExtractedDeviceUUIDs[1])),
	}}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			runTestcase(tc, t)
		})
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
	}}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			runTestcase(tc, t)
		})
	}
}
