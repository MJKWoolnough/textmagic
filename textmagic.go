package textmagic

import (
	"encoding/json"
	"net/http"
	"net/url"
)

const apiUrlPrefix string = "https://www.textmagic.com/app/api?"

const (
	cmdAccount = "account"
)

type TextMagic struct {
	username, password string
}

func New(username, password string) TextMagic {
	return TextMagic{username, password}
}

func (t TextMagic) newParams(cmd string) url.Values {
	var params url.Values
	params.Add("username", t.username)
	params.Add("password", t.password)
	params.Add("cmd", cmd)
	return params
}

type APIError struct {
	Cmd     string
	Code    int    `json:"error_code"`
	Message string `json:"error_message"`
}

func (a APIError) Error() string {
	return "api error: " + a.Message
}

type Balance struct {
	APIError
	Balance float32 `json:"balance"`
}

func (t TextMagic) Account() (float32, error) {
	r, err := http.Get(apiUrlPrefix + t.newParams(cmdAccount).Encode())
	if err != nil {
		return 0, RequestError{cmdAccount, err}
	}
	defer r.Body.Close()
	if r.StatusCode != http.StatusOK {
		return 0, StatusError{cmdAccount, r.StatusCode}
	}
	var b Balance
	err = json.NewDecoder(r.Body).Decode(&b)
	if err != nil {
		return 0, JSONError{cmdAccount, err}
	}
	if b.Code != 0 {
		b.APIError.Cmd = cmdAccount
		return 0, b.APIError
	}
	return b.Balance, nil
}

// Errors

type RequestError struct {
	Cmd string
	Err error
}

func (r RequestError) Error() string {
	return "command " + r.Cmd + " returned the following error while makeing the API call: " + r.Err.Error()
}

type StatusError struct {
	Cmd        string
	StatusCode int
}

func (s StatusError) Error() string {
	return "command " + s.Cmd + " returned a non-200 OK response: " + http.StatusText(s.StatusCode)
}

type JSONError struct {
	Cmd string
	Err error
}

func (j JSONError) Error() string {
	return "command " + j.Cmd + " returned malformed JSON: " + j.Err.Error()
}
