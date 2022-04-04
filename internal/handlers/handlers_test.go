package handlers_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/SavchenkoOleg/shot.git/internal/handlers"
)

func testingPostHandler(t *testing.T) {
	type want struct {
		code     int
		response string
	}

	tests := []struct {
		name string
		body string
		want want
	}{
		{
			name: "positive test #1",
			body: "http://yandex.ru",
			want: want{
				code:     201,
				response: "http://localhost:8080/newURL1",
			},
		},
		{
			name: "positive test #2",
			body: "http://mail.ru",
			want: want{
				code:     201,
				response: "http://localhost:8080/newURL2",
			},
		},
		{
			name: "negative test #1 (empty body quest)",
			body: "",
			want: want{
				code:     400,
				response: "uncorrect reduction URL\n",
			},
		},
	}
	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {
			bodyReader := strings.NewReader(tt.body)
			request := httptest.NewRequest(http.MethodPost, "/", bodyReader)

			w := httptest.NewRecorder()
			h := http.HandlerFunc(handlers.HandlerShot)
			h.ServeHTTP(w, request)
			res := w.Result()

			if res.StatusCode != tt.want.code {
				t.Errorf("Expected status code %d, got %d", tt.want.code, w.Code)
			}

			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)
			if err != nil {
				t.Fatal(err)
			}

			if string(resBody) != tt.want.response {
				t.Errorf("Expected body %s, got %s", tt.want.response, w.Body.String())
			}

		})
	}
}

func testingGetHandler(t *testing.T) {
	type want struct {
		code     int
		response string
	}

	tests := []struct {
		name   string
		target string
		want   want
	}{
		{
			name:   "negative test #2 (empty target)",
			target: "http://localhost:8080",
			want: want{
				code:     400,
				response: "The parameter is missing\n",
			},
		},
		{
			name:   "negative test #3 (bed target)",
			target: "http://localhost:8080/newURL3",
			want: want{
				code:     400,
				response: "URL for the specified id was not found\n",
			},
		},
	}
	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, tt.target, nil)

			w := httptest.NewRecorder()
			h := http.HandlerFunc(handlers.HandlerIndex)
			h.ServeHTTP(w, request)
			res := w.Result()

			if res.StatusCode != tt.want.code {
				t.Errorf("Expected status code %d, got %d", tt.want.code, w.Code)
			}

			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)
			if err != nil {
				t.Fatal(err)
			}

			if string(resBody) != tt.want.response {
				t.Errorf("Expected body %s, got %s", tt.want.response, w.Body.String())
			}

		})
	}
}

func TestHandlerShot(t *testing.T) {

	testingPostHandler(t)

	testingGetHandler(t)
}
