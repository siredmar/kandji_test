package pods

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/gorilla/mux"
	corev1beta1 "github.com/grid-x/ds-k8s/pkg/apis/core/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/grid-x/ds-api/pkg/encoding"
	"github.com/grid-x/ds-api/types"
	v20181203 "github.com/grid-x/ds-api/types/device/2018-12-03/pod"
)

type mockAuthProvider struct {
	deviceID, accountID string
}

func (m *mockAuthProvider) AccountIDFromContext(context.Context) (string, error) {
	return m.accountID, nil
}

func (m *mockAuthProvider) DeviceIDFromContext(context.Context) (string, error) {
	return m.deviceID, nil
}

type getFunc func(context.Context, string, string, bool) (*corev1beta1.DevicePod, error)
type listFunc func(context.Context, string, string, bool) ([]*corev1beta1.DevicePod, error)
type updateFunc func(context.Context, *corev1beta1.DevicePod) (*corev1beta1.DevicePod, error)

type mockPodClient struct {
	get    getFunc
	list   listFunc
	update updateFunc
}

func (m *mockPodClient) ListByDeviceID(ctx context.Context, namespace string, deviceID string, unfiltered bool) ([]*corev1beta1.DevicePod, error) {
	return m.list(ctx, namespace, deviceID, unfiltered)
}
func (m *mockPodClient) UpdateStatus(ctx context.Context, pod *corev1beta1.DevicePod) (*corev1beta1.DevicePod, error) {
	return m.update(ctx, pod)
}
func (m *mockPodClient) Get(ctx context.Context, namespace, name string, unfiltered bool) (*corev1beta1.DevicePod, error) {
	return m.get(ctx, namespace, name, unfiltered)
}

func newReq(id string, t *testing.T) *http.Request {
	req, err := http.NewRequest(http.MethodGet, "test.com", nil)
	if err != nil {
		t.Fatal(err)
	}
	return mux.SetURLVars(req, map[string]string{
		"podID": id,
	})
}

