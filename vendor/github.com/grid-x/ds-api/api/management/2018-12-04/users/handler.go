package users

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"

	"github.com/grid-x/ds-api/api"
	"github.com/grid-x/ds-api/pkg/encoding"
	"github.com/grid-x/ds-api/pkg/errors"
	"github.com/grid-x/ds-api/pkg/model"
)

const (
	version = "2018-12-04"
	group   = "users"
)

var (
	// defaultTimeout is the default timeout used when calling downstream services
	defaultTimeout = 10 * time.Second
)

type authProvider interface {
	AccountIDFromContext(ctx context.Context) (string, error)
	UserIDFromContext(ctx context.Context) (string, error)
}

type usersRepo interface {
	Create(ctx context.Context, accountID string, user *model.User) (*model.User, error)
	Update(ctx context.Context, user *model.User) (*model.User, error)
	Delete(ctx context.Context, userID string) error
	GetByID(ctx context.Context, userID string) (*model.User, error)
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	List(ctx context.Context, accountID string) ([]*model.User, error)
}

// User resource as exposed by this API version
type User struct {
	Metadata  api.Metadata `json:"metadata,omitempty"`
	FirstName *string      `json:"firstName,omitempty"`
	LastName  *string      `json:"lastName,omitempty"`
	Email     string       `json:"email"`
}

func userFromModel(user *model.User) *User {
	return &User{
		Metadata: api.Metadata{
			ID: user.UUID,
		},
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Email:     user.Email,
	}
}

// Service implements the HTTP handlers for user management
type Service struct {
	logger log.FieldLogger

	usersRepo usersRepo
	auth      authProvider
}

// NewService creates a new service and injects all dependencies
func NewService(injections ...interface{}) *Service {
	s := &Service{}

	for _, inj := range injections {
		switch i := inj.(type) {
		case log.FieldLogger:
			s.logger = i
		case usersRepo:
			s.usersRepo = i
		case authProvider:
			s.auth = i
		}
	}

	if s.logger == nil {
		s.logger = log.New().WithFields(log.Fields{
			"version": version,
			"group":   group,
		})
	}

	return s
}

// GetAuthenticatedResponse represents the response type
type GetAuthenticatedResponse struct {
	*User
}

// WriteText creates a text representation from GetAuthenticatedResponse
func (resp *GetAuthenticatedResponse) WriteText(w io.Writer) error {
	fmt.Fprintf(w, "%#v", resp)
	return nil
}

// WriteJSON creates a json representation from GetAuthenticatedResponse
func (resp *GetAuthenticatedResponse) WriteJSON(w io.Writer) error {
	return json.NewEncoder(w).Encode(resp)
}

// GetAuthenticated implements the HTTP handler to get the authenticated user
//
// @name: GetAuthenticatedUser
// @description: Returns the authenticated user
// @action: users:GetAuthenticated
// @resource: users:*
// @endpoint: GET /user
// @middlewares: auth
func (s *Service) GetAuthenticated(req *http.Request) (*encoding.Response, error) {
	userID, err := s.auth.UserIDFromContext(req.Context())
	if err != nil {
		return nil, errors.E(
			errors.Internal,
			fmt.Errorf("Cannot get userID: %+v", err),
		)
	}

	ctx, cancel := context.WithTimeout(req.Context(), defaultTimeout)
	defer cancel()

	user, err := s.usersRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, errors.E(
			errors.NotExists,
			fmt.Errorf("User with ID %s does not exist", userID),
		)
	}

	return &encoding.Response{
		Payload: &GetAuthenticatedResponse{
			User: userFromModel(user),
		},
	}, nil
}

// ListResponse represents the response type
type ListResponse struct {
	Users []*User `json:"users"`
}

// WriteText creates a text representation from ListResponse
func (resp *ListResponse) WriteText(w io.Writer) error {
	fmt.Fprintf(w, "%#v", resp)
	return nil
}

// WriteJSON creates a json representation from ListResponse
func (resp *ListResponse) WriteJSON(w io.Writer) error {
	return json.NewEncoder(w).Encode(resp)
}

