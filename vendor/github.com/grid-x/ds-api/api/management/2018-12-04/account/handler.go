package account

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/grid-x/ds-api/api"
	"github.com/grid-x/ds-api/pkg/encoding"
	"github.com/grid-x/ds-api/pkg/errors"
	"github.com/grid-x/ds-api/pkg/model"
)

const (
	version = "2018-12-04"
	group   = "account"
)

var (
	// defaultTimeout is the default timeout used when calling downstream services
	defaultTimeout = 10 * time.Second
)

type authProvider interface {
	AccountIDFromContext(ctx context.Context) (string, error)
}

type accountRepo interface {
	Get(ctx context.Context, accountID string) (*model.Account, error)
	Update(ctx context.Context, account *model.Account) (*model.Account, error)
}

// Account as exposed by this API version
type Account struct {
	Metadata api.Metadata `json:"metadata,omitempty"`
	Name     string       `json:"name"`
}

func fromModel(acc *model.Account) *Account {
	return &Account{
		Metadata: api.Metadata{
			ID: acc.UUID,
		},
		Name: acc.Name,
	}
}

// Service implements HTTP handlers for account management
type Service struct {
	logger log.FieldLogger

	accRepo accountRepo
	auth    authProvider
}

// NewService creates a new service and injects all dependencies
func NewService(injections ...interface{}) *Service {
	s := &Service{}

	for _, inj := range injections {
		switch i := inj.(type) {
		case log.FieldLogger:
			s.logger = i
		case accountRepo:
			s.accRepo = i
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
	*Account
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

// GetAuthenticated returns the authenticated account
//
// @name: GetAuthenticatedAccount
// @description: Returns the authenticated account
// @action: account:GetAuthenticated
// @resource: account:
// @endpoint: GET /account
// @middlewares: auth
func (s *Service) GetAuthenticated(req *http.Request) (*encoding.Response, error) {
	accountID, err := s.auth.AccountIDFromContext(req.Context())
	if err != nil {
		return nil, errors.E(
			errors.Internal,
			fmt.Errorf("Cannot get accountID: %+v", err),
		)
	}

	ctx, cancel := context.WithTimeout(req.Context(), defaultTimeout)
	defer cancel()

	acc, err := s.accRepo.Get(ctx, accountID)
	if err != nil {
		return nil, errors.E(
			errors.NotExists,
			fmt.Errorf("Account with ID %s does not exist", accountID),
		)
	}

	return &encoding.Response{
		Payload: &GetAuthenticatedResponse{
			Account: fromModel(acc),
		},
	}, nil
}

// UpdateAuthenticatedRequest represents the request type
type UpdateAuthenticatedRequest struct {
	Name string `json:"name"`
}

// Validate validates an UpdateAuthenticatedRequest
func (req *UpdateAuthenticatedRequest) Validate() error {
	if req.Name == "" {
		return errors.E(
			errors.Validation,
			fmt.Errorf("Missing account name"),
		)
	}

	return nil
}

// ReadJSON reads UpdateAuthenticatedRequest from a JSON payload
func (req *UpdateAuthenticatedRequest) ReadJSON(r io.Reader) error {
	return json.NewDecoder(r).Decode(req)
}

// UpdateAuthenticatedResponse represents the response type
type UpdateAuthenticatedResponse struct {
	*Account
}

// WriteText creates a text representation from UpdateAuthenticatedResponse
func (resp *UpdateAuthenticatedResponse) WriteText(w io.Writer) error {
	fmt.Fprintf(w, "%#v", resp)
	return nil
}

// WriteJSON creates a json representation from UpdateAuthenticatedResponse
func (resp *UpdateAuthenticatedResponse) WriteJSON(w io.Writer) error {
	return json.NewEncoder(w).Encode(resp)
}

// UpdateAuthenticated implements the HTTP handler to update the authenticated
// account
//
// @name: UpdateAuthenticatedAccount
// @description: Updates the authenticated account
// @action: account:UpdateAuthenticated
// @resource: account:*
// @endpoint: PATCH /account
// @middlewares: auth
func (s *Service) UpdateAuthenticated(req *http.Request, payload UpdateAuthenticatedRequest) (*encoding.Response, error) {
	accountID, err := s.auth.AccountIDFromContext(req.Context())
	if err != nil {
		return nil, errors.E(
			errors.Internal,
			fmt.Errorf("Cannot get accountID: %+v", err),
		)
	}

	ctx, cancel := context.WithTimeout(req.Context(), defaultTimeout)
	defer cancel()

	acc, err := s.accRepo.Get(ctx, accountID)
	if err != nil {
		return nil, errors.E(
			errors.Exist,
			fmt.Errorf("Account with ID %s does not exist", accountID),
		)
	}
	acc.Name = payload.Name

	acc, err = s.accRepo.Update(ctx, acc)
	if err != nil {
		return nil, errors.E(
			errors.Internal,
			fmt.Errorf("Cannot patch account %s: %+v", accountID, err),
		)
	}

	return &encoding.Response{
		Payload: &UpdateAuthenticatedResponse{
			Account: fromModel(acc),
		},
	}, nil
}
