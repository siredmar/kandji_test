package postgres_test

import (
	"database/sql"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"golang.org/x/net/context"

	"github.com/grid-x/ds-api/pkg/model"
	"github.com/grid-x/ds-api/pkg/postgres"
	testutils "github.com/grid-x/ds-api/pkg/testing"
)

func mkString(s string) *string {
	r := new(string)
	*r = s
	return r
}

func TestGet(t *testing.T) {
	testCases := []struct {
		name       string
		identifier string
		want       *model.User
	}{
		{
			name:       "get-by-email",
			identifier: "first.last+02@gridx.ai",
			want: &model.User{
				UUID:      "@ignore",
				FirstName: mkString("First02"),
				LastName:  mkString("Last02"),
				Email:     "first.last+02@gridx.ai",
				AccountID: mkString("@ignore"),
				Auth0ID:   mkString("auth0|id-02"),
			},
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			ctx := context.Background()
			InitTest()
			accRepo, err := postgres.NewAccountsRepository(DB)
			if err != nil {
				t.Fatalf("cannot create accounts repo: %+v", err)
			}
			defer accRepo.Close()
			acc, err := accRepo.Create(ctx, &model.Account{
				Name: "foo",
			})
			if err != nil {
				t.Fatalf("cannot create account: %+v", err)
			}

			_, repo := createUsers(10, acc, t)
			defer repo.Close()
			got, err := repo.GetByEmail(ctx, tc.identifier)
			if err != nil {
				t.Fatal(err)
			}

			if !cmp.Equal(tc.want, got, testutils.CmpWithIgnore("UUID"), testutils.CmpWithIgnore("AccountID")) {
				t.Errorf("unexpected user: %s", cmp.Diff(tc.want, got, testutils.CmpWithIgnore("UUID"), testutils.CmpWithIgnore("AccountID")))
			}
		})
	}
}