// List implements the HTTP handler to list users
//
// @name: ListUsers
// @description: Returns the list of users in the authenticated account
// @action: users:List
// @resource: users:*
// @endpoint: GET /users
// @middlewares: auth
func (s *Service) List(req *http.Request) (*encoding.Response, error) {
	accountID, err := s.auth.AccountIDFromContext(req.Context())
	if err != nil {
		return nil, errors.E(
			errors.Internal,
			fmt.Errorf("Cannot get accountID: %+v", err),
		)
	}

	ctx, cancel := context.WithTimeout(req.Context(), defaultTimeout)
	defer cancel()

	users, err := s.usersRepo.List(ctx, accountID)
	if err != nil {
		return nil, errors.E(
			errors.Internal,
			fmt.Errorf("Cannot get user list: %+v", err),
		)
	}

	// Sort results to preserve order within responses https://github.com/grid-x/ds-api/issues/113
	sort.Slice(users, func(i, j int) bool {
		return users[i].Email < users[j].Email
	})

	result := make([]*User, len(users))
	for i, u := range users {
		result[i] = userFromModel(u)
	}

	return &encoding.Response{
		Payload: &ListResponse{
			Users: result,
		},
	}, nil
}

// GetResponse represents the response type
type GetResponse struct {
	*User
}

// WriteText creates a text representation from GetResponse
func (resp *GetResponse) WriteText(w io.Writer) error {
	fmt.Fprintf(w, "%#v", resp)
	return nil
}

// WriteJSON creates a json representation from GetResponse
func (resp *GetResponse) WriteJSON(w io.Writer) error {
	return json.NewEncoder(w).Encode(resp)
}

// Get implements the HTTP handler to get a specific user in the account
//
// @name: GetUser
// @description: Returns a specific user in the account
// @action: users:Get
// @resource: users:{userID}
// @endpoint: GET /users/{userID}
// @middlewares: auth
func (s *Service) Get(req *http.Request) (*encoding.Response, error) {
	userID := mux.Vars(req)["userID"]
	if userID == "" {
		return nil, errors.E(
			errors.Validation,
			fmt.Errorf("Missing user ID"),
		)
	}

	ctx, cancel := context.WithTimeout(req.Context(), defaultTimeout)
	defer cancel()

	user, err := s.usersRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, errors.E(
			errors.NotExists,
			fmt.Errorf("User with ID %s does not exist", userID),
		)
	}

	return &encoding.Response{
		Payload: &GetResponse{
			User: userFromModel(user),
		},
	}, nil
}

// CreateRequest represents the request type
type CreateRequest struct {
	FirstName *string `json:"firstName,omitempty"`
	LastName  *string `json:"lastName,omitempty"`
	Email     string  `json:"email"`
}

// Validate validates a CreateRequest
func (req *CreateRequest) Validate() error {
	if req.Email == "" {
		return errors.E(
			errors.Validation,
			fmt.Errorf("Missing email"),
		)
	}

	return nil
}

// ReadJSON reads CreateRequest from a JSON payload
func (req *CreateRequest) ReadJSON(r io.Reader) error {
	return json.NewDecoder(r).Decode(req)
}

// CreateResponse represents the response type
type CreateResponse struct {
	*User
}

// WriteText creates a text representation from CreateResponse
func (resp *CreateResponse) WriteText(w io.Writer) error {
	fmt.Fprintf(w, "%#v", resp)
	return nil
}

// WriteJSON creates a json representation from CreateResponse
func (resp *CreateResponse) WriteJSON(w io.Writer) error {
	return json.NewEncoder(w).Encode(resp)
}

// Create implements the HTTP handler to create users
//
// @name: CreateUser
// @description: Creates a new user
// @action: users:Create
// @resource: users:*
// @endpoint: POST /users
// @middlewares: auth
func (s *Service) Create(req *http.Request, payload CreateRequest) (*encoding.Response, error) {
	accountID, err := s.auth.AccountIDFromContext(req.Context())
	if err != nil {
		return nil, errors.E(
			errors.Internal,
			fmt.Errorf("Cannot get accountID: %+v", err),
		)
	}

	email := payload.Email

	ctx, cancel := context.WithTimeout(req.Context(), defaultTimeout)
	defer cancel()

	_, err = s.usersRepo.GetByEmail(ctx, email)
	if err == nil {
		return nil, errors.E(
			errors.NotExists,
			fmt.Errorf("User with email %s already exists", email),
		)
	}

	ctx, cancel = context.WithTimeout(req.Context(), defaultTimeout)
	defer cancel()

	user, err := s.usersRepo.Create(ctx, accountID, &model.User{
		UUID:      uuid.New().String(),
		FirstName: payload.FirstName,
		LastName:  payload.LastName,
		Email:     email,
	})
	if err != nil {
		return nil, errors.E(
			errors.Internal,
			fmt.Errorf("Cannot create user %s: %+v", email, err),
		)
	}

	return &encoding.Response{
		Status:  201,
		Payload: &CreateResponse{User: userFromModel(user)},
	}, nil
}

