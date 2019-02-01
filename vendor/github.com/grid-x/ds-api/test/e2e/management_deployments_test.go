package main

import (
	"fmt"
	"net/http"
	"testing"
)

func TestDeploymentsEndpoint(t *testing.T) {
	tt := []Testcase{{
		name:            "Get all deployments",
		endpoint:        DeploymentsEndpoint,
		version:         APIVersion,
		method:          http.MethodGet,
		body:            nil,
		expectedCode:    200,
		compareResponse: true,
		expectedResponse: []byte(
			`{
				"deployments":[]
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
							"foo":"bar"
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
							"foo":"bar"
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
								"foo":"bar"
							},
							"matchByDeviceID":"123456"
					},
					"template":{
							"spec":{
								"containers":[
										{
											"name":"monitoring-agent",
											"image":"gridx/monitoring:foobar-u28391389",
											"command":[
													"/usr/bin/foo",
													"-v",
													"-d"
											],
											"env":[
													{
														"name":"foo",
														"value":"bar"
													}
											],
											"volumeMounts":[
													{
														"name":"cache",
														"mountPath":"/data/cache",
														"subPath":"cache"
													}
											],
											"workingDir":"/opt",
											"ports":[
													{
														"name":"http",
														"hostPort":8080,
														"containerPort":8081
													}
											],
											"livelinessProbe":{
													"handler":{
														"exec":{
																"command":[
																	"exit 0"
																]
														}
													},
													"initialDelaySeconds":10,
													"timeoutSeconds":2,
													"periodSeconds":2,
													"successThreshold":3,
													"failureThreshold":3
											},
											"readinessProbe":{
													"handler":{
														"httpGet":{
																"path":"/health",
																"port":8081,
																"host":"gridbox.local",
																"scheme":"HTTP",
																"headers":[
																	{
																			"name":"User-Agent",
																			"value":"supervisor"
																	}
																]
														}
													},
													"initialDelaySeconds":10,
													"timeoutSeconds":2,
													"periodSeconds":2,
													"successThreshold":3,
													"failureThreshold":3
											}
										}
								],
								"network":"Host",
								"restartPolicy":{
										"type":"OnFailure",
										"maxRestartCount":10
								},
								"priority":10
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
					"app":"testapp",
					"selector":{
							"matchByLabels":{
								"foo":"bar"
							},
							"matchByDeviceID":"123456"
					},
					"template":{
							"spec":{
								"containers":[
										{
											"name":"monitoring-agent",
											"image":"gridx/monitoring:foobar-u28391389",
											"command":[
													"/usr/bin/foo",
													"-v",
													"-d"
											],
											"env":[
													{
														"name":"foo",
														"value":"bar"
													}
											],
											"volumeMounts":[
													{
														"name":"cache",
														"mountPath":"/data/cache",
														"subPath":"cache"
													}
											],
											"workingDir":"/opt",
											"ports":[
													{
														"name":"http",
														"hostPort":8080,
														"containerPort":8081
													}
											],
											"livelinessProbe":{
													"handler":{
														"exec":{
																"command":[
																	"exit 0"
																]
														}
													},
													"initialDelaySeconds":10,
													"timeoutSeconds":2,
													"periodSeconds":2,
													"successThreshold":3,
													"failureThreshold":3
											},
											"readinessProbe":{
													"handler":{
														"httpGet":{
																"path":"/health",
																"port":8081,
																"host":"gridbox.local",
																"scheme":"HTTP",
																"headers":[
																	{
																			"name":"User-Agent",
																			"value":"supervisor"
																	}
																]
														}
													},
													"initialDelaySeconds":10,
													"timeoutSeconds":2,
													"periodSeconds":2,
													"successThreshold":3,
													"failureThreshold":3
											}
										}
								],
								"network":"Host",
								"restartPolicy":{
										"type":"OnFailure",
										"maxRestartCount":10
								},
								"priority":10
							}
					}
				},
				"status":{}
			}`),
	}}
	ExtractedDeploymentUUIDs = ExtractedDeploymentUUIDs[:0] // Sanity
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			uuid := runTestcase(tc, t)
			if uuid != "" {
				ExtractedDeploymentUUIDs = append(ExtractedDeploymentUUIDs, uuid)
			}
		})
	}
	tt = []Testcase{{
		// Changing image to testbar-u28391389
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
							"foo":"bar"
						}
					},
					"template":{
						"spec":{
							"containers":[
									{
										"name":"monitoring-agent",
										"image":"gridx/monitoring:testbar-u28391389"
									}
							]
						}
					}
				}
			}`),
		expectedCode:    200,
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
							"foo":"bar"
						}
					},
					"template":{
						"spec":{
							"containers":[
									{
										"name":"monitoring-agent",
										"image":"gridx/monitoring:testbar-u28391389"
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

	// Cleanup
	tt = []Testcase{{
		name:             "Delete deployment",
		endpoint:         fmt.Sprintf("%s/%s", DeploymentsEndpoint, ExtractedDeploymentUUIDs[0]),
		version:          APIVersion,
		method:           http.MethodDelete,
		body:             nil,
		expectedCode:     200,
		compareResponse:  true,
		expectedResponse: []byte(`{}`),
	}, {
		name:             "Delete deployment",
		endpoint:         fmt.Sprintf("%s/%s", DeploymentsEndpoint, ExtractedDeploymentUUIDs[1]),
		version:          APIVersion,
		method:           http.MethodDelete,
		body:             nil,
		expectedCode:     200,
		compareResponse:  true,
		expectedResponse: []byte(`{}`),
	}, {
		name:             "Delete application",
		endpoint:         fmt.Sprintf("%s/%s", ApplicationsEndpoint, "testapp"),
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
