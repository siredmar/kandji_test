package management

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	log "github.com/sirupsen/logrus"
	"k8s.io/api/core/v1"

	"github.com/grid-x/ds-api/pkg/model"
	"github.com/grid-x/ds-api/pkg/router"
)

var (
	defaultTimeout = 10 * time.Second
)

type key int

const (
	userIDKey     = "sub"
	emailKey      = "email"
	userKey   key = 0
)

// UserStore is the interface for an UserStore required by the auth provider
type UserStore interface {
	GetByAuth0ID(context.Context, string) (*model.User, error)
	Create(context.Context, *model.User) (*model.User, error)
}

// AccountStore is the interface for an AccountsStore required by the auth provider
type AccountStore interface {
	Create(context.Context, *model.Account) (*model.Account, error)
}

// NamespaceClient is the interface for creation of namespaces
type NamespaceClient interface {
	Create(*v1.Namespace) (*v1.Namespace, error)
}

// AuthProvider implements the default authentication chain
type AuthProvider struct {
	logger        log.FieldLogger
	rsaSigningKey []byte

	users      UserStore
	accounts   AccountStore
	namespaces NamespaceClient
}

// NewAuthProvider creates a new AuthProvider with the given settings
func NewAuthProvider(logger log.FieldLogger, rsaSigningKey []byte, users UserStore, accounts AccountStore, namespaces NamespaceClient) *AuthProvider {
	return &AuthProvider{
		logger:        logger,
		rsaSigningKey: rsaSigningKey,
		users:         users,
		accounts:      accounts,
		namespaces:    namespaces,
	}
}

// AccountIDFromContext extracts accountID from the given context.
func (ap *AuthProvider) AccountIDFromContext(ctx context.Context) (string, error) {
	user, ok := ap.FromContext(ctx)
	if !ok || user == nil || user.AccountID == nil {
		return "", fmt.Errorf("user's context is not present")
	}

	return *user.AccountID, nil
}

// UserIDFromContext extracts userID from the given context.
func (ap *AuthProvider) UserIDFromContext(ctx context.Context) (string, error) {
	user, ok := ap.FromContext(ctx)
	if !ok {
		return "",
			fmt.Errorf("user's context is not present")
	}

	return user.UUID, nil
}

// FromContext extracts a user from an http.Request.
func (ap *AuthProvider) FromContext(ctx context.Context) (*model.User, bool) {
	user, ok := ctx.Value(userKey).(*model.User)
	return user, ok
}

// NewContext returns a new Context that carries a provided user value:
func (ap *AuthProvider) NewContext(ctx context.Context, user *model.User) context.Context {
	return context.WithValue(ctx, userKey, user)
}

// Middleware returns a new middleware that performs a per-request an
// authorization check.
func (ap *AuthProvider) Middleware() router.Middleware {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == "OPTIONS" {
				// Forward for CORS
				h.ServeHTTP(w, r)
				return
			}

			fields := log.Fields{
				"authorization": r.Header.Get("Authorization"),
			}
			tokenStr, err := extractToken(r.Header.Get("Authorization"))
			if err != nil {
				ap.logger.WithFields(fields).Errorf("extractToken: %+v", err)
				http.Error(w, `{}`, http.StatusUnauthorized)
				return
			}
			fields["extractToken"] = tokenStr
			token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
				if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
					fields["method"] = t.Method.Alg()
					ap.logger.WithFields(fields).Errorf("invalid signing method")
					return nil, fmt.Errorf("invalid signing method: %s", t.Method.Alg())
				}
				return jwt.ParseRSAPublicKeyFromPEM(ap.rsaSigningKey)
			})
			if err != nil || !token.Valid {
				if token != nil {
					fields["valid"] = token.Valid
				}
				ap.logger.WithFields(fields).Errorf("Invalid token: %+v", err)
				http.Error(w, `{}`, http.StatusUnauthorized)
				return
			}
			fields["token"] = token.Raw

			claims, ok := extractClaims(token)
			if !ok {
				ap.logger.WithFields(fields).Errorf("extractAuth0AndEmail: %+v", err)
				http.Error(w, `{}`, http.StatusUnauthorized)
				return
			}
			fields["auth0"] = claims.Auth0
			fields["email"] = claims.Email

			ctx, cancel := context.WithTimeout(r.Context(), defaultTimeout)
			defer cancel()

			user, err := ap.users.GetByAuth0ID(ctx, claims.Auth0)
			if err == sql.ErrNoRows {
				ap.logger.WithFields(fields).Infof("creating new account")
				// Create account
				ctx, cancel = context.WithTimeout(r.Context(), defaultTimeout)
				defer cancel()
				account, err := ap.accounts.Create(ctx, &model.Account{
					Name: claims.Email,
				})
				if err != nil {
					ap.logger.WithFields(fields).Errorf("createAccount: %+v", err)
					http.Error(w, `{}`, http.StatusUnauthorized)
					return
				}
				fields["accountID"] = account.UUID
				ap.logger.WithFields(fields).Infof("new account created")

				// create namespace in k8s
				ctx, cancel = context.WithTimeout(r.Context(), defaultTimeout)
				defer cancel()
				ap.logger.WithFields(fields).Info("creating new namespace...")
				n, err := ap.namespaces.Create(model.AccountNamespace(account))
				if err != nil {
					ap.logger.WithFields(fields).Errorf("createNamespace: %+v", err)
					http.Error(w, `{}`, http.StatusUnauthorized)
					return
				}
				fields["namespace"] = n.Name
				ap.logger.WithFields(fields).Info("namespace created")

				ap.logger.WithFields(fields).Infof("creating new user...")
				user = &model.User{
					Auth0ID:   &claims.Auth0,
					Email:     claims.Email,
					AccountID: &account.UUID,
				}
				ctx, cancel = context.WithTimeout(r.Context(), defaultTimeout)
				defer cancel()
				user, err = ap.users.Create(ctx, user)
				if err != nil {
					ap.logger.WithFields(fields).Errorf("createUser: %+v", err)
					http.Error(w, `{}`, http.StatusUnauthorized)
					return
				}
				fields["userID"] = user.UUID
				ap.logger.WithFields(fields).Infof("new user created")
			} else if err != nil {
				ap.logger.WithFields(fields).Errorf("getUser: %+v", err)
				http.Error(w, `{}`, http.StatusUnauthorized)
				return
			}
			r = r.WithContext(ap.NewContext(r.Context(), user))
			h.ServeHTTP(w, r)
		})

	}
}

func extractToken(t string) (string, error) {
	if t == "" {
		return "", fmt.Errorf("token is empty")
	}
	token := strings.Fields(t)
	if len(token) != 2 {
		return "", fmt.Errorf("token format is invalid; Authorization: TOKEN_TYPE VALUE")
	}
	if !strings.EqualFold(token[0], "Bearer") {
		return "", fmt.Errorf("token type is invalid; expect Bearer")
	}
	return token[1], nil
}

type token struct {
	Auth0 string
	Email string
}

func extractClaims(t *jwt.Token) (*token, bool) {
	claims := t.Claims.(jwt.MapClaims)
	id, ok := claims[userIDKey].(string)
	if !ok {
		// Log error
		return nil, false
	}
	email, ok := claims[emailKey].(string)
	if !ok {
		// Log err
		return nil, false
	}

	return &token{
		Auth0: id,
		Email: email,
	}, true
}
