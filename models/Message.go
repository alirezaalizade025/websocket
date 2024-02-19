package models

import (
	"encoding/json"
)

type Message struct {
	Username    string      `json:"username"`
	ChannelName string      `json:"channel"`
	Action      string      `json:"action"`
	Data        interface{} `json:"data"`
}

func (msg *Message) Decode(message []byte) error {

	return json.Unmarshal(message, &msg)
}