package users

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/gorilla/mux"

	"github.com/grid-x/ds-api/api"
	"github.com/grid-x/ds-api/pkg/encoding"
	"github.com/grid-x/ds-api/pkg/model"
	testutils "github.com/grid-x/ds-api/pkg/testing"
)

type mockAuthProvider struct {
	f func(context.Context) (string, error)
	u func(context.Context) (string, error)
}

func (m *mockAuthProvider) AccountIDFromContext(ctx context.Context) (string, error) {
	return m.f(ctx)
}

func (m *mockAuthProvider) UserIDFromContext(ctx context.Context) (string, error) {
	return m.u(ctx)
}

type createF func(ctx context.Context, accountID string, user *model.User) (*model.User, error)
type updateF func(ctx context.Context, user *model.User) (*model.User, error)
type deleteF func(ctx context.Context, userID string) error
type getByIDF func(ctx context.Context, userID string) (*model.User, error)
type getByEmailF func(ctx context.Context, email string) (*model.User, error)
type listF func(ctx context.Context, accountID string) ([]*model.User, error)

type mockUsersRepo struct {
	create     createF
	update     updateF
	delete     deleteF
	getByID    getByIDF
	getByEmail getByEmailF
	list       listF
}

func (m *mockUsersRepo) Create(ctx context.Context, accountID string, user *model.User) (*model.User, error) {
	return m.create(ctx, accountID, user)
}
func (m *mockUsersRepo) Update(ctx context.Context, user *model.User) (*model.User, error) {
	return m.update(ctx, user)
}
func (m *mockUsersRepo) Delete(ctx context.Context, userID string) error {
	return m.delete(ctx, userID)
}
func (m *mockUsersRepo) GetByID(ctx context.Context, userID string) (*model.User, error) {
	return m.getByID(ctx, userID)
}
func (m *mockUsersRepo) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	return m.getByEmail(ctx, email)
}
func (m *mockUsersRepo) List(ctx context.Context, accountID string) ([]*model.User, error) {
	return m.list(ctx, accountID)
}

func mkString(s string) *string {
	r := new(string)
	*r = s
	return r
}

func newReq(method, id string, t *testing.T) *http.Request {
	req, err := http.NewRequest(method, "test.com", nil)
	if err != nil {
		t.Fatal(err)
	}
	return mux.SetURLVars(req, map[string]string{
		"userID": id,
	})
}

