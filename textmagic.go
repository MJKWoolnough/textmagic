package textmagic

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"github.com/MJKWoolnough/memio"
)

var apiUrlPrefix = "https://www.textmagic.com/app/api?"

const (
	cmdAccount       = "account"
	cmdMessageStatus = "message_status"
)

type TextMagic struct {
	username, password string
}

func New(username, password string) TextMagic {
	return TextMagic{username, password}
}

func (t TextMagic) sendAPI(cmd string, params url.Values, data interface{}) error {
	params.Add("username", t.username)
	params.Add("password", t.password)
	params.Add("cmd", cmd)
	r, err := http.Get(apiUrlPrefix + params.Encode())
	if err != nil {
		return RequestError{cmd, err}
	}
	defer r.Body.Close()
	if r.StatusCode != http.StatusOK {
		return StatusError{cmd, r.StatusCode}
	}
	jsonData := make([]byte, r.ContentLength) // avoid allocation using io.Pipe?
	var apiError APIError
	err = json.NewDecoder(io.TeeReader(r.Body, memio.Create(&jsonData))).Decode(&apiError)
	if err != nil {
		return JSONError{cmd, err}
	}
	if apiError.Code != 0 {
		apiError.Cmd = cmd
		return apiError
	}
	json.Unmarshal(jsonData, data)
	return nil
}

type balance struct {
	Balance float32 `json:"balance"`
}

func (t TextMagic) Account() (float32, error) {
	var b balance
	if err := t.sendAPI(cmdAccount, url.Values{}, &b); err != nil {
		return 0, err
	}
	return b.Balance, nil
}

type Status struct {
	Text      string  `json:"text"`
	Status    string  `json:"status"`
	Created   int64   `json:"created_time"`
	Reply     string  `json:"reply_number"`
	Cost      float32 `json:"credits_cost"`
	Completed int64   `json:"completed_time"`
}

type messageStatuses map[int]Status

func (t TextMagic) MessageStatus(ids ...int) (map[int]Status, error) {
	statuses := make(map[int]Status)
	for len(ids) > 0 {
		var tIds []int
		if len(ids) > 100 {
			tIds = ids[:100]
			ids = ids[100:]
		} else {
			tIds = ids
			ids = nil
		}
		var messageIds string
		for _, id := range tIds {
			if len(messageIds) > 0 {
				messageIds += ","
			}
			messageIds += strconv.Itoa(id)
		}
		strStatuses := make(map[string]Status)
		err := t.sendAPI(cmdMessageStatus, url.Values{"ids": {messageIds}}, strStatuses)
		if err != nil {
			return statuses, err
		}
		for messageID, status := range strStatuses {
			id, err := strconv.Atoi(messageID)
			if err != nil {
				continue
			}
			statuses[id] = status
		}
	}
	return statuses, nil
}

// Errors

type APIError struct {
	Cmd     string
	Code    int    `json:"error_code"`
	Message string `json:"error_message"`
}

func (a APIError) Error() string {
	return "command " + a.Cmd + " returned the following API error: " + a.Message
}

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
