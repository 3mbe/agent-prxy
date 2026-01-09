package httpapi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandleHealthz(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		method       string
		wantStatus   int
		wantAllow    string
		wantCTSubstr string

		wantBody     string
		checkBody    bool
		wantNonEmpty bool
	}{
		{
			name:         "GET returns 200 with ok body",
			method:       http.MethodGet,
			wantStatus:   http.StatusOK,
			wantCTSubstr: "text/plain",
			wantBody:     "ok\n",
			checkBody:    true,
		},
		{
			name:         "HEAD returns 200 with no body",
			method:       http.MethodHead,
			wantStatus:   http.StatusOK,
			wantCTSubstr: "text/plain",
			wantBody:     "",
			checkBody:    true,
		},
		{
			name:         "POST returns 405 and Allow header",
			method:       http.MethodPost,
			wantStatus:   http.StatusMethodNotAllowed,
			wantAllow:    "GET, HEAD",
			wantCTSubstr: "text/plain",
			wantNonEmpty: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(tt.method, "/healthz", nil)
			rr := httptest.NewRecorder()

			handleHealthz(rr, req)

			if rr.Code != tt.wantStatus {
				t.Fatalf("status=%d want=%d", rr.Code, tt.wantStatus)
			}

			ct := rr.Header().Get("Content-Type")
			if tt.wantCTSubstr != "" && !strings.Contains(ct, tt.wantCTSubstr) {
				t.Fatalf("Content-Type=%q want contains %q", ct, tt.wantCTSubstr)
			}

			if tt.wantAllow == "" {
				if got := rr.Header().Get("Allow"); got != "" {
					t.Fatalf("Allow=%q want empty", got)
				}
			} else {
				if got := rr.Header().Get("Allow"); got != tt.wantAllow {
					t.Fatalf("Allow=%q want=%q", got, tt.wantAllow)
				}
			}

			body := rr.Body.String()
			if tt.checkBody {
				if body != tt.wantBody {
					t.Fatalf("body=%q want=%q", body, tt.wantBody)
				}
			} else if tt.wantNonEmpty && body == "" {
				t.Fatalf("expected non-empty body")
			}
		})
	}
}
