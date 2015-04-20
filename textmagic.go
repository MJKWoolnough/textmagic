package textmagic

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
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

// TextMagic contains the data necessary for performing API requests
type TextMagic struct {
	username, password string
}

// New constructs a new TextMagic session
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
	cL := r.ContentLength
	if cL < 0 {
		cL = 1024
	}
	jsonData := make([]byte, 0, cL) // avoid allocation using io.Pipe?
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

// Account returns the balance of the given TextMagic account
func (t TextMagic) Account() (float32, error) {
	var b balance
	if err := t.sendAPI(cmdAccount, url.Values{}, &b); err != nil {
		return 0, err
	}
	return b.Balance, nil
}

// DeliveryNotificationCode is a representation of the status of a delivery
type DeliveryNotificationCode string

// Status returns the type of status based on the code
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

// Status represents all of the information about a sent/pending message
type Status struct {
	Text      string                   `json:"text"`
	Status    DeliveryNotificationCode `json:"status"`
	Created   int64                    `json:"created_time"`
	Reply     string                   `json:"reply_number"`
	Cost      float32                  `json:"credits_cost"`
	Completed int64                    `json:"completed_time"`
}

const joinSep = ","

// MessageStatus gathers information about the messages with the given ids
func (t TextMagic) MessageStatus(ids []string) (map[string]Status, error) {
	statuses := make(map[string]Status)
	for _, tIds := range splitSlice(ids) {
		messageIds := strings.Join(tIds, joinSep)
		strStatuses := make(map[string]Status)
		err := t.sendAPI(cmdMessageStatus, url.Values{"ids": {messageIds}}, strStatuses)
		if err != nil {
			return statuses, err
		}
		for messageID, status := range strStatuses {
			statuses[messageID] = status
		}
	}
	return statuses, nil
}

// Number represents the information about a phone number
type Number struct {
	Price   float32 `json:"price"`
	Country string  `json:"country"`
}

// CheckNumber is used to get the cost and country for the given phone numbers
func (t TextMagic) CheckNumber(numbers []string) (map[string]Number, error) {
	ns := make(map[string]Number)
	if err := t.sendAPI(cmdCheckNumber, url.Values{"phone": {strings.Join(numbers, joinSep)}}, ns); err != nil {
		return nil, err
	}
	return ns, nil
}

type deleted struct {
	Deleted []string `json:"deleted"`
}

// DeleteReply will simple delete message replies with the given ids
func (t TextMagic) DeleteReply(ids []string) ([]string, error) {
	toRet := make([]string, 0, len(ids))
	for _, tIds := range splitSlice(ids) {
		var d deleted
		if err := t.sendAPI(cmdDeleteReply, url.Values{"deleted": {strings.Join(tIds, joinSep)}}, &d); err != nil {
			return toRet, err
		}
		toRet = append(toRet, d.Deleted...)
	}
	return toRet, nil
}

// Message represents the information about a received message
type Message struct {
	ID        uint64 `json:"message_id"`
	From      string `json:"from"`
	Timestamp int64  `json:"timestamp"`
	Text      string `json:"text"`
}

type received struct {
	Messages []Message `json:"messages"`
	Unread   uint64    `json:"unread"`
}

// Receive will retrieve the number of unread messages and the 100 latest
// replies
func (t TextMagic) Receive(lastRetrieved uint64) (uint64, []Message, error) {
	var r received
	err := t.sendAPI(cmdReceive, url.Values{"last_retrieved_id": {utos(lastRetrieved)}}, &r)
	return r.Unread, r.Messages, err
}

// Option is a type representing a message sending option
type Option func(u url.Values)

// From is an option to modify the sender of a message
func From(from string) Option {
	return func(u url.Values) {
		u.Set("from", from)
	}
}

// MaxLength is an option to limit the length of a message
func MaxLength(length uint64) Option {
	if length > 3 {
		length = 3
	}
	return func(u url.Values) {
		u.Set("max_length", utos(length))
	}
}

// CutExtra sets the option to automatically trim overlong messages
func CutExtra() Option {
	return func(u url.Values) {
		u.Set("cut_extra", "1")
	}
}

// SendTime sets the option to schedule the sending of a message for a specific
// time
func SendTime(t time.Time) Option {
	return func(u url.Values) {
		u.Set("send_time", t.Format(time.RFC3339))
	}
}

type messageResponse struct {
	IDs   map[string]string `json:"message_id"`
	Text  string            `json:"sent_text"`
	Parts uint              `json:"parts_count"`
}

// Send will send a message to the given recipients. It takes options to modify
// the scheduling. sender and length of the message
func (t TextMagic) Send(message string, to []string, options ...Option) (map[string]string, string, uint, error) {
	var (
		params = url.Values{}
		text   string
		parts  uint
		ids    = make(map[string]string)
	)
	// check message for unicode/invalid chars
	params.Set("text", message)
	params.Set("unicode", "0")
	for _, o := range options {
		o(params)
	}
	for _, numbers := range splitSlice(to) {
		params.Set("phone", strings.Join(numbers, joinSep))
		var m messageResponse
		if err := t.sendAPI(cmdSend, params, &m); err != nil {
			return ids, text, parts, err
		}
		if parts == 0 {
			parts = m.Parts
			text = m.Text
		}
		for id, number := range m.IDs {
			ids[id] = number
		}
	}
	return ids, text, parts, nil
}

// Errors

// APIError is an error returned when the incorrect or unexpected data is
// received
type APIError struct {
	Cmd     string
	Code    int    `json:"error_code"`
	Message string `json:"error_message"`
}

func (a APIError) Error() string {
	return "command " + a.Cmd + " returned the following API error: " + a.Message
}

// RequestError is an error which wraps an error that occurs while making an
// API call
type RequestError struct {
	Cmd string
	Err error
}

func (r RequestError) Error() string {
	return "command " + r.Cmd + " returned the following error while makeing the API call: " + r.Err.Error()
}

// StatusError is an error that is returned when a non-200 OK http response is
// received
type StatusError struct {
	Cmd        string
	StatusCode int
}

func (s StatusError) Error() string {
	return "command " + s.Cmd + " returned a non-200 OK response: " + http.StatusText(s.StatusCode)
}

// JSONError is an error that wraps a JSON error
type JSONError struct {
	Cmd string
	Err error
}

func (j JSONError) Error() string {
	return "command " + j.Cmd + " returned malformed JSON: " + j.Err.Error()
}
