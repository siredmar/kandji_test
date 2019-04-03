package postgres

import (
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"golang.org/x/net/context"

	"github.com/grid-x/ds-api/pkg/model"
)

// SQL statements to manage users
const (
	usersSelectBase = `
	SELECT
		uuid,
		created_at,
		updated_at,
		deleted_at,
		account_id,
		first_name,
		last_name,
		email,
		auth0_id
	FROM users
	`

	// usersNotDeletedFilter filters out all deleted items.
	usersNotDeletedFilter = `(deleted_at IS NULL OR deleted_at > now())`

	// usersPaginationFilter filters
	usersPaginationFilter = `OFFSET $1 LIMIT $2`

	usersSelect = usersSelectBase +
		` WHERE ` + usersNotDeletedFilter + ` ORDER BY $sort $direction ` + usersPaginationFilter

	usersSelectByID = usersSelectBase +
		` WHERE uuid = $1 AND ` + usersNotDeletedFilter

	usersSelectByIDWithSoftDeleted = usersSelectBase +
		` WHERE uuid = $1`

	usersSelectByEmail = usersSelectBase +
		` WHERE email = $1 AND ` + usersNotDeletedFilter

	usersSelectByAuth0ID = usersSelectBase +
		` WHERE auth0_id = $1 AND ` + usersNotDeletedFilter

	usersInsert = `
	INSERT INTO users
	(
		uuid,
		account_id,
		first_name,
		last_name,
		email,
		auth0_id
	)
	VALUES (
		:uuid,
		:account_id,
		:first_name,
		:last_name,
		:email,
		:auth0_id
	)
	`

	usersUpdate = `
	UPDATE users
	SET
		first_name = :first_name,
		last_name = :last_name,
		updated_at = now()
	WHERE uuid = :uuid
	`

	usersSoftDelete = `
	UPDATE users
	SET
		updated_at = now(),
		deleted_at = now()
	WHERE uuid = $1
	`

	usersHardDelete = `
	DELETE FROM users
	WHERE uuid = $1
	`
)

var (
	// usersSortableColumns defines the columns which are sortable.
	// The first value is used as the default.
	usersSortableColumns = []string{"created_at", "updated_at", "email"}
)

// UsersRepository manages the access to the underlying postgres database.
type UsersRepository struct {
	lists                  map[string]*sqlx.Stmt
	getByID                *sqlx.Stmt
	getByIDWithSoftDeleted *sqlx.Stmt
	getByEmail             *sqlx.Stmt
	getByAuth0ID           *sqlx.Stmt

	create *sqlx.NamedStmt
	update *sqlx.NamedStmt

	softDelete *sqlx.Stmt
	hardDelete *sqlx.Stmt
}

// NewUsersRepository creates a UsersRepository for accessing the
// related tables.
func NewUsersRepository(db *sqlx.DB) (*UsersRepository, error) {
	ctx := context.Background()

	lists := make(map[string]*sqlx.Stmt)
	for _, s := range usersSortableColumns {
		for _, d := range model.AllOrders {
			var err error
			key, query := prepareOrderQuery(usersSelect, s, d)
			lists[key], err = db.PreparexContext(ctx, query)
			if err != nil {
				return nil, err
			}
		}
	}

	getByIDWithSoftDeleted, err := db.PreparexContext(ctx, usersSelectByIDWithSoftDeleted)
	if err != nil {
		return nil, err
	}
	getByID, err := db.PreparexContext(ctx, usersSelectByID)
	if err != nil {
		return nil, err
	}
	getByEmail, err := db.PreparexContext(ctx, usersSelectByEmail)
	if err != nil {
		return nil, err
	}
	getByAuth0ID, err := db.PreparexContext(ctx, usersSelectByAuth0ID)
	if err != nil {
		return nil, err
	}

	create, err := db.PrepareNamedContext(ctx, usersInsert)
	if err != nil {
		return nil, err
	}
	update, err := db.PrepareNamedContext(ctx, usersUpdate)
	if err != nil {
		return nil, err
	}

	softDelete, err := db.PreparexContext(ctx, usersSoftDelete)
	if err != nil {
		return nil, err
	}
	hardDelete, err := db.PreparexContext(ctx, usersHardDelete)
	if err != nil {
		return nil, err
	}

	return &UsersRepository{
		lists:                  lists,
		getByID:                getByID,
		getByIDWithSoftDeleted: getByIDWithSoftDeleted,
		getByEmail:             getByEmail,
		getByAuth0ID:           getByAuth0ID,
		create:                 create,
		update:                 update,
		softDelete:             softDelete,
		hardDelete:             hardDelete,
	}, nil
}

