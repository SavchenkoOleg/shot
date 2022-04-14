package handlers_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/SavchenkoOleg/shot/internal/handlers"
)

func testingPostHandler(t *testing.T) {
	type want struct {
		code     int
		response string
	}

	var HendlerSetting handlers.ServConfig
	HendlerSetting = handlers.HendlerSetting()

	tests := []struct {
		name string
		body string
		want want
	}{
		{
			name: "positive test Post #1",
			body: "http://yandex.ru",
			want: want{
				code:     201,
				response: "http://localhost" + HendlerSetting.ServerAdress + "/" + HendlerSetting.NewURLPref + "1",
			},
		},
		{
			name: "positive test Post #2",
			body: "http://mail.ru",
			want: want{
				code:     201,
				response: "http://localhost" + HendlerSetting.ServerAdress + "/" + HendlerSetting.NewURLPref + "2",
			},
		},
		{
			name: "negative test Post #1 (empty body quest)",
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

	var HendlerSetting handlers.ServConfig
	HendlerSetting = handlers.HendlerSetting()

	tests := []struct {
		name   string
		target string
		want   want
	}{
		{
			name:   "negative GET test #2 (empty target)",
			target: "http://localhost" + HendlerSetting.ServerAdress,
			want: want{
				code:     400,
				response: "The parameter is missing\n",
			},
		},
		{
			name:   "negative GET test #3 (bed target)",
			target: "http://localhost" + HendlerSetting.ServerAdress + "/" + HendlerSetting.NewURLPref + "3",
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

func testingPostHandlerJSON(t *testing.T) {

	type inSt struct {
		URL string `json:"url"`
	}
	type outSt struct {
		Result string `json:"result"`
	}

	type want struct {
		code     int
		response outSt
	}

	var HendlerSetting handlers.ServConfig
	HendlerSetting = handlers.HendlerSetting()

	tests := []struct {
		name string
		body inSt
		want want
	}{
		{
			name: "positive JSON test #1",
			body: inSt{URL: "https://golang-blog.blogspot.com"},
			want: want{
				code:     201,
				response: outSt{Result: "http://localhost" + HendlerSetting.ServerAdress + "/" + HendlerSetting.NewURLPref + "3"},
			},
		},
		{
			name: "positive JSON test #2",
			body: inSt{URL: "https://jsoneditoronline.org"},
			want: want{
				code:     201,
				response: outSt{Result: "http://localhost" + HendlerSetting.ServerAdress + "/" + HendlerSetting.NewURLPref + "4"},
			},
		},
		{
			name: "negative JSON test #1 (empty body quest)",
			body: inSt{},
			want: want{
				code:     400,
				response: outSt{},
			},
		},
	}
	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {
			tx, _ := json.Marshal(tt.body)
			bodyReader := strings.NewReader(string(tx))
			request := httptest.NewRequest(http.MethodPost, "/api/shorten", bodyReader)

			w := httptest.NewRecorder()
			h := http.HandlerFunc(handlers.HandlerShotJSON)
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

			txSt := outSt{}
			json.Unmarshal(resBody, &txSt)

			if txSt != tt.want.response {
				t.Errorf("Expected body %s, got %s", tt.want.response.Result, txSt.Result)
			}

		})
	}
}

func TestHandlerShot(t *testing.T) {

	testingPostHandler(t)

	testingGetHandler(t)

	testingPostHandlerJSON(t)

}
