package models

import (
	"bytes"
	"encoding/json"
)

type Message struct {
	Username string `json:"username"`
	Channel  string `json:"channel"`
	Action   string `json:"action"`
	Data     string `json:"data"`
}

func (msg Message) Decode(message string) error  {

	jsonMsg := string(message)
	decoder := json.NewDecoder(bytes.NewBufferString(jsonMsg))
	return decoder.Decode(&msg)

}