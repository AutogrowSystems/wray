package wray

import (
	"bytes"
	"encoding/json"
	// "fmt"
)

// Message models a message sent over Faye
type Message interface {
	Channel() string
	ID() string
	Data() map[string]interface{}
	Ext() map[string]interface{}
	ConnectionType() string
	Decode(interface{}) error
	HasError() bool
	SetError(string)
	Error() string
	MarshalJSON() ([]byte, error)
}

// Response models a response received from the Faye server
type Response interface {
	OK() bool
	Channel() string
	Error() string
	Advice() Advice
	ClientID() string
	SupportedConnectionTypes() []string
	HasError() bool
}

// Advice given by the Bayeux server about how to reconnect, etc
type Advice interface {
	Interval() float64
	Reconnect() string
	Timeout() float64
}

// Models a Bayeux message that can be for sending or receiving
// TODO: omitempty
type message struct {
	ID                       string                 `json:"id,omitempty"`
	Channel                  string                 `json:"channel,omitempty"`
	Successful               bool                   `json:"successful,omitempty"`
	ClientID                 string                 `json:"clientId,omitempty"`
	SupportedConnectionTypes []string               `json:"supportedConnectionTypes,omitempty"`
	Data                     map[string]interface{} `json:"data,omitempty"`
	Advice                   advice                 `json:"advice,omitempty"`
	Error                    string                 `json:"error,omitempty"`
	decoder                  decoder
	Ext                      map[string]interface{} `json:"ext,omitempty"`
	ConnectionType           string                 `json:"connectionType,omitempty"`
	Subscription             string                 `json:"subscription,omitempty"`
	Version                  string                 `json:"version,omitempty"`
}

type msgWrapper struct {
	msg *message
}

func (w msgWrapper) Data() map[string]interface{} { return w.msg.Data }
func (w msgWrapper) ID() string                   { return w.msg.ID }
func (w msgWrapper) Channel() string              { return w.msg.Channel }

// Ext returns a map of extension data.  As it is a map (and thus a pointer), changes
// to the returned object will modify the content of this field in the message
func (w msgWrapper) Ext() map[string]interface{} {
	if len(w.msg.Ext) == 0 {
		w.msg.Ext = map[string]interface{}{}
	}
	return w.msg.Ext
}
func (w msgWrapper) OK() bool                           { return w.msg.Successful }
func (w msgWrapper) Error() string                      { return w.msg.Error }
func (w msgWrapper) HasError() bool                     { return w.msg.Error != "" }
func (w msgWrapper) SetError(msg string)                { w.msg.Error = msg }
func (w msgWrapper) ConnectionType() string             { return w.msg.ConnectionType }
func (w msgWrapper) Subscription() string               { return w.msg.Subscription }
func (w msgWrapper) Advice() Advice                     { return adviceWrapper{w.msg.Advice} }
func (w msgWrapper) Version() string                    { return w.msg.Version }
func (w msgWrapper) SupportedConnectionTypes() []string { return w.msg.SupportedConnectionTypes }
func (w msgWrapper) ClientID() string                   { return w.msg.ClientID }

// Decodes the message data into the given pointer
func (w msgWrapper) Decode(obj interface{}) error {
	b, err := json.Marshal(w.Data())
	if err != nil {
		return err
	}

	return json.NewDecoder(bytes.NewBuffer(b)).Decode(obj)
}

func (w msgWrapper) MarshalJSON() ([]byte, error) {
	return json.Marshal(w.msg)
}

func decodeResponse(dec decoder) (Response, []Message, error) {
	var msgs = []Message{}

	var raw = []*message{}
	if err := dec.Decode(&raw); err != nil {
		return nil, msgs, err
	}

	for _, msg := range raw {
		msgs = append(msgs, Message(msgWrapper{msg}))
	}

	return msgs[0].(Response), msgs[1:], nil
}

type advice struct {
	Reconnect string  `json:"reconnect,omitempty"`
	Interval  float64 `json:"interval,omitempty"`
	Timeout   float64 `json:"timeout,omitempty"`
}

type adviceWrapper struct {
	adv advice
}

func (w adviceWrapper) Interval() float64 { return w.adv.Interval }
func (w adviceWrapper) Reconnect() string { return w.adv.Reconnect }
func (w adviceWrapper) Timeout() float64  { return w.adv.Timeout }
