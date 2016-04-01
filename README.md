# textmagic
--
    import "github.com/MJKWoolnough/textmagic"

Package textmagic wraps the API for textmagic.com

## Usage

#### type APIError

```go
type APIError struct {
	Cmd     string
	Code    int    `json:"error_code"`
	Message string `json:"error_message"`
}
```

APIError is an error returned when the incorrect or unexpected data is received

#### func (APIError) Error

```go
func (a APIError) Error() string
```

#### type DeliveryNotificationCode

```go
type DeliveryNotificationCode string
```

DeliveryNotificationCode is a representation of the status of a delivery

#### func (DeliveryNotificationCode) Status

```go
func (d DeliveryNotificationCode) Status() string
```
Status returns the type of status based on the code

#### func (DeliveryNotificationCode) String

```go
func (d DeliveryNotificationCode) String() string
```

#### type JSONError

```go
type JSONError struct {
	Cmd string
	Err error
}
```

JSONError is an error that wraps a JSON error

#### func (JSONError) Error

```go
func (j JSONError) Error() string
```

#### type Message

```go
type Message struct {
	ID        uint64 `json:"message_id"`
	From      string `json:"from"`
	Timestamp int64  `json:"timestamp"`
	Text      string `json:"text"`
}
```

Message represents the information about a received message

#### type Number

```go
type Number struct {
	Price   float32 `json:"price"`
	Country string  `json:"country"`
}
```

Number represents the information about a phone number

#### type Option

```go
type Option func(u url.Values)
```

Option is a type representing a message sending option

#### func  CutExtra

```go
func CutExtra() Option
```
CutExtra sets the option to automatically trim overlong messages

#### func  From

```go
func From(from string) Option
```
From is an option to modify the sender of a message

#### func  MaxLength

```go
func MaxLength(length uint64) Option
```
MaxLength is an option to limit the length of a message

#### func  SendTime

```go
func SendTime(t time.Time) Option
```
SendTime sets the option to schedule the sending of a message for a specific
time

#### type RequestError

```go
type RequestError struct {
	Cmd string
	Err error
}
```

RequestError is an error which wraps an error that occurs while making an API
call

#### func (RequestError) Error

```go
func (r RequestError) Error() string
```

#### type Status

```go
type Status struct {
	Text      string                   `json:"text"`
	Status    DeliveryNotificationCode `json:"status"`
	Created   int64                    `json:"created_time"`
	Reply     string                   `json:"reply_number"`
	Cost      float32                  `json:"credits_cost"`
	Completed int64                    `json:"completed_time"`
}
```

Status represents all of the information about a sent/pending message

#### type StatusError

```go
type StatusError struct {
	Cmd        string
	StatusCode int
}
```

StatusError is an error that is returned when a non-200 OK http response is
received

#### func (StatusError) Error

```go
func (s StatusError) Error() string
```

#### type TextMagic

```go
type TextMagic struct {
}
```

TextMagic contains the data necessary for performing API requests

#### func  New

```go
func New(username, password string) TextMagic
```
New constructs a new TextMagic session

#### func (TextMagic) Account

```go
func (t TextMagic) Account() (float32, error)
```
Account returns the balance of the given TextMagic account

#### func (TextMagic) CheckNumber

```go
func (t TextMagic) CheckNumber(numbers []string) (map[string]Number, error)
```
CheckNumber is used to get the cost and country for the given phone numbers

#### func (TextMagic) DeleteReply

```go
func (t TextMagic) DeleteReply(ids []string) ([]string, error)
```
DeleteReply will simple delete message replies with the given ids

#### func (TextMagic) MessageStatus

```go
func (t TextMagic) MessageStatus(ids []string) (map[string]Status, error)
```
MessageStatus gathers information about the messages with the given ids

#### func (TextMagic) Receive

```go
func (t TextMagic) Receive(lastRetrieved uint64) (uint64, []Message, error)
```
Receive will retrieve the number of unread messages and the 100 latest replies

#### func (TextMagic) Send

```go
func (t TextMagic) Send(message string, to []string, options ...Option) (map[string]string, string, uint, error)
```
Send will send a message to the given recipients. It takes options to modify the
scheduling. sender and length of the message