func Test_List(t *testing.T) {
	testcases := []struct {
		req        *http.Request
		injections []interface{}

		wantErr bool
		want    *encoding.Response
	}{
		{
			req: &http.Request{},
			injections: []interface{}{
				&mockAuthProvider{
					accountID: "default",
					deviceID:  "foo",
				},
				&mockPodClient{
					list: func(ctx context.Context, namespace, deviceID string, unfiltered bool) ([]*corev1beta1.DevicePod, error) {
						return nil, fmt.Errorf("not yet implemented")
					},
				},
			},

			wantErr: true,
			want:    nil,
		},
		{
			req: &http.Request{},
			injections: []interface{}{
				&mockAuthProvider{
					accountID: "default",
					deviceID:  "foo",
				},
				&mockPodClient{
					list: func(ctx context.Context, namespace, deviceID string, unfiltered bool) ([]*corev1beta1.DevicePod, error) {
						return []*corev1beta1.DevicePod{
							{
								ObjectMeta: metav1.ObjectMeta{
									Namespace: namespace,
									Name:      "894e3aa3-8beb-4e79-9785-3bd35fa671bc",
								},
								Spec: corev1beta1.DevicePodSpec{
									DeviceID: "ff8a5861-824b-4246-9457-94f04d665d7b",
									Config: corev1beta1.PodConfig{
										Volumes: []corev1beta1.Volume{
											{
												Name: "Testvolume",
												VolumeSource: corev1beta1.VolumeSource{
													HostPath: &corev1beta1.HostPathVolumeSource{
														Path: "Testpath",
														Type: func() *corev1beta1.HostPathType { t := corev1beta1.HostPathType("Type"); return &t }(),
													},
												},
											},
										},
										Containers: []corev1beta1.Container{
											{
												Name:    "Testcontainer",
												Image:   "Testimage",
												Command: []string{"Do", "Stuff"},
												Args:    []string{"Some", "Args"},
												Environment: []corev1beta1.EnvVar{
													{
														Name:  "Some",
														Value: "Variable",
													},
												},
												VolumeMounts: []corev1beta1.VolumeMount{
													{
														Name:      "Mount1",
														ReadOnly:  true,
														MountPath: "Path",
														SubPath:   "SubPath",
													},
												},
												WorkingDir: "/root",
												Ports: []corev1beta1.ContainerPort{
													{
														Name:          "Some port",
														HostPort:      80,
														ContainerPort: 80,
													},
												},
												LivelinessProbe: &corev1beta1.Probe{
													Handler: corev1beta1.Handler{
														Exec: &corev1beta1.ExecAction{
															Command: []string{"Do", "Stuff"},
														},
														HTTPGet: &corev1beta1.HTTPGetAction{
															Path:   "/test",
															Port:   func() *int32 { i := int32(8080); return &i }(),
															Host:   "Testhost",
															Scheme: corev1beta1.URIScheme("http://"),
															HTTPHeaders: []corev1beta1.HTTPHeader{
																{
																	Name:  "Header",
																	Value: "Headervalue",
																},
															},
														},
													},
													InitialDelaySeconds: 20,
													TimeoutSeconds:      20,
													PeriodSeconds:       20,
													SuccessThreshold:    20,
													FailureThreshold:    20,
												},
												ReadinessProbe: &corev1beta1.Probe{
													Handler: corev1beta1.Handler{
														Exec: &corev1beta1.ExecAction{
															Command: []string{"Do", "Stuff"},
														},
														HTTPGet: &corev1beta1.HTTPGetAction{
															Path:   "/test",
															Port:   func() *int32 { i := int32(8080); return &i }(),
															Host:   "Testhost",
															Scheme: corev1beta1.URIScheme("http://"),
															HTTPHeaders: []corev1beta1.HTTPHeader{
																{
																	Name:  "Header",
																	Value: "Headervalue",
																},
															},
														},
													},
													InitialDelaySeconds: 20,
													TimeoutSeconds:      20,
													PeriodSeconds:       20,
													SuccessThreshold:    20,
													FailureThreshold:    20,
												},
											},
										},
										Network:                       "Host",
										TerminationGracePeriodSeconds: func() *int64 { i := int64(60); return &i }(),
										Priority:                      func() *int32 { i := int32(30); return &i }(),
										RestartPolicy: &corev1beta1.RestartPolicy{
											Type: corev1beta1.RestartPolicyOnFailure,
										},
									},
								},
								Status: corev1beta1.DevicePodStatus{},
							},
							{
								ObjectMeta: metav1.ObjectMeta{
									Namespace: namespace,
									Name:      "e76f475d-4d30-4d5e-a10f-2a99afb24e87",
								},
								Spec: corev1beta1.DevicePodSpec{
									DeviceID: "ff8a5861-824b-4246-9457-94f04d665d7b",
									Config:   corev1beta1.PodConfig{},
								},
								Status: corev1beta1.DevicePodStatus{},
							},
						}, nil
					},
				},
			},

			wantErr: false,
			want: &encoding.Response{
				Payload: &ListResponse{
					Pods: []*v20181203.Pod{
						{
							Metadata: types.Metadata{
								ID: "894e3aa3-8beb-4e79-9785-3bd35fa671bc",
							},
							Spec: v20181203.DevicePodSpec{
								DeviceID: "ff8a5861-824b-4246-9457-94f04d665d7b",
								Config: v20181203.PodConfig{
									Volumes: []v20181203.Volume{
										{
											Name: "Testvolume",
											VolumeSource: v20181203.VolumeSource{
												HostPath: &v20181203.HostPathVolumeSource{
													Path: "Testpath",
													Type: func() *v20181203.HostPathType { t := v20181203.HostPathType("Type"); return &t }(),
												},
											},
										},
									},
									Containers: []v20181203.Container{
										{
											Name:    "Testcontainer",
											Image:   "Testimage",
											Command: []string{"Do", "Stuff"},
											Args:    []string{"Some", "Args"},
											Environment: []v20181203.EnvVar{
												{
													Name:  "Some",
													Value: "Variable",
												},
											},
											VolumeMounts: []v20181203.VolumeMount{
												{
													Name:      "Mount1",
													ReadOnly:  true,
													MountPath: "Path",
													SubPath:   "SubPath",
												},
											},
											WorkingDir: "/root",
											Ports: []v20181203.ContainerPort{
												{
													Name:          "Some port",
													HostPort:      80,
													ContainerPort: 80,
												},
											},
											LivelinessProbe: &v20181203.Probe{
												Handler: v20181203.Handler{
													Exec: &v20181203.ExecAction{
														Command: []string{"Do", "Stuff"},
													},
													HTTPGet: &v20181203.HTTPGetAction{
														Path:   "/test",
														Port:   func() *int32 { i := int32(8080); return &i }(),
														Host:   "Testhost",
														Scheme: v20181203.URIScheme("http://"),
														HTTPHeaders: []v20181203.HTTPHeader{
															{
																Name:  "Header",
																Value: "Headervalue",
															},
														},
													},
												},
												InitialDelaySeconds: 20,
												TimeoutSeconds:      20,
												PeriodSeconds:       20,
												SuccessThreshold:    20,
												FailureThreshold:    20,
											},
											ReadinessProbe: &v20181203.Probe{
												Handler: v20181203.Handler{
													Exec: &v20181203.ExecAction{
														Command: []string{"Do", "Stuff"},
													},
													HTTPGet: &v20181203.HTTPGetAction{
														Path:   "/test",
														Port:   func() *int32 { i := int32(8080); return &i }(),
														Host:   "Testhost",
														Scheme: v20181203.URIScheme("http://"),
														HTTPHeaders: []v20181203.HTTPHeader{
															{
																Name:  "Header",
																Value: "Headervalue",
															},
														},
													},
												},
												InitialDelaySeconds: 20,
												TimeoutSeconds:      20,
												PeriodSeconds:       20,
												SuccessThreshold:    20,
												FailureThreshold:    20,
											},
										},
									},
									Network:                       "Host",
									TerminationGracePeriodSeconds: func() *int64 { i := int64(60); return &i }(),
									Priority:                      func() *int32 { i := int32(30); return &i }(),
									RestartPolicy: &v20181203.RestartPolicy{
										Type: v20181203.RestartPolicyOnFailure,
									},
								},
							},
							Status: v20181203.DevicePodStatus{},
						},
						{
							Metadata: types.Metadata{
								ID: "e76f475d-4d30-4d5e-a10f-2a99afb24e87",
							},
							Spec: v20181203.DevicePodSpec{
								DeviceID: "ff8a5861-824b-4246-9457-94f04d665d7b",
								Config:   v20181203.PodConfig{},
							},
							Status: v20181203.DevicePodStatus{},
						},
					},
				},
			},
		},
	}

	for i, tc := range testcases {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			svc := NewService(tc.injections...)
			got, err := svc.List(tc.req)

			if !tc.wantErr && err != nil {
				t.Fatalf("unexpected error: %+v", err)
			} else if tc.wantErr && err == nil {
				t.Fatal("expected error but did not get one")
			}

			if !cmp.Equal(tc.want, got) {
				t.Errorf("unexpected response: %s", cmp.Diff(tc.want, got))
			}
		})
	}
}

