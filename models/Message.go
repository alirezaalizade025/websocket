package models

import (
	"encoding/json"
	"errors"
	"log"
)

// type Message struct {
// 	ChannelName string      `json:"channel"`
// 	Username    string      `json:"username"`
// 	Action      string      `json:"action"`
// 	Data        interface{} `json:"data"`
// }

// func (msg *Message) Decode(message []byte) error {

// 	return json.Unmarshal(message, &msg)
// }

type MessageType string
type SendMode string

const (
	MessageTypeInit           MessageType = "init"
	MessageTypeInitAdmin      MessageType = "init_admin"
	MessageTypeActive         MessageType = "active"
	MessageTypeInactive       MessageType = "inactive"
	MessageTypePong           MessageType = "pong"
	MessageTypeChannelInfo    MessageType = "channel_info"
	MessageTypeChannelJoin    MessageType = "join"
	MessageTypeChannelLeave   MessageType = "leave"
	MessageTypeChannelGeneral MessageType = "general"
	MessageTypeAPI            MessageType = "api"
	MessageTypeP2P            MessageType = "p2p"
)

const (
	SendModeActiveOnly  SendMode = "active_only"
	SendModeActiveFirst SendMode = "active_first"
	SendModeAll         SendMode = "all"
)

type Message struct {
	ChannelName *string     `json:"channel,omitempty"`
	Username    *string     `json:"username,omitempty"`
	SendMode    *string     `json:"send_mode,omitempty"`
	Action      string      `json:"action"`
	Data        interface{} `json:"data"`
}

func (msg *Message) Decode(message []byte) error {
	var tmp struct {
		ChannelName *string `json:"channel,omitempty"`
		Username    *string `json:"username,omitempty"`
		SendMode    *string `json:"send_mode,omitempty"`
		Action      string  `json:"action"`
		Data        interface{}
	}

	if err := json.Unmarshal(message, &tmp); err != nil {
		return err
	}

	action := (MessageType)(tmp.Action)



	if action != "init" &&
		action != "init_admin" &&
		action != "pong" &&
		action != "active" &&
		action != "inactive" &&
		(tmp.ChannelName == nil && tmp.Username == nil) {

		return errors.New("either channel or user name must be provided")
	}

	if action == MessageTypeP2P && tmp.Username == nil {
		log.Println("user name must be provided for p2p message")
		return nil
	}

	msg.ChannelName = tmp.ChannelName
	msg.Username = tmp.Username
	msg.Action = tmp.Action
	msg.Data = tmp.Data

	return nil
}