// Close closes all prepared statements.
func (r *UsersRepository) Close() error {
	var final error
	for _, list := range r.lists {
		if err := list.Close(); err != nil {
			final = err
		}
	}
	if err := r.getByID.Close(); err != nil {
		final = err
	}
	if err := r.getByIDWithSoftDeleted.Close(); err != nil {
		final = err
	}
	if err := r.getByEmail.Close(); err != nil {
		final = err
	}
	if err := r.getByAuth0ID.Close(); err != nil {
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

// List returns a set of users
func (r *UsersRepository) List(ctx context.Context, opts *model.SortedListOptions) ([]*model.User, bool, error) {
	opts = opts.Sanitize(usersSortableColumns[0], usersSortableColumns)
	entities := []*model.User{}
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

// GetByID fetches a users by its id.
func (r *UsersRepository) GetByID(ctx context.Context, id string) (*model.User, error) {
	if id == "" {
		return nil, fmt.Errorf("no model.User uuid provided")
	}
	var entity model.User
	err := r.getByID.GetContext(ctx, &entity, id)
	if err != nil {
		return nil, err
	}
	return &entity, nil
}

// GetByIDWithSoftDeleted fetches a accounts by its id also including accounts which got soft deleted.
func (r *UsersRepository) GetByIDWithSoftDeleted(ctx context.Context, id string) (*model.User, error) {
	if id == "" {
		return nil, fmt.Errorf("no model.User uuid provided")
	}
	var entity model.User
	err := r.getByIDWithSoftDeleted.GetContext(ctx, &entity, id)
	if err != nil {
		return nil, err
	}
	return &entity, nil
}

// GetByEmail fetches a users by its email.
func (r *UsersRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	if email == "" {
		return nil, fmt.Errorf("no model.User email provided")
	}
	var entity model.User
	err := r.getByEmail.GetContext(ctx, &entity, email)
	if err != nil {
		return nil, err
	}
	return &entity, nil
}

// GetByAuth0ID fetches a users by its auth0ID.
func (r *UsersRepository) GetByAuth0ID(ctx context.Context, auth0ID string) (*model.User, error) {
	if auth0ID == "" {
		return nil, fmt.Errorf("no model.User auth0ID provided")
	}
	var entity model.User
	err := r.getByAuth0ID.GetContext(ctx, &entity, auth0ID)
	if err != nil {
		return nil, err
	}
	return &entity, nil
}

// Create creates a new users
func (r *UsersRepository) Create(ctx context.Context, e *model.User) (*model.User, error) {

	if e == nil {
		return nil, fmt.Errorf("couldn't create nil object")
	}
	if e.AccountID == nil {
		return nil, fmt.Errorf("no accountID given")
	}
	if e.Auth0ID == nil {
		return nil, fmt.Errorf("no auth0ID given")
	}

	// New ID for the users
	e.UUID = uuid.New().String()
	_, err := NamedStmt(ctx, r.create).ExecContext(ctx, e)
	if err != nil {
		return nil, err
	}

	var entity model.User
	err = r.getByID.GetContext(ctx, &entity, e.UUID)
	if err != nil {
		return nil, err
	}
	return &entity, nil
}

// Update updates the users
func (r *UsersRepository) Update(ctx context.Context, e *model.User) (*model.User, error) {
	if e == nil {
		return nil, fmt.Errorf("couldn't create nil object")
	}
	if e.UUID == "" {
		return nil, fmt.Errorf("users uuid is missing")
	}
	_, err := NamedStmt(ctx, r.update).ExecContext(ctx, e)
	if err != nil {
		return nil, err
	}
	return e, nil
}

// Delete marks a users as deleted. If hard is set, it deletes the actual table row.
func (r *UsersRepository) Delete(ctx context.Context, id string, hard bool) error {
	if id == "" {
		return fmt.Errorf("users id is missing")
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