func TestUsersList(t *testing.T) {

	testCases := []struct {
		name string
		opts *model.SortedListOptions
		want []*model.User
	}{
		{
			name: "accept-without-list-options",
			want: []*model.User{
				{
					UUID:      "@ignore",
					FirstName: mkString("First00"),
					LastName:  mkString("Last00"),
					Email:     "first.last+00@gridx.ai",
					AccountID: mkString("@ignore"),
					Auth0ID:   mkString("auth0|id-00"),
				},
				{
					UUID:      "@ignore",
					FirstName: mkString("First01"),
					LastName:  mkString("Last01"),
					Email:     "first.last+01@gridx.ai",
					AccountID: mkString("@ignore"),
					Auth0ID:   mkString("auth0|id-01"),
				},
				{
					UUID:      "@ignore",
					FirstName: mkString("First02"),
					LastName:  mkString("Last02"),
					Email:     "first.last+02@gridx.ai",
					AccountID: mkString("@ignore"),
					Auth0ID:   mkString("auth0|id-02"),
				},
				{
					UUID:      "@ignore",
					FirstName: mkString("First03"),
					LastName:  mkString("Last03"),
					Email:     "first.last+03@gridx.ai",
					AccountID: mkString("@ignore"),
					Auth0ID:   mkString("auth0|id-03"),
				},
				{
					UUID:      "@ignore",
					FirstName: mkString("First04"),
					LastName:  mkString("Last04"),
					Email:     "first.last+04@gridx.ai",
					AccountID: mkString("@ignore"),
					Auth0ID:   mkString("auth0|id-04"),
				},
				{
					UUID:      "@ignore",
					FirstName: mkString("First05"),
					LastName:  mkString("Last05"),
					Email:     "first.last+05@gridx.ai",
					AccountID: mkString("@ignore"),
					Auth0ID:   mkString("auth0|id-05"),
				},
				{
					UUID:      "@ignore",
					FirstName: mkString("First06"),
					LastName:  mkString("Last06"),
					Email:     "first.last+06@gridx.ai",
					AccountID: mkString("@ignore"),
					Auth0ID:   mkString("auth0|id-06"),
				},
				{
					UUID:      "@ignore",
					FirstName: mkString("First07"),
					LastName:  mkString("Last07"),
					Email:     "first.last+07@gridx.ai",
					AccountID: mkString("@ignore"),
					Auth0ID:   mkString("auth0|id-07"),
				},
				{
					UUID:      "@ignore",
					FirstName: mkString("First08"),
					LastName:  mkString("Last08"),
					Email:     "first.last+08@gridx.ai",
					AccountID: mkString("@ignore"),
					Auth0ID:   mkString("auth0|id-08"),
				},
				{
					UUID:      "@ignore",
					FirstName: mkString("First09"),
					LastName:  mkString("Last09"),
					Email:     "first.last+09@gridx.ai",
					AccountID: mkString("@ignore"),
					Auth0ID:   mkString("auth0|id-09"),
				},
			},
		},
		{
			name: "accept-per-page",
			opts: &model.SortedListOptions{
				ListOptions: &model.ListOptions{
					Page:    2,
					PerPage: 2,
				},
			},
			want: []*model.User{
				{
					UUID:      "@ignore",
					FirstName: mkString("First02"),
					LastName:  mkString("Last02"),
					Email:     "first.last+02@gridx.ai",
					AccountID: mkString("@ignore"),
					Auth0ID:   mkString("auth0|id-02"),
				},
				{
					UUID:      "@ignore",
					FirstName: mkString("First03"),
					LastName:  mkString("Last03"),
					Email:     "first.last+03@gridx.ai",
					AccountID: mkString("@ignore"),
					Auth0ID:   mkString("auth0|id-03"),
				},
			},
		},
		{
			name: "accept-desc-direction",
			opts: &model.SortedListOptions{
				Sort:  "email",
				Order: "desc",
				ListOptions: &model.ListOptions{
					Page:    1,
					PerPage: 3,
				},
			},
			want: []*model.User{
				{
					UUID:      "@ignore",
					FirstName: mkString("First09"),
					LastName:  mkString("Last09"),
					Email:     "first.last+09@gridx.ai",
					AccountID: mkString("@ignore"),
					Auth0ID:   mkString("auth0|id-09"),
				},
				{
					UUID:      "@ignore",
					FirstName: mkString("First08"),
					LastName:  mkString("Last08"),
					Email:     "first.last+08@gridx.ai",
					AccountID: mkString("@ignore"),
					Auth0ID:   mkString("auth0|id-08"),
				},
				{
					UUID:      "@ignore",
					FirstName: mkString("First07"),
					LastName:  mkString("Last07"),
					Email:     "first.last+07@gridx.ai",
					AccountID: mkString("@ignore"),
					Auth0ID:   mkString("auth0|id-07"),
				},
			},
		},
		{
			name: "accept-asc-direction",
			opts: &model.SortedListOptions{
				Sort:  "email",
				Order: "asc",
				ListOptions: &model.ListOptions{
					Page:    1,
					PerPage: 3,
				},
			},
			want: []*model.User{
				{
					UUID:      "@ignore",
					FirstName: mkString("First00"),
					LastName:  mkString("Last00"),
					Email:     "first.last+00@gridx.ai",
					AccountID: mkString("@ignore"),
					Auth0ID:   mkString("auth0|id-00"),
				},
				{
					UUID:      "@ignore",
					FirstName: mkString("First01"),
					LastName:  mkString("Last01"),
					Email:     "first.last+01@gridx.ai",
					AccountID: mkString("@ignore"),
					Auth0ID:   mkString("auth0|id-01"),
				},
				{
					UUID:      "@ignore",
					FirstName: mkString("First02"),
					LastName:  mkString("Last02"),
					Email:     "first.last+02@gridx.ai",
					AccountID: mkString("@ignore"),
					Auth0ID:   mkString("auth0|id-02"),
				},
			},
		},
		{
			name: "fallback-to-asc-direction",
			opts: &model.SortedListOptions{
				Sort:  "email",
				Order: "abc",
				ListOptions: &model.ListOptions{
					Page:    1,
					PerPage: 3,
				},
			},
			want: []*model.User{
				{
					UUID:      "@ignore",
					FirstName: mkString("First00"),
					LastName:  mkString("Last00"),
					Email:     "first.last+00@gridx.ai",
					AccountID: mkString("@ignore"),
					Auth0ID:   mkString("auth0|id-00"),
				},
				{
					UUID:      "@ignore",
					FirstName: mkString("First01"),
					LastName:  mkString("Last01"),
					Email:     "first.last+01@gridx.ai",
					AccountID: mkString("@ignore"),
					Auth0ID:   mkString("auth0|id-01"),
				},
				{
					UUID:      "@ignore",
					FirstName: mkString("First02"),
					LastName:  mkString("Last02"),
					Email:     "first.last+02@gridx.ai",
					AccountID: mkString("@ignore"),
					Auth0ID:   mkString("auth0|id-02"),
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			InitTest()
			accRepo, err := postgres.NewAccountsRepository(DB)
			if err != nil {
				t.Fatalf("cannot create accounts repo: %+v", err)
			}
			defer accRepo.Close()
			acc, err := accRepo.Create(ctx, &model.Account{
				Name: "foo",
			})
			if err != nil {
				t.Fatalf("cannot create account: %+v", err)
			}

			_, repo := createUsers(10, acc, t)
			defer repo.Close()
			items, _, err := repo.List(ctx, tc.opts)
			if err != nil {
				t.Fatal(err)
			}

			if len(items) != len(tc.want) {
				t.Errorf("users list length = %d, want %d", len(items), len(tc.want))
			}

			if !cmp.Equal(tc.want, items, testutils.CmpWithIgnore("UUID"), testutils.CmpWithIgnore("AccountID")) {
				t.Errorf("unexpected users list: %s", cmp.Diff(tc.want, items, testutils.CmpWithIgnore("UUID"), testutils.CmpWithIgnore("AccountID")))
			}
		})
	}
}

func TestUsersCreate(t *testing.T) {

	testCases := []struct {
		name    string
		input   *model.User
		acc     *model.Account
		wantErr bool
	}{
		{
			name:    "reject-without",
			wantErr: true,
		},
		{
			name: "reject-without-account-id",
			input: &model.User{
				Email: "foo.bar@gridx.ai",
			},
			wantErr: true,
		},
		{
			name: "accept-with-account",
			acc: &model.Account{
				Name: "foo",
			},
			input: &model.User{
				Email:   "foo.bar@gridx.ai",
				Auth0ID: mkString("auth0|id-foo"),
			},
			wantErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			InitTest()
			ctx := context.Background()
			accRepo, err := postgres.NewAccountsRepository(DB)
			if err != nil {
				t.Fatalf("cannot create accounts repo: %+v", err)
			}
			defer accRepo.Close()
			repo, err := postgres.NewUsersRepository(DB)
			if err != nil {
				t.Fatal(err)
			}
			defer repo.Close()

			if tc.acc != nil {
				acc, err := accRepo.Create(ctx, tc.acc)
				if err != nil {
					t.Fatalf("cannot create account: %+v", err)
				}
				tc.input.AccountID = &acc.UUID
			}

			want, err := repo.Create(ctx, tc.input)
			if err != nil && !tc.wantErr {
				t.Fatal(err)
			} else if err == nil && tc.wantErr {
				t.Fatal("create expect error to be returned")
			} else if tc.wantErr {
				return
			}

			got, err := repo.GetByID(ctx, want.UUID)
			if err != nil {
				t.Fatal(err)
			}

			opts := cmpopts.IgnoreFields(model.User{},
				"CreatedAt", "UpdatedAt",
			)
			if !cmp.Equal(got, want, opts) {
				t.Errorf("create = %+v, want %+v", got, want)
			}
		})
	}
}

