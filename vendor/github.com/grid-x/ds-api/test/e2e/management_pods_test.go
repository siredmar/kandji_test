package main

import (
	"fmt"
	"net/http"
	"testing"
	"time"
)

func TestPodsEndpoint(t *testing.T) {
	tt := []Testcase{{
		name:            "Get all pods",
		endpoint:        PodsEndpoint,
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
	}}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			runTestcase(tc, t)
		})
	}

	// Creating Prereqs (Device, App and Deployment)
	tt = []Testcase{{
		name:     "Create device",
		endpoint: DevicesEndpoint,
		version:  APIVersion,
		method:   http.MethodPost,
		body: []byte(
			`{
				"spec":{
					"serialnumber":"abcdefghijklmno-12",
					"macAddress":"99-88-77-66-55-44",
					"publicKey":"ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQDOYFVhg4wr2hr83JjE9C4TOk8/a/4HdX8oA6Rif8JyszqfUX5KKT9DrqD4YzDkHK49ofg/At8pNoNto23pZ5QoK0bkrlf7tpDYX9PP37bt2sGf8OkJq5lEepKKCRP69wAT/t7/6SCkIsVqr4DugLgCxzUqI4aC37boLSmhx4TonjpSK2nmTXLYNhBV2w+SN4DNpvk0GpoxgeKfE4LpGX3zNseqAM8yfLSrUfE5hTbIa/6cAnVxobDYXND8ewBINftF3VZVe9t9xIH8+NzhXEEpYLlSVYZOHZkfY4Ma0UU9csfiCrbsYKRvnbk3V1LU2CrxpDz5fIJcGS3WbkVVb0Xj530sxTubA+hvY0YRl4Rld9KonBgdx36BWlFKLVuIM49YejOpenUi4ghgQEUfW9B6oiu6nYogTmsX1TE7Kft3Ex5jjM5Fper1VkEgeV8Rp6NGyHvuVwiRWMmWN6T3Ava3WySRYBcS61rFabYDHI+kc4lYr3gtoGFhoMM/ivR4n4KlMn+umkv0ILsUh5yW7s4Xih4mZHGjIzlHO82HsOKz2u7PEvt9iibVH94tcBO9Q7n7+6b67pdTiMYL5C5XcdHTjxEfuy24vc81anHFwkjrCrgpxb49ASAm0wwcOQvMS2zDmUX/qJYv7jBiGEEK/wTts0qvXzTvRvvh8U8CPQArBw==",
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
					"macAddress":"99-88-77-66-55-44",
					"publicKey":"ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQDOYFVhg4wr2hr83JjE9C4TOk8/a/4HdX8oA6Rif8JyszqfUX5KKT9DrqD4YzDkHK49ofg/At8pNoNto23pZ5QoK0bkrlf7tpDYX9PP37bt2sGf8OkJq5lEepKKCRP69wAT/t7/6SCkIsVqr4DugLgCxzUqI4aC37boLSmhx4TonjpSK2nmTXLYNhBV2w+SN4DNpvk0GpoxgeKfE4LpGX3zNseqAM8yfLSrUfE5hTbIa/6cAnVxobDYXND8ewBINftF3VZVe9t9xIH8+NzhXEEpYLlSVYZOHZkfY4Ma0UU9csfiCrbsYKRvnbk3V1LU2CrxpDz5fIJcGS3WbkVVb0Xj530sxTubA+hvY0YRl4Rld9KonBgdx36BWlFKLVuIM49YejOpenUi4ghgQEUfW9B6oiu6nYogTmsX1TE7Kft3Ex5jjM5Fper1VkEgeV8Rp6NGyHvuVwiRWMmWN6T3Ava3WySRYBcS61rFabYDHI+kc4lYr3gtoGFhoMM/ivR4n4KlMn+umkv0ILsUh5yW7s4Xih4mZHGjIzlHO82HsOKz2u7PEvt9iibVH94tcBO9Q7n7+6b67pdTiMYL5C5XcdHTjxEfuy24vc81anHFwkjrCrgpxb49ASAm0wwcOQvMS2zDmUX/qJYv7jBiGEEK/wTts0qvXzTvRvvh8U8CPQArBw==",
					"maintenanceWindow":"Sun:02:00-Sun:03:00"
				},
				"status":{}
			}`),
	}, {
		name:     "Create device",
		endpoint: DevicesEndpoint,
		version:  APIVersion,
		method:   http.MethodPost,
		body: []byte(
			`{
				"spec":{
					"serialnumber":"abcdefghijklmno-34",
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
					"serialnumber":"abcdefghijklmno-34",
					"macAddress":"00-14-22-01-23-45",
					"publicKey":"ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC4H3QCpgn/3//9NMPWuOn69wDEgw8PnOqu4ucd7oJid3Yi0a8A63N1U+LvWwPrupKS5GE+D+Q2wDOraOzOHjWgAnZmI7JtqZCZgcDMTzUM6DmvUAwZZ0lCY1FqBHDtdWp/RWEDb6B4ZQIYxEvLqwZRGCCi5mE2aoMargbus6JKQrdtLJbOh4ybhm2F1aqlRmlhWmp4Hl2FvOHytMm5O8GxXMJ8TrjTzDvmMsqCgqAk2nqxe/oD1nTtcgHl/KSAhf/0mtfokMhCMFLLbBL/MtLfhBdXjob+BP6SWJS0E24a43OffnhgpvgY9vFwRdnZStM65khZXfQzzdx62mZdSs0f joel@deep-thought.local",
					"maintenanceWindow":"Sun:02:00-Sun:03:00"
				},
				"status":{}
			}`),
	}, {
		name:     "Create application",
		endpoint: ApplicationsEndpoint,
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
		endpoint: DeploymentsEndpoint,
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
				if tc.endpoint == DevicesEndpoint {
					ExtractedDeviceUUIDs = append(ExtractedDeviceUUIDs, uuid)
				}
				if tc.endpoint == DeploymentsEndpoint {
					ExtractedDeploymentUUIDs = append(ExtractedDeploymentUUIDs, uuid)
				}
			}
		})
	}

	// Checking that there are still no pods and adding the appropiate label to the device
	tt = []Testcase{{
		name:            "Get all pods",
		endpoint:        PodsEndpoint,
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
		// Adding label to "gridx.de/channel": "stable"
		name:     "Patch single device",
		endpoint: fmt.Sprintf("%s/%s", DevicesEndpoint, ExtractedDeviceUUIDs[0]),
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
					"serialnumber":"abcdefghijklmno-12",
					"macAddress":"99-88-77-66-55-44",
					"publicKey":"ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQDOYFVhg4wr2hr83JjE9C4TOk8/a/4HdX8oA6Rif8JyszqfUX5KKT9DrqD4YzDkHK49ofg/At8pNoNto23pZ5QoK0bkrlf7tpDYX9PP37bt2sGf8OkJq5lEepKKCRP69wAT/t7/6SCkIsVqr4DugLgCxzUqI4aC37boLSmhx4TonjpSK2nmTXLYNhBV2w+SN4DNpvk0GpoxgeKfE4LpGX3zNseqAM8yfLSrUfE5hTbIa/6cAnVxobDYXND8ewBINftF3VZVe9t9xIH8+NzhXEEpYLlSVYZOHZkfY4Ma0UU9csfiCrbsYKRvnbk3V1LU2CrxpDz5fIJcGS3WbkVVb0Xj530sxTubA+hvY0YRl4Rld9KonBgdx36BWlFKLVuIM49YejOpenUi4ghgQEUfW9B6oiu6nYogTmsX1TE7Kft3Ex5jjM5Fper1VkEgeV8Rp6NGyHvuVwiRWMmWN6T3Ava3WySRYBcS61rFabYDHI+kc4lYr3gtoGFhoMM/ivR4n4KlMn+umkv0ILsUh5yW7s4Xih4mZHGjIzlHO82HsOKz2u7PEvt9iibVH94tcBO9Q7n7+6b67pdTiMYL5C5XcdHTjxEfuy24vc81anHFwkjrCrgpxb49ASAm0wwcOQvMS2zDmUX/qJYv7jBiGEEK/wTts0qvXzTvRvvh8U8CPQArBw==",
					"maintenanceWindow":"Sun:02:00-Sun:03:00"
				},
				"status":{}
			}`, ExtractedDeviceUUIDs[0])),
	}, {
		// Adding label to "gridx.de/channel": "stable"
		name:     "Patch single device",
		endpoint: fmt.Sprintf("%s/%s", DevicesEndpoint, ExtractedDeviceUUIDs[1]),
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
					"serialnumber":"abcdefghijklmno-34",
					"macAddress":"00-14-22-01-23-45",
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

	// Checking that there are now pods available for the second device matching alpha group
	tt = []Testcase{{
		name:            "Get all pods",
		endpoint:        PodsEndpoint,
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
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			runTestcase(tc, t)
		})
	}

	// Patching the deployment matchByLabels to check if pods getting removed on first device and scheduled on seconed one
	tt = []Testcase{{
		name:     "Patch single deployment",
		endpoint: fmt.Sprintf("%s/%s", DeploymentsEndpoint, ExtractedDeploymentUUIDs[0]),
		version:  APIVersion,
		method:   http.MethodPatch,
		body: []byte(
			`{
						"spec":{
							"app":"testapp",
							"selector":{
								"matchByLabels":{
									"gridx.de/channel":"stable"
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
		expectedCode:    200,
		wanntErr:        false,
		compareResponse: true,
		expectedResponse: []byte(fmt.Sprintf(
			`{
						"metadata":{
							"id":"%s"
						},
						"spec":{
							"app":"testapp",
							"selector":{
								"matchByLabels":{
									"gridx.de/channel":"stable"
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
					}`, ExtractedDeploymentUUIDs[0])),
	}}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			runTestcase(tc, t)
		})
	}

	// Wait for pods to be created
	time.Sleep(time.Second * 5)

	// Checking that there are now pods available for the second device matching stable group
	tt = []Testcase{{
		name:            "Get all pods",
		endpoint:        PodsEndpoint,
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
				 }`, ExtractedDeviceUUIDs[1])),
	}}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			runTestcase(tc, t)
		})
	}
	// Cleanup
	tt = []Testcase{{
		name:             "Delete deployment",
		endpoint:         fmt.Sprintf("%s/%s", DeploymentsEndpoint, ExtractedDeploymentUUIDs[0]),
		version:          APIVersion,
		method:           http.MethodDelete,
		body:             nil,
		expectedCode:     200,
		wanntErr:         false,
		compareResponse:  true,
		expectedResponse: []byte(`{}`),
	}, {
		name:             "Delete device",
		endpoint:         fmt.Sprintf("%s/%s", DevicesEndpoint, ExtractedDeviceUUIDs[0]),
		version:          APIVersion,
		method:           http.MethodDelete,
		body:             nil,
		expectedCode:     200,
		wanntErr:         false,
		compareResponse:  true,
		expectedResponse: []byte(`{}`),
	}, {
		name:             "Delete device",
		endpoint:         fmt.Sprintf("%s/%s", DevicesEndpoint, ExtractedDeviceUUIDs[1]),
		version:          APIVersion,
		method:           http.MethodDelete,
		body:             nil,
		expectedCode:     200,
		wanntErr:         false,
		compareResponse:  true,
		expectedResponse: []byte(`{}`),
	}, {
		name:             "Delete application",
		endpoint:         fmt.Sprintf("%s/%s", ApplicationsEndpoint, "testapp"),
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
