package textmagic

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/MJKWoolnough/memio"
)

var apiURLPrefix = "https://www.textmagic.com/app/api?"

const (
	cmdAccount       = "account"
	cmdCheckNumber   = "check_number"
	cmdDeleteReply   = "delete_reply"
	cmdMessageStatus = "message_status"
	cmdReceive       = "receive"
	cmdSend          = "send"
)

type TextMagic struct {
	username, password string
}

func New(username, password string) TextMagic {
	return TextMagic{username, password}
}

func (t TextMagic) sendAPI(cmd string, params url.Values, data interface{}) error {
	params.Set("username", t.username)
	params.Set("password", t.password)
	params.Set("cmd", cmd)
	r, err := http.Get(apiURLPrefix + params.Encode())
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

type DeliveryNotificationCode string

func (d DeliveryNotificationCode) Status() string {
	switch d {
	case "q", "r", "a", "b", "s":
		return "intermediate"
	case "d", "f", "e", "j", "u":
		return "final"
	}
	return "unknown"
}

func (d DeliveryNotificationCode) String() string {
	switch d {
	case "q":
		return "The message is queued on the TextMagic server."
	case "r":
		return "The message has been sent to the mobile operator."
	case "a":
		return "The mobile operator has acknowledged the message."
	case "b":
		return "The mobile operator has queued the message."
	case "d":
		return "The message has been successfully delivered to the handset."
	case "f":
		return "An error occurred while delivering message."
	case "e":
		return "An error occurred while sending message."
	case "j":
		return "The mobile operator has rejected the message."
	case "s":
		return "This message is scheduled to be sent later."
	default:
		return "The status is unknown."

	}
}

type Status struct {
	Text      string                   `json:"text"`
	Status    DeliveryNotificationCode `json:"status"`
	Created   int64                    `json:"created_time"`
	Reply     string                   `json:"reply_number"`
	Cost      float32                  `json:"credits_cost"`
	Completed int64                    `json:"completed_time"`
}

func (t TextMagic) MessageStatus(ids []uint64) (map[uint64]Status, error) {
	statuses := make(map[uint64]Status)
	for _, tIds := range splitSlice(ids) {
		messageIds := joinUints(tIds)
		strStatuses := make(map[string]Status)
		err := t.sendAPI(cmdMessageStatus, url.Values{"ids": {messageIds}}, strStatuses)
		if err != nil {
			return statuses, err
		}
		for messageID, status := range strStatuses {
			if id, isNum := stou(messageID); isNum {
				statuses[id] = status
			}
		}
	}
	return statuses, nil
}

type Number struct {
	Price   float32 `json:"price"`
	Country string  `json:"country"`
}

func (t TextMagic) CheckNumber(numbers []uint64) (map[uint64]Number, error) {
	ns := make(map[string]Number)
	if err := t.sendAPI(cmdCheckNumber, url.Values{"phone": {joinUints(numbers)}}, ns); err != nil {
		return nil, err
	}
	toRet := make(map[uint64]Number)
	for number, data := range ns {
		if n, isNum := stou(number); isNum {
			toRet[n] = data
		}
	}
	return toRet, nil
}

type deleted struct {
	Deleted []uint64 `json:"deleted"`
}

func (t TextMagic) DeleteReply(ids []uint64) ([]uint64, error) {
	toRet := make([]uint64, 0, len(ids))
	for _, tIds := range splitSlice(ids) {
		var d deleted
		if err := t.sendAPI(cmdDeleteReply, url.Values{"deleted": {joinUints(tIds)}}, &d); err != nil {
			return toRet, err
		}
		toRet = append(toRet, d.Deleted...)
	}
	return toRet, nil
}

type Message struct {
	ID        uint64 `json:"message_id"`
	From      uint64 `json:"from"`
	Timestamp int64  `json:"timestamp"`
	Text      string `json:"text"`
}

type received struct {
	Messages []Message `json:"messages"`
	Unread   uint64    `json:"unread"`
}

func (t TextMagic) Receive(lastRetrieved uint64) (uint64, []Message, error) {
	var r received
	err := t.sendAPI(cmdReceive, url.Values{"last_retrieved_id": {utos(lastRetrieved)}}, &r)
	return r.Unread, r.Messages, err
}

type option func(u url.Values)

func From(from uint64) option {
	return func(u url.Values) {
		u.Set("from", utos(from))
	}
}

func MaxLength(length uint64) option {
	if length > 3 {
		length = 3
	}
	return func(u url.Values) {
		u.Set("max_length", utos(length))
	}
}

func CutExtra() option {
	return func(u url.Values) {
		u.Set("cut_extra", "1")
	}
}

func SendTime(t time.Time) option {
	return func(u url.Values) {
		u.Set("send_time", t.Format(time.RFC3339))
	}
}

type messageResponse struct {
	IDs   map[string]string `json:"message_id"`
	Text  string            `json:"sent_text"`
	Parts uint              `json:"parts_count"`
}

func (t TextMagic) Send(message string, to []uint64, options ...option) (map[string]uint64, string, uint, error) {
	var (
		params url.Values
		text   string
		parts  uint
		ids    = make(map[string]uint64)
	)
	// check message for unicode/invalid chars
	params.Set("text", message)
	for _, o := range options {
		o(params)
	}
	for _, numbers := range splitSlice(to) {
		params.Set("phone", joinUints(numbers))
		var m messageResponse
		if err := t.sendAPI(cmdSend, params, &m); err != nil {
			return ids, text, parts, err
		}
		if parts == 0 {
			parts = m.Parts
			text = m.Text
		}
		for id, number := range m.IDs {
			if n, isNum := stou(number); isNum {
				ids[id] = n
			}
		}
	}
	return ids, text, parts, nil
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
