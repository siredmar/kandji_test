package device

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	corev1beta1 "github.com/grid-x/ds-k8s/pkg/apis/core/v1beta1"
	log "github.com/sirupsen/logrus"

	"github.com/grid-x/ds-api/pkg/model"
	"github.com/grid-x/ds-api/pkg/router"
)

var (
	defaultTimeout = 10 * time.Second
)

type key int

const (
	deviceKey key = 0
)

type device struct {
	deviceID, accountID *string
}

// DeviceRepository is the minimal interface needed by the AuthProvider
type DeviceRepository interface {
	Get(ctx context.Context, namespace, name string) (*corev1beta1.Device, error)
}

// AuthProvider implements authentication functionality for the device-api
type AuthProvider struct {
	logger  log.FieldLogger
	jwt     *JWTGenerator
	devRepo DeviceRepository
}

// NewAuthProvider creates a new auth provider with the given settings
func NewAuthProvider(logger log.FieldLogger, jwt *JWTGenerator, repo DeviceRepository) *AuthProvider {
	return &AuthProvider{
		logger:  logger,
		jwt:     jwt,
		devRepo: repo,
	}
}

// AccountIDFromContext extracts the accountID from the context
func (ap *AuthProvider) AccountIDFromContext(ctx context.Context) (string, error) {
	dev, ok := ap.fromContext(ctx)
	if !ok || dev.accountID == nil {
		return "", fmt.Errorf("accountID not present")
	}

	return *dev.accountID, nil
}

// DeviceIDFromContext extracts the deviceID from the context
func (ap *AuthProvider) DeviceIDFromContext(ctx context.Context) (string, error) {
	dev, ok := ap.fromContext(ctx)
	if !ok || dev.deviceID == nil {
		return "", fmt.Errorf("deviceID not present")
	}

	return *dev.deviceID, nil
}

func (ap *AuthProvider) fromContext(ctx context.Context) (*device, bool) {
	device, ok := ctx.Value(deviceKey).(*device)
	return device, ok
}

func (ap *AuthProvider) newContext(ctx context.Context, dev *device) context.Context {
	return context.WithValue(ctx, deviceKey, dev)
}

// Middleware implements a http middleware for authenticating devices
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

			claims, err := ap.jwt.Valid(tokenStr)
			if err != nil {
				ap.logger.WithFields(fields).Errorf("validating token: %+v", err)
				http.Error(w, `{}`, http.StatusUnauthorized)
				return
			}

			deviceID, ok := claims["sub"].(string)
			if !ok {
				ap.logger.WithFields(fields).Error("cannot extract deviceID claim")
				http.Error(w, `{}`, http.StatusInternalServerError)
				return
			}
			fields["deviceID"] = deviceID

			accountID, ok := claims["accountID"].(string)
			if !ok {
				ap.logger.WithFields(fields).Error("cannot extract deviceID claim")
				http.Error(w, `{}`, http.StatusInternalServerError)
				return
			}
			fields["accountID"] = accountID

			ctx, cancel := context.WithTimeout(r.Context(), defaultTimeout)
			defer cancel()

			_, err = ap.devRepo.Get(ctx, model.AccountNamespaceName(accountID), deviceID)
			if err != nil {
				ap.logger.WithFields(fields).Errorf("cannot get device: %+v", err)
				http.Error(w, `{}`, http.StatusInternalServerError)
				return
			}

			r = r.WithContext(ap.newContext(r.Context(), &device{
				deviceID:  &deviceID,
				accountID: &accountID,
			}))
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