// UpdateRequest represents the request type
type UpdateRequest struct {
	FirstName *string `json:"firstName,omitempty"`
	LastName  *string `json:"lastName,omitempty"`
}

// Validate validates an UpdateRequest
func (req *UpdateRequest) Validate() error {
	if req.FirstName == nil && req.LastName == nil {
		return errors.E(
			errors.Validation,
			fmt.Errorf("Nothing to update"),
		)
	}

	if req.FirstName == nil && *req.LastName == "" {
		return errors.E(
			errors.Validation,
			fmt.Errorf("Nothing to update"),
		)
	}

	if req.LastName == nil && *req.FirstName == "" {
		return errors.E(
			errors.Validation,
			fmt.Errorf("Nothing to update"),
		)
	}

	if *req.LastName == "" && *req.FirstName == "" {
		return errors.E(
			errors.Validation,
			fmt.Errorf("Nothing to update"),
		)
	}

	return nil
}

// ReadJSON reads UpdateRequest from a JSON payload
func (req *UpdateRequest) ReadJSON(r io.Reader) error {
	return json.NewDecoder(r).Decode(req)
}

// UpdateResponse represents the response type
type UpdateResponse struct {
	*User
}

// WriteText creates a text representation from UpdateResponse
func (resp *UpdateResponse) WriteText(w io.Writer) error {
	fmt.Fprintf(w, "%#v", resp)
	return nil
}

// WriteJSON creates a json representation from UpdateResponse
func (resp *UpdateResponse) WriteJSON(w io.Writer) error {
	return json.NewEncoder(w).Encode(resp)
}

// Update implements the HTTP handler to update a user
//
// @name: UpdateUser
// @description: Updates a user
// @action: users:Update
// @resource: users:{userID}
// @endpoint: PATCH /users/{userID}
// @middlewares: auth
func (s *Service) Update(req *http.Request, payload UpdateRequest) (*encoding.Response, error) {
	userID := mux.Vars(req)["userID"]
	if userID == "" {
		return nil, errors.E(
			errors.Validation,
			fmt.Errorf("Missing user ID"),
		)
	}

	ctx, cancel := context.WithTimeout(req.Context(), defaultTimeout)
	defer cancel()

	user, err := s.usersRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, errors.E(
			errors.NotExists,
			fmt.Errorf("User with ID %s does not exist", userID),
		)
	}
	user.FirstName = payload.FirstName
	user.LastName = payload.LastName

	ctx, cancel = context.WithTimeout(req.Context(), defaultTimeout)
	defer cancel()

	user, err = s.usersRepo.Update(ctx, user)
	if err != nil {
		return nil, errors.E(
			errors.Internal,
			fmt.Errorf("Cannot patch user %s: %+v", user.UUID, err),
		)
	}

	return &encoding.Response{
		Payload: &UpdateResponse{
			User: userFromModel(user),
		},
	}, nil
}

// DeleteResponse represents the response type
type DeleteResponse struct{}

// WriteText creates a text representation from DeleteResponse
func (resp *DeleteResponse) WriteText(w io.Writer) error {
	fmt.Fprintf(w, "%#v", resp)
	return nil
}

// WriteJSON creates a json representation from DeleteResponse
func (resp *DeleteResponse) WriteJSON(w io.Writer) error {
	return json.NewEncoder(w).Encode(resp)
}

// Delete implements the HTTP handler for deleting users
//
// @name: DeleteUser
// @description: Deletes the specified user
// @action: users:Delete
// @resource: users:{userID}
// @endpoint: DELETE /users/{userID}
// @middlewares: auth
func (s *Service) Delete(req *http.Request) (*encoding.Response, error) {
	userID := mux.Vars(req)["userID"]
	if userID == "" {
		return nil, errors.E(
			errors.Validation,
			fmt.Errorf("Missing user ID"),
		)
	}

	ctx, cancel := context.WithTimeout(req.Context(), defaultTimeout)
	defer cancel()

	_, err := s.usersRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, errors.E(
			errors.NotExists,
			fmt.Errorf("User with ID %s does not exist", userID),
		)
	}

	ctx, cancel = context.WithTimeout(req.Context(), defaultTimeout)
	defer cancel()

	if err := s.usersRepo.Delete(ctx, userID); err != nil {
		return nil, errors.E(
			errors.Internal,
			fmt.Errorf("Cannot delete user %s: %+v", userID, err),
		)
	}

	return &encoding.Response{
		Payload: &DeleteResponse{},
	}, nil
}
