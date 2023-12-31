package internal

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/glair-ai/glair-vision-go"
	"github.com/stretchr/testify/assert"
)

type mockStruct struct {
	Name string `json:"name,omitempty"`
}

type failingClient struct{}

func (c failingClient) Do(req *http.Request) (*http.Response, error) {
	return nil, errors.New("failed to send request")
}

func TestMakeRequest(t *testing.T) {
	file, _ := os.Open("../examples/ocr/images/ktp.jpeg")

	tests := []struct {
		name       string
		config     *glair.Config
		mockServer *httptest.Server
		want       mockStruct
		wantErr    *glair.Error
	}{
		{
			name: "failed to send request due to bad url",
			config: glair.NewConfig("username", "password", "api-key").
				WithBaseURL("%+0"),
			want: mockStruct{},
			wantErr: &glair.Error{
				Code:    glair.ErrorCodeInvalidURL,
				Message: "Invalid base URL is provided in configuration.",
			},
		},
		{
			name: "failed to send request due to bad client",
			config: glair.NewConfig("username", "password", "api-key").
				WithClient(failingClient{}),
			want: mockStruct{},
			wantErr: &glair.Error{
				Code:    glair.ErrorCodeBadClient,
				Message: "Bad HTTP client is provided in configuration.",
			},
		},
		{
			name:   "response is not OK, handled error",
			config: glair.NewConfig("username", "password", "api-key"),
			mockServer: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(`{"status": "NO_FILE", "reason": "No file in request body"}`))
			})),
			want: mockStruct{},
			wantErr: &glair.Error{
				Code:    glair.ErrorCodeAPIError,
				Message: "GLAIR API returned non-OK response. Please check the Response property for more detailed explanation.",
				Response: glair.Response{
					Code:   400,
					Status: "NO_FILE",
					Reason: "No file in request body",
				},
			},
		},
		{
			name:   "response is not OK, auth error",
			config: glair.NewConfig("username", "password", "api-key"),
			mockServer: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"message": "Access to this API has been disallowed."}`))
			})),
			want: mockStruct{},
			wantErr: &glair.Error{
				Code:    glair.ErrorCodeAPIError,
				Message: "GLAIR API returned non-OK response. Please check the Response property for more detailed explanation.",
				Response: glair.Response{
					Code:   401,
					Reason: "Access to this API has been disallowed.",
				},
			},
		},
		{
			name:   "response is not OK, gateway and miscellanous errors",
			config: glair.NewConfig("username", "password", "api-key"),
			mockServer: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadGateway)
				w.Write([]byte(`28937641y28r12fg`))
			})),
			want: mockStruct{},
			wantErr: &glair.Error{
				Code:    glair.ErrorCodeInvalidResponse,
				Message: "Failed to parse API response. Please contact us about this error.",
				Response: glair.Response{
					Code: 502,
				},
			},
		},
		{
			name:   "success",
			config: glair.NewConfig("username", "password", "api-key"),
			mockServer: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"name":"foo"}`))
			})),
			want: mockStruct{
				Name: "foo",
			},
			wantErr: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			url := tc.config.BaseUrl
			if tc.mockServer != nil {
				url = tc.mockServer.URL
			}

			params := RequestParameters{
				Url:       url,
				RequestID: "samples",
				Payload: map[string]*os.File{
					"image": file,
				},
			}

			res, err := MakeRequest[mockStruct](
				context.TODO(),
				params,
				tc.config,
			)

			assert.Equal(t, tc.want, res)

			if tc.wantErr == nil {
				assert.Equal(t, nil, err)
			} else {
				glairError := err.(*glair.Error)

				assert.Equal(t, tc.wantErr.Code, glairError.Code)
				assert.Equal(t, tc.wantErr.Message, glairError.Message)
				assert.Equal(t, tc.wantErr.Response, glairError.Response)
			}
		})
	}
}
