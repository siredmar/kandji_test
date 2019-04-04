package main

import (
	"fmt"
	"net/http"
	"testing"
)

func TestApplicationsEndpoint(t *testing.T) {
	tt := []Testcase{{
		name:            "Get all applications",
		endpoint:        ManagementApplicationsEndpoint,
		version:         APIVersion,
		method:          http.MethodGet,
		body:            nil,
		expectedCode:    200,
		wanntErr:        false,
		compareResponse: true,
		expectedResponse: []byte(
			`{
				"applications":[]
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
		name:     "Create application",
		endpoint: ManagementApplicationsEndpoint,
		version:  APIVersion,
		method:   http.MethodPost,
		body: []byte(
			`{
				"name":"testapp2"
			}`),
		expectedCode:    201,
		wanntErr:        false,
		compareResponse: true,
		expectedResponse: []byte(
			`{
				"metadata":{
					"id":"testapp2"
				},
				"name":"testapp2"
			}`),
	}, {
		name:             "Get non existing application",
		endpoint:         fmt.Sprintf("%s/%s", ManagementApplicationsEndpoint, "Nothingthere"),
		version:          APIVersion,
		method:           http.MethodGet,
		body:             nil,
		expectedCode:     404,
		wanntErr:         true,
		compareResponse:  false,
		expectedResponse: nil,
	}, {
		name:            "Get single application",
		endpoint:        fmt.Sprintf("%s/%s", ManagementApplicationsEndpoint, "testapp"),
		version:         APIVersion,
		method:          http.MethodGet,
		body:            nil,
		expectedCode:    200,
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
		name:            "Get all applications",
		endpoint:        ManagementApplicationsEndpoint,
		version:         APIVersion,
		method:          http.MethodGet,
		body:            nil,
		expectedCode:    200,
		wanntErr:        false,
		compareResponse: true,
		expectedResponse: []byte(
			`{
				"applications":[{
					"metadata":{
						"id":"testapp"
					},
					"name":"testapp"
				},{
					"metadata":{
						"id":"testapp2"
					},
					"name":"testapp2"
				}]
			}`),
	}}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			runTestcase(tc, t)
		})
	}

	// Cleanup
	tt = []Testcase{{
		name:             "Delete application",
		endpoint:         fmt.Sprintf("%s/%s", ManagementApplicationsEndpoint, "testapp"),
		version:          APIVersion,
		method:           http.MethodDelete,
		body:             nil,
		expectedCode:     200,
		wanntErr:         false,
		compareResponse:  true,
		expectedResponse: []byte(`{}`),
	}, {
		name:             "Delete application",
		endpoint:         fmt.Sprintf("%s/%s", ManagementApplicationsEndpoint, "testapp2"),
		version:          APIVersion,
		method:           http.MethodDelete,
		body:             nil,
		expectedCode:     200,
		wanntErr:         false,
		compareResponse:  true,
		expectedResponse: []byte(`{}`),
	}, {
		name:            "Get all applications",
		endpoint:        ManagementApplicationsEndpoint,
		version:         APIVersion,
		method:          http.MethodGet,
		body:            nil,
		expectedCode:    200,
		wanntErr:        false,
		compareResponse: true,
		expectedResponse: []byte(
			`{
				"applications":[]
			}`),
	}}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			runTestcase(tc, t)
		})
	}
}
