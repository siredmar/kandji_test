package device

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

const (
	PrivKeyBlockType = "RSA PRIVATE KEY"
)

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

func Test_DeviceAuth_Verify(t *testing.T) {

	testcases := []struct {
		privKey *rsa.PrivateKey
		pubKey  string
		content string
		want    error
	}{
		{
			privKey: loadPrivKey("testdata/private.pem", t),
			pubKey:  loadPubKeyStr("testdata/public.pem", t),
			content: "foo",
			want:    nil,
		},
		{
			privKey: loadPrivKey("testdata/private.pem", t),
			pubKey:  "",
			content: "foo",
			want:    fmt.Errorf("verification failed"),
		},
	}

	for i, tc := range testcases {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			signature := sign([]byte(tc.content), tc.privKey, t)
			err := Verify(string(signature), tc.pubKey, []byte(tc.content))
			if err == nil && tc.want != nil {
				t.Fatalf("expected error: %+v", tc.want)
			} else if err != nil && tc.want == nil {
				t.Fatalf("got unexpected error: %+v", err)
			} else if err != nil && tc.want != nil && tc.want.Error() != err.Error() {
				t.Fatalf("unexpected error! want: %+v, but got: %+v", tc.want, err)
			}
		})
	}
}
func Test_DeviceAuth_GenerateToken(t *testing.T) {

	testPrivKey := "testdata/private.pem"

	raw, err := ioutil.ReadFile(testPrivKey)
	if err != nil {
		t.Fatalf("can't read test private key: %+v", err)
	}
	block, _ := pem.Decode([]byte(raw))
	if block == nil {
		t.Fatalf("error decoding private key")
	}
	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		t.Fatalf("error parsing private key: %+v", err)
	}

	testcases := []struct {
		jwtIssuer, deviceID string
		jwtIssuedAt         time.Time
		jwtNotBefore        time.Time
		jwtExpTime          time.Time
		jwtPrivKey          *rsa.PrivateKey

		want    string
		wantErr error
	}{
		{
			jwtIssuer:    "device-api.gridx.de",
			deviceID:     "83d8769dc416a9203c2a86c72c325ac030a99bee776184d29e289e1b6fc4a22e",
			jwtIssuedAt:  time.Date(2018, time.January, 1, 0, 30, 17, 0, time.UTC),
			jwtNotBefore: time.Date(2018, time.January, 1, 1, 30, 37, 0, time.UTC),
			jwtExpTime:   time.Date(2018, time.January, 1, 21, 30, 37, 0, time.UTC),
			jwtPrivKey:   key,

			want:    "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE1MTQ4NDIyMzcsImlhdCI6MTUxNDc2NjYxNywiaXNzIjoiZGV2aWNlLWFwaS5ncmlkeC5kZSIsIm5iZiI6MTUxNDc3MDIzNywic3ViIjoiODNkODc2OWRjNDE2YTkyMDNjMmE4NmM3MmMzMjVhYzAzMGE5OWJlZTc3NjE4NGQyOWUyODllMWI2ZmM0YTIyZSJ9.ONtXYcQNakYJjLPyIhTEEO_xbxduTRMLgkXASmXapzXDlPUGIGXxbTJpGDDn5yS-2aWbUCFOq2AHEuusjtm3bJ9TwRu1gaW6-Hs5Y3QoeTTUYP24Qwi3UMYzvJ57ITDtc-NVVGMK39CW3R4iDkePSeqZpdBXvoddgE0sDOgeRiuFlgujFrL7JZgQTDBQVlxnXD05d4fKLCT8UJd5ZvyN5c275TxZyfsNaGxJeZA-oyEvo2L3HQoR1IXSZNMJMx32dJAUQXMhAM7f9IDSMh0fw8GiGZHES8OZfRWP2k2tgr8zR4K2uso8PFuyodxIOfwmGStaAlPGpHvnbhrzFUmgdg",
			wantErr: nil,
		},
	}

	for i, tc := range testcases {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			gen := NewJWTGenerator(tc.jwtIssuer, tc.jwtPrivKey)
			got, err := gen.genToken(tc.jwtIssuedAt, tc.jwtNotBefore, tc.jwtExpTime, tc.deviceID)
			if !cmp.Equal(tc.wantErr, err) {
				t.Fatalf("unexpected error: %s", cmp.Diff(tc.wantErr, err))
			}
			if !cmp.Equal(tc.want, got) {
				t.Errorf("unexpected token: %s", cmp.Diff(tc.want, got))
			}
		})
	}
}