func TestUsersUpdate(t *testing.T) {

	testCases := []struct {
		name    string
		create  *model.User
		update  *model.User
		want    *model.User
		wantErr bool
	}{
		{
			name:    "reject-without-item",
			wantErr: true,
		},
		{
			name:    "reject-without-id",
			update:  &model.User{},
			wantErr: true,
		},
		{
			name: "overwrite-first-name",
			create: &model.User{
				Email:   "foo.bar@gridx.de",
				Auth0ID: mkString("auth0|id-foo"),
			},
			update: &model.User{
				FirstName: mkString("Foo"),
				Email:     "foo.bar@gridx.de",
			},
			want: &model.User{
				UUID:      "@ignore",
				FirstName: mkString("Foo"),
				Email:     "foo.bar@gridx.de",
				AccountID: mkString("@ignore"),
				Auth0ID:   mkString("auth0|id-foo"),
			},
			wantErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			InitTest()
			ctx := context.Background()

			accRepo, err := postgres.NewAccountsRepository(DB)
			if err != nil {
				t.Fatalf("cannot create accounts repository: %+v", err)
			}
			acc, err := accRepo.Create(ctx, &model.Account{
				Name: "foo",
			})
			if err != nil {
				t.Fatalf("cannot create account: %+v", err)
			}
			if tc.create != nil {
				tc.create.AccountID = &acc.UUID
			}

			repo, err := postgres.NewUsersRepository(DB)
			if err != nil {
				t.Fatalf("cannot create users repo: %+v", err)
			}

			if tc.create != nil {
				user, err := repo.Create(ctx, tc.create)
				if err != nil {
					t.Fatalf("cannot create user: %+v", err)
				}
				if tc.update != nil {
					tc.update.UUID = user.UUID
				}
			}

			_, err = repo.Update(ctx, tc.update)
			if err != nil && !tc.wantErr {
				t.Fatal(err)
			} else if err == nil && tc.wantErr {
				t.Fatal("create expect error to be returned")
			} else if tc.wantErr {
				return
			}

			got, err := repo.GetByID(ctx, tc.update.UUID)
			if err != nil {
				t.Fatal(err)
			}

			opts := cmpopts.IgnoreFields(model.User{},
				"CreatedAt", "UpdatedAt",
			)
			if !cmp.Equal(got, tc.want, opts, testutils.CmpWithIgnore("UUID"), testutils.CmpWithIgnore("AccountID")) {
				t.Errorf("unexpected user: %s", cmp.Diff(tc.want, got, testutils.CmpWithIgnore("UUID"), testutils.CmpWithIgnore("AccountID")))
			}
		})
	}
}

