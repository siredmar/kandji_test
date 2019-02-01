package auth

import (
	"bytes"
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	corev1beta1 "github.com/grid-x/ds-k8s/pkg/apis/core/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/grid-x/ds-api/pkg/auth/device"
)

const (
	PrivKeyBlockType = "RSA PRIVATE KEY"
)

type getByPubKeyF func(ctx context.Context, pubKey string) (*corev1beta1.Device, error)

type mockDeviceRepository struct {
	get getByPubKeyF
}

func (m *mockDeviceRepository) GetByPublicKey(ctx context.Context, publicKey string) (*corev1beta1.Device, error) {
	return m.get(ctx, publicKey)
}

func newReq(signature string) *http.Request {
	req, _ := http.NewRequest(http.MethodPost, "test.com", nil)
	req.Header.Add("X-GRIDX", signature)
	req.Header.Add("content-type", "application/json;")
	return req
}

func sign(data []byte, privkey *rsa.PrivateKey, t *testing.T) []byte {
	hash := sha256.New()
	if _, err := bytes.NewReader(data).WriteTo(hash); err != nil {
		t.Fatal(err)
	}

	sig, err := rsa.SignPKCS1v15(rand.Reader, privkey, crypto.SHA256, hash.Sum(nil))
	if err != nil {
		t.Fatal(err)
	}

	b64 := make([]byte, base64.StdEncoding.EncodedLen(len(sig)))
	base64.StdEncoding.Encode(b64, sig)

	return b64
}

func loadPrivKey(path string, t *testing.T) *rsa.PrivateKey {
	pemData, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	block, _ := pem.Decode(pemData)

	if block == nil ||
		block.Type != PrivKeyBlockType {
		t.Fatal(err)
	}

	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		t.Fatal(err)
	}

	return key
}

func loadPubKeyStr(path string, t *testing.T) string {
	pemData, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	return string(pemData)
}

func Test_GetToken(t *testing.T) {
	publicKey := loadPubKeyStr("testdata/public_device.pem", t)
	privKey := loadPrivKey("testdata/private_device.pem", t)

	privateJWTKey := loadPrivKey("testdata/private.pem", t)

	jwtGen := device.NewJWTGenerator("foobar", privateJWTKey)

	testcases := []struct {
		injections []interface{}
		input      GetTokenRequest

		wantErr         bool
		wantTokenLength int
	}{
		{
			injections: []interface{}{
				&mockDeviceRepository{
					get: func(ctx context.Context, pubKey string) (*corev1beta1.Device, error) {
						return nil, fmt.Errorf("not implemented")
					},
				},
				jwtGen,
			},
			input:           GetTokenRequest{},
			wantErr:         true,
			wantTokenLength: 0,
		},
		{
			injections: []interface{}{
				&mockDeviceRepository{
					get: func(ctx context.Context, pubKey string) (*corev1beta1.Device, error) {
						return &corev1beta1.Device{
							ObjectMeta: metav1.ObjectMeta{
								Namespace: "default",
								Name:      "foobar",
							},
						}, nil
					},
				},
				jwtGen,
			},
			input: GetTokenRequest{
				PublicKey: publicKey,
				IDData:    "{}",
			},
			wantErr:         false,
			wantTokenLength: 490,
		},
	}

	for i, tc := range testcases {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			data, err := json.Marshal(tc.input)
			if err != nil {
				t.Fatal(err)
			}
			signature := sign(data, privKey, t)
			req := newReq(string(signature))

			svc := NewService(tc.injections...)
			got, err := svc.GetToken(req, tc.input)

			if !tc.wantErr && err != nil {
				t.Fatalf("unexpected error: %+v", err)
			} else if tc.wantErr && err == nil {
				t.Fatal("expected error but did not get one")
			}

			token := ""
			if got != nil {
				if got.Payload != nil {
					if resp := got.Payload.(*GetTokenResponse); resp != nil {
						token = resp.Token
					}
				}
			}

			if tc.wantTokenLength != len(token) {
				t.Errorf("unexpected token length: want: %d, but got: %d", tc.wantTokenLength, len(token))
			}
		})
	}
}
