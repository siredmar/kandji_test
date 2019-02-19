package main

import (
	"fmt"
	"net/http"
	"testing"
)

func TestMaintenanceEndpoint(t *testing.T) {
	tt := []Testcase{{
		name:            "Get all maintenances",
		endpoint:        MaintenanceEndpoint,
		version:         APIVersion,
		method:          http.MethodGet,
		body:            nil,
		expectedCode:    200,
		compareResponse: true,
		expectedResponse: []byte(
			`{
				"maintenanceTasks":[]
			}`),
	}, {
		name:     "Create maintenance task",
		endpoint: MaintenanceEndpoint,
		version:  APIVersion,
		method:   http.MethodPost,
		body: []byte(
			`{
				"spec":{
					"type":"Restart",
					"selector":{
						"matchByLabels":{
							"foo":"bar"
						}
					}
				}
			}`),
		expectedCode:    201,
		compareResponse: true,
		expectedResponse: []byte(
			`{
				"metadata":{
					"id":"@UUID"
				},
				"spec":{
					"type":"Restart",
					"selector":{
						"matchByLabels":{
							"foo":"bar"
						}
					}
				},
				"status":{
					"successful":0,
					"failed":0,
					"running":0
				}
			}`),
	}, {
		name:     "Create maintenance task",
		endpoint: MaintenanceEndpoint,
		version:  APIVersion,
		method:   http.MethodPost,
		body: []byte(
			`{
				"spec":{
					"type":"Shutdown",
					"selector":{
						"matchByLabels":{
							"foo":"bar"
						},
						"matchByDeviceID":"123456"
					}
				}
		}`),
		expectedCode:    201,
		compareResponse: true,
		expectedResponse: []byte(
			`{
				"metadata":{
					"id":"@UUID"
				},
				"spec":{
					"type":"Shutdown",
					"selector":{
						"matchByLabels":{
							"foo":"bar"
						},
						"matchByDeviceID":"123456"
					}
				},
				"status":{
					"successful":0,
					"failed":0,
					"running":0
				}
			}`),
	}}
	ExtractedMaintenanceTaskUUIDs = ExtractedMaintenanceTaskUUIDs[:0] // Sanity
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			uuid := runTestcase(tc, t)
			if uuid != "" {
				ExtractedMaintenanceTaskUUIDs = append(ExtractedMaintenanceTaskUUIDs, uuid)
			}
		})
	}

	tt = []Testcase{{
		name:             "Get non existing task",
		endpoint:         fmt.Sprintf("%s/%s", MaintenanceEndpoint, "Nothingthere"),
		version:          APIVersion,
		method:           http.MethodGet,
		body:             nil,
		expectedCode:     404,
		wanntErr:         true,
		compareResponse:  false,
		expectedResponse: nil,
	}, {
		name:            "Get single task",
		endpoint:        fmt.Sprintf("%s/%s", MaintenanceEndpoint, ExtractedMaintenanceTaskUUIDs[0]),
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
					"type":"Restart",
					"selector":{
						"matchByLabels":{
							"foo":"bar"
						}
					}
				},
				"status":{
					"successful":0,
					"failed":0,
					"running":0
				}
			}`, ExtractedMaintenanceTaskUUIDs[0])),
	}, {
		name:            "Get all maintenances",
		endpoint:        MaintenanceEndpoint,
		version:         APIVersion,
		method:          http.MethodGet,
		body:            nil,
		expectedCode:    200,
		compareResponse: true,
		expectedResponse: []byte(fmt.Sprintf(
			`{
				"maintenanceTasks":[
					{
						"metadata":{
							"id":"%s"
						},
						"spec":{
							"type":"Restart",
							"selector":{
								"matchByLabels":{
									"foo":"bar"
								}
							}
						},
						"status":{
							"successful":0,
							"failed":0,
							"running":0
						}
					},
					{
						"metadata":{
							"id":"%s"
						},
						"spec":{
							"type":"Shutdown",
							"selector":{
								"matchByLabels":{
									"foo":"bar"
								},
								"matchByDeviceID":"123456"
							}
						},
						"status":{
							"successful":0,
							"failed":0,
							"running":0
						}
					}
				]
			}`, ExtractedMaintenanceTaskUUIDs[0], ExtractedMaintenanceTaskUUIDs[1])),
	}}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			runTestcase(tc, t)
		})
	}

	// Cleanup
	tt = []Testcase{{
		name:             "Delete task",
		endpoint:         fmt.Sprintf("%s/%s", MaintenanceEndpoint, ExtractedMaintenanceTaskUUIDs[0]),
		version:          APIVersion,
		method:           http.MethodDelete,
		body:             nil,
		expectedCode:     200,
		compareResponse:  true,
		expectedResponse: []byte(`{}`),
	}, {
		name:             "Delete task",
		endpoint:         fmt.Sprintf("%s/%s", MaintenanceEndpoint, ExtractedMaintenanceTaskUUIDs[1]),
		version:          APIVersion,
		method:           http.MethodDelete,
		body:             nil,
		expectedCode:     200,
		compareResponse:  true,
		expectedResponse: []byte(`{}`),
	}}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			runTestcase(tc, t)
		})
	}
}
