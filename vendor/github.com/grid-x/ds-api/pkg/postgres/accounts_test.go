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
)

func TestAccountsList(t *testing.T) {

	testCases := []struct {
		name string
		opts *model.SortedListOptions
		want []*model.Account
	}{
		{
			name: "accept-without-list-options",
			want: []*model.Account{
				{
					UUID: "@ignore",
					Name: "account_00",
				},
				{
					UUID: "@ignore",
					Name: "account_01",
				},
				{
					UUID: "@ignore",
					Name: "account_02",
				},
				{
					UUID: "@ignore",
					Name: "account_03",
				},
				{
					UUID: "@ignore",
					Name: "account_04",
				},
				{
					UUID: "@ignore",
					Name: "account_05",
				},
				{
					UUID: "@ignore",
					Name: "account_06",
				},
				{
					UUID: "@ignore",
					Name: "account_07",
				},
				{
					UUID: "@ignore",
					Name: "account_08",
				},
				{
					UUID: "@ignore",
					Name: "account_09",
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
			want: []*model.Account{
				{
					UUID: "@ignore",
					Name: "account_02",
				},
				{
					UUID: "@ignore",
					Name: "account_03",
				},
			},
		},
		{
			name: "accept-desc-direction",
			opts: &model.SortedListOptions{
				Sort:  "account_name",
				Order: "desc",
				ListOptions: &model.ListOptions{
					Page:    1,
					PerPage: 3,
				},
			},
			want: []*model.Account{
				{
					UUID: "@ignore",
					Name: "account_09",
				},
				{
					UUID: "@ignore",
					Name: "account_08",
				},
				{
					UUID: "@ignore",
					Name: "account_07",
				},
			},
		},
		{
			name: "accept-asc-direction",
			opts: &model.SortedListOptions{
				Sort:  "account_name",
				Order: "asc",
				ListOptions: &model.ListOptions{
					Page:    1,
					PerPage: 3,
				},
			},
			want: []*model.Account{
				{
					UUID: "@ignore",
					Name: "account_00",
				},
				{
					UUID: "@ignore",
					Name: "account_01",
				},
				{
					UUID: "@ignore",
					Name: "account_02",
				},
			},
		},
		{
			name: "fallback-to-asc-direction",
			opts: &model.SortedListOptions{
				Sort:  "account_name",
				Order: "abc",
				ListOptions: &model.ListOptions{
					Page:    1,
					PerPage: 3,
				},
			},
			want: []*model.Account{
				{
					UUID: "@ignore",
					Name: "account_00",
				},
				{
					UUID: "@ignore",
					Name: "account_01",
				},
				{
					UUID: "@ignore",
					Name: "account_02",
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			InitTest()
			list, repo := createAccounts(10, t)
			var _ = list
			defer repo.Close()
			ctx := context.Background()
			items, _, err := repo.List(ctx, tc.opts)
			if err != nil {
				t.Fatal(err)
			}

			if len(items) != len(tc.want) {
				t.Errorf("accounts list length = %d, want %d", len(items), len(tc.want))
			}

			opts := cmpopts.IgnoreFields(model.Account{},
				"UUID", "CreatedAt", "UpdatedAt",
			)
			if !cmp.Equal(tc.want, items, opts) {
				t.Errorf("accounts list: %s ", cmp.Diff(tc.want, items, opts))
			}
		})
	}
}

func TestAccountsCreate(t *testing.T) {
	InitTest()
	repo, err := postgres.NewAccountsRepository(DB)
	if err != nil {
		t.Fatal(err)
	}
	defer repo.Close()

	testCases := []struct {
		name    string
		input   *model.Account
		wantErr bool
	}{
		{
			name:    "reject-without",
			wantErr: true,
		},
		{
			name: "accept-simple",
			input: &model.Account{
				Name: "account_00",
			},
			wantErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()

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

			opts := cmpopts.IgnoreFields(model.Account{},
				"CreatedAt", "UpdatedAt",
			)
			if !cmp.Equal(got, want, opts) {
				t.Errorf("create = %+v, want %+v", got, want)
			}

			if got.CreatedAt.String() == "0001-01-01 00:00:00 +0000 UTC" {
				t.Errorf("Accounts creation time should be set with now()")
			}
		})
	}
}

func TestAccountsUpdate(t *testing.T) {
	InitTest()
	items, repo := createAccounts(1, t)
	defer repo.Close()

	testCases := []struct {
		name    string
		item    *model.Account
		want    *model.Account
		wantErr bool
	}{
		{
			name:    "reject-without-item",
			wantErr: true,
		},
		{
			name:    "reject-without-id",
			item:    &model.Account{},
			wantErr: true,
		},
		{
			name: "update-simple",
			item: &model.Account{
				UUID: items[0].UUID,
				Name: "account_99",
			},
			wantErr: false,
			want: &model.Account{
				UUID: items[0].UUID,
				Name: "account_99",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()

			_, err := repo.Update(ctx, tc.item)
			if err != nil && !tc.wantErr {
				t.Fatal(err)
			} else if err == nil && tc.wantErr {
				t.Fatal("create expect error to be returned")
			} else if tc.wantErr {
				return
			}

			got, err := repo.GetByID(ctx, tc.item.UUID)
			if err != nil {
				t.Fatal(err)
			}

			opts := cmpopts.IgnoreFields(model.Account{},
				"CreatedAt", "UpdatedAt",
			)
			if !cmp.Equal(got, tc.want, opts) {
				t.Errorf("create = %+v, want %+v", got, tc.want)
			}

			if got.UpdatedAt.String() == "0001-01-01 00:00:00 +0000 UTC" {
				t.Errorf("Accounts update time should be set with now()")
			}
		})
	}
}

func TestAccountsDelete(t *testing.T) {
	testCases := []struct {
		name        string
		accountName string
		hard        bool
	}{
		{
			name:        "accept-soft-delete",
			accountName: "foo",
			hard:        false,
		},
		{
			name:        "accept-hard-delete",
			accountName: "foo",
			hard:        true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			InitTest()

			_, repo := createAccounts(10, t)
			defer repo.Close()

			wantItems, _, err := repo.List(context.Background(), nil)
			if err != nil {
				t.Fatal(err)
			}

			ctx := context.Background()
			item, err := repo.Create(ctx, &model.Account{
				Name: tc.accountName,
			})
			if err != nil {
				t.Fatal(err)
			}

			if err := repo.Delete(ctx, item.UUID, tc.hard); err != nil {
				t.Error(err)
			}

			got, err := repo.GetByIDWithSoftDeleted(ctx, item.UUID)
			if tc.hard && err != sql.ErrNoRows {
				t.Error(err)
			}
			if !tc.hard && err == sql.ErrNoRows {
				t.Error(err)
			}

			items, _, err := repo.List(ctx, nil)
			if err != nil {
				t.Fatal(err)
			}

			if len(items) != len(wantItems) {
				t.Errorf("Delete = %d, want %d", len(items), len(wantItems))
			}

			if !tc.hard && (got.DeletedAt == nil || got.DeletedAt.String() == "0001-01-01 00:00:00 +0000 UTC") {
				t.Errorf("Accounts soft deletion time should be set with now()")
			}
		})
	}

}

func TestAccountsPagination(t *testing.T) {

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
			_, repo := createAccounts(model.PerPageMax+1, t)
			defer repo.Close()

			items, hasNext, err := repo.List(context.Background(), tc.opts)

			if err != nil {
				t.Fatal("Failed to list accounts")
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

// createAccounts creates accounts. It returns the list of created
// accounts and the AccountsRepository
func createAccounts(count int, t *testing.T) ([]*model.Account, *postgres.AccountsRepository) {
	list := make([]*model.Account, count)
	repo, err := postgres.NewAccountsRepository(DB)
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < count; i++ {
		item, err := repo.Create(context.Background(), &model.Account{
			Name: fmt.Sprintf("account_%02d", i),
		})

		if err != nil {
			t.Fatalf("could not create account: %+v", err)
		}

		list[i] = item
	}

	return list, repo
}
