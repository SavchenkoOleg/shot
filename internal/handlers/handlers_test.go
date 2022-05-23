package handlers_test

import (
	"encoding/json"
	"flag"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/SavchenkoOleg/shot/internal/handlers"
	"github.com/SavchenkoOleg/shot/internal/storage"
)

func testingPostHandler(conf storage.AppContext, t *testing.T) {
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
			name: "positive test Post #1",
			body: "http://yandex.ru",
			want: want{
				code:     201,
				response: conf.FullPathTest + "1",
			},
		},
		{
			name: "positive test Post #2",
			body: "http://mail.ru",
			want: want{
				code:     201,
				response: conf.FullPathTest + "2",
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
			h := http.HandlerFunc(handlers.HandlerShot(&conf))
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

func testingGetHandler(conf storage.AppContext, t *testing.T) {
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
			name:   "negative GET test #2 (empty target)",
			target: conf.BaseURL + conf.ServerAdress,
			want: want{
				code:     400,
				response: "The parameter is missing\n",
			},
		},
		{
			name:   "negative GET test #3 (bed target)",
			target: conf.FullPathTest + "3",
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
			h := http.HandlerFunc(handlers.HandlerIndex(&conf))
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

func testingPostHandlerJSON(conf storage.AppContext, t *testing.T) {

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
				response: outSt{Result: conf.FullPathTest + "3"},
			},
		},
		{
			name: "positive JSON test #2",
			body: inSt{URL: "https://jsoneditoronline.org"},
			want: want{
				code:     201,
				response: outSt{Result: conf.FullPathTest + "4"},
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
			h := http.HandlerFunc(handlers.HandlerShotJSON(&conf))
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

type flagConfigStruct struct {
	serverAdress    string
	baseURL         string
	fileStoragePath string
}

var flagConfig flagConfigStruct

func hendlerSetting(flags flagConfigStruct) (outConf storage.AppContext) {

	// значения по умолчанию
	outConf.NewURLPref = "newURL"
	outConf.ServerAdress = "localhost:8080"
	outConf.BaseURL = "shot"
	outConf.FileStorage = false
	outConf.FileStoragePath = ""

	// переменные окружения
	BaseURL, exp := os.LookupEnv("BASE_URL")
	if exp {
		outConf.BaseURL = BaseURL
	}

	serverAdress, exp := os.LookupEnv("SERVER_ADDRESS")
	if exp {
		outConf.ServerAdress = serverAdress
	}

	outConf.FileStoragePath, outConf.FileStorage = os.LookupEnv("FILE_STORAGE_PATH")

	// флаги, если они есть
	if flags.baseURL != "" {
		outConf.BaseURL = flags.baseURL

	}

	if flags.serverAdress != "" {
		outConf.ServerAdress = flags.serverAdress

	}

	if flags.fileStoragePath != "" {
		outConf.FileStoragePath = flags.fileStoragePath
		outConf.FileStorage = true
	}

	//  синтез пути, используемый в тестах
	outConf.FullPathTest = "http://" + outConf.ServerAdress + "/" + outConf.BaseURL + "/" + outConf.NewURLPref

	return outConf
}

func TestHandlerShot(t *testing.T) {

	var flagConfig flagConfigStruct

	// init conf
	flag.StringVar(&flagConfig.serverAdress, "a", "", "analog of environment variable SERVER_ADDRESS")
	flag.StringVar(&flagConfig.baseURL, "b", "", "analog of environment variable BASE_URL")
	flag.StringVar(&flagConfig.fileStoragePath, "f", "", "analog of environment variable FILE_STORAGE_PATH")
	flag.Parse()

	conf := hendlerSetting(flagConfig)

	testingPostHandler(conf, t)

	testingGetHandler(conf, t)

	testingPostHandlerJSON(conf, t)

}