func TestUsersDelete(t *testing.T) {

	InitTest()

	accRepo, err := postgres.NewAccountsRepository(DB)
	if err != nil {
		t.Fatalf("cannot create account repository: %+v", err)
	}
	defer accRepo.Close()

	ctx := context.Background()
	acc, err := accRepo.Create(ctx, &model.Account{
		Name: "foo",
	})
	if err != nil {
		t.Fatalf("cannot create account: %+v", err)
	}

	_, repo := createUsers(10, acc, t)

	testCases := []struct {
		name string
		hard bool
	}{
		{
			name: "accept-soft-delete",
			hard: false,
		},
		{
			name: "accept-hard-delete",
			hard: true,
		},
	}

	wantItems, _, err := repo.List(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()

			item, err := repo.Create(ctx, &model.User{
				Email:     "foo.bar@gridx.de",
				AccountID: &acc.UUID,
				Auth0ID:   mkString("auth0|id-foo"),
			})
			if err != nil {
				t.Fatal(err)
			}

			if err := repo.Delete(ctx, item.UUID, tc.hard); err != nil {
				t.Error(err)
			}

			_, err = repo.GetByID(ctx, item.UUID)
			if err != sql.ErrNoRows {
				t.Error(err)
			}

			items, _, err := repo.List(ctx, nil)
			if err != nil {
				t.Fatal(err)
			}

			if len(items) != len(wantItems) {
				t.Errorf("Delete = %d, want %d", len(items), len(wantItems))
			}
		})
	}

}

