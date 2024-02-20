package models

import (
	"encoding/json"
	"errors"
	"log"

	"github.com/olahol/melody"
)

type Channel struct {
	ChannelName    string          `json:"channel_name"`
	ChannelClients []ChannelClient `json:"users"`
}

type Status string

const (
	Active   Status = "active"
	Inactive Status = "inactive"
)

type ChannelClient struct {
	ID     string `json:"id"`
	Status Status `json:"status"`
}

var Channels = map[string]*Channel{}

func (c *Channel) Join(id string) {

	// append if not exists
	for _, client := range c.ChannelClients {
		if client.ID == id {
			return
		}
	}

	c.ChannelClients = append(c.ChannelClients, ChannelClient{
		ID:     id,
		Status: Active,
	})

	Channels[c.ChannelName] = &Channel{
		ChannelName:    c.ChannelName,
		ChannelClients: c.ChannelClients,
	}

	// add channel to client channels
	for i, item := range Clients {
		if item.ID == id {
			Clients[i].Channels = append(Clients[i].Channels, c.ChannelName)
		}
	}
}

func (c *Channel) Leave(id string) {

	if id == "" {
		return
	}

	for i, u := range c.ChannelClients {
		if u.ID == id {
			c.ChannelClients = append(c.ChannelClients[:i], c.ChannelClients[i+1:]...)
		}
	}

	Channels[c.ChannelName] = &Channel{
		ChannelName:    c.ChannelName,
		ChannelClients: c.ChannelClients,
	}

	// remove channel name from client channels
	for i, item := range Clients {
		if item.ID == id {
			for j, ch := range Clients[i].Channels {
				if ch == c.ChannelName {
					Clients[i].Channels = append(Clients[i].Channels[:j], Clients[i].Channels[j+1:]...)
				}
			}
		}
	}
}

func (c *Channel) ActiveClient(id string) {

	if id == "" {
		return
	}

	for i, ChannelClient := range c.ChannelClients {
		if ChannelClient.ID == id {
			c.ChannelClients[i].Status = Active
		}
	}

	Channels[c.ChannelName] = &Channel{
		ChannelName:    c.ChannelName,
		ChannelClients: c.ChannelClients,
	}

}

func (c *Channel) InactiveClient(id string) {

	if id == "" {
		return
	}

	for i, ChannelClient := range c.ChannelClients {
		if ChannelClient.ID == id {
			c.ChannelClients[i].Status = Inactive
		}
	}

	Channels[c.ChannelName] = &Channel{
		ChannelName:    c.ChannelName,
		ChannelClients: c.ChannelClients,
	}
}

func (c *Channel) FirstOrCreate(channelName string) error {

	if !channelExists(channelName) {

		newChannel := Channel{
			ChannelName:    channelName,
			ChannelClients: []ChannelClient{},
		}

		Channels[channelName] = &newChannel

	}

	c.ChannelName = Channels[channelName].ChannelName
	c.ChannelClients = Channels[channelName].ChannelClients

	if c.ChannelName == "" {
		return errors.New("channel not found")
	}

	return nil
}

func (c *Channel) InfoMessage() []byte {

	var joinedUsernames []map[string]interface{}

	for _, ChannelClient := range c.ChannelClients {
		client := FindByID(ChannelClient.ID)

		if client.Username == "" {
			// 	user.Username = id
			panic("username not found")
		}

		joinedUsernames = append(joinedUsernames, map[string]interface{}{
			"username":   client.Username,
			"connect_at": client.ConnectAt,
			"avatar":     client.Avatar,
			"status":     ChannelClient.Status,
		})
	}

	type Data struct {
		// ChannelName    string                   `json:"channel_name"`
		ChannelClients []map[string]interface{} `json:"users"`
	}

	info, err := json.Marshal(Message{
		Username:    "SERVER",
		ChannelName: c.ChannelName,
		Action:      "channel_info",
		Data: Data{
			// ChannelName:    c.ChannelName,
			ChannelClients: joinedUsernames,
		},
	})

	if err != nil {
		log.Panicln(err)
	}
	return info
}

func (c *Channel) InChannel(id string) bool {
	for _, u := range c.ChannelClients {
		if u.ID == id && u.Status == Active {
			return true
		}
	}
	return false
}

func (c *Channel) Broadcast(message []byte, m *melody.Melody) {

	m.BroadcastFilter([]byte(message), func(q *melody.Session) bool {

		// return true if the session id is in channel users
		return c.InChannel(q.Keys["id"].(string))
	})
}

func (c *Channel) BroadcastOther(s *melody.Session, message []byte, m *melody.Melody) {

	m.BroadcastFilter([]byte(message), func(q *melody.Session) bool {

		// return true if the session id is in channel users
		return c.InChannel(q.Keys["id"].(string)) && q != s
	})
}

func ChannelFirst(channelName string) (c Channel, err error) {

	if channelExists(channelName) {

		return *Channels[channelName], nil
	}

	return c, errors.New("channel not found")
}

func channelExists(name string) bool {

	return Channels[name] != nil

}
