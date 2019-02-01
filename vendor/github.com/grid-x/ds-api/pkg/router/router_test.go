package router

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func mkRequest(headers map[string]string) *http.Request {
	req, _ := http.NewRequest(http.MethodGet, "http://example.com", nil)
	for k, v := range headers {
		req.Header.Add(k, v)
	}
	return req
}

func Test_extractVersion(t *testing.T) {
	testcases := []struct {
		req     *http.Request
		want    string
		wantErr bool
	}{
		{
			req:     &http.Request{},
			wantErr: true,
		},
		{
			req: mkRequest(map[string]string{
				"Accept": "application/vnd.gridx.ai.2018-2-07+json",
			}),
			wantErr: true,
		},
		{
			req: mkRequest(map[string]string{
				"Accept": "application/vnd.gridx.2018-12-07+json",
			}),
			wantErr: true,
		},
		{
			req: mkRequest(map[string]string{
				"Accept": "application/vnd.gridx.ai.2018-12-7+json",
			}),
			wantErr: true,
		},
		{
			req: mkRequest(map[string]string{
				"Accept": "application/vnd.gridx.ai.2018-12-41+json",
			}),
			wantErr: true,
		},
		{
			req: mkRequest(map[string]string{
				"Accept": "application/vnd.gridx.ai.2118-12-01+json",
			}),
			wantErr: true,
		},
		{
			req: mkRequest(map[string]string{
				"Accept": "application/vnd.gridx.ai.2018-12-07+json",
			}),
			want:    "2018-12-07",
			wantErr: false,
		},
		{
			req: mkRequest(map[string]string{
				"Accept": "application/vnd.gridx.ai.2099-12-31+json",
			}),
			want:    "2099-12-31",
			wantErr: false,
		},
		{
			req: mkRequest(map[string]string{
				"Accept": "application/vnd.gridx.ai.2001-01-01+json",
			}),
			want:    "2001-01-01",
			wantErr: false,
		},
	}

	for i, tc := range testcases {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			got, err := extractVersion(tc.req)
			if tc.wantErr {
				if err != nil {
					return
				}
				t.Fatalf("expected error but got nil")
			} else if !tc.wantErr && err != nil {
				t.Fatalf("unexpected error: %+v", err)
			}

			if !cmp.Equal(tc.want, got) {
				t.Errorf("unexpected result: %s", cmp.Diff(tc.want, got))
			}
		})
	}
}