func Test_GetAuthenticated(t *testing.T) {
	testcases := []struct {
		req        *http.Request
		injections []interface{}

		wantErr bool
		want    *encoding.Response
	}{
		{
			req: &http.Request{},
			injections: []interface{}{
				&mockUsersRepo{
					getByID: func(ctx context.Context, userID string) (*model.User, error) {
						return nil, fmt.Errorf("not implemented")
					},
				},
				&mockAuthProvider{
					f: func(ctx context.Context) (string, error) {
						return "default", nil
					},
					u: func(ctx context.Context) (string, error) {
						return "2777841d-edb1-480e-9952-e9aa8859e0ac", nil
					},
				},
			},
			wantErr: true,
			want:    nil,
		},
		{
			req: &http.Request{},
			injections: []interface{}{
				&mockUsersRepo{
					getByID: func(ctx context.Context, userID string) (*model.User, error) {
						return &model.User{
							UUID:      userID,
							FirstName: mkString("John"),
							LastName:  mkString("Doe"),
							Email:     "j.doe@gridx.ai",
						}, nil
					},
				},
				&mockAuthProvider{
					f: func(ctx context.Context) (string, error) {
						return "default", nil
					},
					u: func(ctx context.Context) (string, error) {
						return "2777841d-edb1-480e-9952-e9aa8859e0ac", nil
					},
				},
			},
			wantErr: false,
			want: &encoding.Response{
				Payload: &GetAuthenticatedResponse{
					User: &User{
						Metadata: api.Metadata{
							ID: "2777841d-edb1-480e-9952-e9aa8859e0ac",
						},
						FirstName: mkString("John"),
						LastName:  mkString("Doe"),
						Email:     "j.doe@gridx.ai",
					},
				},
			},
		},
	}

	for i, tc := range testcases {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			svc := NewService(tc.injections...)
			got, err := svc.GetAuthenticated(tc.req)

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
				&mockUsersRepo{
					list: func(ctx context.Context, accountID string) ([]*model.User, error) {
						return nil, fmt.Errorf("not implemented")
					},
				},
				&mockAuthProvider{
					f: func(ctx context.Context) (string, error) {
						return "default", nil
					},
				},
			},

			wantErr: true,
			want:    nil,
		},
		{
			req: &http.Request{},
			injections: []interface{}{
				&mockUsersRepo{
					list: func(ctx context.Context, accountID string) ([]*model.User, error) {
						return []*model.User{
							{
								UUID:      "be670564-8d5c-468f-b1c0-3965b4ab3a64",
								FirstName: mkString("John"),
								LastName:  mkString("Doe"),
								Email:     "j.doe@gridx.ai",
							},
							{
								UUID:      "2cd8a8a8-3700-49d6-a0d4-2b36da1d5d9b",
								FirstName: mkString("Jane"),
								LastName:  mkString("Doe"),
								Email:     "ja.doe@gridx.ai",
							},
						}, nil
					},
				},
				&mockAuthProvider{
					f: func(ctx context.Context) (string, error) {
						return "default", nil
					},
				},
			},

			wantErr: false,
			want: &encoding.Response{
				Payload: &ListResponse{
					Users: []*User{
						{
							Metadata: api.Metadata{
								ID: "be670564-8d5c-468f-b1c0-3965b4ab3a64",
							},
							FirstName: mkString("John"),
							LastName:  mkString("Doe"),
							Email:     "j.doe@gridx.ai",
						},
						{
							Metadata: api.Metadata{
								ID: "2cd8a8a8-3700-49d6-a0d4-2b36da1d5d9b",
							},
							FirstName: mkString("Jane"),
							LastName:  mkString("Doe"),
							Email:     "ja.doe@gridx.ai",
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

func Test_Get(t *testing.T) {
	testcases := []struct {
		req        *http.Request
		injections []interface{}

		wantErr bool
		want    *encoding.Response
	}{
		{
			req: &http.Request{},
			injections: []interface{}{
				&mockUsersRepo{
					getByID: func(ctx context.Context, userID string) (*model.User, error) {
						return nil, fmt.Errorf("not implemented")
					},
				},
				&mockAuthProvider{
					f: func(ctx context.Context) (string, error) {
						return "default", nil
					},
				},
			},

			wantErr: true,
			want:    nil,
		},
		{
			req: newReq(http.MethodGet, "b9abdbc0-fcec-4264-9b37-4c589643d305", t),
			injections: []interface{}{
				&mockUsersRepo{
					getByID: func(ctx context.Context, userID string) (*model.User, error) {
						return &model.User{
							UUID:      userID,
							FirstName: mkString("John"),
							LastName:  mkString("Doe"),
							Email:     "j.doe@gridx.de",
						}, nil
					},
				},
				&mockAuthProvider{
					f: func(ctx context.Context) (string, error) {
						return "default", nil
					},
				},
			},

			wantErr: false,
			want: &encoding.Response{
				Payload: &GetResponse{
					User: &User{
						Metadata: api.Metadata{
							ID: "b9abdbc0-fcec-4264-9b37-4c589643d305",
						},
						FirstName: mkString("John"),
						LastName:  mkString("Doe"),
						Email:     "j.doe@gridx.de",
					},
				},
			},
		},
	}

	for i, tc := range testcases {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			svc := NewService(tc.injections...)
			got, err := svc.Get(tc.req)

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

func Test_Create_Validate(t *testing.T) {
	testcases := []struct {
		req        *http.Request
		injections []interface{}
		input      CreateRequest

		wantErr bool
		want    error
	}{
		{ // Test empty request
			req:        &http.Request{},
			injections: []interface{}{},
			input:      CreateRequest{},
			wantErr:    true,
			want:       nil,
		},
		{ // Test empty email
			req:        &http.Request{},
			injections: []interface{}{},
			input: CreateRequest{
				Email: "",
			},
			wantErr: true,
			want:    nil,
		},
	}

	for i, tc := range testcases {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			got := tc.input.Validate()

			if !tc.wantErr && got != nil {
				t.Fatalf("unexpected error: %+v", got)
			} else if tc.wantErr && got == nil {
				t.Fatal("expected error but did not get one")
			}
		})
	}
}

func Test_Create(t *testing.T) {
	testcases := []struct {
		req        *http.Request
		injections []interface{}
		input      CreateRequest

		wantErr bool
		want    *encoding.Response
	}{
		{
			req: &http.Request{},
			injections: []interface{}{
				&mockUsersRepo{
					create: func(ctx context.Context, accountID string, user *model.User) (*model.User, error) {
						return nil, fmt.Errorf("not implemented")
					},
					getByEmail: func(ctx context.Context, email string) (*model.User, error) {
						return nil, fmt.Errorf("not implemented")
					},
				},
				&mockAuthProvider{
					f: func(ctx context.Context) (string, error) {
						return "default", nil
					},
				},
			},
			input:   CreateRequest{},
			wantErr: true,
			want:    nil,
		},
		{ // User with email already exists
			req: &http.Request{},
			injections: []interface{}{
				&mockUsersRepo{
					create: func(ctx context.Context, accountID string, user *model.User) (*model.User, error) {
						return user, nil
					},
					getByEmail: func(ctx context.Context, email string) (*model.User, error) {
						return &model.User{
							UUID:      "be670564-8d5c-468f-b1c0-3965b4ab3a64",
							FirstName: mkString("Jane"),
							LastName:  mkString("Doe"),
							Email:     "jane.doe@gridx.ai",
						}, nil
					},
				},
				&mockAuthProvider{
					f: func(ctx context.Context) (string, error) {
						return "default", nil
					},
				},
			},
			input: CreateRequest{
				FirstName: mkString("Jane"),
				LastName:  mkString("Doe"),
				Email:     "jane.doe@gridx.ai",
			},
			wantErr: true,
			want:    nil,
		},
		{
			req: &http.Request{},
			injections: []interface{}{
				&mockUsersRepo{
					create: func(ctx context.Context, accountID string, user *model.User) (*model.User, error) {
						return user, nil
					},
					getByEmail: func(ctx context.Context, email string) (*model.User, error) {
						return nil, fmt.Errorf("not found")
					},
				},
				&mockAuthProvider{
					f: func(ctx context.Context) (string, error) {
						return "default", nil
					},
				},
			},
			input: CreateRequest{
				FirstName: mkString("Jane"),
				LastName:  mkString("Doe"),
				Email:     "jane.doe@gridx.ai",
			},
			wantErr: false,
			want: &encoding.Response{
				Status: 201,
				Payload: &CreateResponse{
					User: &User{
						Metadata: api.Metadata{
							ID: "@ignore",
						},
						FirstName: mkString("Jane"),
						LastName:  mkString("Doe"),
						Email:     "jane.doe@gridx.ai",
					},
				},
			},
		},
	}

	for i, tc := range testcases {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			svc := NewService(tc.injections...)
			got, err := svc.Create(tc.req, tc.input)

			if !tc.wantErr && err != nil {
				t.Fatalf("unexpected error: %+v", err)
			} else if tc.wantErr && err == nil {
				t.Fatal("expected error but did not get one")
			}

			if !cmp.Equal(tc.want, got, testutils.CmpWithIgnore("ID")) {
				t.Errorf("unexpected response: %s", cmp.Diff(tc.want, got, testutils.CmpWithIgnore("ID")))
			}
		})
	}
}

func Test_Update_Validate(t *testing.T) {
	testcases := []struct {
		req        *http.Request
		injections []interface{}
		input      UpdateRequest

		wantErr bool
		want    error
	}{
		{ // Test empty request
			req:        &http.Request{},
			injections: []interface{}{},
			input:      UpdateRequest{},
			wantErr:    true,
			want:       nil,
		},
		{ // Test empty firstname
			req:        &http.Request{},
			injections: []interface{}{},
			input: UpdateRequest{
				FirstName: mkString(""),
			},
			wantErr: true,
			want:    nil,
		},
		{ // Test empty lastname
			req:        &http.Request{},
			injections: []interface{}{},
			input: UpdateRequest{
				LastName: mkString(""),
			},
			wantErr: true,
			want:    nil,
		},
		{ // Test empty first an lastname
			req:        &http.Request{},
			injections: []interface{}{},
			input: UpdateRequest{
				FirstName: mkString(""),
				LastName:  mkString(""),
			},
			wantErr: true,
			want:    nil,
		},
	}

	for i, tc := range testcases {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			got := tc.input.Validate()

			if !tc.wantErr && got != nil {
				t.Fatalf("unexpected error: %+v", got)
			} else if tc.wantErr && got == nil {
				t.Fatal("expected error but did not get one")
			}
		})
	}
}

func Test_Update(t *testing.T) {
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
				&mockUsersRepo{
					update: func(ctx context.Context, user *model.User) (*model.User, error) {
						return nil, fmt.Errorf("not implemented")
					},
					getByID: func(ctx context.Context, userID string) (*model.User, error) {
						return nil, fmt.Errorf("not implemented")
					},
				},
				&mockAuthProvider{
					f: func(ctx context.Context) (string, error) {
						return "default", nil
					},
				},
			},
			input:   UpdateRequest{},
			wantErr: true,
			want:    nil,
		},
		{ // User not found
			req: newReq(http.MethodPatch, "bb9f4eef-e9f1-4994-9831-ab42efde243d", t),
			injections: []interface{}{
				&mockUsersRepo{
					update: func(ctx context.Context, user *model.User) (*model.User, error) {
						return user, nil
					},
					getByID: func(ctx context.Context, userID string) (*model.User, error) {
						return nil, fmt.Errorf("not found")
					},
				},
				&mockAuthProvider{
					f: func(ctx context.Context) (string, error) {
						return "default", nil
					},
				},
			},
			input: UpdateRequest{
				FirstName: mkString("Jane"),
				LastName:  mkString("Smitt"),
			},
			wantErr: true,
			want:    nil,
		},
		{
			req: newReq(http.MethodPatch, "bb9f4eef-e9f1-4994-9831-ab42efde243d", t),
			injections: []interface{}{
				&mockUsersRepo{
					update: func(ctx context.Context, user *model.User) (*model.User, error) {
						return user, nil
					},
					getByID: func(ctx context.Context, userID string) (*model.User, error) {
						return &model.User{
							UUID:      userID,
							FirstName: mkString("John"),
							LastName:  mkString("Doe"),
							Email:     "j.doe@gridx.ai",
						}, nil
					},
				},
				&mockAuthProvider{
					f: func(ctx context.Context) (string, error) {
						return "default", nil
					},
				},
			},
			input: UpdateRequest{
				FirstName: mkString("Jane"),
				LastName:  mkString("Smitt"),
			},
			wantErr: false,
			want: &encoding.Response{
				Payload: &UpdateResponse{
					User: &User{
						Metadata: api.Metadata{
							ID: "bb9f4eef-e9f1-4994-9831-ab42efde243d",
						},
						Email:     "j.doe@gridx.ai",
						FirstName: mkString("Jane"),
						LastName:  mkString("Smitt"),
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

func Test_Delete(t *testing.T) {
	testcases := []struct {
		req        *http.Request
		injections []interface{}

		wantErr bool
		want    *encoding.Response
	}{
		{
			req: &http.Request{},
			injections: []interface{}{
				&mockUsersRepo{
					delete: func(ctx context.Context, userID string) error {
						return fmt.Errorf("not implemented")
					},
					getByID: func(ctx context.Context, userID string) (*model.User, error) {
						return nil, fmt.Errorf("not implemented")
					},
				},
				&mockAuthProvider{
					f: func(ctx context.Context) (string, error) {
						return "default", nil
					},
				},
			},

			wantErr: true,
			want:    nil,
		},
		{ // User not found
			req: newReq(http.MethodDelete, "f52ff020-a64c-4a51-a818-a931f4a7b727", t),
			injections: []interface{}{
				&mockUsersRepo{
					delete: func(ctx context.Context, userID string) error {
						return nil
					},
					getByID: func(ctx context.Context, userID string) (*model.User, error) {
						return nil, fmt.Errorf("not found")
					},
				},
				&mockAuthProvider{
					f: func(ctx context.Context) (string, error) {
						return "default", nil
					},
				},
			},

			wantErr: true,
			want:    nil,
		},
		{
			req: newReq(http.MethodDelete, "f52ff020-a64c-4a51-a818-a931f4a7b727", t),
			injections: []interface{}{
				&mockUsersRepo{
					delete: func(ctx context.Context, userID string) error {
						return nil
					},
					getByID: func(ctx context.Context, userID string) (*model.User, error) {
						return &model.User{
							UUID:      userID,
							FirstName: mkString("John"),
							LastName:  mkString("Doe"),
							Email:     "j.doe@gridx.ai",
						}, nil
					},
				},
				&mockAuthProvider{
					f: func(ctx context.Context) (string, error) {
						return "default", nil
					},
				},
			},

			wantErr: false,
			want:    &encoding.Response{Payload: &DeleteResponse{}},
		},
	}

	for i, tc := range testcases {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			svc := NewService(tc.injections...)
			got, err := svc.Delete(tc.req)

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
