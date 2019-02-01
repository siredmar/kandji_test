package postgres

import (
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/satori/go.uuid"
	"golang.org/x/net/context"

	"github.com/grid-x/ds-api/pkg/model"
)

// SQL statements to manage accounts
const (
	accountsSelectBase = `
	SELECT
		uuid,
		created_at,
		updated_at,
		deleted_at,
		account_name
	FROM accounts
	`

	// accountsNotDeletedFilter filters out all deleted items.
	accountsNotDeletedFilter = `(deleted_at IS NULL OR deleted_at > now())`

	// accountsPaginationFilter filters
	accountsPaginationFilter = `OFFSET $1 LIMIT $2`

	accountsSelect = accountsSelectBase +
		` WHERE ` + accountsNotDeletedFilter + ` ORDER BY $sort $direction ` + accountsPaginationFilter

	accountsSelectByID = accountsSelectBase +
		` WHERE uuid = $1 AND ` + accountsNotDeletedFilter

	accountsInsert = `
	INSERT INTO accounts
	(
		uuid,
		created_at,
		updated_at,
		deleted_at,
		account_name
	)
	VALUES (
		:uuid,
		:created_at,
		:updated_at,
		:deleted_at,
		:account_name
	)
	`

	accountsUpdate = `
	UPDATE accounts
	SET
		account_name = :account_name,
		updated_at = DEFAULT
	WHERE uuid = :uuid
	`

	accountsSoftDelete = `
	UPDATE accounts
	SET
		updated_at = DEFAULT,
		deleted_at = now()
	WHERE uuid = $1
	`

	accountsHardDelete = `
	DELETE FROM accounts
	WHERE uuid = $1
	`
)

var (
	// accountsSortableColumns defines the columns which are sortable.
	// The first value is used as the default.
	accountsSortableColumns = []string{"created_at", "updated_at", "account_name"}
)

// AccountsRepository manages the access to the underlying postgres database.
type AccountsRepository struct {
	lists   map[string]*sqlx.Stmt
	getByID *sqlx.Stmt

	create *sqlx.NamedStmt
	update *sqlx.NamedStmt

	softDelete *sqlx.Stmt
	hardDelete *sqlx.Stmt
}

// NewAccountsRepository creates a AccountsRepository for accessing the
// related tables.
func NewAccountsRepository(db *sqlx.DB) (*AccountsRepository, error) {
	ctx := context.Background()

	lists := make(map[string]*sqlx.Stmt)
	for _, s := range accountsSortableColumns {
		for _, d := range model.AllOrders {
			var err error
			key, query := prepareOrderQuery(accountsSelect, s, d)
			lists[key], err = db.PreparexContext(ctx, query)
			if err != nil {
				return nil, err
			}
		}
	}

	getByID, err := db.PreparexContext(ctx, accountsSelectByID)
	if err != nil {
		return nil, err
	}

	create, err := db.PrepareNamedContext(ctx, accountsInsert)
	if err != nil {
		return nil, err
	}
	update, err := db.PrepareNamedContext(ctx, accountsUpdate)
	if err != nil {
		return nil, err
	}

	softDelete, err := db.PreparexContext(ctx, accountsSoftDelete)
	if err != nil {
		return nil, err
	}
	hardDelete, err := db.PreparexContext(ctx, accountsHardDelete)
	if err != nil {
		return nil, err
	}

	return &AccountsRepository{
		lists:      lists,
		getByID:    getByID,
		create:     create,
		update:     update,
		softDelete: softDelete,
		hardDelete: hardDelete,
	}, nil
}

// Close closes all prepared statements.
func (r *AccountsRepository) Close() error {
	var final error
	for _, list := range r.lists {
		if err := list.Close(); err != nil {
			final = err
		}
	}
	if err := r.getByID.Close(); err != nil {
		final = err
	}
	if err := r.create.Close(); err != nil {
		final = err
	}
	if err := r.update.Close(); err != nil {
		final = err
	}
	if err := r.softDelete.Close(); err != nil {
		final = err
	}
	if err := r.hardDelete.Close(); err != nil {
		final = err
	}
	return final
}

// List returns a set of accounts
func (r *AccountsRepository) List(ctx context.Context, opts *model.SortedListOptions) ([]*model.Account, bool, error) {
	opts = opts.Sanitize(accountsSortableColumns[0], accountsSortableColumns)
	entities := []*model.Account{}
	hasNext := false

	q, ok := r.lists[fmt.Sprintf("%s:%s", opts.Sort, opts.Order)]
	if !ok {
		return nil, false, fmt.Errorf("couldn't find appropriate query")
	}
	err := q.SelectContext(ctx, &entities, (opts.Page-1)*opts.PerPage, opts.PerPage+1)

	// Check if we will have another page of results
	if len(entities) > opts.PerPage {
		entities = entities[:opts.PerPage]
		hasNext = true
	}

	return entities, hasNext, err
}

// GetByID fetches a accounts by its id.
func (r *AccountsRepository) GetByID(ctx context.Context, id string) (*model.Account, error) {
	if id == "" {
		return nil, fmt.Errorf("no model.Account uuid provided")
	}
	var entity model.Account
	err := r.getByID.GetContext(ctx, &entity, id)
	if err != nil {
		return nil, err
	}
	return &entity, nil
}

// Create creates a new accounts
func (r *AccountsRepository) Create(ctx context.Context, e *model.Account) (*model.Account, error) {

	if e == nil {
		return nil, fmt.Errorf("couldn't create nil object")
	}

	// New ID for the accounts
	e.UUID = uuid.NewV4().String()
	_, err := NamedStmt(ctx, r.create).ExecContext(ctx, e)
	if err != nil {
		return nil, err
	}
	return e, nil
}

// Update updates the accounts
func (r *AccountsRepository) Update(ctx context.Context, e *model.Account) (*model.Account, error) {
	if e == nil {
		return nil, fmt.Errorf("couldn't create nil object")
	}
	if e.UUID == "" {
		return nil, fmt.Errorf("accounts uuid is missing")
	}
	_, err := NamedStmt(ctx, r.update).ExecContext(ctx, e)
	if err != nil {
		return nil, err
	}
	return e, nil
}

// Delete marks a accounts as deleted. If hard is set, it deletes the actual table row.
func (r *AccountsRepository) Delete(ctx context.Context, id string, hard bool) error {
	if id == "" {
		return fmt.Errorf("accounts id is missing")
	}
	if hard {
		result, err := Stmt(ctx, r.hardDelete).ExecContext(ctx, id)
		if count, _ := result.RowsAffected(); count == 0 {
			return sql.ErrNoRows
		}
		return err
	}
	result, err := Stmt(ctx, r.softDelete).ExecContext(ctx, id)
	if count, _ := result.RowsAffected(); count == 0 {
		return sql.ErrNoRows
	}
	return err
}