func Test_Update(t *testing.T) {
	now := metav1.Now()
	testcases := []struct {
		req        *http.Request
		injections []interface{}
		input      UpdateRequest

		wantErr bool
		want    *encoding.Response
	}{
		{
			req: &http.Request{},
			injections: []interface{}{
				&mockAuthProvider{
					accountID: "default",
					deviceID:  "foo",
				},
				&mockPodClient{
					update: func(ctx context.Context, pod *corev1beta1.DevicePod) (*corev1beta1.DevicePod, error) {
						return nil, fmt.Errorf("not yet implemented")
					},
					get: func(ctx context.Context, namespace, name string, unfiltered bool) (*corev1beta1.DevicePod, error) {
						return nil, fmt.Errorf("not yet implemented")
					},
				},
			},
			input: UpdateRequest{
				Status: &v20181203.DevicePodStatus{
					StartTime: &now,
				},
			},
			wantErr: true,
			want:    nil,
		},
		{
			req: newReq("c91ffe91-e44b-4fa3-946a-b153d9a9ecbb", t),
			injections: []interface{}{
				&mockAuthProvider{
					accountID: "default",
					deviceID:  "foo",
				},
				&mockPodClient{
					update: func(ctx context.Context, pod *corev1beta1.DevicePod) (*corev1beta1.DevicePod, error) {
						return pod, nil
					},
					get: func(ctx context.Context, namespace, name string, unfiltered bool) (*corev1beta1.DevicePod, error) {
						return &corev1beta1.DevicePod{
							ObjectMeta: metav1.ObjectMeta{
								Namespace: namespace,
								Name:      "c91ffe91-e44b-4fa3-946a-b153d9a9ecbb",
							},
							Spec:   corev1beta1.DevicePodSpec{},
							Status: corev1beta1.DevicePodStatus{},
						}, nil
					},
				},
			},
			input: UpdateRequest{
				Status: &v20181203.DevicePodStatus{
					StartTime: &now,
				},
			},
			wantErr: false,
			want: &encoding.Response{
				Payload: &UpdateResponse{
					Pod: &v20181203.Pod{
						Metadata: types.Metadata{
							ID: "c91ffe91-e44b-4fa3-946a-b153d9a9ecbb",
						},
						Status: v20181203.DevicePodStatus{
							StartTime: &now,
						},
					},
				},
			},
		},
	}

	for i, tc := range testcases {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			svc := NewService(tc.injections...)
			got, err := svc.Update(tc.req, tc.input)

			if !tc.wantErr && err != nil {
				t.Fatalf("unexpected error: %+v", err)
			} else if tc.wantErr && err == nil {
				t.Fatal("expected error but did not get one")
			}

			if !cmp.Equal(tc.want, got) {
				t.Errorf("unexpected response: %s", cmp.Diff(tc.want, got))
			}
		})
	}
}
