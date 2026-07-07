package security

import (
	"io"
	"net/http"
	"strings"
	"testing"
)

type mockTransport struct {
	roundTripFunc func(req *http.Request) (*http.Response, error)
}

func (m *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return m.roundTripFunc(req)
}

func TestVerifyTurnstile(t *testing.T) {
	origTransport := http.DefaultTransport
	defer func() {
		http.DefaultTransport = origTransport
	}()

	tests := []struct {
		name       string
		token      string
		mockResp   string
		mockStatus int
		mockErr    error
		want       bool
		wantErr    bool
	}{
		{
			name:       "Success",
			token:      "valid-token",
			mockResp:   `{"success": true}`,
			mockStatus: http.StatusOK,
			want:       true,
			wantErr:    false,
		},
		{
			name:       "Failed Verification",
			token:      "invalid-token",
			mockResp:   `{"success": false, "error-codes": ["invalid-input-response"]}`,
			mockStatus: http.StatusOK,
			want:       false,
			wantErr:    false,
		},
		{
			name:       "HTTP Error Status",
			token:      "some-token",
			mockResp:   `Internal Server Error`,
			mockStatus: http.StatusInternalServerError,
			want:       false,
			wantErr:    true,
		},
		{
			name:    "Empty Token",
			token:   "",
			want:    false,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name != "Empty Token" {
				http.DefaultTransport = &mockTransport{
					roundTripFunc: func(req *http.Request) (*http.Response, error) {
						if tt.mockErr != nil {
							return nil, tt.mockErr
						}
						return &http.Response{
							StatusCode: tt.mockStatus,
							Body:       io.NopCloser(strings.NewReader(tt.mockResp)),
							Header:     make(http.Header),
						}, nil
					},
				}
			}

			got, err := VerifyTurnstile("secret", tt.token, "127.0.0.1")
			if (err != nil) != tt.wantErr {
				t.Fatalf("VerifyTurnstile() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("VerifyTurnstile() = %v, want %v", got, tt.want)
			}
		})
	}
}
func TestVerifyTurnstileRequestPayload(t *testing.T) {
	origTransport := http.DefaultTransport
	defer func() {
		http.DefaultTransport = origTransport
	}()

	var capturedBody string
	http.DefaultTransport = &mockTransport{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			if err := req.ParseForm(); err != nil {
				return nil, err
			}
			capturedBody = req.PostForm.Encode()
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(`{"success": true}`)),
				Header:     make(http.Header),
			}, nil
		},
	}

	_, err := VerifyTurnstile("my-secret", "my-token", "192.168.1.1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedParts := []string{
		"secret=my-secret",
		"response=my-token",
		"remoteip=192.168.1.1",
	}

	for _, p := range expectedParts {
		if !strings.Contains(capturedBody, p) {
			t.Errorf("expected payload to contain %q, got %q", p, capturedBody)
		}
	}
}
