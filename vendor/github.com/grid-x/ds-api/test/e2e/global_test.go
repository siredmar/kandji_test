package main

import (
	"net/http"
	"testing"
)

func TestVersion(t *testing.T) {
	tt := []Testcase{{
		name:             "Without version",
		endpoint:         DevicesEndpoint,
		version:          "",
		method:           http.MethodGet,
		body:             nil,
		expectedCode:     400,
		wanntErr:         true,
		compareResponse:  false,
		expectedResponse: nil,
	}, {
		name:             "Wrong version",
		endpoint:         DevicesEndpoint,
		version:          "vnd.gridx.ai.2018-12-04",
		method:           http.MethodGet,
		body:             nil,
		expectedCode:     400,
		wanntErr:         true,
		compareResponse:  false,
		expectedResponse: nil,
	}, {
		name:             "Correct version",
		endpoint:         DevicesEndpoint,
		version:          APIVersion,
		method:           http.MethodGet,
		body:             nil,
		expectedCode:     200,
		wanntErr:         false,
		compareResponse:  false,
		expectedResponse: nil,
	}}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			runTestcase(tc, t)
		})
	}
}

func TestEndpoint(t *testing.T) {
	tt := []Testcase{{
		name:             "Wrong Endpoint",
		endpoint:         "/Nothingthere",
		version:          "",
		method:           http.MethodGet,
		body:             nil,
		expectedCode:     404,
		wanntErr:         true,
		compareResponse:  false,
		expectedResponse: nil,
	}}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			runTestcase(tc, t)
		})
	}
}
