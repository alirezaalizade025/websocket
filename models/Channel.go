package models

import (
	"encoding/json"
	"errors"
	"log"
)

type Channel struct {
	ChannelName string   `json:"chanel_name"`
	ChannelUsers       []string `json:"users"`
}

var Channels = []Channel{}


func (c Channel) Join(username string, channelName string) {

	c.ChannelUsers = append(c.ChannelUsers, username)

}

func (c Channel) Leave(username string, channelName string) {


	for i, u := range c.ChannelUsers {
		if u == username {
			c.ChannelUsers = append(c.ChannelUsers[:i], c.ChannelUsers[i+1:]...)
		}
	}

}

func (c Channel) FirstOrCreate(channelName string) error {

	if !channelExists(channelName) {

		newChannel := Channel{
			ChannelName:  channelName,
			ChannelUsers: []string{},
		}

		Channels = append(Channels, newChannel)
	}

	for _, item := range Channels {
		if item.ChannelName == channelName {
			c = item
		}
	}

	return errors.New("channel not found")
}

func (c Channel) InfoMessage() []byte {

	infoJson, err := json.Marshal(c)
	if err != nil {
		log.Panicln(err)
	}

	info, err := json.Marshal(Message{
		Username: "SERVER",
		Channel:  c.ChannelName,
		Action:   "channel-info",
		Data:     string(infoJson),
	})

	if err != nil {
		log.Panicln(err)
	}
	return info
}

func channelFirst(channelName string) (c Channel, err error) {

	if channelExists(channelName) {
		for _, c := range Channels {
			if c.ChannelName == channelName {
				return c, nil
			}
		}
	}

	return c, errors.New("channel not found")
}

func channelExists(name string) bool {
	for _, c := range Channels {
		if c.ChannelName == name {
			return true
		}
	}
	return false
}