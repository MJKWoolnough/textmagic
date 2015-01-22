package textmagic

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"reflect"
	"testing"
)

var apiConfig TextMagic

const (
	testUsername = "testUser"
	testPassword = "testPassword"
)

func apiServer(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "invalid method", http.StatusMethodNotAllowed)
		return
	}
	r.ParseForm()
	if r.Form.Get("username") != testUsername || r.Form.Get("password") != testPassword {
		w.Write([]byte(`{"error_code":5,"error_message":"Invalid username & password combination"}`))
		return
	}
	switch r.Form.Get("cmd") {
	case "send":
	case "account":
		w.Write([]byte(`{"balance":417.7}`))
	case "message_status":
	case "receive":
	case "delete_reply":
	case "check_number":
	default:
		w.Write([]byte(`{"error_code":3,"error_message":"Command is undefined"}`))
	}

}

func TestAccount(t *testing.T) {
	balance, err := apiConfig.Account()
	if err != nil {
		t.Errorf("test 1: unexpected error, %s", err)
	} else if balance != 417.7 {
		t.Errorf("test 1: expecting balance 417.7, got %f", balance)
	}
	_, err = New("", "").Account()
	if e, ok := err.(APIError); !ok {
		t.Errorf("test 2: expecting error of type APIError, got %s", err)
	} else if e.Code != 5 {
		t.Errorf("test 2: expecting error value of 5, got %d", e.Code)
	}

}

func TestWrongCommand(t *testing.T) {
	err := apiConfig.sendAPI("unknown_command", url.Values{}, struct{}{})
	if err == nil {
		t.Errorf("expecting error, received nil")
	} else if e, ok := err.(APIError); !ok {
		t.Errorf("expecting error of type APIError, got %s: %s", reflect.TypeOf(err).Name, err)
	} else if e.Code != 3 {
		t.Errorf("expecting error code 3 (\"Command is undefined\"), got %d (%q)", e.Code, e.Message)
	}
}

func TestMain(m *testing.M) {
	username := os.Getenv("TMusername")
	password := os.Getenv("TMpassword")
	if username == "" || password == "" {
		username = testUsername
		password = testPassword
		s := httptest.NewServer(http.HandlerFunc(apiServer))
		apiURLPrefix = s.URL + "?"
	}
	apiConfig = New(username, password)
	os.Exit(m.Run())
}