func TestUsersPagination(t *testing.T) {

	testCases := []struct {
		name        string
		opts        *model.SortedListOptions
		resultCount int
		hasNext     bool
	}{
		{
			name:        "No list options",
			opts:        nil,
			resultCount: model.PerPageDefault,
			hasNext:     true,
		},
		{
			name: "PerPage underflow",
			opts: &model.SortedListOptions{
				ListOptions: &model.ListOptions{Page: 0, PerPage: -20},
			},
			resultCount: model.PerPageMin,
			hasNext:     true,
		},
		{
			name: "PerPage right",
			opts: &model.SortedListOptions{
				ListOptions: &model.ListOptions{Page: 0, PerPage: 10},
			},
			resultCount: 10,
			hasNext:     true,
		},
		{
			name: "PerPage overflow",
			opts: &model.SortedListOptions{
				ListOptions: &model.ListOptions{Page: 0, PerPage: 1337},
			},
			resultCount: model.PerPageMax,
			hasNext:     true,
		},
		{
			name: "Page underflow",
			opts: &model.SortedListOptions{
				ListOptions: &model.ListOptions{Page: -10, PerPage: -20},
			},
			resultCount: model.PerPageMin,
			hasNext:     true,
		},
		{
			name: "Page right",
			opts: &model.SortedListOptions{
				ListOptions: &model.ListOptions{Page: 2, PerPage: 20},
			},
			resultCount: 20,
			hasNext:     true,
		},
		{
			name: "Page overflow",
			opts: &model.SortedListOptions{
				ListOptions: &model.ListOptions{Page: 100, PerPage: 20},
			},
			resultCount: 0,
			hasNext:     false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			InitTest()

			accRepo, err := postgres.NewAccountsRepository(DB)
			if err != nil {
				t.Fatalf("cannot create account repository: %+v", err)
			}
			defer accRepo.Close()

			ctx := context.Background()
			acc, err := accRepo.Create(ctx, &model.Account{
				Name: "foo",
			})

			_, repo := createUsers(model.PerPageMax+1, acc, t)
			defer repo.Close()

			items, hasNext, err := repo.List(context.Background(), tc.opts)

			if err != nil {
				t.Fatal("Failed to list users")
			}

			if len(items) != tc.resultCount {
				t.Fatalf("%s: len(items) = %d, expected %d", tc.name, len(items), tc.resultCount)
			}
			if hasNext != tc.hasNext {
				t.Fatalf("%s: hasNext = %v, expected %v", tc.name, hasNext, tc.hasNext)
			}
		})
	}
}

// createUsers creates users. It returns the list of created
// users and the UsersRepository
func createUsers(count int, acc *model.Account, t *testing.T) ([]*model.User, *postgres.UsersRepository) {
	list := make([]*model.User, count)
	repo, err := postgres.NewUsersRepository(DB)
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < count; i++ {
		item, err := repo.Create(context.Background(), &model.User{
			FirstName: mkString(fmt.Sprintf("First%02d", i)),
			LastName:  mkString(fmt.Sprintf("Last%02d", i)),
			Email:     fmt.Sprintf("first.last+%02d@gridx.ai", i),
			AccountID: &acc.UUID,
			Auth0ID:   mkString(fmt.Sprintf("auth0|id-%02d", i)),
		})

		if err != nil {
			t.Fatalf("could not create user: %+v", err)
		}

		list[i] = item
	}

	return list, repo
}
