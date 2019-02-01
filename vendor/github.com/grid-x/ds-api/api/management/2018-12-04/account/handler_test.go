package account

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/grid-x/ds-api/api"
	"github.com/grid-x/ds-api/pkg/encoding"
	"github.com/grid-x/ds-api/pkg/model"
)

type mockAuthProvider struct {
	f func(context.Context) (string, error)
}

func (m *mockAuthProvider) AccountIDFromContext(ctx context.Context) (string, error) {
	return m.f(ctx)
}

type getF func(ctx context.Context, accountID string) (*model.Account, error)
type updateF func(ctx context.Context, account *model.Account) (*model.Account, error)

type mockAccountRepo struct {
	get    getF
	update updateF
}

func (m *mockAccountRepo) Get(ctx context.Context, accountID string) (*model.Account, error) {
	return m.get(ctx, accountID)
}
func (m *mockAccountRepo) Update(ctx context.Context, account *model.Account) (*model.Account, error) {
	return m.update(ctx, account)
}

func mkString(s string) *string {
	r := new(string)
	*r = s
	return r
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
				&mockAccountRepo{
					get: func(ctx context.Context, accountID string) (*model.Account, error) {
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
				&mockAccountRepo{
					get: func(ctx context.Context, accountID string) (*model.Account, error) {
						return &model.Account{
							UUID: accountID,
							Name: "foobar",
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
				Payload: &GetAuthenticatedResponse{
					Account: &Account{
						Metadata: api.Metadata{
							ID: "default",
						},
						Name: "foobar",
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

func Test_UpdateAuthenticated_Validate(t *testing.T) {
	testcases := []struct {
		req        *http.Request
		injections []interface{}
		input      UpdateAuthenticatedRequest

		wantErr bool
		want    error
	}{
		{ // Test empty request
			req:        &http.Request{},
			injections: []interface{}{},
			input:      UpdateAuthenticatedRequest{},
			wantErr:    true,
			want:       nil,
		},
		{ // Test empty name
			req:        &http.Request{},
			injections: []interface{}{},
			input: UpdateAuthenticatedRequest{
				Name: "",
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

func Test_UpdateAuthenticated(t *testing.T) {
	testcases := []struct {
		req        *http.Request
		injections []interface{}
		input      UpdateAuthenticatedRequest

		wantErr bool
		want    *encoding.Response
	}{
		{
			req: &http.Request{},
			injections: []interface{}{
				&mockAccountRepo{
					update: func(ctx context.Context, account *model.Account) (*model.Account, error) {
						return nil, fmt.Errorf("not implemented")
					},
					get: func(ctx context.Context, accountID string) (*model.Account, error) {
						return nil, fmt.Errorf("not implemented")
					},
				},
				&mockAuthProvider{
					f: func(ctx context.Context) (string, error) {
						return "default", nil
					},
				},
			},
			input: UpdateAuthenticatedRequest{
				Name: "foobaz",
			},
			wantErr: true,
			want:    nil,
		},
		{
			req: &http.Request{},
			injections: []interface{}{
				&mockAccountRepo{
					update: func(ctx context.Context, account *model.Account) (*model.Account, error) {
						return account, nil
					},
					get: func(ctx context.Context, accountID string) (*model.Account, error) {
						return nil, fmt.Errorf("not found")
					},
				},
				&mockAuthProvider{
					f: func(ctx context.Context) (string, error) {
						return "default", nil
					},
				},
			},
			input: UpdateAuthenticatedRequest{
				Name: "foobaz",
			},
			wantErr: true,
			want:    nil,
		},
		{
			req: &http.Request{},
			injections: []interface{}{
				&mockAccountRepo{
					update: func(ctx context.Context, account *model.Account) (*model.Account, error) {
						return account, nil
					},
					get: func(ctx context.Context, accountID string) (*model.Account, error) {
						return &model.Account{
							UUID: accountID,
							Name: "foobar",
						}, nil
					},
				},
				&mockAuthProvider{
					f: func(ctx context.Context) (string, error) {
						return "default", nil
					},
				},
			},
			input: UpdateAuthenticatedRequest{
				Name: "foobaz",
			},
			wantErr: false,
			want: &encoding.Response{
				Payload: &UpdateAuthenticatedResponse{
					Account: &Account{
						Metadata: api.Metadata{
							ID: "default",
						},
						Name: "foobaz",
					},
				},
			},
		},
	}

	for i, tc := range testcases {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			svc := NewService(tc.injections...)
			got, err := svc.UpdateAuthenticated(tc.req, tc.input)

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
